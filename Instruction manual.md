# Installation
## keentuned
### 1. 安装golang
安装golang编译环境
```sh
wget https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo "export PATH=\$PATH:/usr/local/go/bin" >>~/.bashrc    
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct
```
### 2. 编译安装keentuned
```
git clone https://gitee.com/anolis/keentuned.git -b dev-1.3.3
cd keentuned
bash keentuned_install.sh
```
### 3. 配置target ip
```sh
vim /etc/keentune/conf/keentuned.conf
# 修改配置文件中TARGET_IP的值
```
```conf
[target-group-1]
#* Topology of target group and knobs to be tuned in target. *#
# The machine ip address to depoly keentune-target.
TARGET_IP   = [target ip address]
# The service port of keentune-target.
TARGET_PORT = 9873
# Knobs to be tuned in this target
PARAMETER   = sysctl.json
```
### 4. 启动
```
keentuned
```

## keentune-target
### 1. 下载keentune-target
```
git clone https://gitee.com/anolis/keentune_target.git -b dev-1.3.3
```
### 2. 安装python依赖
```
cd keentune_target
pip3 install -r requirements.txt
```
### 3. 安装keentune-target
```
python3 setup.py install
```
### 4. 启动
```
keentune-target
```