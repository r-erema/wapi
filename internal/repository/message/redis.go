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

func (r *RedisRepository) SaveMessageTime(msgID string, msgTime time.Time) error {
	return r.client.Set(timeKey(msgID), msgTime.UnixNano(), r.storeExpirationTime).Err()
}

func (r *RedisRepository) GetMessageTime(msgID string) (*time.Time, error) {
	timestamp, err := r.client.Get(timeKey(msgID)).Result()
	if err != nil {
		return nil, err
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, err
	}

	t := time.Unix(ts, 0)
	return &t, nil
}

func timeKey(msgID string) string {
	return "msg_timestamp:" + msgID
}
