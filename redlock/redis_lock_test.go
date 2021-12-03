package redlock

import (
	"context"
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

func TestBase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	rlm := NewRedisLockManager(ctx, testGetRedis(t), nil)
	lock, err := rlm.TryLockWithOwnedTimeout("lock-1", 2*time.Second, false)
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	err = lock.Unlock()
	assert.Nil(t, err)

	rlm.Terminal()
	rlm.Wait()
}

func TestReentrantBase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	rlm := NewRedisLockManager(ctx, testGetRedis(t), nil)
	lock, _, err := rlm.TryReentrantLockWithOwnedTimeout("lock-1", "gid1", 2*time.Second, false)
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	lock2, _, err := rlm.TryReentrantLockWithOwnedTimeout("lock-1", "gid1", 2*time.Second, false)
	assert.Nil(t, err)
	assert.NotNil(t, lock2)

	err = lock.Unlock()
	assert.Nil(t, err)

	err = lock2.Unlock()
	assert.Nil(t, err)

	rlm.Terminal()
	rlm.Wait()
}

func TestTTL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	rlm := NewRedisLockManager(ctx, testGetRedis(t), nil)
	lock, err := rlm.TryLockWithOwnedTimeout("lock-1", 2*time.Second, false)
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	time.Sleep(time.Second * 10)

	err = lock.Unlock()
	assert.Nil(t, err)

	rlm.Terminal()
	rlm.Wait()
}
