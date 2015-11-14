package gui

import (
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

//todo add multiple db support
func mainWindow(db pwsafe.DB, conf config.PWSafeDBConfig) {

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

	updateRecords(db, recordStore, "")
	recordTree.ExpandAll()
	recordTree.Connect("row_activated", func() {
		recordWindow(getSelectedRecord(recordStore, recordTree, db))
	})

	searchPaned := gtk.NewHPaned()
	searchLabel := gtk.NewLabel("Search: ")
	searchPaned.Pack1(searchLabel, false, false)
	searchBox := gtk.NewEntry()
	searchBox.Connect("changed", func() {
		updateRecords(db, recordStore, searchBox.GetText())
		recordTree.ExpandAll()
	})
	searchPaned.Pack2(searchBox, false, false)

	//todo add a status bar that will be updated based on the recent actions performed

	// layout
	vbox := gtk.NewVBox(false, 1)
	vbox.PackStart(standardMenuBar(window), false, false, 0)
	vbox.PackStart(selectedRecordMenuBar(window, recordStore, recordTree, db), false, false, 0)
	vbox.PackStart(searchPaned, false, false, 0)
	vbox.Add(recordFrame)
	window.Add(vbox)
	window.SetSizeRequest(800, 800)
	window.ShowAll()
}

// return a db.Record matching the selected entry
func getSelectedRecord(recordStore *gtk.TreeStore, recordTree *gtk.TreeView, db pwsafe.DB) *pwsafe.Record {
	var path *gtk.TreePath
	var column *gtk.TreeViewColumn
	var iter gtk.TreeIter
	var rowValue glib.GValue
	model := recordStore.ToTreeModel()
	recordTree.GetCursor(&path, &column)
	model.GetIter(&iter, path)
	model.GetValue(&iter, 1, &rowValue)

	record, _ := db.GetRecord(rowValue.GetString())
	/* todo rather than _ have success and check but then I need to pass in the gtk window also, altenatively return the status and check in the main function
	if !success {
		errorDialog(window, "Error retrieving record.")
	}
	*/
	return &record
}

func updateRecords(db pwsafe.DB, store *gtk.TreeStore, search string) {
	store.Clear()
	var dbRoot gtk.TreeIter
	store.Append(&dbRoot, nil)
	store.Set(&dbRoot, gtk.NewImage().RenderIcon(gtk.STOCK_DIRECTORY, gtk.ICON_SIZE_SMALL_TOOLBAR, "").GPixbuf, db.GetName())

	searchLower := strings.ToLower(search)
	for _, groupName := range db.Groups() {
		var matches []string
		for _, item := range db.ListByGroup(groupName) {
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

//todo add a status bar and have it display messages like, copied username to clipboard, etc
// Configures the standard menubar and keyboard shortcuts
func standardMenuBar(window *gtk.Window) *gtk.Widget {
	actionGroup := gtk.NewActionGroup("standard")
	actionGroup.AddAction(gtk.NewAction("FileMenu", "File", "", ""))
	fileQuit := gtk.NewAction("FileQuit", "", "", gtk.STOCK_QUIT)
	fileQuit.Connect("activate", gtk.MainQuit)
	actionGroup.AddActionWithAccel(fileQuit, "<control>q")

	uiInfo := `
<ui>
  <menubar name='MenuBar'>
    <menu action='FileMenu'>
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
func selectedRecordMenuBar(window *gtk.Window, recordStore *gtk.TreeStore, recordTree *gtk.TreeView, db pwsafe.DB) *gtk.Widget {
	clipboard := gtk.NewClipboardGetForDisplay(gdk.DisplayGetDefault(), gdk.SELECTION_CLIPBOARD)

	actionGroup := gtk.NewActionGroup("record")
	actionGroup.AddAction(gtk.NewAction("RecordMenu", "Record", "", ""))

	copyUser := gtk.NewAction("CopyUsername", "Copy username to clipboard", "", "")
	copyUser.Connect("activate", func() { clipboard.SetText(getSelectedRecord(recordStore, recordTree, db).Username) })
	actionGroup.AddActionWithAccel(copyUser, "<control>u")

	copyPassword := gtk.NewAction("CopyPassword", "Copy password to clipboard", "", "")
	copyPassword.Connect("activate", func() { clipboard.SetText(getSelectedRecord(recordStore, recordTree, db).Password) })
	actionGroup.AddActionWithAccel(copyPassword, "<control>p")

	openURL := gtk.NewAction("OpenURL", "Open URL", "", "")
	// gtk-go hasn't yet implemented gtk_show_uri so using github.com/skratchdot/open-golang/open
	// todo it opens the url but should switch to that app also.
	openURL.Connect("activate", func() { open.Start(getSelectedRecord(recordStore, recordTree, db).URL) })
	actionGroup.AddActionWithAccel(openURL, "<control>o")

	copyURL := gtk.NewAction("CopyURL", "Copy URL to clipboard", "", "")
	copyURL.Connect("activate", func() { clipboard.SetText(getSelectedRecord(recordStore, recordTree, db).URL) })
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

//Start Begins execution of the gui
func Start(dbFile string) int {
	gtk.Init(nil)
	openWindow(dbFile)
	gtk.Main()
	return 0
}
