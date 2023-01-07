package app

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
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

type editPresetUI struct {
	entry            *widget.Entry
	indexList        *widget.Select
	deleteButton     *widget.Button
	updateButton     *widget.Button
	container        *fyne.Container
	selectedPreset   *Preset
	selectedPosition int
}

func newEditPresetUI(
	onDelete func(int, *Preset),
	onUpdate func(int, *Preset),
) *editPresetUI {
	index32List := make([]string, 32)
	for i := 0; i < 32; i++ {
		index32List[i] = fmt.Sprintf("%d", i+1)
	}

	ui := &editPresetUI{
		entry:        widget.NewEntry(),
		indexList:    widget.NewSelect(index32List, nil),
		deleteButton: widget.NewButton("Delete", nil),
		updateButton: widget.NewButton("Update", nil),
	}
	ui.container = container.NewVBox(
		ui.entry,
		ui.indexList,
		container.NewHBox(
			ui.deleteButton,
			ui.updateButton,
		),
	)
	ui.container.Hide()

	ui.deleteButton.OnTapped = func() {
		onDelete(ui.selectedPosition, ui.selectedPreset)
		ui.container.Hide()
	}
	ui.updateButton.OnTapped = func() {
		ui.selectedPreset.Name = ui.entry.Text
		indexPlusOne := 1
		_, _ = fmt.Sscanf(ui.indexList.Selected, "%d", &indexPlusOne)
		ui.selectedPreset.PresetIndex = indexPlusOne - 1

		onUpdate(ui.selectedPosition, ui.selectedPreset)
	}

	return ui
}

func (epui *editPresetUI) selectPreset(position int, preset *Preset) {
	epui.selectedPreset = &Preset{
		Name:        preset.Name,
		PresetIndex: preset.PresetIndex,
	}
	epui.selectedPosition = position
	epui.entry.SetText(preset.Name)
	epui.indexList.SetSelected(fmt.Sprintf("%d", preset.PresetIndex+1))
	epui.container.Show()
}

// TODO(manuel, 2023-01-06) we really need to copy the list of presets here, not use a.Presets
// or maybe we should...
func (a *App) ShowPresetEditor() {
	editedPresets := copyPresets(a.Presets)

	if a.ui.presetEditorWindow != nil {
		a.ui.presetEditorWindow.Close()
	}
	a.ui.presetEditorWindow = a.ui.fyneApp.NewWindow("Preset Editor")

	var addPresetButton *widget.Button

	presetUndoStack := newPresetUndoStack()
	var undoButton *widget.Button
	var redoButton *widget.Button
	var applyButton *widget.Button

	refreshUndoRedoButtons := func() {
		if len(presetUndoStack.undoStack) == 0 {
			undoButton.Disable()
			applyButton.Disable()
		} else {
			undoButton.Enable()
			applyButton.Enable()
		}
		if len(presetUndoStack.redoStack) == 0 {
			redoButton.Disable()
		} else {
			redoButton.Enable()
		}
	}
	pushUndo := func(presets []*Preset) {
		log.Debug().Msg("Pushing undo")
		presetUndoStack.push(presets)
		refreshUndoRedoButtons()
	}

	presetButtonContainer := container.New(layout.NewGridLayout(4))
	var editUI *editPresetUI
	updatePresetEditorButtons := func() {
		presetButtonContainer.RemoveAll()
		for idx, preset := range editedPresets {
			preset_ := preset
			idx_ := idx
			presetButtonContainer.Add(widget.NewButton(
				fmt.Sprintf("%s (%d)", preset_.Name, preset_.PresetIndex+1),
				func() {
					log.Info().Str("presetName", preset_.Name).Msg("Preset edit")
					editUI.selectPreset(idx_, preset_)
				}))
		}
		if len(editedPresets) >= 32 {
			addPresetButton.Disabled()
		}
	}

	editUI = newEditPresetUI(
		func(updatedPosition int, preset *Preset) {
			log.Info().Int("position", updatedPosition).
				Str("preset", preset.Name).Msg("Deleting preset")

			pushUndo(editedPresets)
			updatedPresets := append(editedPresets[:updatedPosition], editedPresets[updatedPosition+1:]...)
			editedPresets = updatedPresets
			updatePresetEditorButtons()
		},
		func(updatedPosition int, preset *Preset) {
			log.Info().Int("position", updatedPosition).
				Str("preset", preset.Name).Msg("Updating preset")
			pushUndo(editedPresets)
			copiedPresets := copyPresets(editedPresets)
			copiedPresets[updatedPosition] = preset
			editedPresets = copiedPresets
			updatePresetEditorButtons()
		})

	undoButton = widget.NewButton("Undo", func() {
		editedPresets = presetUndoStack.undo(editedPresets)
		updatePresetEditorButtons()
		refreshUndoRedoButtons()
	})
	redoButton = widget.NewButton("Redo", func() {
		editedPresets = presetUndoStack.redo(editedPresets)
		updatePresetEditorButtons()
		refreshUndoRedoButtons()
	})
	redoButton.Disable()

	addPresetButton = widget.NewButton("Add Preset", func() {
		pushUndo(editedPresets)

		editedPresets = append(editedPresets, &Preset{
			Name:        "New Preset",
			PresetIndex: len(editedPresets),
		})
		updatePresetEditorButtons()
	})
	settingsButtonContainer := container.NewHBox(
		widget.NewButton("Reset Presets", func() {
			pushUndo(editedPresets)
			a.resetPresets()
			editedPresets = copyPresets(a.Presets)
			updatePresetEditorButtons()
		}),
		widget.NewButton("Reload Presets", func() {
			pushUndo(editedPresets)

			a.loadPresets()
			editedPresets = copyPresets(a.Presets)
			updatePresetEditorButtons()
		}),
		addPresetButton,
	)

	applyButton = widget.NewButton("Apply", func() {
		a.Presets = editedPresets
		a.updatePresets()
		a.ui.presetEditorWindow.Close()
	})
	mainGridContainer := container.NewVBox(
		settingsButtonContainer,
		container.NewHBox(
			undoButton,
			redoButton,
		),
		widget.NewSeparator(),
		presetButtonContainer,
		widget.NewSeparator(),
		editUI.container,
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewButton("Cancel", func() {
				a.ui.presetEditorWindow.Close()
			}),
			applyButton,
		),
	)

	a.ui.presetEditorWindow.SetCloseIntercept(func() {
		// can't close, only through cancel or apply
	})

	updatePresetEditorButtons()
	refreshUndoRedoButtons()

	a.ui.presetEditorWindow.SetContent(mainGridContainer)
	a.ui.presetEditorWindow.SetOnClosed(func() {
		a.loadPresets()
		a.ui.presetEditorWindow = nil
	})
	a.ui.presetEditorWindow.Show()
}
