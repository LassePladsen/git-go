package main

import (
	"fmt"
	"os"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	var output Output
	var err error
	switch command := os.Args[1]; command {
	case "init":
		// disabled for safety
		// output, err = init_git() 

	default:
		err = fmt.Errorf("Unknown command %s", command)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(output) != 0 {
		os.Stdout.Write(output)
	}
}
