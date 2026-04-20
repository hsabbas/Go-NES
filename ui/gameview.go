package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/hsabbas/Go-NES-Emulator/nes"
)

type gameview struct {
	console  *nes.NES
	texture  *rl.Texture2D
	srcRect  *rl.Rectangle
	destRect *rl.Rectangle

	colors     []color.RGBA
	frameReady bool

	firstFrame bool
}

func createGameView(console *nes.NES) *gameview {
	img := rl.NewImage(make([]byte, 240*256*4), 256, 240, 1, rl.UncompressedR8g8b8a8)
	texture := rl.LoadTextureFromImage(img)

	srcRect := rl.NewRectangle(0, 0, 256, 240)
	destRect := &rl.Rectangle{}
	view := &gameview{
		console:    console,
		texture:    &texture,
		srcRect:    &srcRect,
		destRect:   destRect,
		colors:     make([]color.RGBA, 256*240),
		firstFrame: true,
	}
	view.updateDestRect()
	return view
}

func (g *gameview) processInput() {
	g.console.UpdatePlayer1Register(g.pollPlayer1Input())
	g.console.UpdatePlayer2Register(g.pollPlayer2Input())
}

func (g *gameview) pollPlayer1Input() byte {
	var buttons byte
	if player1PressingUp() {
		buttons |= byte(nes.Up)
	}
	if player1PressingDown() {
		buttons |= byte(nes.Down)
	}
	if player1PressingLeft() {
		buttons |= byte(nes.Left)
	}
	if player1PressingRight() {
		buttons |= byte(nes.Right)
	}
	if player1PressingA() {
		buttons |= byte(nes.A)
	}
	if player1PressingB() {
		buttons |= byte(nes.B)
	}
	if player1PressingSelect() {
		buttons |= byte(nes.Select)
	}
	if player1PressingStart() {
		buttons |= byte(nes.Start)
	}
	return buttons
}

func (g *gameview) pollPlayer2Input() byte {
	var buttons byte
	if player2PressingUp() {
		buttons |= byte(nes.Up)
	}
	if player2PressingDown() {
		buttons |= byte(nes.Down)
	}
	if player2PressingLeft() {
		buttons |= byte(nes.Left)
	}
	if player2PressingRight() {
		buttons |= byte(nes.Right)
	}
	if player2PressingA() {
		buttons |= byte(nes.A)
	}
	if player2PressingB() {
		buttons |= byte(nes.B)
	}
	if player2PressingSelect() {
		buttons |= byte(nes.Select)
	}
	if player2PressingStart() {
		buttons |= byte(nes.Start)
	}
	return buttons
}

func (g *gameview) update() {
	g.console.RunFrame()
	g.frameReady = true
}

func (g *gameview) render() {
	if rl.IsWindowResized() {
		g.updateDestRect()
	}

	if g.frameReady {
		pixels := g.console.GetImage()

		// Hopefully this can be removed soon. The raylib-go master branch has
		// support for passing []byte pixel data to UpdateTexture, but current
		// release only supports []color.RGBA.
		i := 0
		for y := range 240 {
			for x := 0; x < 256; x++ {
				g.colors[i] = color.RGBA{
					R: pixels[y][x*3],
					G: pixels[y][x*3+1],
					B: pixels[y][x*3+2],
					A: 255,
				}
				i++
			}
		}
		rl.UpdateTexture(*g.texture, g.colors)
		g.frameReady = false
	}

	rl.BeginDrawing()
	rl.DrawTexturePro(*g.texture, *g.srcRect, *g.destRect, rl.NewVector2(0, 0), 0, rl.White)
	rl.EndDrawing()
}

// The destRect is a rectangle to render the texture (the actual NES output) to.
// It is sized to be the as large as the window allows with the correct aspect
// ratio, then positioned to be in the center.
func (g *gameview) updateDestRect() {
	w := rl.GetScreenWidth()
	h := rl.GetScreenHeight()

	scale := min(w/256, h/240)
	if scale < 1 {
		scale = 1
	}

	g.destRect.Width = float32(256 * scale)
	g.destRect.Height = float32(240 * scale)
	g.destRect.X = float32((w - int(g.destRect.Width)) / 2)
	g.destRect.Y = float32((h - int(g.destRect.Height)) / 2)
}

func (g *gameview) close() {
	rl.UnloadTexture(*g.texture)
}
