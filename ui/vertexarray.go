package ui

import (
	"unsafe"

	"github.com/go-gl/gl/v4.4-core/gl"
)

type vertexArray struct {
	vao, vbo, ibo uint32
	indexCount    int
}

func newVertexArray() vertexArray {
	va := vertexArray{}
	gl.GenVertexArrays(1, &va.vao)
	gl.BindVertexArray(va.vao)
	return va
}

func (va *vertexArray) addVertexBuffer(data unsafe.Pointer, size int) {
	gl.BindVertexArray(va.vao)
	gl.GenBuffers(1, &va.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, size, data, gl.STATIC_DRAW)
}

func (va *vertexArray) addIndexBuffer(data unsafe.Pointer, count int) {
	va.indexCount = count
	gl.BindVertexArray(va.vao)
	gl.GenBuffers(1, &va.ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, va.ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, count*4, data, gl.STATIC_DRAW)
}

func (va *vertexArray) bind() {
	gl.BindVertexArray(va.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, va.ibo)
}

func (va *vertexArray) unbind() {
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
}

func (va *vertexArray) delete() {
	gl.DeleteVertexArrays(1, &va.vao)
}

func (va *vertexArray) setBufferLayout(bf bufferLayout) {
	va.bind()
	offset := 0
	for i, element := range bf.bufferElements {
		currOffset := offset
		gl.EnableVertexAttribArray(uint32(i))
		gl.VertexAttribPointerWithOffset(uint32(i), element.count, element.glType, false, bf.stride, uintptr(currOffset))
		// gl.VertexAttribPointer(uint32(i), element.count, element.glType, false, bf.stride, gl.Ptr(&offset))
		offset += int(getSizeOfGlType(element.glType, element.count))
	}
}

type bufferLayout struct {
	stride         int32
	bufferElements []bufferElement
}

func (b *bufferLayout) addElement(glType uint32, count int32) {
	be := bufferElement{
		glType: glType,
		count:  count,
	}
	b.bufferElements = append(b.bufferElements, be)
	b.stride += getSizeOfGlType(glType, count)
}

type bufferElement struct {
	glType uint32
	count  int32
}

func getSizeOfGlType(glType uint32, count int32) int32 {
	switch glType {
	case gl.FLOAT:
		return 4 * count
	}
	return 0
}
