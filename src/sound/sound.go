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
	chan1_curr_freq uint32;
	chan2_curr_freq uint32;
	chan3_curr_freq uint32;
	chan2_real_freq uint32;
	chan3_real_freq uint32;



	dac_power bool

	chan1_duty uint8;
	chan1_hi_freq uint16;
	chan1_lo_freq uint16;
	chan1_freq uint16;
	chan1_timer uint16;
	chan1_len_mode uint16;
	chan1_len uint8;
	chan1_len_enable uint8;
	chan1_enabled bool

	chan1_swp_shadow uint16
	chan1_swp_period uint8
	chan1_swp_period_load uint8
	chan1_swp_negate uint8
	chan1_swp_shift uint8
	chan1_swp_enable uint8

	chan1_vol uint8;
	chan1_vol_initial uint8
	chan1_duty_pointer uint8;
	chan1_vol_period uint8;
	chan1_vol_period_load uint8;
	chan1_vol_op uint8;


	chan2_duty uint8;
	chan2_hi_freq uint16;
	chan2_lo_freq uint16;
	chan2_freq uint16;
	chan2_timer uint16;
	chan2_len_mode uint16;
	chan2_len uint8;
	chan2_len_enable uint8;
	chan2_enabled bool
	chan2_duty_pointer uint8;

	chan2_vol uint8;
	chan2_vol_initial uint8
	chan2_vol_period uint8;
	chan2_vol_period_load uint8;

	chan2_vol_op uint8;


	chan1_sample_p uint32;
	sample_timer uint32;
	csample[sample_size] byte;

	square_duty[4][8] bool;
	channel_enables[4][2] uint8;

	frame_seq_counter uint32
	frame_seq_step uint8



	


}

const (
	sample_rate = 48000
	channels    = 2
	samples     = 2048
	sample_size=4096
	frame_seq_clocks = 8192
)

func (g *Sound) Get_reg_list() component.RegList {
	return g.reg_list
}

func (s *Sound) Setup_SDL() {
	var desired sdl.AudioSpec
	desired.Freq = sample_rate
	desired.Format =sdl.AUDIO_U8
	desired.Channels=2
	desired.Silence = 0
	desired.Samples = sample_size
	var recv sdl.AudioSpec

	sdl.OpenAudio(&desired,&recv)

	sdl.PauseAudio(false)

}
func (s *Sound) Reset() {
}
func (s *Sound) Update_channel1() {
	s.chan1_timer-=1
	if s.chan1_timer == 0 {

		s.chan1_duty_pointer += 1
		s.chan1_duty_pointer &= 7
		//fmt.Println(s.clocks,"Current Timer",s.chan1_timer)
		//fmt.Println(s.clocks,"Duty",s.chan1_duty_pointer,s.square_duty[s.chan1_duty][s.chan1_duty_pointer],s.chan1_sample_p)
		s.chan1_timer = (2048 - s.chan1_freq)*4;

	}
	if ! s.chan1_enabled {

	}
}

