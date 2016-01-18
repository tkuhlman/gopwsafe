package gui

import (
	"strconv"
	"strings"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/skratchdot/open-golang/open"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

// GUI docs
// https://godoc.org/github.com/mattn/go-gtk
// https://developer.gnome.org/gtk-tutorial/stable/
// https://developer.gnome.org/gtk2/2.24/

func mainWindow(dbs []*pwsafe.DB, conf config.PWSafeDBConfig, dbFile string) {

	//todo revisit the structure of the gui code, splitting more out into functions and in general better organizing things.

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "Main Window")

	recordFrame := gtk.NewFrame("Records")
	recordWin := gtk.NewScrolledWindow(nil, nil)
	recordWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	recordWin.SetShadowType(gtk.SHADOW_IN)
	recordFrame.Add(recordWin)
	recordTree := gtk.NewTreeView()
	recordWin.Add(recordTree)
	recordStore := gtk.NewTreeStore(gdkpixbuf.GetType(), glib.G_TYPE_STRING)
	recordTree.SetModel(recordStore.ToTreeModel())
	recordTree.AppendColumn(gtk.NewTreeViewColumnWithAttributes("", gtk.NewCellRendererPixbuf(), "pixbuf", 0))
	recordTree.AppendColumn(gtk.NewTreeViewColumnWithAttributes("Name", gtk.NewCellRendererText(), "text", 1))

	updateRecords(&dbs, recordStore, "")
	recordTree.ExpandAll()

	// Prepare to select the first record in the tree on update
	treeSelection := recordTree.GetSelection()
	treeSelection.SetMode(gtk.SELECTION_SINGLE)

	recordTree.Connect("row_activated", func() {
		db, record := getSelectedRecord(recordStore, recordTree, &dbs)
		recordWindow(db, record)
	})

	searchPaned := gtk.NewHPaned()
	searchLabel := gtk.NewLabel("Search: ")
	searchPaned.Pack1(searchLabel, false, false)
	searchBox := gtk.NewEntry()
	searchBox.Connect("changed", func() {
		updateRecords(&dbs, recordStore, searchBox.GetText())
		recordTree.ExpandAll()
		for i := range dbs {
			firstEntryPath := gtk.NewTreePathFromString(strconv.Itoa(i) + ":0:0")
			treeSelection.SelectPath(firstEntryPath)
			if treeSelection.PathIsSelected(firstEntryPath) {
				break
			}
		}
	})
	searchPaned.Pack2(searchBox, false, false)

	// If the window regains focus, select the entire selection of the searchBox
	window.Connect("focus-in-event", func() {
		searchBox.SelectRegion(0, -1)
	})

	//todo add a status bar that will be updated based on the recent actions performed

	// layout
	vbox := gtk.NewVBox(false, 1)
	vbox.PackStart(mainMenuBar(window, &dbs, conf, recordStore), false, false, 0)
	vbox.PackStart(selectedRecordMenuBar(window, recordStore, recordTree, &dbs), false, false, 0)
	vbox.PackStart(searchPaned, false, false, 0)
	vbox.Add(recordFrame)
	window.Add(vbox)
	window.SetSizeRequest(800, 800)
	window.Hide() // Start hidden, expose when a db is opened

	// On first startup show the login window
	if len(dbs) == 0 {
		openWindow(dbFile, &dbs, conf, window, recordStore)
		recordTree.ExpandAll()
	}
}

// return a pwsafe.DB and pwsafe.Record matching the selected entry. If nothing is selected default to dbs[0] and an empty record
func getSelectedRecord(recordStore *gtk.TreeStore, recordTree *gtk.TreeView, dbs *[]*pwsafe.DB) (*pwsafe.DB, *pwsafe.Record) {
	var iter gtk.TreeIter
	var rowValue glib.GValue
	selection := recordTree.GetSelection()
	selection.GetSelected(&iter)
	model := recordStore.ToTreeModel()
	model.GetValue(&iter, 1, &rowValue)
	path := model.GetPath(&iter)
	pathStr := path.String()
	activeDB, err := strconv.Atoi(strings.Split(pathStr, ":")[0])
	if err != nil {
		db := (*dbs)[0] // Default to the first db if none is selected
		var record pwsafe.Record
		return db, &record
	}
	db := (*dbs)[activeDB]

	// todo fail gracefully if a non-leaf is selected.

	record, _ := (*db).GetRecord(rowValue.GetString())
	/* todo rather than _ have success and check but then I need to pass in the gtk window also, altenatively return the status and check in the main function
	if !success {
		errorDialog(window, "Error retrieving record.")
	}
	*/
	return db, &record
}

func updateRecords(dbs *[]*pwsafe.DB, store *gtk.TreeStore, search string) {
	store.Clear()
	for i, db := range *dbs {
		name := (*db).GetName()
		if name == "" {
			name = strconv.Itoa(i)
		}
		var dbRoot gtk.TreeIter
		store.Append(&dbRoot, nil)
		store.Set(&dbRoot, gtk.NewImage().RenderIcon(gtk.STOCK_DIRECTORY, gtk.ICON_SIZE_SMALL_TOOLBAR, "").GPixbuf, name)

		searchLower := strings.ToLower(search)
		for _, groupName := range (*db).Groups() {
			var matches []string
			for _, item := range (*db).ListByGroup(groupName) {
				if strings.Contains(strings.ToLower(item), searchLower) {
					matches = append(matches, item)
				}
			}
			if len(matches) > 0 {
				var group gtk.TreeIter
				store.Append(&group, &dbRoot)
				store.Set(&group, gtk.NewImage().RenderIcon(gtk.STOCK_DIRECTORY, gtk.ICON_SIZE_SMALL_TOOLBAR, "").GPixbuf, groupName)
				for _, recordName := range matches {
					var record gtk.TreeIter
					store.Append(&record, &group)
					store.Set(&record, gtk.NewImage().RenderIcon(gtk.STOCK_FILE, gtk.ICON_SIZE_SMALL_TOOLBAR, "").GPixbuf, recordName)
				}
			}
		}
	}
}

