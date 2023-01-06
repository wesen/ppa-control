package app

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/rs/zerolog/log"
	bucheron "github.com/wesen/bucheron/pkg"
	"golang.org/x/sync/errgroup"
)

type SettingsUI struct {
	uploadButton *widget.Button
	cancelButton *widget.Button
	closeButton  *widget.Button
	progressBar  *widget.ProgressBar
	label        *widget.Label
	popup        *widget.PopUp
	ui           *UI
	uploadCancel func()
}

func (s *SettingsUI) handleUpload() {
	progressChannel := make(chan bucheron.ProgressEvent)
	// TODO(manuel, 2023-01-06): this needs to be an app Context
	ctx, cancel := context.WithCancel(context.Background())
	errGroup, ctx2 := errgroup.WithContext(ctx)

	s.uploadButton.Disable()
	s.cancelButton.Enable()
	s.closeButton.Disable()
	s.uploadCancel = cancel

	go func() {
		errGroup.Go(func() error {
			for {
				select {
				case <-ctx2.Done():
					return ctx2.Err()
				case progress, ok := <-progressChannel:
					if !ok {
						return nil
					}
					s.label.SetText(progress.Step)
					s.progressBar.SetValue(progress.StepProgress)
					fmt.Printf("Progress: %s %f\n", progress.Step, progress.StepProgress)
				}
			}
		})

		errGroup.Go(func() error {
			return s.ui.app.UploadLogs(ctx2, progressChannel)
		})

		err := errGroup.Wait()
		if err != nil {
			log.Error().Err(err).Msg("Error uploading logs")
		} else {
			s.label.SetText("Upload complete")
		}

		s.uploadCancel = nil
		s.uploadButton.Enable()
		s.cancelButton.Disable()
		s.closeButton.Enable()
		s.progressBar.SetValue(0)
		s.progressBar.Hide()
	}()
}

func (s *SettingsUI) handleCancel() {
	fmt.Println("Cancel")
	if s.uploadCancel != nil {
		s.uploadCancel()
	}
}

func (s *SettingsUI) handleClose() {
	s.popup.Hide()
}

func (s *SettingsUI) Show() {
	s.popup.Show()
}

func (ui *UI) BuildSettingsUI() *SettingsUI {
	settings := &SettingsUI{
		ui: ui,
	}
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

	settings.label = widget.NewLabel("Upload a file to the PPA")

	// add the button and progress bar to the window
	settingsLayout := container.NewVBox(buttonLayout, settings.label, settings.progressBar)
	settings.popup = widget.NewModalPopUp(settingsLayout, ui.window.Canvas())

	return settings
}
