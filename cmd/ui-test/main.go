package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"time"
)

func main() {
	a := app.New()
	w := a.NewWindow("ppa-control")

	//w.SetContent(widget.NewLabel("Hello Fyne!"))
	clock := widget.NewLabel("")
	w.SetContent(clock)

	updateTime(clock)
	go func() {
		for range time.Tick(time.Second) {
			updateTime(clock)
		}
	}()

	w.Resize(fyne.NewSize(300, 300))
	w.Show()

	w2 := a.NewWindow("ppa-control")
	w2.SetContent(widget.NewButton("Open new window", func() {
		w3 := a.NewWindow("ppa-control new")
		w3.SetContent(widget.NewButton("Close", func() {
			w3.Close()
		}))
		w3.Show()
	}))
	w2.Resize(fyne.NewSize(100, 100))
	w2.Show()

	// XXX how do we deal with contexts and cancellation of goroutines?
	a.Run()
}

func updateTime(clock *widget.Label) {
	formatted := time.Now().Format("Time: 03:04:05")
	clock.SetText(formatted)
}
