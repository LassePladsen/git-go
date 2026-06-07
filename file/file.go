package file

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
)

func ReadCompressedFile(path string) (data []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	zr, err := zlib.NewReader(file)
	if err != nil {
		err = fmt.Errorf("Could not decompress file '%v': %w", path, err)
		return
	}
	defer zr.Close()

	data, err = io.ReadAll(zr)
	if err != nil {
		err = fmt.Errorf("Could not read decompressed data for '%v': %w", path, err)
		return
	}
	return
}

func ReadFile(path string) (data []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	data, err = io.ReadAll(file)
	if err != nil {
		return
	}
	return
}

// Whether file or dir exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
