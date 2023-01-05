package ui

import (
	"fmt"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SettingsUI struct {
	uploadButton *widget.Button
	cancelButton *widget.Button
	closeButton  *widget.Button
	progressBar  *widget.ProgressBar
	popup        *widget.PopUp
}

func (s *SettingsUI) handleUpload() {
	fmt.Println("Upload")
	s.uploadButton.Disable()
	s.cancelButton.Enable()
	s.progressBar.SetValue(0.5)
	s.progressBar.Show()
}

func (s *SettingsUI) handleCancel() {
	fmt.Println("Cancel")
	s.uploadButton.Enable()
	s.cancelButton.Disable()
	s.progressBar.SetValue(0)
	s.progressBar.Hide()
}

func (s *SettingsUI) handleClose() {
	s.popup.Hide()
}

func (s *SettingsUI) Show() {
	s.popup.Show()
}

func (ui *UI) BuildSettingsUI() *SettingsUI {
	settings := &SettingsUI{}
	// create the progress bar
	settings.progressBar = widget.NewProgressBar()
	settings.progressBar.Hide()

	settings.uploadButton = widget.NewButton("Upload File", func() {
		fmt.Println("Uploading file...")
		settings.handleUpload()
	})

	settings.cancelButton = widget.NewButton("Cancel", func() {
		settings.handleCancel()
	})
	settings.cancelButton.Disable()

	settings.closeButton = widget.NewButton("Close", func() {
		settings.handleClose()
	})

	buttonLayout := container.NewHBox(
		settings.uploadButton,
		settings.cancelButton,
		settings.closeButton,
	)

	// add the button and progress bar to the window
	settingsLayout := container.NewVBox(buttonLayout, settings.progressBar)
	settings.popup = widget.NewModalPopUp(settingsLayout, ui.window.Canvas())

	return settings
}
