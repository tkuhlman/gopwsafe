package gui

import (
	"github.com/tkuhlman/gopwsafe/pwsafe"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/themes/light"
)

func mainWindow(driver gxui.Driver, db pwsafe.DB) {
	theme := light.CreateTheme(driver)
	font, err := driver.CreateFont(gxfont.Default, 20)
	if err != nil {
		panic(err)
	}
	theme.SetDefaultFont(font)
	window := theme.CreateWindow(500, 500, "GoPWSafe")
	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetDirection(gxui.TopToBottom)
	window.AddChild(layout)
	window.OnClose(driver.Terminate)

	for _, item := range db.List() {
		layout.AddChild(listEntry(theme, item))
	}
}

func listEntry(theme gxui.Theme, name string) gxui.Label {
	item := theme.CreateLabel()
	item.SetText(name)
	return item
}

//Start Begins execution of the gui
func Start(dbFile string) int {
	gl.StartDriver(loginWindow)
	return 0
}
