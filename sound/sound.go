package sound

import (
	"fmt"
	"gb/component"

	"github.com/veandco/go-sdl2/sdl"
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
	clocks          uint64
	chan1_curr_freq uint32
	chan2_curr_freq uint32
	chan3_curr_freq uint32
	chan2_real_freq uint32
	chan3_real_freq uint32
	doNothing       bool
	dac_power       bool

	chan1_duty       uint8
	chan1_hi_freq    uint16
	chan1_lo_freq    uint16
	chan1_freq       uint16
	chan1_timer      uint16
	chan1_len_mode   uint16
	chan1_len        uint8
	chan1_len_enable uint8
	chan1_enabled    bool

	chan1_swp_shadow      uint16
	chan1_swp_period      uint8
	chan1_swp_period_load uint8
	chan1_swp_negate      uint8
	chan1_swp_shift       uint8
	chan1_swp_enable      uint8

	chan1_vol             uint8
	chan1_vol_initial     uint8
	chan1_duty_pointer    uint8
	chan1_vol_period      uint8
	chan1_vol_period_load uint8
	chan1_vol_op          uint8

	chan2_duty         uint8
	chan2_hi_freq      uint16
	chan2_lo_freq      uint16
	chan2_freq         uint16
	chan2_timer        uint16
	chan2_len_mode     uint16
	chan2_len          uint8
	chan2_len_enable   uint8
	chan2_enabled      bool
	chan2_duty_pointer uint8

	chan2_vol             uint8
	chan2_vol_initial     uint8
	chan2_vol_period      uint8
	chan2_vol_period_load uint8
	chan2_vol_op          uint8

	//Chan3
	chan3_hi_freq    uint16
	chan3_lo_freq    uint16
	chan3_freq       uint16
	chan3_timer      uint16
	chan3_len_mode   uint16
	chan3_len        uint8
	chan3_len_enable uint8
	chan3_enabled    bool
	chan3_pos        uint8

	chan3_vol         uint8
	chan3_vol_current uint8

	//Chan4

	chan4_timer      uint16
	chan4_len_mode   uint16
	chan4_len        uint8
	chan4_len_enable uint8
	chan4_enabled    bool
	chan4_divisor    uint8
	chan4_clk_shift  uint8
	chan4_width_mode uint8

	chan4_vol             uint8
	chan4_vol_current     uint8
	chan4_vol_initial     uint8
	chan4_vol_period      uint8
	chan4_vol_period_load uint8
	chan4_vol_op          uint8
	chan4_lfsr            uint16

	sample_p     uint32
	sample_timer uint32
	csample      [sample_size]byte

	square_duty        [4][8]bool
	channel_enables    [4][2]uint8
	chan4_divisor_code [8]uint8

	frame_seq_counter uint32
	frame_seq_step    uint8
}

const (
	sample_rate         = 48000
	channels            = 2
	samples             = 2048
	sample_size         = 1024
	frame_seq_clocks    = 8192
	chan1_global_enable = 1
	chan2_global_enable = 1
	chan3_global_enable = 1
	chan4_global_enable = 1
)

func (g *Sound) Get_reg_list() component.RegList {
	return g.reg_list
}

func (s *Sound) Setup_SDL() {
	var desired sdl.AudioSpec
	desired.Freq = sample_rate / 2
	desired.Format = sdl.AUDIO_U16
	desired.Channels = 2
	desired.Silence = 0
	desired.Samples = sample_size
	var recv sdl.AudioSpec

	sdl.OpenAudio(&desired, &recv)

	sdl.PauseAudio(false)

}
func (s *Sound) Reset() {
}
func (s *Sound) Update_channel1() {
	s.chan1_timer -= 1
	if s.chan1_timer == 0 {

		s.chan1_duty_pointer += 1
		s.chan1_duty_pointer &= 7
		//fmt.Println(s.clocks,"Current Timer",s.chan1_timer)
		//fmt.Println(s.clocks,"Duty",s.chan1_duty_pointer,s.square_duty[s.chan1_duty][s.chan1_duty_pointer],s.sample_p)
		s.chan1_timer = (2048 - s.chan1_freq) * 4

	}
	if !s.chan1_enabled {
		s.chan1_vol = 0
	}
}

