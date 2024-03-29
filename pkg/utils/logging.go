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
	errorString := ""
	if err != nil {
		errorString = err.Error()
	}
	log.Warnf("%v\n%v", s, errorString)

	return os.Getenv("TEST_PASS") != ""
}
