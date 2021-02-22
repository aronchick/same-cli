package utils

import (
	"fmt"
	"io/fs"

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
func LoadConfig(path string) (err error) {
	log.Info("in utils.LoadConfig")
	viper.AutomaticEnv() // read in environment variables that match

	viper.SetConfigFile(path)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
		return nil
	} else {
		_, notFound := err.(viper.ConfigFileNotFoundError)
		_, badPath := err.(*fs.PathError)
		if badPath || notFound {
			message := fmt.Errorf("No config file found at: %v", path)
			return message
		} else {
			message := fmt.Errorf("Config file found at '%v', but error: %v", path, err)
			return message
		}
	}
}
