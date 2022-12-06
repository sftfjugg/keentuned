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
CONFIG = "cpu2017-20220610-32039.cfg"
LOADTYPE="502 520 525"
COMMAND = "cd /home/walter/cpu2017 && source /home/walter/cpu2017/shrc && source /etc/profile && runcpu --config={} --action=run --copies=32 --threads=1 --iterations=1 --size=ref --output_format=txt --noreportable {}".format(CONFIG,LOADTYPE)

class Benchmark():
    def __init__(self, command=COMMAND):
        """Init benchmark
        """
        self.command = command

    def run(self):
        """Run benchmark and parse output

        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        time.sleep(30)
        cmd = self.command
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
            Txt_name = re.compile(r'([/\w\.]+)intrate.refrate.txt')
            if not re.search(Txt_name, self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []
            txt_name = re.search(Txt_name, self.out).group()

            with open(txt_name, 'r') as f:
                data = f.read()

            RUN_TIME = 0
            RATE = 0
            for load_name in LOADTYPE.split(' '):
                pattern = re.compile(r'{}\..*r\s+(\d+)\s+(\d+)\s+(\d+)'.format(load_name))
                if not re.search(pattern, data):
                    logger.error("can not parse output: {}".format(data))
                    return False, []
                run_time = re.search(pattern, data).group(2)
                rate = re.search(pattern, data).group(3)

                RUN_TIME += int(run_time)
                RATE += int(rate)

            result = {
                    "Run_time": RUN_TIME,
                    "Rate": RATE
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
