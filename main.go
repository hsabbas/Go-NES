package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hsabbas/Go-NES-Emulator/nes"
	"github.com/hsabbas/Go-NES-Emulator/ui"
)

func main() {
	rom, err := loadROM()
	if err != nil {
		panic(err)
	}

	nes := nes.BootNES(rom)

	window, err := ui.CreateWindow(nes)
	if err != nil {
		panic(err)
	}
	defer window.Close()

	display, err := ui.CreateDisplay()
	if err != nil {
		panic(err)
	}
	defer display.Close()

	nes.SetFrameCallback(func(pixels [240][256]uint16) {
		display.ReceiveNESFrame(pixels, 256, 240)
	})

	frameDuration := time.Millisecond * 16
	t := time.NewTicker(frameDuration)
	for !window.ShouldClose() {
		<-t.C
		display.RenderFrame()
		window.Update()
		nes.RunFrame()
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
				rom, err := os.ReadFile(entry.Name())
				if err != nil {
					return nil, err
				}
				return rom, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find rom")
}
