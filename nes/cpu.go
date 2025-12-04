package nes

import "log"

const nmiVector uint16 = 0xFFFA
const resetVector uint16 = 0xFFFC
const irqVector uint16 = 0xFFFE
const oamdmaAddr uint16 = 0x4014

type cpu struct {
	bus *cpuBus

	irqPending bool
	nmiPending bool

	// Registers
	a  byte //Accumulator
	x  byte //Indexes
	y  byte
	pc uint16 //Program Counter
	s  byte   //Stack Pointer

	//Flags
	c bool //Carry
	z bool //Zero
	i bool //Interrupt Disable
	d bool //Decimal
	v bool //Overflow
	n bool //Negative

	cyclesToSkip uint
	oddCycle     bool
}

func createCPU(mem *cpuBus) *cpu {
	cpu := &cpu{
		bus: mem,
		i:   true,
	}

	cpu.pc = cpu.readAddress(resetVector)
	return cpu
}

func (cpu *cpu) write(address uint16, value byte) {
	if address == oamdmaAddr {
		cpu.cyclesToSkip += 513
		if cpu.oddCycle {
			cpu.cyclesToSkip++
		}
	}

	cpu.bus.write(address, value)
}

func (cpu *cpu) read(address uint16) byte {
	return cpu.bus.read(address)
}

func (cpu *cpu) readAddress(address uint16) uint16 {
	return uint16(cpu.read(address)) + uint16(cpu.read(address+1))<<8
}

func (cpu *cpu) pushStack(value byte) {
	cpu.bus.write(0x100|uint16(cpu.s), value)
	cpu.s--
}

func (cpu *cpu) popStack() byte {
	cpu.s++
	return cpu.read(0x100 | uint16(cpu.s))
}

func (cpu *cpu) getFlagByte() byte {
	var flags byte = bit5
	if cpu.n {
		flags |= bit7
	}
	if cpu.v {
		flags |= bit6
	}
	if cpu.d {
		flags |= bit3
	}
	if cpu.i {
		flags |= bit2
	}
	if cpu.z {
		flags |= bit1
	}
	if cpu.c {
		flags |= bit0
	}
	return flags
}

func (cpu *cpu) step() {
	if cpu.cyclesToSkip > 0 {
		cpu.cyclesToSkip--
		if cpu.cyclesToSkip != 0 {
			return
		}
	}

	if cpu.nmiPending || (cpu.irqPending && !cpu.i) {
		cpu.interrupt()
		return
	}

	opcode := cpu.read(cpu.pc)
	cpu.pc++

	cpu.execute(opcode)
	cpu.oddCycle = !cpu.oddCycle
}

func (cpu *cpu) execute(opcode byte) {
	// https://llx.com/Neil/a2/opcodes.html
	executed := cpu.executeImplied(opcode) ||
		cpu.executeBranch(opcode) ||
		cpu.executeGroupOne(opcode) ||
		cpu.executeGroupTwo(opcode) ||
		cpu.executeGroupThree(opcode)

	if !executed {
		log.Printf("Failed to execute opcode: %x\n", opcode)
	}
}

