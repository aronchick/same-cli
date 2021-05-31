package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// // Config stores all configuration of the application.
// // The values are read by viper from a config file or environment variable.
// type Config struct {
// 	MetadataUri    string `mapstructure:"METADATA_URI,omitempty"`
// 	ActivePipeline string `mapstructure:"ACTIVE_PIPELINE"`
// }

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(configPath string) (err error) {
	log.Trace("- In utils.LoadConfig")
	viper.AutomaticEnv() // read in environment variables that match

	viper.SetConfigFile(configPath)
	log.Tracef("Setting config file to: %v\n", configPath)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Tracef("Using config file: %v", viper.ConfigFileUsed())
		return nil
	} else {
		_, notFound := err.(viper.ConfigFileNotFoundError)
		_, badPath := err.(*fs.PathError)

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("Failed to write to temporary file: %v", err)
		}
		defaultSameConfigDir := path.Join(home, ".same")
		_, noSameConfigDirFound := os.Stat(defaultSameConfigDir)
		if noSameConfigDirFound != nil {
			// There was no .same directory, so we'll eat the BadPath error and convert
			// it to a notFound error
			badPath = false
			notFound = true
		}

		_, err = os.Stat(path.Join(defaultSameConfigDir, "config.yaml"))
		if err != nil {
			// The directory is present, but the config file is missing
			// fixing the errors
			badPath = false
			notFound = true
		}

		if notFound {
			log.Infoln("No config file found, writing a default one.")
			log.Tracef("Current User's home dir: %v", home)
			_, findDirErr := os.Stat(defaultSameConfigDir)
			if findDirErr != nil {
				log.Infof("No config directory found at %v, creating one.\n", defaultSameConfigDir)
				createDirErr := os.Mkdir(defaultSameConfigDir, 0750)
				if createDirErr != nil {
					return fmt.Errorf("Failed to write to create config directory: %v", createDirErr)
				}
			}
			viper.Set("created-at", time.Now().Format(time.RFC3339))
			err = viper.SafeWriteConfigAs(path.Join(defaultSameConfigDir, "config.yaml"))
			if err != nil {
				return fmt.Errorf("Error while writing a default config file: %v", err)
			}
			return nil
		} else if badPath {
			message := fmt.Errorf("No config file found at: %v", configPath)
			return message
		} else {
			message := fmt.Errorf("Config file found at '%v', but error: %v", configPath, err)
			return message
		}
	}

}

func DetectRootless() (detected bool) {
	return false
}
