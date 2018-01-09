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

// GoPWSafeGTK wraps gtk.Application adding variables for state needed for this particular application
// I'm using gtk.Application though gotk3 support for gio is lacking making the menus and associated
// accelerators more painful to implement.
// https://wiki.gnome.org/HowDoI/GtkApplication
type GoPWSafeGTK struct {
	*gtk.Application
	accelGroup *gtk.AccelGroup
	conf       config.PWSafeDBConfig
	dbs        []pwsafe.DB
	// TODO having the recordStore/recordTree and window ids in here doesn't seem to be the best structure
	recordStore  *gtk.TreeStore
	recordTree   *gtk.TreeView
	mainWindowID uint
}

func NewGoPWSafeGTK() (*GoPWSafeGTK, error) {
	// TODO add glib.APPLICATION_HANDLES_OPEN and figure out how to pass in a path to openWindow
	gtkApp, err := gtk.ApplicationNew("tkuhlman.gopwsafe", glib.APPLICATION_NON_UNIQUE)
	if err != nil {
		return nil, err
	}

	ag, err := gtk.AccelGroupNew()
	logError(err, "")

	app := &GoPWSafeGTK{
		Application: gtkApp,
		accelGroup:  ag,
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

//startUp handles the startUp signal for the GTK application by defining the main window.
func (app *GoPWSafeGTK) startUp(gtkApp *gtk.Application) {
	app.AddWindow(app.mainWindow())
}

//open handles the open and activate signals for the GTK application by starting the open window.
func (app *GoPWSafeGTK) open(gtkApp *gtk.Application) {
	app.openWindow("") // TODO handle a path passed to the application see, glib.APPLICATION_HANDLES_OPEN todo item
}

// shutdown handles the shutdown signal for the GTK application.
func (app *GoPWSafeGTK) shutdown(gtkApp *gtk.Application) {
	app.Quit()
}

// mainWindow is the primary application window all other windows are accesories to it.
func (app *GoPWSafeGTK) mainWindow() *gtk.Window {

	//TODO revisit the structure of the gui code, splitting more out into functions and in general better organizing things.

	window, err := gtk.ApplicationWindowNew(app.Application)
	logError(err, "Failed to create main window")
	app.mainWindowID = window.GetID()
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.AddAccelGroup(app.accelGroup)
	window.Window.Connect("destroy", func() {
		// Check if any dbs need to be saved
		for _, db := range app.dbs {
			if db.NeedsSave() {
				app.propertiesWindow(db)
				app.errorDialog(fmt.Sprintf("Unsaved changes for db %v", db.GetName()))
			}
		}
	})

	recordFrame, err := gtk.FrameNew("Records")
	logError(err, "Failed to create record frame")
	recordWin, err := gtk.ScrolledWindowNew(nil, nil)
	logError(err, "Failed to create scrolled window")
	recordWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	recordFrame.SetShadowType(gtk.SHADOW_IN)
	recordFrame.Add(recordWin)

	app.recordTree, err = gtk.TreeViewNew()
	logError(err, "Failed to create tree view")
	recordWin.Add(app.recordTree)

	cellPB, err := gtk.CellRendererPixbufNew()
	logError(err, "")
	col1, err := gtk.TreeViewColumnNewWithAttribute("", cellPB, "pixbuf", 0)
	logError(err, "")
	app.recordTree.AppendColumn(col1)
	cellText, err := gtk.CellRendererTextNew()
	logError(err, "")
	col2, err := gtk.TreeViewColumnNewWithAttribute("Name", cellText, "text", 1)
	logError(err, "")
	app.recordTree.AppendColumn(col2)

	app.recordStore, err = gtk.TreeStoreNew(glib.TYPE_OBJECT, glib.TYPE_STRING)
	logError(err, "")
	app.recordTree.SetModel(app.recordStore)

	app.updateRecords("")
	app.recordTree.ExpandAll()

	// Prepare to select the first record in the tree on update
	treeSelection, err := app.recordTree.GetSelection()
	logError(err, "Failed to get tree selection")
	treeSelection.SetMode(gtk.SELECTION_SINGLE)

	app.recordTree.Connect("row_activated", func() {
		db, record := app.getSelectedRecord()
		if record != nil {
			app.recordWindow(db, record)
		}
	})

	searchPaned, err := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	logError(err, "")
	searchLabel, err := gtk.LabelNew("Search: ")
	logError(err, "")
	searchPaned.Pack1(searchLabel, false, false)
	searchBox, err := gtk.EntryNew()
	logError(err, "")
	searchBox.Connect("changed", func() {
		text, err := searchBox.GetText()
		logError(err, "")
		app.updateRecords(text)
		app.recordTree.ExpandAll()
		for i := range app.dbs {
			firstEntryPath, err := gtk.TreePathNewFromString(strconv.Itoa(i) + ":0:0")
			logError(err, "")
			treeSelection.SelectPath(firstEntryPath)
		}
	})
	searchBox.Connect("activate", func() {
		db, record := app.getSelectedRecord()
		if record != nil {
			app.recordWindow(db, record)
		}
	})
	// Only one or the other of the searchbox or selected tree value should be hilighted
	searchBox.Connect("focus-in-event", func() {
		searchBox.SelectRegion(0, -1)
		recordFrame.Bin.Container.Widget.SetStateFlags(gtk.STATE_FLAG_BACKDROP, true)
		searchBox.Widget.SetStateFlags(gtk.STATE_FLAG_FOCUSED, true)
	})
	searchBox.Connect("focus-out-event", func() {
		recordFrame.Bin.Container.Widget.SetStateFlags(gtk.STATE_FLAG_FOCUSED, true)
		searchBox.Widget.SetStateFlags(gtk.STATE_FLAG_BACKDROP, true)
	})
	// By default when switch focus back to this window it hilights both the search box and selected
	// item in the tree, making it hard to tell where the focus is, this fixes it
	window.Window.Connect("style-updated", func() {
		if searchBox.Widget.HasFocus() {
			recordFrame.Bin.Container.Widget.SetStateFlags(gtk.STATE_FLAG_BACKDROP, true)
			searchBox.Widget.SetStateFlags(gtk.STATE_FLAG_FOCUSED, true)
		} else {
			recordFrame.Bin.Container.Widget.SetStateFlags(gtk.STATE_FLAG_FOCUSED, true)
			searchBox.Widget.SetStateFlags(gtk.STATE_FLAG_BACKDROP, true)
		}
	})
	searchPaned.Pack2(searchBox, false, false)

	// layout
	vbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	logError(err, "")
	vbox.PackStart(app.mainMenuBar(), false, false, 0)
	vbox.PackStart(searchPaned, false, false, 0)
	vbox.PackEnd(recordFrame, true, true, 0)
	window.Add(vbox)
	window.SetDefaultSize(800, 800)
	window.Hide() // Start hidden, expose when a db is opened

	return &window.Window
}

// getSelectedRecord returns a pwsafe.DB and pwsafe.Record matching the selected entry.
// If nothing is selected, returns nil,nil.
func (app *GoPWSafeGTK) getSelectedRecord() (pwsafe.DB, *pwsafe.Record) {
	selection, err := app.recordTree.GetSelection()
	if err != nil {
		log.Printf("Failed to determine record tree selection: %v", err)
	}
	_, iter, ok := selection.GetSelected()
	if !ok {
		return nil, nil
	}
	rowValue, err := app.recordStore.GetValue(iter, 1)
	logError(err, "")
	path, err := app.recordStore.GetPath(iter)
	if err != nil {
		log.Printf("Failed to determine selected path: %v", err)
	}
	activeDB, err := strconv.Atoi(strings.Split(path.String(), ":")[0])
	logError(err, "")
	db := app.dbs[activeDB]

	// TODO fail gracefully if a non-leaf is selected.

	value, err := rowValue.GetString()
	logError(err, "")
	record, success := db.GetRecord(value)
	if !success {
		return db, nil
	}
	return db, &record
}

func (app *GoPWSafeGTK) updateRecords(search string) {
	// TODO it would be ideal if updateRecords could read the search field itself
	var iconErrors []error
	icons, err := gtk.IconThemeGetDefault()
	iconErrors = append(iconErrors, err)
	rootIcon, err := icons.LoadIcon("dialog-password", 16, gtk.ICON_LOOKUP_FORCE_SIZE)
	iconErrors = append(iconErrors, err)
	folderIcon, err := icons.LoadIcon("folder", 16, gtk.ICON_LOOKUP_FORCE_SIZE)
	iconErrors = append(iconErrors, err)
	recordIcon, err := icons.LoadIcon("text-x-generic", 16, gtk.ICON_LOOKUP_FORCE_SIZE)
	iconErrors = append(iconErrors, err)
	for _, e := range iconErrors {
		if e != nil {
			log.Print(e)
		}
	}

	app.recordStore.Clear()
	for i, db := range app.dbs {
		name := db.GetName()
		if name == "" {
			name = strconv.Itoa(i)
		}
		dbRoot := app.recordStore.Append(nil)
		err := app.recordStore.SetValue(dbRoot, 0, rootIcon)
		logError(err, "")
		err = app.recordStore.SetValue(dbRoot, 1, name)
		logError(err, "")

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
				err := app.recordStore.SetValue(group, 0, folderIcon)
				logError(err, "")
				err = app.recordStore.SetValue(group, 1, groupName)
				logError(err, "")

				for _, recordName := range matches {
					record := app.recordStore.Append(group)
					err := app.recordStore.SetValue(record, 0, recordIcon)
					logError(err, "")
					err = app.recordStore.SetValue(record, 1, recordName)
					logError(err, "")
				}
			}
		}
	}
}

