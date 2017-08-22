package gui

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func (app *GoPWSafeGTK) openDB(path string, password string) bool {

	// TODO make sure the dbFile is not already opened and in dbs
	db, err := pwsafe.OpenPWSafeFile(path, password)
	if err != nil {
		app.errorDialog(fmt.Sprintf("Error Opening file %s\n%s", path, err))
		return false
	}
	err = app.conf.AddToPathHistory(path)
	if err != nil {
		app.errorDialog(fmt.Sprintf("Error adding %s to History\n%s", path, err))
		return false
	}
	app.dbs = append(app.dbs, db)
	app.updateRecords("")
	return true
}

// TODO if this is the only window open on close it should shutdown the app, likely just handle close by
// sending the shutdown signal
func (app *GoPWSafeGTK) openWindow(dbFile string) {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal(err)
	}
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")

	window.Connect("destroy", func() {
		window.Close()
		app.GetWindowByID(app.mainWindowID).ShowAll()
	})

	pathLabel, err := gtk.LabelNew("Password DB path: ")
	if err != nil {
		log.Fatal(err)
	}

	pathBox, err := gtk.ComboBoxTextNewWithEntry()
	if err != nil {
		log.Fatal(err)
	}
	if dbFile != "" {
		pathBox.AppendText(dbFile)
	}
	for _, entry := range app.conf.GetPathHistory() {
		pathBox.AppendText(entry)
	}
	pathBox.AppendText("Choose a file")
	pathBox.SetActive(0)
	pathBox.Connect("changed", func() {
		if pathBox.GetActiveText() == "Choose a file" {
			filechooserdialog, err := gtk.FileChooserDialogNewWith1Button(
				"Choose Password Safe file...",
				window,
				gtk.FILE_CHOOSER_ACTION_OPEN,
				"Okay",
				gtk.RESPONSE_ACCEPT)
			if err != nil {
				log.Fatal(err)
			}
			if gtk.ResponseType(filechooserdialog.Run()) == gtk.RESPONSE_ACCEPT {
				pathBox.PrependText(filechooserdialog.GetFilename())
				pathBox.SetActive(0)
				filechooserdialog.Destroy()
			}
		}
	})

	passwdLabel, err := gtk.LabelNew("Password: ")
	if err != nil {
		log.Fatal(err)
	}

	passwordBox, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	passwordBox.SetVisibility(false)
	// Pressing enter in the password box opens the db
	dbDecrypt := func() {
		text, err := passwordBox.GetText()
		if err != nil {
			log.Fatal(err)
		}
		opened := app.openDB(pathBox.GetActiveText(), text)
		if opened {
			window.Close()
			app.GetWindowByID(app.mainWindowID).ShowAll()
		}
	}
	passwordBox.Connect("activate", dbDecrypt)

	openButton, err := gtk.ButtonNewWithLabel("Open")
	if err != nil {
		log.Fatal(err)
	}

	openButton.Connect("clicked", dbDecrypt)

	//layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	if err != nil {
		log.Fatal(err)
	}
	vbox.PackStart(app.loginMenuBar(), false, false, 0)
	vbox.Add(pathLabel)
	vbox.Add(pathBox)
	vbox.Add(passwdLabel)
	vbox.Add(passwordBox)
	vbox.Add(openButton)
	window.Add(vbox)
	window.SetSizeRequest(500, 150)

	window.ShowAll()
}

func (app *GoPWSafeGTK) loginMenuBar() *gtk.MenuBar {
	mb, err := gtk.MenuBarNew()
	if err != nil {
		log.Fatal(err)
	}
	mb.Append(app.fileMenu())

	// TODO below is too much a duplicate of what is in the mainMenuBar, problems and all
	dbMenuItem, err := gtk.MenuItemNewWithLabel("DB")
	if err != nil {
		log.Fatal(err)
	}
	mb.Append(dbMenuItem)
	dbMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal(err)
	}
	dbMenuItem.SetSubmenu(dbMenu)

	newDB, err := gtk.MenuItemNewWithLabel("NewDB")
	if err != nil {
		log.Fatal(err)
	}
	newDB.Connect("activate", func() {
		db := pwsafe.NewV3("", "")
		app.propertiesWindow(db)
		app.dbs = append(app.dbs, db)
	})
	dbMenu.Append(newDB)

	return mb
}
