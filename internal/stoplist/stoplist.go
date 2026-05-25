package stoplist

import "sync"

type StopList struct {
	mu    sync.RWMutex
	words map[string]struct{}
}

func New() *StopList {
	return &StopList{words: make(map[string]struct{})}
}

func (s *StopList) Add(word string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.words[word] = struct{}{}
}

func (s *StopList) Remove(word string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.words, word)
}

func (s *StopList) Words() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]string, 0, len(s.words))
	for w := range s.words {
		res = append(res, w)
	}
	return res
}
