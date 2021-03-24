package utils

import (
	"os"

	"github.com/spf13/cobra"
)

type UtilsInterface interface {
	DetectK3s() (string, error)
	K3sRunning(*cobra.Command) (bool, error)
	GetCmd() *cobra.Command
	SetCmd(*cobra.Command)
	GetCmdArgs() []string
	SetCmdArgs([]string)
}

func GetUtils() UtilsInterface {
	if os.Getenv("GITHUB_ACTIONS") == "" {
		return &UtilsLive{}
	} else {
		return &UtilsMock{}
	}
}
