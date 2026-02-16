package nes

import (
	"fmt"
	"log"
)

type mapper interface {
	cpuRead(address uint16) byte
	cpuWrite(address uint16, value byte)
	ppuRead(address uint16) byte
	ppuWrite(address uint16, value byte)
}

func createMapper(rom []byte) (mapper, error) {
	cart, err := readCartridge(rom)
	if err != nil {
		log.Fatal("Failed to read cartridge:", err)
	}

	switch cart.mapperNumber {
	case 0:
		return createNromMapper(cart), nil
	case 1:
		return createMMC1Mapper(cart), nil
	case 2:
		return createUxromMapper(cart), nil
	case 7:
		return createAxromMapper(cart), nil
	}

	return nil, fmt.Errorf("unknown mapper number: %d", cart.mapperNumber)
}
