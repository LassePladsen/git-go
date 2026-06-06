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
	var commandFn Command
	commandName := os.Args[1]
	commandFn, ok := commands[commandName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %s\n", commandName)
		os.Exit(1)
	}

	output, err = commandFn(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(output) != 0 {
		os.Stdout.Write(output)
	}
}
