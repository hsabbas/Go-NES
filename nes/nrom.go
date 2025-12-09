package nes

import (
	"log"
)

type nrom struct {
	is16Kb    bool
	prgRom    []byte
	chr       []byte
	vram      [0x800]byte
	mirroring byte
}

func createNromMapper(rom []byte, prgRom []byte, chrRom []byte, prg16kbBanks int) *nrom {
	return &nrom{
		is16Kb:    prg16kbBanks == 1,
		prgRom:    prgRom,
		chr:       chrRom,
		mirroring: rom[6] & bit0,
	}
}

func (n *nrom) cpuRead(address uint16) byte {
	if address < 0x8000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}

	address -= 0x8000
	if n.is16Kb {
		address %= 0x4000
	}

	return n.prgRom[address]
}

func (n *nrom) cpuWrite(address uint16, value byte) {
	log.Println("invalid cpu write for NROM")
}

func (n *nrom) ppuRead(address uint16) byte {
	if address >= 0x3000 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < kb8 {
		return n.chr[address]
	}

	address = getVramIndex(address, n.mirroring)

	return n.vram[address]
}

func (n *nrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3000 || address < 0x2000 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	address = getVramIndex(address, n.mirroring)

	n.vram[address] = value
}
