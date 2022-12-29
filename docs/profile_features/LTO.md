# 优化原理
LTO，顾名思义，就是在链接的时候，继续做一系列的优化。又由于编译器的优化器是以IR为输入，因此编译器在链接时优化的编译过程中将每个源文件生成的不再是真正的二进制文件，而是"LTO对象文件"。这些文件中存储的不是二进制代码，而是编译器生成的IR。
链接时，编译器会将所有"LTO对象文件"合并到一起，合并后生成的巨大IR文件包含了所有源代码信息，这样就能跨文件对所有源代码（合并后的IR文件）进行集中的分析和优化，并生成一个真正的二进制文件，链接并得到可执行文件。因为编译器能够一次性看到所有源代码的信息，所以它做的分析具有更高的确定性和准确性，能够实施的优化也要多得多。
但在使用LTO优化大型的应用程序的时候，“链接”阶段将所有"LTO对象文件"合并成一个IR文件并生成的单个二进制文件可能非常庞大，这给系统虚拟内存造成很大压力，导致频繁内存换页，最终影响编译、链接的效率，甚至出现out of memory错误。
实际上，GCC编译器在LTO阶段首先会将所有文件中的符号表（symbol table）信息和调用图（call graph）信息合并到一起，这样不仅能跨文件对所有的函数进行全局的分析，而且内存压力也较小。WPA（whole program analysis, 全局分析）阶段会确定具体的优化策略，并依据优化策略将所有的"LTO对象文件"合并成一或多个IR文件。
![image.png](https://intranetproxy.alipay.com/skylark/lark/0/2022/png/337836/1661235609734-7ba8d8fa-ad68-48ab-b4dd-d5067bfbe1f8.png#clientId=u03c343e1-5eb0-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=208&id=u38b8d83b&margin=%5Bobject%20Object%5D&name=image.png&originHeight=415&originWidth=750&originalType=binary&ratio=1&rotation=0&showTitle=false&size=50664&status=done&style=none&taskId=uae858605-9a2f-42f2-b4fd-a200b49becc&title=&width=375)
这些IR文件再次并发经过优化器优化后，生成真正的二进制对象文件并传递给真正的链接器进行链接，最后得到可执行文件。
![image.png](https://intranetproxy.alipay.com/skylark/lark/0/2022/png/337836/1661235644272-166afe2c-01ca-4321-9cff-32cb57d98471.png#clientId=u03c343e1-5eb0-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=144&id=ub8d56a1c&margin=%5Bobject%20Object%5D&name=image.png&originHeight=287&originWidth=750&originalType=binary&ratio=1&rotation=0&showTitle=false&size=37694&status=done&style=none&taskId=ue2789136-02a4-4e5b-9b29-83087681c78&title=&width=375)
# 使用方法
举例说明如何使用LTO来进行编译优化：

- 编写了一个包含两个源文件（source1.c 和 source2.c）和一个头文件（source2.h）的程序。由于头文件中非常简单地包含了source2.c中的函数原型， 所以并没有列出。
```
/* source1.c */
#include <stdio.h> // scanf_s and printf.
#include "source2.h"
int square(int x) { return x*x; }
int main() {
  int n = 5, m;
  scanf("%d", &m);
  printf("The square of %d is %d.\n", n, square(n));
  printf("The square of %d is %d.\n", m, square(m));
  printf("The cube of %d is %d.\n", n, cube(n));
  printf("The sum of %d is %d.\n", n, sum(n));
  printf("The sum of cubes of %d is %d.\n", n, sumOfCubes(n));
  printf("The %dth prime number is %d.\n", n, getPrime(n));
}
```
```
/* source2.c */
#include <math.h> // sqrt.
#include <stdbool.h> // bool, true and false.
#include "source2.h"
int cube(int x) { return x*x*x; }
int sum(int x) {
  int result = 0;
  for (int i = 1; i <= x; ++i) result += i;
  return result;
}
int sumOfCubes(int x) {
  int result = 0;
  for (int i = 1; i <= x; ++i) result += cube(i);
  return result;
}
static
bool isPrime(int x) {
  for (int i = 2; i <= (int)sqrt(x); ++i) {
    if (x % i == 0) return false;
  }
  return true;
}
int getPrime(int x) {
  int count = 0;
  int candidate = 2;
  while (count != x) {
    if (isPrime(candidate))
      ++count;
  }
  return candidate;
}
```
source1.c文件包含两个函数，有一个参数并返回这个参数的平方的square函数，以及程序的main函数。main函数调用source2.c中除了isPrime之外的所有函数。source2.c有5个函数。cube返回一个数的三次方；sum函数返回从1到给定数的和；sumOfCubes返回1到给定数的三次方的和；isPrime用于判断一个数是否是质数；getPrime函数返回第x个质数。笔者省略掉了容错处理因为那并非本文的重点。
这些代码简单但是很有用。其中一些函数只进行简单的运算，一些需要简单的循环。getPrime是当中最复杂的函数，包含一个while循环且在循环内部调用了也包含一个循环的isPrime函数。通过这些函数，我们将在LTO下看到函数内联，常量折叠等编译器最重要，最常见的优化。

- 编译命令（gcc版本是9.0.1）
```
# Non-LTO:
gcc source1.c source2.c -O2 -o source-nolto
# LTO:
gcc source1.c source2.c -flto -O2 -o source-lto
```

- 检查source-nolto的汇编
```
1. 可以发现square被main函数inline了，对于第一次square(n)的调用，由于n在编译时刻可知，所以代码
   被常数25替代，对于square的第二次调用square（m），由于m的值在编译时是未知的，所以编译器不能对
   计算估值。
2. 可以发现isPrime被getPrime函数inline了，这是由于candidate有初始值，inline后可以做常量折叠；
   另外一方面，isPrime是static函数，且只有getPrime调用了它，inline之后，可以将这个函数当作dead
   code删除，减小代码体积。
```

- 检查source-lto的汇编
```
1. 可以看到除了scanf之外的所有函数都被作为了内联函数，因为在GCC LTO的WPA阶段, 这些函数的定义都是
   可见的。并且，由n编译时刻已知，编译器在编译时刻就会完成函数square(n), cube(n)，sum(n)和
   sumOfCubes(n)的计算。
2. LTO获得了30%以上的性能提升
```

- 结论
```
1. 某些优化，当其作用于整个程序级别时，往往比其作用于局部时更加有效，函数内联就是这种类型的优化之一。
2. 事实上，大多数优化都在全局范围内都更加有效，这也就是我们编译大型应用程序时，推荐使用LTO的原因和理由。
3. FireFox、Chrome、 Mysql等大型应用使用LTO build都获得了5%-15%不等的性能提升
```
