# Configuration of KeenTune
---  
## KeenTuned
<https://gitee.com/anolis/keentuned/blob/master/keentuned.conf>  
**Installed in /etc/keentune/conf/keentuned.conf**  
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
---  
## KeenTune-brain
<https://gitee.com/anolis/keentune_brain/blob/master/brain/brain.conf>  
**Installed in /etc/keentune/conf/brain.conf**  
```conf
[brain]
# Basic Configuration
KeenTune_HOME       = /etc/keentune/    ; KeenTune-brain install path.
KeenTune_WORKSPACE  = /var/keentune/    ; KeenTune-brain workspace.
BRAIN_PORT          = 9872              ; KeenTune-brain service port

[tuning]
# Auto-tuning Algorithm Configuration.
MAX_SEARCH_SPACE    = 1000              ; Limitation of the Max-number of available value of a single knob to avoid dimension explosion.
SURROGATE           = RBFInterpolant    ; Surrogate in tuning algorithm - HORD i.e. RBFInterpolant, PolyRegressor, GPRegressor.
STRATEGY            = DYCORSStrategy    ; Strategy in tuning algorithm - HORD i.e. DYCORSStrategy, SRBFStrategy, SOPStrategy, EIStrategy.

[sensitize]
# Sensitization Algorithm Configuration.
EPOCH       = 5         ; Modle train epoch in Sensitization Algorithm, improve the accuracy and running time
TOPN        = 10        ; The top number to select sensitive knobs.
THRESHOLD   = 0.9       ; The sensitivity threshold to select sensitive knobs.

[log]
# Configuration about log
LOGFILE_PATH        = /var/log/keentune/brain.log   ; Log file of brain
CONSOLE_LEVEL       = INFO                          ; Console Log level
LOGFILE_LEVEL       = DEBUG                         ; File saved log level
LOGFILE_INTERVAL    = 1                             ; The interval of log file replacing
LOGFILE_BACKUP_COUNT= 14                            ; The count of backup log file  
```
---  
## KeenTune-target
<https://gitee.com/anolis/keentune_target/blob/master/agent/target.conf>  
**Installed in /etc/keentune/conf/target.conf**  
```conf
[agent]
# Basic Configuration
KeenTune_HOME       = /etc/keentune/                ; KeenTune-target install path.
KeenTune_WORKSPACE  = /var/keentune/                ; KeenTune-target workspace.
AGENT_PORT          = 9873                          ; KeenTune-target service port
ORIGINAL_CONF       = /var/keentune/OriginalBackup  ; Original configuration backup path.

[log]
# Configuration about log
LOGFILE_PATH        = /var/log/keentune/brain.log   ; Log file of brain
CONSOLE_LEVEL       = INFO                          ; Console Log level
LOGFILE_LEVEL       = DEBUG                         ; File saved log level
LOGFILE_INTERVAL    = 1                             ; The interval of log file replacing
LOGFILE_BACKUP_COUNT= 14                            ; The count of backup log file  
```
---  
## KeenTune-bench
<https://gitee.com/anolis/keentune_bench/blob/master/bench/bench.conf>  
**Installed in /etc/keentune/conf/bench.conf**  
```conf
[bench]
# Basic Configuration
KeenTune_HOME       = /etc/keentune/    ; KeenTune-bench install path.
KeenTune_WORKSPACE  = /var/keentune/    ; KeenTune-bench workspace.
BENCH_PORT          = 9874              ; KeenTune-bench service port

[log]
# Configuration about log
LOGFILE_PATH        = /var/log/keentune/bench.log   ; Log file of bench
CONSOLE_LEVEL       = INFO                          ; Console Log level
LOGFILE_LEVEL       = DEBUG                         ; File saved log level
LOGFILE_INTERVAL    = 1                             ; The interval of log file replacing
LOGFILE_BACKUP_COUNT= 14                            ; The count of backup log file  
```
