<div align="center">
  <img src="https://github.com/easy-go-123/red-lock/workflows/ut/badge.svg?branch=main&event=push" alt="Unit Test">
  <img src="https://github.com/easy-go-123/red-lock/workflows/golangci-lint/badge.svg?branch=main&event=push" alt="GolangCI Linter">
</div>

# red-lock

---

基于redis实现的分布式锁

## 应用场景

能访问一个`redis`(包括集群)的多个应用之间加锁同步数据

### 非递归锁

* `easyredlock.TryLock`
* `easyredlock.TryLockWithTimeout` - 支持尝试加锁和自定义锁的过期时间

### 递归锁

* `easyredlock.TryReentrantLock`
* `easyredlock.TryReentrantLockWithTimeout` - 支持尝试加锁和自定义锁的过期时间

### 自动续期锁

* `easyredlock.Manager`


## 用例

### 非递归锁

```
err := TryLock(redisCli, "key1", func(key string, owned bool) error {
    assert.Equal(t, "key1", key)
    assert.Equal(t, true, owned)

    return nil
})
assert.Nil(t, err)
```

### 递归锁

```
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
	}
```

## 自动续期锁

```bash
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
```