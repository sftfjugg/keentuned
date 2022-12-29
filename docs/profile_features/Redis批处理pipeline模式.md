# 优化原理
- pipeline：client一个请求，redis server一个响应，期间client阻塞

![](https://intranetproxy.alipay.com/skylark/lark/0/2022/png/337836/1661499928149-01e754ff-9592-4f7b-98b2-8973b02757da.png#clientId=u17a851c2-92c9-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=231&id=u11098e73&margin=%5Bobject%20Object%5D&originHeight=597&originWidth=248&originalType=url&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u9aaf7a2c-90a4-4a00-98f6-41ef11824b3&title=&width=96)

- Pipeline：redis的管道命令，允许client将多个请求依次发给[服务器](https://cloud.tencent.com/product/cvm?from=10680)（redis的客户端，如jedisCluster，lettuce等都实现了对pipeline的封装），过程中而不需要等待请求的回复，在最后再一并读取结果即可。

![](https://intranetproxy.alipay.com/skylark/lark/0/2022/png/337836/1661499928614-ed12aa3c-d1a3-4058-a4de-3dec47e25243.png#clientId=u17a851c2-92c9-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=181&id=u1d7baedd&margin=%5Bobject%20Object%5D&originHeight=276&originWidth=219&originalType=url&ratio=1&rotation=0&showTitle=false&status=done&style=none&taskId=u4828cfab-876d-4419-aec4-36d9aad6b4a&title=&width=144)
在实际业务使用中，Pipeline的模式更为常见，并且倚天因为其自身高频及多core的能力，在Pipeline模式下的性能优势更为明显。
# 使用方法
单pipeline的测试命令示例：
```
memtier_benchmark  -s 10.0.3.112 -p 9400 -t 8 --test-time=60  -c 10 --ratio=1:0 --pipeline=1 -d 32 --key-maximum=100000
```
```
memtier_benchmark  -s 10.0.3.112 -p 9400 -t 8 --test-time=60  -c 10 --ratio=1:0 -d 32 --pipeline=50 -key-maximum=10000000

```
