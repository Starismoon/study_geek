package failover

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms"
	"gitee.com/geekbang/basic-go/webook/pkg/limiter"
	"sync/atomic"
	"time"
)

type ResponseFailoverSMSService struct {
	svc sms.Service
	// 响应时间，只读的，超过此值视为响应超时
	responseTime float64
	// 连续几个有问题
	cnt int32
	// 切换的阈值，只读的
	threshold int32
	limiter   limiter.Limiter
	key       string
	reTryNum  int
}

func NewResponseFailoverSMSService(svc sms.Service, threshold int32, responseTime float64, l limiter.Limiter, reTryNum int) *ResponseFailoverSMSService {
	return &ResponseFailoverSMSService{
		svc:          svc,
		threshold:    threshold,
		limiter:      l,
		responseTime: responseTime,
		reTryNum:     reTryNum,
	}
}

func (t *ResponseFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := t.limiter.Limit(ctx, t.key)
	if err != nil {
		return err
	}
	if limited {
		// 限流了，记录数据后其他goroutine重试
		go t.ReTrySend(ctx, tplId, args, numbers...)
		return errors.New("限流了，稍后发送验证码")
	}
	cnt := atomic.LoadInt32(&t.cnt)

	// 超过阈值，记录数据后其他goroutine重试
	if cnt >= t.threshold {
		return errors.New("服务商可能存在问题，响应较慢，稍后发送验证码")
	}
	// 判断响应时间，超过t.responseTime秒则为异常
	before := time.Now()
	err = t.svc.Send(ctx, tplId, args, numbers...)
	if time.Since(before).Seconds() < t.responseTime && err == nil {
		// 响应正常且响应时间小于t.responseTime 后重置为0
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	} else {
		// 如果响应时间超过t.responseTime秒或者遇到错误提示，则判断服务商可能存在问题加1
		// 只是响应慢 计数器加1，发送失败开启goroutine重试
		atomic.AddInt32(&t.cnt, 1)
		if err != nil {
			go t.ReTrySend(ctx, tplId, args, numbers...)
		}
		return errors.New("响应时间过长")
	}
	//return err
}
func (t *ResponseFailoverSMSService) ReTrySend(ctx context.Context, tplId string, args []string, numbers ...string) {
	// 5s后进行第一次重试，每次失败后等待时间翻倍
	sleepTime := 5
	for i := 0; i < t.reTryNum; i++ {
		time.Sleep(time.Second * time.Duration(sleepTime))
		before := time.Now()
		err := t.svc.Send(ctx, tplId, args, numbers...)
		if time.Since(before).Seconds() < t.responseTime && err == nil {
			// 响应正常后重置为0，并结束重试
			atomic.StoreInt32(&t.cnt, 0)
			return
		}
		sleepTime *= 2
	}
}
