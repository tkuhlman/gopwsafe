// miscellaneous subordinate windows

package gui

import (
	"fmt"

	"github.com/tkuhlman/gopwsafe/pwsafe"

	"github.com/mattn/go-gtk/gtk"
)

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

func propertiesWindow(db pwsafe.DB) {
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle(db.GetName())

	v3db := (db).(*pwsafe.V3)

	name := gtk.NewLabel("DB Name")
	nameValue := gtk.NewEntry()
	nameValue.SetText(db.GetName())

	savePath := gtk.NewLabel("Save path")
	savePathValue := gtk.NewEntry()
	savePathValue.SetText(v3db.LastSavePath)

	saveButton := gtk.NewButtonWithLabel("Save")
	saveButton.Clicked(func() {
		v3db.Name = nameValue.GetText()

		err := pwsafe.WritePWSafeFile(db, "")
		if err != nil {
			errorDialog(window, fmt.Sprintf("Error Saving database to a file\n%s", err))
		}
		window.Destroy()
	})
	cancelButton := gtk.NewButtonWithLabel("Cancel")
	cancelButton.Clicked(func() {
		//todo if this is a new DB that was cancelled it will still show in the list
		window.Destroy()
	})

	//layout
	vbox := gtk.NewVBox(false, 0)
	vbox.PackStart(quitMenuBar(window), false, false, 0)

	hbox := gtk.NewHBox(true, 1)
	hbox.Add(name)
	hbox.Add(nameValue)
	vbox.PackStart(hbox, false, false, 0)

	hbox = gtk.NewHBox(true, 1)
	hbox.Add(savePath)
	hbox.Add(savePathValue)
	vbox.PackStart(hbox, false, false, 0)

	hbox = gtk.NewHBox(true, 1)
	hbox.Add(saveButton)
	hbox.Add(cancelButton)
	vbox.PackStart(hbox, false, false, 0)

	window.Add(vbox)
	window.SetSizeRequest(500, 500)
	window.ShowAll()
}
