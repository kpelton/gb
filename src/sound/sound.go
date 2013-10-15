package sound

import (
		"fmt"
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

func (s *Sound) Write_mmio(addr uint16 ,val uint8) {
	switch (addr) {
	case 0xff10:
		s.SND_MODE_1_SWP = val
	case 0xff11:
		s.SND_MODE_1_LEN = val
	case 0xff12:
		s.SND_MODE_1_ENVP = val
	case 0xff13:
		s.SND_MODE_1_FREQ_LOW = val
	case 0xff14:
		s.SND_MODE_1_FREQ_HI = val

	case 0xff16:
		s.SND_MODE_2_LEN = val
	case 0xff17:
		s.SND_MODE_2_ENVP = val
	case 0xff18:
		s.SND_MODE_2_FREQ_LOW = val
	case 0xff19:
		s.SND_MODE_2_FREQ_HI = val

	case 0xff1a:
		s.SND_MODE_3 = val
	case 0xff1b:
		s.SND_MODE_3_LEN = val
		//fmt.Println(val)
	case 0xff1c:
		s.SND_MODE_3_OUTPUT = val
	case 0xff1d:
		s.SND_MODE_3_FREQ_HI = val
	case 0xff1e:
		s.SND_MODE_3_FREQ_HI = val

	case 0xff20:
		s.SND_MODE_4_LEN = val
	case 0xff21:
		s.SND_MODE_4_ENVP = val
	case 0xff22:
		s.SND_MODE_4_POLY = val
	case 0xff23:
		s.SND_MODE_4_COUNTER = val

	case 0xff24:
		s.SND_CHN_CTRL = val
	case 0xff25:
		s.SND_TERM_OUTPUT = val
	case 0xff26:
		s.SND_MASTER_CTRL = val

	default:
		fmt.Println("SOUND: unhandled sound write",addr)


	}
}

func (s *Sound) Read_mmio(addr uint16) uint8 {
	var val uint8
	switch addr {

	case 0xff10:
		val = s.SND_MODE_1_SWP
	case 0xff11:
		val = s.SND_MODE_1_LEN
	case 0xff12:
		val = s.SND_MODE_1_ENVP
	case 0xff13:
		val = s.SND_MODE_1_FREQ_LOW
	case 0xff14:
		val = s.SND_MODE_1_FREQ_HI

	case 0xff16:
		val = s.SND_MODE_2_LEN
	case 0xff17:
		val = s.SND_MODE_2_ENVP
	case 0xff18:
		val = s.SND_MODE_2_FREQ_LOW
	case 0xff19:
		val = s.SND_MODE_2_FREQ_HI

	case 0xff1a:
		val = s.SND_MODE_3
	case 0xff1b:
		val = s.SND_MODE_3_LEN
	case 0xff1c:
		val = s.SND_MODE_3_OUTPUT
	case 0xff1d:
		val = s.SND_MODE_3_FREQ_HI
	case 0xff1e:
		val = s.SND_MODE_3_FREQ_HI

	case 0xff20:
		val = s.SND_MODE_4_LEN
	case 0xff21:
		val = s.SND_MODE_4_ENVP
	case 0xff22:
		val = s.SND_MODE_4_POLY
	case 0xff23:
		val = s.SND_MODE_4_COUNTER

	case 0xff24:
		val = s.SND_CHN_CTRL
	case 0xff25:
		val = s.SND_TERM_OUTPUT
	case 0xff26:
		val = s.SND_MASTER_CTRL

	default:
		fmt.Printf("SOUND: unhandled sound read %x\n",addr)
	}
	return val
}

