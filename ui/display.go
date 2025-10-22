package ui

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v4.4-core/gl"
)

type Display struct {
	f frame
}

func CreateDisplay() (*Display, error) {
	err := gl.Init()
	if err != nil {
		return nil, fmt.Errorf("error initializing gl: %s", err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version:", version)

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(debugCallback, nil)

	frame, err := createFrame()
	if err != nil {
		return nil, fmt.Errorf("error creating frame: %s", err)
	}

	return &Display{
		f: frame,
	}, nil
}

func (d *Display) RenderFrame() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	d.f.bind()
	gl.DrawElements(gl.TRIANGLES, int32(d.f.va.indexCount), gl.UNSIGNED_INT, nil)
}

func (d *Display) Close() {
	d.f.delete()
}

func (d *Display) ReceiveNESFrame(pixels [240][256]uint16, width int32, height int32) {
	data := make([]byte, 256*240*3)

	ind := 0
	for i := 239; i >= 0; i-- {
		for _, val := range pixels[i] {
			data[ind] = colorMap[val].r
			data[ind+1] = colorMap[val].g
			data[ind+2] = colorMap[val].b
			ind += 3
		}
	}
	d.f.t.updateTexture(data, width, height)
}

func debugCallback(source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {
	fmt.Println("OpenGL Debug:", message)
}

func createImage(width int, height int) []byte {
	data := make([]byte, width*height*3)
	i := 0
	for row := range height {
		r := byte(255 * (float32(height-row) / float32(height)))
		for col := range width {
			data[i] = r
			data[i+1] = byte(255 * float32(col) / float32(width))
			data[i+2] = 0
			i += 3
		}
	}
	return data
}
