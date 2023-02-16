# Bench Group
---

### 默认配置
当前 Bench 服务只支持一组配置
```conf
[bench-group-1]
BENCH_SRC_IP = localhost
BENCH_DEST_IP = localhost
BENCH_SRC_PORT = 9874
BENCH_CONFIG = wrk_http_long.json
```
### 参数说明
```conf
BENCH_SRC_IP：Bench服务ip，支持多个ip输入，用","或", "隔开
BENCH_DEST_IP：目标打压机器，可以是普通机器，也可以是target群组中的机器，目标机器只能有一个
BENCH_SRC_PORT：Bench服务的端口号
BENCH_CONFIG：Bench向目标机器打压的配置文件
```
常用的 BENCH_CONFIG 文件有以下几个
```conf
wrk_http_long.json：http 长链接
wrk_http_short.json：http 短链接
wrk_https_long.json：https 长链接
wrk_https_short.json：https 短链接
```