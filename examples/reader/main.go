package main

import (
	"fmt"
	"strings"

	"github.com/markbates/envy"
)

func main() {
	input := strings.NewReader("APP_ENV=dev; PORT=4000; DEBUG=true;")
	env, err := envy.FromReader(input, ';')
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(strings.Join(env.Environ(), ";"))
}
