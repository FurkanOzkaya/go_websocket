package common

import (
	"sync"

	"log"

	"github.com/go-redis/redis"
)

// Redis client
type Redis struct {
	*redis.Client
}

var redisInstance *Redis

var once sync.Once

func GetRedis() *Redis {
	once.Do(func() {

		redisInstance = &Redis{
			Client: redis.NewClient(&redis.Options{
				Addr:     "redis:6379",
				Password: "",
				DB:       0,
			}),
		}
	})
	pong, err := redisInstance.Client.Ping().Result()
	if err != nil {
		log.Fatal(err)

	}
	log.Println("redis", pong)

	return redisInstance
}
