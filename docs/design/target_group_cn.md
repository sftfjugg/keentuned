# Target Group
---

### 默认配置
当前 Target 服务支持多组配置
```conf
[target-group-1]
TARGET_IP = localhost
TARGET_PORT = 9873
PARAMETER = sysctl.json
```
### 参数说明
```conf
TARGET_IP：Target服务ip，支持多个ip输入，用","或", "隔开
TARGET_PORT：Target服务的端口号
PARAMETER：调优参数文件，支持多个场景联合调优，用","或", "隔开
```
多组配置示例
```conf
[target-group-1]
TARGET_IP = localhost, ip1
TARGET_PORT = 9873
PARAMETER = nginx_conf.json
[target-group-2]
TARGET_IP = ip2, ip3
TARGET_PORT = 9873
PARAMETER = sysctl.json, nginx_conf.json
```
常用的 PARAMETER 文件有以下几个
```conf
wrk.json：wrk 发压参数调优
sysctl.json：内核参数调优
my_cnf.json：mysql 配置参数调优
nginx_conf.json：nginx 配置参数调优
```