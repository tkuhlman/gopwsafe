package gui

import (
	"fmt"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

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
	passwordBox.SetActivatesDefault(true)

	openButton := gtk.NewButtonWithLabel("Open")
	openButton.Clicked(func() {
		toOpen := pathBox.GetActiveText()
		// todo make sure the dbFile is not already opened and in dbs
		db, err := pwsafe.OpenPWSafeFile(toOpen, passwordBox.GetText())
		if err != nil {
			errorDialog(window, fmt.Sprintf("Error Opening file %s\n%s", toOpen, err))
			return
		}
		err = conf.AddToPathHistory(toOpen)
		if err != nil {
			errorDialog(window, fmt.Sprintf("Error adding %s to History\n%s", toOpen, err))
		}
		newdbs := append(*dbs, &db)
		*dbs = newdbs
		updateRecords(dbs, recordStore, "")
		window.Hide()
		mainWindow.ShowAll()
	})

	// I want enter in the passwordBox to work for opening the db but am unsure how to do it.
	//window.SetDefault(openButton)

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
