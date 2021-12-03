package easyredlock

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/libeasygo/helper"
	"github.com/sgostarter/libeasygo/ut"
	"github.com/stretchr/testify/assert"
)

func testGetRedis(t *testing.T) *redis.Client {
	cfg := ut.SetupUTConfig4Redis(t)
	redisCli, err := helper.NewRedisClient(cfg.RedisDNS)
	assert.Nil(t, err)

	return redisCli
}

func TestTryLock(t *testing.T) {
	redisCli := testGetRedis(t)
	assert.NotNil(t, redisCli)

	err := TryLock(redisCli, "key1", func(key string, owned bool) error {
		assert.Equal(t, "key1", key)
		assert.Equal(t, true, owned)

		return nil
	})
	assert.Nil(t, err)

	cnt := 0
	err = TryLockWithTimeout(redisCli, "key2", 0, time.Second, func(key string, owned bool) error {
		assert.Equal(t, "key2", key)
		assert.Equal(t, true, owned)
		cnt++

		_ = TryLockWithTimeout(redisCli, "key2", 0, time.Second, func(key string, owned bool) error {
			assert.Equal(t, "key2", key)
			assert.Equal(t, false, owned)
			cnt++

			return nil
		})

		time.Sleep(2 * time.Second)

		_ = TryLockWithTimeout(redisCli, "key2", 0, time.Second, func(key string, owned bool) error {
			assert.Equal(t, "key2", key)
			assert.Equal(t, true, owned)
			cnt++

			return nil
		})

		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, cnt, 3)
}

func TestReentrantLock(t *testing.T) {
	redisCli := testGetRedis(t)
	assert.NotNil(t, redisCli)

	cnt := 0
	err := TryReentrantLock(redisCli, "key1", "token1", func(key, token string, reentrant, owned bool) error {
		assert.Equal(t, "key1", key)
		assert.Equal(t, "token1", token)
		assert.Equal(t, false, reentrant)
		assert.Equal(t, true, owned)
		cnt++

		return nil
	})
	assert.Nil(t, err)

	err = TryReentrantLockWithTimeout(redisCli, "key", "token", 0, time.Second, func(key, token string, reentrant, owned bool) error {
		assert.Equal(t, "key", key)
		assert.Equal(t, "token", token)
		assert.Equal(t, false, reentrant)
		assert.Equal(t, true, owned)
		cnt++

		err = TryReentrantLockWithTimeout(redisCli, "key", "token", 0, time.Second, func(key, token string, reentrant, owned bool) error {
			assert.Equal(t, "key", key)
			assert.Equal(t, "token", token)
			assert.Equal(t, true, reentrant)
			assert.Equal(t, true, owned)
			cnt++

			return nil
		})
		assert.Nil(t, err)

		err = TryReentrantLockWithTimeout(redisCli, "key", "token1", 0, time.Second, func(key, token string, reentrant, owned bool) error {
			assert.Equal(t, "key", key)
			assert.Equal(t, "token1", token)
			assert.Equal(t, false, reentrant)
			assert.Equal(t, false, owned)
			cnt++

			return nil
		})
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)

		err = TryReentrantLockWithTimeout(redisCli, "key", "token1", 0, time.Second, func(key, token string, reentrant, owned bool) error {
			assert.Equal(t, "key", key)
			assert.Equal(t, "token1", token)
			assert.Equal(t, reentrant, false)
			assert.Equal(t, true, owned)
			cnt++

			return nil
		})
		assert.Nil(t, err)

		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 5, cnt)
}
