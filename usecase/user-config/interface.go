package user_config

import "gogoscrap/repository"

type Service interface {
	Save(config repository.UserConfig)
	RemoveConfigByUser(userID int64)
	GetByUser(userID int64) (*repository.UserConfig, error)
}
