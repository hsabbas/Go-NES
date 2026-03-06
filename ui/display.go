package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/hsabbas/Go-NES-Emulator/nes"
)

type Display struct {
	console  *nes.NES
	texture  *rl.Texture2D
	srcRect  *rl.Rectangle
	destRect *rl.Rectangle
}

func Init(console *nes.NES) *Display {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(256*3, 240*3, "NES Emulator by Hassan :)")
	img := rl.NewImage(make([]byte, 240*256*4), 256, 240, 1, rl.UncompressedR8g8b8a8)
	texture := rl.LoadTextureFromImage(img)
	rl.SetTargetFPS(60)

	srcRect := rl.NewRectangle(0, 0, 256, 240)

	destRect := &rl.Rectangle{}

	d := &Display{
		console:  console,
		texture:  &texture,
		srcRect:  &srcRect,
		destRect: destRect,
	}
	d.updateDestRect()
	return d
}

func (d *Display) ProcessInput() {
	var buttons byte
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		buttons |= byte(nes.Up)
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		buttons |= byte(nes.Down)
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		buttons |= byte(nes.Left)
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		buttons |= byte(nes.Right)
	}
	if rl.IsKeyDown(rl.KeyPeriod) || rl.IsKeyDown(rl.KeyX) {
		buttons |= byte(nes.A)
	}
	if rl.IsKeyDown(rl.KeyComma) || rl.IsKeyDown(rl.KeyZ) {
		buttons |= byte(nes.B)
	}
	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		buttons |= byte(nes.Select)
	}
	if rl.IsKeyDown(rl.KeyEnter) {
		buttons |= byte(nes.Start)
	}

	d.console.UpdatePlayer1Register(buttons)
}

func (d *Display) Render() {
	if rl.IsWindowResized() {
		d.updateDestRect()
	}

	pixels := d.console.GetImage()

	// Hopefully this can be removed soon. The raylib-go master branch has
	// support for passing []byte pixel data to UpdateTexture, but current
	// release only supports []color.RGBA.
	colors := make([]color.RGBA, 256*240)
	i := 0
	for y := range 240 {
		for x := 0; x < 256; x++ {
			colors[i] = color.RGBA{
				R: pixels[y][x*3],
				G: pixels[y][x*3+1],
				B: pixels[y][x*3+2],
				A: 255,
			}
			i++
		}
	}

	rl.UpdateTexture(*d.texture, colors)
	rl.BeginDrawing()
	rl.DrawTexturePro(*d.texture, *d.srcRect, *d.destRect, rl.NewVector2(0, 0), 0, rl.White)
	rl.EndDrawing()
}

// The destRect is a rectangle to render the texture (the actual NES output) to.
// It is sized to be the as large as the window allows with the correct aspect
// ratio, then positioned to be in the center.
func (d *Display) updateDestRect() {
	w := rl.GetScreenWidth()
	h := rl.GetScreenHeight()
	scale := min(float32(w)/float32(256), float32(h)/float32(240))
	d.destRect.Width = 256 * scale
	d.destRect.Height = 240 * scale
	d.destRect.X = float32((w - int(d.destRect.Width)) / 2)
	d.destRect.Y = float32((h - int(d.destRect.Height)) / 2)
}

func (d *Display) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (d *Display) Close() {
	rl.UnloadTexture(*d.texture)
}
