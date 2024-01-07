在初始化NewRankingJob的时候生成一个随机数作为当前实例的id，启动一个goroutine每分钟更新一次负载，并判断redis中记录的如果是当前实例的负载直接更新redis

在执行job之前判断当前实例负载和redis中记录的最小负载，如果当前负载更小或者redis记录的也是当前实例的负载情况则更新redis，当前实例执行job，否则直接返回让其他实例执行

### 有没有可能选中最差的节点
答：有可能在redis中没有记录最小值的情况下，以当前节点执行job选中最差的节点
### 如果选中的节点宕机了，会发生什么
答： 如果宕机前此节点的负载值很低，其他节点负载一直高于此值，在redis记录过期前，则一直不会有节点执行job，但redis值过期后会有新的负载值更新记录去执行


代码实现路径 webook/internal/job/ranking_job.go