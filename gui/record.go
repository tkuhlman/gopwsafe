package gui

import (
	"log"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func (app *GoPWSafeGTK) recordWindow(db pwsafe.DB, record *pwsafe.Record) {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal(err)
	}
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle(record.Title)

	title, err := gtk.LabelNew("Title")
	if err != nil {
		log.Fatal(err)
	}
	titleValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	titleValue.SetText(record.Title)
	titleValue.SetHExpand(true)

	group, err := gtk.LabelNew("Group")
	if err != nil {
		log.Fatal(err)
	}
	groupValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	groupValue.SetText(record.Group)
	groupValue.SetHExpand(true)

	user, err := gtk.LabelNew("Username")
	if err != nil {
		log.Fatal(err)
	}
	userValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	userValue.SetText(record.Username)
	userValue.SetHExpand(true)

	url, err := gtk.LabelNew("URL")
	if err != nil {
		log.Fatal(err)
	}
	urlValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	urlValue.SetText(record.URL)
	urlValue.SetHExpand(true)

	password, err := gtk.LabelNew("Password")
	if err != nil {
		log.Fatal(err)
	}
	passwordValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	passwordValue.SetVisibility(false)
	passwordValue.SetText(record.Password)
	passwordValue.SetHExpand(true)
	showPassword, err := gtk.ButtonNewWithLabel("show/hide")
	if err != nil {
		log.Fatal(err)
	}
	showPassword.Connect("clicked", func() {
		passwordValue.SetVisibility(!passwordValue.GetVisibility())
	})

	modTime, err := gtk.LabelNew("Last Modification")
	if err != nil {
		log.Fatal(err)
	}
	modValue, err := gtk.LabelNew(record.ModTime.Format(time.UnixDate))
	if err != nil {
		log.Fatal(err)
	}
	modValue.SetHExpand(true)

	notesFrame, err := gtk.FrameNew("Notes")
	if err != nil {
		log.Fatal(err)
	}
	notesWin, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	notesWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	textView, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal(err)
	}
	textView.SetWrapMode(gtk.WRAP_WORD)
	buffer, err := textView.GetBuffer()
	if err != nil {
		log.Fatal(err)
	}
	buffer.SetText(record.Notes)
	notesWin.Add(textView)
	notesFrame.Add(notesWin)

	okayButton, err := gtk.ButtonNewWithLabel("Okay")
	if err != nil {
		log.Fatal(err)
	}
	okayButton.Connect("clicked", func() {
		// Grab values
		origName := record.Title
		record.Title, err = titleValue.GetText()
		if err != nil {
			log.Fatal(err)
		}
		record.Group, err = groupValue.GetText()
		if err != nil {
			log.Fatal(err)
		}
		record.Username, err = userValue.GetText()
		if err != nil {
			log.Fatal(err)
		}
		record.URL, err = urlValue.GetText()
		if err != nil {
			log.Fatal(err)
		}
		record.Password, err = passwordValue.GetText()
		if err != nil {
			log.Fatal(err)
		}
		start := buffer.GetStartIter()
		end := buffer.GetEndIter()
		record.Notes, err = buffer.GetText(start, end, true)
		if err != nil {
			log.Fatal(err)
		}

		// Update the record
		if origName != record.Title { // The Record title has changed
			db.DeleteRecord(origName)
			app.updateRecords("")
		}
		db.SetRecord(*record)
		window.Destroy()
	})
	cancelButton, err := gtk.ButtonNewWithLabel("Cancel")
	if err != nil {
		log.Fatal(err)
	}
	cancelButton.Connect("clicked", func() {
		window.Destroy()
	})

	//layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	if err != nil {
		log.Fatal(err)
	}
	// TODO figure out how to add record specific menu bar including ctrl-w to close this window and
	// all the mainMenuBar record items but targetted at the window record not the selection
	//	vbox.PackStart(recordMenuBar(window, record), false, false, 0)

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal(err)
	}
	vbox.PackStart(grid, false, true, 1)
	grid.SetColumnSpacing(2)

	grid.Attach(title, 0, 0, 1, 1)
	grid.Attach(titleValue, 1, 0, 2, 1)

	grid.Attach(group, 0, 1, 1, 1)
	grid.Attach(groupValue, 1, 1, 2, 1)

	grid.Attach(user, 0, 2, 1, 1)
	grid.Attach(userValue, 1, 2, 2, 1)

	grid.Attach(url, 0, 3, 1, 1)
	grid.Attach(urlValue, 1, 3, 2, 1)

	grid.Attach(password, 0, 4, 1, 1)
	grid.Attach(passwordValue, 1, 4, 1, 1)
	grid.Attach(showPassword, 2, 4, 1, 1)

	grid.Attach(modTime, 0, 5, 1, 1)
	grid.Attach(modValue, 1, 5, 2, 1)

	vbox.PackStart(notesFrame, true, true, 0)
	hbox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	if err != nil {
		log.Fatal(err)
	}
	hbox.Add(okayButton)
	hbox.Add(cancelButton)
	vbox.PackStart(hbox, false, false, 0)

	window.Add(vbox)
	window.SetDefaultSize(500, 500)
	window.ShowAll()
}

