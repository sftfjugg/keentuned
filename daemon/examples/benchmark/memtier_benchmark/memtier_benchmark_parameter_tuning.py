#!/usr/bin/python3
# -*- coding: utf-8 -*-
import re
import sys
import time
import subprocess
import logging

"""
memtier_benchmark benchmark
"""

logger = logging.getLogger(__name__)

# baseline parameters
KEY_MAXIMUM=100000
RATIO="1:1"
PIPELINE=1
TEST_TIME=6

DEFAULT = "--threads=8 --clients=10 --data-size=32"

script_file="benchmark_run.sh"

class Benchmark:
    def __init__(self, url, default=DEFAULT, ratio=RATIO, pipeline=PIPELINE, test_time=TEST_TIME, key_maximum=KEY_MAXIMUM):
        """Init benchmark

        Args:
            url (string): url.
            default (string): Number of threads to use and Connections to keep open.
            duration (int, optional): Duration of test.
        """
        # Modify the test command based on the actual scenario
        self.DEFAULT_CMD = " memtier_benchmark  -s {} -p $port {} --test-time={} --ratio={} --pipeline={}  --key-maximum={} ".format(url, default, test_time, ratio, pipeline, key_maximum)

    def __transfMeasurement(self,value,measurement):
        if measurement == '':
            return value

        # measurement of Latency
        elif measurement == 'h':
            return value * 60 * 60 * 10 ** 6
        elif measurement == 'm':
            return value * 60 * 10 ** 6
        elif measurement == "s":
            return value * 10 ** 6
        elif measurement == "ms":
            return value * 10 ** 3
        elif measurement == 'us':
            return value

        # measurement of Req/Sec
        elif measurement == "k":
            return value * 10 ** 3
        elif measurement == 'M':
            return value * 10 ** 6
        elif measurement == 'G':
            return value * 10 ** 9

        # measurement of Transfer/sec
        elif measurement == "KB":
            return value * 10 ** 3
        elif measurement == 'MB':
            return value * 10 ** 6
        elif measurement == 'GB':
            return value * 10 ** 9

        else:
            logger.warning("Unknown measurement: {}".format(measurement))
            return value

    def run(self):
        """Run benchmark and parse output

        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        redisshell="""#!/bin/bash
instance=2
REDIS_PORT=9400
cpu_cores=`cat /proc/cpuinfo | grep processor | wc -l`
process_cpu=$(($cpu_cores / $instance))

client_pids=\"\"
for i in $(seq 1 $instance); do
    cpu_start=$((($i - 1) * $process_cpu))
    cpu_end=$(($i * $process_cpu - 1))
    port=$(($REDIS_PORT + i - 1))
    client_cmd=\"numactl -C $cpu_start-$cpu_end %s\"
    ${client_cmd} &
    client_pids=\"${client_pids} $!\"
done

for pid in ${client_pids}; do wait ${pid} ; done

""" % self.DEFAULT_CMD

        with open(script_file, "w", encoding='UTF-8') as f:
            f.write(redisshell)

        cmd = "sh {}".format(script_file)
        result = subprocess.run(
                    cmd,
                    shell=True,
                    close_fds=True,
                    stderr=subprocess.PIPE,
                    stdout=subprocess.PIPE
                )
        self.out = result.stdout.decode('UTF-8','strict')
        print(self.out)
        self.error = result.stderr.decode('UTF-8','strict')
        if result.returncode == 0:
            logger.info(self.out)
            cmd = "echo \"%s\" | awk 'BEGIN{ops_per_sec = 0} /^Totals/{ops_per_sec += $2} END{print ops_per_sec}'" % self.out
            result = subprocess.run(
                    cmd,
                    shell=True,
                    close_fds=True,
                    stderr=subprocess.PIPE,
                    stdout=subprocess.PIPE
                )
            self.options = result.stdout.decode('UTF-8','strict')

            result = {"OPS": float(self.options)}
            result_str = ", ".join(["{} = {}".format(k,v) for k,v in result.items()])
            print(result_str)
            return True, result_str

        else:
            logger.error(self.error)
            return False, []


if __name__ == "__main__":
    if sys.argv.__len__() <= 1:
        print("'Target ip' is wanted: python3  wrk_parameter_tuning.py [Target ip]")
        exit(1)
    bench = Benchmark(sys.argv[1])
    suc, res = bench.run()

