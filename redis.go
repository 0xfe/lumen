package cli

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// DataStore represents the conntection to the Google Cloud Datastore.
type Redis struct {
	store  *Store
	client *redis.Client
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
		store: &Store{
			driver:     "redis",
			parameters: address,
		},
		client: client,
	}, nil
}

func (store *Redis) Set(k string, v string, ttl time.Duration) error {
	err := store.client.Set(k, v, ttl).Err()
	if err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Errorf("Set: %v", err)
	}
	return err
}

func (store *Redis) Get(k string) (string, error) {
	val, err := store.client.Get(k).Result()
	if err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Debugf("Get: %v", err)
	}
	return val, err
}

func (store *Redis) Delete(k string) error {
	err := store.client.Del(k).Err()
	if err != nil {
		log.WithFields(log.Fields{"type": "store", "store": "redis"}).Errorf("Del: %v", err)
	}
	return err
}
