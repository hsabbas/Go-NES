package ui

import (
	"runtime"

	"github.com/go-gl/gl/v4.4-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hsabbas/Go-NES-Emulator/nes"
)

type window struct {
	w                  *glfw.Window
	controllerListener func(byte, bool)
}

const (
	width  = 256
	height = 240
)

func CreateWindow(console *nes.NES) (window, error) {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		return window{}, err
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	w, err := glfw.CreateWindow(width*3, height*3, "NES Emulator by Hassan :)", nil, nil)
	if err != nil {
		return window{}, err
	}

	w.MakeContextCurrent()
	w.SetAspectRatio(16, 15)
	glfw.SwapInterval(1)

	w.SetFramebufferSizeCallback(windowResize)

	w.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Repeat {
			return
		}

		pressed := action == glfw.Press

		if key == glfw.KeyW || key == glfw.KeyUp {
			console.ReceiveInput(nes.Up, pressed)

		}

		if key == glfw.KeyS || key == glfw.KeyDown {
			console.ReceiveInput(nes.Down, pressed)
		}

		if key == glfw.KeyA || key == glfw.KeyLeft {
			console.ReceiveInput(nes.Left, pressed)
		}

		if key == glfw.KeyD || key == glfw.KeyRight {
			console.ReceiveInput(nes.Right, pressed)
		}

		if key == glfw.KeyPeriod || key == glfw.KeyX {
			console.ReceiveInput(nes.A, pressed)
		}

		if key == glfw.KeyComma || key == glfw.KeyZ {
			console.ReceiveInput(nes.B, pressed)
		}

		if key == glfw.KeyLeftShift || key == glfw.KeyRightShift {
			console.ReceiveInput(nes.Select, pressed)
		}

		if key == glfw.KeyEnter {
			console.ReceiveInput(nes.Start, pressed)
		}
	})

	return window{
		w: w,
	}, nil
}

func (w *window) Close() {
	w.w.Destroy()
	glfw.Terminate()
}

func (w *window) ShouldClose() bool {
	return w.w.ShouldClose()
}

func (w *window) Update() {
	w.w.SwapBuffers()
	glfw.PollEvents()
}

func windowResize(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
