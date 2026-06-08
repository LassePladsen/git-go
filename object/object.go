package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"mygit/file"
	"os"
	"path/filepath"
)

// RawObject kind enum
type Kind string

const (
	KindBlob   Kind = "blob"
	KindTree   Kind = "tree"
	KindCommit Kind = "commit"
	KindTag    Kind = "tag"
)

func ParseKind(s string) (Kind, error) {
	switch kind := Kind(s); kind {
	case KindBlob, KindTree:
		return kind, nil
	}
	return "", fmt.Errorf("Unsupported kind: %v", s)
}

// represents a stored object file
type RawObject struct {
	Hash string
	Kind Kind
	Data []byte
}

func (o RawObject) String() string {
	return fmt.Sprintf("RawObject{\n\tPath: %v\n\tHash:% v\n\tKind: %v\n\tSize: %v\n\tData: %v\n}",
		o.Path(), o.Hash, o.Kind, o.Size(), string(o.Data))
}

// Size of data
func (o RawObject) Size() int {
	return len(o.Data)
}

// File path to object
func (o RawObject) Path() string {
	return HashToPath(o.Hash)
}

// Creates file path from object hash. Example: 1eadkl351341k123jlk21WDad -> .mygit/objects/1e/adkl351341k123jlk21WDad
func HashToPath(hash string) string {
	dir := hash[0:2]
	filename := hash[2:]
	return fmt.Sprintf(".mygit/objects/%v/%v", dir, filename)
}

// Reads object to RawObject
func OpenObject(hash string) (*RawObject, error) {
	filePath := HashToPath(hash)
	data, err := file.ReadCompressedFile(filePath)
	if err != nil {
		return nil, err
	}

	// Header: <object_kind> <size>\0
	// Read object kind up to a space
	obj := RawObject{Hash: hash}
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
	// TODO: do i need size at all? maybe try to use it in slice below..?
	// _, err = strconv.Atoi(string(buf))
	// if err != nil {
	// 	return nil, err
	// }

	// The rest is the actual object data, we are done with header after null byte
	// we don't actually need the size since i've loaded the entire data into a slice
	obj.Data = rest[i+1:]
	return &obj, nil
}

// Compress data and write to object file, and return the RawObject
func WriteObject(data []byte, kind Kind) (*RawObject, error) {
	// Object format: <object_kind> <size>\0<data>
	size := len(data)

	objData := fmt.Appendf(nil, "%v %v\x00%v", kind, size, string(data))
	sum := sha1.Sum(objData)
	hash := fmt.Sprintf("%x", sum)
	obj := RawObject{Kind: kind, Data: data, Hash: hash}

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
	Entries []TreeEntry
}
type TreeEntry struct {
	// Permissions mode. NB: Directory mode 040000 is stored as 40000
	Mode []byte
	Name []byte
	Hash []byte
}

func (e TreeEntry) Kind() (kind Kind, err error) {
	switch mode := string(e.Mode); mode {
	case "40000":
		kind = KindTree
	case "100644", "100755", "120000":
		kind = KindBlob
	case "160000":
		kind = KindCommit
	default:
		err = fmt.Errorf("TreeEntry has unsupported mode: %v", mode)
	}
	return
}

// Parse tree entries. if openEntryObjects then each entry object is opened and read into TreeEntry.Object
func ParseTree(treeObj *RawObject) (*Tree, error) {
	if treeObj.Kind != KindTree {
		return nil, errors.New("Not a tree object")
	}
	// read format: <mode> <name>\0<20_byte_object_hash>
	var tree Tree

	// Loop entries until rest data is empty
	rest := treeObj.Data
	for len(rest) > 0 {
		var entry TreeEntry

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

		tree.Entries = append(tree.Entries, entry)
	}
	return &tree, nil

}

// serialize a Tree into tree payload bytes ready to WriteObject
func EncodeTree(tree *Tree) ([]byte, error) {
	return []byte{}, nil
}

// TODO:
// Walks directory path recursively, writes RawObjects for each entry, and finally writes the root Tree object
func WriteTree(path string) (*RawObject, error) {
	dirEntries, err := os.ReadDir(path) // entires are sorted by name, so we don't need to handle this ourselves
	if err != nil {
		return nil, fmt.Errorf("Could not read working directory: %v\n", err)
	}

	// Iterate over files in path, create blobs for files and trees for dirs
	entries := make([]TreeEntry, len(dirEntries))
	for i, dirEntry := range dirEntries {
		i = i // FIXME: delete
		fmt.Println("LP dirEntry: ", dirEntry)
		// TODO: implement mygitignore
		if dirEntry.IsDir() {
			continue // TODO: dirs
		} else { // file
			// open file and read
			filePath := filepath.Join(path, dirEntry.Name())
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("Could not open file '%v' in to write tree: %w\n", dirEntry.Name(), err)
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("Could not read file '%v' in to write tree: %w\n", dirEntry.Name(), err)
			}

			// Write object
			_, err = WriteObject(data, KindBlob)
			if err != nil {
				return nil, fmt.Errorf("Could not write object for file '%v' to write tree: %w\n", dirEntry.Name(), err)
			}
			stat, err := file.Stat()
			if err != nil {
				return nil, fmt.Errorf("Could not stat file '%v' to write tree: %w\n", dirEntry.Name(), err)
			}

			// Store entry for tree
			mode, err := gitMode(stat.Mode())
			if err != nil {
				return nil, fmt.Errorf("Could not parse file mode to git mode for file '%v' to write tree: %w\n", dirEntry.Name(), err)
			}
			entries[i] = TreeEntry{Name: []byte(dirEntry.Name()), Mode: []byte(mode)}
		}

	}
	fmt.Println("LP entries: ", entries)
	os.Exit(1)

	// // FIXME: delete this block
	// for _, e := range entries {
	// 	if e == nil {continue}
	// 	fmt.Println("LP entry: ", e.name)
	// }

	tree := Tree{Entries: entries}

	data, err := EncodeTree(&tree)
	data = data // FIXME: delete
	if err != nil {
		return nil, err
	}

	// return WriteObject(KindTree, data)
	return nil, nil
}

func gitMode(mode os.FileMode) (string, error) {
	switch {
	case mode.IsDir():
		return "40000", nil
	case mode&os.ModeSymlink != 0:
		return "120000", nil
	case mode.IsRegular():
		if mode&0o111 != 0 {
			return "100755", nil
		}
		return "100644", nil
	default:
		return "", fmt.Errorf("unsupported file mode: %v", mode)
	}
}
