package store

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// Redis represents a Redis-based backing store.
type Redis struct {
	*Store
	client *redis.Client
	prefix string // all redis keys are prefixed with this
}

func NewRedisStore(address string) (*Redis, error) {
	// Try to connect
	log.WithFields(log.Fields{"type": "store", "store": "redis"}).Infof("Connecting to redis at %v", address)
	client := redis.NewClient(&redis.Options{Addr: address})

	if err := client.Ping().Err(); err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Infof("connection failed:", err)
		return nil, fmt.Errorf("can't reach redis server at %s: %v", address, err)
	}

	return &Redis{
		Store: &Store{
			driver:     "redis",
			parameters: address,
		},
		client: client,
	}, nil
}

func (store *Redis) WithPrefix(prefix string) *Redis {
	store.prefix = prefix + ":"
	return store
}

func (store *Redis) Set(k string, v string, ttl time.Duration) error {
	err := store.client.Set(store.prefix+k, v, ttl).Err()
	if err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Errorf("Set: %v", err)
	}
	return err
}

func (store *Redis) Get(k string) (string, error) {
	val, err := store.client.Get(store.prefix + k).Result()
	if err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Debugf("Get: %v", err)
	}
	return val, err
}

func (store *Redis) Delete(k string) error {
	err := store.client.Del(store.prefix + k).Err()
	if err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Errorf("Del: %v", err)
	}
	return err
}
