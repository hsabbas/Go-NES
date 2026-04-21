package nes

import "log"

type mmc3 struct {
	c            *cartridge
	prgMode      bool
	chrMode      bool
	selectedBank byte
	banks        []byte
	prgOffsets   []int
	chrOffsets   []int

	irqEnable  bool
	irqReload  byte
	irqCounter byte
}

func createMMC3Mapper(c *cartridge) *mmc3 {
	m := &mmc3{
		c:          c,
		banks:      make([]byte, 8),
		prgOffsets: make([]int, 4),
		chrOffsets: make([]int, 8),
	}
	m.banks[6] = 0
	m.banks[7] = 1
	m.updateOffsets()
	return m
}

func (m *mmc3) cpuRead(address uint16) byte {
	if address < 0x6000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}

	if address < 0x8000 {
		return m.c.prgRam[address-0x6000]
	}

	address -= 0x8000
	bank := address / kb8
	offset := address % kb8
	return m.c.prgRom[m.prgOffsets[bank]+int(offset)]
}

func (m *mmc3) cpuWrite(address uint16, value byte) {
	if address < 0x6000 {
		log.Println("Invalid cpu write to address", address)
		return
	}

	if address < 0x8000 {
		m.c.prgRam[address-0x6000] = value
		return
	}

	even := address%2 == 0
	if address < 0xA000 {
		if even {
			m.writeBankSelect(value)
		} else {
			m.writeBankData(value)
		}
	} else if address < 0xC000 {
		if even {
			m.setMirroring(value)
		} else {
			m.writeRamProtect(value)
		}
	} else if address < 0xE000 {
		if even {
			m.writeIRQLatch(value)
		} else {
			m.writeIRQReload(value)
		}
	} else {
		if even {
			m.writeIRQDisable(value)
		} else {
			m.writeIRQEnable(value)
		}
	}
}

func (m *mmc3) ppuRead(address uint16) byte {
	if address >= 0x3F00 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		bank := address / 0x400
		offset := address % 0x400
		return m.c.chr[m.chrOffsets[bank]+int(offset)]
	}

	return m.c.readVram(address)
}

func (m *mmc3) ppuWrite(address uint16, value byte) {
	if address >= 0x3F00 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		bank := address / 0x400
		offset := address % 0x400
		m.c.chr[m.chrOffsets[bank]+int(offset)] = value
		return
	}

	m.c.writeVram(address, value)
}

func (m *mmc3) writeBankSelect(value byte) {
	m.chrMode = hasBit7(value)
	m.prgMode = hasBit6(value)
	m.selectedBank = value & 7

	m.updateOffsets()
}

func (m *mmc3) writeBankData(value byte) {
	if m.selectedBank > 5 {
		m.banks[m.selectedBank] = value & 0x3F
	}
	if m.selectedBank < 3 {
		m.banks[m.selectedBank] = value & 0xFE
	} else {
		m.banks[m.selectedBank] = value
	}
	m.updateOffsets()
}

func (m *mmc3) setMirroring(value byte) {
	if hasBit0(value) {
		m.c.mirroring = horizontalMirroring
	} else {
		m.c.mirroring = verticalMirroring
	}
}

func (m *mmc3) writeRamProtect(value byte) {
	// Do nothing?
	// https://www.nesdev.org/wiki/MMC3#iNES_Mapper_004_and_MMC6
}

func (m *mmc3) writeIRQLatch(value byte) {
	m.irqReload = value
}

func (m *mmc3) writeIRQReload(value byte) {
	m.irqCounter = 0
}

func (m *mmc3) writeIRQDisable(value byte) {
	m.irqEnable = false
}

func (m *mmc3) writeIRQEnable(value byte) {
	m.irqEnable = true
}

func (m *mmc3) step() {
	ppu := m.c.nes.ppu
	if !ppu.renderingEnabled() || (ppu.scanline >= 241 && ppu.scanline <= 260) {
		return
	}
	if !ppu.rightTableBg && ppu.rightTableSprites && ppu.cycle == 260 {
		m.clockIRQ()
	}
	if ppu.rightTableBg && !ppu.rightTableSprites && ppu.cycle == 324 {
		m.clockIRQ()
	}
}

func (m *mmc3) clockIRQ() {
	if m.irqCounter == 0 {
		m.irqCounter = m.irqReload
		return
	}

	m.irqCounter--
	if m.irqCounter == 0 {
		if m.irqEnable {
			m.c.nes.cpu.sendIRQ()
		}
	}
}

func (m *mmc3) updateOffsets() {
	if m.prgMode {
		m.prgOffsets[0] = kb16 * (m.c.prgBanks - 1)
		m.prgOffsets[1] = kb8 * int(m.banks[7])
		m.prgOffsets[2] = kb8 * int(m.banks[6])
		m.prgOffsets[3] = kb8 + m.prgOffsets[0]
	} else {
		m.prgOffsets[0] = kb8 * int(m.banks[6])
		m.prgOffsets[1] = kb8 * int(m.banks[7])
		m.prgOffsets[2] = kb16 * (m.c.prgBanks - 1)
		m.prgOffsets[3] = kb8 + m.prgOffsets[2]
	}

	if m.chrMode {
		m.chrOffsets[0] = 0x400 * int(m.banks[2])
		m.chrOffsets[1] = 0x400 * int(m.banks[3])
		m.chrOffsets[2] = 0x400 * int(m.banks[4])
		m.chrOffsets[3] = 0x400 * int(m.banks[5])
		m.chrOffsets[4] = 0x400 * int(m.banks[0]&0xFE)
		m.chrOffsets[5] = 0x400 + m.chrOffsets[4]
		m.chrOffsets[6] = 0x400 * int(m.banks[1]&0xFE)
		m.chrOffsets[7] = 0x400 + m.chrOffsets[6]
	} else {
		m.chrOffsets[0] = 0x400 * int(m.banks[0]&0xFE)
		m.chrOffsets[1] = 0x400 + m.chrOffsets[0]
		m.chrOffsets[2] = 0x400 * int(m.banks[1]&0xFE)
		m.chrOffsets[3] = 0x400 + m.chrOffsets[2]
		m.chrOffsets[4] = 0x400 * int(m.banks[2])
		m.chrOffsets[5] = 0x400 * int(m.banks[3])
		m.chrOffsets[6] = 0x400 * int(m.banks[4])
		m.chrOffsets[7] = 0x400 * int(m.banks[5])
	}
}
