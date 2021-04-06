package storage

import (
	"errors"
	"gogoscrap/repository"
	"sync"
)

type Storage struct {
	configs []repository.UserConfig
	mu      sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Insert(userConfig repository.UserConfig) {
	s.mu.Lock()
	s.configs = append(s.configs, userConfig)
	s.mu.Unlock()
}

func (s *Storage) Remove(userID int64) {
	id := s.FindID(userID)

	if id == -1 {
		return
	}

	s.mu.Lock()
	s.configs = append(s.configs[:id], s.configs[:id+1]...)
	s.mu.Unlock()
}

func (s *Storage) FindID(userID int64) int {
	id := -1

	for i, c := range s.configs {
		if c.UserID == userID {
			id = i
			break
		}
	}

	return id
}

func (s *Storage) Get(userID int64) (*repository.UserConfig, error) {
	id := s.FindID(userID)
	if id == -1 {
		return nil, errors.New("not found")
	}
	return &s.configs[id], nil
}
