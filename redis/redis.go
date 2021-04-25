package redis

import (
	"reflect"

	"github.com/gomodule/redigo/redis"
)

/**
 * 设置缓存
 */
func (r *Redis) Set(key string, value interface{}, expire int) error {
	_, err := r.do("Set", key, value, "EX", expire)
	return err
}

/**
 * key不存在时设置缓存
 */
func (r *Redis) SetNX(key string, value interface{}, expire int) (bool, error) {
	reply, err := r.do("Set", key, value, "EX", expire, "NX")
	//成功返回OK，失败返回nil
	return reply != nil, err
}

/**
 * 计数器加
 */
func (r *Redis) Incr(key string, value interface{}, expire int) (int, error) {
	reply, err := redis.Int(r.do("INCR", key))
	if err != nil {
		return 0, err
	}
	if reply == 1 {
		//初次设置的时候，设置有效期
		if _, err := r.do("EXPIRE", key, expire); err != nil {
			return 0, err
		}
	}
	return reply, nil
}

/**
 * 计数器减
 */
func (r *Redis) Decr(key string) (int, error) {
	reply, err := redis.Int(r.do("DECR", key))
	if err != nil {
		return 0, err
	}
	return reply, nil
}

/**
 * 设置过期时间
 */
func (r *Redis) Expire(key string, expire int) error {
	_, err := r.do("EXPIRE", key, expire)
	return err
}

/**
 * 获取数据
 */
func (r *Redis) Get(key string) (interface{}, error) {
	reply, err := r.do("Get", key)
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取整数
 */
func (r *Redis) Int(key string) (int, error) {
	reply, err := redis.Int(r.Get(key))
	if err == redis.ErrNil {
		return 0, nil
	}
	return reply, err
}

/**
 * 获取整数
 */
func (r *Redis) Int64(key string) (int64, error) {
	reply, err := redis.Int64(r.Get(key))
	if err == redis.ErrNil {
		return 0, nil
	}
	return reply, err
}

/**
 * 获取浮点数
 */
func (r *Redis) Float64(key string) (float64, error) {
	reply, err := redis.Float64(r.Get(key))
	if err == redis.ErrNil {
		return 0, nil
	}
	return reply, err
}

/**
 * 获取字符串
 */
func (r *Redis) String(key string) (string, error) {
	reply, err := redis.String(r.Get(key))
	if err == redis.ErrNil {
		return "", nil
	}
	return reply, err
}

/**
 * 获取字节切片
 */
func (r *Redis) Bytes(key string) ([]byte, error) {
	reply, err := redis.Bytes(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取整数切片
 */
func (r *Redis) Ints(key string) ([]int, error) {
	reply, err := redis.Ints(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取整数切片
 */
func (r *Redis) Int64s(key string) ([]int64, error) {
	reply, err := redis.Int64s(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取浮点数切片
 */
func (r *Redis) Float64s(key string) ([]float64, error) {
	reply, err := redis.Float64s(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取字符串切片
 */
func (r *Redis) Strings(key string) ([]string, error) {
	reply, err := redis.Strings(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取字节切片的切片
 */
func (r *Redis) ByteSlices(key string) ([][]byte, error) {
	reply, err := redis.ByteSlices(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取整数map
 */
func (r *Redis) IntMap(key string) (map[string]int, error) {
	reply, err := redis.IntMap(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取整数map
 */
func (r *Redis) Int64Map(key string) (map[string]int64, error) {
	reply, err := redis.Int64Map(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 获取字符串map
 */
func (r *Redis) StringMap(key string) (map[string]string, error) {
	reply, err := redis.StringMap(r.Get(key))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, err
}

/**
 * 删除缓存
 */
func (r *Redis) Delete(key string) error {
	_, err := r.do("DEL", key)
	return err
}

func (r *Redis) Exists(key string) (exists bool, err error) {
	exists, err = redis.Bool(r.do("EXISTS", key))
	return
}

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
		//初次设置的时候，设置有效期
		if _, err := r.do("EXPIRE", key, expire); err != nil {
			return err
		}
	}
	return
}

func (r *Redis) SrandMemberInt64(key string, count int) ([]int64, error) {
	vals, err := redis.Int64s(r.do("srandmember", key, count))
	return vals, err
}

func (r *Redis) ZAdd(key string, score int, val interface{}) error {
	_, err := r.do("ZADD", key, score, val)
	return err
}

func (r *Redis) ZrevRank(key string, item string) (int32, error) {
	rank, err := redis.Int(r.do("zrevrank", key, item))
	return int32(rank), err
}

func (r *Redis) Zrange(key string, start int, end int) (map[string]int64, error) {
	ans, err := redis.Int64Map(r.do("ZRANGE", key, start, end, "withscores"))
	return ans, err
}

func (r *Redis) ZScore(key string, item string) (int, error) {
	score, err := redis.Int(r.do("ZSCORE", key, item))
	return score, err
}

func (r *Redis) ZRem(key string, item string) error {
	_, err := r.do("ZREM", key, item)
	return err
}

/**
 * ZREVRANGEBYSCORE用例：
	for _,v := range list {
		val := string(v.([]byte)))
        ...
	}
*/
func (r *Redis) ZRevRangeByScore(key string, max int, min int) ([]interface{}, error) {
	reply, err := redis.Values(r.do("ZREVRANGEBYSCORE", key, max, min))
	return reply, err
}

func (r *Redis) SisMember(key string, value interface{}) (int, error) {
	reply, err := redis.Int(r.do("SISMEMBER", key, value))
	return reply, err
}
