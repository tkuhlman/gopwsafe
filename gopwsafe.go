// Start up the gtk interface if possible, if not fall back to the cli interface

package main

import (
	"flag"
	"os"

	"github.com/tkuhlman/gopwsafe/cli"
	"github.com/tkuhlman/gopwsafe/gui"
)

func main() {
	useCli := flag.Bool("c", false, "Use the cli interface, normal behavior is to try gtk and fall back to the cli")
	dbFile := flag.String("f", "", "Path of the password database to open.")
	flag.Parse()

	var exitCode int
	if !*useCli {
		exitCode = gui.Start(*dbFile)
	} else {
		exitCode = cli.Start(*dbFile)
	}

	os.Exit(exitCode)
}
