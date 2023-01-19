# KeenTune-daemon(Keentuned)
## INTRODUCTION
KeenTune 是一款 AI 算法与专家知识库双轮驱动的操作系统全栈式智能优化产品，为主流的操作系统提供轻量化、跨平台的一键式性能调优，让应用在智能定制的运行环境发挥最优性能。

KeenTuned 是 KeenTune 的调度管理组件，包含 CLI 和 Daemon 两个部分。CLI 模块提供用户可用的命令行接口，命令分为基础命令、静态调优相关命令、动态调优相关命令三个部分。keentuned 作为核心管控模块，负责监控其他组件、接收解析来自 CLI 的命令、按照业务处理顺序调度相关组件等。

## Build & Install
First, we can use keentuned either build 'keentuned' by source code or install by yum repo. Choose one of the following ways.

### Build by source code
```s
>> sh keentuned_install.sh
``` 
### Install by yum install
First add the yum repo. If it is an Ali8 series system, directly by modifying `enabled=1` in /etc/yum.repos.d/AnolisOS-Plus.repo to enable Plus source.
```s
[KeenTune]
baseurl=https://mirrors.openanolis.cn/anolis/8.6/Plus/$basearch/os
enabled=1
gpgkey=https://mirrors.openanolis.cn/anolis/RPM-GPG-KEY-ANOLIS
gpgcheck=0
```
And then install
```s
yum clean all
yum makecache
yum install keentuned -y
```

## Configuration
we can find configuration file in /etc/keentune/conf/keentund.conf
```conf
[keentuned]
# Basic configuration of KeenTune-Daemon(KeenTuned).
VERSION_NUM     = 2.0.1                     ; Record the version number of keentune
PORT            = 9871                      ; KeenTuned access port
HEARTBEAT_TIME  = 30                        ; Heartbeat detection interval(unit: seconds), recommended value 30
KEENTUNED_HOME  = /etc/keentune             ; KeenTuned  default configuration root location
DUMP_HOME       = /var/keentune             ; Dump home is the working directory for KeenTune job execution result

; configuration about configuration dumping
DUMP_BASELINE_CONFIGURATION = false         ; If dump the baseline configuration.
DUMP_TUNING_CONFIGURATION   = false         ; If dump the intermediate configuration.
DUMP_BEST_CONFIGURATION     = true          ; If dump the best configuration.

; benchmark replay duplicately round
BASELINE_BENCH_ROUND    = 5                 ; Benchmark execution rounds of baseline
TUNING_BENCH_ROUND      = 3                 ; Benchmark execution rounds during tuning execution
RECHECK_BENCH_ROUND     = 4                 ; Benchmark execution rounds after tuning for recheck

; configuration about log
LOGFILE_LEVEL           = DEBUG             ; logfile log level, i.e. INFO, DEBUG, WARN, FATAL
LOGFILE_NAME            = keentuned.log     ; logfile name.
LOGFILE_INTERVAL        = 2                 ; logfile interval
LOGFILE_BACKUP_COUNT    = 14                ; logfile backup count

[brain]
# Topology of brain and basic configuration about brain.
BRAIN_IP                = localhost         ; The machine ip address to depoly keentune-brain.
BRAIN_PORT              = 9872              ; The service port of keentune-brain.
AUTO_TUNING_ALGORITHM   = tpe               ; Brain optimization algorithm. i.e. tpe, hord, random
SENSITIZE_ALGORITHM     = shap              ; Explainer of sensitive parameter training. i.e. shap, lasso, univariate

[target-group-1]
# Topology of target group and knobs to be tuned in target.
TARGET_IP   = localhost                     ; The machine ip address to depoly keentune-target.
TARGET_PORT = 9873                          ; The service port of keentune-target.
PARAMETER   = sysctl.json                   ; Knobs to be tuned in this target

[bench-group-1]
# Topology of bench group and benchmark script to be performed in bench.
BENCH_SRC_IP    = localhost                 ; The machine ip address to depoly keentune-bench.
BENCH_SRC_PORT  = 9874                      ; The service port of keentune-bench.
BENCH_DEST_IP   = localhost                 ; The destination ip address in benchmark workload.
BENCH_CONFIG    = bench_wrk_nginx_long.json ; The configuration file of benchmark to be performed

```

## Run
After modify KeenTuned configuration file, we can deploy KeenTuned and listening to requests as
```s
>> keentuned
or depoly keentuned by systemctl
>> systemctl start keentuned
```

## Code structure
### CLI
```
cli
├── api.go
├── checkinput.go
├── common.go
├── config.go
├── main.go
├── param.go
├── profile.go
└── sensitize.go

0 directories, 8 files
```
### Daemon
```
daemon
├── api                 # Control layer: receiving cli and ui requests, checking parameters, and simple business processing
│   ├── common
│   │   ├── common.go
│   │   ├── handle.go
│   │   ├── heartbeat.go
│   │   └── read.go
│   ├── param
│   │   ├── delete.go
│   │   ├── dump.go
│   │   ├── job.go
│   │   ├── list.go
│   │   ├── rollback.go
│   │   ├── stop.go
│   │   └── tune.go
│   ├── profile
│   │   ├── delete.go
│   │   ├── generate.go
│   │   ├── info.go
│   │   ├── list.go
│   │   ├── rollback.go
│   │   └── set.go
│   ├── sensitize
│   │   ├── delete.go
│   │   ├── job.go
│   │   ├── stop.go
│   │   └── train.go
│   └── system
│       └── benchmark.go
├── common              # Common libraries layer: including configuration, file, log processing and tool functions.
│   ├── config
│   │   ├── check.go
│   │   ├── config.go
│   │   ├── priority.go
│   │   ├── update.go
│   │   └── workpath.go
│   ├── file
│   │   ├── file.go
│   │   └── gocsv.go
│   ├── log
│   │   └── log.go
│   └── utils
│       ├── calculator.go
│       ├── http
│       │   └── http.go
│       ├── parsejson.go
│       └── utils.go
├── examples
│   ├── benchmark       # sample benchmark script and json configuration lib
│   ├── detect          # sample configuration for Resource Detection
│   ├── parameter       # sample parameter lib 
│   └── profile         # sample profile lib
└── modules             # Model layer: tuning and other complex business processing
│   ├── analyse.go
│   ├── apply.go
│   ├── assemble.go
│   ├── benchmark.go
│   ├── best.go
│   ├── common.go
│   ├── configuration.go
│   ├── init.go
│   ├── jobs.go
│   ├── loop.go
│   ├── parameter.go
│   ├── setter.go
│   ├── stop.go
│   ├── trainer.go
│   └── tuner.go
└── daemon.go           # keentuned daemon main entrance

24 directories, 75 files
```
