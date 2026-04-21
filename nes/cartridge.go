package nes

import (
	"fmt"
	"log"
)

type nametableArrangement uint8

const (
	horizontalMirroring nametableArrangement = iota
	verticalMirroring
	oneScreenLower
	oneScreenUpper
)

type cartridge struct {
	nes          *NES
	mapperNumber int
	prgBanks     int
	chrBanks     int
	prgRom       []byte
	prgRam       []byte
	chr          []byte
	mirroring    nametableArrangement
	vram         [0x800]byte
}

func readCartridge(nes *NES, rom []byte) (*cartridge, error) {
	if rom[0] != 0x4E || rom[1] != 0x45 || rom[2] != 0x53 || rom[3] != 0x1A {
		return nil, fmt.Errorf("malformed .NES header")
	}

	hasTrainer := hasBit2(rom[6])
	if hasTrainer {
		return nil, fmt.Errorf("trainers are not supported")
	}

	isNES20 := !hasBit2(rom[7]) && hasBit3(rom[7])
	mapperNum := int(rom[6]&0xF0>>4) + int(rom[7]&0xF0)
	prgBanks := int(rom[4])
	chrBanks := int(rom[5])
	hasPrgRam := hasBit2(rom[6])

	prgSize := prgBanks * kb16
	chrSize := chrBanks * kb8

	log.Println("--Cartridge Info--")
	if isNES20 {
		log.Println(".NES 2.0 headers are not supported. Game may not run as expected.")
	}
	log.Println("Mapper Number:", mapperNum)
	log.Println("Total size:", len(rom), "bytes")
	log.Println("16 KB PRG Rom banks:", prgBanks)
	log.Println("8 KB CHR Rom banks:", chrBanks)
	if chrBanks == 0 {
		log.Println("Using 8 KB of CHR Ram")
	}
	if hasPrgRam {
		log.Println("Has battery backed RAM")
	}

	prgStart := 16
	chrStart := prgStart + prgSize
	prg := rom[prgStart:chrStart]
	chr := rom[chrStart : chrStart+chrSize]
	if len(chr) == 0 {
		chr = make([]byte, kb8)
	}

	return &cartridge{
		nes:          nes,
		mapperNumber: mapperNum,
		prgBanks:     prgBanks,
		chrBanks:     chrBanks,
		prgRom:       prg,
		prgRam:       make([]byte, kb8),
		chr:          chr,
		mirroring:    nametableArrangement(rom[6] & bit0),
	}, nil
}

func (c *cartridge) readVram(address uint16) byte {
	return c.vram[c.getVramIndex(address)]
}

func (c *cartridge) writeVram(address uint16, value byte) {
	c.vram[c.getVramIndex(address)] = value
}

func (c *cartridge) getVramIndex(address uint16) uint16 {
	address = (address - 0x2000) % 0x1000
	switch c.mirroring {
	case horizontalMirroring:
		if address >= 0x400 {
			address -= 0x400
		}

		if address >= 0x800 {
			address -= 0x400
		}
		return address
	case verticalMirroring:
		if address >= 0x800 {
			address -= 0x800
		}
		return address
	case oneScreenLower:
		return address % 0x400
	case oneScreenUpper:
		return address%0x400 + 0x400
	}
	log.Fatal("Invalid nametable arrangement.")
	return address
}
