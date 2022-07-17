package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("ppa-control")
	w.SetContent(widget.NewEntry()) // This is a text entry field
	w.ShowAndRun()
}
