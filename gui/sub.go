// miscellaneous subordinate windows
package gui

import "github.com/mattn/go-gtk/gtk"

func errorDialog(parent *gtk.Window, msg string) {
	messagedialog := gtk.NewMessageDialog(
		parent,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_OK,
		msg)
	messagedialog.Response(func() {
		messagedialog.Destroy()
	})
	messagedialog.Run()
}
