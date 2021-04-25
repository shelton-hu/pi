package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/shelton-hu/logger"
	"github.com/shelton-hu/util/idutil"

	"github.com/shelton-hu/pi/config"
)

// Redis is a instance for calling most of the package's methods.
type Redis struct {
	// conn is the connection between the server and the client.
	conn *redis.PubSubConn

	// rdsKeyPrefix is the prefix of all redis key.
	rdsKeyPrefix string

	// ctx is used for logger and jeager record.
	ctx context.Context
}

var rdsPool *redis.Pool
var rdsKeyPrefix string

// ConnectRedis connects to redis.
func ConnectRedis(ctx context.Context, redisConfig config.Redis) {
	address := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
	rdsPool = &redis.Pool{
		Wait:        true,
		MaxIdle:     redisConfig.MaxIdle,
		MaxActive:   redisConfig.MaxActive,
		IdleTimeout: time.Duration(redisConfig.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address, redis.DialPassword(redisConfig.Password))
			if err != nil {
				logger.Error(ctx, err.Error())
				return nil, err
			}
			return conn, nil
		},
	}
	rdsKeyPrefix = redisConfig.KeyPrefix
	if rdsKeyPrefix == "" {
		rdsKeyPrefix = idutil.Gen8LetterUuid()
	}
}

// CloseRedis closes connect to redis.
func CloseRedis(ctx context.Context) {
	if rdsPool != nil {
		if err := rdsPool.Close(); err != nil {
			logger.Error(ctx, err.Error())
		}
	}
}

// GetConnect get a instance from this redis pool.
func GetConnect(ctx context.Context) *Redis {
	r := &Redis{
		conn: &redis.PubSubConn{
			Conn: rdsPool.Get(),
		},
		rdsKeyPrefix: rdsKeyPrefix,
		ctx:          ctx,
	}
	return r
}
