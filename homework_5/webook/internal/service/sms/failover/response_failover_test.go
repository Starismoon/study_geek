package failover

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms"
	smsmocks "gitee.com/geekbang/basic-go/webook/internal/service/sms/mocks"
	"gitee.com/geekbang/basic-go/webook/pkg/limiter"
	limitermocks "gitee.com/geekbang/basic-go/webook/pkg/limiter/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestResonseFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter)

		responseTime float64
		// 连续几个有问题
		cnt int32
		// 切换的阈值，只读的
		threshold int32
		// 预期输出
		wantErr error
	}{
		{
			name: "超过设定时长",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Do(func(_ context.Context, _ string, _ []string, _ string) {
					time.Sleep(time.Second * 6)
				}).Return(nil)
				return svc, l
			},
			cnt:       1,
			threshold: 10,
			wantErr:   errors.New("响应时间过长"),
		},
		{
			name: "被限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, l
			},
			wantErr: errors.New("限流了，稍后发送验证码"),
		},
		{
			name: "限流器错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("redis限流器错误"))
				return svc, l
			},
			wantErr: errors.New("redis限流器错误"),
		},
		{
			name: "超过阈值",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				return svc, l
			},
			cnt:       10,
			threshold: 10,
			wantErr:   errors.New("服务商可能存在问题，响应较慢，稍后发送验证码"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			smsSvc, l := tc.mock(ctrl)
			svc := NewResponseFailoverSMSService(smsSvc, tc.threshold, 5, l, 3)
			svc.cnt = tc.cnt
			err := svc.Send(context.Background(), "abc",
				[]string{"123"}, "123456")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
