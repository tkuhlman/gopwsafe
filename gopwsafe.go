// Start up the gtk interface if possible, if not fall back to the cli interface

package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/tkuhlman/gopwsafe/cli"
)

func main() {
	useCli := flag.Bool("c", false, "Use the cli interface, normal behavior is to try gtk and fall back to the cli")
	dbFile := flag.String("f", "", "Path of the password database to open.")
	flag.Parse()

	if !*useCli {
		log.Error("No gui interface yet implemented")
	}

	cli.Start(*dbFile)
}
