package ui

import "github.com/go-gl/gl/v4.4-core/gl"

const (
	vertexShaderSource = `
    #version 460
    layout (location=0) in vec4 position;
	layout (location=1) in vec2 texCoord;

	out vec2 vTexCoord;

    void main() {
        gl_Position = position;
		vTexCoord = texCoord;
    }
` + "\x00"

	fragmentShaderSource = `
    #version 460
    layout(location = 0) out vec4 color;

	in vec2 vTexCoord;

	uniform vec4 uColor;
	uniform sampler2D uTexture;

    void main() {
		vec4 texColor = texture(uTexture, vTexCoord);
        color = texColor;
    }
` + "\x00"
)

var (
	square = []float32{
		-1, -1, 0, 0,
		1, -1, 1, 0,
		1, 1, 1, 1,
		-1, 1, 0, 1,
	}

	indices = []uint32{
		0, 1, 2,
		2, 3, 0,
	}
)

type frame struct {
	s  shader
	va vertexArray
	t  texture
}

func createFrame() (frame, error) {
	shader, err := newShader(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return frame{}, err
	}

	va := newVertexArray()
	va.addVertexBuffer(gl.Ptr(square), 4*len(square))
	va.addIndexBuffer(gl.Ptr(indices), len(indices))

	bl := bufferLayout{}
	bl.addElement(gl.FLOAT, 2)
	bl.addElement(gl.FLOAT, 2)
	va.setBufferLayout(bl)

	texture := newTexture(createImage(width, height), width, height)
	texture.bind()

	err = shader.setUniform1i("uTexture", 0)
	if err != nil {
		return frame{}, err
	}

	return frame{
		s:  shader,
		va: va,
		t:  texture,
	}, nil
}

func (f *frame) bind() {
	f.s.bind()
	f.va.bind()
}

func (f *frame) delete() {
	f.s.delete()
}

func (f *frame) newFrame(data []byte, width int32, height int32) {
	f.t.updateTexture(data, width, height)
}
