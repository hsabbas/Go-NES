package ui

import (
	"log"
	"os"

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
	rl.InitWindow(256*3, 240*3, "Go-NES")

	d := &Display{}

	if isNesRom(romPath) {
		gameview, err := startGameEarly(romPath)
		if err == nil {
			d.view = gameview
			return d, nil
		}
		log.Println("Failed to start with given ROM:", err)
	}

	d.view = createMenuView(romPath, d.startGameView)

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

func (d *Display) startGameView(rom []byte) error {
	console, err := nes.BootNES(rom)
	if err != nil {
		return err
	}
	d.view.close()
	d.view = createGameView(console)
	return nil
}

func (d *Display) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (d *Display) Close() {
	d.view.close()
	rl.CloseWindow()
}

func startGameEarly(romPath string) (*gameview, error) {
	rom, err := os.ReadFile(romPath)
	if err != nil {
		return nil, err
	}

	console, err := nes.BootNES(rom)
	if err != nil {
		return nil, err
	}

	return createGameView(console), nil
}
