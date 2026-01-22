package nes

import (
	"fmt"
	"log"
)

const (
	horizontalMirroring = iota
	verticalMirroring
)

const (
	kb4  = 0x1000
	kb8  = 0x2000
	kb16 = 0x4000
	kb32 = 0x8000

	headerSize = 16
)

type mapper interface {
	cpuRead(address uint16) byte
	cpuWrite(address uint16, value byte)
	ppuRead(address uint16) byte
	ppuWrite(address uint16, value byte)
}

func mapCartridge(rom []byte) (mapper, error) {
	if rom[0] != 0x4E || rom[1] != 0x45 || rom[2] != 0x53 || rom[3] != 0x1A {
		return nil, fmt.Errorf("malformed .NES header")
	}

	hasTrainer := hasBit2(rom[6])
	isNES20 := !hasBit2(rom[7]) && hasBit3(rom[7])
	mapperNum := int(rom[6]&0xF0>>4) + int(rom[7]&0xF0)
	prgBanks := int(rom[4])
	chrBanks := int(rom[5])

	if hasTrainer {
		return nil, fmt.Errorf("trainers are not supported")
	}

	if isNES20 {
		log.Println(".NES 2.0 headers are not supported. Game may not run as expected.")
	}

	prgSize := prgBanks * kb16
	chrSize := chrBanks * kb8

	log.Println("ROM mapper:", mapperNum)
	log.Println("ROM size:", len(rom), "bytes")
	log.Println("PRG Rom size:", prgSize, "bytes")
	log.Println("CHR Rom size:", chrSize, "bytes")

	prgStart := headerSize
	chrStart := prgStart + prgSize
	prgRom := rom[prgStart:chrStart]
	chrRom := rom[chrStart : chrStart+chrSize]

	if mapperNum == 0 {
		return createNromMapper(rom, prgRom, chrRom, prgBanks), nil
	}
	if mapperNum == 2 {
		return createUxromMapper(rom, prgRom, prgBanks), nil
	}
	if mapperNum == 7 {
		return createAxromMapper(prgRom, prgBanks), nil
	}

	return nil, fmt.Errorf("unknown mapper number: %d", mapperNum)
}

func getVramIndex(address uint16, mirroring byte) uint16 {
	if mirroring == verticalMirroring {
		if address >= 0x2800 {
			address -= 0x800
		}
		return address - 0x2000
	}

	if address >= 0x2400 {
		address -= 0x400
	}

	if address >= 0x2800 {
		address -= 0x400
	}

	return address - 0x2000
}
