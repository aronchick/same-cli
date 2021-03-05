package mocks

type MockConfigFileIO struct {
}

func (mcfio *MockConfigFileIO) ConfigWriter() error {
	return nil
}
