package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tenteedee/gopher-social/internal/store"
)

const UserExpTime = 5 * time.Minute

type UserStore struct {
	rdb *redis.Client
}

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%d", id)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%d", user.ID)
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	if err := s.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err(); err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {

	return nil
}
