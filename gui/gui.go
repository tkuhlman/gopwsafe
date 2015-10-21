package gui

import (
	"github.com/tkuhlman/gopwsafe/pwsafe"

	log "github.com/Sirupsen/logrus"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
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

	window := theme.CreateWindow(380, 250, "GoPWSafe")

	//todo validate common keyboard shortcuts.
	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetDirection(gxui.TopToBottom)
	layout.SetHorizontalAlignment(gxui.AlignCenter)
	layout.HorizontalAlignment().AlignCenter()

	pathLabel := theme.CreateLabel()
	pathLabel.SetFont(font)
	pathLabel.SetText("Password DB path: changed")
	layout.AddChild(pathLabel)

	//todo add selectable entries from history
	//todo add a file selection dialog box
	pathBox := theme.CreateTextBox()
	pathBox.SetDesiredWidth(math.MaxSize.W)
	pathBox.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	pathBox.SetMargin(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	layout.AddChild(pathBox)

	passwdLabel := theme.CreateLabel()
	passwdLabel.SetFont(font)
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
		openDB(driver, pathBox.Text(), passwordBox.Text())
		window.Close()
	})
	layout.AddChild(openButton)

	passwordBox.OnKeyDown(func(ev gxui.KeyboardEvent) {
		if ev.Key == gxui.KeyEnter || ev.Key == gxui.KeyKpEnter {
			openDB(driver, pathBox.Text(), passwordBox.Text())
			window.Close()
		}
	})

	//todo I need a driver.Terminate to trigger when the login window is closed without the main window open

	window.AddChild(layout)
}

func openDB(driver gxui.Driver, dbFile string, passwd string) {
	db, err := pwsafe.OpenPWSafeFile(dbFile, passwd)
	if err != nil {
		//todo ditch logging and instead pop up an error dialog
		log.WithFields(log.Fields{"File": dbFile, "Error": err}).Error("Error Opening file")
		loginWindow(driver)
		return
	} else {
		mainWindow(driver, db)
	}
}

func mainWindow(driver gxui.Driver, db pwsafe.DB) {
	theme := light.CreateTheme(driver)
	window := theme.CreateWindow(500, 500, "GoPWSafe")
	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetDirection(gxui.TopToBottom)
	window.AddChild(layout)
	window.OnClose(driver.Terminate)
}

//Start Begins execution of the gui
func Start(dbFile string) int {
	gl.StartDriver(loginWindow)
	return 0
}
