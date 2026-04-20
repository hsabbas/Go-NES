package nes

const frameCounterRate float64 = cpuFreq / 240.0

var lcTable = []byte{
	10, 254, 20, 2, 40, 4, 80, 6,
	160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22,
	192, 24, 72, 26, 16, 28, 32, 30,
}
var pulseDutyTable = [4][8]byte{
	{0, 1, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 0, 0, 0},
	{1, 0, 0, 1, 1, 1, 1, 1},
}
var triangleSequenceTable = []byte{
	15, 14, 13, 12, 11, 10, 9, 8,
	7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15,
}
var noisePeriodLU = []uint16{
	4, 8, 16, 32, 64, 96, 128, 160, 202,
	254, 380, 508, 762, 1016, 2034, 4068,
}
var dmcRateLU = []uint16{
	428, 380, 340, 320, 286, 254, 226, 214,
	190, 160, 142, 128, 106, 84, 72, 54,
}
var pulseOutLU = [31]float32{}
var tndLU = [203]float32{}

func init() {
	for i := range len(pulseOutLU) {
		pulseOutLU[i] = 95.52 / (8128/float32(i) + 100)
	}

	for i := range len(tndLU) {
		tndLU[i] = 163.67 / (24329/float32(i) + 100)
	}
}

type apu struct {
	cpu *cpu
	p1  *pulse
	p2  *pulse
	t   *triangle
	n   *noise
	dmc *dmc

	frameIRQInhibit  bool
	frame5Step       bool
	frameCounterStep byte

	cycle           uint64
	cyclesPerSample float64

	output chan<- float32
}

func createAPU() *apu {
	pulse1 := createPulse(true)
	pulse2 := createPulse(false)
	t := &triangle{}
	n := &noise{
		shiftRegister: 1,
		envelope:      &envelope{},
	}
	d := &dmc{}

	return &apu{
		p1:  pulse1,
		p2:  pulse2,
		t:   t,
		n:   n,
		dmc: d,
	}
}

func (a *apu) linkCpu(cpu *cpu) {
	a.cpu = cpu
	a.dmc.cpu = cpu
}

func (apu *apu) writeRegister(address uint16, value byte) {
	switch address {
	case 0x4000:
		writePulseRegisters(apu.p1, value)
	case 0x4001:
		writeSweep(apu.p1.sweep, value)
	case 0x4002:
		writePulseTimerLow(apu.p1, value)
	case 0x4003:
		writePulseLengthCounter(apu.p1, value)
	case 0x4004:
		writePulseRegisters(apu.p2, value)
	case 0x4005:
		writeSweep(apu.p2.sweep, value)
	case 0x4006:
		writePulseTimerLow(apu.p2, value)
	case 0x4007:
		writePulseLengthCounter(apu.p2, value)
	case 0x4008:
		apu.writeTriangleLinearControl(value)
	case 0x400a:
		apu.t.period = (apu.t.period & 0xFF00) | uint16(value)
	case 0x400b:
		apu.writeTriangleLengthCounter(value)
	case 0x400c:
		apu.writeNoiseRegisters(value)
	case 0x400e:
		apu.writeNoisePeriod(value)
	case 0x400f:
		apu.writeNoiseLengthCounter(value)
	case 0x4010:
		apu.writeDMCFlags(value)
	case 0x4011:
		apu.dmc.value = value & 0x7F
	case 0x4012:
		apu.dmc.startAddress = 0xC000 + uint16(value)*64
	case 0x4013:
		apu.dmc.length = uint16(value)*16 + 1
	case 0x4015:
		apu.writeStatus(value)
	case 0x4017:
		apu.writeFrameCounter(value)
	}
}

func writePulseRegisters(p *pulse, value byte) {
	p.duty = (value >> 6) & 3
	p.lcHalt = hasBit5(value)
	p.envelope.loop = p.lcHalt
	p.envelope.constant = hasBit4(value)
	p.envelope.volume = value & 0xF
}

func writeSweep(s *sweep, value byte) {
	s.enabled = hasBit7(value)
	s.period = (value >> 4) & 7
	s.negate = hasBit3(value)
	s.shift = value & 7
	s.reload = true
}

