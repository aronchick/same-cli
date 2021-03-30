package loaders

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// // Empty struct - used to implement Converter interface.
// type V1 struct {
// }

// LoadSameConfig takes the loaded definiton file, loads it and then unmarshalls it into a SameConfig struct.
func (v V1) LoadSAME(def interface{}) (sameConfig *SameConfig, err error) {
	// First create the struct to unmarshall the yaml into
	sameConfigFromFile := &SameSpec{}

	bytes, err := yaml.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("could not marshal input file into bytes: %v", err)
	}
	err = yaml.Unmarshal(bytes, sameConfigFromFile)
	if err != nil {
		return nil, fmt.Errorf("could not unpack same configuration file: %v", err)
	}

	sameConfig = &SameConfig{
		Spec: SameSpec{
			APIVersion: sameConfigFromFile.APIVersion,
			Version:    sameConfigFromFile.Version,
		},
	}

	sameConfig.Spec.Metadata = sameConfigFromFile.Metadata
	sameConfig.Spec.Bases = sameConfigFromFile.Bases
	sameConfig.Spec.EnvFiles = sameConfigFromFile.EnvFiles
	sameConfig.Spec.Resources = sameConfigFromFile.Resources
	sameConfig.Spec.Workflow.Parameters = sameConfigFromFile.Workflow.Parameters
	sameConfig.Spec.Pipeline = sameConfigFromFile.Pipeline
	sameConfig.Spec.DataSets = sameConfigFromFile.DataSets
	sameConfig.Spec.Run = sameConfigFromFile.Run
	sameConfig.Spec.ConfigFilePath = sameConfigFromFile.ConfigFilePath

	// a, _ := yaml.Marshal(sameConfig)
	// fmt.Println(string(a))

	return sameConfig, nil
}
