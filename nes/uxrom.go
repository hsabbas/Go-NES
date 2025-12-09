package nes

import (
	"log"
)

type uxrom struct {
	prgRom       []byte
	prgBanks     byte
	selectedBank byte
	chr          [kb8]byte
	vram         [0x800]byte
	mirroring    byte
}

func createUxromMapper(rom []byte, prgRom []byte, prg16kbBanks int) *uxrom {
	return &uxrom{
		prgRom:       prgRom,
		prgBanks:     byte(prg16kbBanks),
		selectedBank: 0,
		mirroring:    rom[6] & bit0,
	}
}

func (u *uxrom) cpuRead(address uint16) byte {
	if address < 0x8000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}

	newAddress := int(address)
	if address >= 0xC000 {
		newAddress -= 0xC000
		newAddress += int(u.prgBanks-1) * kb16
	} else if address >= 0x8000 {
		newAddress -= 0x8000
		newAddress += int(u.selectedBank) * kb16
	}
	return u.prgRom[newAddress]
}

func (u *uxrom) cpuWrite(address uint16, value byte) {
	if address < 0x8000 {
		log.Println("Invalid cpu write to address", address)
		return
	}

	if address >= 0x8000 {
		if value >= u.prgBanks {
			log.Println("Cpu tried selecting a PRG bank that doesn't exist.")
			u.selectedBank = u.prgBanks - 1
			return
		}
		u.selectedBank = value
	}
}

func (u *uxrom) ppuRead(address uint16) byte {
	if address >= 0x3000 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		return u.chr[address]
	}

	address = getVramIndex(address, u.mirroring)

	return u.vram[address]
}

func (u *uxrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3000 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		u.chr[address] = value
		return
	}

	address = getVramIndex(address, u.mirroring)

	u.vram[address] = value
}
