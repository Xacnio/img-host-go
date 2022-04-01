package database

import (
	"context"
	"fmt"
	"github.com/Xacnio/img-host-go/pkg/configs"
	"github.com/go-redis/redis/v8"
	"time"
)

var ctx = context.Background()

type RedisCon struct {
	rdb *redis.Client
}

func NewRConnection() RedisCon {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", configs.Get("REDIS_HOST"), configs.Get("REDIS_PORT")),
		Password: configs.Get("REDIS_PASSWORD"), // no password set
		DB:       configs.GetInt("REDIS_DB", 0),  // use default DB
	})
	return RedisCon{rdb: rdb}
}

func (rdb* RedisCon) RClose() {
	rdb.rdb.Close()
}

func (rdb* RedisCon) RPing() error {
	_, errP := rdb.rdb.Ping(ctx).Result()
	if errP != nil {
		return errP
	}
	return nil
}


func (rdb* RedisCon) RGet(key string) (string, error) {
	val2, err := rdb.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val2, nil
	}
}

func (rdb* RedisCon) RSet(key, value string) error {
	errs := rdb.rdb.Set(ctx, key, value, 0).Err()
	return errs
}

func (rdb* RedisCon) RSetTTL(key, value string, ttl_second time.Duration) error {
	errs := rdb.rdb.Set(ctx, key, value, ttl_second * time.Second).Err()
	return errs
}