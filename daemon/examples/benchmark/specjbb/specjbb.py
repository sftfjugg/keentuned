import re
import sys
import time
import subprocess
import logging
logger = logging.getLogger(__name__)

"""
SPEC_cpu test ...
e.g.
runcpu --rebuild --config=cpu2017-20220610-32039.cfg --action=build --copies=32 --threads=1 --iterations=3 --size=ref --output_format=txt --noreportable 502 520 525
runcpu --config=cpu2017-20220610-32039.cfg --action=run --copies=32 --threads=1 --iterations=1 --size=ref --output_format=txt --noreportable  502 520 525
"""

#const
COMMAND = "cd /root/specjbb2015_103; bash ./cfg_os.sh; bash ./64C.sh"

class Benchmark():
    def __init__(self, command=COMMAND):
        """Init benchmark
        """
        self.command = command

    def run(self):
        """Run benchmark and parse output

        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        #time.sleep(30)
        #cmd = self.command
        #logger.info(cmd)
        #result = subprocess.run(
        #            cmd,
        #            shell=True,
        #            close_fds=True,
        #            stderr=subprocess.PIPE,
        #            stdout=subprocess.PIPE
        #        )
        #self.out = result.stdout.decode('UTF-8','strict')
        #self.error = result.stderr.decode('UTF-8','strict')
        with open("/root/specjbb2015_103/out.txt", 'r') as f:
            self.out = f.read()
        if 1:
            Out_name = re.compile(r'([/\w\.-]+)composite.out')
            if not re.search(Out_name, self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []

            out_name = re.search(Out_name, self.out).group()
            with open("/root/specjbb2015_103/" + out_name, 'r') as f:
                data = f.read()

            MAX_JOPS = 0
            CRITICAL_JOPS = 0
            MAX_JOPS_pattern = re.compile(r'max-jOPS = ((\d)+)')
            CRITICAL_JOPS_pattern = re.compile(r'critical-jOPS = ((\d)+)')
            if not re.search(MAX_JOPS_pattern, data) or not re.search(CRITICAL_JOPS_pattern, data):
                logger.error("can not parse output: {}".format(data))
                return False, []

            MAX_JOPS = int(re.search(MAX_JOPS_pattern, data).group(1))
            CRITICAL_JOPS = int(re.search(CRITICAL_JOPS_pattern, data).group(1))
            print(MAX_JOPS)
            print(CRITICAL_JOPS)
            
            AVG_JOPS = float(MAX_JOPS/102342) + float(CRITICAL_JOPS/52457)


            result = {
                    "agg_jOPS": AVG_JOPS
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

