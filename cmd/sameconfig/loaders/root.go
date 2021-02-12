package loaders

import (
	"fmt"

	"io/ioutil"
	netUrl "net/url"
	"path"

	"github.com/ghodss/yaml"
	gogetter "github.com/hashicorp/go-getter"

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

// isValidURL reports if the URL is correct.
func isValidURL(toTest string) bool {
	_, err := netUrl.ParseRequestURI(toTest)
	return err == nil
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

	isRemoteFile, err := IsRemoteFile(configFilePath)
	if err != nil {
		return nil, err
	}

	// appFile is configFilePath if configFilePath is local.
	// Otherwise (configFile is remote), appFile points to a downloaded copy of configFilePath in tmp.
	appFile := configFilePath
	// If config is remote, download it to a temp dir.
	if isRemoteFile {
		// TODO(jlewi): We should check if configFilePath doesn't specify a protocol or the protocol
		// is file:// then we can just read it rather than fetching with go-getter.
		appDir, err := ioutil.TempDir("", "")
		if err != nil {
			return nil, fmt.Errorf("unable to create a temporary directory to copy the file to")
		}
		// Open config file
		appFile = path.Join(appDir, "tmp_app.yaml")

		log.Infof("Downloading %v to %v", configFilePath, appFile)
		configFileURI, err := netUrl.Parse(configFilePath)
		if err != nil {
			log.Errorf("could not parse configFile url")
		}
		if isValidURL(configFilePath) {
			errGet := gogetter.GetFile(appFile, configFilePath)
			if errGet != nil {
				return nil, fmt.Errorf("could not fetch specified config %s: %v", configFilePath, errGet)
			}
		} else {
			g := new(gogetter.FileGetter)
			g.Copy = true
			errGet := g.GetFile(appFile, configFileURI)
			if errGet != nil {
				return nil, fmt.Errorf("could not fetch specified config %s: %v", configFilePath, err)
			}
		}
	}

	// Read contents
	configFileBytes, err := ioutil.ReadFile(appFile)
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
