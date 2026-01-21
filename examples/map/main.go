package main

import (
	"fmt"
	"strings"

	"github.com/markbates/envy"
)

func main() {
	env := envy.FromMap(map[string]string{
		"APP_ENV": "dev",
		"PORT":    "4000",
	})

	fmt.Println(strings.Join(env.Environ(), ";"))
}
