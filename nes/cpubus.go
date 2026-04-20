package nes

type cpuBus struct {
	internalRam [0x800]byte
	m           mapper
	ppu         *ppu
	apu         *apu
	controller1 *controller
	controller2 *controller
}

func (cb *cpuBus) read(address uint16) byte {
	if address < 0x2000 {
		return cb.internalRam[address%0x800]
	}

	// PPU registers
	if address <= 0x3FFF {
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
		return cb.apu.readStatus()
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
	} else if address <= 0x3FFF {
		cb.ppu.writeRegister(address, value)
	} else if address == 0x4014 {
		cb.oamdma(value)
	} else if address == 0x4016 {
		if value&bit0 == bit0 {
			cb.controller1.startPoll()
			cb.controller2.startPoll()
		} else {
			cb.controller1.stopPoll()
			cb.controller2.stopPoll()
		}
	} else if address <= 0x4017 {
		cb.apu.writeRegister(address, value)
	} else if address >= 0x4020 {
		cb.m.cpuWrite(address, value)
	}
}

func (cb *cpuBus) oamdma(page byte) {
	next := uint16(page) << 8
	for i := 0; i < 256; i++ {
		cb.ppu.writeToOAMDATA(cb.internalRam[next])
		next++
	}
}
