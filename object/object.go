package object

func Read() {}
/*
type ObjectKind
 enum ObjectKind {
    Blob,
    Tree,
}

type object struct {
     kind ObjectKind,
     size uint
     contents []byte,
}

impl Object {
     func is_kind(&self, kind: ObjectKind) -> bool {
        self.kind == kind
    }

    /// Ensures the object is of the given kind by returning Err with premade message.
     func ensure_kind(&self, kind: ObjectKind) -> Result<()> {
        if !self.is_kind(kind) {
            bail!(
                "Unexpected object kind '{:?}', expected '{:?}'\n",
                self.kind,
                ObjectKind::Blob
            );
        }
        Ok(())
    }
}

impl Object {
     func write() -> Object {
        todo!()
    }

    /// Reads and decompressed object contents
     func read(object_hash: &str) -> Result<Object> {
        let file_path = get_path(object_hash);
        let file = fs::File::open(&file_path)
            .with_context(|| format!("Could not open file '{file_path}'"))?;
        let mut decoder = compression::Decoder::new(file);
        let mut decompressed = Vec::new();
        let _ = decoder.read_to_end(&mut decompressed).with_context(|| {
            format!("Could not read decompressed bytes to end for file '{file_path}'")
        })?;

        // Convert contents to vec to efficiently pop from the front
        let mut decompressed = VecDeque::from(decompressed);

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
        let type_: ObjectKind =
            match str::from_utf8(&type_buf).expect("Invalid UTF in object type buffer") {
                "blob" => ObjectKind::Blob,
                "tree" => ObjectKind::Tree,
                s => panic!("Invalid type: {s}"),
            };
        drop(type_buf);

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
            kind: type_,
            contents,
        })
    }
}

 func get_path(object_hash: &str) -> String {
    let dir = &object_hash[0..2];
    let filename = &object_hash[2..];
    format!(".git/objects/{dir}/{filename}")
}
*/
