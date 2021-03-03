package utils

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// PrettyPrint returns a pretty format output of any value.
func PrettyPrint(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}
	valueJson, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal value; error %v", err)
		return fmt.Sprintf("%+v", value)
	}
	return string(valueJson)
}

func PrintError(s string, err error) (exit bool) {
	log.Tracef("Error output: %v", err.Error())
	message := fmt.Sprintf(s, err.Error())
	log.Fatalf(message + "\n")

	return os.Getenv("TEST_PASS") != ""
}
