package redis

import (
	"ferdinand/library-management-system-api/internal/config"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.RedisConfig) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",cfg.Host,cfg.Port),
		Password: "",
		DB: 0,
	})



	return  client

}