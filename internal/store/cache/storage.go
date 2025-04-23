package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/tenteedee/gopher-social/internal/store"
)

type Storage struct {
	User interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
		Delete(context.Context, int64) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		User: &UserStore{rdb},
	}
}
