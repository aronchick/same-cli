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

	// // Set the AppDir and ConfigFileName for kfconfig
	// if isRemoteFile {
	// 	cwd, err := os.Getwd()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("could not get current directory for KfDef %v", err)
	// 	}
	// 	sameconfig.Spec.AppDir = cwd
	// } else {
	// 	sameconfig.Spec.AppDir = filepath.Dir(configFile)
	// }
	// kfconfig.Spec.ConfigFileName = filepath.Base(configFile)
	return sameconfig, nil
}

// func isCwdEmpty() string {
// 	cwd, _ := os.Getwd()
// 	files, _ := ioutil.ReadDir(cwd)
// 	if len(files) > 1 {
// 		return ""
// 	}
// 	return cwd
// }

// func WriteConfigToFile(config kfconfig.KfConfig) error {
// 	if config.Spec.AppDir == "" {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INVALID_ARGUMENT),
// 			Message: "No AppDir, cannot write to file.",
// 		}
// 	}
// 	if config.Spec.ConfigFileName == "" {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INVALID_ARGUMENT),
// 			Message: "No ConfigFileName, cannot write to file.",
// 		}
// 	}
// 	filename := filepath.Join(config.Spec.AppDir, config.Spec.ConfigFileName)
// 	converters := map[string]Loader{
// 		"v1alpha1": V1alpha1{},
// 		"v1beta1":  V1beta1{},
// 		"v1":       V1{},
// 	}
// 	apiVersionSeparated := strings.Split(config.APIVersion, "/")
// 	if len(apiVersionSeparated) < 2 || apiVersionSeparated[0] != Api {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INVALID_ARGUMENT),
// 			Message: fmt.Sprintf("invalid config: apiVersion must be in the format of %v/<version>, got %v", Api, config.APIVersion),
// 		}
// 	}

// 	converter, ok := converters[apiVersionSeparated[1]]
// 	if !ok {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INVALID_ARGUMENT),
// 			Message: fmt.Sprintf("invalid config: unable to find converter for version %v", config.APIVersion),
// 		}
// 	}

// 	var kfdef interface{}
// 	if err := converter.LoadKfDef(config, &kfdef); err != nil {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INVALID_ARGUMENT),
// 			Message: fmt.Sprintf("error when loading KfDef: %v", err),
// 		}
// 	}
// 	kfdefBytes, err := yaml.Marshal(kfdef)
// 	if err != nil {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INVALID_ARGUMENT),
// 			Message: fmt.Sprintf("error when marshaling KfDef: %v", err),
// 		}
// 	}

// 	err = ioutil.WriteFile(filename, kfdefBytes, 0644)
// 	if err != nil {
// 		return &kfapis.KfError{
// 			Code:    int(kfapis.INTERNAL_ERROR),
// 			Message: fmt.Sprintf("error when writing KfDef: %v", err),
// 		}
// 	}
// 	return nil
// }
