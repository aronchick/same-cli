package mocks

import (
	"github.com/spf13/viper"
)

type MockConfigFileIO struct {
}

func (mcfio *MockConfigFileIO) ConfigWriter(viper viper.Viper) error {
	return nil
}
