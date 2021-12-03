package redlock

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/easy-go-123/red-lock/redlockdef"
	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/i/logger"
)

const (
	channelCacheSize       = 10
	keyTTLIntervalDuration = time.Millisecond * 500
)

type lockKeyInfo struct {
	key     string
	timeout time.Duration
}

type RedisLockManager struct {
	wg        sync.WaitGroup
	ctx       context.Context
	ctxCancel context.CancelFunc
	redisCli  *redis.Client
	log       logger.Wrapper
	hostName  string

	addLockChannel    chan *lockKeyInfo
	removeLockChannel chan string
	lockKeys          map[string]*lockKeyInfo
}

func NewRedisLockManager(ctx context.Context, redisCli *redis.Client, log logger.Wrapper) *RedisLockManager {
	if log == nil {
		log = logger.NewWrapper(&logger.NopLogger{})
	}

	hostName, _ := os.Hostname()

	ctx, cancel := context.WithCancel(ctx)
	rlm := &RedisLockManager{
		ctx:               ctx,
		ctxCancel:         cancel,
		redisCli:          redisCli,
		log:               log.WithFields(logger.FieldString("clsRedisLockManager", "1")),
		hostName:          hostName,
		addLockChannel:    make(chan *lockKeyInfo, channelCacheSize),
		removeLockChannel: make(chan string, channelCacheSize),
		lockKeys:          make(map[string]*lockKeyInfo),
	}

	rlm.wg.Add(1)

	go rlm.keyWatchRoutine()

	return rlm
}

func (rlm *RedisLockManager) Wait() {
	rlm.wg.Wait()
}

func (rlm *RedisLockManager) Terminal() {
	rlm.ctxCancel()
}

func (rlm *RedisLockManager) keyWatchRoutine() {
	defer rlm.wg.Done()

	rlm.log.Infof("enterKeyWatchRoutine")

	loop := true

	d := time.Hour

	for loop {
		select {
		case <-rlm.ctx.Done():
			rlm.log.Debug("ctxDone")

			loop = false

			break
		case <-time.After(d):
			for _, ki := range rlm.lockKeys {
				_ = rlm.redisCli.Expire(context.TODO(), ki.key, ki.timeout).Err()
			}
		case ki := <-rlm.addLockChannel:
			rlm.lockKeys[ki.key] = ki
			d = keyTTLIntervalDuration

			rlm.log.WithFields(logger.FieldString("key", ki.key)).Debug("newLockKey")
		case key := <-rlm.removeLockChannel:
			_, ok := rlm.lockKeys[key]
			if !ok {
				rlm.log.WithFields(logger.FieldString("key", key)).Error("removeNotExistsLockKey")

				return
			}

			rlm.log.WithFields(logger.FieldString("key", key)).Error("removeLockKey")

			delete(rlm.lockKeys, key)

			if len(rlm.lockKeys) == 0 {
				d = time.Hour
			}
		}
	}

	rlm.log.Info("leaveKeyWatchRoutine")
}

func (rlm *RedisLockManager) TryLockWithOwnedTimeout(key string, timeout time.Duration, autoRenewal bool) (lock redlockdef.Lock, err error) {
	lock, err = tryLockWithOwnedTimeout(rlm.redisCli, key, timeout, func(key string) {
		rlm.removeLockChannel <- key
	})
	if err != nil {
		return
	}

	if lock != nil && autoRenewal {
		rlm.addLockChannel <- &lockKeyInfo{
			key:     lock.RedisKey(),
			timeout: timeout,
		}
	}

	return
}

func (rlm *RedisLockManager) TryReentrantLockWithOwnedTimeout(key, reentrantToken string, timeout time.Duration,
	autoRenewal bool) (lock redlockdef.Lock, reentrant bool, err error) {
	lockToken := fmt.Sprintf("%s:%d:%s", rlm.hostName, os.Getpid(), reentrantToken)

	lock, reentrant, err = tryReentrantLockWithOwnedTimeout(rlm.redisCli, key, lockToken, timeout, func(key string) {
		rlm.removeLockChannel <- key
	})
	if err != nil {
		return
	}

	if lock != nil && autoRenewal {
		rlm.addLockChannel <- &lockKeyInfo{
			key:     lock.RedisKey(),
			timeout: timeout,
		}
	}

	return
}
