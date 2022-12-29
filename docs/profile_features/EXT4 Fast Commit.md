# 优化原理
Fast Commit 是 Linux 5.10 引入的一个新的轻量级日志方案，根据 ATC-17 的论文 “iJournaling: Fine-Grained Journaling for Improving the Latency of Fsync System Call” 实现。
在常用的 ext4 data=ordered 日志模式下，fsync() 系统调用会因为无关 IO 操作导致显著的时延。Fast Commit 根据写入日志中的元数据反推，只提交与当前 transaction 相关的操作，从而优化 fsync() 的时延。
启用 Fast Commit 特性后，系统中将会有两个日志，快速提交日志用于可以优化的操作，常规日志用于标准提交。其中 Fast Commit 日志包含上一次标准提交之后执行的操作。
从作者的 benchmark 测试数据来看，打开 Fast Commit 特性后，本地 ext4 文件系统有 20% ~ 200% 的性能提升；NFS 场景也有 30% ~ 75% 的性能提升。

| Benchmark | Config | w/o Fast Commit | w/ Fast Commit | Delta |
| --- | --- | --- | --- | --- |
| Fsmark
Fsmark | Local, 8 threads
NFS, 4 threads | 1475.1 files/s
299.4 files/s | 4309.8 files/s
409.45 files/s | +192.2%
+36.8% |
| Dbench
Dbench | Local, 2 procs
NFS, 2 procs | 33.32 MB/s
8.84 MB/s | 70.87 MB/s
11.88 MB/s | +112.7%
+34.4% |
| Dbench
Dbench | Local, 10 procs
NFS, 10 procs | 90.48 MB/s
34.62 MB/s | 110.12 MB/s
52.83 MB/s | +21.7%
+52.6% |
| FileBench
FileBench | Local, 16 threads
NFS, 16 threads | 10442.3 ops/s
1531.3 ops/s | 18617.8 ops/s
2681.5 ops/s | +78.3%
+75.1% |

# 使用方法
## 获取支持Fast Commit特性的e2fsprogs
### 使用Alinux3.2208及以后版本
Alinux3在2208版本（e2fsprogs版本高于1.46.0）已默认启用该特性
### 下载最新的 e2fsprogs 包并编译
```
wget https://git.kernel.org/pub/scm/fs/ext2/e2fsprogs.git/snapshot/e2fsprogs-1.46.2.tar.gz
tar -xvf e2fsprogs-1.46.2.tar.gz
cd e2fsprogs-1.46.2
./configure
make
```
## 格式化打开Fast Commit特性
```
./misc/mke2fs -t ext4 -O fast_commit /dev/vdc1
```
dumpe2fs 可以看到已经打开 fast commit:

```
Filesystem features: has_journal ext_attr resize_inode dir_index fast_commit filetype extent 64bit flex_bg sparse_super large_file huge_file uninit_bg dir_nlink extra_isize
```
此外，超级块中多出 Overhead blocks 字段：

```
Overhead blocks: 126828
```
同时，Journal size 也由默认的 128M 变成 130M，应该是默认 fast commit journal size 为 journal size / 64 带来的增量：
```
Journal size: 130M
```
