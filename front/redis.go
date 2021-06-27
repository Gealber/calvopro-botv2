package front

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	log "gopkg.in/inconshreveable/log15.v2"
)

var (
	REDIS_URL      = os.Getenv("REDIS_URL")
	REDIS_PASSWORD = os.Getenv("REDIS_PASSWORD")
)

type redisRepo struct {
	Logger log.Logger
	rdb    *redis.Client
}

//NewRedisRepo ...
func NewRedisRepo() *redisRepo {
	logger := log.New("function", "redis")
	return &redisRepo{
		Logger: logger,
	}
}

//Connect to redis repo
func (repo *redisRepo) Connect() {
	repo.rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_URL,
		Password: REDIS_PASSWORD,
		DB:       0,
	})
	//make a ping to check connectivity
	pong, err := repo.rdb.Ping(context.Background()).Result()
	if err != nil {
		repo.Logger.Error("Unable to make Ping to redis server", "err", err)
		return
	}
	repo.Logger.Info(fmt.Sprintf("Ping <--> %s", pong), "host", REDIS_URL)
}

//Get the specified Key from Redis
func (repo *redisRepo) Get(key string) string {
	ctx := context.Background()
	val, err := repo.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		repo.Logger.Error(fmt.Sprintf("Unable to retrieve Key: %s", key), "err", err)
		return ""
	}
	return val
}

//Set the specified Key:Value pair into Redis
func (repo *redisRepo) Set(key, value string, expiration time.Duration) {
	ctx := context.Background()
	//I don't care if fail
	repo.rdb.Set(ctx, key, value, expiration)
}

func (repo *redisRepo) BRPopLPush() (string, error) {
	ctx := context.Background()
	return repo.rdb.BRPopLPush(ctx, RESULT_QUEUE, BACKUP_RESULT_QUEUE, 10*time.Second).Result()
}

func (repo *redisRepo) LPush(data string) (int64, error) {
	ctx := context.Background()
	return repo.rdb.LPush(ctx, TASK_QUEUE, data).Result()
}

func (repo *redisRepo) Incr(key string) (int64, error) {
	ctx := context.Background()
	return repo.rdb.Incr(ctx, key).Result()
}
