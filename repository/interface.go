package repository

type Repository interface {
	Insert(config UserConfig)
	Remove(userID int64)
	FindID(userID int64) int
	Get(userID int64) (*UserConfig, error)
}

type UserConfig struct {
	UserID  int64
	Website string
}
