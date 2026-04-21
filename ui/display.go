package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/hsabbas/Go-NES-Emulator/nes"
)

type view interface {
	processInput()
	update()
	render()
	close()
}

type Display struct {
	view view
}

func Init(romPath string) (*Display, error) {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(256*3, 240*3, "NES Emulator by Hassan :)")

	d := &Display{}

	rom, err := loadROM(romPath)
	if err != nil {
		return nil, err
	}

	console, err := nes.BootNES(rom)
	if err != nil {
		return nil, err
	}

	d.view = createGameView(console)
	return d, nil
}

func (d *Display) SetTargetFPS(fps int32) {
	rl.SetTargetFPS(fps)
}

func (d *Display) Run() {
	for !d.ShouldClose() {
		d.view.processInput()
		d.view.update()
		d.view.render()
	}
}

func (d *Display) startGameView(rom []byte) {
	console, err := nes.BootNES(rom)
	if err != nil {
		log.Println("failed to boot NES")
		return
	}
	d.view.close()
	d.view = createGameView(console)
}

func (d *Display) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (d *Display) Close() {
	d.view.close()
	rl.CloseWindow()
}

func loadROM(romPath string) ([]byte, error) {
	if isNesRom(romPath) {
		return os.ReadFile(romPath)
	}

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
