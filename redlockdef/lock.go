package redlockdef

type Lock interface {
	Unlock() (err error)
	RedisKey() string
}
