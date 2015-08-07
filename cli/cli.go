package cli

import (
	"fmt"
	"regexp"
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
		cmdStr, err := term.Basic("> ", false)
		cmd := strings.SplitN(cmdStr, " ", 2)
		if err == prompt.ErrEOF || err == prompt.ErrCTRLC {
			cmd = []string{"exit"}
		}
		switch strings.ToLower(cmd[0]) {
		case "help", "h":
			fmt.Println("Valid commands: help, exit, find, groups, list, listgroup, quit, save, show")
		// Todo: Support ^d for quitting also
		case "exit", "quit", "q":
			break CLILoop
		case "find":
			if len(cmd) == 1 {
				fmt.Println("Please specify a search string")
				continue
			}
			records := db.List()
			// Todo - figure out how to default case insensitive without hindering more complicated regexes
			search, err := regexp.Compile("(?i)" + cmd[1])
			if err != nil {
				fmt.Println("Invalid regexp" + err.Error())
				continue
			}
			for _, record := range records {
				if search.MatchString(record) {
					fmt.Printf("\"%v\"\n\r", record)
				}
			}
		case "groups":
			for _, item := range db.Groups() {
				fmt.Printf("\"%v\"\n\r", item)
			}
		case "list":
			for _, item := range db.List() {
				fmt.Printf("\"%v\"\n\r", item)
			}
		case "listgroup":
			group, _ := prompt.Basic("Which group: ", false)
			for _, item := range db.ListByGroup(group) {
				fmt.Printf("\"%v\"\n\r", item)
			}
		case "save":
			fmt.Println("Unimplemented")
		case "show":
			if len(cmd) == 1 {
				fmt.Println("Please specify a record title to show")
				continue
			}
			// todo - I should handle quotes around a name
			r, prs := db.GetRecord(cmd[1])
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
			h := fmt.Sprintf("Unknown command %s, type 'help' for valid commands", cmd)
			fmt.Println(h)
		}
	}
	return 0
}
