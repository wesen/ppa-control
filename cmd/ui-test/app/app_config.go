package app

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
)

const DEFAULT_COMPONENT_ID = 0xFF

type AppConfigFolders struct {
	configDirs   configdir.ConfigDir
	queryFolders []*configdir.Config
	ConfigFile   string
}

func CreateAppConfigFolders() *AppConfigFolders {
	acf := &AppConfigFolders{}
	acf.configDirs = configdir.New("Hoffmann Audio", "ppa-control")
	acf.queryFolders = acf.configDirs.QueryFolders(configdir.Global)
	queryFolder := acf.configDirs.QueryFolderContainsFile("config.json")
	if queryFolder != nil {
		acf.ConfigFile = path.Join(queryFolder.Path, "config.json")
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

	SaveConfig bool `json:"-"`

	LogUploadAPI    string `json:"logUploadAPI"`
	LogUploadBucket string `json:"logUploadBucket"`
	LogUploadRegion string `json:"logUploadRegion"`

	ConfigFolders *AppConfigFolders `json:"-"`
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
	if acf.ConfigFile != "" {
		log.Info().Str("path", acf.ConfigFile).Msg("Found config file")
		f, err := os.Open(acf.ConfigFile)
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

	cmd.Flags().Bool("save-config", defaultConfig.SaveConfig, "Save config to file")
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
	defaultConfig.ConfigFolders = acf
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
	config.SaveConfig = saveConfig

	return config
}
