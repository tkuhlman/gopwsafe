package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func (app *GoPWSafeGTK) openDB(path string, password string) bool {

	for _, db := range app.dbs {
		v3db, ok := db.(*pwsafe.V3)
		if !ok {
			continue
		}
		if path == v3db.LastSavePath {
			app.errorDialog(fmt.Sprintf("A password database at path %q is already open", path))
			return false
		}
	}
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

func (app *GoPWSafeGTK) openWindow(dbFile string) {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	logError(err, "")
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.AddAccelGroup(app.accelGroup)

	window.Connect("destroy", func() {
		window.Close()
		app.GetWindowByID(app.mainWindowID).ShowAll()
	})

	pathLabel, err := gtk.LabelNew("Password DB path: ")
	logError(err, "")

	pathBox, err := gtk.ComboBoxTextNewWithEntry()
	logError(err, "")
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
			logError(err, "")
			if gtk.ResponseType(filechooserdialog.Run()) == gtk.RESPONSE_ACCEPT {
				pathBox.PrependText(filechooserdialog.GetFilename())
				pathBox.SetActive(0)
				filechooserdialog.Destroy()
			}
		}
	})

	passwdLabel, err := gtk.LabelNew("Password: ")
	logError(err, "")

	passwordBox, err := gtk.EntryNew()
	logError(err, "")
	passwordBox.SetVisibility(false)
	// Pressing enter in the password box opens the db
	dbDecrypt := func() {
		text, err := passwordBox.GetText()
		logError(err, "")
		app.openDB(pathBox.GetActiveText(), text)
		window.Close()
		app.GetWindowByID(app.mainWindowID).ShowAll()
	}
	passwordBox.Connect("activate", dbDecrypt)

	openButton, err := gtk.ButtonNewWithLabel("Open")
	logError(err, "")

	openButton.Connect("clicked", dbDecrypt)

	//layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	logError(err, "")
	vbox.PackStart(app.openWindowMenuBar(), false, false, 0)
	vbox.Add(pathLabel)
	vbox.Add(pathBox)
	vbox.Add(passwdLabel)
	vbox.Add(passwordBox)
	vbox.Add(openButton)
	window.Add(vbox)
	window.SetSizeRequest(500, 150)

	window.ShowAll()
}

func (app *GoPWSafeGTK) openWindowMenuBar() *gtk.MenuBar {
	mb, err := gtk.MenuBarNew()
	logError(err, "")
	mb.Append(app.fileMenu())

	dbMenuItem, err := gtk.MenuItemNewWithLabel("DB")
	logError(err, "")
	mb.Append(dbMenuItem)
	dbMenu, err := gtk.MenuNew()
	logError(err, "")
	dbMenuItem.SetSubmenu(dbMenu)

	newDB, err := gtk.MenuItemNewWithLabel("NewDB")
	logError(err, "")
	newDB.Connect("activate", func() {
		db := pwsafe.NewV3("", "")
		app.propertiesWindow(db)
		app.dbs = append(app.dbs, db)
	})
	dbMenu.Append(newDB)

	return mb
}
