package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/fs"
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

type Object struct {
	Hash     string
	Kind     Kind
	Size     uint
	Contents []byte
}

// Creates file path from object hash. Example: 1eadkl351341k123jlk21WDad -> .git/objects/1e/adkl351341k123jlk21WDad
func HashToPath(hash string) string {
	dir := hash[0:2]
	filename := hash[2:]
	return fmt.Sprintf(".git/objects/%v/%v", dir, filename)
}

// Reads object to Object struct
func Open(hash string) (*Object, error) {
	filePath := HashToPath(hash)
	data, err := file.ReadCompressedFile(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Not a valid object name: %v\n", hash)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	// Header: <object_kind> <size>\0
	// Read object kind up to a space
	var buf []byte
	var i int
	var b byte
	for i, b = range data {
		if b == ' ' {
			break
		}
		buf = append(buf, b)
	}

	obj := Object{Hash: hash}
	switch kind := Kind(buf); kind {
	case "blob":
		obj.Kind = KindBlob
	case "tree":
		obj.Kind = KindTree
	default:
		fmt.Fprintf(os.Stderr, "Unsupported kind %v\n", kind)
		os.Exit(1)
	}

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
	obj.Size = uint(size)

	// The rest is the actual object contents, we are done with header after null byte
	// we don't actually need the size since i've loaded the entire data into a slice
	obj.Contents = rest[i+1:]
	return &obj, nil
}

// Compress data and write to object file, also returns the Object
func Write(data []byte) (*Object, error) {
	// Object format: <object_kind> <size>\0<data>
	kind := KindBlob // for now only supports blobs. TODO:
	size := len(data)
	out := fmt.Sprintf("%v %v\x00%v", kind, size, string(data))
	sum := sha1.Sum([]byte(out))
	hash := fmt.Sprintf("%x", sum)
	obj := Object{Kind: kind, Size: uint(size), Contents: data, Hash: hash}

	// Compress hash and write to file
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	zw.Write([]byte(hash))

	path := HashToPath(hash)
	dir := filepath.Dir(path)
	if !file.Exists(dir) {
		if err := os.Mkdir(dir, 0775); err != nil {
			fmt.Fprintf(os.Stderr, "Could not mkdir for object '%v': %v\n", path, err)
			os.Exit(1)
		}
	}
	if err := os.WriteFile(path, []byte(hash), 0664); err != nil {
		fmt.Fprintf(os.Stderr, "Could not write to file: %v\n", path)
		os.Exit(1)
	}
	return &obj, nil
}
