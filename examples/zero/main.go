package main

import (
	"fmt"

	"github.com/markbates/envy"
)

func main() {
	// create a new empty Env
	env := envy.Zero()

	// print the environment entries
	fmt.Println("env:", env.Environ())
}
