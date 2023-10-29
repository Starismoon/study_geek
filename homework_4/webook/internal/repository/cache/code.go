package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/coocood/freecache"
	_ "github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"
	"strconv"
	"sync"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证太频繁")
)

type CodeRedisCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}
type RedisCodeCache struct {
	cmd redis.Cmdable
}
type LocalCodeCache struct {
	cmd  *freecache.Cache
	lock sync.RWMutex
}

func NewCodeCache(cache *freecache.Cache) CodeCache {
	return &LocalCodeCache{cmd: cache}
}

// Set NewCodeCache
//
//	func NewCodeCache(cmd redis.Cmdable) CodeRedisCache {
//		return &RedisCodeCache{
//			cmd: cmd,
//		}
//	}
func (c *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
	}()
	k := c.key(biz, phone)
	ttl, err := c.cmd.TTL([]byte(k))
	if err != nil {
		return errors.New("验证码存在，但是没有过期时间")
	}
	if ttl < 540 {
		err := c.cmd.Set([]byte(k), []byte(code), 600)
		if err != nil {
			return err
		}
		err = c.cmd.Set([]byte(k+":cnt"), []byte("3"), 600)
		if err != nil {
			return err
		}
	}
	return ErrCodeSendTooMany
}

func (c *LocalCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	c.lock.Lock()
	defer func() {
		c.lock.Unlock()
	}()
	k := c.key(biz, phone)
	cntByte, err := c.cmd.Get([]byte(k + ":cnt"))
	if err != nil {
		return false, nil
	}
	cntByteInt, err := strconv.Atoi(string(cntByte))
	if err != nil {
		return false, nil
	}
	if cntByteInt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	expectedCodeByte, err := c.cmd.Get([]byte(k))
	if string(expectedCodeByte) == code {
		_ = c.cmd.Set([]byte(k+":cnt"), []byte("0"), 0)
		return true, nil
	} else {
		_ = c.cmd.Set([]byte(k+":cnt"), []byte(strconv.Itoa(cntByteInt-1)), 0)
		return false, nil
	}
}
func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 redis 出了问题
		return err
	}
	switch res {
	case -2:
		return errors.New("验证码存在，但是没有过期时间")
	case -1:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 redis 出了问题
		return false, err
	}
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return true, nil
	}
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
func (c *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
