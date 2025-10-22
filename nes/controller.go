package nes

const (
	A uint8 = 1 << iota
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
		val = c.register & A
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

func (c *controller) updateButton(buttonBit byte, pressed bool) {
	if pressed {
		c.buttonStates |= buttonBit
	} else {
		c.buttonStates &= ^buttonBit
	}
	if c.polling {
		c.register = c.buttonStates
	}
}
