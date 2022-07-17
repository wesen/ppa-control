package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
)

func main() {
	a := app.New()
	w := a.NewWindow("ppa-control")

	serverConsole := canvas.NewText("ServerConsole\nServerConsole", color.White)
	//serverConsole := widget.NewLabel("ServerConsole\nServerConsole\nServerConsole")

	clientConsole := canvas.NewText("Hello There", color.White)

	var presetButtons = make([]fyne.CanvasObject, 8)
	for i := 0; i < 8; i++ {
		presetButtons[i] = widget.NewButton(fmt.Sprintf("Preset %d", i+1),
			func() {
				log.Println(fmt.Sprintf("Preset %d clicked", i+1))
			})
	}
	presetButtonContainer := container.New(layout.NewGridLayout(4), presetButtons...)

	var controlButtons []fyne.CanvasObject = make([]fyne.CanvasObject, 8)
	for i := 0; i < 8; i++ {
		controlButtons[i] = widget.NewButton(fmt.Sprintf("Control %d", i+1),
			func() {
				log.Println(fmt.Sprintf("Control %d clicked", i+1))
			})
	}
	controlButtonContainer := container.New(layout.NewGridLayout(4), controlButtons...)

	var volumeButtons = make([]fyne.CanvasObject, 4)
	volumeButtons[0] = widget.NewButton("High", func() {
		log.Println("Volume HIGH")
	})
	volumeButtons[1] = widget.NewButton("Mid", func() {
		log.Println("Volume MID")
	})
	volumeButtons[2] = widget.NewButton("Low", func() {
		log.Println("Volume LOW")
	})
	volumeButtons[3] = widget.NewButton("Mute", func() {
		log.Println("Volume MUTE TOGGLE")
	})
	volumeContainer := container.New(layout.NewVBoxLayout(), volumeButtons...)

	// we also need title fields
	// 4 buttons for the fixed volumes
	// a side bar and the volume

	mainGridContainer := container.NewVBox(
		serverConsole,
		widget.NewSeparator(),
		clientConsole,
		widget.NewSeparator(),
		presetButtonContainer,
		widget.NewSeparator(),
		volumeContainer,
		widget.NewSeparator(),
		controlButtonContainer)

	serverConsole.Text = "foobar\nFoo Foo\nblablabla"
	serverConsole.Refresh()

	masterSlider := widget.NewSlider(0, 1)
	masterSlider.Step = 0.01
	masterSlider.Orientation = widget.Vertical
	masterSlider.OnChanged = func(value float64) {
		log.Println(fmt.Sprintf("Master Volume: %f", value))
	}
	masterContainer := container.NewVBox(masterSlider)
	masterContainer.Resize(fyne.NewSize(50, 700))

	mainHBox := container.NewHBox(mainGridContainer, widget.NewSeparator(), masterContainer)
	w.SetContent(mainHBox) // This is a text entry field
	w.Resize(fyne.NewSize(800, 800))

	w.ShowAndRun()
}
