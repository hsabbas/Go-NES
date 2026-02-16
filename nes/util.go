package nes

const (
	bit0 uint8 = 1 << iota
	bit1
	bit2
	bit3
	bit4
	bit5
	bit6
	bit7
)

const (
	kb4  = 0x1000
	kb8  = 0x2000
	kb16 = 0x4000
	kb32 = 0x8000
)

func hasBit0(b byte) bool {
	return b&bit0 == bit0
}

func hasBit1(b byte) bool {
	return b&bit1 == bit1
}

func hasBit2(b byte) bool {
	return b&bit2 == bit2
}

func hasBit3(b byte) bool {
	return b&bit3 == bit3
}

func hasBit4(b byte) bool {
	return b&bit4 == bit4
}

func hasBit5(b byte) bool {
	return b&bit5 == bit5
}

func hasBit6(b byte) bool {
	return b&bit6 == bit6
}

func hasBit7(b byte) bool {
	return b&bit7 == bit7
}
