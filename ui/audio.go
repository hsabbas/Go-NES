package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	sampleRate = 48000
)

type audio struct {
	stream rl.AudioStream
	output chan float32

	bufferStart int
	bufferSize  int
	lastSample  float32
}

func initAudio() *audio {
	rl.InitAudioDevice()
	stream := rl.LoadAudioStream(sampleRate, 32, 1)
	a := &audio{
		stream: stream,
		output: make(chan float32, sampleRate),
	}

	return a
}

func (a *audio) startAudio() {
	rl.SetAudioStreamBufferSizeDefault(2048)
	rl.PlayAudioStream(a.stream)
	rl.SetAudioStreamCallback(a.stream, a.streamCallback)
}

func (a *audio) cleanupAudio() {
	rl.UnloadAudioStream(a.stream)
	rl.CloseAudioDevice()
	close(a.output)
}

func (a *audio) streamCallback(data []float32, frames int) {
	for i := range frames {
		select {
		case sample := <-a.output:
			data[i] = sample
		default:
			data[i] = 0
		}
	}
}
