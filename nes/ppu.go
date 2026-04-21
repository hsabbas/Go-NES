package nes

type spritePixel struct {
	isSprite0 bool
	inFront   bool
	value     byte
	palette   byte
}

type ppu struct {
	bus          *ppuBus
	oam          [256]byte
	secondaryOam [32]byte

	//Control Register
	currentNametable  byte
	vramAdrInc        bool
	rightTableSprites bool
	rightTableBg      bool
	bigSprites        bool
	msSelect          bool
	vblankNmi         bool

	//Mask Register
	greyscale         bool
	showBgLeft8       bool
	showSpritesLeft8  bool
	backgroundEnabled bool
	spriteEnabled     bool
	colorEmphasis     byte

	//Status Register
	spriteOverflow bool
	sprite0Hit     bool
	vblank         bool

	//Sprite RAM Address
	oamAddr byte

	//VRAM Data
	dataBuffer byte

	//Internal Registers
	v uint16
	t uint16
	x byte
	w bool

	//background rendering
	backgroundValues   [256]byte
	backgroundPalettes [256]byte
	tileId             byte
	attribute          byte
	patternAddr        uint16
	lowPattern         byte
	highPattern        byte

	//sprite rendering
	spritesFound     uint8
	renderingSprite0 bool
	spritePixels     map[uint16]spritePixel

	cycle     uint16
	scanline  uint16
	evenFrame bool

	paletteVals [240][256]uint16
	frame       [240][256 * 3]byte

	signalNMI func()
}

func createPPU(pb *ppuBus) *ppu {
	return &ppu{
		bus: pb,
	}
}

func (ppu *ppu) setNMICallback(callback func()) {
	ppu.signalNMI = callback
}

func (ppu *ppu) writeRegister(address uint16, value byte) {
	address = 0x2000 + (address % 8)
	switch address {
	case 0x2000:
		ppu.writeToPPUCTRL(value)
	case 0x2001:
		ppu.writeToPPUMASK(value)
	case 0x2003:
		ppu.writeToOAMADDR(value)
	case 0x2004:
		ppu.writeToOAMDATA(value)
	case 0x2005:
		ppu.writeToPPUSCROLL(value)
	case 0x2006:
		ppu.writeToPPUADDR(value)
	case 0x2007:
		ppu.writeToPPUDATA(value)
	}
}

func (ppu *ppu) writeToPPUCTRL(data byte) {
	vbNmiWasSet := ppu.vblankNmi

	ppu.currentNametable = data & (bit0 + bit1)
	ppu.vramAdrInc = hasBit2(data)
	ppu.rightTableSprites = hasBit3(data)
	ppu.rightTableBg = hasBit4(data)
	ppu.bigSprites = hasBit5(data)
	ppu.msSelect = hasBit6(data)
	ppu.vblankNmi = hasBit7(data)

	if !vbNmiWasSet && ppu.vblankNmi && ppu.vblank {
		ppu.signalNMI()
	}

	ppu.t = ppu.t&0x7000 + uint16(ppu.currentNametable)<<10 + ppu.t&0x3ff
}

func (ppu *ppu) writeToPPUMASK(data byte) {
	ppu.greyscale = hasBit0(data)
	ppu.showBgLeft8 = hasBit1(data)
	ppu.showSpritesLeft8 = hasBit2(data)
	ppu.backgroundEnabled = hasBit3(data)
	ppu.spriteEnabled = hasBit4(data)
	ppu.colorEmphasis = data >> 5
}

func (ppu *ppu) writeToPPUSCROLL(data byte) {
	if !ppu.w {
		ppu.t = ppu.t&0b111111111100000 + uint16(data>>3)
		ppu.x = data & 0b111
		ppu.w = true
	} else {
		ppu.t = ppu.t & 0b000110000011111
		ppu.t += uint16(data&0b111)<<12 + uint16(data&0b11111000)<<2
		ppu.w = false
	}
}

func (ppu *ppu) writeToPPUADDR(data byte) {
	if !ppu.w {
		ppu.t = ppu.t&0b000000011111111 + uint16(data&0b00111111)<<8
		ppu.w = true
	} else {
		ppu.t = ppu.t&0b111111100000000 + uint16(data)
		ppu.v = ppu.t
		ppu.w = false
	}
}

