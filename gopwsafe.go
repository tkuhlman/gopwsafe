// Start up the gtk interface if possible, if not fall back to the cli interface

package main

import (
	"./cli"
	"flag"
	"fmt"
)

func main() {
	useCli := flag.Bool("cli", false, "Use the cli interface, normal behavior is to try gtk and fall back to the cli")
	dbFile := flag.String("file", "", "Path of the password database to open.")
	flag.Parse()

	if !*useCli {
		fmt.Println("No gui interface yet implemented")
	}

	cli.CLIInterface(*dbFile)
}
