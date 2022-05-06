package apiserver

import "fmt"

// Run runs the specified APISERVER. This should never exit.
func Run() error {
	fmt.Println("Hello World!")
	return nil
}
