package app

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/rs/zerolog/log"
	"os"
)

type presetUndoStack struct {
	undoStack [][]*Preset
	redoStack [][]*Preset
}

func newPresetUndoStack() *presetUndoStack {
	return &presetUndoStack{
		undoStack: make([][]*Preset, 0),
		redoStack: make([][]*Preset, 0),
	}
}

func copyPresets(presets []*Preset) []*Preset {
	copyPresets := make([]*Preset, len(presets))
	copy(copyPresets, presets)
	return copyPresets
}

func (p *presetUndoStack) push(presets []*Preset) {
	p.undoStack = append(p.undoStack, copyPresets(presets))
	p.redoStack = make([][]*Preset, 0)
}

func (p *presetUndoStack) undo(current []*Preset) []*Preset {
	if len(p.undoStack) == 0 {
		return nil
	}
	ret := p.undoStack[len(p.undoStack)-1]
	p.undoStack = p.undoStack[:len(p.undoStack)-1]
	p.redoStack = append(p.redoStack, copyPresets(current))
	return ret
}

func (p *presetUndoStack) redo(current []*Preset) []*Preset {
	if len(p.redoStack) == 0 {
		return nil
	}
	ret := p.redoStack[len(p.redoStack)-1]
	p.redoStack = p.redoStack[:len(p.redoStack)-1]
	p.undoStack = append(p.undoStack, copyPresets(current))
	return ret
}

func (a *App) resetPresets() {
	a.Presets = []*Preset{}
	// create 16 presets per default
	for i := 0; i < 4; i++ {
		a.Presets = append(a.Presets, &Preset{
			Name:        fmt.Sprintf("Preset %d", i+1),
			PresetIndex: i,
		})
	}
}

func (a *App) loadPresets() {
	presetFile := a.Config.GetConfigFilePath("presets.json")
	if _, err := os.Stat(presetFile); err == nil {
		// decode presetFile as json to app.Presets
		f, err := os.Open(presetFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open presets file")
		} else {
			err = json.NewDecoder(f).Decode(&a.Presets)
			if err != nil {
				log.Error().Err(err).Msg("Failed to decode presets file")
			}
		}
	}
}

func (a *App) updatePresets() {
	presetButtons := a.createPresetButtons()
	a.presetButtonContainer.RemoveAll()
	for _, button := range presetButtons {
		a.presetButtonContainer.Add(button)
	}
	a.presetButtonContainer.Refresh()

	presetFile := a.Config.GetConfigFilePath("presets.json")
	f, err := os.Create(presetFile)
	if err != nil {
		log.Error().Err(err).Msg("Error creating presets file")
		return
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(a.Presets)
	if err != nil {
		log.Error().Err(err).Msg("Error encoding presets")
		return
	}
}

func (a *App) ShowPresetEditor() {
	if a.ui.presetEditorWindow != nil {
		a.ui.presetEditorWindow.Close()
	}
	a.ui.presetEditorWindow = a.ui.fyneApp.NewWindow("Preset Editor")

	presetButtonContainer := container.New(layout.NewGridLayout(4))

	updatePresetEditorButtons := func() {
		presetButtonContainer.RemoveAll()
		for _, preset := range a.Presets {
			presetButtonContainer.Add(widget.NewButton(preset.Name, func() {
				log.Info().Str("presetName", preset.Name).Msg("Preset edit")
			}))
		}
	}

	updatePresetEditorButtons()

	presetUndoStack := newPresetUndoStack()
	var undoButton *widget.Button
	var redoButton *widget.Button

	refreshUndoRedoButtons := func() {
		if len(presetUndoStack.undoStack) == 0 {
			undoButton.Disable()
		} else {
			undoButton.Enable()
		}
		if len(presetUndoStack.redoStack) == 0 {
			redoButton.Disable()
		} else {
			redoButton.Enable()
		}
	}
	pushUndo := func(presets []*Preset) {
		presetUndoStack.push(presets)
		refreshUndoRedoButtons()
	}

	undoButton = widget.NewButton("Undo", func() {
		a.Presets = presetUndoStack.undo(a.Presets)
		updatePresetEditorButtons()
		refreshUndoRedoButtons()
	})
	redoButton = widget.NewButton("Redo", func() {
		a.Presets = presetUndoStack.redo(a.Presets)
		updatePresetEditorButtons()
		refreshUndoRedoButtons()
	})
	redoButton.Disable()

	settingsButtonContainer := container.NewHBox(
		widget.NewButton("Reset Presets", func() {
			pushUndo(a.Presets)

			a.resetPresets()
			updatePresetEditorButtons()
		}),
		widget.NewButton("Reload Presets", func() {
			pushUndo(a.Presets)

			a.loadPresets()
			updatePresetEditorButtons()
		}),
		widget.NewButton("Save Presets", func() {
			a.updatePresets()
		}),
		widget.NewButton("Add Preset", func() {
			pushUndo(a.Presets)

			a.Presets = append(a.Presets, &Preset{
				Name:        "New Preset",
				PresetIndex: len(a.Presets),
			})
			updatePresetEditorButtons()
		}),
	)

	mainGridContainer := container.NewVBox(
		settingsButtonContainer,
		container.NewHBox(
			undoButton,
			redoButton,
		),
		widget.NewSeparator(),
		presetButtonContainer,
		widget.NewSeparator(),
	)

	a.ui.presetEditorWindow.SetContent(mainGridContainer)
	a.ui.presetEditorWindow.SetOnClosed(func() {
		a.loadPresets()
		a.ui.presetEditorWindow = nil
	})
	a.ui.presetEditorWindow.Show()
}
