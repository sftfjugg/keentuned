# 优化原理
代码大页特性主要为大代码段业务服务，可以降低程序的iTLB miss，从而提升程序性能。
# 使用方法
在Alinux3/Anolis8上，该特性默认是关闭的。可以使用sysfs接口进行启用，启用方式有三种不同的方式。
## 启用方法
方式一：仅打开二进制和动态库大页
```
echo 1 > /sys/kernel/mm/transparent_hugepage/hugetext_enabled
```
方式二：仅打开可执行匿名大页
```
echo 2 > /sys/kernel/mm/transparent_hugepage/hugetext_enabled
```
方式三：同时打开以上两类大页
```
echo 3 > /sys/kernel/mm/transparent_hugepage/hugetext_enabled
```
## 关闭方法
使用sysfs接口关闭代码大页：
```
echo 0 > /sys/kernel/mm/transparent_hugepage/hugetext_enabled
```
同时，注意释放已使用的大页，也有几种不同方式：
方式一：清理整个系统的page cache
```
echo 3 > /proc/sys/vm/drop_caches 
```
方式二：清理单个文件的page cache
```
vmtouch -e /<path>/target
```
方式三：清理遗留大页
```
echo 1 > /sys/kernel/debug/split_huge_pages
```
## 检查是否启用代码大页
查看/proc/<pid>/smaps中FilePmdMapped字段可确定是否使用了代码大页。
扫描进程代码大页使用数量（**单位KB**）：
```
cat /proc/<pid>/smaps | grep FilePmdMapped | awk '{sum+=$2}END{print"Sum= ",sum}'
```
# 性能收益
该特性在不同平台优化效果不同，主要原因在于平台TLB的设计。当前已知较适用场景包括mysql、postgresql以及oceanbase，优化效果在5~8%。

---

# 附录-1：进一步优化——Padding
## 优化原理
Padding特性是对代码大页特性的优化，主要解决在分配给大页使用的内存剩余量不足以再分配出一个新的大页时导致的内存碎片问题。该特性需要在启用代码大页的基础上使用，不可独立使用。
> 举例说明：当二进制文件末尾剩余text段由于不足2M而无法使用大页，当剩余text大小超过hugetext_pad_threshold值，可将其填充为2M text，保证可使用上大页。

## 使用方法
### 启用方法
同样使用sysfs接口启用padding特性：
```
echo [0~2097151] >  /sys/kernel/mm/transparent_hugepage/hugetext_pad_threshold
```
> 建议一般情况写4096即可：echo 4096 >  /sys/kernel/mm/transparent_hugepage/hugetext_pad_threshold

### 关闭方法
使用sysfs接口关闭padding特性，同时注意释放已使用的大页（参考“代码大页”的关闭方法）。
```
echo 0 >  /sys/kernel/mm/transparent_hugepage/hugetext_pad_threshold
```
# 附录-2：注意事项

1. 打开、关闭并不意味着立即合并、拆散大页，hugetext 是异步的。
1. 如果一段代码曾经被整理成大页，即使关闭 hugetext 功能，还是会大页映射。
1. 在测试性能时，为了消除这些影响，可以通过 `echo 3 > /proc/sys/vm/drop_caches` 来回收整理的大页，下次就是小页映射了。
1. 想确认代码段是否大页映射，可以通过 `grep FilePmdMapped /proc/$(pidof mysqld)/smaps` 来确认。