//todo add a status bar and have it display messages like, copied username to clipboard, etc
// Configures the main menubar and keyboard shortcuts
func mainMenuBar(window *gtk.Window, dbs *[]*pwsafe.DB, conf config.PWSafeDBConfig, recordStore *gtk.TreeStore) *gtk.Widget {
	actionGroup := gtk.NewActionGroup("main")
	actionGroup.AddAction(gtk.NewAction("FileMenu", "File", "", ""))

	openDB := gtk.NewAction("OpenDB", "Open a DB", "", "")
	openDB.Connect("activate", func() { openWindow("", dbs, conf, window, recordStore) })
	actionGroup.AddActionWithAccel(openDB, "<control>t")

	//todo - I need a save option.

	//todo, this doesn't actually work
	//todo close the selected or pop up a dialog not just the last
	closeDB := gtk.NewAction("CloseDB", "Close an open DB", "", "")
	closeDB.Connect("activate", func() {
		dbsValue := (*dbs)[:len(*dbs)-1]
		dbs = &dbsValue
	})
	actionGroup.AddActionWithAccel(closeDB, "")

	fileQuit := gtk.NewAction("FileQuit", "", "", gtk.STOCK_QUIT)
	fileQuit.Connect("activate", gtk.MainQuit)
	actionGroup.AddActionWithAccel(fileQuit, "<control>q")

	uiInfo := `
<ui>
  <menubar name='MenuBar'>
    <menu action='FileMenu'>
      <menuitem action='OpenDB' />
      <menuitem action='CloseDB' />
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

// todo this is remarkably similar to the recordMenuBar in gui/record.go the difference being this
// one doesn't get a record passed in but finds it from selection. I should think about how I could
// clearly and idiomatically reduce the duplication.
func selectedRecordMenuBar(window *gtk.Window, recordStore *gtk.TreeStore, recordTree *gtk.TreeView, dbs *[]*pwsafe.DB) *gtk.Widget {
	clipboard := gtk.NewClipboardGetForDisplay(gdk.DisplayGetDefault(), gdk.SELECTION_CLIPBOARD)

	actionGroup := gtk.NewActionGroup("record")
	actionGroup.AddAction(gtk.NewAction("RecordMenu", "Record", "", ""))

	newRecord := gtk.NewAction("NewRecord", "Add a new record to the selected db", "", "")
	newRecord.Connect("activate", func() {
		db, _ := getSelectedRecord(recordStore, recordTree, dbs)
		var record pwsafe.Record
		recordWindow(db, &record)
	})
	actionGroup.AddActionWithAccel(newRecord, "<control>n")

	deleteRecord := gtk.NewAction("DeleteRecord", "Deleted the selected record", "", "")
	deleteRecord.Connect("activate", func() {
		db, record := getSelectedRecord(recordStore, recordTree, dbs)
		//todo Pop up an are you sure dialog.
		(*db).DeleteRecord(record.Title)
	})
	actionGroup.AddActionWithAccel(deleteRecord, "Delete")

	//todo all of the getSelectedRecord calls for menu items could fail more gracefully if nothing is selected or a non-leaf selected.
	copyUser := gtk.NewAction("CopyUsername", "Copy username to clipboard", "", "")
	copyUser.Connect("activate", func() {
		_, record := getSelectedRecord(recordStore, recordTree, dbs)
		clipboard.SetText(record.Username)
	})
	actionGroup.AddActionWithAccel(copyUser, "<control>u")

	copyPassword := gtk.NewAction("CopyPassword", "Copy password to clipboard", "", "")
	copyPassword.Connect("activate", func() {
		_, record := getSelectedRecord(recordStore, recordTree, dbs)
		clipboard.SetText(record.Password)
	})
	actionGroup.AddActionWithAccel(copyPassword, "<control>p")

	openURL := gtk.NewAction("OpenURL", "Open URL", "", "")
	// gtk-go hasn't yet implemented gtk_show_uri so using github.com/skratchdot/open-golang/open
	// todo it opens the url but should switch to that app also.
	openURL.Connect("activate", func() {
		_, record := getSelectedRecord(recordStore, recordTree, dbs)
		open.Start(record.URL)
	})
	actionGroup.AddActionWithAccel(openURL, "<control>o")

	copyURL := gtk.NewAction("CopyURL", "Copy URL to clipboard", "", "")
	copyURL.Connect("activate", func() {
		_, record := getSelectedRecord(recordStore, recordTree, dbs)
		clipboard.SetText(record.URL)
	})
	actionGroup.AddActionWithAccel(copyURL, "<control>l")

	uiInfo := `
<ui>
  <menubar name='MenuBar'>
    <menu action='RecordMenu'>
      <menuitem action='NewRecord' />
      <menuitem action='DeleteRecord' />
      <menuitem action='CopyUsername' />
      <menuitem action='CopyPassword' />
      <menuitem action='OpenURL' />
      <menuitem action='CopyURL' />
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

//Start Begins execution of the gui
func Start(dbFile string) int {
	gtk.Init(nil)
	var dbs []*pwsafe.DB
	conf := config.Load()
	mainWindow(dbs, conf, dbFile)
	gtk.Main()
	return 0
}
