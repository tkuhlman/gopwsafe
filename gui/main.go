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

	// todo add a menu
	// todo add and about dialog
	menubar := gtk.NewMenuBar()

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
	})
	searchPaned.Pack2(searchBox, false, false)

	//todo add a status bar that will be updated based on the recent actions performed

	// layout
	vbox := gtk.NewVBox(false, 1)
	vbox.PackStart(menubar, false, false, 0)
	vbox.Add(searchPaned)
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

//Start Begins execution of the gui
func Start(dbFile string) int {
	// todo ctrl-q should work for exit for all windows
	gtk.Init(nil)
	openWindow(dbFile)
	gtk.Main()
	return 0
}
