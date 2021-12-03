package redlock

import (
	"context"
	"fmt"
	"time"

	"github.com/easy-go-123/red-lock/redlockdef"
	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/libeasygo/cuserror"
)

var lockReentrantScript = redis.NewScript(`
	if (redis.call('exists', KEYS[1]) == 0)
	then
		redis.call('hset', KEYS[1], ARGV[1], 1);
		redis.call('pexpire', KEYS[1], ARGV[2]);
		return 1;
	end;
	
	if (redis.call('hexists', KEYS[1], ARGV[1]) == 1)
	then
		redis.call('hincrby', KEYS[1], ARGV[1], 1);
		redis.call('pexpire', KEYS[1], ARGV[2]);
		return 2;
	end;
	
	return 0;
`)

var unlockReentrantScript = redis.NewScript(`
	if (redis.call('hexists', KEYS[1], ARGV[1]) == 1)
	then
		if(redis.call('hincrby', KEYS[1], ARGV[1], -1) <= 0)
		then
			return redis.call("del", KEYS[1])
		end
		return 2
	else
		return 0
	end
`)

type redisReentrantLock struct {
	redisCli   *redis.Client
	key        string
	token      string
	timeout    time.Duration
	fnUnlockOb func(key string)
}

func (lock *redisReentrantLock) tryLock() (ok, reentrant bool, err error) {
	sha1, err := lockReentrantScript.Load(context.TODO(), lock.redisCli).Result()
	if err != nil {
		return
	}

	f, err := lock.redisCli.EvalSha(lock.redisCli.Context(), sha1, []string{lock.key}, lock.token,
		int64(lock.timeout/time.Millisecond)).Int()

	if err != nil {
		return
	}

	if f == 0 {
		return
	}

	ok = true

	if f == 2 {
		reentrant = true
	}

	return
}

func (lock *redisReentrantLock) Unlock() (err error) {
	sha1, err := unlockReentrantScript.Load(context.TODO(), lock.redisCli).Result()
	if err != nil {
		return
	}

	f, err := lock.redisCli.EvalSha(lock.redisCli.Context(), sha1, []string{lock.key}, lock.token).Result()

	if err != nil {
		return
	}

	if f.(int64) != 1 && f.(int64) != 2 {
		err = cuserror.NewWithErrorMsg(fmt.Sprintf("%d", f))

		return
	}

	if f.(int64) == 1 {
		if lock.fnUnlockOb != nil {
			lock.fnUnlockOb(lock.key)
		}
	}

	return
}

func (lock *redisReentrantLock) RedisKey() string {
	return lock.key
}

func TryReentrantLockWithOwnedTimeout(redisCli *redis.Client, key, token string, ownedTimeout time.Duration) (lock redlockdef.Lock, reentrant bool, err error) {
	return tryReentrantLockWithOwnedTimeout(redisCli, key, token, ownedTimeout, nil)
}

func tryReentrantLockWithOwnedTimeout(redisCli *redis.Client, key, token string, timeout time.Duration,
	unlockOb func(key string)) (lock redlockdef.Lock, reentrant bool, err error) {
	lockImpl := &redisReentrantLock{
		redisCli:   redisCli,
		key:        key,
		token:      token,
		timeout:    timeout,
		fnUnlockOb: unlockOb,
	}

	ok, reentrant, err := lockImpl.tryLock()
	if err == nil && ok {
		lock = lockImpl
	}

	return
}
