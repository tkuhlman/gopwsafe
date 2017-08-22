package gui

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// GUI docs
// http://godoc.org/github.com/gotk3/gotk3
// https://developer.gnome.org/gtk3/stable/

// TODO I should try to define my Application and the associated windows, actions and menus in a
// more coordinated way, see https://developer.gnome.org/gtk3/stable/ch01s04.html and
// http://python-gtk-3-tutorial.readthedocs.io/en/latest/application.html and
// https://wiki.gnome.org/HowDoI/GtkApplication

// TODO godocs
// TODO work on the naming
type GoPWSafeGTK struct {
	*gtk.Application
	conf config.PWSafeDBConfig
	dbs  []pwsafe.DB
	// TODO having the recordStore/recordTree and window ids in here doesn't seem to be the best structure
	recordStore  *gtk.TreeStore
	recordTree   *gtk.TreeView
	mainWindowID uint
}

func NewGoPWSafeGTK() (*GoPWSafeGTK, error) {
	gtkApp, err := gtk.ApplicationNew("tkuhlman.gopwsafe", glib.APPLICATION_HANDLES_OPEN|glib.APPLICATION_NON_UNIQUE)
	if err != nil {
		return nil, err
	}

	app := &GoPWSafeGTK{
		Application: gtkApp,
		conf:        config.Load(),
		dbs:         make([]pwsafe.DB, 0),
	}

	if _, err := app.Connect("startup", app.startUp, nil); err != nil {
		return nil, fmt.Errorf("error connecting startup signal handler:%v", err)
	}
	if _, err := app.Connect("activate", app.open, nil); err != nil {
		return nil, fmt.Errorf("error connecting activate signal handler:%v", err)
	}
	if _, err := app.Connect("open", app.open, nil); err != nil {
		return nil, fmt.Errorf("error connecting open signal handler:%v", err)
	}
	if _, err := app.Connect("shutdown", app.shutdown, nil); err != nil {
		return nil, fmt.Errorf("error connecting shutdown signal handler:%v", err)
	}
	return app, nil
}

func (app *GoPWSafeGTK) Open(initialDB string) int {
	return app.Run([]string{initialDB})
}

//startUp handles the startUp signal for the GTK application
func (app *GoPWSafeGTK) startUp(gtkApp *gtk.Application) {
	app.AddWindow(app.mainWindow(""))
	// TODO should I add all the windows here or leave them all flowing from the main window?
	// app.setActions()
	// TODO add a popup menu, I think that is a context menu
}

func (app *GoPWSafeGTK) open(gtkApp *gtk.Application) {
	// TODO where does the passed in initial db path come into play?
	//app.openWindow(dbFile)
	app.openWindow("")
}

//TODO doc
func (app *GoPWSafeGTK) setActions() {
	// TODO implment actiongroups and accelgroups
	// https://developer.gnome.org/gtk3/stable/gtk3-Keyboard-Accelerators.html
	// https://developer.gnome.org/gtk3/stable/GtkActionGroup.html
	// ** What I believe is I define acctions via the .Connect method, attaching to whatever is appropriate
	// or possibly to an actionGroup. Actiongroups are attached to objects via the top level gtk methods
	// Once actions are defined accelerators are defined that map keyboard shortcuts to these actions.
	// acelerators can also be put into groups to be more easily attached to multiple windows or
	// individually assigned. Starting off I should just put accelerators onto the main menubar then start
	// to migrate the actions and accels into groups defined here or in a similar function

}

//TODO doc
func (app *GoPWSafeGTK) shutdown(gtkApp *gtk.Application) {
	app.Quit()
}

