package message

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type RedisRepository struct {
	client              *redis.Client
	storeExpirationTime time.Duration
}

func NewRedisRepository(host string) (*RedisRepository, error) {
	redisClient := redis.NewClient(&redis.Options{Addr: host})
	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisRepository{client: redisClient, storeExpirationTime: time.Hour * 24 * 30}, nil
}

func (r *RedisRepository) SaveMessageTime(msgId string, msgTime time.Time) error {
	return r.client.Set(timeKey(msgId), msgTime.UnixNano(), r.storeExpirationTime).Err()
}

func (r *RedisRepository) GetMessageTime(msgId string) (time.Time, error) {
	timestamp, err := r.client.Get(timeKey(msgId)).Result()
	if err != nil {
		return time.Time{}, err
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(ts, 0), nil
}

func timeKey(msgId string) string {
	return "msg_timestamp:" + msgId
}