// Configures the main menubar and keyboard shortcuts
func (app *GoPWSafeGTK) mainMenuBar() *gtk.MenuBar {
	// Note of this writing the gotk3 implementation of gtkBuilder is causing me errors hence building menus
	// directly in the code
	parent := app.GetWindowByID(app.mainWindowID)

	mb, err := gtk.MenuBarNew()
	logError(err, "")
	mb.Append(app.fileMenu())

	dbMenuItem, err := gtk.MenuItemNewWithLabel("DB")
	logError(err, "")
	mb.Append(dbMenuItem)
	dbMenu, err := gtk.MenuNew()
	logError(err, "")
	dbMenuItem.SetSubmenu(dbMenu)
	dbAG, err := gtk.AccelGroupNew()
	logError(err, "")
	parent.AddAccelGroup(dbAG)
	dbMenu.SetAccelGroup(dbAG)

	openDB, err := gtk.MenuItemNewWithLabel("Open")
	logError(err, "")
	openDB.Connect("activate", func() { app.openWindow("") })
	openDB.AddAccelerator("activate", dbAG, 't', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	dbMenu.Append(openDB)

	saveDB, err := gtk.MenuItemNewWithLabel("Save")
	logError(err, "")
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

	newDB, err := gtk.MenuItemNewWithLabel("New")
	logError(err, "")
	newDB.Connect("activate", func() {
		db := pwsafe.NewV3("", "")
		app.propertiesWindow(db)
	})
	dbMenu.Append(newDB)

	newRecord, err := gtk.MenuItemNewWithLabel("New Record")
	logError(err, "")
	newRecord.Connect("activate", func() {
		db, _ := app.getSelectedRecord()
		app.recordWindow(db, &pwsafe.Record{})
	})
	newRecord.AddAccelerator("activate", app.accelGroup, 'n', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	dbMenu.Append(newRecord)

	deleteRecord, err := gtk.MenuItemNewWithLabel("Delete Record")
	logError(err, "")
	deleteRecord.Connect("activate", func() {
		db, record := app.getSelectedRecord()
		if record == nil {
			app.errorDialog("Error retrieving record.")
		}
		app.recordWindow(db, &pwsafe.Record{})
		db.DeleteRecord(record.Title)
	})
	dbMenu.Append(deleteRecord)

	closeDB, err := gtk.MenuItemNewWithLabel("Close")
	logError(err, "")
	closeDB.Connect("activate", func() {
		//TODO close the selected or pop up a dialog not just the last
		app.dbs = app.dbs[:len(app.dbs)-1]
		// TODO either use the current selection in the search box or clear it out
		app.updateRecords("")
	})
	dbMenu.Append(closeDB)

	mb.Append(app.recordMenu(parent, nil))

	return mb
}

func (app *GoPWSafeGTK) fileMenu() *gtk.MenuItem {
	fileMenuItem, err := gtk.MenuItemNewWithLabel("File")
	logError(err, "")
	fileMenu, err := gtk.MenuNew()
	logError(err, "")
	fileMenuItem.SetSubmenu(fileMenu)
	fileMenu.SetAccelGroup(app.accelGroup)

	quit, err := gtk.MenuItemNewWithLabel("Quit")
	logError(err, "")
	quit.Connect("activate", app.Quit)
	quit.AddAccelerator("activate", app.accelGroup, 'q', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	fileMenu.Append(quit)

	return fileMenuItem
}

func (app *GoPWSafeGTK) recordMenu(parent *gtk.Window, record *pwsafe.Record) *gtk.MenuItem {
	recordMenuItem, err := gtk.MenuItemNewWithLabel("Record")
	logError(err, "")
	recordMenu, err := gtk.MenuNew()
	logError(err, "")
	recordMenuItem.SetSubmenu(recordMenu)
	recordAG, err := gtk.AccelGroupNew()
	logError(err, "")
	parent.AddAccelGroup(recordAG)
	recordMenu.SetAccelGroup(recordAG)

	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	logError(err, "")
	copyUser, err := gtk.MenuItemNewWithLabel("Copy Username")
	logError(err, "")
	copyUser.Connect("activate", func() {
		if record == nil {
			if selectedRecord, ok := app.checkSelectedRecord(); ok {
				clipboard.SetText(selectedRecord.Username)
			}
		} else {
			clipboard.SetText(record.Username)
		}
	})
	copyUser.AddAccelerator("activate", recordAG, 'u', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(copyUser)

	copyPassword, err := gtk.MenuItemNewWithLabel("Copy Password")
	logError(err, "")
	copyPassword.Connect("activate", func() {
		if record == nil {
			if selectedRecord, ok := app.checkSelectedRecord(); ok {
				clipboard.SetText(selectedRecord.Password)
			}
		} else {
			clipboard.SetText(record.Password)
		}
	})
	copyPassword.AddAccelerator("activate", recordAG, 'p', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(copyPassword)

	openURL, err := gtk.MenuItemNewWithLabel("Open URL")
	logError(err, "")
	openURL.Connect("activate", func() {
		if record == nil {
			if selectedRecord, ok := app.checkSelectedRecord(); ok {
				open.Start(selectedRecord.URL)
			}
		} else {
			open.Start(record.URL)
		}
	})
	openURL.AddAccelerator("activate", recordAG, 'o', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(openURL)

	copyURL, err := gtk.MenuItemNewWithLabel("Copy URL")
	logError(err, "")
	copyURL.Connect("activate", func() {
		if record == nil {
			if selectedRecord, ok := app.checkSelectedRecord(); ok {
				clipboard.SetText(selectedRecord.URL)
			}
		} else {
			clipboard.SetText(record.URL)
		}
	})
	copyURL.AddAccelerator("activate", recordAG, 'l', gdk.GDK_CONTROL_MASK, gtk.ACCEL_VISIBLE)
	recordMenu.Append(copyURL)

	return recordMenuItem
}

// checkSelectedRecord retrieves the selected record, validates it and either
// pops up an error dialog or returns the record.
func (app *GoPWSafeGTK) checkSelectedRecord() (*pwsafe.Record, bool) {
	_, selectedRecord := app.getSelectedRecord()
	if selectedRecord == nil {
		app.errorDialog("Error retrieving record.")
		return nil, false
	}
	return selectedRecord, true
}

// logError handles errors that are unexpected to occur in normal funtioning of the app. If provided
// preface string will be used to build a new error before logging.
// string is used to preface the error
func logError(err error, preface string) {
	if err != nil {
		if preface == "" {
			log.Fatal(err)
		} else {
			log.Fatalf("%s: %v", preface, err)
		}
	}
}
