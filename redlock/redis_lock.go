package redlock

import (
	"context"
	"fmt"
	"time"

	"github.com/easy-go-123/red-lock/redlockdef"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sgostarter/libeasygo/cuserror"
)

type redisLock struct {
	redisCli   *redis.Client
	key        string
	token      string
	timeout    time.Duration
	fnUnlockOb func(key string)
}

var unlockScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1]
	then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

func (lock *redisLock) tryLock() (bool, error) {
	return lock.redisCli.SetNX(context.TODO(), lock.key, lock.token, lock.timeout).Result()
}

func (lock *redisLock) Unlock() (err error) {
	sha1, err := unlockScript.Load(context.TODO(), lock.redisCli).Result()
	if err != nil {
		return
	}

	f, err := lock.redisCli.EvalSha(lock.redisCli.Context(), sha1, []string{lock.key}, lock.token).Result()

	if err != nil {
		return
	}

	if lock.fnUnlockOb != nil {
		lock.fnUnlockOb(lock.key)
	}

	if f.(int64) != 1 {
		err = cuserror.NewWithErrorMsg(fmt.Sprintf("%d", f))

		return
	}

	return
}

func (lock *redisLock) RedisKey() string {
	return lock.key
}

func TryLockWithOwnedTimeout(redisCli *redis.Client, key string, ownedTimeout time.Duration) (lock redlockdef.Lock, err error) {
	return tryLockWithOwnedTimeout(redisCli, key, ownedTimeout, nil)
}

func tryLockWithOwnedTimeout(redisCli *redis.Client, key string, ownedTimeout time.Duration, unlockOb func(key string)) (lock redlockdef.Lock, err error) {
	lockImpl := &redisLock{
		redisCli:   redisCli,
		key:        key,
		token:      uuid.New().String(),
		timeout:    ownedTimeout,
		fnUnlockOb: unlockOb,
	}

	ok, err := lockImpl.tryLock()
	if err == nil && ok {
		lock = lockImpl
	}

	return
}
