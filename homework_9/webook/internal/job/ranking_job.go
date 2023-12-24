package job

import (
	"context"
	"errors"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"math/rand"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	cache   cache.LoadCache
	l       logger.LoggerV1
	timeout time.Duration
	client  *rlock.Client
	key     string

	localLock *sync.Mutex
	lock      *rlock.Lock

	// 作业提示
	// 随机生成一个，就代表当前负载。你可以每隔一分钟生成一个
	load int
	// 模拟实例id
	instanceId string
}

func NewRankingJob(
	svc service.RankingService,
	cache cache.LoadCache,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration) *RankingJob {
	var rankingJob *RankingJob
	rankingJob = &RankingJob{svc: svc,
		cache:      cache,
		key:        "job:ranking",
		l:          l,
		client:     client,
		localLock:  &sync.Mutex{},
		timeout:    timeout,
		instanceId: string(rand.Int31n(1000)),
	}
	// 模拟每分钟获取一次负载
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for range t.C {
			rankingJob.load = rand.Intn(100)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, instanceId, err := rankingJob.cache.Get(ctx)
			if err != nil {
				rankingJob.l.Error("获取负载失败")
				_ = rankingJob.cache.Set(ctx, fmt.Sprintf("%s_%d", rankingJob.instanceId, rankingJob.load))
			} else {
				// 如果redis中记录的最小负载是本机的负载，更新当前负载到redis
				if instanceId == rankingJob.instanceId {
					err := rankingJob.cache.Set(ctx, fmt.Sprintf("%s_%d", instanceId, rankingJob.load))
					if err != nil {
						//return
						rankingJob.l.Error("负载设置失败")
					}
				}
			}

			cancel()
		}
	}()
	return rankingJob
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// go fun() { r.Run()}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {

		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		load, instanceId, err := r.cache.Get(ctx)
		if err != nil {
			_ = r.cache.Set(ctx, fmt.Sprintf("%s_%d", r.instanceId, r.load))
			return err
		}
		if r.load < load || instanceId == r.instanceId {
			// 当前实例负载比redis记录的更小或记录最小负载正好是当前实例
			err := r.cache.Set(ctx, fmt.Sprintf("%s_%d", r.instanceId, r.load))
			if err != nil {
				//return
				r.l.Error("负载设置失败")
			}
		} else {
			return errors.New("当前服务器负载不是最小")
		}
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				//lock.Unlock()
				r.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

//func (r *RankingJob) Run() error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//	lock, err := r.client.Lock(ctx, r.key, r.timeout,
//		&rlock.FixIntervalRetry{
//			Interval: time.Millisecond * 100,
//			Max:      3,
//			// 重试的超时
//		}, time.Second)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		// 解锁
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		er := lock.Unlock(ctx)
//		if er != nil {
//			r.l.Error("ranking job释放分布式锁失败", logger.Error(er))
//		}
//	}()
//	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
//	defer cancel()
//
//	return r.svc.TopN(ctx)
//}
