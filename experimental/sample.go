package main

import (
	"fmt"
	"net/url"

	"github.com/azure-octo/same-cli/pkg/utils"
)

func main() {
	b, err := url.Parse("google.com")

	fmt.Println(b)
	fmt.Println(err)

	a, err := utils.IsRemoteFilePath("google.com")
	fmt.Println(err)
	fmt.Println(a)
}
