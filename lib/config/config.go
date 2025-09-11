package config

import (
	"bitbox-editor/lib/logging"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	APP_NAME = "bitbox-editor"
)

var defaultConfig = []byte(`
# Default pre-generated config
[global]
loglevel = "DEBUG"
theme = "future dark"
`)

var log *zap.Logger

// Structure of config
type Config struct {
	Global struct {
		Loglevel string
		Theme    string
	}
}

func init() {
	log = logging.NewLogger("config")

	configdir, _ := os.UserConfigDir()
	appConfigDir := filepath.Join(configdir, APP_NAME)

	viper.AutomaticEnv()
	viper.SetEnvPrefix(APP_NAME)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(appConfigDir)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError

		if errors.As(err, &configFileNotFoundError) {
			log.Debug("Config file missing!")

			err = viper.ReadConfig(bytes.NewBuffer(defaultConfig))
			if err != nil {
				panic(err)
			}

			// ensure config directory exists; create it if not ...
			_, err := os.Open(fmt.Sprintf(appConfigDir))
			if err != nil {
				// Missing, create it.
				log.Debug(fmt.Sprintf("Creating directory: %s ...", appConfigDir))
				mkdirerr := os.Mkdir(appConfigDir, 0750)
				if mkdirerr != nil {
					panic(mkdirerr)
				}
			}

			err = viper.SafeWriteConfig()
			if err != nil {
				panic(err)
			}
			log.Debug("Wrote default config successfully.")

			if err := viper.ReadInConfig(); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		log.Debug("Config file found!")
	}
}
