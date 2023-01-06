package app

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shibukawa/configdir"
	bucheron "github.com/wesen/bucheron/pkg"
	"golang.org/x/sync/errgroup"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
	logger "ppa-control/lib/log"
	"syscall"
	"time"
)

type App struct {
	Config      *AppConfig
	LogsDir     string
	MultiClient *client.MultiClient
}

func (a *App) InitLogger() error {
	fmt.Println("withCaller", a.Config.WithCaller)
	logger.InitializeLogger(a.Config.WithCaller)
	// default is json
	var logWriter io.Writer
	if a.Config.LogFormat == "text" {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	configDirs := configdir.New("Hoffmann Audio", "ppa-control")
	cacheFolder := configDirs.QueryCacheFolder()
	a.LogsDir = path.Join(cacheFolder.Path, "logs")
	log.Info().Msgf("Writing logs to %s", a.LogsDir)

	// ensure logs dir exists
	err := os.MkdirAll(a.LogsDir, 0755)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create logs dir")
	} else {
		logWriter = io.MultiWriter(
			logWriter,
			zerolog.ConsoleWriter{
				NoColor: true,
				Out: &lumberjack.Logger{
					Filename:   path.Join(a.LogsDir, "ppa-control.log"),
					MaxSize:    10, // megabytes
					MaxBackups: 3,
					MaxAge:     28,    //days
					Compress:   false, // disabled by default
				},
			})
	}

	log.Logger = log.Output(logWriter)

	switch a.Config.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	}

	return nil
}

func (a *App) UploadLogs(ctx context.Context, progressCh chan bucheron.ProgressEvent) error {
	defer close(progressCh)

	log.Info().Msg("Uploading logs")
	progressCh <- bucheron.ProgressEvent{
		StepProgress: 0,
		Step:         "Getting upload credentials",
		IsError:      false,
	}
	credentials, err := bucheron.GetUploadCredentials(ctx, a.Config.LogUploadAPI)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get upload credentials")
		progressCh <- bucheron.ProgressEvent{
			StepProgress: 0,
			Step:         "Failed to get upload credentials",
			IsError:      true,
		}
		return err
	}

	settings := &bucheron.BucketSettings{
		Region:      a.Config.LogUploadRegion,
		Bucket:      a.Config.LogUploadBucket,
		Credentials: credentials,
	}

	files := []string{}

	// gather all files in a.LogsDir
	err = filepath.Walk(a.LogsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to walk logs dir")
	}

	if a.Config.ConfigFolders.ConfigFile != "" {
		files = append(files, a.Config.ConfigFolders.ConfigFile)
	}

	data := &bucheron.UploadData{
		Files:    files,
		Comment:  fmt.Sprintf("ppa-control upload at %s", time.Now().Format(time.RFC3339)),
		Metadata: nil,
	}

	return bucheron.UploadLogs(ctx, settings, data, progressCh)
}

func (a *App) Run() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer func() {
		log.Debug().Msg("Cancelling context")
		cancel()
	}()

	grp, ctx2 := errgroup.WithContext(ctx)

	receivedCh := make(chan client.ReceivedMessage)
	discoveryCh := make(chan discovery.PeerInformation)

	if a.Config.SaveConfig {
		if a.Config.ConfigFolders.ConfigFile != "" {
			configFilePath := a.Config.ConfigFolders.ConfigFile
			log.Info().Str("path", configFilePath).Msg("Found config file")
			err := a.Config.SaveToFile(configFilePath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to save config file to %s", configFilePath)
			}
		}
	}

	a.MultiClient = client.NewMultiClient("ui")
	for _, addr := range a.Config.Addresses {
		if addr == "" {
			continue
		}
		log.Info().Msgf("adding pkg %s", addr)
		_, err := a.MultiClient.AddClient(ctx2, addr, "", a.Config.ComponentId)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add pkg")
		}
	}

	if a.Config.Discover {
		grp.Go(func() error {
			return discovery.Discover(ctx, discoveryCh, a.Config.Addresses, uint16(a.Config.Port))
		})
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ui_ := BuildUI(a, cancel)
	ui_.Log("ppa-control started, waiting for devices...")

	grp.Go(func() error {
		return a.MultiClient.Run(ctx2, &receivedCh)
	})

	grp.Go(func() error {
		return bucheron.CancelOnSignal(ctx, syscall.SIGINT, cancel)
	})

	// This is the peer handling
	grp.Go(func() error {
		for {
			select {
			case <-ctx2.Done():
				return ctx2.Err()
			case msg := <-receivedCh:
				if msg.Header != nil {
					log.Info().Str("from", msg.RemoteAddress.String()).
						Str("type", msg.Header.MessageType.String()).
						Str("pkg", msg.Client.Name()).
						Str("status", msg.Header.Status.String()).
						Msg("received message")
				} else {
					log.Debug().
						Str("from", msg.RemoteAddress.String()).
						Str("pkg", msg.Client.Name()).
						Msg("received unknown message")
				}
			case msg := <-discoveryCh:
				log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
				switch msg.(type) {
				case discovery.PeerDiscovered:
					log.Info().
						Str("addr", msg.GetAddress()).
						Str("iface", msg.GetInterface()).
						Msg("peer discovered")
					ui_.Log("Peer discovered: " + msg.GetAddress())
					c, err := a.MultiClient.AddClient(ctx, msg.GetAddress(), msg.GetInterface(), a.Config.ComponentId)
					if err != nil {
						log.Error().Err(err).Msg("failed to add pkg")
						cancel()
						return err
					}
					// send immediate ping
					c.SendPing()
				case discovery.PeerLost:
					log.Info().
						Str("addr", msg.GetAddress()).
						Str("iface", msg.GetInterface()).
						Msg("peer lost")
					ui_.Log("Peer lost: " + msg.GetAddress())
					err := a.MultiClient.CancelClient(msg.GetAddress())
					if err != nil {
						log.Error().Err(err).Msg("failed to remove pkg")
						cancel()
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

		// TODO(manuel, 2023-01-06) We need an error dialog here
		ui_.Close()
	}()

	ui_.Run()
}