func (ppu *ppu) writeToOAMADDR(data byte) {
	ppu.oamAddr = data
}

func (ppu *ppu) writeToOAMDATA(data byte) {
	ppu.oam[ppu.oamAddr] = data
	ppu.oamAddr++
}

func (ppu *ppu) readPPUSTATUS() byte {
	ppustatus := byte(0x1D)
	if ppu.spriteOverflow {
		ppustatus |= bit5
	}
	if ppu.sprite0Hit {
		ppustatus |= bit6
	}
	if ppu.vblank {
		ppustatus |= bit7
	}
	ppu.vblank = false
	ppu.w = false
	return ppustatus
}

func (ppu *ppu) readOAMDATA() byte {
	return ppu.oam[ppu.oamAddr]
}

func (ppu *ppu) writeToPPUDATA(data byte) {
	ppu.bus.write(ppu.v, data)
	if ppu.vramAdrInc {
		ppu.v += 32
	} else {
		ppu.v += 1
	}
}

func (ppu *ppu) readPPUDATA() byte {
	var val byte
	if ppu.v >= 0x3F00 {
		ppu.dataBuffer = ppu.bus.read(ppu.v - 0x1000)
		val = ppu.bus.read(ppu.v)
	} else {
		val = ppu.dataBuffer
		ppu.dataBuffer = ppu.bus.read(ppu.v)
	}
	if ppu.vramAdrInc {
		ppu.v += 32
	} else {
		ppu.v += 1
	}
	return val
}

func (ppu *ppu) renderingEnabled() bool {
	return ppu.backgroundEnabled || ppu.spriteEnabled
}

func (ppu *ppu) endFrame() {
	ppu.scanline = 0
	ppu.cycle = 0
	ppu.evenFrame = !ppu.evenFrame

	for y := 0; y < 240; y++ {
		for x, val := range ppu.paletteVals[y] {
			ppu.frame[y][x*3] = colorMap[val].r
			ppu.frame[y][x*3+1] = colorMap[val].g
			ppu.frame[y][x*3+2] = colorMap[val].b
		}
	}
}

func (ppu *ppu) step() bool {
	if ppu.cycle == 0 {
		ppu.cycle++
		if !(ppu.scanline == 0 && ppu.evenFrame && ppu.renderingEnabled()) {
			return false
		}
	}

	visibleFrame := ppu.scanline <= 239
	prerenderLine := ppu.scanline == 261
	renderLine := visibleFrame || prerenderLine
	earlyFetchCycle := ppu.cycle >= 321 && ppu.cycle <= 336
	visibleCycle := ppu.cycle <= 256
	fetchCycle := earlyFetchCycle || visibleCycle

	if visibleFrame && ppu.cycle == 1 {
		for i := range 32 {
			ppu.secondaryOam[i] = 0xFF
		}
	}

	if ppu.renderingEnabled() {
		if renderLine && fetchCycle {
			ppu.renderBackground()
		}

		if renderLine && ppu.cycle%8 == 0 && fetchCycle {
			ppu.incrementCoarseX()
		}

		if renderLine && ppu.cycle == 256 {
			ppu.incrementCoarseY()
		}

		// Reset coarse x of v to coarse x of t
		if renderLine && ppu.cycle == 257 {
			ppu.v = (ppu.v & 0x7BE0) | (ppu.t & 0x41F)
		}

		// Reset coarse y of v to coarse y of t
		if prerenderLine && ppu.cycle >= 280 && ppu.cycle <= 304 {
			ppu.v = (ppu.v & 0x41F) | (ppu.t & 0x7BE0)
		}

		if visibleFrame && ppu.cycle == 65 {
			ppu.fillSecondaryOAM()
		}
		if renderLine {
			if ppu.cycle == 257 {
				ppu.fetchSpriteData()
			}
			if ppu.cycle >= 257 && ppu.cycle <= 320 {
				ppu.oamAddr = 0
			}
		}

		if visibleFrame && visibleCycle {
			ppu.createPixel()
		}
	}

	if ppu.scanline == 241 && ppu.cycle == 1 {
		ppu.vblank = true
		if ppu.vblankNmi {
			ppu.signalNMI()
		}
	}

	if prerenderLine {
		if ppu.cycle == 1 {
			ppu.spritePixels = make(map[uint16]spritePixel, 64)
			ppu.vblank = false
			ppu.sprite0Hit = false
			ppu.spriteOverflow = false
		}
	}

	if !fetchCycle {
		ppu.oamAddr = 0
	}

	ppu.cycle++
	if ppu.cycle > 340 {
		ppu.cycle = 0
		ppu.scanline++
		if ppu.scanline > 261 {
			ppu.endFrame()
			return true
		}
	}
	return false
}

