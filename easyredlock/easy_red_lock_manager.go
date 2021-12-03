package easyredlock

import (
	"context"
	"time"

	"github.com/easy-go-123/red-lock/redlock"
	"github.com/easy-go-123/red-lock/redlockdef"
	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/i/logger"
	"github.com/sgostarter/libeasygo/cuserror"
	"github.com/sgostarter/libeasygo/helper"
)

type Manager struct {
	lockMan *redlock.RedisLockManager
}

func NewManager(ctx context.Context, redisCli *redis.Client, log logger.Wrapper) *Manager {
	return &Manager{
		lockMan: redlock.NewRedisLockManager(ctx, redisCli, log),
	}
}

func (man *Manager) TryLock(key string, cb LockResultCB) (err error) {
	return man.TryLockWithTimeout(key, 0, DefaultOwnedTimeout, cb)
}

func (man *Manager) TryLockWithTimeout(key string, tryTimeout, ownedTimeout time.Duration, cb LockResultCB) (err error) {
	if cb == nil {
		err = cuserror.NewWithErrorMsg("noCB")

		return
	}

	var lock redlockdef.Lock

	_, _ = helper.TryWithTimeout(tryTimeout, func(_ time.Duration) bool {
		lock, err = man.lockMan.TryLockWithOwnedTimeout(key, ownedTimeout, true)
		if err != nil {
			return true
		}

		return lock != nil
	})

	if err != nil {
		return
	}

	err = cb(key, lock != nil)

	return
}

func (man *Manager) TryReentrantLock(key, token string, cb ReentrantLockResultCB) (err error) {
	return man.TryReentrantLockWithTimeout(key, token, 0, DefaultOwnedTimeout, cb)
}

func (man *Manager) TryReentrantLockWithTimeout(key, token string, tryTimeout, ownedTimeout time.Duration, cb ReentrantLockResultCB) (err error) {
	if cb == nil {
		err = ErrNoResultCB

		return
	}

	var lock redlockdef.Lock

	var reentrant bool

	_, _ = helper.TryWithTimeout(tryTimeout, func(_ time.Duration) bool {
		lock, reentrant, err = man.lockMan.TryReentrantLockWithOwnedTimeout(key, token, ownedTimeout, true)
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
