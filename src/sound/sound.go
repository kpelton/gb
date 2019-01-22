package sound

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"component"
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
	reg_list        component.RegList
	clocks		uint64
}

const (
	sample_rate = 44100
	channels    = 2
	samples     = 2048
)

func (g *Sound) Get_reg_list() component.RegList {
	return g.reg_list
}
func (s *Sound) Play_sound(freq uint32) {
	SamplesPerSecond := uint32(48000);
	ToneHz := freq;
	//ToneVolume := 3000;
	//RunningSampleIndex := 0;
	SquareWavePeriod := uint32(SamplesPerSecond / ToneHz);
	HalfSquareWavePeriod := uint32(SquareWavePeriod );
	//BytesPerSample := 2 * 2;
	//BytesToWrite := 800 * BytesPerSample;
	var RunningSampleIndex uint32
	var sampleout [48000]byte
	SampleCount:=1000
	for SampleIndex := 0; SampleIndex < SampleCount; SampleIndex++ {
		if (RunningSampleIndex / HalfSquareWavePeriod) % 2 == 1 {
			sampleout[SampleIndex] =byte(0xff)
		}else {
			sampleout[SampleIndex] =byte(0x0)
		}
		RunningSampleIndex+=1;
	}
	sdl.QueueAudio(1,sampleout[0:SampleCount])
}
func (s *Sound) Setup_SDL() {
	var desired sdl.AudioSpec
	desired.Freq = 48000
	desired.Format =sdl.AUDIO_U8
	desired.Channels=2
	desired.Silence = 0
	desired.Samples = 1024
	desired.Size = 48000
	var recv sdl.AudioSpec

	sdl.OpenAudio(&desired,&recv)

	sdl.PauseAudio(false)

}
func (s *Sound) Reset() {
}

func (s *Sound) Update(clocks uint16) {
	s.clocks += uint64(clocks)
}
func (s *Sound) Update_channel1() {
	//First 3 bits of Freq_Hi is part of the 11bit freq
	hi_freq := uint16(s.SND_MODE_1_FREQ_HI &0x7)
	hi_freq = hi_freq <<8
	lo_freq := uint16(s.SND_MODE_1_FREQ_LOW)
	snd_len :=s.SND_MODE_1_LEN &0x1f
	freq:=hi_freq+lo_freq
	real_freq := 131072/uint32((2048-freq))
	fmt.Println(131072/uint32((2048-freq)),snd_len)
	if real_freq > 100 {
		s.Play_sound(real_freq)
	}
}
func (s *Sound) Update_channel2() {
	//First 3 bits of Freq_Hi is part of the 11bit freq
	hi_freq := uint16(s.SND_MODE_2_FREQ_HI &0x7)
	hi_freq = hi_freq <<8
	lo_freq := uint16(s.SND_MODE_2_FREQ_LOW)
	snd_len :=s.SND_MODE_1_LEN &0x1f
	freq:=hi_freq+lo_freq
	real_freq := 131072/uint32((2048-freq))
	fmt.Println("2nd",131072/uint32((2048-freq)),snd_len)
	if real_freq > 100 {
		s.Play_sound(real_freq)
	}
}

func NewSound() *Sound {
	s := new(Sound)
	s.Setup_SDL()
	s.reg_list = component.RegList{
		{Name: "SND_MODE_1_SWP ", Addr: 0xff10},
		{Name: "SND_MODE_1_LEN", Addr: 0xff11},
		{Name: "SND_MODE_1_ENVP ", Addr: 0xff12},
		{Name: "SND_MODE_1_FREQ_LOW", Addr: 0xff13},
		{Name: "SND_MODE_1_FREQ_HI", Addr: 0xff14},
		{Name: "SND_MODE_2_LEN ", Addr: 0xff16},
		{Name: "SND_MODE_2_ENVP", Addr: 0xff17},
		{Name: "SND_MODE_2_FREQ_LOW", Addr: 0xff18},
		{Name: "SND_MODE_2_FREQ_HI", Addr: 0xff19},
		{Name: "SND_MODE_3", Addr: 0xff1a},
		{Name: "SND_MODE_3_LEN", Addr: 0xff1b},
		{Name: "SND_MODE_3_OUTPUT", Addr: 0xff1c},
		{Name: "SND_MODE_3_FREQ_LOW", Addr: 0xff1d},
		{Name: "SND_MODE_3_FREQ_HI", Addr: 0xff1e},
		{Name: "SND_MODE_4_LEN", Addr: 0xff20},
		{Name: "SND_MODE_4_ENVP", Addr: 0xff21},
		{Name: "SND_MODE_4_POLY", Addr: 0xff22},
		{Name: "SND_MODE_4_COUNTER", Addr: 0xff23},
		{Name: "SND_CHN_CTRL", Addr: 0xff24},
		{Name: "SND_TERM_OUTPUT", Addr: 0xff25},
		{Name: "SND_MASTER_CTRL", Addr: 0xff26},
	}

	return s
}

func (s *Sound) Write_mmio(addr uint16, val uint8) {
	switch addr {
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
		s.Update_channel1()
		//s.Update_channel2()
	case 0xff26:
		s.SND_MASTER_CTRL = val

	default:
		fmt.Println("SOUND: unhandled sound write", addr)

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
		fmt.Printf("SOUND: unhandled sound read %x\n", addr)
	}
	return val
}
