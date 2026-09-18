package keys

import "sync"

type Store struct {
	mu   sync.RWMutex
	keys map[string]Key
}

func NewStore() *Store {
	return &Store{
		keys: make(map[string]Key),
	}
}

func (s *Store) SetKeys(keys []Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		s.keys[key.KeyId] = key
	}

	return nil
}

func (s *Store) GetKeys() ([]Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]Key, 0, len(s.keys))
	for _, key := range s.keys {
		keys = append(keys, key)
	}

	return keys, nil
}

func (s *Store) DeleteKeys(keys []Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.keys, key.KeyId)
	}

	return nil
}