func (s *Sound) Update_channel2() {
	s.chan2_timer -= 1
	if s.chan2_timer == 0 {

		s.chan2_duty_pointer += 1
		s.chan2_duty_pointer &= 7
		//fmt.Println(s.clocks,"Current Timer",s.chan1_timer)
		//fmt.Println(s.clocks,"Duty",s.chan1_duty_pointer,s.square_duty[s.chan1_duty][s.chan1_duty_pointer],s.sample_p)
		s.chan2_timer = (2048 - s.chan2_freq) * 4

	}
	if !s.chan2_enabled {
		s.chan2_vol = 0
	}
}

func (s *Sound) Update_channel3() {
	s.chan3_timer -= 1
	if s.chan3_timer == 0 {
		s.chan3_timer = (2048 - s.chan3_freq) * 2
		//fmt.Println("SND_MODE_3_TICK")

		pos := s.chan3_pos / 2

		data := s.Wram[pos]
		//if pos is even then grab 2nd nibble else first
		if s.chan3_pos%2 == 0 {
			data = data >> 4
			//fmt.Println("WRAM UPPER nibble")
		}
		data &= 0xf
		//fmt.Printf("WRAM update %x %x %x %x\n",s.chan3_pos,pos,data,s.Wram[pos])

		//if code != 0 do the shift otherwise output is 0
		/*
			Code   Shift   Volume
			-----------------------
			0      4         0% (silent)
			1      0       100%
			2      1        50%
			3      2        25%
		*/
		if s.dac_power && s.chan3_enabled && s.chan3_vol > 0 {
			switch s.chan3_vol {

			case 2:
				data >>= 1 //50%
			case 3:
				data >>= 2 //25%
			}
		} else {
			data = 0
		}
		s.chan3_vol_current = data
		s.chan3_pos += 1
		if s.chan3_pos == 32 {
			s.chan3_pos = 0

		}
	}
	if !s.chan3_enabled {
		s.chan3_vol_current = 0
	}
}

func (s *Sound) Update_channel4() {
	s.chan4_timer -= 1
	if s.chan4_enabled && s.chan4_timer == 0 {
		s.chan4_timer = uint16(s.chan4_divisor_code[s.chan4_divisor]) << uint16(s.chan4_clk_shift)

		val := (s.chan4_lfsr & 1) ^ ((s.chan4_lfsr & 2) >> 1)
		s.chan4_lfsr >>= 1
		s.chan4_lfsr |= val << 14
		//fmt.Println("Update Channel 4",s.chan4_timer,s.chan4_lfsr)

		if s.chan4_width_mode == 1 {
			s.chan4_lfsr &^= 0x40
			s.chan4_lfsr |= val << 6

		}

		if s.chan4_lfsr&1 == 1 {
			s.chan4_vol_current = s.chan4_vol
		} else {
			s.chan4_vol_current = 0
		}

	}
	if !s.chan4_enabled {
		s.chan4_vol_current = 0
	}
}

