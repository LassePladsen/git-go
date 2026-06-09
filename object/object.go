package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
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
	return fmt.Sprintf("RawObject{\n\tPath: %v\n\tHash: %v\n\tKind: %v\n\tSize: %v\n\tData: %v\n}",
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
	Entries []*TreeEntry
}
type TreeEntry struct {
	// Permissions mode. NB: Directory mode 040000 is stored as 40000
	Mode string
	Name string
	Hash string
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
	var tree Tree

	// Loop entries until rest data is empty
	// Each entry's format: <mode> <name>\0<20_byte_object_hash>
	rest := treeObj.Data
	for len(rest) > 0 {
		var entry TreeEntry

		var i int
		var mode []byte
		for i = range rest {
			b := rest[i]
			if b == ' ' {
				break
			}
			mode = append(mode, b)
		}
		if i >= len(rest) || rest[i] != ' ' {
			return nil, errors.New("Malformed tree entry mode")
		}
		entry.Mode = string(mode)

		rest = rest[i+1:] // skip the space

		// Read name
		var name []byte
		for i = range rest {
			b := rest[i]
			if b == 0 {
				break
			}
			name = append(name, b)
		}
		if i >= len(rest) || rest[i] != 0 {
			return nil, errors.New("Malformed tree entry name")
		}
		entry.Name = string(name)

		rest = rest[i+1:] // skip the null byte

		// Read 20 byte hash (its not stored as hex)
		if len(rest) < 20 {
			return nil, errors.New("Malformed tree entry hash")
		}
		entry.Hash = string(rest[:20])
		tree.Entries = append(tree.Entries, &entry)

		rest = rest[20:]
	}
	return &tree, nil

}

// serialize a Tree into tree object's data bytes ready for WriteObject
// TODO: some error here in output...
func EncodeTree(tree *Tree) ([]byte, error) {
	var data []byte

	// Loop entries until rest data is empty
	for _, entry := range tree.Entries {
		// dir mode 040000 should be stored as 40000
		mode := entry.Mode
		if mode == "040000" {
			mode = "40000"
		}

		// Each entry's format: <mode> <name>\0<20_byte_object_hash>
		// and there is no separator between entries. 

		//note: hash will be output as 20 byte raw hash, not hex
		rawHash, err := hex.DecodeString(entry.Hash)
		if err != nil {
			return nil, fmt.Errorf("Could not decode hash '%v' encode tree: %w", entry.Hash, err)
		}

		data = fmt.Appendf(data, "%v %v\x00%v", mode, entry.Name, string(rawHash))
	}

	return data, nil
}

// These files or dirs will be ignored when writing a tree
var filesIgnored map[string]bool = map[string]bool{
	".mygit": true,
	".git":   true,
	// TODO: implement .mygitignore
}

// Walks directory path recursively, writes RawObjects for each entry, and finally writes the root Tree object
func WriteTree(path string) (*RawObject, error) {
	dirEntries, err := os.ReadDir(path) // entires are sorted by name, so we don't need to handle this ourselves
	if err != nil {
		return nil, fmt.Errorf("Could not read working directory: %v\n", err)
	}

	// Iterate over files in path, create blobs for files and trees for dirs
	entries := make([]*TreeEntry, len(dirEntries))
	var n int
	for _, dirEntry := range dirEntries {
		if _, ok := filesIgnored[dirEntry.Name()]; ok {
			// ignored path, should make entires dynamically sliced with append() or somehow slice this index away?
			continue
		}
		entryPath := filepath.Join(path, dirEntry.Name())

		if dirEntry.IsDir() {
			obj, err := WriteTree(entryPath)
			fmt.Printf("LP writing dir path '%v' to hash: %v\n", entryPath, obj.Hash)
			if err != nil {
				return nil, err
			}
			entries[n] = &TreeEntry{Hash: obj.Hash, Mode: "40000", Name: dirEntry.Name()}
		} else { // file
			// open file and read
			file, err := os.Open(entryPath)
			if err != nil {
				return nil, fmt.Errorf("Could not open file '%v' in to write tree: %w", dirEntry.Name(), err)
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("Could not read file '%v' in to write tree: %w", dirEntry.Name(), err)
			}

			// Write object
			obj, err := WriteObject(data, KindBlob)
			fmt.Printf("LP wrote filepath '%v' to object hash: %v\n", entryPath, obj.Hash)

			if err != nil {
				return nil, fmt.Errorf("Could not write object for file '%v' to write tree: %w", dirEntry.Name(), err)
			}

			// Store entry for tree
			mode, err := ParseFileMode(file)
			if err != nil {
				return nil, fmt.Errorf("Could not parse file mode to git mode for file '%v' to write tree: %w", dirEntry.Name(), err)
			}
			entries[n] = &TreeEntry{Name: dirEntry.Name(), Mode: mode, Hash: obj.Hash}
		}
		n++
	}

	tree := Tree{Entries: entries[:n]} // slice away the end to skip the nil entries we ignored
	data, err := EncodeTree(&tree)
	if err != nil {
		return nil, fmt.Errorf("Could not encode tree for path '%v': %w", path, err)
	}
	return WriteObject(data, KindTree)
}

// File mode to git mode, e.g "40000" for dirs
func ParseFileMode(f *os.File) (string, error) {
	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("Could not stat file '%v' to parse mode: %w", f.Name(), err)
	}
	mode := stat.Mode()
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
