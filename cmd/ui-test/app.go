package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"os/signal"
	"path"
	"ppa-control/cmd/ui-test/ui"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
	logger "ppa-control/lib/log"
)

type AppConfigFolders struct {
	configDirs   configdir.ConfigDir
	queryFolders []*configdir.Config
	configFile   string
}

func CreateAppConfigFolders() *AppConfigFolders {
	acf := &AppConfigFolders{}
	acf.configDirs = configdir.New("Hoffmann Audio", "ppa-control")
	acf.queryFolders = acf.configDirs.QueryFolders(configdir.Global)
	queryFolder := acf.configDirs.QueryFolderContainsFile("config.json")
	if queryFolder != nil {
		acf.configFile = path.Join(queryFolder.Path, "config.json")
	}

	return acf

}

type AppConfig struct {
	WithCaller bool   `json:"withCaller"`
	LogFormat  string `json:"logFormat"`
	LogLevel   string `json:"logLevel"`

	Addresses   []string `json:"addresses"`
	ComponentId uint     `json:"componentId"`
	Discover    bool     `json:"discover"`
	Port        uint     `json:"port"`
	Interfaces  []string `json:"interfaces"`

	saveConfig bool

	LogUploadAPI    string `json:"logUploadAPI"`
	LogUploadBucket string `json:"logUploadBucket"`
	LogUploadRegion string `json:"logUploadRegion"`

	configFolders *AppConfigFolders
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		WithCaller:  false,
		LogFormat:   "text",
		LogLevel:    "debug",
		Addresses:   []string{},
		Discover:    true,
		Port:        5001,
		Interfaces:  []string{},
		ComponentId: DEFAULT_COMPONENT_ID,

		LogUploadAPI:    "https://npyksyvjqj.execute-api.us-east-1.amazonaws.com/v1/",
		LogUploadBucket: "wesen-ppa-control-logs",
		LogUploadRegion: "us-east-1",
	}
}

func NewAppConfigFromFile(acf *AppConfigFolders) (*AppConfig, error) {
	if acf.configFile != "" {
		log.Info().Str("path", acf.configFile).Msg("Found config file")
		f, err := os.Open(acf.configFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		config := NewAppConfig()
		err = json.Unmarshal(data, config)
		if err != nil {
			return nil, err
		}
		return config, nil
	} else {
		return nil, nil
	}
}

func AddAppConfigFlags(cmd *cobra.Command) {
	defaultConfig := CreateDefaultAppConfig()

	cmd.PersistentFlags().String("log-level", defaultConfig.LogLevel, "Log level")
	cmd.PersistentFlags().String("log-format", defaultConfig.LogFormat, "Log format (json, text)")
	cmd.PersistentFlags().Bool("with-caller", defaultConfig.WithCaller, "Log caller")
	cmd.PersistentFlags().StringArrayP(
		"addresses", "a", defaultConfig.Addresses,
		"Addresses to ping, comma separated",
	)
	// disable discovery by default when pinging
	cmd.PersistentFlags().BoolP(
		"discover", "d", defaultConfig.Discover,
		"Send broadcast discovery messages",
	)

	cmd.PersistentFlags().StringArray("interfaces", defaultConfig.Interfaces, "Interfaces to use for discovery")

	cmd.PersistentFlags().UintP(
		"componentId", "c", defaultConfig.ComponentId,
		"Component ID to use for devices")

	cmd.PersistentFlags().UintP("port", "p", defaultConfig.Port, "Port to ping on")

	cmd.PersistentFlags().String(
		"api",
		defaultConfig.LogUploadAPI,
		"URL of the bucheron API")
	cmd.PersistentFlags().String("bucket", defaultConfig.LogUploadBucket, "S3 bucket to upload to")
	cmd.PersistentFlags().String("region", defaultConfig.LogUploadRegion, "Region of the S3 bucket")

	cmd.Flags().Bool("save-config", defaultConfig.saveConfig, "Save config to file")
}

func CreateDefaultAppConfig() *AppConfig {
	acf := CreateAppConfigFolders()
	defaultConfig, err := NewAppConfigFromFile(acf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config file")
	}
	if defaultConfig == nil {
		defaultConfig = NewAppConfig()
	}
	defaultConfig.configFolders = acf
	return defaultConfig
}

func (ac *AppConfig) SaveToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(ac)
	if err != nil {
		return err
	}

	return nil
}

func NewAppConfigFromCommand(cmd *cobra.Command) *AppConfig {
	config := CreateDefaultAppConfig()

	withCaller, _ := cmd.Flags().GetBool("with-caller")
	logFormat, _ := cmd.Flags().GetString("log-format")
	logLevel, _ := cmd.Flags().GetString("log-level")

	addresses, _ := cmd.Flags().GetStringArray("addresses")
	componentId, _ := cmd.Flags().GetUint("componentId")
	discover, _ := cmd.Flags().GetBool("discover")
	port, _ := cmd.Flags().GetUint("port")
	interfaces, _ := cmd.Flags().GetStringArray("interfaces")

	saveConfig, _ := cmd.Flags().GetBool("save-config")

	config.WithCaller = withCaller
	config.LogFormat = logFormat
	config.LogLevel = logLevel
	config.Addresses = addresses
	config.ComponentId = componentId
	config.Discover = discover
	config.Port = port
	config.Interfaces = interfaces
	config.saveConfig = saveConfig

	return config
}

type App struct {
	Config  *AppConfig
	LogsDir string
}

func (a *App) initLogger() error {
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

func (a *App) UploadLogs(ctx context.Context) {
	//credentials, err := bucheron.GetUploadCredentials(ctx, a.Config.logUploadAPI)
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

	if a.Config.saveConfig {
		if a.Config.configFolders.configFile != "" {
			configFilePath := a.Config.configFolders.configFile
			log.Info().Str("path", configFilePath).Msg("Found config file")
			err := a.Config.SaveToFile(configFilePath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to save config file to %s", configFilePath)
			}
		}
	}

	multiClient := client.NewMultiClient("ui")
	for _, addr := range a.Config.Addresses {
		if addr == "" {
			continue
		}
		log.Info().Msgf("adding pkg %s", addr)
		_, err := multiClient.AddClient(ctx2, addr, "", a.Config.ComponentId)
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

	ui_ := ui.BuildUI(multiClient, cancel)
	ui_.Log("ppa-control started, waiting for devices...")

	grp.Go(func() error {
		return multiClient.Run(ctx2, &receivedCh)
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
					c, err := multiClient.AddClient(ctx, msg.GetAddress(), msg.GetInterface(), a.Config.ComponentId)
					if err != nil {
						log.Error().Err(err).Msg("failed to add pkg")
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
					err := multiClient.CancelClient(msg.GetAddress())
					if err != nil {
						log.Error().Err(err).Msg("failed to remove pkg")
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

	ui_.Run()

}
