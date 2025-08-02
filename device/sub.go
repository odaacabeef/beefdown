package device

import "sync"

type sub struct {
	ch map[string]chan struct{}
	mu sync.RWMutex
}

func (s *sub) Sub(name string, ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ch[name] = ch
}

func (s *sub) Unsub(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ch, name)
}

func (s *sub) Pub() {
	s.mu.RLock()
	channels := make([]chan struct{}, 0, len(s.ch))
	for _, ch := range s.ch {
		channels = append(channels, ch)
	}
	s.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- struct{}{}:
			// Message sent successfully
		default:
			// Channel is full or closed, skip it
		}
	}
}
