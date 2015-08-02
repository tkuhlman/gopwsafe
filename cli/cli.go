/*

* update the Readme with info on what I decide to use.
Options include:
	- termbox go - https://github.com/nsf/termbox-go - this could lead to a nice cli interface but may be overkill.
	- just using bufio.ReadLine(),
		- Add color to my output, https://github.com/aybabtme/rgbterm or https://github.com/alecthomas/colour
*/
package cli

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/Bowery/prompt"
	"github.com/davecgh/go-spew/spew"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func CLIInterface(dbFile string) int {
	if dbFile == "" {
		dbFile, _ = prompt.Basic("Please enter the path to the password database file to open:", true)
	}
	passwd, _ := prompt.Password("Password:")
	db, err := pwsafe.OpenPWSafeFile(dbFile, passwd)
	if err == nil {
		fmt.Printf("Opened file %s, enter a command or 'help' for information", dbFile)
	} else {
		log.WithFields(log.Fields{"File": dbFile, "Error": err}).Error("Error Opening file")
		return 1
	}

	term, _ := prompt.NewTerminal()
	defer term.Close()
CLILoop:
	for {
		cmd, err := term.Basic("> ", false)
		if err == prompt.ErrEOF || err == prompt.ErrCTRLC {
			cmd = "exit"
		}
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
			entry, _ := prompt.Basic("Which entry: ", true)
			r, prs := db.GetRecord(entry)
			if prs {
				// todo - the problem of closing the term is I loose history, but without it spew is mangled.
				term.Close()
				spew.Dump(r)
				term, _ = prompt.NewTerminal()
				defer term.Close()
			} else {
				fmt.Println("Record not found")
			}
		default:
			fmt.Println("Unknown command %s, type 'help' for valid commands", cmd)
		}
	}
	return 0
}
