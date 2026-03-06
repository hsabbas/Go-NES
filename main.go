package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hsabbas/Go-NES-Emulator/nes"
	"github.com/hsabbas/Go-NES-Emulator/ui"
)

func main() {
	var console *nes.NES
	if len(os.Args) > 1 {
		rom, err := os.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		console = nes.BootNES(rom)
	} else {
		rom, err := loadROM()
		if err != nil {
			panic(err)
		}
		console = nes.BootNES(rom)
	}

	display := ui.Init(console)
	defer display.Close()

	for !display.ShouldClose() {
		display.ProcessInput()

		console.RunFrame()

		display.Render()
	}
}

func loadROM() ([]byte, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Type().IsRegular() {
			if filepath.Ext(entry.Name()) == ".nes" {
				return os.ReadFile(entry.Name())
			}
		}
	}

	return nil, fmt.Errorf("cannot find rom")
}
