package gui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

// GUI docs
// https://godoc.org/github.com/mattn/go-gtk
// https://developer.gnome.org/gtk-tutorial/stable/
// https://developer.gnome.org/gtk2/2.24/

func mainWindow(db pwsafe.DB, conf config.PWSafeDBConfig) {

	//todo revisit the structure of the gui code, splitting more out into functions and in general better organizing things.

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "Main Window")

	recordFrame := gtk.NewFrame("Records")
	//todo Make into a tree view so I can easily distinguish multiple DBs, also use for grouping by pw group
	//	recordTree := gtk.NewTreeView()
	recordWin := gtk.NewScrolledWindow(nil, nil)
	recordWin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	recordWin.SetShadowType(gtk.SHADOW_IN)
	recordTextView := gtk.NewTextView()
	recordTextView.SetEditable(false)
	recordWin.Add(recordTextView)
	recordBuffer := recordTextView.GetBuffer()
	updateRecords(db, recordBuffer, "")
	recordFrame.Add(recordWin)

	searchPaned := gtk.NewHPaned()
	searchLabel := gtk.NewLabel("Search: ")
	searchPaned.Pack1(searchLabel, false, false)
	searchBox := gtk.NewEntry()
	searchBox.Connect("changed", func() {
		updateRecords(db, recordBuffer, searchBox.GetText())
		// todo look at GtkEntryCompletion to see if that is a better way to approach this, https://developer.gnome.org/gtk3/stable/GtkEntryCompletion.html
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

func updateRecords(db pwsafe.DB, buffer *gtk.TextBuffer, search string) {
	var end, start gtk.TextIter
	buffer.GetStartIter(&start)
	buffer.GetEndIter(&end)
	buffer.Delete(&start, &end)
	searchLower := strings.ToLower(search)

	for _, item := range db.List() {
		// todo make sure the default font doesn't do stupid things like mix up I l 1, etc
		if strings.Contains(strings.ToLower(item), searchLower) {
			buffer.Insert(&start, fmt.Sprintf("%s\n", item))
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
