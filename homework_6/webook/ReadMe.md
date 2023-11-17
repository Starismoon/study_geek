此方案在触发限流或者连续N个请求时间大于M秒的时候将请求交给新开的goroutine处理。
### 适用场景
此方案适合在开发者想要控制响应速度并且有一定的重试机制确保验证码能够成功发送的场景下。
#### 优点
- 开发者可控制验证码发送响应时间
- 支持重试机制最大化保证验证码发送成功
- 可设置重试次数，不会导致goroutine占用过多内存

#### 缺点
- 可能自身网络问题导致误判为服务商响应过慢

### 改进方案
结合使用日志系统或其他方案判断其他请求自身网络问题是否正常

代码位置 internal/service/sms/failover/response_failover.go