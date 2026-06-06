package main

import (
	"os"
)

type Output = []byte

func init_git() (output Output, err error) {
	if err = os.Mkdir(".git", 0775); err != nil {
		return
	}
	if err = os.Mkdir(".git/objects", 0775); err != nil {
		return
	}
	if err = os.Mkdir(".git/refs", 0775); err != nil {
		return
	}
	if err = os.WriteFile(".git/HEAD", []byte("ref: refs/heads/main\n"), 0664); err != nil {
		return
	}
	output = []byte("Initialized git directory")
	return
}

/*
/// Read blob
pub fn cat_file(args: &[String]) -> Output {
    // Get blob hash from positional arg
    let mut hash: Option<&str> = None;
    for arg in &args[2..] {
        // Flag argument, skip for now. TODO: support flags?
        if arg.starts_with('-') {
            continue;
        }
        hash = Some(arg);
    }
    let Some(hash) = hash else {
        return "Missing hash\n".into();
    };
    if hash.len() < 3 {
        return "Hash name too short\n".into();
    }
    let object = match Object::read(hash) {
        Ok(object) => object,
        Err(e) => return format!("{e}").into(),
    };
    object.contents
}

/// Hash object to blob
pub fn hash_object(args: &[String]) {
    // Get path from positional arg
    let mut path: Option<&str> = None;
    for arg in &args[2..] {
        // Flag argument, skip for now. TODO: support flags?
        if arg.starts_with('-') {
            continue;
        }
        path = Some(arg);
    }
    let Some(path) = path else {
        println!("Missing path");
        return;
    };

    // Read file
    let bytes = match fs::read(path) {
        Ok(bytes) => bytes,
        Err(err) => {
            println!("{err}");
            return;
        }
    };
    let object = Object {contents: bytes, kind: ObjectKind::Blob, size: bytes.len(): usize};
}

/// Inspect tree object
pub fn ls_tree(args: &[String]) {
    // Get tree hash input from positional arg
    let mut hash: Option<&str> = None;
    let mut print_name_only = false;
    for arg in &args[2..] {
        // Supports flag: --name-only
        if "--name-only" == arg {
            print_name_only = true;
        }
        if arg.starts_with('-') {
            continue;
        }
        hash = Some(arg);
    }
    let Some(hash) = hash else {
        println!("Missing hash");
        return;
    };

    // TODO: also support full print (where print_name_only=false)

    // Read file
    let path = object::get_path(hash);
    // let file = File
}
*/
