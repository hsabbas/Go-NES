package ui

import (
	"fmt"

	"github.com/go-gl/gl/v4.4-core/gl"
)

type shader struct {
	program      uint32
	uniformCache map[string]int32
}

func newShader(vertexShaderSource string, fragmentShaderSource string) (shader, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return shader{}, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return shader{}, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	gl.ValidateProgram(program)
	gl.UseProgram(program)

	return shader{
		program:      program,
		uniformCache: make(map[string]int32),
	}, nil
}

func (s *shader) bind() {
	gl.UseProgram(s.program)
}

func (s *shader) delete() {
	gl.DeleteProgram(s.program)
}

func (s *shader) setUniform4f(name string, v0 float32, v1 float32, v2 float32, v3 float32) error {
	location := s.getUniformLocation(name)
	if location == -1 {
		return fmt.Errorf("could not find uniform %v", name)
	}
	gl.Uniform4f(location, v0, v1, v2, v3)
	return nil
}

func (s *shader) setUniform1i(name string, value int32) error {
	location := s.getUniformLocation(name)
	if location == -1 {
		return fmt.Errorf("could not find uniform %v", name)
	}
	gl.Uniform1i(location, value)
	return nil
}

func (s *shader) getUniformLocation(name string) int32 {
	if _, ok := s.uniformCache[name]; ok {
		return s.uniformCache[name]
	}

	location := gl.GetUniformLocation(s.program, gl.Str(name+"\x00"))
	s.uniformCache[name] = location
	return location
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := string(make([]byte, logLength))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