func (cpu *cpu) executeImplied(opcode byte) bool {
	switch opcode {
	// TAX
	case 0xAA:
		cpu.x = cpu.a
		cpu.setZN(cpu.x)
		cpu.cyclesToSkip += 2
	// TAY
	case 0xA8:
		cpu.y = cpu.a
		cpu.setZN(cpu.y)
		cpu.cyclesToSkip += 2
	// TXA
	case 0x8A:
		cpu.a = cpu.x
		cpu.setZN(cpu.a)
		cpu.cyclesToSkip += 2
	// TYA
	case 0x98:
		cpu.a = cpu.y
		cpu.setZN(cpu.a)
		cpu.cyclesToSkip += 2
	// DEX
	case 0xCA:
		cpu.x--
		cpu.setZN(cpu.x)
		cpu.cyclesToSkip += 2
	// DEY
	case 0x88:
		cpu.y--
		cpu.setZN(cpu.y)
		cpu.cyclesToSkip += 2
	// INX
	case 0xE8:
		cpu.x++
		cpu.setZN(cpu.x)
		cpu.cyclesToSkip += 2
	// INY
	case 0xC8:
		cpu.y++
		cpu.setZN(cpu.y)
		cpu.cyclesToSkip += 2
	// BRK
	case 0x00:
		cpu.pushStack(byte((cpu.pc + 1) >> 8))
		cpu.pushStack(byte(cpu.pc + 1))
		cpu.pushStack(cpu.getFlagByte() | bit4)
		cpu.i = true
		cpu.pc = cpu.readAddress(irqVector)
		cpu.cyclesToSkip += 7
	// JSR
	case 0x20:
		cpu.pushStack(byte((cpu.pc + 1) >> 8))
		cpu.pushStack(byte(cpu.pc + 1))
		cpu.pc = cpu.readAddress(cpu.pc)
		cpu.cyclesToSkip += 6
	// RTS
	case 0x60:
		cpu.pc = uint16(cpu.popStack()) + (uint16(cpu.popStack()) << 8) + 1
		cpu.cyclesToSkip += 6
	// RTI
	case 0x40:
		flags := cpu.popStack()
		cpu.n = hasBit7(flags)
		cpu.v = hasBit6(flags)
		cpu.d = hasBit3(flags)
		cpu.i = hasBit2(flags)
		cpu.z = hasBit1(flags)
		cpu.c = hasBit0(flags)
		cpu.pc = uint16(cpu.popStack()) + uint16(cpu.popStack())<<8
		cpu.cyclesToSkip += 6
	// PHA
	case 0x48:
		cpu.pushStack(cpu.a)
		cpu.cyclesToSkip += 3
	// PHP
	case 0x08:
		cpu.pushStack(cpu.getFlagByte() | 0x10)
		cpu.cyclesToSkip += 3
	// PLA
	case 0x68:
		cpu.a = cpu.popStack()
		cpu.setZN(cpu.a)
		cpu.cyclesToSkip += 4
	// PLP
	case 0x28:
		flags := cpu.popStack()
		cpu.n = hasBit7(flags)
		cpu.v = hasBit6(flags)
		cpu.d = hasBit3(flags)
		cpu.i = hasBit2(flags)
		cpu.z = hasBit1(flags)
		cpu.c = hasBit0(flags)
		cpu.cyclesToSkip += 4
	// TSX
	case 0xBA:
		cpu.x = cpu.s
		cpu.setZN(cpu.x)
		cpu.cyclesToSkip += 2
	// TXS
	case 0x9A:
		cpu.s = cpu.x
		cpu.cyclesToSkip += 2
	// CLC
	case 0x18:
		cpu.c = false
		cpu.cyclesToSkip += 2
	// CLD
	case 0xD8:
		cpu.d = false
		cpu.cyclesToSkip += 2
	// CLI
	case 0x58:
		cpu.i = false
		cpu.cyclesToSkip += 2
	// CLV
	case 0xB8:
		cpu.v = false
		cpu.cyclesToSkip += 2
	// SEC
	case 0x38:
		cpu.c = true
		cpu.cyclesToSkip += 2
	// SED
	case 0xF8:
		cpu.d = true
		cpu.cyclesToSkip += 2
	// SEI
	case 0x78:
		cpu.i = true
		cpu.cyclesToSkip += 2
	// NOP
	case 0xEA:
		//Do nothing
		cpu.cyclesToSkip += 2
	default:
		return false
	}
	return true
}

