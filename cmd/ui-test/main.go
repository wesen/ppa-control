package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"time"
)

func main() {
	a := app.New()
	w := a.NewWindow("ppa-control")
	c := w.Canvas()

	blue := color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	rect := canvas.NewRectangle(blue)
	c.SetContent(rect)

	go func() {
		time.Sleep(2 * time.Second)
		green := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
		rect.FillColor = green
		rect.Refresh()

		time.Sleep(2 * time.Second)
		rect.Refresh()
		setContentToText(c)

		time.Sleep(2 * time.Second)
		rect.Refresh()
		setContentToCircle(c)
	}()

	w.ShowAndRun()
}

func setContentToText(c fyne.Canvas) {
	green := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	text := canvas.NewText("Hello World", green)
	text.TextStyle.Bold = true
	c.SetContent(text)
}

func setContentToCircle(c fyne.Canvas) {
	green := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	circle := canvas.NewCircle(green)
	circle.StrokeWidth = 4
	red := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	circle.StrokeColor = red
	c.SetContent(circle)
}
