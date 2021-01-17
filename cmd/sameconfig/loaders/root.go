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

// const (
// 	Api = "kfdef.apps.kubeflow.org"
// )

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

// LoadConfigFromURI reads the samedef from a remote URI or local file,
// and returns the sameconfig.
func LoadConfigFromURI(configFile string) (*SameConfig, error) {
	if configFile == "" {
		return nil, fmt.Errorf("config file must be the URI of a SameDef spec")
	}

	isRemoteFile, err := IsRemoteFile(configFile)
	if err != nil {
		return nil, err
	}

	// appFile is configFile if configFile is local.
	// Otherwise (configFile is remote), appFile points to a downloaded copy of configFile in tmp.
	appFile := configFile
	// If config is remote, download it to a temp dir.
	if isRemoteFile {
		// TODO(jlewi): We should check if configFile doesn't specify a protocol or the protocol
		// is file:// then we can just read it rather than fetching with go-getter.
		appDir, err := ioutil.TempDir("", "")
		if err != nil {
			return nil, fmt.Errorf("unable to create a temporary directory to copy the file to")
		}
		// Open config file
		appFile = path.Join(appDir, "tmp_app.yaml")

		log.Infof("Downloading %v to %v", configFile, appFile)
		configFileURI, err := netUrl.Parse(configFile)
		if err != nil {
			log.Errorf("could not parse configFile url")
		}
		if isValidURL(configFile) {
			errGet := gogetter.GetFile(appFile, configFile)
			if errGet != nil {
				return nil, fmt.Errorf("could not fetch specified config %s: %v", configFile, errGet)
			}
		} else {
			g := new(gogetter.FileGetter)
			g.Copy = true
			errGet := g.GetFile(appFile, configFileURI)
			if errGet != nil {
				return nil, fmt.Errorf("could not fetch specified config %s: %v", configFile, err)
			}
		}
	}

	// Read contents
	configFileBytes, err := ioutil.ReadFile(appFile)
	if err != nil {
		return nil, fmt.Errorf("could not read from config file %s: %v", configFile, err)
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
