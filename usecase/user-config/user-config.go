package user_config

import (
	"gogoscrap/repository"
)

type UserService struct {
	repo repository.Repository
}

func NewService(r repository.Repository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) Save(config repository.UserConfig) {
	s.repo.Remove(config.UserID)
	s.repo.Insert(config)
}

func (s *UserService) RemoveConfigByUser(userID int64) {
	s.repo.Remove(userID)
}

func (s *UserService) GetByUser(userID int64) (*repository.UserConfig, error) {
	userConfig, err := s.repo.Get(userID)
	return userConfig, err
}