func (s *Sound) Update_channel2() {
	s.chan2_timer-=1
	if s.chan2_timer == 0 {

		s.chan2_duty_pointer += 1
		s.chan2_duty_pointer &= 7
		//fmt.Println(s.clocks,"Current Timer",s.chan1_timer)
		//fmt.Println(s.clocks,"Duty",s.chan1_duty_pointer,s.square_duty[s.chan1_duty][s.chan1_duty_pointer],s.chan1_sample_p)
		s.chan2_timer = (2048 - s.chan2_freq)*4;

	}
	if ! s.chan2_enabled {
		s.chan2_vol = 0
	}
}
func (s *Sound) Sampler() {
	s.sample_timer -=1
	if s.sample_timer == 0 {

		s.sample_timer= (4194304 / sample_rate)
		s.csample[s.chan1_sample_p]=0
		s.csample[s.chan1_sample_p+1]=0
		
		chan1_vol :=s.square_duty[s.chan1_duty][s.chan1_duty_pointer]
		chan2_vol :=s.square_duty[s.chan2_duty][s.chan2_duty_pointer]

		if s.channel_enables[0][0] == 1  && s.dac_power {
			if chan1_vol == true{
				s.csample[s.chan1_sample_p] += 0;
			}else {
				
				s.csample[s.chan1_sample_p] +=(uint8(int(s.chan1_vol)*30)/15);

			}
		}

		if s.channel_enables[1][0] == 1 && s.dac_power {
			if chan2_vol == true{
				s.csample[s.chan1_sample_p] += 0;
			}else {
				s.csample[s.chan1_sample_p] +=(uint8(int(s.chan2_vol)*30)/15);
			}
		}

		fmt.Println("wave Left:",s.csample[s.chan1_sample_p])

		if s.channel_enables[0][1] == 1  && s.dac_power {
			if chan1_vol == true{
				s.csample[s.chan1_sample_p+1] += 0;
			}else {
				s.csample[s.chan1_sample_p+1] +=(uint8(int(s.chan1_vol)*30)/15)
			}
		}	

		if s.channel_enables[1][1] == 1 && s.dac_power {
			if chan2_vol == true{
				s.csample[s.chan1_sample_p+1] += 0;
			}else {
				s.csample[s.chan1_sample_p+1] += uint8((int(s.chan2_vol)*30)/15);
			}
		}
		fmt.Println("wave right:",s.csample[s.chan1_sample_p])


		if s.chan1_sample_p+2 >= sample_size-1 {
			p := s.csample
			s.chan1_sample_p =0
			sdl.QueueAudio(1,p[0:sample_size])


		}else
		{
			s.chan1_sample_p += 2
		}


		for ((sdl.GetQueuedAudioSize(1)) > sample_size) {
		}

	}

}
func (s *Sound) channel1_len_clock() {
	if s.chan1_len_enable  == 1  {
		s.chan1_len-=1
		if s.chan1_len ==0 {
			fmt.Println("Disabled due to timer")
			s.chan1_enabled = false
			s.chan1_len_enable= 0
		}
	}
}

func (s *Sound) channel2_len_clock() {
	if s.chan2_len_enable  == 1  {
		s.chan2_len-=1
		if s.chan2_len ==0 {
			fmt.Println("Disabled 2 due to timer")
			s.chan2_enabled = false
			s.chan2_len_enable= 0
		}
	}
}

func (s *Sound) channel1_vol_clock() {
	if s.chan1_vol_period  != 0  {
		s.chan1_vol_period -=1
		if s.chan1_vol_period ==0 {
			if s.chan1_vol_op == 1  && s.chan1_vol <15 {
				s.chan1_vol +=1
			}
			if s.chan1_vol_op == 0  && s.chan1_vol >0 {
				s.chan1_vol -=1
			}
			s.chan1_vol_period = s.chan1_vol_period_load
			fmt.Println("KYLEtest chan1",s.chan1_vol,s.chan1_vol_period)

		}
	}
}

func (s *Sound) channel2_vol_clock() {
	if s.chan2_vol_period  != 0  {
		s.chan2_vol_period -=1
		if s.chan2_vol_period ==0 {
			if s.chan2_vol_op == 1  && s.chan2_vol <15 {
				s.chan2_vol +=1
			}
			if s.chan2_vol_op == 0  && s.chan2_vol >0 {
				s.chan2_vol -=1
			}
			s.chan2_vol_period = s.chan2_vol_period_load
			fmt.Println("KYLEtest chan2",s.chan2_vol,s.chan2_vol_period)


		}
	}
}
func (s *Sound) channel1_swp_calc () uint16 {
	new_freq := s.chan1_swp_shadow >>  s.chan1_swp_shift

	if s.chan1_swp_negate == 1 {
		new_freq -= s.chan1_swp_shadow
	}else {
		new_freq += s.chan1_swp_shadow

	}
	if new_freq > 2047 {
		s.chan1_enabled = false
		s.chan1_swp_enable = 0
		fmt.Println("SWP Disabled channel 1 due to freq overflow")
	}
	return new_freq

}
func (s *Sound) channel1_swp_clock() {

	s.chan1_swp_period -=1 
	if s.chan1_swp_enable == 1 &&  s.chan1_swp_period_load >0 {
		if s.chan1_swp_period == 0 {
			s.chan1_swp_period = s.chan1_swp_period_load
			new_calc := s.channel1_swp_calc()
			if s.chan1_swp_shift >0 && new_calc < 2047 {
				fmt.Println("SWP New freq",new_calc,s.chan1_freq)
				s.chan1_freq = new_calc
				s.chan1_swp_shadow = new_calc
			}
			s.channel1_swp_calc()
				

		}
	}
}