func writePulseTimerLow(p *pulse, value byte) {
	p.period = (p.period & 0xFF00) | uint16(value)
	p.sweep.calcTarget()
}

func writePulseLengthCounter(p *pulse, value byte) {
	p.period = (p.period & 0xFF) | uint16(value&7)<<8
	p.sweep.calcTarget()
	if p.enabled {
		p.lengthCounter = lcTable[(value>>3)&0x1F]
	}
	p.envelope.start = true
	p.sequencerStep = 0
}

func (a *apu) writeTriangleLinearControl(value byte) {
	a.t.controlFlag = hasBit7(value)
	a.t.linearReload = value & 0x7F
}

func (a *apu) writeTriangleLengthCounter(value byte) {
	if a.t.enabled {
		a.t.lengthCounter = lcTable[(value>>3)&0x1F]
	}
	a.t.period = a.t.period&0xFF | uint16(value&3)<<8
	a.t.reloadFlag = true
}

func (a *apu) writeNoiseRegisters(value byte) {
	a.n.lcHalt = hasBit5(value)
	a.n.envelope.loop = a.n.lcHalt
	a.n.envelope.constant = hasBit4(value)
	a.n.envelope.volume = value & 0xF
}

func (a *apu) writeNoisePeriod(value byte) {
	a.n.mode = hasBit7(value)
	a.n.period = noisePeriodLU[value&0xF]
}

func (a *apu) writeNoiseLengthCounter(value byte) {
	a.n.lengthCounter = lcTable[(value>>3)&0x1F]
	a.n.envelope.start = true
}

func (a *apu) writeDMCFlags(value byte) {
	a.dmc.irqEnable = hasBit7(value)
	a.dmc.loop = hasBit6(value)
	a.dmc.period = dmcRateLU[value&0xF]
}

func (apu *apu) readStatus() byte {
	var status byte
	if apu.p1.lengthCounter > 0 {
		status |= bit0
	}
	if apu.p2.lengthCounter > 0 {
		status |= bit1
	}
	if apu.t.lengthCounter > 0 && apu.t.linearCounter > 0 {
		status |= bit2
	}
	if apu.n.lengthCounter > 0 {
		status |= bit3
	}
	if apu.dmc.bytesLeft > 0 {
		status |= bit4
	}
	if apu.frameIRQInhibit {
		status |= bit6
	}
	if apu.dmc.irqEnable {
		status |= bit7
	}
	apu.frameIRQInhibit = false
	return status
}

func (apu *apu) writeStatus(value byte) {
	apu.p1.enabled = hasBit0(value)
	apu.p2.enabled = hasBit1(value)
	apu.t.enabled = hasBit2(value)
	apu.n.enabled = hasBit3(value)
	apu.dmc.enabled = hasBit4(value)
	apu.dmc.irqEnable = false

	if !apu.p1.enabled {
		apu.p1.lengthCounter = 0
	}
	if !apu.p2.enabled {
		apu.p2.lengthCounter = 0
	}
	if !apu.t.enabled {
		apu.t.lengthCounter = 0
	}
	if !apu.n.enabled {
		apu.n.lengthCounter = 0
	}
	if !apu.dmc.enabled {
		apu.dmc.bytesLeft = 0
	} else {
		if apu.dmc.bytesLeft == 0 {
			apu.dmc.currAddress = apu.dmc.startAddress
			apu.dmc.bytesLeft = apu.dmc.length
		}
	}
}

func (apu *apu) writeFrameCounter(value byte) {
	apu.frameIRQInhibit = hasBit6(value)
	apu.frame5Step = hasBit7(value)
}

func (apu *apu) stepFrameCounter() {
	if apu.frameCounterStep == 3 && apu.frame5Step {
		apu.frameCounterStep++
		return
	}

	apu.clockEnvelopes()
	apu.t.clockLinC()
	if apu.frameCounterStep == 1 || apu.frameCounterStep == 3 {
		apu.clockLengthCounters()
		apu.clockSweeps()
		if !apu.frameIRQInhibit {
			apu.cpu.sendIRQ()
		}
	}

	if apu.frameCounterStep == 4 {
		apu.clockLengthCounters()
		apu.clockSweeps()
	}

	apu.frameCounterStep++
	if apu.frameCounterStep >= 5 || (apu.frameCounterStep >= 4 && !apu.frame5Step) {
		apu.frameCounterStep = 0
	}
}

