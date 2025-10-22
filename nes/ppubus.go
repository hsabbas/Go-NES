package nes

import "fmt"

type ppuBus struct {
	palette [0x20]byte
	m       mapper
}

func (pb *ppuBus) read(address uint16) byte {
	if address > 0x3FFF {
		fmt.Println("Invalid PPU read")
		return 0
	}
	if address < 0x3F00 {
		return pb.m.ppuRead(address)
	} else {
		i := (address - 0x3F00) % 0x20
		if i%4 == 0 {
			return pb.palette[0]
		}
		return pb.palette[i]
	}
}

func (pb *ppuBus) write(address uint16, value byte) {
	if address > 0x3FFF {
		fmt.Println("Invalid PPU write")
		return
	}
	if address < 0x3F00 {
		pb.m.ppuWrite(address, value)
	} else {
		i := (address - 0x3F00) % 0x20
		if i == 0 || i == 0x10 {
			pb.palette[0] = value
		} else if i%4 != 0 {
			pb.palette[i] = value
		}
	}
}
