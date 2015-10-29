package gui

import (
	"fmt"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func openWindow(dbFile string) {
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "Open Window")

	conf := config.Load()

	vbox := gtk.NewVBox(false, 1)
	menubar := gtk.NewMenuBar()
	vbox.PackStart(menubar, false, false, 0)

	pathLabel := gtk.NewLabel("Password DB path: ")
	vbox.Add(pathLabel)

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
	vbox.Add(pathBox)

	passwdLabel := gtk.NewLabel("Password: ")
	vbox.Add(passwdLabel)

	passwordBox := gtk.NewEntry()
	passwordBox.SetVisibility(false)
	vbox.Add(passwordBox)

	openButton := gtk.NewButtonWithLabel("Open")
	openButton.Clicked(func() {
		openDB(window, conf, pathBox.GetActiveText(), passwordBox.GetText())
	})
	vbox.Add(openButton)

	window.Add(vbox)
	window.SetSizeRequest(500, 300)
	//todo DB should open after hitting enter (rather than tab-enter)
	// the way to do this is probably setting a default widget but none of my gui elements are a widget, perhaps I need a builder
	// and builder.GetObject() and gtk.WidgetFromObject() in conjunction?
	//window.SetDefault(openButton)
	window.ShowAll()
}

func openDB(previousWindow *gtk.Window, conf config.PWSafeDBConfig, dbFile string, passwd string) {
	db, err := pwsafe.OpenPWSafeFile(dbFile, passwd)
	if err != nil {
		errorDialog(previousWindow, fmt.Sprintf("Error Opening file %s\n%s", dbFile, err))
		return
	}
	err = conf.AddToPathHistory(dbFile)
	if err != nil {
		errorDialog(previousWindow, fmt.Sprintf("Error adding %s to History\n%s", dbFile, err))
	}
	previousWindow.Hide()
	mainWindow(db, conf)
}

func errorDialog(parent *gtk.Window, msg string) {
	messagedialog := gtk.NewMessageDialog(
		parent,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_OK,
		msg)
	messagedialog.Response(func() {
		messagedialog.Destroy()
	})
	messagedialog.Run()
}
