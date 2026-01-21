package main

import (
	"fmt"

	"github.com/markbates/envy"
)

func main() {
	// create a new Env from the current process environment
	env := envy.New()

	// get and print the value of the HOME environment variable
	fmt.Println("HOME:", env.Getenv("HOME"))
}
