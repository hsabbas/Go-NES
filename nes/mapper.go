package nes

import "fmt"

type mapper interface {
	cpuRead(address uint16) byte
	cpuWrite(address uint16, value byte)
	ppuRead(address uint16) byte
	ppuWrite(address uint16, value byte)
}

const (
	cpuMinAddr = 0x4020
)

func mapCartridge(rom []byte) (mapper, error) {
	if rom[0] != 0x4E || rom[1] != 0x45 || rom[2] != 0x53 || rom[3] != 0x1A {
		return nil, fmt.Errorf("malformed .NES header")
	}

	fmt.Println("Header bytes 4-7")
	fmt.Printf("Byte 4: %b\n", rom[4])
	fmt.Printf("Byte 5: %b\n", rom[5])
	fmt.Printf("Byte 6: %b\n", rom[6])
	fmt.Printf("Byte 7: %b\n", rom[7])

	var prgSize, chrSize, mapperNum int
	prgSize = int(rom[4]) * 0x4000
	chrSize = int(rom[5]) * 0x2000
	mapperNum = int(rom[6]&0xF0>>4) + int(rom[7]&0xF0)

	fmt.Println("PRG-ROM size:", prgSize)
	fmt.Println("CHR-ROM size:", chrSize)
	fmt.Println("Mapper number:", mapperNum)

	verticalMirroring := rom[6]&bit0 == bit0
	trainer := rom[6]&bit2 == bit2

	if mapperNum == 0 {
		return createNromMapper(rom, prgSize, chrSize, trainer, verticalMirroring), nil
	}

	return nil, fmt.Errorf("unknown mapper number: %d", mapperNum)
}
