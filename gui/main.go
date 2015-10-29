package gui

import (
	"fmt"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"
)

// GUI docs
// https://godoc.org/github.com/mattn/go-gtk
// https://developer.gnome.org/gtk2/2.24/

func mainWindow(db pwsafe.DB, conf config.PWSafeDBConfig) {

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GoPWSafe")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		gtk.MainQuit()
	}, "Main Window")

	swin := gtk.NewScrolledWindow(nil, nil)
	swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	swin.SetShadowType(gtk.SHADOW_IN)
	textview := gtk.NewTextView()
	var start gtk.TextIter
	buffer := textview.GetBuffer()
	buffer.GetStartIter(&start)

	for _, item := range db.List() {
		// todo make sure the default font doesn't do stupid things like mix up I l 1, etc
		buffer.Insert(&start, fmt.Sprintf("%v\n%", item))
	}
	swin.Add(textview)
	window.Add(swin)

	// todo add a menu
	window.SetSizeRequest(800, 800)
	window.ShowAll()
}

//Start Begins execution of the gui
func Start(dbFile string) int {
	// todo ctrl-q should work for exit for all windows
	gtk.Init(nil)
	openWindow(dbFile)
	gtk.Main()
	return 0
}
