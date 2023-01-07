package app

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/rs/zerolog/log"
)

type UI struct {
	window             fyne.Window
	console            binding.String
	fyneApp            fyne.App
	presetEditorWindow fyne.Window
}

func (ui *UI) Log(line string) {
	s, err := ui.console.Get()
	if err != nil {
		log.Error().Err(err).Msg("Error getting console string")
	}
	err = ui.console.Set(s + line + "\n")
	if err != nil {
		log.Error().Err(err).Msg("Error setting console string")
	}
}

func (ui *UI) Run() {
	ui.window.ShowAndRun()
}

func (ui *UI) Close() {
	ui.window.Close()
}

func (a *App) createPresetButtons() []fyne.CanvasObject {
	ret := make([]fyne.CanvasObject, 0)

	for _, preset := range a.Presets {
		ret = append(ret, widget.NewButton(preset.Name, func() {
			a.switchPreset(preset)
		}))
	}

	return ret
}

func (a *App) switchPreset(preset *Preset) {
	log.Info().Int("presetIndex", preset.PresetIndex).
		Str("presetName", preset.Name).
		Msg("Preset clicked")
	a.MultiClient.SendPresetRecallByPresetIndex(preset.PresetIndex)
}

// TODO(manuel, 2023-01-06) this whole UI struct containing the app containing the UI is not great
func (a *App) BuildUI(cancel context.CancelFunc) *UI {
	fyneApp := app.New()
	ui := &UI{
		fyneApp: fyneApp,
		window:  fyneApp.NewWindow("PPA Control"),
		console: binding.NewString(),
	}
	_ = ui.console.Set("")

	serverConsole := widget.NewLabelWithData(ui.console)
	serverScrollContainer := container.NewVScroll(serverConsole)
	serverScrollContainer.SetMinSize(fyne.NewSize(600, 80))

	//clientConsole := canvas.NewText("Hello There", color.White)
	//clientScrollContainer := container.NewVScroll(clientConsole)
	//clientScrollContainer.SetMinSize(fyne.NewSize(600, 150))

	uploadLogs := widget.NewButton("Upload Logs", func() {
		uploadLogsPopup := a.BuildUploadLogsUI()
		uploadLogsPopup.Show()
	})

	settingsButtonContainer := container.NewHBox(
		uploadLogs,
		widget.NewButton("Edit Presets", func() {
			a.ShowPresetEditor()
		}),
	)

	//presetCount := 16
	//var presetButtons = make([]fyne.CanvasObject, presetCount)
	//for i := 0; i < presetCount; i++ {
	//	j := i
	//	presetButtons[i] = widget.NewButton(fmt.Sprintf("Preset %d", i+1),
	//		func() {
	//			a.MultiClient.SendPresetRecallByPresetIndex(j)
	//		})
	//}
	presetButtons := a.createPresetButtons()
	a.presetButtonContainer = container.New(layout.NewGridLayout(4), presetButtons...)

	//var controlButtons []fyne.CanvasObject = make([]fyne.CanvasObject, 8)
	//for i := 0; i < 8; i++ {
	//	j := i
	//	controlButtons[i] = widget.NewButton(fmt.Sprintf("Control %d", i+1),
	//		func() {
	//			log.Info().Msg(fmt.Sprintf("Control %d clicked", j))
	//		})
	//}
	//controlButtonContainer := container.New(layout.NewGridLayout(4), controlButtons...)

	//var volumeButtons = make([]fyne.CanvasObject, 4)
	//volumeButtons[0] = widget.NewButton("High", func() {
	//	log.Info().Msg("Volume HIGH")
	//})
	//volumeButtons[1] = widget.NewButton("Mid", func() {
	//	log.Info().Msg("Volume MID")
	//})
	//volumeButtons[2] = widget.NewButton("Low", func() {
	//	log.Info().Msg("Volume LOW")
	//})
	//volumeButtons[3] = widget.NewButton("Mute", func() {
	//	log.Info().Msg("Volume MUTE TOGGLE")
	//})
	//volumeContainer := container.New(layout.NewVBoxLayout(), volumeButtons...)
	//
	// we also need title fields
	// 4 buttons for the fixed volumes
	// fyneApp side bar and the volume

	mainGridContainer := container.NewVBox(
		a.presetButtonContainer,
		widget.NewSeparator(),
		settingsButtonContainer,
		//widget.NewSeparator(),
		//clientScrollContainer,
		widget.NewSeparator(),
		serverScrollContainer,
		//widget.NewSeparator(),
		//volumeContainer,
		//widget.NewSeparator(),
		//controlButtonContainer,
	)

	//masterTitle := canvas.NewText("master", color.White)
	//masterSlider := widget.NewSlider(0, 10)
	//masterSlider.Step = 0.01
	//masterSlider.Orientation = widget.Vertical
	//masterSlider.MinSize()
	//masterSlider.OnChanged = func(value float64) {
	//	log.Info().Float64("volume", value).Msg("Master Volume")
	//}
	//sliderContainer := container.New(layout.NewBorderLayout(
	//	//container.NewVBox(masterTitle, widget.NewSeparator()),
	//	masterTitle,
	//	nil, nil, nil),
	//	masterTitle, masterSlider)

	//mainHBox := container.NewHBox(
	//	mainGridContainer,
	//	widget.NewSeparator(),
	//	//sliderContainer,
	//)
	ui.window.SetContent(mainGridContainer) // This is fyneApp text entry field
	//ui.window.Resize(fyne.NewSize(800, 800))

	ui.window.SetOnClosed(func() {
		log.Info().Msg("Closing")
		cancel()
		log.Info().Msg("After cancel")
	})

	a.ui = ui
	return ui
}
