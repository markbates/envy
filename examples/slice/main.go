package main

import (
	"fmt"
	"strings"

	"github.com/markbates/envy"
)

func main() {
	env := envy.FromSlice([]string{
		"HOME=/tmp/usr",
		"PORT=4000",
	})

	fmt.Println(strings.Join(env.Environ(), ";"))
}
