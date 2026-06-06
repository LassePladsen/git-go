package object

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

type ObjectType string

const (
	TypeBlob ObjectType = "blob"
	TypeTree ObjectType = "tree"
)

type Object struct {
	kind     ObjectType
	size     uint
	contents []byte
}

// Creates file path from object hash. Example: 1eadkl351341k123jlk21WDad -> .git/objects/1e/adkl351341k123jlk21WDad
func HashToPath(hash string) string {
	dir := hash[0:2]
	filename := hash[2:]
	return fmt.Sprintf(".git/objects/%v/%v}", dir, filename)
}

// Read file and decompresses
func readCompressedFile(path string) (data []byte, err error) {
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

// Reads object to Object struct
func Open(hash string) (obj Object, err error) {
	filePath := HashToPath(hash)
	data, err := readCompressedFile(filePath)
	if err != nil {
		return
	}
	fmt.Println("LP data: ", data)
	return

	/*
		// Header: <object_type><size>null_byte
		// Read type up to space
		let mut type_buf = Vec::new();
		while let Some(byte) = decompressed.pop_front() {
			if b' ' == byte {
				break;
			}
			type_buf.push(byte);
		}
		// convert type to string and interpret it
		let type_: ObjectType =
		match str::from_utf8(&type_buf).expect("Invalid UTF in object type buffer") {
		"blob" => ObjectType::Blob,
		"tree" => ObjectType::Tree,
		s => panic!("Invalid type: {s}"),
	*/
}

/*
// Read size up to null byte
let mut size_buf = Vec::new();
while let Some(byte) = decompressed.pop_front() {
	if 0 == byte {
		break;
	}
	size_buf.push(byte);
}
let size: usize = str::from_utf8(&size_buf)
.expect("Invalid UTF in size type buffer")
.parse()
.expect("Could not parse size byte to integer");

// Read rest of contents
let contents = Vec::from(decompressed);

Ok(Object {
	size,
	type: type_,
	contents,
})
}
}

/*
impl Object {
func is_type(&self, type: ObjectType) -> bool {
self.type == type
}

/// Ensures the object is of the given type by returning Err with premade message.
func ensure_type(&self, type: ObjectType) -> Result<()> {
if !self.is_type(type) {
bail!(
"Unexpected object type '{:?}', expected '{:?}'\n",
self.type,
ObjectType::Blob
);
}
Ok(())
}
}

impl Object {
func write() -> Object {
todo!()
}

}

func get_path(object_hash: &str) -> String {
let dir = &object_hash[0..2];
let filename = &object_hash[2..];
format!(".git/objects/{dir}/{filename}")
}
*/
