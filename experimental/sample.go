package main

import "fmt"

func main() {
	// os.Setenv("PATH", "/sbin")
	// path, err := exec.LookPath("kubectl")
	// if err != nil {
	// 	log.Fatal("installing kubectl is in your future")
	// }
	// fmt.Printf("fortune is available at %s\n", path)

	a := []string{"a", "b", "c"}
	fmt.Println(a)
	b := a[3:]
	_ = b
	fmt.Println(b)
}
