package gui

import (
	"fmt"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func openDB(path string, password string, dbs *[]*pwsafe.DB, parent *gtk.Window, conf config.PWSafeDBConfig, recordStore *gtk.TreeStore) {

	// todo make sure the dbFile is not already opened and in dbs
	db, err := pwsafe.OpenPWSafeFile(path, password)
	if err != nil {
		errorDialog(parent, fmt.Sprintf("Error Opening file %s\n%s", path, err))
		return
	}
	err = conf.AddToPathHistory(path)
	if err != nil {
		errorDialog(parent, fmt.Sprintf("Error adding %s to History\n%s", path, err))
	}
	newdbs := append(*dbs, &db)
	*dbs = newdbs
	updateRecords(dbs, recordStore, "")
}

func openWindow(dbFile string, dbs *[]*pwsafe.DB, conf config.PWSafeDBConfig, mainWindow *gtk.Window, recordStore *gtk.TreeStore) {
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "Open Window")

	pathLabel := gtk.NewLabel("Password DB path: ")

	pathBox := gtk.NewComboBoxTextWithEntry()
	if dbFile != "" {
		pathBox.AppendText(dbFile)
	}
	for _, entry := range conf.GetPathHistory() {
		pathBox.AppendText(entry)
	}
	pathBox.AppendText("Choose a file")
	pathBox.SetActive(0)
	pathBox.Connect("changed", func() {
		if pathBox.GetActiveText() == "Choose a file" {
			filechooserdialog := gtk.NewFileChooserDialog(
				"Choose Password Safe file...",
				window,
				gtk.FILE_CHOOSER_ACTION_OPEN,
				gtk.STOCK_OK,
				gtk.RESPONSE_ACCEPT)
			filechooserdialog.Response(func() {
				pathBox.PrependText(filechooserdialog.GetFilename())
				//todo This triggers a bug in go-gtk causing a crash
				//pathBox.SetActive(0)
				filechooserdialog.Destroy()
			})
			filechooserdialog.Run()
		}
	})

	passwdLabel := gtk.NewLabel("Password: ")

	passwordBox := gtk.NewEntry()
	passwordBox.SetVisibility(false)
	// Pressing enter in the password box opens the db
	passwordBox.Connect("activate", func() {
		openDB(pathBox.GetActiveText(), passwordBox.GetText(), dbs, window, conf, recordStore)
		window.Hide()
		mainWindow.ShowAll()
	})

	openButton := gtk.NewButtonWithLabel("Open")
	openButton.Clicked(func() {
		openDB(pathBox.GetActiveText(), passwordBox.GetText(), dbs, window, conf, recordStore)
		window.Hide()
		mainWindow.ShowAll()
	})

	//layout
	vbox := gtk.NewVBox(false, 1)
	vbox.PackStart(quitMenuBar(window), false, false, 0)
	vbox.Add(pathLabel)
	vbox.Add(pathBox)
	vbox.Add(passwdLabel)
	vbox.Add(passwordBox)
	vbox.Add(openButton)
	window.Add(vbox)
	window.SetSizeRequest(500, 300)

	window.ShowAll()
}

func quitMenuBar(window *gtk.Window) *gtk.Widget {
	actionGroup := gtk.NewActionGroup("standard")
	actionGroup.AddAction(gtk.NewAction("FileMenu", "File", "", ""))
	fileQuit := gtk.NewAction("FileQuit", "", "", gtk.STOCK_QUIT)
	fileQuit.Connect("activate", gtk.MainQuit)
	actionGroup.AddActionWithAccel(fileQuit, "<control>q")

	uiInfo := `
<ui>
  <menubar name='MenuBar'>
    <menu action='FileMenu'>
      <menuitem action='FileQuit' />
    </menu>
  </menubar>
</ui>
`
	// todo add a popup menu, I think that is a context menu
	uiManager := gtk.NewUIManager()
	uiManager.AddUIFromString(uiInfo)
	uiManager.InsertActionGroup(actionGroup, 0)
	accelGroup := uiManager.GetAccelGroup()
	window.AddAccelGroup(accelGroup)

	return uiManager.GetWidget("/MenuBar")
}