// TODO dbFile should not be here, also add godoc
func (app *GoPWSafeGTK) mainWindow(dbFile string) *gtk.Window {

	//TODO revisit the structure of the gui code, splitting more out into functions and in general better organizing things.

	window, err := gtk.ApplicationWindowNew(app.Application)
	// TODO gotk3 handles errors different update my code accordingly
	if err != nil {
		log.Fatalf("Failed to create main window: %v", err)
	}
	app.mainWindowID = window.GetID()
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.Connect("destroy", func() {
		// Check if any dbs need to be saved
		for _, db := range app.dbs {
			if db.NeedsSave() {
				app.propertiesWindow(db)
				app.errorDialog(fmt.Sprintf("Unsaved changes for db %v", db.GetName()))
				// TODO it seems that right now cancelling the error dialog kills all the windows
			}
		}
	})

	recordFrame, err := gtk.FrameNew("Records")
	if err != nil {
		log.Fatalf("Failed to create record frame: %v", err)
	}
	recordWin, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatalf("Failed to create scrolled window: %v", err)
	}
	recordWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	recordFrame.SetShadowType(gtk.SHADOW_IN)
	recordFrame.Add(recordWin)

	app.recordTree, err = gtk.TreeViewNew()
	if err != nil {
		log.Fatalf("Failed to create tree view: %v", err)
	}
	recordWin.Add(app.recordTree)

	cellPB, err := gtk.CellRendererPixbufNew()
	if err != nil {
		log.Fatal(err)
	}
	// TODO verify the icons are showing up correctly
	col1, err := gtk.TreeViewColumnNewWithAttribute("", cellPB, "pixbuf", 0)
	if err != nil {
		log.Fatal(err)
	}
	app.recordTree.AppendColumn(col1)
	cellText, err := gtk.CellRendererTextNew()
	if err != nil {
		log.Fatal(err)
	}
	col2, err := gtk.TreeViewColumnNewWithAttribute("Name", cellText, "text", 1)
	if err != nil {
		log.Fatal(err)
	}
	app.recordTree.AppendColumn(col2)

	app.recordStore, err = gtk.TreeStoreNew(glib.TYPE_OBJECT, glib.TYPE_STRING)
	if err != nil {
		log.Fatal(err)
	}
	app.recordTree.SetModel(app.recordStore)

	app.updateRecords("")
	app.recordTree.ExpandAll()

	// Prepare to select the first record in the tree on update
	treeSelection, err := app.recordTree.GetSelection()
	if err != nil {
		log.Printf("Failed to get tree selection: %v", err)
	}
	treeSelection.SetMode(gtk.SELECTION_SINGLE)

	app.recordTree.Connect("row_activated", func() {
		db, record := app.getSelectedRecord()
		if record != nil {
			recordWindow(db, record)
		}
	})

	searchPaned, err := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		log.Fatal(err)
	}
	searchLabel, err := gtk.LabelNew("Search: ")
	if err != nil {
		log.Fatal(err)
	}
	searchPaned.Pack1(searchLabel, false, false)
	searchBox, err := gtk.EntryNew()
	if err != nil {
		log.Fatal(err)
	}
	searchBox.Connect("changed", func() {
		text, err := searchBox.GetText()
		if err != nil {
			log.Fatal(err)
		}
		app.updateRecords(text)
		app.recordTree.ExpandAll()
		for i := range app.dbs {
			firstEntryPath, err := gtk.TreePathNewFromString(strconv.Itoa(i) + ":0:0")
			if err != nil {
				log.Fatal(err)
			}
			treeSelection.SelectPath(firstEntryPath)
		}
	})
	searchBox.Connect("activate", func() {
		// TODO this duplicates the recordTree behaver, dedup
		db, record := app.getSelectedRecord()
		if record != nil {
			recordWindow(db, record)
		}
	})
	searchPaned.Pack2(searchBox, false, false)

	// If the window regains focus, select the entire selection of the searchBox
	window.Connect("focus-in-event", func() {
		searchBox.SelectRegion(0, -1)
	})

	// layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	if err != nil {
		log.Fatal(err)
	}
	vbox.PackStart(app.mainMenuBar(), false, false, 0) // TODO consider app.SetAppMenu or app.SetMenubar
	vbox.PackStart(searchPaned, false, false, 0)
	vbox.PackStart(recordFrame, true, true, 0)
	window.Add(vbox)
	window.SetDefaultSize(800, 800)
	window.Hide() // Start hidden, expose when a db is opened

	return &window.Window // TODO is this really what I want?
}

