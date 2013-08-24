package sound

import (
	//	"fmt"
	//"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	//"github.com/banthar/Go-SDL/sdl"
	"banthar/mixer"
)

type Sound struct {
	SND_MODE_1_SWP      uint8 //0xff10
	SND_MODE_1_LEN      uint8 //0xff11
	SND_MODE_1_ENVP     uint8 //0xff12
	SND_MODE_1_FREQ_LOW uint8 //0xff13
	SND_MODE_1_FREQ_HI  uint8 //0xff14

	SND_MODE_2_LEN      uint8 //0xff16
	SND_MODE_2_ENVP     uint8 //0xff17
	SND_MODE_2_FREQ_LOW uint8 //0xff18
	SND_MODE_2_FREQ_HI  uint8 //0xff19

	SND_MODE_3          uint8 //0xff1a on or off
	SND_MODE_3_LEN      uint8 //0xff1b
	SND_MODE_3_OUTPUT   uint8 //0xff1c
	SND_MODE_3_FREQ_LOW uint8 //0xff1d
	SND_MODE_3_FREQ_HI  uint8 //0xff1e

	SND_MODE_4_LEN     uint8 //0xff20
	SND_MODE_4_ENVP    uint8 //0xff21
	SND_MODE_4_POLY    uint8 //0xff22
	SND_MODE_4_COUNTER uint8 //0xff23

	SND_CHN_CTRL    uint8 //0xff24
	SND_TERM_OUTPUT uint8 //0xff25
	SND_MASTER_CTRL uint8 //0xff26
	Wram            [0x10]uint8
}

const (
	sample_rate = 44100
	channels    = 2
	samples     = 2048
)

func NewSound() *Sound {
	s := new(Sound)
	mixer.OpenAudio(sample_rate, mixer.AUDIO_S16, channels, samples)
	mixer.ResumeMusic()
	return s
}
