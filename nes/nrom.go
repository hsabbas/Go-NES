package nes

import (
	"log"
)

type nromMapper struct {
	is16Kb         bool
	prgRom         []byte
	chr            [0x2000]byte
	vram           [0x1000]byte
	verticalMirror bool
}

func createNromMapper(rom []byte, prgSize, chrSize int, trainer, verticalMirror bool) *nromMapper {
	prgStart := 16
	if trainer {
		prgStart += 512
	}
	chrAddr := prgStart + prgSize

	var prgRom [0x8000]byte
	var chr [0x2000]byte
	copy(prgRom[:], rom[prgStart:prgStart+prgSize])
	copy(chr[:], rom[chrAddr:chrAddr+chrSize])

	return &nromMapper{
		is16Kb:         prgSize <= 0x4000,
		prgRom:         prgRom[:],
		chr:            chr,
		verticalMirror: verticalMirror,
	}
}

func (n *nromMapper) cpuRead(address uint16) byte {
	if address < 0x8000 {
		log.Println("invalid cpu read for NROM")
		return 0
	}

	address -= 0x8000
	if n.is16Kb {
		address %= 0x4000
	}

	return n.prgRom[address]
}

func (n *nromMapper) cpuWrite(address uint16, value byte) {
	if address < 0x8000 {
		log.Println("invalid cpu read for NROM")
		return
	}

	address -= 0x8000
	if n.is16Kb {
		address %= 0x4000
	}

	n.prgRom[address] = value
}

func (n *nromMapper) ppuRead(address uint16) byte {
	if address >= 0x3000 {
		log.Println("invalid ppu read for NROM")
		return 0
	}

	if address < 0x2000 {
		return n.chr[address]
	}

	if n.verticalMirror {
		if address >= 0x2800 {
			address -= 0x800
		}
	} else {
		if address >= 0x2400 && address < 0x2800 {
			address -= 0x400
		}

		if address >= 0x2C00 {
			address -= 0x400
		}
	}

	address -= 0x2000

	return n.vram[address]
}

func (n *nromMapper) ppuWrite(address uint16, value byte) {
	if address >= 0x3000 {
		log.Println("invalid ppu write for NROM")
		return
	}

	if address < 0x2000 {
		n.chr[address] = value
		return
	}

	if n.verticalMirror {
		if address >= 0x2800 {
			address -= 0x800
		}
	} else {
		if address >= 0x2400 && address < 0x2800 {
			address -= 0x400
		}

		if address >= 0x2C00 {
			address -= 0x400
		}
	}

	address -= 0x2000

	n.vram[address] = value
}
