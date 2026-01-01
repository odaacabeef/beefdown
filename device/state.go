package device

import "sync/atomic"

type state uint32

const (
	playing state = 1
	stopped state = 0
)

func newState() state {
	return stopped
}

func (s *state) play() {
	atomic.StoreUint32((*uint32)(s), uint32(playing))
}

func (s *state) stop() {
	atomic.StoreUint32((*uint32)(s), uint32(stopped))
}

func (s *state) playing() bool {
	return atomic.LoadUint32((*uint32)(s)) == uint32(playing)
}

func (s *state) stopped() bool {
	return atomic.LoadUint32((*uint32)(s)) == uint32(stopped)
}

func (s *state) string() string {
	switch atomic.LoadUint32((*uint32)(s)) {
	case uint32(playing):
		return "playing"
	case uint32(stopped):
		return "stopped"
	}
	return ""
}

func (d *Device) State() string {
	return d.state.string()
}

func (d *Device) Playing() bool {
	return d.state.playing()
}

func (d *Device) Stopped() bool {
	return d.state.stopped()
}
