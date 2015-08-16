package gui

import (
	"github.com/tkuhlman/gopwsafe/pwsafe"

	"github.com/andlabs/ui"
)

var window ui.Window

//Start Begins execution of the gui
func Start(dbFile string) int {
	// Consider using OpenFile more details at http://godoc.org/github.com/andlabs/ui
	path := ui.NewTextField()
	password := ui.NewTextField()
	button := ui.NewButton("Open")
	stack := ui.NewVerticalStack(
		ui.NewLabel("Password file path:"),
		path,
		ui.NewLabel("Password:"),
		password,
		button)
	window = ui.NewWindow("GoPWSafe", 200, 100, stack)

	button.OnClicked(func() {
		_, err := pwsafe.OpenPWSafeFile(path.Text(), password.Text())
		if err == nil {
			errorDialog("It worked")
		} else {
			errorDialog("Error opening file: " + err.Error())
		}
	})
	window.OnClosing(func() bool {
		ui.Stop()
		return true
	})
	window.Show()

	err := ui.Go()
	if err != nil {
		panic(err)
	}
	return 0
}

func errorDialog(msg string) {
	stack := ui.NewVerticalStack(ui.NewLabel(msg))
	window = ui.NewWindow("Error", 100, 100, stack)
	window.OnClosing(func() bool {
		ui.Stop()
		return true
	})
	window.Show()
}
