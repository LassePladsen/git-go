package main

import (
	"errors"
	"os"
	"mygit/object"
)

type Output = []byte
type Args = []string
type Command = func(args Args) (Output, error)

func initGit(_ Args) (output Output, err error) {
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

// Get file contents
func cat_file(args Args) (output Output, err error) {
	// Get blob hash from positional arg
	var hash string
	for _, arg := range args[2:] {
		// Flag argument, skip for now. TODO: support flags?
		if arg[0] == byte('-') {
			hash = arg
			break
		}
	}
	if hash == "" {
		err = errors.New("Missing hash")
		return
	}
	if len(hash) < 3 {
		err = errors.New("Hash name too short")
		return
	}
	object, err := object.Read(hash)
	if err != nil {
		return
	}
	output = object.contents
	return
}

/*
/// Hash object to blob
func hash_object(args Args) {
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
func ls_tree(args Args) {
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
