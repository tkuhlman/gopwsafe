// miscellaneous subordinate windows

package gui

import (
	"fmt"
	"time"

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

	saveTime := gtk.NewLabel(fmt.Sprintf("Last Save at %v", v3db.LastSave.Format(time.RFC3339)))

	passwordLabel := gtk.NewLabel("New Password")
	passwordValue := gtk.NewEntry()
	passwordValue.SetVisibility(false)

	password2Label := gtk.NewLabel("Repeated New Password")
	password2Value := gtk.NewEntry()
	password2Value.SetVisibility(false)

	descriptionFrame := gtk.NewFrame("Description")
	descriptionWin := gtk.NewScrolledWindow(nil, nil)
	descriptionWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	descriptionWin.SetShadowType(gtk.SHADOW_IN)
	textView := gtk.NewTextView()
	buffer := textView.GetBuffer()
	buffer.SetText(v3db.Description)
	descriptionWin.Add(textView)
	descriptionFrame.Add(descriptionWin)

	saveButton := gtk.NewButtonWithLabel("Save")
	// todo it would be nice if pressing enter in an field activated this.
	saveButton.Clicked(func() {
		v3db.Name = nameValue.GetText()

		var start, end gtk.TextIter
		buffer.GetStartIter(&start)
		buffer.GetEndIter(&end)
		v3db.Description = buffer.GetText(&start, &end, true)

		pw := passwordValue.GetText()
		if pw != "" {
			pw2 := password2Value.GetText()
			if pw != pw2 {
				errorDialog(window, "Error Passwords don't match")
			} else if err := db.SetPassword(pw); err != nil {
				errorDialog(window, fmt.Sprintf("Error Updating password\n%s", err))
			}

		}

		err := pwsafe.WritePWSafeFile(db, savePathValue.GetText())
		if err != nil {
			errorDialog(window, fmt.Sprintf("Error Saving database to a file\n%s", err))
		}
		window.Destroy()
		gtk.MainQuit()
	})
	cancelButton := gtk.NewButtonWithLabel("Cancel")
	cancelButton.Clicked(func() {
		//todo if this is a new DB that was cancelled it will still show in the list
		window.Destroy()
		gtk.MainQuit()
	})

	window.Connect("destroy", gtk.MainQuit)

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
	hbox.Add(saveTime)
	vbox.PackStart(hbox, false, false, 0)

	hbox = gtk.NewHBox(true, 1)
	hbox.Add(passwordLabel)
	hbox.Add(passwordValue)
	vbox.PackStart(hbox, false, false, 0)

	hbox = gtk.NewHBox(true, 1)
	hbox.Add(password2Label)
	hbox.Add(password2Value)
	vbox.PackStart(hbox, false, false, 0)

	vbox.Add(descriptionFrame)

	hbox = gtk.NewHBox(true, 1)
	hbox.Add(saveButton)
	hbox.Add(cancelButton)
	vbox.PackStart(hbox, false, false, 0)

	window.Add(vbox)
	window.SetSizeRequest(500, 200)
	window.ShowAll()
	gtk.Main()
}