func (apu *apu) clockLengthCounters() {
	apu.p1.clockLC()
	apu.p2.clockLC()
	apu.t.clockLenC()
	apu.n.clockLC()
}

func (apu *apu) clockEnvelopes() {
	apu.p1.envelope.clock()
	apu.p2.envelope.clock()
	apu.n.envelope.clock()
}

func (apu *apu) clockSweeps() {
	apu.p1.sweep.clock()
	apu.p2.sweep.clock()
}

func (apu *apu) createSample() float32 {
	var sample float32
	p1 := apu.p1.output()
	p2 := apu.p2.output()
	t := apu.t.output()
	n := apu.n.output()
	d := apu.dmc.output()

	sample = pulseOutLU[p1+p2]
	sample += tndLU[3*t+2*n+d]
	return sample
}

func (apu *apu) step() {
	next := apu.cycle + 1
	if int(float64(apu.cycle)/apu.cyclesPerSample) != int(float64(next)/apu.cyclesPerSample) {
		apu.output <- apu.createSample()
	}
	if int(float64(apu.cycle)/frameCounterRate) != int(float64(next)/frameCounterRate) {
		apu.stepFrameCounter()
	}

	apu.cycle++

	if apu.cycle%2 == 0 {
		apu.p1.clock()
		apu.p2.clock()
	}
	apu.n.clock()
	apu.dmc.clock()
	apu.t.clock()
}

type pulse struct {
	isPulse1      bool
	enabled       bool
	duty          byte
	lcHalt        bool
	lengthCounter byte
	period        uint16
	timer         uint16
	sequencerStep byte
	sequencerVal  byte

	envelope *envelope
	sweep    *sweep
}

func createPulse(isChannel1 bool) *pulse {
	p := &pulse{
		isPulse1: isChannel1,
	}
	e := &envelope{}
	s := &sweep{
		pulse: p,
	}
	p.envelope = e
	p.sweep = s
	return p
}

func (p *pulse) clock() {
	if p.timer == 0 {
		p.timer = p.period
		p.sequencerStep++
		if p.sequencerStep >= 8 {
			p.sequencerStep = 0
		}
		p.sequencerVal = pulseDutyTable[p.duty][p.sequencerStep]
	} else {
		p.timer--
	}
}

func (p *pulse) clockLC() {
	if !p.enabled {
		p.lengthCounter = 0
	}

	if !p.lcHalt && p.lengthCounter > 0 {
		p.lengthCounter--
	}
}

func (p *pulse) output() byte {
	if !p.sweep.muting && p.sequencerVal != 0 && p.lengthCounter > 0 {
		if p.envelope.constant {
			return p.envelope.volume
		} else {
			return p.envelope.decayLevel
		}
	}

	return 0
}

type envelope struct {
	loop       bool
	counter    byte
	start      bool
	constant   bool
	volume     byte
	decayLevel byte
}

func (e *envelope) clock() {
	if e.start {
		e.start = false
		e.decayLevel = 15
		e.counter = e.volume
		return
	}

	if e.counter > 0 {
		e.counter--
		return
	}

	e.counter = e.volume
	if e.decayLevel == 0 {
		if e.loop {
			e.decayLevel = 15
		}
	} else {
		e.decayLevel--
	}
}

type sweep struct {
	pulse   *pulse
	enabled bool
	period  byte
	counter byte
	negate  bool
	shift   byte
	reload  bool
	target  uint16
	muting  bool
}

func (s *sweep) calcTarget() {
	muted := s.pulse.period < 8
	change := int(s.pulse.period >> uint16(s.shift))
	if s.negate {
		change = -change
		if s.pulse.isPulse1 {
			change--
		}
	}

	target := int(s.pulse.period) + change
	if target < 0 {
		target = 0
	}

	muted = muted || target > 0x7FF
	s.muting = muted
	s.target = uint16(target)
}

