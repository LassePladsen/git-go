package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"mygit/file"
	"mygit/object"
	"os"
	"slices"
)

type Command = func(args []string) []byte

var commands = map[string]Command{
	"init":        initGit,
	"cat-file":    catFile,
	"hash-object": hashObject,
	"ls-tree":     lsTree,
	"write-tree":  writeTree,
}

func initGit(_ []string) (output []byte) {
	if err := os.Mkdir(".mygit", 0775); err != nil {
		fmt.Fprintf(os.Stderr, "Could not mkdir '.mygit': %v\n", err)
		os.Exit(1)
	}
	if err := os.Mkdir(".mygit/objects", 0775); err != nil {
		fmt.Fprintf(os.Stderr, "Could not mkdir '.mygit/objects' %v\n", err)
		os.Exit(1)
	}
	if err := os.Mkdir(".mygit/refs", 0775); err != nil {
		fmt.Fprintf(os.Stderr, "Could not mkdir '.mygit/refs' %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(".mygit/HEAD", []byte("ref: refs/heads/main\n"), 0664); err != nil {
		fmt.Fprintf(os.Stderr, "Could not write to file '.mygit/HEAD': %v\n", err)
		os.Exit(1)
	}
	output = []byte("Initialized mygit directory\n")
	return
}

// return object data
func catFile(args []string) []byte {
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
		fmt.Fprintf(os.Stderr, "Not a valid object name: %v\n", hash)
		os.Exit(1)
	}
	obj, err := object.OpenObject(hash)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Not a valid object name: %v\n", hash)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	return obj.Data
}

// Hash file to object
func hashObject(args []string) []byte {
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

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var kind object.Kind
	var data []byte
	if fileInfo.IsDir() {
		// TODO: hash dir to tree?
		kind = object.KindTree

	} else { // file
		kind = object.KindBlob
		// Read file
		data, err = file.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Write to object
	obj, err := object.WriteObject(data, kind)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return fmt.Appendln(nil, obj.Hash)
}

// Inspect tree object
// Tree structure: <object_kind> <size>\0<tree_entries>
// where <tree_entries> is each a <mode> <name>\0<20_byte_object_hash>
func lsTree(args []string) []byte {
	helpMsg := "usage: mygit ls-tree [arguments] <object_hash>"
	// Get tree hash input from positional arg
	var hash string
	nameOnly := false
	for _, arg := range args[2:] {
		// Supports flag: --name-only
		if "--name-only" == arg {
			nameOnly = true
		}
		if arg[0] != byte('-') {
			hash = arg
		}
	}

	if hash == "" {
		fmt.Println(helpMsg)
		os.Exit(0)
	}
	if len(hash) < 3 {
		fmt.Fprintf(os.Stderr, "Not a valid object name: %v\n", hash)
		os.Exit(1)
	}

	obj, err := object.OpenObject(hash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read object: %v\n", err)
		os.Exit(1)
	}

	tree, err := object.ParseTree(obj)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var out []byte
	for _, entry := range tree.Entries {
		if nameOnly {
			// only output name
			out = append(out, entry.Name...)
		} else {
			// output: <mode> <kind> <hash>    <name>
			hash := make([]byte, hex.EncodedLen(len(entry.Hash)))
			hex.Encode(hash, entry.Hash)

			kind, err := entry.Kind()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not get tree entry kind: %v", err)
				os.Exit(1)
			}

			// dir mode 040000 is stored as 40000, so prepend the zero
			if string(entry.Mode) == "40000" {
				entry.Mode = slices.Insert(entry.Mode, 0, '0')
			}

			out = append(out, entry.Mode...)
			out = append(out, ' ')
			out = append(out, []byte(kind)...)
			out = append(out, ' ')
			out = append(out, hash...)
			out = append(out, []byte("    ")...)
			out = append(out, entry.Name...)
			out = fmt.Appendln(out)
		}
	}

	return fmt.Appendln(out)
}

// Create a tree object from current state of "staging area" (from git add)
// For now, every file in the working dir are already staged. TODO: implement git add and staging area
func writeTree(args []string) []byte {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get cwd: %v\n", err)
		os.Exit(1)
	}
	obj, err := object.WriteTree(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not write tree for working dir: %v\n", err)
		os.Exit(1)
	}
	return fmt.Appendln(nil, obj.Hash)
}
