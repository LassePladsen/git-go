package main

import (
	"fmt"
	"mygit/file"
	"mygit/object"
	"os"
)

type Output = []byte
type Args = []string
type Command = func(args Args) Output

var commands = map[string]Command{
	"init":        initGit,
	"cat-file":    catFile,
	"hash-object": hashObject,
}

func initGit(_ Args) (output Output) {
	if err := os.Mkdir(".git", 0775); err != nil {
		fmt.Fprintln(os.Stderr, "Could not mkdir for .git")
		os.Exit(1)
	}
	if err := os.Mkdir(".git/objects", 0775); err != nil {
		fmt.Fprintln(os.Stderr, "Could not mkdir: .git/objects")
		os.Exit(1)
	}
	if err := os.Mkdir(".git/refs", 0775); err != nil {
		fmt.Fprintln(os.Stderr, "Could not mkdir: .git/refs")
		os.Exit(1)
	}
	if err := os.WriteFile(".git/HEAD", []byte("ref: refs/heads/main\n"), 0664); err != nil {
		fmt.Fprintln(os.Stderr, "Could not write to file: .git/HEAD")
		os.Exit(1)
	}
	output = []byte("Initialized git directory")
	return
}

// return object contents
func catFile(args Args) Output {
	helpMsg := "usage: mygit cat-file <object_hash>"
	// Get blob hash from positional arg
	var hash string
	for _, arg := range args[2:] {
		// Flag argument, skip for now. TODO: support flags?
		if arg[0] != byte('-') {
			hash = arg
			break
		}
	}

	if hash == "" {
		fmt.Println(helpMsg)
		os.Exit(0)
	}
	if len(hash) < 3 {
		fmt.Fprintf(os.Stderr, "Not a valid object name: %v", hash)
		os.Exit(1)
	}
	obj, err := object.Open(hash)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return obj.Contents
}

// TODO:
// Hash file to object
func hashObject(args Args) Output {
	helpMsg := "usage: mygit hash-object <file_path>"
	// Get path from positional arg
	var path string
	for _, arg := range args[2:] {
		// Flag argument, skip for now. TODO: support flags?
		if arg[0] != byte('-') {
			path = arg
			break
		}
	}
	if path == "" {
		fmt.Println(helpMsg)
		os.Exit(0)
	}

	// Read file
	data, err := file.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Write to object
	obj, err := object.Write(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return Output(fmt.Sprintln(obj.Hash))
}

/*
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
