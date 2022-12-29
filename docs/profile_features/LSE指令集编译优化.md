# 优化原理
LSE（Large System Extensions）是ARMv8.1新增的原子操作指令集。
在LSE之前，如果想实现某个原子操作，必须要使用带有load_acquire/store_release的指令，如LDXR和STXR,但这两个指令的操作本质上是很多CPU核去抢某个内存变量的独占访问，以前ARM主要用来在低功耗设备上运行，CPU核并不多，不会存在太大的问题。但在数据中心发展场景下，ARM处理器已经发展到几十上百核，如果还是独占访问会存在严重的性能问题。因此，为了支持这种大型系统，在ARMv8.1中特意加入了大量原生原子操作指令以优化性能。在有较多多线程竞争的场景下，使用LSE指令集会有比较明显的性能提升。
# 使用方法
倚天710建议指定march=armv8.6+sve2 mtune=neoverse-n1
PS: LSE在armv8.1以后by default支持，指定armv8.6或neoverse-n1 都会使用LSE进行编译
