package easyredlock

import (
	"context"
	"testing"
	"time"

	"github.com/sgostarter/i/logger"
	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	ctx := context.Background()

	redisCli := testGetRedis(t)
	assert.NotNil(t, redisCli)

	lockMan := NewManager(ctx, redisCli, logger.NewWrapper(logger.NewCommLogger(&logger.FmtRecorder{})))

	err := lockMan.TryLock("key", func(key string, owned bool) error {
		assert.Equal(t, "key", key)
		assert.Equal(t, true, owned)

		time.Sleep(5 * time.Second)

		err := lockMan.TryLock("key", func(key string, owned bool) error {
			assert.Equal(t, "key", key)
			assert.Equal(t, false, owned)

			return nil
		})
		assert.Nil(t, err)

		return nil
	})
	assert.Nil(t, err)
}

func Test2(t *testing.T) {
	ctx := context.Background()

	redisCli := testGetRedis(t)
	assert.NotNil(t, redisCli)

	lockMan := NewManager(ctx, redisCli, logger.NewWrapper(logger.NewCommLogger(&logger.FmtRecorder{})))

	startTime := time.Now()
	err := lockMan.TryReentrantLockWithTimeout("key1", "token1", 0, 1*time.Second,
		func(key, token string, reentrant, owned bool) error {
			assert.Equal(t, "key1", key)
			assert.Equal(t, "token1", token)
			assert.Equal(t, false, reentrant)
			assert.Equal(t, true, owned)

			go func() {
				_ = lockMan.TryReentrantLockWithTimeout("key1", "token2", 3*time.Second, time.Second, func(key, token string, reentrant, owned bool) error {
					assert.Equal(t, "key1", key)
					assert.Equal(t, "token2", token)
					assert.Equal(t, false, reentrant)
					assert.Equal(t, false, owned)

					tu := time.Since(startTime)
					assert.True(t, tu >= 3*time.Second)

					return nil
				})
			}()

			time.Sleep(4 * time.Second)

			return nil
		})
	assert.Nil(t, err)
}