// getSelected returns a pwsafe.DB and pwsafe.Record matching the selected entry.
// If nothing is selected, returns nil,nil.
func (app *GoPWSafeGTK) getSelectedRecord() (pwsafe.DB, *pwsafe.Record) {
	selection, err := app.recordTree.GetSelection()
	if err != nil {
		// TODO amoung the many errors to decide on how to handle
		log.Printf("Failed to determine record tree selection: %v", err)
	}
	_, iter, ok := selection.GetSelected()
	if !ok {
		return nil, nil
	}
	rowValue, err := app.recordStore.GetValue(iter, 1)
	if err != nil {
		log.Fatal(err)
	}
	path, err := app.recordStore.GetPath(iter)
	if err != nil {
		log.Printf("Failed to determine selected path: %v", err)
	}
	activeDB, err := strconv.Atoi(strings.Split(path.String(), ":")[0])
	if err != nil {
		log.Fatal(err)
	}
	db := app.dbs[activeDB]

	// TODO fail gracefully if a non-leaf is selected.

	value, err := rowValue.GetString()
	if err != nil {
		log.Fatal(err)
	}
	record, success := db.GetRecord(value)
	if !success {
		app.errorDialog("Error retrieving record.")
		return db, nil
	}
	return db, &record
}

func (app *GoPWSafeGTK) updateRecords(search string) {
	app.recordStore.Clear()
	for i, db := range app.dbs {
		name := db.GetName()
		if name == "" {
			name = strconv.Itoa(i)
		}
		dbRoot := app.recordStore.Append(nil)
		err := app.recordStore.SetValue(dbRoot, 1, name)
		if err != nil {
			log.Fatal(err)
		}

		searchLower := strings.ToLower(search)
		for _, groupName := range db.Groups() {
			var matches []string
			for _, item := range db.ListByGroup(groupName) {
				if strings.Contains(strings.ToLower(item), searchLower) {
					matches = append(matches, item)
				}
			}
			if len(matches) > 0 {
				group := app.recordStore.Append(dbRoot)
				err := app.recordStore.SetValue(group, 1, groupName)
				if err != nil {
					log.Fatal(err)
				}

				for _, recordName := range matches {
					record := app.recordStore.Append(group)
					err := app.recordStore.SetValue(record, 1, recordName)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}
}

//TODO add a status bar and have it display messages like, copied username to clipboard, etc, see gtk.Statusbar
// Configures the main menubar and keyboard shortcuts
func (app *GoPWSafeGTK) mainMenuBar() *gtk.MenuBar {
	// Note of this writing the gotk3 implementation of gtkBuilder is causing me errors hence building menus
	// directly in the code
	parent := app.GetWindowByID(app.mainWindowID)

	// TODO I need to think about how to split up the menu into sections I can reuse for different windows.
	// fix all other menubar methods also
	mb, err := gtk.MenuBarNew()
	if err != nil {
		log.Fatal(err)
	}
	mb.Append(app.fileMenu())

	dbMenuItem, err := gtk.MenuItemNewWithLabel("DB")
	if err != nil {
		log.Fatal(err)
	}
	mb.Append(dbMenuItem)
	dbMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal(err)
	}
	dbMenuItem.SetSubmenu(dbMenu)
	dbAG, err := gtk.AccelGroupNew()
	if err != nil {
		log.Fatal(err)
	}
	parent.AddAccelGroup(dbAG)
	dbMenu.SetAccelGroup(dbAG)

	openDB, err := gtk.MenuItemNewWithLabel("OpenDB")
	if err != nil {
		log.Fatal(err)
	}
	openDB.Connect("activate", func() { app.openWindow("") })
	openDB.AddAccelerator("activate", dbAG, 't', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	dbMenu.Append(openDB)

	saveDB, err := gtk.MenuItemNewWithLabel("SaveDB")
	if err != nil {
		log.Fatal(err)
	}
	saveDB.Connect("activate", func() {
		db, _ := app.getSelectedRecord()
		if db != nil {
			app.propertiesWindow(db)
		} else {
			app.errorDialog("No DB is selected, please select a DB in the tree view to save")
		}
	})
	saveDB.AddAccelerator("activate", dbAG, 's', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	dbMenu.Append(saveDB)

	newDB, err := gtk.MenuItemNewWithLabel("NewDB")
	if err != nil {
		log.Fatal(err)
	}
	newDB.Connect("activate", func() {
		db := pwsafe.NewV3("", "")
		app.propertiesWindow(db)
		// TODO this annoying adds in the new DB even when cancel was clicked, fix that
		app.dbs = append(app.dbs, db)
	})
	dbMenu.Append(newDB)

	closeDB, err := gtk.MenuItemNewWithLabel("CloseDB")
	if err != nil {
		log.Fatal(err)
	}
	closeDB.Connect("activate", func() {
		//TODO close the selected or pop up a dialog not just the last
		app.dbs = app.dbs[:len(app.dbs)-1]
		// TODO either use the current selection in the search box or clear it out
		app.updateRecords("")
	})
	dbMenu.Append(closeDB)
	// TODO some shortcut ctrl-w ? something better?

	// TODO gui.go has a similar menuBar I need to combine into one with this
	recordMenuItem, err := gtk.MenuItemNewWithLabel("Record")
	if err != nil {
		log.Fatal(err)
	}
	mb.Append(recordMenuItem)
	recordMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal(err)
	}
	recordMenuItem.SetSubmenu(recordMenu)
	recordAG, err := gtk.AccelGroupNew()
	if err != nil {
		log.Fatal(err)
	}
	parent.AddAccelGroup(recordAG)
	recordMenu.SetAccelGroup(recordAG)

	newRecord, err := gtk.MenuItemNewWithLabel("New")
	if err != nil {
		log.Fatal(err)
	}
	newRecord.Connect("activate", func() {
		db, _ := app.getSelectedRecord()
		recordWindow(db, &pwsafe.Record{})
	})
	newRecord.AddAccelerator("activate", recordAG, 'n', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(newRecord)

	deleteRecord, err := gtk.MenuItemNewWithLabel("Delete")
	if err != nil {
		log.Fatal(err)
	}
	deleteRecord.Connect("activate", func() {
		db, record := app.getSelectedRecord()
		recordWindow(db, &pwsafe.Record{})
		db.DeleteRecord(record.Title)
	})
	recordMenu.Append(deleteRecord)

	// TODO all of the getSelectedRecord calls for menu items could fail more gracefully if nothing is selected or a non-leaf selected.
	// Also what happens if the selection is different than the open window right now it will follow the
	// selection which is bad behavior
	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err != nil {
		log.Fatal(err)
	}
	copyUser, err := gtk.MenuItemNewWithLabel("Copy Username")
	if err != nil {
		log.Fatal(err)
	}
	copyUser.Connect("activate", func() {
		_, record := app.getSelectedRecord()
		clipboard.SetText(record.Username)
	})
	copyUser.AddAccelerator("activate", recordAG, 'u', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(copyUser)

	copyPassword, err := gtk.MenuItemNewWithLabel("Copy Password")
	if err != nil {
		log.Fatal(err)
	}
	copyPassword.Connect("activate", func() {
		_, record := app.getSelectedRecord()
		clipboard.SetText(record.Password)
	})
	copyPassword.AddAccelerator("activate", recordAG, 'p', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(copyPassword)

	openURL, err := gtk.MenuItemNewWithLabel("Open URL")
	if err != nil {
		log.Fatal(err)
	}
	openURL.Connect("activate", func() {
		_, record := app.getSelectedRecord()
		open.Start(record.URL)
	})
	openURL.AddAccelerator("activate", recordAG, 'o', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(openURL)

	copyURL, err := gtk.MenuItemNewWithLabel("Copy URL")
	if err != nil {
		log.Fatal(err)
	}
	copyURL.Connect("activate", func() {
		_, record := app.getSelectedRecord()
		clipboard.SetText(record.URL)
	})
	copyURL.AddAccelerator("activate", recordAG, 'l', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(copyURL)

	return mb
}

func (app *GoPWSafeGTK) fileMenu() *gtk.MenuItem {
	fileMenuItem, err := gtk.MenuItemNewWithLabel("File")
	if err != nil {
		log.Fatal(err)
	}
	fileMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal(err)
	}
	fileMenuItem.SetSubmenu(fileMenu)

	// TODO the accelGroup should be defined somewhere else and just looked up
	fileAG, err := gtk.AccelGroupNew()
	if err != nil {
		log.Fatal(err)
	}
	parent := app.GetWindowByID(app.mainWindowID)
	parent.AddAccelGroup(fileAG)
	fileMenu.SetAccelGroup(fileAG)

	quit, err := gtk.MenuItemNewWithLabel("Quit")
	if err != nil {
		log.Fatal(err)
	}
	quit.Connect("activate", app.Quit)
	quit.AddAccelerator("activate", fileAG, 'q', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	fileMenu.Append(quit)

	return fileMenuItem
}