func (cpu *cpu) executeBranch(opcode byte) bool {
	if opcode&0b11111 != 0b10000 {
		return false
	}

	xx := opcode >> 6
	y := opcode >> 5 & 1

	shouldBranch := false
	isFlagSet := false
	switch xx {
	case 0b00:
		isFlagSet = cpu.n
	case 0b01:
		isFlagSet = cpu.v
	case 0b10:
		isFlagSet = cpu.c
	case 0b11:
		isFlagSet = cpu.z
	}

	if y == 0 {
		shouldBranch = !isFlagSet
	} else {
		shouldBranch = isFlagSet
	}

	if shouldBranch {
		page := cpu.pc >> 8
		offset := uint16(cpu.read(cpu.pc))
		cpu.pc += 1 + offset
		if offset >= 128 {
			cpu.pc -= 256
		}
		cpu.cyclesToSkip++
		if cpu.pc>>8 != page {
			cpu.cyclesToSkip++
		}
	} else {
		cpu.pc++
	}

	cpu.cyclesToSkip += 2

	return true
}

func (cpu *cpu) executeGroupOne(opcode byte) bool {
	cc := opcode & 0b11
	if cc != 0b01 {
		return false
	}

	aaa := opcode >> 5
	bbb := opcode >> 2 & 0b111

	var address uint16
	// Addressing modes
	switch bbb {
	// Indexed Indirect X
	case 0b000:
		arg := cpu.read(cpu.pc)
		cpu.pc++
		ind := arg + cpu.x
		address = uint16(cpu.read(uint16(ind))) + uint16(cpu.read(uint16(ind+1)))<<8
		cpu.cyclesToSkip += 6
	// Zero page
	case 0b001:
		address = uint16(cpu.read(cpu.pc))
		cpu.pc++
		cpu.cyclesToSkip += 3
	// #immediate
	case 0b010:
		address = cpu.pc
		cpu.pc++
		cpu.cyclesToSkip += 2
	// Absolute
	case 0b011:
		address = cpu.readAddress(cpu.pc)
		cpu.pc += 2
		cpu.cyclesToSkip += 4
	// Indexed Indirect Y
	case 0b100:
		arg := uint16(cpu.read(cpu.pc))
		cpu.pc++
		address = uint16(cpu.read(arg)) + uint16(cpu.read((arg+1)%256))<<8 + uint16(cpu.y)
		cpu.cyclesToSkip += 5
		if (address-uint16(cpu.y))&0xFF00 != address&0xFF00 || aaa == 0b100 {
			cpu.cyclesToSkip += 1
		}
	// Zero page indexed X
	case 0b101:
		address = uint16(cpu.read(cpu.pc) + cpu.x)
		cpu.pc++
		cpu.cyclesToSkip += 4
	// Absolute indexed Y
	case 0b110:
		arg := cpu.readAddress(cpu.pc)
		address = arg + uint16(cpu.y)
		cpu.pc += 2
		cpu.cyclesToSkip += 4
		if arg&0xFF00 != address&0xFF00 || aaa == 0b100 {
			cpu.cyclesToSkip += 1
		}
	// Absolute indexed X
	case 0b111:
		arg := cpu.readAddress(cpu.pc)
		address = arg + uint16(cpu.x)
		cpu.pc += 2
		cpu.cyclesToSkip += 4
		if arg&0xFF00 != address&0xFF00 || aaa == 0b100 {
			cpu.cyclesToSkip += 1
		}
	}

	// Instructions
	switch aaa {
	// ORA
	case 0b000:
		cpu.a |= cpu.read(address)
		cpu.setZN(cpu.a)
	// AND
	case 0b001:
		cpu.a &= cpu.read(address)
		cpu.setZN(cpu.a)
	// EOR
	case 0b010:
		cpu.a ^= cpu.read(address)
		cpu.setZN(cpu.a)
	// ADC
	case 0b011:
		operand := cpu.read(address)
		sum := uint16(cpu.a) + uint16(operand)
		if cpu.c {
			sum++
		}
		cpu.c = sum > 0xFF
		cpu.v = (sum^uint16(cpu.a))&(sum^uint16(operand))&0x80 != 0
		cpu.a = byte(sum)
		cpu.setZN(cpu.a)
	// STA
	case 0b100:
		cpu.write(address, cpu.a)
	// LDA
	case 0b101:
		cpu.a = cpu.read(address)
		cpu.setZN(cpu.a)
	// CMP
	case 0b110:
		operand := cpu.read(address)
		cpu.c = cpu.a >= operand
		cpu.setZN(cpu.a - operand)
	// SBC
	case 0b111:
		operand := cpu.read(address)
		diff := cpu.a - operand
		if !cpu.c {
			diff--
			cpu.c = operand+1 <= cpu.a
		} else {
			cpu.c = operand <= cpu.a
		}
		cpu.v = (diff^cpu.a)&(diff^^operand)&0x80 == 0x80
		cpu.a = diff
		cpu.setZN(cpu.a)
	default:
		return false
	}
	return true
}

