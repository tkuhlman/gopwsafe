package gui

import (
	"github.com/google/gxui"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/light"
)

func ErrorDialog(driver gxui.Driver, msg string) {
	theme := light.CreateTheme(driver)
	// todo figure out how to resize based on the message size.
	window := theme.CreateWindow(500, 250, "Error")
	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)
	layout.SetDirection(gxui.TopToBottom)
	layout.SetHorizontalAlignment(gxui.AlignCenter)

	font, err := driver.CreateFont(gxfont.Default, 25)
	if err != nil {
		panic(err)
	}
	theme.SetDefaultFont(font)

	errLabel := theme.CreateLabel()
	errLabel.SetMultiline(true)
	errLabel.SetText(msg)
	layout.AddChild(errLabel)

	OKButton := theme.CreateButton()
	OKButton.SetText("Okay")
	OKButton.OnClick(func(gxui.MouseEvent) {
		window.Close()
	})
	layout.AddChild(OKButton)

	window.AddChild(layout)
	// todo be smarter about positioning
	window.SetPosition(math.Point{1000, 1000})
	window.SetFocus(OKButton)
}