func (ppu *ppu) renderBackground() {
	switch ppu.cycle % 8 {
	case 1:
		nametableAddr := ppu.v&0xFFF | 0x2000
		ppu.tileId = ppu.bus.read(nametableAddr)
	case 3:
		attributeAddr := 0x23C0 | (ppu.v & 0x0C00) | ((ppu.v >> 4) & 0x38) | ((ppu.v >> 2) & 0x07)
		ppu.attribute = ppu.bus.read(attributeAddr)
	case 5:
		ppu.patternAddr = (uint16(ppu.tileId) << 4) | (ppu.v >> 12)
		if ppu.rightTableBg {
			ppu.patternAddr |= 0x1000
		}
		ppu.lowPattern = ppu.bus.read(ppu.patternAddr)
	case 7:
		ppu.highPattern = ppu.bus.read(ppu.patternAddr + 8)
		attribute := ppu.attribute
		if ppu.v&0x2 == 2 { // Right half of attribute
			attribute = attribute >> 2
		}
		if ppu.v&0x40 == 0x40 { // Bottom half of attribute
			attribute = attribute >> 4
		}
		palette := attribute & 0x3

		var xCoord int
		if ppu.cycle > 320 {
			xCoord = int(ppu.cycle) - 327 - int(ppu.x)
		} else {
			xCoord = int(ppu.cycle) + 9 - int(ppu.x)
		}
		for i := 7; i >= 0; i-- {
			patternVal := ((ppu.highPattern>>i)&bit0)<<1 | (ppu.lowPattern>>i)&bit0

			if xCoord >= 0 && xCoord < 256 {
				ppu.backgroundValues[xCoord] = patternVal
				ppu.backgroundPalettes[xCoord] = palette
			}
			xCoord++
		}
	}
}

func (ppu *ppu) fillSecondaryOAM() {
	start := ppu.oamAddr
	ppu.spritesFound = 0
	ppu.renderingSprite0 = false
	var m byte = 0
	for n := range 64 {
		ppu.oamAddr = 4*byte(n) + m + start
		if ppu.spritesFound >= 8 {
			m = (m + 1) % 4
		}

		yCoord := ppu.oam[ppu.oamAddr]
		if ppu.scanline < uint16(yCoord) {
			continue
		}

		fineY := ppu.scanline - uint16(yCoord)
		if (!ppu.bigSprites && fineY < 8) || (ppu.bigSprites && fineY < 16) {
			if n == 0 {
				ppu.renderingSprite0 = true
			}
			if ppu.spritesFound < 8 {
				secOamAddr := ppu.spritesFound * 4
				ppu.secondaryOam[secOamAddr] = ppu.oam[ppu.oamAddr]
				ppu.secondaryOam[secOamAddr+1] = ppu.oam[ppu.oamAddr+1]
				ppu.secondaryOam[secOamAddr+2] = ppu.oam[ppu.oamAddr+2]
				ppu.secondaryOam[secOamAddr+3] = ppu.oam[ppu.oamAddr+3]
				ppu.spritesFound++
			} else {
				ppu.spriteOverflow = true
				break
			}
		}
	}
}

