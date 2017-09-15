// Start up the gtk interface if possible, if not fall back to the cli interface

package main

import (
	"flag"
	"log"
	"os"

	"github.com/tkuhlman/gopwsafe/gui"
)

func main() {
	dbFile := flag.String("f", "", "Path of the password database to open.")
	flag.Parse()

	app, err := gui.NewGoPWSafeGTK()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(app.Open(*dbFile))
}