func (s *sweep) clock() {
	if s.counter == 0 && s.enabled && s.shift != 0 && !s.muting {
		s.pulse.period = s.target
		s.calcTarget()
	}

	if s.counter == 0 || s.reload {
		s.counter = s.period
		s.reload = false
	} else {
		s.counter--
	}
}

type triangle struct {
	enabled       bool
	controlFlag   bool
	lengthCounter byte
	linearCounter byte
	reloadFlag    bool
	linearReload  byte
	period        uint16
	timer         uint16
	sequencerStep byte
}

func (t *triangle) clock() {

	if t.timer == 0 {
		t.timer = t.period
		if t.linearCounter > 0 && t.lengthCounter > 0 {
			t.sequencerStep++
			if t.sequencerStep >= 32 {
				t.sequencerStep = 0
			}
		}
	} else {
		t.timer--
	}
}

func (t *triangle) clockLenC() {
	if !t.enabled {
		t.lengthCounter = 0
	}

	if !t.controlFlag && t.lengthCounter > 0 {
		t.lengthCounter--
	}

}

func (t *triangle) clockLinC() {
	if t.reloadFlag {
		t.linearCounter = t.linearReload
	} else {
		if t.linearCounter > 0 {
			t.linearCounter--
		}
	}

	if !t.controlFlag {
		t.reloadFlag = false
	}
}

func (t *triangle) output() byte {
	if t.enabled && t.lengthCounter > 0 && t.linearCounter > 0 {
		return triangleSequenceTable[t.sequencerStep]
	}
	return 0
}

type noise struct {
	shiftRegister uint16
	enabled       bool
	lcHalt        bool
	mode          bool
	period        uint16
	timer         uint16
	lengthCounter byte

	envelope *envelope
}

func (n *noise) clock() {
	if n.timer == 0 {
		var val uint16
		if n.mode {
			val = (n.shiftRegister & 1) ^ (n.shiftRegister >> 6 & 1)
		} else {
			val = (n.shiftRegister & 1) ^ (n.shiftRegister >> 1 & 1)
		}
		n.shiftRegister >>= 1
		n.shiftRegister |= val << 14
		n.timer = n.period
	} else {
		n.timer--
	}
}

func (n *noise) output() byte {
	if n.lengthCounter == 0 || hasBit0(byte(n.shiftRegister)) {
		return 0
	}

	if n.envelope.constant {
		return n.envelope.volume
	} else {
		return n.envelope.decayLevel
	}
}

func (n *noise) clockLC() {
	if !n.enabled {
		n.lengthCounter = 0
	}

	if !n.lcHalt && n.lengthCounter > 0 {
		n.lengthCounter--
	}
}

type dmc struct {
	cpu           *cpu
	enabled       bool
	irqEnable     bool
	loop          bool
	period        uint16
	timer         uint16
	startAddress  uint16
	currAddress   uint16
	length        uint16
	bytesLeft     uint16
	value         byte
	buffer        byte
	shiftRegister byte
	bitsLeft      byte
	silence       bool
}

func (d *dmc) clock() {
	if !d.enabled {
		return
	}

	d.readAddress()
	if d.timer == 0 {
		d.timer = d.period
		d.clockShiftRegister()
	} else {
		d.timer--
	}
}

func (d *dmc) clockShiftRegister() {
	if d.bitsLeft != 0 {
		if hasBit0(d.shiftRegister) {
			if d.value <= 125 {
				d.value += 2
			}
		} else {
			if d.value >= 2 {
				d.value -= 2
			}
		}
		d.shiftRegister >>= 1
		d.bitsLeft--
	}
}

func (d *dmc) readAddress() {
	if d.bytesLeft > 0 && d.bitsLeft == 0 {
		d.shiftRegister = d.cpu.read(d.currAddress)
		d.bitsLeft = 8
		if d.currAddress == 0xFFFF {
			d.currAddress = 0x8000
		} else {
			d.currAddress++
		}
		d.bytesLeft--
		if d.bytesLeft == 0 && d.loop {
			d.bytesLeft = d.length
			d.currAddress = d.startAddress
		}
	}
}

func (d *dmc) output() byte {
	return d.value
}