func (s *Sound) Sampler() {
	s.sample_timer -= 1
	if s.sample_timer == 0 {

		s.sample_timer = (4194304 / sample_rate)
		s.csample[s.sample_p] = 0
		s.csample[s.sample_p+1] = 0

		chan1_vol := s.square_duty[s.chan1_duty][s.chan1_duty_pointer]
		chan2_vol := s.square_duty[s.chan2_duty][s.chan2_duty_pointer]

		if s.channel_enables[0][0] == 1 && s.chan1_vol_initial > 0 {
			if chan1_vol == false {
				s.csample[s.sample_p] += 0
			} else {

				s.csample[s.sample_p] += (uint8(s.chan1_vol))
			}
		}

		if s.channel_enables[1][0] == 1 && s.chan2_vol_initial > 0 {
			if chan2_vol == false {
				s.csample[s.sample_p] += 0
			} else {
				s.csample[s.sample_p] += (uint8(s.chan2_vol))
			}
		}

		if s.channel_enables[2][0] == 1 && s.dac_power {
			s.csample[s.sample_p] += (uint8(s.chan3_vol_current))

		}

		if s.channel_enables[3][0] == 1 && s.chan4_vol_initial > 0 {
			s.csample[s.sample_p] += (uint8(s.chan4_vol_current))

		}

		if s.channel_enables[0][1] == 1 && s.chan1_vol_initial > 0 {
			if chan1_vol == true {
				s.csample[s.sample_p+1] += 0
			} else {
				s.csample[s.sample_p+1] += (uint8(s.chan1_vol))
			}
		}

		if s.channel_enables[1][1] == 1 && s.chan2_vol_initial > 0 {
			if chan2_vol == true {
				s.csample[s.sample_p+1] += 0
			} else {
				s.csample[s.sample_p+1] += uint8((s.chan2_vol))
			}
		}

		if s.channel_enables[2][1] == 1 && s.dac_power {
			//fmt.Println("SND_MODE_3_SAMPLE",s.chan3_vol_current)
			if s.channel_enables[3][1] == 1 && s.chan4_vol_initial > 0 {
				s.csample[s.sample_p+1] += (uint8(s.chan4_vol_current))
			}
		}

		if s.sample_p+2 >= sample_size-1 {
			p := s.csample
			s.sample_p = 0
			sdl.QueueAudio(1, p[0:sample_size])

		} else {
			s.sample_p += 2
		}

		for (sdl.GetQueuedAudioSize(1)) > sample_size {
		}

	}

}
func (s *Sound) channel1_len_clock() {
	if s.chan1_len_enable == 1 {
		s.chan1_len -= 1
		if s.chan1_len == 0 {
			//fmt.Println(" SWP Disabled due to timer")
			s.chan1_enabled = false
			s.chan1_len_enable = 0
		}
	}
}

func (s *Sound) channel2_len_clock() {
	if s.chan2_len_enable == 1 {
		s.chan2_len -= 1
		if s.chan2_len == 0 {
			//fmt.Println("Disabled 2 due to timer")
			s.chan2_enabled = false
			s.chan2_len_enable = 0
		}
	}
}

func (s *Sound) channel3_len_clock() {
	if s.chan3_len_enable == 1 {
		s.chan3_len -= 1
		if s.chan3_len == 0 {
			//fmt.Println("Disabled 3 due to timer")
			s.chan3_enabled = false
			s.chan3_len_enable = 0
		}
	}
}
func (s *Sound) channel4_len_clock() {
	if s.chan4_len_enable == 1 {
		s.chan4_len -= 1
		if s.chan4_len == 0 {
			//fmt.Println("LEN4 Disabled 2 due to timer")
			s.chan4_enabled = false
			s.chan4_len_enable = 0
		}
	}
}

func (s *Sound) channel1_vol_clock() {
	if s.chan1_vol_period != 0 {
		s.chan1_vol_period -= 1
		if s.chan1_vol_period == 0 {
			if s.chan1_vol_op == 1 && s.chan1_vol < 15 {
				s.chan1_vol += 1
			}
			if s.chan1_vol_op == 0 && s.chan1_vol > 0 {
				s.chan1_vol -= 1
			}
			s.chan1_vol_period = s.chan1_vol_period_load
			//fmt.Println("KYLEtest chan1",s.chan1_vol,s.chan1_vol_period)

		}
	}
}

