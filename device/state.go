package device

type state string

const (
	playing = "playing"
	stopped = "stopped"
)

func newState() state {
	return stopped
}

func (s *state) play() {
	*s = playing
}

func (s *state) stop() {
	*s = stopped
}

func (s state) playing() bool {
	return s == playing
}

func (s state) stopped() bool {
	return s == stopped
}
