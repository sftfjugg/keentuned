#!/usr/bin/python3
# -*- coding: utf-8 -*-
import re
import sys
import time
import subprocess
import logging

"""
wrk benchmark
"""

logger = logging.getLogger(__name__)

# baseline parameters
DEFAULT_DURATION = 30
DEFAULT = "--threads 2 --connections 10"


class Benchmark:
    def __init__(self, url, default=DEFAULT, duration=DEFAULT_DURATION):
        """Init benchmark

        Args:
            url (string): url.
            default (string): Number of threads to use and Connections to keep open.
            duration (int, optional): Duration of test.
        """
        # Modify the test command based on the actual scenario
        self.CMD = "wrk {} --duration {} --latency http://{}".format(default, duration, url)

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
        time.sleep(5)
        cmd = self.CMD
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
            logger.info(self.out)

            pattern_latency_90 = re.compile(r'90%\s+([\d.]+)(\w+)')
            pattern_latency_99 = re.compile(r'99%\s+([\d.]+)(\w+)')
            pattern_Requests_sec = re.compile(r'Requests/sec:\s+([\d.]+)')
            pattern_Transfer_sec = re.compile(r'Transfer/sec:\s+([\d.]+)(\w+)')

            if not re.search(pattern_latency_90,self.out) \
                or not re.search(pattern_latency_99,self.out) \
                or not re.search(pattern_Requests_sec,self.out) \
                or not re.search(pattern_Transfer_sec,self.out):
                logger.error("can not parse output: {}".format(self.out))
                print(self.error, cmd)
                return False, []

            _latency_90 = float(re.search(pattern_latency_90,self.out).group(1))
            latency_90_measurement = re.search(pattern_latency_90,self.out).group(2)
            latency_90 = self.__transfMeasurement(_latency_90, latency_90_measurement)

            _latency_99 = float(re.search(pattern_latency_99,self.out).group(1))
            latency_99_measurement = re.search(pattern_latency_99,self.out).group(2)
            latency_99 = self.__transfMeasurement(_latency_99, latency_99_measurement)

            Requests_sec = float(re.search(pattern_Requests_sec,self.out).group(1))

            _Transfer_sec = float(re.search(pattern_Transfer_sec,self.out).group(1))
            Transfer_sec_measurement = re.search(pattern_Transfer_sec,self.out).group(2)
            Transfer_sec = self.__transfMeasurement(_Transfer_sec, Transfer_sec_measurement)

            result = {
                "Latency_90": latency_90,
                "Latency_99": latency_99,
                "Requests_sec": Requests_sec,
                "Transfer_sec": Transfer_sec,
            }

            result_str = ", ".join(["{} = {}".format(k,v) for k,v in result.items()])
            print(result_str)
            return True, result_str

        else:
            logger.error(self.error)
            print(self.error, cmd)
            return False, []



if __name__ == "__main__":
    if sys.argv.__len__() <= 1:
        print("'Target ip' is wanted: python3  wrk_parameter_tuning.py [Target ip]")
        exit(1)
    bench = Benchmark(sys.argv[1])
    suc, res = bench.run()



