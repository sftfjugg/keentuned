# 优化原理
TLB flush range是arm64芯片在armv8.4-TLBI版本上支持的一种指令集批量处理地址刷新的特性，需要内核配置文件开启CONFIG_ARM64_TLB_RANGE功能，此外还需要arm64芯片支持armv8.4-TLBI特性。
传统的TLB flush操作方式，以stride为粒度去进行TLB flush,  这种方式存在明显弊端，对于需要刷新大量的地址范围，需要将其拆分为一个个stride粒度去进行刷新，导致耗时较多。TLB flush range的意义在于动态切割地址范围， 按照如下设计思想:
每一次刷新页数目由numa和scale两个变量决定，num值方位为（-1， 31）， 通过__TLBI_RANGE_NUM传入的剩余需要刷新的pages数以及scale的值获得num具体的值，当num=-1时，表示刷新完成。
```bash
#define TLBI_RANGE_MASK                 GENMASK_ULL(4, 0)
#define __TLBI_RANGE_NUM(pages, scale)  \
          ((((pages) >> (5 * (scale) + 1)) & TLBI_RANGE_MASK) - 1)
```
scale的值按照0逐渐递增，结合num值，统计出此次刷新的页数。
```bash
#define __TLBI_RANGE_PAGES(num, scale)   \
                  ((unsigned long)((num) + 1) << (5 * (scale) + 1))
```
按照rvale1is格式，将地址start, 以及scale和num值，写入rvale1is寄存器中，对页进行刷新操作。
```bash
__TLBI_VADDR_RANGE(start, asid, scale, num, tlb_level);
__tlbi(rvale1is, addr);
__tlbi_user(rvale1is, addr);
```
更多可参考
[https://developer.arm.com/documentation/ddi0595/2021-12/AArch64-Instructions/TLBI-RVALE1IS--TLBI-RVALE1ISNXS--TLB-Range-Invalidate-by-VA--Last-level--EL1--Inner-Shareable](https://developer.arm.com/documentation/ddi0595/2021-12/AArch64-Instructions/TLBI-RVALE1IS--TLBI-RVALE1ISNXS--TLB-Range-Invalidate-by-VA--Last-level--EL1--Inner-Shareable)
# 使用方法
### 使用Alinux3.2208及以后版本
Alinux3在2208版本（内核版本5.10.134-12_rc1）已默认启用该特性
