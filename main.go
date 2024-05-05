package main

import (
	"os"

	"github.com/leonardinius/golox/cmd"
)

func main() {
	app := cmd.NewLoxApp()
	os.Exit(app.Main(os.Args[1:]))
}
