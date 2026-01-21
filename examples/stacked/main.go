package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/markbates/envy"
)

func main() {
	// new env with one variable
	env := envy.FromMap(map[string]string{
		"HOME": "/usr/home",
	})

	sep := ";"

	// print the environment
	fmt.Println("Initial environment:")
	fmt.Println(strings.Join(env.Environ(), sep))
	fmt.Println()

	// create a fs.FS for the current directory
	cab := os.DirFS(".")

	// update env with files
	env, err := envy.With(env, func() (*envy.Env, error) {
		// load base.env
		e, err := envy.FromFile(cab, "base.env")
		if err != nil {
			return nil, err
		}

		// load dev.env on top of base.env
		return envy.With(e, func() (*envy.Env, error) {
			return envy.FromFile(cab, "dev.env")
		})
	})

	if err != nil {
		fmt.Println("error loading env files:", err)
		return
	}

	// print the updated environment
	fmt.Println("Final environment:")
	fmt.Println(strings.Join(env.Environ(), sep))
}
