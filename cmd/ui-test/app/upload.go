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

type UploadLogsUI struct {
	uploadButton *widget.Button
	cancelButton *widget.Button
	closeButton  *widget.Button
	progressBar  *widget.ProgressBar
	label        *widget.Label
	popup        *widget.PopUp
	app          *App
	uploadCancel func()
}

func (s *UploadLogsUI) handleUpload() {
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
			return s.app.UploadLogs(ctx2, progressChannel)
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

func (s *UploadLogsUI) handleCancel() {
	fmt.Println("Cancel")
	if s.uploadCancel != nil {
		s.uploadCancel()
	}
}

func (s *UploadLogsUI) handleClose() {
	s.popup.Hide()
}

func (s *UploadLogsUI) Show() {
	s.popup.Show()
}

func (a *App) BuildUploadLogsUI() *UploadLogsUI {
	uploadLogsUI := &UploadLogsUI{
		app: a,
	}
	// create the progress bar
	uploadLogsUI.progressBar = widget.NewProgressBar()
	uploadLogsUI.progressBar.Hide()

	uploadLogsUI.uploadButton = widget.NewButton("Upload File", func() {
		fmt.Println("Uploading file...")
		uploadLogsUI.handleUpload()
	})

	uploadLogsUI.cancelButton = widget.NewButton("Cancel", func() {
		uploadLogsUI.handleCancel()
	})
	uploadLogsUI.cancelButton.Disable()

	uploadLogsUI.closeButton = widget.NewButton("Close", func() {
		uploadLogsUI.handleClose()
	})

	buttonLayout := container.NewHBox(
		uploadLogsUI.uploadButton,
		uploadLogsUI.cancelButton,
		uploadLogsUI.closeButton,
	)

	uploadLogsUI.label = widget.NewLabel("Upload a file to the PPA")

	// add the button and progress bar to the window
	settingsLayout := container.NewVBox(
		buttonLayout,
		uploadLogsUI.label,
		uploadLogsUI.progressBar,
	)

	uploadLogsUI.popup = widget.NewModalPopUp(settingsLayout, a.ui.window.Canvas())

	return uploadLogsUI
}
