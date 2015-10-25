package gui

import (
	"fmt"

	"github.com/tkuhlman/gopwsafe/config"
	"github.com/tkuhlman/gopwsafe/pwsafe"

	"github.com/google/gxui"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/light"
)

func loginWindow(driver gxui.Driver) {
	conf := config.Load()
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

	pathBox := theme.CreateTextBox()
	pathBox.SetDesiredWidth(math.MaxSize.W)
	pathBox.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	pathBox.SetMargin(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	//todo look into textbox_controller to allow a selection list of mulitple options
	hist := conf.GetPathHistory()
	if len(hist) > 0 {
		pathBox.SetText(hist[0])
	}
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
		openDB(driver, window, conf, pathBox.Text(), passwordBox.Text())
	})
	layout.AddChild(openButton)

	passwordBox.OnKeyDown(func(ev gxui.KeyboardEvent) {
		if ev.Key == gxui.KeyEnter || ev.Key == gxui.KeyKpEnter {
			window.Hide()
			openDB(driver, window, conf, pathBox.Text(), passwordBox.Text())
		}
	})

	window.AddChild(layout)
	window.SetFocus(pathBox)
	window.OnClose(driver.Terminate)
}

func openDB(driver gxui.Driver, previousWindow gxui.Window, conf config.PWSafeDBConfig, dbFile string, passwd string) {
	db, err := pwsafe.OpenPWSafeFile(dbFile, passwd)
	if err != nil {
		//todo figure out how to make the error dialog stay on top
		errorDialog(driver, fmt.Sprintf("Error Opening file %s\n%s", dbFile, err))
		previousWindow.Show()
		return
	}
	//todo handle duplicates and handle only keeping a certain amount of history
	err = conf.AddToPathHistory(dbFile)
	if err != nil {
		errorDialog(driver, fmt.Sprintf("Error adding %s to History\n%s", dbFile, err))
	}
	mainWindow(driver, db, conf)
}
