package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("ppa-control")

	w.SetContent(container.NewVBox(makeUI()))

	w.Resize(fyne.NewSize(300, 300))
	w.Show()

	a.Run()
}

func makeUI() (*widget.Label, *widget.Entry) {
	in := widget.NewEntry()
	out := widget.NewLabel("Hello World")
	in.OnChanged = func(content string) {
		out.SetText("Hello " + content)
	}
	return out, in
}
