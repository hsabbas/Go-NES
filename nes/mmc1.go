package nes

import (
	"log"
)

type mmc1 struct {
	c *cartridge

	writeCount int

	// Internal Registers
	shiftRegister byte
	control       byte
	prgMode       byte
	twoCHRbanks   bool
	chrBank0      byte
	chrBank1      byte
	prgBank       byte

	prgOffsets []int
	chrOffsets []int
}

func createMMC1Mapper(c *cartridge) *mmc1 {
	m := mmc1{
		c:          c,
		prgOffsets: make([]int, 2),
		chrOffsets: make([]int, 2),
	}
	m.prgOffsets[1] = (m.c.prgBanks - 1) * kb16
	return &m
}

func (m *mmc1) cpuRead(address uint16) byte {
	if address < 0x6000 {
		log.Println("Invalid cpu read for MMC1")
		return 0
	}
	if address < 0x8000 {
		return m.c.prgRam[address-0x6000]
	}

	address -= 0x8000

	bank := address / kb16
	offset := address % kb16
	return m.c.prgRom[m.prgOffsets[bank]+int(offset)]
}

func (m *mmc1) cpuWrite(address uint16, value byte) {
	if address < 0x6000 {
		log.Println("Invalid cpu write for MMC1")
		return
	}

	if address < 0x8000 {
		m.c.prgRam[address-0x6000] = value
		return
	}

	if hasBit7(value) {
		// reset
		m.shiftRegister = 0
		m.writeCount = 0
		m.writeControl(m.control | 0x0C)
		return
	}

	m.shiftRegister |= (value & bit0) << m.writeCount

	m.writeCount++
	if m.writeCount == 5 {
		m.writeCount = 0
		//write to register
		if address < 0xA000 {
			m.writeControl(m.shiftRegister)
		} else if address < 0xC000 {
			m.chrBank0 = m.shiftRegister
		} else if address < 0xE000 {
			m.chrBank1 = m.shiftRegister
		} else {
			m.prgBank = m.shiftRegister & 0xF
		}
		m.shiftRegister = 0
		m.updateOffsets()
	}
}

func (m *mmc1) ppuRead(address uint16) byte {
	if address >= 0x3F00 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		bank := address / kb4
		offset := address % kb4
		return m.c.chr[m.chrOffsets[bank]+int(offset)]
	}

	return m.c.readVram(address)
}

func (m *mmc1) ppuWrite(address uint16, value byte) {
	if address >= 0x3F00 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		bank := address / kb4
		offset := address % kb4
		m.c.chr[m.chrOffsets[bank]+int(offset)] = value
		return
	}

	m.c.writeVram(address, value)
}

func (m *mmc1) writeControl(value byte) {
	m.control = value
	switch value & (bit0 | bit1) {
	case 0:
		m.c.mirroring = oneScreenLower
	case 1:
		m.c.mirroring = oneScreenUpper
	case 2:
		m.c.mirroring = verticalMirroring
	case 3:
		m.c.mirroring = horizontalMirroring
	}
	m.prgMode = (value >> 2) & 3
	if m.prgMode == 0 {
		m.prgMode++
	}
	m.twoCHRbanks = hasBit4(value)
}

func (m *mmc1) updateOffsets() {
	switch m.prgMode {
	case 0, 1:
		m.prgOffsets[0] = kb16 * int(m.prgBank&0xFE)
		m.prgOffsets[1] = m.prgOffsets[0] + kb16
	case 2:
		m.prgOffsets[0] = 0
		m.prgOffsets[1] = int(m.prgBank) * kb16
	case 3:
		m.prgOffsets[0] = int(m.prgBank) * kb16
		m.prgOffsets[1] = (m.c.prgBanks - 1) * kb16
	}

	if m.twoCHRbanks {
		m.chrOffsets[0] = int(m.chrBank0) * kb4
		m.chrOffsets[1] = int(m.chrBank1) * kb4
	} else {
		m.chrOffsets[0] = int(m.chrBank0&0xFE) * kb4
		m.chrOffsets[1] = m.chrOffsets[0] + kb4
	}
}

func (m *mmc1) step() {}
