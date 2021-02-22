package loaders

import (
	"fmt"

	"io/ioutil"
	netUrl "net/url"

	"github.com/ghodss/yaml"

	log "github.com/sirupsen/logrus"
)

// V1 Empty struct - used to implement Converter interface.
type V1 struct {
}

// Loader is an interface describing LoadSameConfig and LoadSameDef for this version of the API.
type Loader interface {
	LoadSameConfig(samedef interface{}) (*SameConfig, error)
	LoadSameDef(config SameConfig, out interface{}) error
}

// IsRemoteFile checks if the path configFile is remote (e.g. http://github...)
func IsRemoteFile(configFile string) (bool, error) {
	if configFile == "" {
		message := fmt.Errorf("config file must be a URI or a path")
		log.Errorf(message.Error())
		return false, message
	}
	url, err := netUrl.Parse(configFile)
	if err != nil {
		message := fmt.Errorf("error parsing file path: %v", err)
		log.Errorf(message.Error())
		return false, message
	}
	if url.Scheme != "" {
		return true, nil
	}
	return false, nil
}

// LoadSAMEConfig reads the samedef from a remote URI or local file,
// and returns the sameconfig.
func LoadSAME(configFilePath string) (*SameConfig, error) {
	if configFilePath == "" {
		return nil, fmt.Errorf("config file must be the URI of a SameDef spec")
	}

	// Read contents
	resolvedConfigFilePath, err := netUrl.Parse(configFilePath)
	if err != nil {
		log.Errorf("root.go: could not resolve same config file path: %v", err)
		return nil, err
	}
	configFileBytes, err := ioutil.ReadFile(resolvedConfigFilePath.Path)
	if err != nil {
		message := fmt.Errorf("root.go: could not read from config file %s: %v", configFilePath, err)
		log.Errorf(message.Error())
		return nil, message
	}

	// Check API version.
	var obj map[string]interface{}
	if err = yaml.Unmarshal(configFileBytes, &obj); err != nil {
		return nil, fmt.Errorf("unable to unmarshal the yaml file - invalid config file format: %v", err)
	}
	// apiVersion, ok := obj["apiVersion"]
	// if !ok {
	// 	return nil, fmt.Errorf("invalid config: apiVersion is not found.")
	// }

	v1 := V1{}
	sameconfig, err := v1.LoadSAME(obj)
	if err != nil {
		log.Errorf("Failed to convert kfdef to kfconfig: %v", err)
		return nil, err
	}

	return sameconfig, nil
}
