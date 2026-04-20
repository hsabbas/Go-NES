package nes

import "log"

type NES struct {
	cpu         *cpu
	ppu         *ppu
	controller1 *controller
	controller2 *controller
}

func BootNES(rom []byte) (*NES, error) {
	mapper, err := createMapper(rom)
	if err != nil {
		return nil, err
	}

	controller1 := &controller{}
	controller2 := &controller{}

	ppuBus := &ppuBus{
		m: mapper,
	}
	ppu := createPPU(ppuBus)

	cpuBus := &cpuBus{
		m:           mapper,
		ppu:         ppu,
		controller1: controller1,
		controller2: controller2,
	}
	cpu := createCPU(cpuBus)

	ppu.setNMICallback(cpu.sendNMI)

	return &NES{
		cpu:         cpu,
		ppu:         ppu,
		controller1: controller1,
		controller2: controller2,
	}, nil
}

func (nes *NES) RunFrame() {
	frameDone := false
	for !frameDone {
		frameDone = frameDone || nes.ppu.step()
		frameDone = frameDone || nes.ppu.step()
		frameDone = frameDone || nes.ppu.step()
		nes.cpu.step()
	}
}

func (nes *NES) UpdatePlayer1Button(button Button, pressed bool) {
	nes.controller1.updateButton(button, pressed)
}

func (nes *NES) UpdatePlayer2Button(button Button, pressed bool) {
	nes.controller2.updateButton(button, pressed)
}

func (nes *NES) UpdatePlayer1Register(register byte) {
	nes.controller1.updateRegister(register)
}

func (nes *NES) UpdatePlayer2Register(register byte) {
	nes.controller2.updateRegister(register)
}

func (nes *NES) GetImage() [240][256 * 3]byte {
	return nes.ppu.frame
}
