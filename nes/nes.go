package nes

import "log"

type NES struct {
	cpu         *cpu
	ppu         *ppu
	controller1 *controller
	controller2 *controller
}

func BootNES(rom []byte) *NES {
	mapper, err := createMapper(rom)
	if err != nil {
		log.Fatal("Failed to read ROM\n", err)
	}

	controller1 := &controller{}
	controller2 := &controller{}

	ppuBus := &ppuBus{
		m: mapper,
	}
	ppu := &ppu{
		bus: ppuBus,
	}

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
	}
}

func (nes *NES) RunFrame() {
	for i := 0; i < 29781; i++ {
		nes.ppu.step()
		nes.ppu.step()
		nes.ppu.step()
		nes.cpu.step()
	}
}

func (nes *NES) ReceivePlayer1Input(button Button, pressed bool) {
	nes.controller1.updateButton(button, pressed)
}

func (nes *NES) SetFrameCallback(frameCallback func([240 * 256 * 3]byte)) {
	nes.ppu.setFrameCallback(frameCallback)
}
