# QuickStart
###### Version: 2.0.1

## Dependencies
安装python3运行环境  
```s
yum install python36 python36-devel
pip3 install --upgrade pip
```
安装python依赖包
```s
pip3 install hyperopt==0.1.2
pip3 install numpy==1.19.5
pip3 install POAP==0.1.26
pip3 install pySOT==0.3.3
pip3 install scikit_learn==1.1.1
pip3 install setuptools==39.2.0
pip3 install shap==0.40.0
pip3 install tornado==6.1
pip3 install xgboost==1.5.2
pip3 install pynginxconfig==0.3.4
pip3 install torch
```

## Installation
### 配置yum源
将以下keentune源追加进系统的yum源
```s
vim /etc/yum.repos.d/epel.repo
```
```s
[keentune]
name=keentune-os
baseurl=https://mirrors.openanolis.cn/anolis/8.4/Plus/x86_64/os/
gpgkey=https://mirrors.openanolis.cn/anolis/RPM-GPG-KEY-ANOLIS
enabled=1
```   
更新yum源缓存
```s
yum clean all
yum makecache
```

### 安装KeenTune
使用yum安装KeenTune各组件，使用源码安装的方法参考[这里](./2.Installation_cn.md)
```
yum install keentuned keentune-brain keentune-bench keentune-target
```

## Usages
### 启动KeenTune
```s
# 启动keentuned服务
systemctl start keentuned
# 启动keentune-brain服务
systemctl start keentune-brain
# 启动keentune-bench服务
systemctl start keentune-bench
# 启动keentune-target服务
systemctl start keentune-target
```

### Auto-Tuning(参数调优)
使用[默认配置](./install/Configuration_cn.md)开始一次[参数调优](./how-to/How%20to%20Auto-Tuning%20Knobs_cn.md)，总轮次为10轮。  
```s
keentune param tune --job tune_demo --iteration 10
```

### Profile(专家调优方案)
为[默认配置](./install/Configuration_cn.md)中[Group1](./design/target_group_cn.md)的机器设置[CPU高负载](./design/buildin_profile_cn.md)的[专家调优](./how-to/How%20to%20Auto-Tuning%20Knobs_cn.md)方案。  
```s
keentune profile set --group1 cpu_high_load.conf
```

### Sensitize Knobs(敏感参数识别)
使用[默认配置](./install/Configuration_cn.md)和Auto-Tuning中产生的数据对[参数敏感性](./how-to/How%20to%20sentivize%20knob.md)进行识别
```s
keentune sensitize train --data tune_demo --output tune_demo
```

### Pressure Control(智能控压)
在http长链接工作负载下对[wrk](./5.benchmkar_wrk.md)的参数进行[智能化控制](./how-to/How%20to%20sentivize%20knob.md)，总轮次为10轮。
```s
keentune param tune --param wrk.json --bench wrk_nginx_long.json   --job wrk_demo --iteration 10
```