func (cpu *cpu) executeGroupTwo(opcode byte) bool {
	cc := opcode & 0b11
	if cc != 0b10 {
		return false
	}

	aaa := opcode >> 5
	bbb := opcode >> 2 & 0b111

	var address uint16

	// Addressing modes
	switch bbb {
	// #immediate
	case 0b000:
		address = cpu.pc
		cpu.pc++
		cpu.cyclesToSkip += 2
	// Zero page
	case 0b001:
		address = uint16(cpu.read(cpu.pc))
		cpu.pc++
		if aaa == 0b100 || aaa == 0b101 {
			cpu.cyclesToSkip += 3
		} else {
			cpu.cyclesToSkip += 5
		}
	// Accumulator
	case 0b010:
		cpu.cyclesToSkip += 2
	// Absolute
	case 0b011:
		address = cpu.readAddress(cpu.pc)
		cpu.pc += 2
		if aaa == 0b100 || aaa == 0b101 {
			cpu.cyclesToSkip += 4
		} else {
			cpu.cyclesToSkip += 6
		}
	// Zero page indexed X/Y
	case 0b101:
		if aaa == 0b100 || aaa == 0b101 {
			address = uint16(cpu.read(cpu.pc) + cpu.y)
		} else {
			address = uint16(cpu.read(cpu.pc) + cpu.x)
		}
		cpu.pc++
		cpu.cyclesToSkip += 6
	// Absolute indexed X/Y
	case 0b111:
		address = cpu.readAddress(cpu.pc)
		cpu.pc += 2
		if aaa == 0b101 {
			address += uint16(cpu.y)
		} else {
			address += uint16(cpu.x)
		}
		cpu.cyclesToSkip += 7
	default:
		return false
	}

	switch aaa {
	// ASL
	case 0b000:
		var value byte
		// Accumulator
		if bbb == 0b010 {
			value = cpu.a
			cpu.a = value << 1
		} else {
			value = cpu.read(address)
			cpu.write(address, value<<1)
		}

		cpu.c = hasBit7(value)
		cpu.setZN(value << 1)
	// ROL
	case 0b001:
		var value uint16
		// Accumulator
		if bbb == 0b010 {
			value = uint16(cpu.a) << 1
			if cpu.c {
				value += 1
			}
			cpu.a = byte(value)
		} else {
			value = uint16(cpu.read(address)) << 1
			if cpu.c {
				value += 1
			}
			cpu.write(address, byte(value))
		}

		cpu.c = value&0x100 == 0x100
		cpu.setZN(byte(value))
	// LSR
	case 0b010:
		var value byte
		// Accumulator
		if bbb == 0b010 {
			value = cpu.a
			cpu.a = value >> 1
		} else {
			value = cpu.read(address)
			cpu.write(address, value>>1)
		}

		cpu.c = value&1 == 1
		cpu.setZN(value >> 1)
	// ROR
	case 0b011:
		var value uint16
		// Accumulator
		if bbb == 0b010 {
			value = uint16(cpu.a)
			if cpu.c {
				value += 0x100
			}
			cpu.c = value&1 == 1
			cpu.a = byte(value >> 1)
		} else {
			value = uint16(cpu.read(address))
			if cpu.c {
				value += 0x100
			}
			cpu.c = value&1 == 1
			cpu.write(address, byte(value>>1))
		}

		cpu.setZN(byte(value >> 1))
	// STX
	case 0b100:
		cpu.write(address, cpu.x)
	// LDX
	case 0b101:
		cpu.x = cpu.read(address)
		cpu.setZN(cpu.x)
	// DEC
	case 0b110:
		value := cpu.read(address) - 1
		cpu.setZN(value)
		cpu.write(address, value)
	// INC
	case 0b111:
		value := cpu.read(address) + 1
		cpu.setZN(value)
		cpu.write(address, value)
	default:
		return false
	}
	return true
}

