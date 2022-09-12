package main

import (
	"context"
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/sync/errgroup"
	"image/color"
	"log"
	"ppa-control/lib/client"
	"ppa-control/lib/server"
	"time"
)

var (
	address        = flag.String("address", "127.0.0.1", "server address")
	port           = flag.Uint("port", 5001, "server port")
	runServer      = flag.Bool("run-server", false, "Run as server too")
	presetPosition = flag.Int("position", 1, "preset")
	componentId    = flag.Int("component-id", 0xff, "component ID (default: 0xff)")
)

func main() {
	flag.Parse()
	serverString := fmt.Sprintf("%s:%d", *address, *port)
	c := client.NewClient(serverString, *componentId)

	a := app.New()
	w := a.NewWindow("ppa-control")

	serverConsole := widget.NewLabel("ServerConsole\nServerConsole\nServerConsole")
	serverScrollContainer := container.NewVScroll(serverConsole)
	serverScrollContainer.SetMinSize(fyne.NewSize(600, 150))

	clientConsole := canvas.NewText("Hello There", color.White)
	clientScrollContainer := container.NewVScroll(clientConsole)
	clientScrollContainer.SetMinSize(fyne.NewSize(600, 150))

	var presetButtons = make([]fyne.CanvasObject, 8)
	for i := 0; i < 8; i++ {
		j := i
		presetButtons[i] = widget.NewButton(fmt.Sprintf("Preset %d", i+1),
			func() {
				c.SendPresetRecallByPresetIndex(j)
				log.Println(fmt.Sprintf("Preset %d clicked", j+1))
			})
	}
	presetButtonContainer := container.New(layout.NewGridLayout(4), presetButtons...)

	var controlButtons []fyne.CanvasObject = make([]fyne.CanvasObject, 8)
	for i := 0; i < 8; i++ {
		j := i
		controlButtons[i] = widget.NewButton(fmt.Sprintf("Control %d", i+1),
			func() {
				log.Println(fmt.Sprintf("Control %d clicked", j))
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
		serverScrollContainer,
		widget.NewSeparator(),
		clientScrollContainer,
		widget.NewSeparator(),
		presetButtonContainer,
		widget.NewSeparator(),
		volumeContainer,
		widget.NewSeparator(),
		controlButtonContainer)

	serverConsole.Text = "foobar\nFoo Foo\nblablabla\nfunkfunk\nyo"
	serverConsole.Refresh()

	masterTitle := canvas.NewText("master", color.White)
	masterSlider := widget.NewSlider(0, 10)
	masterSlider.Step = 0.01
	masterSlider.Orientation = widget.Vertical
	masterSlider.MinSize()
	masterSlider.OnChanged = func(value float64) {
		log.Println(fmt.Sprintf("Master Volume: %f", value))
	}
	sliderContainer := container.New(layout.NewBorderLayout(
		//container.NewVBox(masterTitle, widget.NewSeparator()),
		masterTitle,
		nil, nil, nil),
		masterTitle, masterSlider)

	mainHBox := container.NewHBox(mainGridContainer, widget.NewSeparator(), sliderContainer)
	w.SetContent(mainHBox) // This is a text entry field
	w.Resize(fyne.NewSize(800, 800))

	ctx, cancel := context.WithCancel(context.Background())
	fmt.Printf("Connecting to %s\n", serverString)

	if *runServer {
		fmt.Printf("Starting test server")
		go server.RunServer(ctx, serverString)
		time.Sleep(1 * time.Second)
	}

	w.SetOnClosed(func() {
		log.Println("Closing")
		cancel()
		log.Println("After cancel")
	})

	grp, ctx2 := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return c.Run(ctx2)
	})
	go func() {
		log.Println("Waiting for main loop")
		err := grp.Wait()
		log.Println("Waited for main loop")
		if err != nil {
			log.Printf("Error in main loop: %v\n", err)
		}
	}()

	w.ShowAndRun()
}
