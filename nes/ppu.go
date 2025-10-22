package nes

type spritePixel struct {
	isSprite0 bool
	inFront   bool
	value     byte
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
	bgX              int
	backgroundPixels [256]byte

	//sprite rendering
	spritesFound     uint8
	renderingSprite0 bool
	spritePixels     map[int]spritePixel

	cycle    uint16
	scanline uint16
	oddFrame bool

	pixels [240][256]uint16

	sendFrame func([240][256]uint16)
	signalNMI func()
}

func (ppu *ppu) setFrameCallback(callback func([240][256]uint16)) {
	ppu.sendFrame = callback
}

func (ppu *ppu) setNMICallback(callback func()) {
	ppu.signalNMI = callback
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
	val := ppu.dataBuffer
	ppu.dataBuffer = ppu.bus.read(ppu.v)
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

func (ppu *ppu) renderFrame() {
	ppu.scanline = 0
	ppu.cycle = 0
	ppu.oddFrame = !ppu.oddFrame
	ppu.sendFrame(ppu.pixels)
}

func (ppu *ppu) step() {
	if ppu.cycle == 0 {
		if ppu.scanline == 0 {
			ppu.cycle++
			if ppu.oddFrame {
				return
			}
		} else {
			ppu.cycle++
			return
		}
	}

	if ppu.backgroundEnabled {
		ppu.renderBackground()
		ppu.updateV()
	}

	if ppu.scanline < 239 && ppu.spriteEnabled {
		ppu.evaluateSprites()
	}

	// Create pixel
	if ppu.scanline <= 239 && ppu.cycle <= 256 {
		transparent := ppu.bus.read(0x3F00)
		spritePix, ok := ppu.spritePixels[int(ppu.cycle-1)]
		bgPixel := ppu.backgroundPixels[ppu.cycle-1]

		if (ppu.cycle < 8 && !ppu.showSpritesLeft8) || !ppu.spriteEnabled {
			ok = false
		}
		if (ppu.cycle < 8 && !ppu.showBgLeft8) || !ppu.backgroundEnabled {
			bgPixel = transparent
		}

		var pixelVal byte
		if !ok {
			pixelVal = bgPixel
		} else {
			pixelVal = spritePix.value

			if bgPixel != transparent {
				if !spritePix.inFront {
					pixelVal = bgPixel
				}

				if spritePix.isSprite0 {
					ppu.sprite0Hit = true
				}
			}
		}
		colorVal := uint16(pixelVal)
		if ppu.greyscale {
			colorVal &= 0x30
		}

		colorVal |= uint16(ppu.colorEmphasis) << 8

		ppu.pixels[ppu.scanline][ppu.cycle-1] = colorVal

	}

	if ppu.scanline == 241 && ppu.cycle == 1 {
		ppu.vblank = true
		if ppu.vblankNmi {
			ppu.signalNMI()
		}
	}

	if ppu.scanline == 261 {
		if ppu.cycle == 1 {
			ppu.vblank = false
			ppu.sprite0Hit = false
			ppu.spriteOverflow = false
		}

		if ppu.cycle >= 280 && ppu.cycle <= 304 && ppu.renderingEnabled() {
			ppu.v = (ppu.v & 0x1F) | (ppu.v & 0x400) | (ppu.t & 0x3E0) | (ppu.t & 7800)
		}
	}

	ppu.cycle++
	if ppu.cycle > 340 {
		ppu.cycle = 0
		ppu.scanline++
		if ppu.scanline > 261 {
			ppu.renderFrame()
		}
	}
}

func (ppu *ppu) renderBackground() {

	if (ppu.scanline <= 239 && (ppu.cycle <= 256 || ppu.cycle > 320)) || (ppu.scanline == 261 && ppu.cycle > 320) {
		if ppu.cycle%8 == 0 {
			nametableAddr := ppu.v&0xFFF | 0x2000
			attributeAddr := 0x23C0 | (ppu.v & 0x0C00) | ((ppu.v >> 4) & 0x38) | ((ppu.v >> 2) & 0x07)
			tileId := ppu.bus.read(nametableAddr)
			attribute := ppu.bus.read(attributeAddr)
			if ppu.v&0x2 == 2 { // Right half of attribute
				attribute = attribute >> 2
			}
			if ppu.v&0x40 == 0x40 { // Bottom half of attribute
				attribute = attribute >> 4
			}
			palette := attribute & 0x3
			patternAddr := (uint16(tileId) << 4) | (ppu.v >> 12)
			if ppu.rightTableBg {
				patternAddr |= 0x1000
			}
			lowPattern := ppu.bus.read(patternAddr)
			highPattern := ppu.bus.read(patternAddr + 8)
			paletteAddr := 0x3F00 | (uint16(palette) << 2)
			paletteColors := [4]uint8{
				ppu.bus.read(0x3F00),
				ppu.bus.read(paletteAddr + 1),
				ppu.bus.read(paletteAddr + 2),
				ppu.bus.read(paletteAddr + 3),
			}

			for i := 7; i >= 0; i-- {
				patternVal := ((highPattern>>i)&bit0)<<1 | (lowPattern>>i)&bit0

				xCoord := ppu.bgX - int(ppu.x)
				if xCoord >= 0 && xCoord < 256 {
					ppu.backgroundPixels[xCoord] = paletteColors[patternVal]
				}
				ppu.bgX++
			}
		}
	}

	if ppu.cycle == 320 {
		ppu.bgX = 0
	}
}

func (ppu *ppu) evaluateSprites() {
	if ppu.cycle == 1 {
		for i := range 32 {
			ppu.secondaryOam[i] = 0xFF
		}
	}

	if ppu.cycle > 256 && ppu.cycle <= 320 {
		ppu.oamAddr = 0
	}

	// Fill secondary OAM
	if ppu.cycle == 65 {
		start := ppu.oamAddr
		spritesFound := 0
		var n byte = 0
		for n < 64 && spritesFound < 8 {
			addr := 4*n + start
			yCoord := ppu.oam[addr]
			if ppu.scanline < uint16(yCoord) {
				n++
				continue
			}

			fineY := ppu.scanline - uint16(yCoord)
			if (!ppu.bigSprites && fineY < 8) || (ppu.bigSprites && fineY < 16) {
				if n == 0 {
					ppu.renderingSprite0 = true
				}
				secOamAddr := spritesFound * 4
				ppu.secondaryOam[secOamAddr] = ppu.oam[addr]
				ppu.secondaryOam[secOamAddr+1] = ppu.oam[addr+1]
				ppu.secondaryOam[secOamAddr+2] = ppu.oam[addr+2]
				ppu.secondaryOam[secOamAddr+3] = ppu.oam[addr+3]
				spritesFound++
			}
			n++
		}

		//Check for overflow
		var m byte = 0
		for n < 64 {
			addr := 4*n + m + start
			n++
			m = (m + 1) % 4
			yCoord := ppu.oam[addr]
			if ppu.scanline < uint16(yCoord) {
				continue
			}

			fineY := ppu.scanline - uint16(yCoord)
			if (!ppu.bigSprites && fineY < 8) || (ppu.bigSprites && fineY < 16) {
				ppu.spriteOverflow = true
				break
			}
		}

		ppu.spritesFound = uint8(spritesFound)
	}

	// Fetch sprite data
	if ppu.cycle == 257 {
		ppu.spritePixels = make(map[int]spritePixel, 64)
		for n := range ppu.spritesFound {
			addr := n * 4
			yCoord := ppu.secondaryOam[addr]
			xCoord := ppu.secondaryOam[addr+3]
			tileId := ppu.secondaryOam[addr+1]
			attributes := ppu.secondaryOam[addr+2]
			paletteAddr := 0x3F10 | uint16(attributes&3)*4
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
					if fineY < 8 {
						//go to next tile
						fineY = 7 - fineY
						patternAddr += 16
					}
				} else {
					if fineY > 7 {
						//go to next tile
						patternAddr += 16
						fineY -= 8
					}
				}
				patternAddr += fineY
			} else {
				if flipV {
					patternAddr = uint16(tileId)<<4 + (7 - fineY)
				} else {
					patternAddr = uint16(tileId)<<4 + fineY
				}
				if ppu.rightTableSprites {
					paletteAddr |= 0x1000
				}
			}
			lowByte = ppu.bus.read(uint16(patternAddr))
			highByte = ppu.bus.read(uint16(patternAddr) + 8)

			for i := range 8 {
				x := int(xCoord) + i
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
						value:     ppu.bus.read(paletteAddr + uint16(val)),
					}
				}
			}
		}

		ppu.spritesFound = 0
		ppu.renderingSprite0 = false
	}
}

func (ppu *ppu) updateV() {
	if ppu.cycle%8 == 0 && (ppu.cycle < 256 || ppu.cycle > 320) {
		// Increment coarse x of v
		if ppu.v&0x01F == 31 {
			ppu.v = ((ppu.v >> 5) << 5) ^ 0x400
		} else {
			ppu.v++
		}
	}

	// Increment fine and coarse y of v
	if ppu.cycle == 256 {
		if ppu.v&0x7000 == 0x7000 {
			ppu.v = ppu.v & 0xFFF
			coarseY := (ppu.v & 0x03E0) >> 5
			if coarseY == 29 {
				coarseY = 0
				ppu.v ^= 0x0800
			} else if coarseY == 31 {
				coarseY = 0
			} else {
				coarseY++
			}
			ppu.v = (ppu.v & 0x7C1F) | (coarseY << 5)
		} else {
			ppu.v += 0x1000
		}
	}

	// Reset coarse x of v to coarse x of t
	if ppu.cycle == 257 {
		ppu.v = (ppu.v & 0x3E0) | (ppu.v & 0x7800) | (ppu.t & 0x1F) | (ppu.t & 0x400)
	}
}
