package gui

import (
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

	window := theme.CreateWindow(380, 200, "GoPWSafe")

	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetDirection(gxui.TopToBottom)

	pathLabel := theme.CreateLabel()
	pathLabel.SetFont(font)
	pathLabel.SetText("Password DB path:")
	layout.AddChild(pathLabel)

	//todo add selectable entries from history
	//todo add a file selector option
	pathBox := theme.CreateTextBox()
	//	pathBox.SetText(dbFile)
	pathBox.SetPadding(math.ZeroSpacing)
	pathBox.SetMargin(math.ZeroSpacing)
	layout.AddChild(pathBox)

	passwdLabel := theme.CreateLabel()
	passwdLabel.SetFont(font)
	passwdLabel.SetText("Password:")
	layout.AddChild(passwdLabel)

	passwordBox := theme.CreateTextBox()
	passwordBox.SetPadding(math.ZeroSpacing)
	passwordBox.SetMargin(math.ZeroSpacing)
	layout.AddChild(passwordBox)

	openButton := theme.CreateButton()
	openButton.SetText("Open")
	//	openButton.OnClick(func(gxui.MouseEvent) { action(); update() })
	layout.AddChild(openButton)

	window.AddChild(layout)
	window.OnClose(driver.Terminate)
}

//Start Begins execution of the gui
func Start(dbFile string) int {
	gl.StartDriver(loginWindow)
	return 0
}
