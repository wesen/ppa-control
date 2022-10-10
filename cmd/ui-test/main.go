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
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"image/color"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
	"strings"
)

var (
	address     = flag.String("address", "127.0.0.1", "board address")
	port        = flag.Uint("port", 5001, "default port")
	addresses   = flag.String("addresses", "", "multiple board addresses")
	componentId = flag.Uint("component-id", 0xff, "default component ID (default: 0xff)")
)

func main() {
	flag.Parse()
	serverString := fmt.Sprintf("%s:%d", *address, *port)

	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx2 := errgroup.WithContext(ctx)

	receivedCh := make(chan client.ReceivedMessage)
	discoveryCh := make(chan discovery.PeerInformation)

	multiClient := client.NewMultiClient()
	if *addresses != "" {
		for _, addr := range strings.Split(*addresses, ",") {
			if addr == "" {
				continue
			}
			log.Info().Msgf("adding client %s", addr)
			_, err := multiClient.StartClient(ctx2, addr, *componentId)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to add client")
			}
		}
	} else {
		_, err := multiClient.StartClient(ctx2, serverString, *componentId)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add client")
		}
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

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
				multiClient.SendPresetRecallByPresetIndex(j)
				log.Info().Msg(fmt.Sprintf("Preset %d clicked", j+1))
			})
	}
	presetButtonContainer := container.New(layout.NewGridLayout(4), presetButtons...)

	var controlButtons []fyne.CanvasObject = make([]fyne.CanvasObject, 8)
	for i := 0; i < 8; i++ {
		j := i
		controlButtons[i] = widget.NewButton(fmt.Sprintf("Control %d", i+1),
			func() {
				log.Info().Msg(fmt.Sprintf("Control %d clicked", j))
			})
	}
	controlButtonContainer := container.New(layout.NewGridLayout(4), controlButtons...)

	var volumeButtons = make([]fyne.CanvasObject, 4)
	volumeButtons[0] = widget.NewButton("High", func() {
		log.Info().Msg("Volume HIGH")
	})
	volumeButtons[1] = widget.NewButton("Mid", func() {
		log.Info().Msg("Volume MID")
	})
	volumeButtons[2] = widget.NewButton("Low", func() {
		log.Info().Msg("Volume LOW")
	})
	volumeButtons[3] = widget.NewButton("Mute", func() {
		log.Info().Msg("Volume MUTE TOGGLE")
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
		log.Info().Float64("volume", value).Msg("Master Volume")
	}
	sliderContainer := container.New(layout.NewBorderLayout(
		//container.NewVBox(masterTitle, widget.NewSeparator()),
		masterTitle,
		nil, nil, nil),
		masterTitle, masterSlider)

	mainHBox := container.NewHBox(mainGridContainer, widget.NewSeparator(), sliderContainer)
	w.SetContent(mainHBox) // This is a text entry field
	w.Resize(fyne.NewSize(800, 800))

	fmt.Printf("Connecting to %s\n", serverString)

	w.SetOnClosed(func() {
		log.Info().Msg("Closing")
		cancel()
		log.Info().Msg("After cancel")
	})

	grp.Go(func() error {
		return multiClient.Run(ctx2, &receivedCh)
	})
	grp.Go(func() error {
		var interfaces []string
		return discovery.Discover(ctx2, discoveryCh, interfaces, uint16(*port))
	})

	grp.Go(func() error {
		for {
			select {
			case <-ctx2.Done():
				return ctx2.Err()
			case msg := <-receivedCh:
				if msg.Header != nil {
					log.Info().Str("from", msg.RemoteAddress.String()).
						Str("type", msg.Header.MessageType.String()).
						Str("client", msg.Client.Name()).
						Str("status", msg.Header.Status.String()).
						Msg("received message")
				} else {
					log.Debug().
						Str("from", msg.RemoteAddress.String()).
						Str("client", msg.Client.Name()).
						Msg("received unknown message")
				}
			case msg := <-discoveryCh:
				log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
				switch msg.(type) {
				case discovery.PeerDiscovered:
					log.Info().Str("addr", msg.GetAddress()).Msg("peer discovered")
					c, err := multiClient.StartClient(ctx, msg.GetAddress(), *componentId)
					if err != nil {
						log.Error().Err(err).Msg("failed to add client")
						return err
					}
					// send immediate ping
					c.SendPing()
				case discovery.PeerLost:
					log.Info().Str("addr", msg.GetAddress()).Msg("peer lost")
					err := multiClient.CancelClient(msg.GetAddress())
					if err != nil {
						log.Error().Err(err).Msg("failed to remove client")
						return err
					}
				}
			}
		}
	})

	// TODO this feels quite odd, let's learn more about fyne next
	go func() {
		log.Debug().Msg("Waiting for main loop")
		err := grp.Wait()
		log.Debug().Msg("Waited for main loop")

		if err != nil {
			log.Error().Err(err).Msg("Error in main loop")
		}
	}()

	w.ShowAndRun()
}
