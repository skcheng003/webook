package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/skcheng003/webook/internal/domain"
	"time"
)

// ErrKeyNotExist is the error when key not exist
var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

// RedisUserCache Programing with interface
type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// NewRedisUserCache Dependency injection
// If A uses B, B should be an interface
// If A uses B, B should be a property of A 吗
// If A uses B, A should not initialize B, B should be initialized outside A
func NewRedisUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// Get gets user info from cache
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err == ErrKeyNotExist {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal(val, &user)
	return user, err
}

// Set sets user info to cache
func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	err = cache.client.Set(ctx, key, val, cache.expiration).Err()
	return err
}

// key generates key for user info
func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
