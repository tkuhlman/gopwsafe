/*

* update the Readme with info on what I decide to use.
Options include:
	- termbox go - https://github.com/nsf/termbox-go - this could lead to a nice cli interface but may be overkill.
	- just using bufio.ReadLine(),
		- Add color to my output, https://github.com/aybabtme/rgbterm or https://github.com/alecthomas/colour
*/
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/davecgh/go-spew/spew"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func CLIInterface(dbFile string) int {
	console := bufio.NewScanner(os.Stdin)
	if dbFile == "" {
		fmt.Print("Please enter the path to the password database file to open:")
		console.Scan()
		dbFile = console.Text()
	}
	fmt.Print("Password:")
	console.Scan()
	passwd := console.Text()
	db, err := pwsafe.OpenPWSafeFile(dbFile, passwd)
	if err == nil {
		fmt.Printf("Opened file %s, enter a command or 'help' for information", dbFile)
	} else {
		log.WithFields(log.Fields{"File": dbFile, "Error": err}).Error("Error Opening file")
		return 1
	}

CLILoop:
	for {
		fmt.Print("\n> ")
		console.Scan()
		cmd := console.Text()
		switch strings.ToLower(cmd) {
		case "help", "h":
			fmt.Println("Valid commands: help, exit, list, quit, save, show")
		// Todo: Support ^d for quitting also
		case "exit", "quit", "q":
			break CLILoop
		case "list":
			for _, item := range db.List() {
				fmt.Printf("\"%v\"\n", item)
			}
		case "save":
			fmt.Println("Unimplemented")
		case "show":
			fmt.Println("\tWhich entry")
			console.Scan()
			r, prs := db.GetRecord(console.Text())
			if prs {
				spew.Dump(r)
			} else {
				fmt.Println("Record not found")
			}
		default:
			fmt.Printf("Unknown command %s, type 'help' for valid commands", cmd)
		}
	}
	return 0
}