func (s *Sound) channel2_vol_clock() {
	if s.chan2_vol_period != 0 {
		s.chan2_vol_period -= 1
		if s.chan2_vol_period == 0 {
			if s.chan2_vol_op == 1 && s.chan2_vol < 15 {
				s.chan2_vol += 1
			}
			if s.chan2_vol_op == 0 && s.chan2_vol > 0 {
				s.chan2_vol -= 1
			}
			s.chan2_vol_period = s.chan2_vol_period_load
			//fmt.Println("KYLEtest chan2",s.chan2_vol,s.chan2_vol_period)

		}
	}
}

func (s *Sound) channel4_vol_clock() {
	if s.chan4_vol_period != 0 {
		s.chan4_vol_period -= 1
		if s.chan4_vol_period == 0 {
			if s.chan4_vol_op == 1 && s.chan4_vol < 15 {
				s.chan4_vol += 1
			}
			if s.chan4_vol_op == 0 && s.chan4_vol > 0 {
				s.chan4_vol -= 1
			}
			s.chan4_vol_period = s.chan4_vol_period_load
			//fmt.Println("KYLEtest chan4", s.chan4_vol, s.chan4_vol_period)

		}
	}
}
func (s *Sound) channel1_swp_calc() uint16 {

	curr_val := s.chan1_swp_shadow
	val := s.chan1_swp_shadow >> s.chan1_swp_shift

	if s.chan1_swp_negate == 1 {
		curr_val -= val
		//fmt.Println("SWP negate calc", curr_val, s.chan1_swp_shift, val)

	} else {
		curr_val += val
		//fmt.Println("SWP add calc", curr_val, s.chan1_swp_shift, val)

	}
	if curr_val > 2047 {
		s.chan1_enabled = false
		s.chan1_swp_enable = 0
		//fmt.Println("SWP Disabled channel 1 due to freq overflow")
	}
	return curr_val

}
func (s *Sound) channel1_swp_clock() {

	s.chan1_swp_period -= 1
	if s.chan1_swp_enable == 1 && s.chan1_swp_period_load > 0 {
		if s.chan1_swp_period == 0 {
			s.chan1_swp_period = s.chan1_swp_period_load
			new_calc := s.channel1_swp_calc()
			if s.chan1_swp_shift > 0 && new_calc < 2047 {
				if s.chan1_enabled {
					//fmt.Println("SWP New freq", new_calc, s.chan1_freq)
					s.chan1_freq = new_calc
					s.chan1_swp_shadow = new_calc
				}
			}
			s.channel1_swp_calc()

		}
	}
}

func (s *Sound) Freq_sampler() {
	s.frame_seq_counter -= 1

	if s.frame_seq_counter == 0 {
		s.frame_seq_counter = frame_seq_clocks
		switch s.frame_seq_step {

		case 0:
			s.channel1_len_clock()
			s.channel2_len_clock()
			s.channel3_len_clock()
			s.channel4_len_clock()

		case 2:
			s.channel1_swp_clock()
			s.channel1_len_clock()
			s.channel2_len_clock()
			s.channel3_len_clock()
			s.channel4_len_clock()

		case 4:
			s.channel1_len_clock()
			s.channel2_len_clock()
			s.channel3_len_clock()
			s.channel4_len_clock()
		case 6:
			s.channel1_swp_clock()

			s.channel1_len_clock()
			s.channel2_len_clock()
			s.channel3_len_clock()
			s.channel4_len_clock()
		case 7:
			s.channel1_vol_clock()
			s.channel2_vol_clock()
			s.channel4_vol_clock()
		}
		//0-7 steps
		s.frame_seq_step += 1
		s.frame_seq_step &= 7

	}

}
func (s *Sound) Update(clocks uint16) {
	if s.doNothing {
		return
	}
	for i := 1; i <= int(clocks); i++ {
		s.Update_channel1()
		s.Update_channel2()
		s.Update_channel3()
		s.Update_channel4()
		s.Freq_sampler()
		s.Sampler()
	}
}

