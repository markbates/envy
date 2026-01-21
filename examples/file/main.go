package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/markbates/envy"
)

func main() {
	cab := os.DirFS(".")
	env, err := envy.FromFile(cab, "app.env")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(strings.Join(env.Environ(), ";"))
}
