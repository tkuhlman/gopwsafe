package gui

import (
	"fmt"

	"github.com/tkuhlman/gopwsafe/pwsafe"

	"github.com/google/gxui"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/light"
)

func loginWindow(driver gxui.Driver) {
	theme := light.CreateTheme(driver)

	font, err := driver.CreateFont(gxfont.Default, 25)
	if err != nil {
		panic(err)
	}
	theme.SetDefaultFont(font)

	window := theme.CreateWindow(500, 300, "GoPWSafe")

	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetDirection(gxui.TopToBottom)
	layout.SetHorizontalAlignment(gxui.AlignCenter)

	pathLabel := theme.CreateLabel()
	pathLabel.SetText("Password DB path: changed")
	layout.AddChild(pathLabel)

	//todo add selectable entries from history
	pathBox := theme.CreateTextBox()
	pathBox.SetDesiredWidth(math.MaxSize.W)
	pathBox.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	pathBox.SetMargin(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	layout.AddChild(pathBox)

	passwdLabel := theme.CreateLabel()
	passwdLabel.SetText("Password:")
	layout.AddChild(passwdLabel)

	//todo hide password
	//I am subscribed to this bug for this https://github.com/google/gxui/issues/119
	// If needed I could consider possibly using OnTextChanged handler + SetText
	passwordBox := theme.CreateTextBox()
	passwordBox.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	passwordBox.SetMargin(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	passwordBox.SetDesiredWidth(math.MaxSize.W)
	layout.AddChild(passwordBox)

	openButton := theme.CreateButton()
	openButton.SetText("Open")
	openButton.OnClick(func(gxui.MouseEvent) {
		window.Hide()
		openDB(driver, window, pathBox.Text(), passwordBox.Text())
	})
	layout.AddChild(openButton)

	passwordBox.OnKeyDown(func(ev gxui.KeyboardEvent) {
		if ev.Key == gxui.KeyEnter || ev.Key == gxui.KeyKpEnter {
			window.Hide()
			openDB(driver, window, pathBox.Text(), passwordBox.Text())
		}
	})

	window.AddChild(layout)
	window.SetFocus(pathBox)
	window.OnClose(driver.Terminate)
}

func openDB(driver gxui.Driver, previousWindow gxui.Window, dbFile string, passwd string) {
	db, err := pwsafe.OpenPWSafeFile(dbFile, passwd)
	if err != nil {
		ErrorDialog(driver, fmt.Sprintf("Error Opening file %s\n%s", dbFile, err))
		//todo figure out how to
		previousWindow.Show()
		return
	} else {
		previousWindow.Show()
		mainWindow(driver, db)
	}
}
