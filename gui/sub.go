package gui

import (
	"fmt"
	"log"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

func (app *GoPWSafeGTK) errorDialog(msg string) {
	parent := app.GetWindowByID(app.mainWindowID)

	messagedialog := gtk.MessageDialogNew(
		parent,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_CLOSE,
		"%s",
		msg)
	messagedialog.Response(gtk.RESPONSE_CLOSE)
	messagedialog.Run()
	messagedialog.Destroy()
}

func (app *GoPWSafeGTK) propertiesWindow(db pwsafe.DB) {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal(err)
	}
	window.SetPosition(gtk.WIN_POS_CENTER)
	dbName := db.GetName()
	window.SetTitle(dbName)

	v3db, ok := db.(*pwsafe.V3)
	if !ok {
		log.Fatalf("Failed to cast Password DB %q as a V3 password safe", dbName)
	}

	name, err := gtk.LabelNew("DB Name")
	if err != nil {
		log.Fatal(err)
	}
	nameValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	nameValue.SetText(dbName)
	nameValue.SetHExpand(true)

	savePath, err := gtk.LabelNew("Save path")
	if err != nil {
		log.Fatal(err)
	}
	savePathValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	savePathValue.SetText(v3db.LastSavePath)
	savePathValue.SetHExpand(true)

	saveTime, err := gtk.LabelNew(fmt.Sprintf("Last Save at %v", v3db.LastSave.Format(time.RFC3339)))
	if err != nil {
		log.Fatal(err)
	}

	passwordLabel, err := gtk.LabelNew("New Password")
	if err != nil {
		log.Fatal(err)
	}
	passwordValue, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	passwordValue.SetVisibility(false)
	passwordValue.SetHExpand(true)

	password2Label, err := gtk.LabelNew("Repeated New Password")
	if err != nil {
		log.Fatal(err)
	}
	password2Value, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	password2Value.SetVisibility(false)
	password2Value.SetHExpand(true)

	descriptionFrame, err := gtk.FrameNew("Description")
	if err != nil {
		log.Fatal(err)
	}
	descriptionWin, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	descriptionWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	textView, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal(err)
	}
	textView.SetWrapMode(gtk.WRAP_WORD)
	buffer, err := textView.GetBuffer()
	if err != nil {
		log.Fatal(err)
	}
	buffer.SetText(v3db.Description)
	descriptionWin.Add(textView)
	descriptionFrame.Add(descriptionWin)

	saveButton, err := gtk.ButtonNewWithLabel("Save")
	if err != nil {
		log.Fatal(err)
	}
	saveButton.Connect("clicked", func() {
		v3db.Name, err = nameValue.GetText()
		if err != nil {
			log.Fatal(err)
		}

		start := buffer.GetStartIter()
		end := buffer.GetEndIter()
		v3db.Description, err = buffer.GetText(start, end, true)
		if err != nil {
			log.Fatal(err)
		}

		pw, err := passwordValue.GetText()
		if err != nil {
			log.Fatal(err)
		}
		if pw != "" {
			pw2, err := password2Value.GetText()
			if err != nil {
				log.Fatal(err)
			}
			if pw != pw2 {
				app.errorDialog("Error Passwords don't match")
			} else if err := db.SetPassword(pw); err != nil {
				app.errorDialog(fmt.Sprintf("Error Updating password\n%s", err))
			}

		}

		path, err := savePathValue.GetText()
		if err != nil {
			log.Fatal(err)
		}

		var new bool
		v3db := db.(*pwsafe.V3)
		if v3db.LastSavePath == "" {
			new = true
		}
		if err := pwsafe.WritePWSafeFile(db, path); err != nil {
			app.errorDialog(fmt.Sprintf("Error Saving database to a file\n%s", err))
		} else if new {
			app.dbs = append(app.dbs, db)
			app.updateRecords("")
		}

		window.Destroy()
	})
	cancelButton, err := gtk.ButtonNewWithLabel("Cancel")
	if err != nil {
		log.Fatal(err)
	}
	cancelButton.Connect("clicked", func() {
		window.Destroy()
	})

	window.Connect("destroy", window.Close)

	//layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	if err != nil {
		log.Fatal(err)
	}

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal(err)
	}
	vbox.PackStart(grid, false, true, 1)
	grid.SetColumnSpacing(2)

	hbox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	if err != nil {
		log.Fatal(err)
	}
	grid.Attach(name, 0, 0, 1, 1)
	grid.Attach(nameValue, 1, 0, 1, 1)

	grid.Attach(savePath, 0, 1, 1, 1)
	grid.Attach(savePathValue, 1, 1, 1, 1)

	grid.Attach(saveTime, 0, 2, 2, 1)

	grid.Attach(passwordLabel, 0, 3, 1, 1)
	grid.Attach(passwordValue, 1, 3, 1, 1)

	grid.Attach(password2Label, 0, 4, 1, 1)
	grid.Attach(password2Value, 1, 4, 1, 1)

	vbox.PackStart(descriptionFrame, true, true, 0)

	hbox, err = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	if err != nil {
		log.Fatal(err)
	}
	hbox.Add(saveButton)
	hbox.Add(cancelButton)
	vbox.PackStart(hbox, false, false, 0)

	window.Add(vbox)
	window.SetDefaultSize(500, 400)
	window.ShowAll()
}
