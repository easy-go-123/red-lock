package easyredlock

import (
	"time"

	"github.com/easy-go-123/red-lock/redlock"
	"github.com/easy-go-123/red-lock/redlockdef"
	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/libeasygo/helper"
)

// DefaultOwnedTimeout is the duration for which the lock is valid
const DefaultOwnedTimeout = 2 * time.Second

type LockResultCB func(key string, owned bool) error
type ReentrantLockResultCB func(key, token string, reentrant, owned bool) error

func TryLock(redisCli *redis.Client, key string, cb LockResultCB) (err error) {
	return TryLockWithTimeout(redisCli, key, 0, DefaultOwnedTimeout, cb)
}

func TryLockWithTimeout(redisCli *redis.Client, key string, tryTimeout, ownedTimeout time.Duration, cb LockResultCB) (err error) {
	if cb == nil {
		err = ErrNoResultCB

		return
	}

	var lock redlockdef.Lock

	_, _ = helper.TryWithTimeout(tryTimeout, func(_ time.Duration) bool {
		lock, err = redlock.TryLockWithOwnedTimeout(redisCli, key, ownedTimeout)
		if err != nil {
			return true
		}

		return lock != nil
	})

	if err != nil {
		return
	}

	err = cb(key, lock != nil)

	if lock != nil {
		_ = lock.Unlock()
	}

	return
}

func TryReentrantLock(redisCli *redis.Client, key, token string, cb ReentrantLockResultCB) (err error) {
	return TryReentrantLockWithTimeout(redisCli, key, token, 0, DefaultOwnedTimeout, cb)
}

func TryReentrantLockWithTimeout(redisCli *redis.Client, key, token string, tryTimeout, ownedTimeout time.Duration, cb ReentrantLockResultCB) (err error) {
	if cb == nil {
		err = ErrNoResultCB

		return
	}

	var lock redlockdef.Lock

	var reentrant bool

	_, _ = helper.TryWithTimeout(tryTimeout, func(_ time.Duration) bool {
		lock, reentrant, err = redlock.TryReentrantLockWithOwnedTimeout(redisCli, key, token, ownedTimeout)
		if err != nil {
			return true
		}

		return lock != nil
	})

	if err != nil {
		return
	}

	err = cb(key, token, reentrant, lock != nil)

	if lock != nil {
		_ = lock.Unlock()
	}

	return
}
