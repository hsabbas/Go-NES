package nes

type Button byte

const (
	A Button = 1 << iota
	B
	Select
	Start
	Up
	Down
	Left
	Right
)

type controller struct {
	bitsRead     byte
	buttonStates byte
	register     byte
	polling      bool
}

func (c *controller) startPoll() {
	c.bitsRead = 0
	c.register = c.buttonStates
	c.polling = true
}

func (c *controller) stopPoll() {
	c.polling = false
}

func (c *controller) read() byte {
	var val byte
	if c.polling {
		val = c.register & byte(A)
	} else {
		if c.bitsRead >= 8 {
			val = 1
		} else {
			val = (c.register >> c.bitsRead) & bit0
			c.bitsRead++
		}
	}

	return val
}

func (c *controller) updateButton(buttonBit Button, pressed bool) {
	if pressed {
		c.buttonStates |= byte(buttonBit)
	} else {
		c.buttonStates &= ^byte(buttonBit)
	}
	if c.polling {
		c.register = c.buttonStates
	}
}
