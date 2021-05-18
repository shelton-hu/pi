package redis

import (
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// Set ...
func (r *Redis) Set(key string, value interface{}, expire int) error {
	_, err := r.do("Set", key, value, "EX", expire)
	return err
}

// SetNX returns (true, nil) if success, or returns (false, error).
func (r *Redis) SetNX(key string, value interface{}, expire int) (bool, error) {
	reply, err := r.do("Set", key, value, "EX", expire, "NX")
	return reply != nil, err
}

// Incr ...
func (r *Redis) Incr(key string, value interface{}, expire int) (int, error) {
	reply, err := redis.Int(r.do("INCR", key))
	if err != nil {
		return 0, err
	}
	if reply == 1 {
		// if reply is 1, set expire.
		if _, err := r.do("EXPIRE", key, expire); err != nil {
			return 0, err
		}
	}
	return reply, nil
}

// Decr ...
func (r *Redis) Decr(key string) (int, error) {
	reply, err := redis.Int(r.do("DECR", key))
	if err != nil {
		return 0, err
	}
	return reply, nil
}

// Expire ...
func (r *Redis) Expire(key string, expire int) error {
	_, err := r.do("EXPIRE", key, expire)
	return err
}

// Get ...
func (r *Redis) Get(key string) (interface{}, error) {
	reply, err := r.do("Get", key)
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Int
func (r *Redis) Int(key string) (int, error) {
	reply, err := redis.Int(r.Get(key))
	if err == redis.ErrNil {
		return 0, nil
	}
	return reply, err
}

// Int64 ...
func (r *Redis) Int64(key string) (int64, error) {
	reply, err := redis.Int64(r.Get(key))
	if err == redis.ErrNil {
		return 0, nil
	}
	return reply, err
}

// Float64 ...
func (r *Redis) Float64(key string) (float64, error) {
	reply, err := redis.Float64(r.Get(key))
	if err == redis.ErrNil {
		return 0, nil
	}
	return reply, err
}

// String ...
func (r *Redis) String(key string) (string, error) {
	reply, err := redis.String(r.Get(key))
	if err == redis.ErrNil {
		return "", nil
	}
	return reply, err
}

// Bytes ...
func (r *Redis) Bytes(key string) ([]byte, error) {
	reply, err := redis.Bytes(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Ints ...
func (r *Redis) Ints(key string) ([]int, error) {
	reply, err := redis.Ints(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Int64s ...
func (r *Redis) Int64s(key string) ([]int64, error) {
	reply, err := redis.Int64s(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Float64s ...
func (r *Redis) Float64s(key string) ([]float64, error) {
	reply, err := redis.Float64s(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Strings ...
func (r *Redis) Strings(key string) ([]string, error) {
	reply, err := redis.Strings(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// ByteSlices ...
func (r *Redis) ByteSlices(key string) ([][]byte, error) {
	reply, err := redis.ByteSlices(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// IntMap ...
func (r *Redis) IntMap(key string) (map[string]int, error) {
	reply, err := redis.IntMap(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Int64Map ...
func (r *Redis) Int64Map(key string) (map[string]int64, error) {
	reply, err := redis.Int64Map(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// StringMap ...
func (r *Redis) StringMap(key string) (map[string]string, error) {
	reply, err := redis.StringMap(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

// Delete ...
func (r *Redis) Delete(key string) error {
	_, err := r.do("DEL", key)
	return err
}

// Exists ...
func (r *Redis) Exists(key string) (exists bool, err error) {
	exists, err = redis.Bool(r.do("EXISTS", key))
	return
}

// Sadd ...
func (r *Redis) Sadd(key string, vals interface{}, expire int) (err error) {
	args := []interface{}{key}

	if reflect.TypeOf(vals).Kind() == reflect.Slice {
		s := reflect.ValueOf(vals)
		for i := 0; i < s.Len(); i++ {
			ele := s.Index(i)
			args = append(args, ele.Interface())
		}
	} else {
		args = append(args, vals)
	}
	_, err = redis.Int64(r.do("SADD", args...))
	if err == nil {
		if _, err := r.do("EXPIRE", key, expire); err != nil {
			return err
		}
	}
	return
}

// SrandMemberInt64 ...
func (r *Redis) SrandMemberInt64(key string, count int) ([]int64, error) {
	vals, err := redis.Int64s(r.do("srandmember", key, count))
	return vals, err
}

// ZAdd ...
func (r *Redis) ZAdd(key string, score int, val interface{}) error {
	_, err := r.do("ZADD", key, score, val)
	return err
}

// ZrevRank ...
func (r *Redis) ZrevRank(key string, item string) (int32, error) {
	rank, err := redis.Int(r.do("zrevrank", key, item))
	return int32(rank), err
}

// Zrange ...
func (r *Redis) Zrange(key string, start int, end int) (map[string]int64, error) {
	ans, err := redis.Int64Map(r.do("ZRANGE", key, start, end, "withscores"))
	return ans, err
}

// ZScore ...
func (r *Redis) ZScore(key string, item string) (int, error) {
	score, err := redis.Int(r.do("ZSCORE", key, item))
	return score, err
}

// ZRem ...
func (r *Redis) ZRem(key string, item string) error {
	_, err := r.do("ZREM", key, item)
	return err
}

// ZRevRangeByScore ...
func (r *Redis) ZRevRangeByScore(key string, max int, min int) ([]interface{}, error) {
	reply, err := redis.Values(r.do("ZREVRANGEBYSCORE", key, max, min))
	return reply, err
}

// SisMember ...
func (r *Redis) SisMember(key string, value interface{}) (int, error) {
	reply, err := redis.Int(r.do("SISMEMBER", key, value))
	return reply, err
}