func (cpu *cpu) executeGroupThree(opcode byte) bool {
	cc := opcode & 0b11
	if cc != 0b00 {
		return false
	}

	aaa := opcode >> 5
	bbb := opcode >> 2 & 0b111

	var address uint16

	// Addressing modes
	switch bbb {
	// #immediate
	case 0b000:
		address = cpu.pc
		cpu.pc++
		cpu.cyclesToSkip += 2
	// Zero page
	case 0b001:
		address = uint16(cpu.read(cpu.pc))
		cpu.pc++
		cpu.cyclesToSkip += 3
	// Absolute
	case 0b011:
		// Skip for JMP
		if aaa == 0b010 || aaa == 0b011 {
			break
		}
		address = cpu.readAddress(cpu.pc)
		cpu.pc += 2
		cpu.cyclesToSkip += 4
	// Zero page indexed X
	case 0b101:
		address = uint16(cpu.read(cpu.pc) + cpu.x)
		cpu.pc++
		cpu.cyclesToSkip += 4
	// Absolute indexed X
	case 0b111:
		arg := cpu.readAddress(cpu.pc)
		address = arg + uint16(cpu.x)
		cpu.pc += 2
		cpu.cyclesToSkip += 4
		if arg&0xFF00 != address&0xFF00 {
			cpu.cyclesToSkip += 1
		}
	}

	switch aaa {
	// BIT
	case 0b001:
		operand := cpu.read(address)
		cpu.z = operand&cpu.a == 0
		cpu.v = hasBit6(operand)
		cpu.n = hasBit7(operand)
	// JMP
	case 0b010:
		cpu.pc = cpu.readAddress(cpu.pc)
		cpu.cyclesToSkip += 3
	// JMP (indirect)
	case 0b011:
		address = cpu.readAddress(cpu.pc)
		page := address & 0xFF00
		highByteAddr := page + (address+1)&0xff
		cpu.pc = uint16(cpu.read(address)) + uint16(cpu.read(highByteAddr))<<8
		cpu.cyclesToSkip += 5
	// STY
	case 0b100:
		cpu.write(address, cpu.y)
	// LDY
	case 0b101:
		cpu.y = cpu.read(address)
		cpu.setZN(cpu.y)
	// CPY
	case 0b110:
		operand := cpu.read(address)
		cpu.c = cpu.y >= operand
		cpu.setZN(cpu.y - operand)
	// CPX
	case 0b111:
		operand := cpu.read(address)
		cpu.c = cpu.x >= operand
		cpu.setZN(cpu.x - operand)
	default:
		return false
	}
	return true
}

func (cpu *cpu) sendNMI() {
	cpu.nmiPending = true
}

func (cpu *cpu) sendIRQ() {
	cpu.irqPending = true
}

func (cpu *cpu) interrupt() {
	cpu.pushStack(byte(cpu.pc >> 8))
	cpu.pushStack(byte(cpu.pc))
	cpu.pushStack(cpu.getFlagByte())
	cpu.i = true

	if cpu.nmiPending {
		cpu.pc = cpu.readAddress(nmiVector)
		cpu.nmiPending = false
	} else {
		cpu.pc = cpu.readAddress(irqVector)
		cpu.irqPending = false
	}
	cpu.cyclesToSkip += 7
}

func (cpu *cpu) setZN(value byte) {
	cpu.z = value == 0
	cpu.n = hasBit7(value)
}
