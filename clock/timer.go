package clock

type Clock struct {
	Cycles uint64
}

func NewClock() *Clock {
	timer := new(Clock)
	return timer
}
