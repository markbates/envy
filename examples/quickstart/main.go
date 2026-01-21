package main

import (
	"fmt"

	"github.com/markbates/envy"
)

func main() {
	// new env based on os.Environ
	env := envy.New()

	// print the HOME environment variable
	fmt.Println("HOME: ", env.Getenv("HOME"))

	// set a new HOME variable
	err := env.Setenv("HOME", "/tmp/home")
	if err != nil {
		fmt.Println("Error setting HOME:", err)
		return
	}

	// print the updated HOME variable
	fmt.Println("Updated HOME: ", env.Getenv("HOME"))
}
