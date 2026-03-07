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
	defer s.mu.RUnlock()

	for _, ch := range s.ch {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