func (s *Sound) chan1_trigger() {
	s.chan1_enabled = true
	s.chan1_vol_period = s.chan1_vol_period_load
	s.chan1_vol = s.chan1_vol_initial
	//sweep
	s.chan1_swp_shadow = s.chan1_freq
	s.chan1_swp_period = s.chan1_swp_period_load

	//The internal enabled flag is set if either the sweep period or shift are non-zero, cleared otherwise.
	if s.chan1_swp_shift != 0 || s.chan1_swp_period != 0 {
		s.chan1_swp_enable = 1
	} else {
		s.chan1_swp_enable = 0
	}
	//If the sweep shift is non-zero, frequency calculation and the overflow check are performed immediately.
	if s.chan1_swp_shift != 0 {
		s.channel1_swp_calc()
	}
	//len
	s.chan1_len_enable = 0
	if s.chan1_len == 0 {
		s.chan1_len = 64
	}
	s.chan1_timer = (2048 - s.chan1_freq) * 4
	//fmt.Println("SND_MODE_1 Trigger 1")

}

func (s *Sound) chan2_trigger() {
	s.chan2_enabled = true
	s.chan2_len_enable = 0
	if s.chan2_len == 0 {
		s.chan2_len = 64
	}
	s.chan2_vol_period = s.chan2_vol_period_load
	s.chan2_vol = s.chan2_vol_initial

	s.chan2_timer = (2048 - s.chan2_freq) * 4
	//fmt.Println("SND_MODE_2 Trigger 2")

}

func (s *Sound) chan3_trigger() {
	s.chan3_enabled = true
	s.chan3_len_enable = 0
	s.chan3_pos = 0
	if s.chan3_len == 0 {
		s.chan3_len = 0xff
	}
	s.chan3_timer = (2048 - s.chan3_freq) * 2
	//fmt.Println("SND_MODE_3 Trigger 3")

}

func (s *Sound) chan4_trigger() {
	s.chan4_enabled = true
	s.chan4_len_enable = 0
	s.chan4_vol_period = s.chan4_vol_period_load
	s.chan4_vol = s.chan4_vol_initial

	if s.chan4_len == 0 {
		s.chan4_len = 64
	}
	s.chan4_timer = uint16(s.chan4_divisor_code[s.chan4_divisor]) << uint16(s.chan4_clk_shift)
	fmt.Println("SND_MODE_4 Trigger 4", s.chan4_timer)
	s.chan4_lfsr = 0x7FFF

}

func NewSound(doNothing bool) *Sound {
	s := new(Sound)
	s.doNothing = doNothing
	s.Setup_SDL()
	/*	Duty   Waveform    Ratio
		-------------------------
		0      00000001    12.5%
		1      10000001    25%
		2      10000111    50%
		3      01111110    75%
	*/
	s.square_duty = [4][8]bool{
		{false, false, false, false, false, false, false, true},
		{true, false, false, false, false, false, false, true},
		{true, false, false, false, false, true, true, true},
		{false, true, true, true, true, true, true, false},
	}
	s.chan4_divisor_code = [8]uint8{8, 16, 32, 48, 64, 80, 96, 112}
	s.sample_p = 0
	s.sample_timer = 4194304 / sample_rate
	s.frame_seq_counter = frame_seq_clocks
	s.chan1_enabled = true
	s.chan2_enabled = true
	s.chan4_enabled = true
	s.chan1_vol = 0xf
	s.chan2_vol = 0xf
	s.chan4_vol = 0xf
	s.dac_power = false

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
		{Name: "SND_WRAM_1", Addr: 0xff30},
		{Name: "SND_WRAM_2", Addr: 0xff31},
		{Name: "SND_WRAM_3", Addr: 0xff32},
		{Name: "SND_WRAM_4", Addr: 0xff33},
		{Name: "SND_WRAM_5", Addr: 0xff34},
		{Name: "SND_WRAM_6", Addr: 0xff35},
		{Name: "SND_WRAM_7", Addr: 0xff36},
		{Name: "SND_WRAM_8", Addr: 0xff37},
		{Name: "SND_WRAM_9", Addr: 0xff38},
		{Name: "SND_WRAM_10", Addr: 0xff39},
		{Name: "SND_WRAM_11", Addr: 0xff3a},
		{Name: "SND_WRAM_12", Addr: 0xff3b},
		{Name: "SND_WRAM_13", Addr: 0xff3c},
		{Name: "SND_WRAM_14", Addr: 0xff3e},
		{Name: "SND_WRAM_15", Addr: 0xff3f},
	}

	return s
}

