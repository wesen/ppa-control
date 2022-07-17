package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"image/color"
)

func main() {
	a := app.New()
	w := a.NewWindow("ppa-control")

	green := color.NRGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}

	text1 := canvas.NewText("Hello World", green)
	text2 := canvas.NewText("Hello There", green)
	text2.Move(fyne.NewPos(20, 20))
	//content := container.NewWithoutLayout(text1, text2)
	content := container.New(layout.NewGridLayout(2), text1, text2)
	w.SetContent(content) // This is a text entry field
	w.Resize(fyne.NewSize(200, 200))

	w.ShowAndRun()
}
