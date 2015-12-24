package gui

import (
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
	"github.com/skratchdot/open-golang/open"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

// The default ubuntu font is okay but using something like hack would be better.
func recordWindow(db *pwsafe.DB, record *pwsafe.Record) {
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle(record.Title)

	title := gtk.NewLabel("Title")
	titleValue := gtk.NewEntry()
	titleValue.SetText(record.Title)

	group := gtk.NewLabel("Group")
	groupValue := gtk.NewEntry()
	groupValue.SetText(record.Group)

	user := gtk.NewLabel("Username")
	userValue := gtk.NewEntry()
	userValue.SetText(record.Username)

	url := gtk.NewLabel("URL")
	urlValue := gtk.NewEntry()
	urlValue.SetText(record.URL)

	password := gtk.NewLabel("Password")
	passwordValue := gtk.NewEntry()
	passwordValue.SetVisibility(false)
	passwordValue.SetText(record.Password)
	showPassword := gtk.NewButtonWithLabel("show/hide")
	showPassword.Clicked(func() {
		passwordValue.SetVisibility(!passwordValue.GetVisibility())
	})

	notesFrame := gtk.NewFrame("Notes")
	notesWin := gtk.NewScrolledWindow(nil, nil)
	notesWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	notesWin.SetShadowType(gtk.SHADOW_IN)
	textView := gtk.NewTextView()
	buffer := textView.GetBuffer()
	buffer.SetText(record.Notes)
	notesWin.Add(textView)
	notesFrame.Add(notesWin)

	okayButton := gtk.NewButtonWithLabel("Okay")
	okayButton.Clicked(func() {
		// Grab values
		origName := record.Title
		record.Title = titleValue.GetText()
		record.Group = groupValue.GetText()
		record.Username = userValue.GetText()
		record.URL = urlValue.GetText()
		record.Password = passwordValue.GetText()
		var start, end gtk.TextIter
		buffer.GetStartIter(&start)
		buffer.GetEndIter(&end)
		record.Notes = buffer.GetText(&start, &end, true)

		// Update the record
		if origName != record.Title { // The Record title has changed
			(*db).DeleteRecord(origName)
			// Todo, this invalidates the current search so a new run of updateRecords should be done, or at least it made obvious what is going on
		}
		//todo detect if there have been changes and only update if needed
		(*db).SetRecord(*record)
		window.Destroy()
	})
	cancelButton := gtk.NewButtonWithLabel("Cancel")
	cancelButton.Clicked(func() {
		window.Destroy()
	})

	//layout
	vbox := gtk.NewVBox(false, 0)
	vbox.PackStart(quitMenuBar(window), false, false, 0)
	vbox.PackStart(recordMenuBar(window, record), false, false, 0)

	hbox := gtk.NewHBox(true, 1)
	hbox.Add(title)
	hbox.Add(titleValue)
	vbox.PackStart(hbox, false, false, 0)
	hbox = gtk.NewHBox(true, 1)
	hbox.Add(group)
	hbox.Add(groupValue)
	vbox.PackStart(hbox, false, false, 0)
	hbox = gtk.NewHBox(true, 1)
	hbox.Add(user)
	hbox.Add(userValue)
	vbox.PackStart(hbox, false, false, 0)
	hbox = gtk.NewHBox(true, 1)
	hbox.Add(url)
	hbox.Add(urlValue)
	vbox.PackStart(hbox, false, false, 0)
	hbox = gtk.NewHBox(true, 1)
	hbox.Add(password)
	hbox.Add(passwordValue)
	hbox.Add(showPassword)
	vbox.PackStart(hbox, false, false, 0)

	vbox.Add(notesFrame)
	hbox = gtk.NewHBox(true, 1)
	hbox.Add(okayButton)
	hbox.Add(cancelButton)
	vbox.PackStart(hbox, false, false, 0)

	window.Add(vbox)
	window.SetSizeRequest(500, 500)
	window.ShowAll()
}

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
