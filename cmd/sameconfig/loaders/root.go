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
		return false, fmt.Errorf("config file must be a URI or a path")
	}
	url, err := netUrl.Parse(configFile)
	if err != nil {
		return false, fmt.Errorf("error parsing file path: %v", err)
	}
	if url.Scheme != "" {
		return true, nil
	}
	return false, nil
}

// LoadSAMEConfig reads the samedef from a remote URI or local file,
// and returns the sameconfig.
func LoadSAMEConfig(configFilePath string) (*SameConfig, error) {
	if configFilePath == "" {
		return nil, fmt.Errorf("config file must be the URI of a SameDef spec")
	}

	// Read contents
	configFileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read from config file %s: %v", configFilePath, err)
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
	sameconfig, err := v1.LoadSameConfig(obj)
	if err != nil {
		log.Errorf("Failed to convert kfdef to kfconfig: %v", err)
		return nil, err
	}

	return sameconfig, nil
}
