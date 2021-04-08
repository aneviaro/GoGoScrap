package storage

import (
	"errors"
	"gogoscrap/repository"
	"sync"
)

type Repo struct {
	configs []repository.UserConfig
	mu      sync.RWMutex
}

func NewRepo() *Repo {
	return &Repo{}
}

func (s *Repo) Insert(userConfig repository.UserConfig) {
	s.mu.Lock()
	s.configs = append(s.configs, userConfig)
	s.mu.Unlock()
}

func (s *Repo) Remove(userID int64) {
	id := s.FindID(userID)

	if id == -1 {
		return
	}

	s.mu.Lock()
	s.configs[id] = s.configs[len(s.configs)-1]
	s.configs[len(s.configs)-1] = repository.UserConfig{
		UserID:  0,
		Website: "",
	}
	s.configs = s.configs[:len(s.configs)-1]
	s.mu.Unlock()
}

func (s *Repo) FindID(userID int64) int {
	id := -1

	for i, c := range s.configs {
		if c.UserID == userID {
			id = i
			break
		}
	}

	return id
}

func (s *Repo) Get(userID int64) (*repository.UserConfig, error) {
	id := s.FindID(userID)
	if id == -1 {
		return nil, errors.New("not found")
	}
	return &s.configs[id], nil
}
