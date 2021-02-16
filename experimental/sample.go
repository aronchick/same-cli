package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	// os.Setenv("PATH", "/sbin")
	path, err := exec.LookPath("kubectl")
	if err != nil {
		log.Fatal("installing kubectl is in your future")
	}
	fmt.Printf("fortune is available at %s\n", path)
}
