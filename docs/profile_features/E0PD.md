# 优化原理
E0PD是ARMv8.5扩展引入的一个硬件防护特性，它用来替代KPTI避免来自用户态的Meltdown漏洞攻击。KPTI技术通过在返回用户态时unmap kernel space的方式避免内核地址空间的暴露，因此在退出内核态时unmap内核页表以及在进入内核态时重新映射内核页表会带来极大性能开销。E0PD在硬件侧防护内核地址空间从而使内核可以绕过KPTI，因而带来性能提升。

在内核优化方面，开启E0PD绕过KPTI的代码如下（省略一些无关的代码片段）。在boot阶段，内核调用unmap_kernel_at_el0函数检测是否需要开启KPTI unmap内核地址空间功能，其中的重要一项检测就是调用kaslr_requires_kpti，在此函数内检查硬件是否支持E0PD并且内核是否打开E0PD，若是则返回false，表示不需要启用KPTI功能。

```
static int __kpti_forced; /* 0: not forced, >0: forced on, <0: forced off */

static bool unmap_kernel_at_el0(const struct arm64_cpu_capabilities *entry,
				int scope)
{
	…………

	/* Useful for KASLR robustness */
	if (kaslr_requires_kpti()) {
		if (!__kpti_forced) {
			str = "KASLR";
			__kpti_forced = 1;
		}
	}

  …………

	/* Forced? */
	if (__kpti_forced) {
		pr_info_once("kernel page table isolation forced %s by %s\n",
			     __kpti_forced > 0 ? "ON" : "OFF", str);
		return __kpti_forced > 0;
	}

	…………
}
```
```
/*
 * This check is triggered during the early boot before the cpufeature
 * is initialised. Checking the status on the local CPU allows the boot
 * CPU to detect the need for non-global mappings and thus avoiding a
 * pagetable re-write after all the CPUs are booted. This check will be
 * anyway run on individual CPUs, allowing us to get the consistent
 * state once the SMP CPUs are up and thus make the switch to non-global
 * mappings if required.
 */
bool kaslr_requires_kpti(void)
{
	if (!IS_ENABLED(CONFIG_RANDOMIZE_BASE))
		return false;

	/*
	 * E0PD does a similar job to KPTI so can be used instead
	 * where available.
	 */
	if (IS_ENABLED(CONFIG_ARM64_E0PD)) {
		u64 mmfr2 = read_sysreg_s(SYS_ID_AA64MMFR2_EL1);
		if (cpuid_feature_extract_unsigned_field(mmfr2,
						ID_AA64MMFR2_E0PD_SHIFT))
			return false;
	}

	/*
	 * Systems affected by Cavium erratum 24756 are incompatible
	 * with KPTI.
	 */
	if (IS_ENABLED(CONFIG_CAVIUM_ERRATUM_27456)) {
		extern const struct midr_range cavium_erratum_27456_cpus[];

		if (is_midr_in_range_list(read_cpuid_id(),
					  cavium_erratum_27456_cpus))
			return false;
	}

	return kaslr_offset() > 0;
}
```
# 
使用方法
## 使用Alinux3.2208及以后版本
Alinux3在2208版本（内核版本5.10.134）已默认启用
说明：从upstream打上E0PD相关patch（anck 5.10已经合入），

- 3e6c69a arm64: Add initial support for E0PD
- 92ac6fd arm64: Don't use KPTI where we have E0PD
## 编译内核时启用E0PD特性
如果需要自己编译内核，可以在编译参数中打开E0PD配置
```
CONFIG_ARM64_E0PD=y
```
# 性能收益
KPTI的内核地址空间map、unmap操作在通用路径上，所以E0PD对KPTI的优化影响广泛。在我们的测试中，E0PD优化对基础benchmark和E2E应该都带来了5-23%的大幅度性能提升。
