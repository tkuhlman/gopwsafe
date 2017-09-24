// Start up the gtk interface if possible, if not fall back to the cli interface

package main

import (
	"log"
	"os"

	"github.com/tkuhlman/gopwsafe/gui"
)

func main() {
	app, err := gui.NewGoPWSafeGTK()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(app.Run(os.Args))
}
