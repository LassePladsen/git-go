package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"mygit/file"
	"os"
	"path/filepath"
	"strconv"
)

// Object kind enum
type Kind string

const (
	KindBlob Kind = "blob"
	KindTree Kind = "tree"
)

func ParseKind(s string) (Kind, error) {
	switch kind := Kind(s); kind {
	case KindBlob, KindTree:
		return kind, nil
	}
	return "", fmt.Errorf("Unsupported kind: %v", s)
}

type Object struct {
	Hash     string
	Kind     Kind
	Contents []byte
}

func (o Object) String() string {
	return fmt.Sprintf("Object{\n\tPath: %v\n\tHash:% v\n\tKind: %v\n\tSize: %v\n\tContents: %v\n}",
		o.Path(), o.Hash, o.Kind, o.Size(), string(o.Contents))
}

// Size of data
func (o Object) Size() int {
	return len(o.Contents)
}

// File path to object
func (o Object) Path() string {
	return HashToPath(o.Hash)
}

// Creates file path from object hash. Example: 1eadkl351341k123jlk21WDad -> .mygit/objects/1e/adkl351341k123jlk21WDad
func HashToPath(hash string) string {
	dir := hash[0:2]
	filename := hash[2:]
	return fmt.Sprintf(".mygit/objects/%v/%v", dir, filename)
}

// Reads object to Object struct
func Open(hash string) (*Object, error) {
	filePath := HashToPath(hash)
	data, err := file.ReadCompressedFile(filePath)
	if err != nil {
		return nil, err
	}

	// Header: <object_kind> <size>\0
	// Read object kind up to a space
	obj := Object{Hash: hash}
	var buf []byte
	var i int
	var b byte
	for i, b = range data {
		if b == ' ' {
			break
		}
		buf = append(buf, b)
	}

	kind, err := ParseKind(string(buf))
	if err != nil {
		return nil, err
	}
	obj.Kind = kind

	// Read size up to null byte
	buf = buf[:0]
	rest := data[i+1:] // skip the space with i+1
	for i, b = range rest {
		if b == 0 {
			break
		}
		buf = append(buf, b)

	}
	size, err := strconv.Atoi(string(buf))
	if err != nil {
		return nil, err
	}
	fmt.Println("LP size: ", size) // TODO: do i need size? maybe use size in slice below just to use it?

	// The rest is the actual object contents, we are done with header after null byte
	// we don't actually need the size since i've loaded the entire data into a slice
	obj.Contents = rest[i+1:]
	return &obj, nil
}

// Compress data and write to object file, also returns the Object
func Create(data []byte, kind Kind) (*Object, error) {
	// Object format: <object_kind> <size>\0<data>
	size := len(data)

	objData := fmt.Appendf(nil, "%v %v\x00%v", kind, size, string(data))
	sum := sha1.Sum(objData)
	hash := fmt.Sprintf("%x", sum)
	obj := Object{Kind: kind, Contents: data, Hash: hash}

	// Compress data and write to file
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(objData); err != nil {
		return nil, fmt.Errorf("Could not compress data for hash '%v': %w", hash, err)
	}
	zw.Close()

	path := HashToPath(hash)
	dir := filepath.Dir(path)
	if !file.Exists(dir) {
		if err := os.Mkdir(dir, 0775); err != nil {
			return nil, fmt.Errorf("Could not mkdir for object '%v': %w", path, err)
		}
	}
	if err := os.WriteFile(path, buf.Bytes(), 0664); err != nil {
		return nil, fmt.Errorf("Could not write to file '%v': %w", path, err)
	}
	return &obj, nil
}

type Tree struct {
	Obj     Object
	Entries []TreeEntry
}
type TreeEntry struct {
	Mode   []byte
	Name   []byte
	Hash   []byte
	Object *Object
}

// Parse tree entries. if openEntryObjects then each entry object is opened and read into TreeEntry.Object
func ReadTree(treeObj *Object, openEntryObjects bool) (*Tree, error) {
	if treeObj.Kind != KindTree {
		return nil, errors.New("Not a tree object")
	}
	// read format: <mode> <name>\0<20_byte_object_hash>
	tree := Tree{Obj: *treeObj}

	// Loop entries until rest data is empty
	rest := treeObj.Contents
	for len(rest) > 0 {
		var entry TreeEntry

		// Read mode
		var i int
		for i = range rest {
			b := rest[i]
			if b == ' ' {
				break
			}
			entry.Mode = append(entry.Mode, b)
		}
		if i >= len(rest) || rest[i] != ' ' {
			return nil, errors.New("Malformed tree entry mode")
		}
		rest = rest[i+1:] // skip space

		// Read name
		for i = range rest {
			b := rest[i]
			if b == 0 {
				break
			}
			entry.Name = append(entry.Name, b)
		}
		if i >= len(rest) || rest[i] != 0 {
			return nil, errors.New("Malformed tree entry name")
		}
		rest = rest[i+1:] // skip null byte

		// Read 20 byte hash (its not stored as hex)
		if len(rest) < 20 {
			return nil, errors.New("Malformed tree entry hash")
		}
		entry.Hash, rest = rest[:20], rest[20:]

		// Unless openEntryObjects, we need to open each object
		if openEntryObjects {
			entryObj, err := Open(fmt.Sprintf("%x", entry.Hash))
			if err != nil {
				return nil, fmt.Errorf("Could not open tree entry object '%x': %w", entry.Hash, err)
			}
			entry.Object = entryObj
		}

		tree.Entries = append(tree.Entries, entry)
	}
	return &tree, nil

}

// Write tree object recursively for the given dir path
func WriteTree(path string) (*Tree, error) {
	dirEntries, err := os.ReadDir(path) // entires are sorted by name, so we don't need to handle this ourselves
	if err != nil {
		return nil, fmt.Errorf("Could not read working directory: %v\n", err)
	}

	// Iterate over files in cwd, create blobs for files and trees for dirs
	type entry struct {
		obj  *Object
		name string
	}
	entries := make([]*entry, len(dirEntries))
	for i, dirEntry := range dirEntries {
		fmt.Println("LP dirEntry: ", dirEntry)
		// TODO: implement mygitignore
		if dirEntry.IsDir() {
			continue // TODO: dirs
		} else { // file
			data, err := file.ReadFile(filepath.Join(path, dirEntry.Name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read file in working directory to write tree: %v\n", err)
				os.Exit(1)
			}
			entryObj, err := Create(data, KindBlob)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not write data for file '%v' to blob for write tree: %v\n", dirEntry.Name(), err)
				os.Exit(1)
			}
			entries[i] = &entry{obj: entryObj, name: dirEntry.Name()}
		}

	}

	// // FIXME: delete this block
	// for _, e := range entries {
	// 	if e == nil {continue}
	// 	fmt.Println("LP entry: ", e.name)
	// }

	return &Tree{}, nil
}
