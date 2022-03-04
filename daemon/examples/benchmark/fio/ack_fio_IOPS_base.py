#!/usr/bin/python3
#-*- coding: utf-8 -*-
import re
import sys
import subprocess
import logging
logger = logging.getLogger(__name__)

"""
FIO test IOPS,Stress testing on the hardware
e.g.
'fio -filename=/dev/sdb -ioengine=psync -time_based=1 -rw=read -direct=1 -buffered=0 -thread -size=110g -bs=512B -numjobs=16 -iodepth=1 -runtime=300 -lockmem=1G -group_reporting -name=read'

-filename 对整个磁盘或分区测试
-ioengine 定义使用哪种IO
-rw 测试顺序（随机）读（写）的I/O,可选参数: write|read|readwrite|randwrite|randread|randrw
-direct 测试过程绕过机器自带的buffer
-thread fio使用线程而不是进程
-bs 单次io的块文件大小
-size 本次的测试文件大小
-runtime 测试时间为300秒，如果不定义时间，则一直将110g文件分512B每次写完为止
-lockmem 只使用1g内存进行测试
-iodepth 设置IO队列的深度，表示在这个文件上同一时刻运行n个I/O
"""

#const
FileName = "/dev/sdb"
Block_Size = "512B"
DEFAULT_rw = "read"
SIZE = "110"
NumJobs = "16"

class Benchmark():
    def __init__(self, filename=FileName, bs=Block_Size, rw=DEFAULT_rw, size=SIZE, numjobs=NuJobs):
        """Init benchmark
        """
        self.filename = filename
        self.bs = bs
        self.rw = rw
        self.size = size
        self.numjobs = numjobs

    def run(self):
        """Run benchmark and parse output

        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        cmd = 'fio -filename={} -ioengine=psync -time_based=1 -rw={} -direct=1 -buffered=0 -thread -size={} -bs={} -numjobs={} -iodepth=1 -runtime=300 -lockmem=1G -group_reporting -name=read'.format(
                self.filename,self.rw,self.size,self.bs,self.numjobs)
        logger.info(cmd)
        result = subprocess.run(
                    cmd,
                    shell=True,
                    close_fds=True,
                    stderr=subprocess.PIPE,
                    stdout=subprocess.PIPE
                )
        self.out = result.stdout.decode('UTF-8','strict')
        self.error = result.stderr.decode('UTF-8','strict')
        if result.returncode == 0:
            pattern_iops = re.compile(r'IOPS=([\d.]+)')
            pattern_bw = re.compile(r'BW=([\d.]+)')

            if not re.search(pattern_iops,self.out) \
                or not re.search(pattern_bw,self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []

            iops = float(re.search(pattern_iops,self.out).group(1))
            bw = float(re.search(pattern_bw,self.out).group(1))
            result = {
                "IOPS": iops,
                "BW": bw,
            }

            result_str = ", ".join(["{} = {}".format(k,v) for k,v in result.items()])
            print(result_str)
            return True, result_str
        else:
            logger.error(self.error)
            return False, []
if __name__ == "__main__":
    bench = Benchmark()
    suc, result = bench.run()


