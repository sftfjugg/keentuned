# 优化原理
PGO全称profile guided optimization，主要是为了解决传统编译器在执行优化的时候，只是是基于静态代码信息，而不去考虑用户可能的输入，从而无法有效对代码进行有效优化的问题。
PGO可以分为三个阶段，分别是instrument，train，optimize三个阶段。在instrument阶段中，会先对应用做一次编译。在这次编译中，编译器会向代码中插入一下指令，以便下一阶段可以收集数据。插入的指令分为三种类型，分别用来统计：

1. 每个函数被执行了多少次
1. 每个分支被执行了多少次（例如if-else的场景）
1. 某些变量的值（主要用于switch-case的场景）

在train阶段中，用户需要使用最常用的输入来运行上一阶段编译生成的应用。由于上一阶段已经做好了收集数据的准备，在经过train阶段之后，该应用最常见的使用场景对应的数据就会被收集下来。
最后阶段是optimization阶段。在该阶段中，编译器会利用上一阶段收集到的数据，对应用进行重新编译。由于上一阶段的数据来自于用户输入的最常见的用户场景，那么最后优化得到的结果就能在该场景下有更好的优化。
# 使用方法
举例说明如何使用PGO来进行编译优化：
编写一段C++代码，该代码用较为低效的方式来判断一个数字是否为质数。代码如下：
```
//test.cpp
#include<iostream>
#include<stdlib.h>
using namespace std;
int main(int argc, char** argv){
    int num0 = atoi(argv[1]);
    int num1 = atoi(argv[1]);
    int branch = atoi(argv[2]);
    if (branch < 1){
        for (int i=2;i<=num0;i++){
            if (num0%i==0){
                cout<<i<<endl;
                break;
            }
        }
    } else {
        for (int i=2;i<num1;i++){
            if (num1%i==0){
                cout<<i<<endl;
                break;
            }
        }
    }
    return 0;
}
```
可以看到代码中根据branch的值不同，分为了两个分支。这两个分支的代码完全相同。这个是为了后续测试的目的。另外，2147483647是int范围内最大的质数，后面会用到。
#### 不使用PGO
先看下不使用PGO的情况。用下列命令编译：
```
g++ test.cpp -O3 -o test
```
执行下面两条命令得到两个分支的时间
```
time ./test 2147483647 0

real    0m6.904s
user    0m6.902s
sys     0m0.000s

time ./test 2147483647 1

real    0m6.907s
user    0m6.905s
sys     0m0.000s
```
可以看到两个分支的执行时间几乎是相同的。
#### 使用PGO
使用下面的命令做第一次编译
```
g++ test.cpp -O3 -fprofile-generate -o test.pgo_generate
```
这里得到的test.pog_generate即是前文提到的第一阶段生成用户收集数据的binary。
执行下列命令进行训练：
```
time ./test.pgo_generate 2147483647 0

real    0m11.894s
user    0m11.890s
sys     0m0.001s
```
这边只训练branch=0这个分支。可以看到由于需要收集数据，执行速度慢了很多。
接下来再做一次编译：
```
g++ test.cpp -O3 -fprofile-use -o test.pgo_use
```
这里得到的test.pgo_use即是最终经过PGO优化完成的binary。
执行下列命令测试时间
```
time ./test.pgo_use 2147483647 0

real    0m6.258s
user    0m6.255s
sys     0m0.001s

time ./test.pgo_use 2147483647 1

real    0m6.905s
user    0m6.903s
sys     0m0.000s
```
可以看到，被优化了的branch=0分支，运行速度得到了提升；而没有被优化的branch=1分支，执行时间保持不变。
这也就说明了PGO这样的优化是有效的。