func (s *Sound) Freq_sampler() {
	s.frame_seq_counter -= 1

	if 	s.frame_seq_counter	== 0 {
		s.frame_seq_counter = frame_seq_clocks
		switch s.frame_seq_step {

			case 0:
				s.channel1_len_clock()
				s.channel2_len_clock()
			case 2:
				s.channel1_len_clock()
				s.channel1_swp_clock()
				s.channel2_len_clock()
			case 4:
				s.channel1_len_clock()
				s.channel2_len_clock()
			case 6:
				s.channel1_len_clock()
				s.channel1_swp_clock()
				s.channel2_len_clock()
			case 7:
				s.channel1_vol_clock()
				s.channel2_vol_clock()
		}
		//0-7 steps
		s.frame_seq_step +=1
		s.frame_seq_step &=7

	}

}
func (s *Sound) Update(clocks uint16) {
	for i := 1;  i<=int(clocks); i++ {

		s.Update_channel1()
		s.Update_channel2()
		s.Freq_sampler()
		s.Sampler()

}
		

}
func (s *Sound) Update_channel1_regs()  {
	hi_freq := uint16(s.SND_MODE_1_FREQ_HI &0x7)
	s.chan1_hi_freq = hi_freq <<8
	s.chan1_lo_freq = uint16(s.SND_MODE_1_FREQ_LOW)
	s.chan1_len_enable = uint8(s.SND_MODE_1_FREQ_HI &0x40) >>6
	s.chan1_len =s.SND_MODE_1_LEN &0x1f
	s.chan1_freq =s.chan1_hi_freq+s.chan1_lo_freq
	//real_freq := 131072/uint32((2048-freq))
	//NR11 FF11 DDLL LLLL Duty, Length load (64-L)
	s.chan1_duty = s.SND_MODE_1_LEN >> 6
	s.chan1_timer = (2048 - s.chan1_freq)*4;

	//fmt.Println("Duty Cycle1:",s.chan1_duty)
	//fmt.Println(s.chan1_freq)
	//fmt.Println("Len:",s.chan1_len)
}
func (s *Sound) Update_channel2_regs()  {
	hi_freq := uint16(s.SND_MODE_2_FREQ_HI &0x7)
	s.chan2_hi_freq = hi_freq <<8
	s.chan2_lo_freq = uint16(s.SND_MODE_2_FREQ_LOW&0xfe)
	s.chan2_len_enable = uint8(s.SND_MODE_2_FREQ_HI &0x40) >>6
	s.chan2_len =s.SND_MODE_2_LEN &0x1f
	s.chan2_freq =s.chan2_hi_freq+s.chan2_lo_freq 
	//real_freq := 131072/uint32((2048-freq))
	//NR11 FF11 DDLL LLLL Duty, Length load (64-L)
	s.chan2_duty = s.SND_MODE_2_LEN >> 6
	s.chan2_timer = (2048 - s.chan2_freq)*4;

	//fmt.Println(s.chan2_freq)

}
func (s *Sound) chan1_trigger()  {
	s.chan1_enabled = true
	s.chan1_vol_period = s.chan1_vol_period_load
	s.chan1_vol =  s.chan1_vol_initial
	//sweep
	s.chan1_swp_shadow = s.chan1_freq
	s.chan1_swp_period = s.chan1_swp_period_load
	
	//The internal enabled flag is set if either the sweep period or shift are non-zero, cleared otherwise.
	if s.chan1_swp_shift != 0 || s.chan1_swp_negate != 0 {
		s.chan1_swp_enable = 1
	}else {
		s.chan1_swp_enable = 0
	}
	//If the sweep shift is non-zero, frequency calculation and the overflow check are performed immediately.
	if s.chan1_swp_shift != 0 {
		s.channel1_swp_calc()
	}
	//len
	s.chan1_len_enable = 0
	if s.chan1_len == 0 {
		s.chan1_len =64
	}
	s.chan1_timer = (2048 - s.chan1_freq)*4
	fmt.Println("Trigger 1")

}

