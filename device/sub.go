package device

import "sync"

type sub struct {
	ch map[string]chan struct{}
	mu sync.Mutex
}

func (s *sub) sub(name string, ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ch[name] = ch
}

func (s *sub) unsub(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ch, name)
}

func (s *sub) pub() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.ch {
		ch <- struct{}{}
	}
}