func (s *Sound) Write_mmio(addr uint16, val uint8) {
	switch addr {
	///CHANN1
	case 0xff10:
		//NR10 FF10 -PPP NSSS Sweep period, negate, shift
		s.SND_MODE_1_SWP = val
		s.chan1_swp_period = (val & 0x70) >> 4
		s.chan1_swp_period_load = s.chan1_swp_period
		s.chan1_swp_negate = (val & 0x8) >> 3
		s.chan1_swp_shift = val & 0x7
		//fmt.Println("SWP negate", s.chan1_swp_negate)
		//fmt.Println("SWP shift", s.chan1_swp_shift)
		//fmt.Println("SWP period", s.chan1_swp_period)

	case 0xff11:
		s.SND_MODE_1_LEN = val
		s.chan1_len = val & 0x1f
		s.chan1_duty = val >> 6
	case 0xff12:
		s.SND_MODE_1_ENVP = val
		s.chan1_vol = (val >> 4) & 0xF
		s.chan1_vol_initial = s.chan1_vol
		s.chan1_vol_period_load = val & 7
		s.chan1_vol_op = (val >> 3) & 1
	case 0xff13:
		s.SND_MODE_1_FREQ_LOW = val
		s.chan1_lo_freq = uint16(val & 0xfe)
		s.chan1_freq = s.chan1_hi_freq + s.chan1_lo_freq

	case 0xff14:
		s.SND_MODE_1_FREQ_HI = val
		hi_freq := uint16(val & 0x7)
		s.chan1_hi_freq = hi_freq << 8
		s.chan1_len_enable = uint8(val&0x40) >> 6
		s.chan1_freq = s.chan1_hi_freq + s.chan1_lo_freq
		//trigger bit has been set

		if val&0x80 == 0x80 {
			s.chan1_trigger()
		}
		s.chan1_len_enable = uint8(val&0x40) >> 6
		//chan2
	case 0xff16:
		s.SND_MODE_2_LEN = val
		s.chan2_len = s.SND_MODE_2_LEN & 0x1f
		s.chan2_duty = val >> 6

	case 0xff17:
		s.SND_MODE_2_ENVP = val
		s.chan2_vol = (val >> 4) & 0xF
		s.chan2_vol_initial = s.chan2_vol
		s.chan2_vol_period_load = val & 7
		s.chan2_vol_op = (val >> 3) & 1
	case 0xff18:
		s.SND_MODE_2_FREQ_LOW = val
		s.chan2_lo_freq = uint16(val & 0xfe)
		s.chan2_freq = s.chan2_hi_freq + s.chan2_lo_freq

	case 0xff19:
		s.SND_MODE_2_FREQ_HI = val
		hi_freq := uint16(val & 0x7)
		s.chan2_hi_freq = hi_freq << 8
		s.chan2_len_enable = uint8(val&0x40) >> 6
		s.chan2_freq = s.chan2_hi_freq + s.chan2_lo_freq
		if val&0x80 == 0x80 {
			s.chan2_trigger()
		}

		s.chan2_len_enable = uint8(val&0x40) >> 6

	case 0xff1a:
		s.SND_MODE_3 = val
		//if DAC is 0x80  bit is set
		if s.SND_MODE_3&0xf8 != 0x00 {
			s.dac_power = true
		} else {
			s.dac_power = false
		}

	case 0xff1b:
		s.SND_MODE_3_LEN = val
		s.chan3_len = val
	case 0xff1c:
		s.SND_MODE_3_OUTPUT = (val & 0x60) >> 5
		s.chan3_vol = (val & 0x60) >> 5
		fmt.Println("SND_3_VOL", s.chan3_vol)

	case 0xff1d:
		s.SND_MODE_3_FREQ_LOW = val
		s.chan3_lo_freq = uint16(val & 0xfe)
		s.chan3_freq = s.chan3_hi_freq | s.chan3_lo_freq

	case 0xff1e:
		s.SND_MODE_3_FREQ_HI = val
		hi_freq := uint16(val & 0x7)
		s.chan3_hi_freq = hi_freq << 8
		s.chan3_len_enable = uint8(val&0x40) >> 6
		s.chan3_freq = s.chan3_hi_freq | s.chan3_lo_freq
		fmt.Println("SND_3_FREQ", s.chan3_freq)
		//s.chan3_timer = (2048 - s.chan3_freq)*2;
		if val&0x80 == 0x80 {
			s.chan3_trigger()
		}

		s.chan3_len_enable = uint8(val&0x40) >> 6

	case 0xff20:
		s.SND_MODE_4_LEN = val & 0x3f
		s.chan4_len = s.SND_MODE_4_LEN

	case 0xff21:
		s.SND_MODE_4_ENVP = val
		s.chan4_vol = (val >> 4) & 0xF
		s.chan4_vol_period_load = val & 7
		s.chan4_vol_op = (val >> 3) & 1
		s.chan4_vol_initial = s.chan4_vol
	case 0xff22:
		s.SND_MODE_4_POLY = val
		s.chan4_divisor = val & 7
		s.chan4_clk_shift = (0xf0 & val) >> 4
		s.chan4_width_mode = (val & 0x8) >> 3
	case 0xff23:
		s.SND_MODE_4_COUNTER = val
		//not sure what happens if both trigger bit and len-en bits are set
		fmt.Println("LEN4 enabled 2 due to timer")

		if val&0x80 == 0x80 {
			s.chan4_trigger()
		}
		s.chan4_len_enable = val & 0x40 >> 6

	case 0xff24:
		s.SND_CHN_CTRL = val
	case 0xff25:
		s.SND_TERM_OUTPUT = val

		//chan 1
		s.channel_enables[0][0] = (val & 1) & chan1_global_enable
		s.channel_enables[0][1] = ((val & 0x10) >> 4) & chan1_global_enable
		//chan 2
		s.channel_enables[1][0] = ((val & 2) >> 1) & chan2_global_enable
		s.channel_enables[1][1] = ((val & 0x20) >> 5) & chan2_global_enable
		//chan3
		s.channel_enables[2][0] = ((val & 4) >> 2) & chan3_global_enable
		s.channel_enables[2][1] = ((val & 0x40) >> 6) & chan3_global_enable

		//chan4
		s.channel_enables[3][0] = ((val & 8) >> 3) & chan4_global_enable
		s.channel_enables[3][1] = ((val & 0x80) >> 7) & chan4_global_enable

	case 0xff26:
		s.SND_MASTER_CTRL = val

	default:
		if addr >= 0xff30 && addr < 0xff40 {
			if s.dac_power {
				//panic("Accessing RAM while channel is enabled")
			}
			baseaddr := (addr & 0x3f) - 0x30
			s.Wram[baseaddr] = val
			fmt.Println("wram", baseaddr, s.Wram)
		} else {
			fmt.Println("SOUND: unhandled sound write", addr)

		}

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
		val = 0x80
	//	val = s.SND_MODE_3

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
		val = 0x0
		//val = s.SND_MASTER_CTRL

	default:
		fmt.Printf("SOUND: unhandled sound read %x\n", addr)
	}
	return val
}
