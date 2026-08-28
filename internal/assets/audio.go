package assets

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

var (
	AudioContext *audio.Context
	HitSound     []byte
	ShoveSound   []byte
)

func InitAudio() {
	if AudioContext == nil {
		AudioContext = audio.NewContext(44100)
	}
	
	// Generate Hit Sound (Short white noise burst)
	hitData := make([]byte, 44100/10*4) // 0.1 seconds, 4 bytes per sample (stereo 16-bit)
	for i := 0; i < len(hitData); i += 4 {
		// Fade out
		vol := 1.0 - float64(i)/float64(len(hitData))
		val := int16((rand.Float64()*2 - 1) * 32767 * 0.3 * vol)
		
		hitData[i] = byte(val)
		hitData[i+1] = byte(val >> 8)
		hitData[i+2] = byte(val)
		hitData[i+3] = byte(val >> 8)
	}
	HitSound = hitData

	// Generate Shove Sound (Low thump - sine wave)
	shoveData := make([]byte, 44100/5*4) // 0.2 seconds
	for i := 0; i < len(shoveData)/4; i++ {
		vol := 1.0 - float64(i)/float64(len(shoveData)/4)
		// Frequency sweep down from 150hz to 50hz
		freq := 150.0 - 100.0*(float64(i)/float64(len(shoveData)/4))
		val := int16(math.Sin(float64(i)*freq*2*math.Pi/44100) * 32767 * 0.5 * vol)
		
		idx := i * 4
		shoveData[idx] = byte(val)
		shoveData[idx+1] = byte(val >> 8)
		shoveData[idx+2] = byte(val)
		shoveData[idx+3] = byte(val >> 8)
	}
	ShoveSound = shoveData
}

func PlaySound(data []byte) {
	if AudioContext == nil {
		return
	}
	p := AudioContext.NewPlayerFromBytes(data)
	p.Play()
}