func (ppu *ppu) fetchSpriteData() {
	ppu.spritePixels = make(map[uint16]spritePixel, 64)
	for n := range ppu.spritesFound {
		addr := n * 4
		yCoord := ppu.secondaryOam[addr]
		tileId := ppu.secondaryOam[addr+1]
		attributes := ppu.secondaryOam[addr+2]
		xCoord := ppu.secondaryOam[addr+3]
		priority := hasBit5(attributes)
		flipH := hasBit6(attributes)
		flipV := hasBit7(attributes)
		var patternAddr uint16
		var lowByte byte
		var highByte byte
		fineY := ppu.scanline - uint16(yCoord)
		if ppu.bigSprites {
			patternAddr = uint16(tileId & ^bit0)<<4 | uint16(tileId&bit0)<<12
			if flipV {
				fineY = 15 - fineY
				}
				if fineY > 7 {
				// go to next tile
					patternAddr += 16
					fineY -= 8
			}
			patternAddr += fineY
		} else {
			if flipV {
				patternAddr = uint16(tileId)<<4 + (7 - fineY)
			} else {
				patternAddr = uint16(tileId)<<4 + fineY
			}
			if ppu.rightTableSprites {
				patternAddr |= 0x1000
			}
		}
		lowByte = ppu.bus.read(uint16(patternAddr))
		highByte = ppu.bus.read(uint16(patternAddr) + 8)

		for i := range 8 {
			x := uint16(xCoord) + uint16(i)
			if _, ok := ppu.spritePixels[x]; ok || x > 255 {
				continue
			}

			var val, lowBit, highBit byte
			if flipH {
				lowBit = (lowByte & (1 << i)) >> i
				highBit = (highByte & (1 << i)) >> i
			} else {
				offset := 7 - i
				lowBit = (lowByte & (1 << offset)) >> byte(offset)
				highBit = (highByte & (1 << offset)) >> byte(offset)
			}
			val = highBit<<1 + lowBit

			if val > 0 {
				ppu.spritePixels[x] = spritePixel{
					isSprite0: n == 0 && ppu.renderingSprite0,
					inFront:   !priority,
					value:     val,
					palette:   attributes & 3,
				}
			}
		}
	}
}

func (ppu *ppu) createPixel() {
	x := ppu.cycle - 1
	spritePixelData, ok := ppu.spritePixels[x]
	backgroundValue := ppu.backgroundValues[x]
	spritePaletteAddr := 0x3F10 | uint16(spritePixelData.palette)*4
	backgroundPaletteAddr := 0x3F00 | (uint16(ppu.backgroundPalettes[x]) * 4)

	if (x < 8 && !ppu.showSpritesLeft8) || !ppu.spriteEnabled {
		ok = false
	}
	if (x < 8 && !ppu.showBgLeft8) || !ppu.backgroundEnabled {
		backgroundValue = 0
	}

	backgroundColor := ppu.bus.read(backgroundPaletteAddr + uint16(backgroundValue))
	spriteColor := ppu.bus.read(spritePaletteAddr + uint16(spritePixelData.value))
	var pixel byte
	if !ok || spritePixelData.value == 0 {
		pixel = backgroundColor
	} else {
		pixel = spriteColor
		if backgroundValue != 0 {
			if !spritePixelData.inFront {
				pixel = backgroundColor
			}

			if spritePixelData.isSprite0 {
				ppu.sprite0Hit = true
			}
		}
	}
	colorVal := uint16(pixel)
	if ppu.greyscale {
		colorVal &= 0x30
	}

	colorVal |= uint16(ppu.colorEmphasis) << 8

	ppu.paletteVals[ppu.scanline][x] = colorVal
}

func (ppu *ppu) incrementCoarseX() {
	if ppu.v&0x01F == 31 {
		ppu.v = ((ppu.v >> 5) << 5) ^ 0x400
	} else {
		ppu.v++
	}
}

func (ppu *ppu) incrementCoarseY() {
	if ppu.v&0x7000 == 0x7000 {
		ppu.v = ppu.v & 0xFFF
		coarseY := (ppu.v & 0x03E0) >> 5
		switch coarseY {
		case 29:
			coarseY = 0
			ppu.v ^= 0x0800
		case 31:
			coarseY = 0
		default:
			coarseY++
		}
		ppu.v = (ppu.v & 0xFC1F) | (coarseY << 5)
	} else {
		ppu.v += 0x1000
	}
}
