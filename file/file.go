package file

import (
	"compress/zlib"
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
		return
	}
	defer zr.Close()

	data, err = io.ReadAll(zr)
	if err != nil {
		return
	}
	return
}
