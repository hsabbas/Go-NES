package nes

type cpuBus struct {
	internalRam [0x800]byte
	m           mapper
	ppu         *ppu
	controller1 *controller
	controller2 *controller
}

func (cb *cpuBus) read(address uint16) byte {
	if address < 0x2000 {
		return cb.internalRam[address%0x800]
	}

	// PPU registers
	if address < 0x3FFF {
		address = 0x2000 + (address % 8)
	}

	if address == 0x2002 {
		return cb.ppu.readPPUSTATUS()
	}
	if address == 0x2004 {
		return cb.ppu.readOAMDATA()
	}
	if address == 0x2007 {
		return cb.ppu.readPPUDATA()
	}

	//APU registers
	if address == 0x4015 {
		return 0
	}

	//Controller registers
	if address == 0x4016 {
		return cb.controller1.read()
	}
	if address == 0x4017 {
		return cb.controller2.read()
	}

	//External memory
	if address >= 0x4020 {
		return cb.m.cpuRead(address)
	}

	return 0
}

func (cb *cpuBus) write(address uint16, value byte) {
	if address < 0x2000 {
		cb.internalRam[address%0x800] = value
		return
	}

	if address <= 0x3FFF {
		address = 0x2000 + (address % 8)
	}

	//PPU Registers
	if address == 0x2000 {
		cb.ppu.writeToPPUCTRL(value)
	}
	if address == 0x2001 {
		cb.ppu.writeToPPUMASK(value)
	}
	if address == 0x2003 {
		cb.ppu.writeToOAMADDR(value)
	}
	if address == 0x2004 {
		cb.ppu.writeToOAMDATA(value)
	}
	if address == 0x2005 {
		cb.ppu.writeToPPUSCROLL(value)
	}
	if address == 0x2006 {
		cb.ppu.writeToPPUADDR(value)
	}
	if address == 0x2007 {
		cb.ppu.writeToPPUDATA(value)
	}

	//OAMDMA
	if address == 0x4014 {
		next := uint16(value) << 8
		for i := 0; i < 256; i++ {
			cb.ppu.writeToOAMDATA(cb.internalRam[next])
			next++
		}
	}

	//Controller registers
	if address == 0x4016 {
		if value&bit0 == bit0 {
			cb.controller1.startPoll()
			cb.controller2.startPoll()
		} else {
			cb.controller1.stopPoll()
			cb.controller2.stopPoll()
		}
	}

	//External memory
	if address >= 0x4020 {
		cb.m.cpuWrite(address, value)
	}
}