func (s *Sound) chan2_trigger()  {
	s.chan2_enabled = true
	s.chan2_len_enable = 0
	if s.chan2_len == 0 {
		s.chan2_len =64
	}
	s.chan2_vol_period = s.chan2_vol_period_load
	s.chan2_vol =  s.chan2_vol_initial

	s.chan2_timer = (2048 - s.chan2_freq)*4
	fmt.Println("Trigger 2")

}


func NewSound() *Sound {
	s := new(Sound)
	s.Setup_SDL()
/*	Duty   Waveform    Ratio
-------------------------
0      00000001    12.5%
1      10000001    25%
2      10000111    50%
3      01111110    75%
*/
	s.square_duty = [4][8]bool{
		{false,false,false,false,false,false,false,true},
		{true,false,false,false,false,false,false,true},
		{true,false,false,false,false,true,true,true},
		{false,true,true,true,true,true,true,false},
	}
	s.chan1_sample_p = 0
	s.sample_timer = 4194304 / sample_rate
	s.frame_seq_counter  = frame_seq_clocks
	s.chan1_enabled = true
	s.chan2_enabled = true
	s.chan1_vol = 30
	s.chan2_vol =30
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
	}

	return s
}

func (s *Sound) Write_mmio(addr uint16, val uint8) {
	switch addr {
	case 0xff10:
		//NR10 FF10 -PPP NSSS Sweep period, negate, shift
		s.SND_MODE_1_SWP = val
		s.chan1_swp_period = (val & 0x70) >>4
		s.chan1_swp_period_load = s.chan1_swp_period
		s.chan1_swp_negate = (val &0x8) >> 3
		s.chan1_swp_shift = val &0x7
		
	case 0xff11:
		s.SND_MODE_1_LEN = val
		s.Update_channel1_regs()
	case 0xff12:
		s.SND_MODE_1_ENVP = val
		s.Update_channel1_regs()
		s.chan1_vol = (val >> 4) & 0xF
		s.chan1_vol_initial = s.chan1_vol
		s.chan1_vol_period = val & 7
		s.chan1_vol_period_load = s.chan1_vol_period

		s.chan1_vol_op = (val >>3) &1
	case 0xff13:
		s.SND_MODE_1_FREQ_LOW = val
		s.Update_channel1_regs()
	case 0xff14:
		s.SND_MODE_1_FREQ_HI = val
		s.Update_channel1_regs()

		//trigger bit has been set

		if val & 0x80 == 0x80 { 
			s.chan1_trigger()
		}

	case 0xff16:
		s.SND_MODE_2_LEN = val
		s.Update_channel2_regs()

	case 0xff17:
		s.SND_MODE_2_ENVP = val
		s.chan2_vol = (val >> 4) & 0xF
		s.chan2_vol_period = val & 7
		s.chan2_vol_initial  =s.chan2_vol
		s.chan2_vol_period_load = s.chan2_vol_period
		s.chan2_vol_op = (val >>3) &1
	case 0xff18:
		s.SND_MODE_2_FREQ_LOW = val
		s.Update_channel2_regs()


	case 0xff19:
		s.SND_MODE_2_FREQ_HI = val
		s.Update_channel2_regs()

		//trigger bit has been set
		if val & 0x80 == 0x80 { 
			s.chan2_trigger()
		}


	case 0xff1a:
		s.SND_MODE_3 = val
		//if DAC is 0x80  bit is set
		if s.SND_MODE_3 & 0xf8 != 0x00 {
			s.dac_power = true
		}else{
			s.dac_power = false
		}
		
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
		//chan 1
		s.channel_enables[0][0] = val &1
		s.channel_enables[0][1] = (val & 0x10) >>4
		//chan 2
		s.channel_enables[1][0] = (val &2) >>1
		s.channel_enables[1][1] = (val & 0x20) >>5
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
