package gui

import (
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func (app *GoPWSafeGTK) recordWindow(db pwsafe.DB, record *pwsafe.Record) {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	logError(err, "")
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle(record.Title)
	window.AddAccelGroup(app.accelGroup)

	title, err := gtk.LabelNew("Title")
	logError(err, "")
	titleValue, err := gtk.EntryNew()
	logError(err, "")
	titleValue.SetText(record.Title)
	titleValue.SetHExpand(true)

	group, err := gtk.LabelNew("Group")
	logError(err, "")
	groupValue, err := gtk.EntryNew()
	logError(err, "")
	groupValue.SetText(record.Group)
	groupValue.SetHExpand(true)

	user, err := gtk.LabelNew("Username")
	logError(err, "")
	userValue, err := gtk.EntryNew()
	logError(err, "")
	userValue.SetText(record.Username)
	userValue.SetHExpand(true)

	url, err := gtk.LabelNew("URL")
	logError(err, "")
	urlValue, err := gtk.EntryNew()
	logError(err, "")
	urlValue.SetText(record.URL)
	urlValue.SetHExpand(true)

	password, err := gtk.LabelNew("Password")
	logError(err, "")
	passwordValue, err := gtk.EntryNew()
	logError(err, "")
	passwordValue.SetVisibility(false)
	passwordValue.SetText(record.Password)
	passwordValue.SetHExpand(true)
	showPassword, err := gtk.ButtonNewWithLabel("show/hide")
	logError(err, "")
	showPassword.Connect("clicked", func() {
		passwordValue.SetVisibility(!passwordValue.GetVisibility())
	})

	modTime, err := gtk.LabelNew("Last Modification")
	logError(err, "")
	modValue, err := gtk.LabelNew(record.ModTime.Format(time.UnixDate))
	logError(err, "")
	modValue.SetHExpand(true)

	notesFrame, err := gtk.FrameNew("Notes")
	logError(err, "")
	notesWin, err := gtk.ScrolledWindowNew(nil, nil)
	logError(err, "")
	notesWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	textView, err := gtk.TextViewNew()
	logError(err, "")
	textView.SetWrapMode(gtk.WRAP_WORD)
	buffer, err := textView.GetBuffer()
	logError(err, "")
	buffer.SetText(record.Notes)
	notesWin.Add(textView)
	notesFrame.Add(notesWin)

	okayButton, err := gtk.ButtonNewWithLabel("Okay")
	logError(err, "")
	okayButton.Connect("clicked", func() {
		// Grab values
		origName := record.Title
		record.Title, err = titleValue.GetText()
		logError(err, "")
		record.Group, err = groupValue.GetText()
		logError(err, "")
		record.Username, err = userValue.GetText()
		logError(err, "")
		record.URL, err = urlValue.GetText()
		logError(err, "")
		record.Password, err = passwordValue.GetText()
		logError(err, "")
		start := buffer.GetStartIter()
		end := buffer.GetEndIter()
		record.Notes, err = buffer.GetText(start, end, true)
		logError(err, "")

		// Update the record
		if origName != record.Title { // The Record title has changed
			db.DeleteRecord(origName)
			app.updateRecords("")
		}
		db.SetRecord(*record)
		window.Destroy()
	})
	cancelButton, err := gtk.ButtonNewWithLabel("Cancel")
	logError(err, "")
	cancelButton.Connect("clicked", func() {
		window.Destroy()
	})

	//layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	logError(err, "")
	vbox.PackStart(app.recordMenuBar(window, record), false, false, 0)

	grid, err := gtk.GridNew()
	logError(err, "")
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
	logError(err, "")
	hbox.Add(okayButton)
	hbox.Add(cancelButton)
	vbox.PackStart(hbox, false, false, 0)

	window.Add(vbox)
	window.SetDefaultSize(500, 500)
	window.ShowAll()
}

// Configures the record menubar and keyboard shortcuts
func (app *GoPWSafeGTK) recordMenuBar(parent *gtk.Window, record *pwsafe.Record) *gtk.MenuBar {
	mb, err := gtk.MenuBarNew()
	logError(err, "")

	fileMenuItem, err := gtk.MenuItemNewWithLabel("File")
	logError(err, "")
	fileMenu, err := gtk.MenuNew()
	logError(err, "")
	fileMenuItem.SetSubmenu(fileMenu)

	ag, err := gtk.AccelGroupNew()
	logError(err, "")
	fileMenu.SetAccelGroup(ag)
	parent.AddAccelGroup(ag)

	close, err := gtk.MenuItemNewWithLabel("Close")
	logError(err, "")
	close.Connect("activate", parent.Destroy)
	close.AddAccelerator("activate", ag, 'w', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)
	fileMenu.Append(close)

	mb.Append(fileMenuItem)
	mb.Append(app.recordMenu(parent, record))

	return mb
}
