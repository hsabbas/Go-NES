package main

import (
	"os"

	"github.com/hsabbas/Go-NES-Emulator/ui"
)

func main() {
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	display, err := ui.Init(path)
	if err != nil {
		panic(err)
	}
	defer display.Close()

	display.SetTargetFPS(60)
	display.Run()
}
