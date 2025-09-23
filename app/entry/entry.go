package entry

import "time"

type Entry interface {
	Type() string
}

type RedisString struct {
	value      string
	expiryTime time.Time
}

func NewRedisString(v string, et time.Time) *RedisString {
	return &RedisString{
		value:      v,
		expiryTime: et,
	}
}

func (r *RedisString) Value() string {
	return r.value
}

func (r *RedisString) ExpiryTime() time.Time {
	return r.expiryTime
}

func (r *RedisString) Type() string {
	return "string"
}
