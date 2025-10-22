package ui

import (
	"github.com/go-gl/gl/v4.4-core/gl"
)

type texture struct {
	id uint32
}

func newTexture(data []byte, width int, height int) texture {
	texture := texture{}
	gl.GenTextures(1, &texture.id)
	gl.BindTexture(gl.TEXTURE_2D, texture.id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB8, int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(data))

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return texture
}

func (t texture) bind() {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.id)
}

func (t texture) unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (t *texture) updateTexture(data []byte, width int32, height int32) {
	t.bind()
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, width, height, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(data))
}
