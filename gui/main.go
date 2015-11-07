package gui

import (
	"strings"

	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
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
		// Find the record name for the row being activated
		var path *gtk.TreePath
		var column *gtk.TreeViewColumn
		var iter gtk.TreeIter
		var rowValue glib.GValue
		model := recordStore.ToTreeModel()
		recordTree.GetCursor(&path, &column)
		model.GetIter(&iter, path)
		model.GetValue(&iter, 1, &rowValue)

		record, _ := db.GetRecord(rowValue.GetString())
		recordWindow(&record)
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
	vbox.PackStart(searchPaned, false, false, 0)
	vbox.Add(recordFrame)
	window.Add(vbox)
	window.SetSizeRequest(800, 800)
	window.ShowAll()
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