/*
// Configures the record menubar and keyboard shortcuts
func recordMenuBar(window *gtk.Window, record *pwsafe.Record) *gtk.Widget {
	clipboard := gtk.NewClipboardGetForDisplay(gdk.DisplayGetDefault(), gdk.SELECTION_CLIPBOARD)

	actionGroup := gtk.NewActionGroup("record")
	actionGroup.AddAction(gtk.NewAction("RecordMenu", "Record", "", ""))

	copyUser := gtk.NewAction("CopyUsername", "Copy username to clipboard", "", "")
	copyUser.Connect("activate", func() { clipboard.SetText(record.Username) })
	actionGroup.AddActionWithAccel(copyUser, "<control>u")

	copyPassword := gtk.NewAction("CopyPassword", "Copy password to clipboard", "", "")
	copyPassword.Connect("activate", func() { clipboard.SetText(record.Password) })
	actionGroup.AddActionWithAccel(copyPassword, "<control>p")

	openURL := gtk.NewAction("OpenURL", "Open URL", "", "")
	// gtk-go hasn't yet implemented gtk_show_uri so using github.com/skratchdot/open-golang/open
	// todo it opens the url but should switch to that app also.
	openURL.Connect("activate", func() { open.Start(record.URL) })
	actionGroup.AddActionWithAccel(openURL, "<control>o")

	copyURL := gtk.NewAction("CopyURL", "Copy URL to clipboard", "", "")
	copyURL.Connect("activate", func() { clipboard.SetText(record.URL) })
	actionGroup.AddActionWithAccel(copyURL, "<control>l")

	closeWindow := gtk.NewAction("CloseWindow", "", "", gtk.STOCK_CLOSE)
	closeWindow.Connect("activate", window.Destroy)
	actionGroup.AddActionWithAccel(closeWindow, "<control>w")

	uiInfo := `
<ui>
  <menubar name='MenuBar'>
    <menu action='RecordMenu'>
      <menuitem action='CopyUsername' />
      <menuitem action='CopyPassword' />
      <menuitem action='OpenURL' />
      <menuitem action='CopyURL' />
      <menuitem action='CloseWindow' />
    </menu>
  </menubar>
</ui>
`
	// todo add a popup menu, at least I think that is a context menu
	uiManager := gtk.NewUIManager()
	uiManager.AddUIFromString(uiInfo)
	uiManager.InsertActionGroup(actionGroup, 0)
	accelGroup := uiManager.GetAccelGroup()
	window.AddAccelGroup(accelGroup)

	return uiManager.GetWidget("/MenuBar")
}
*/
