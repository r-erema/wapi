package message

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// RedisRepository stores messages metadata via Redis.
type RedisRepository struct {
	client              *redis.Client
	storeExpirationTime time.Duration
}

// NewRedis creates redis repository.
func NewRedis(host string) (*RedisRepository, error) {
	redisClient := redis.NewClient(&redis.Options{Addr: host})
	if _, err := redisClient.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "redis client creation error")
	}
	return &RedisRepository{client: redisClient, storeExpirationTime: time.Hour * 24 * 30}, nil
}

// SaveMessageTime stores message time in repository.
func (r *RedisRepository) SaveMessageTime(msgID string, msgTime time.Time) error {
	return r.client.Set(timeKey(msgID), msgTime.UnixNano(), r.storeExpirationTime).Err()
}

// MessageTime retrieves message time from repository.
func (r *RedisRepository) MessageTime(msgID string) (*time.Time, error) {
	timestamp, err := r.client.Get(timeKey(msgID)).Result()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get timestamp by message id")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse timestamp")
	}

	t := time.Unix(ts, 0)
	return &t, nil
}

func timeKey(msgID string) string {
	return "msg_timestamp:" + msgID
}
