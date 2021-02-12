package cmd

import (
	"context"
	"fmt"

	"github.com/azure-octo/same-cli/cmd/sameconfig/loaders"
)

// ParseConfig takes a flat string and a version and converts it into a strongly typed struct.
func ParseConfig(ctx context.Context, fileURI string) (sameConfig *loaders.SameConfig, err error) {
	// Only works (for right now) against file in the root of the directory
	sameConfig, err = loaders.LoadSAMEConfig(fileURI)

	if err != nil {
		fmt.Printf("failed to load config: %v", err.Error())
	}

	return sameConfig, nil
}
