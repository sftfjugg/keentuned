import os
import sys
import logging
import re
import subprocess
import time
import unittest

logger = logging.getLogger(__name__)
sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runParamTune
from common import getTuneTaskResult
from common import getTaskLogPath
from common import copyTmpFile

logger = logging.getLogger(__name__)

class TestParamTune(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestParamDump begin...")

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestParamDump end...")

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_param_tune testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_tune testcase finished')

    def check_param_tune_job(self, name):
        cmd = 'keentune param jobs'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__(name))

    def reset_keentuned(self, config, file):
        cmd = r"sh conf/reset_keentuned.sh {} '{}'".format(config, file)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("restart keentuned server successfully!", self.out)

    def test_param_tune_RBT_lose_job_param(self):
        cmd = 'keentune param tune -i 10'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_param_tune_RBT_lose_job_value(self):
        cmd = 'keentune param tune -i 10 --job'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --job'))   

    def test_param_tune_RBT_job_value_null(self):
        cmd = "keentune param tune -i 10 --job ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_param_tune_RBT_job_value_empty(self):
        cmd = "keentune param tune -i 10 --job ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("Incomplete or Unmatched command"))

    def test_param_tune_RBT_job_value_repeat(self):
        self.status = runParamTune("param1")
        self.assertEqual(self.status, 0)
        cmd = 'keentune param tune -i 10 --job param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("already exists"))
        deleteDependentData("param1")

    def test_param_tune_RBT_lose_iteration_value(self):
        cmd = "keentune param tune -i --job param3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

    def test_param_tune_RBT_iteration_value_error(self):
        cmd = "keentune param tune -i -o --job param3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

    def test_param_tune_RBT_iteration_value_null(self):
        cmd = "keentune param tune -i '' --job param3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

    def test_param_tune_RBT_iteration_value_empty(self):
        cmd = "keentune param tune -i ' ' --job param3"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

    def test_param_tune_RBT_lose_debug_value(self):
        cmd = 'keentune param tune -i 10 --job param4 --debug'
        path = getTaskLogPath(cmd)

        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if "Time cost statistical information" in res_data or "[ERROR]" in res_data:
                break
            time.sleep(8)
        word_list = ["Step1", "Step6", "[BEST] Tuning improvement", "Time cost statistical information"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)
        deleteDependentData("param4")

    def test_param_tune_RBT_debug_value_error(self):
        cmd = "keentune param tune -i 10 --job param4 --debug -0"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("unknown shorthand flag: '0' in -0"))

    def test_param_tune_RBT_lose_optional_params(self):
        cmd = 'keentune param tune --job param5'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        pattern = re.compile(r'iteration:\s*(\d+)')
        iteration = re.search(pattern, self.out).group(1)
        self.assertEqual(iteration, "100")
        
        cmd = 'keentune param stop'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Abort parameter optimization job'))
        time.sleep(5)
        deleteDependentData("param5")

    def test_param_tune_RBT_job_special_chars_01(self):
        cmd = 'keentune param tune -i 10 --job _param11'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("_param11")
        deleteDependentData("_param11")

    def test_param_tune_RBT_job_special_chars_02(self):
        cmd = 'keentune param tune -i 10 --job /root/test'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--job find unexpected characters '/'"))

    def test_param_tune_RBT_job_special_chars_03(self):
        cmd = 'keentune param tune -i 10 --job test,a'
        self.status, self.out, _ = sysCommand(cmd)
        time.sleep(20)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--job find unexpected characters ','"))

    def test_param_tune_RBT_job_special_chars_04(self):
        cmd = 'keentune param tune -i 10 --job _test+1.1'
        self.status, self.out, _ = sysCommand(cmd)
        time.sleep(20)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("--job find unexpected characters '+' '.'"))

    def test_param_tune_RBT_job_special_chars_05(self):
        cmd = 'keentune param tune -i 10 --job param3 a'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("param3")
        deleteDependentData("param3")

    def test_param_tune_RBT_iteration_value_negative(self):
        cmd = 'keentune param tune -i -1 --job param3'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("iteration >= 10"))

    def test_param_tune_RBT_iteration_value_zero(self):
        cmd = 'keentune param tune -i 0 --job param3'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("iteration >= 10"))

    def test_param_tune_RBT_iteration_value_float(self):
        cmd = 'keentune param tune -i 5.2 --job param3'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("Error: invalid argument"))

    def test_param_tune_RBT_iteration_value_str(self):
        cmd = 'keentune param tune -i "hello" --job param3'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__("Error: invalid argument"))

    def test_param_tune_RBT_iteration_value_str_number(self):
        cmd = 'keentune param tune -i "10" --job param3'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("param3")
        deleteDependentData("param3")

    def test_param_tune_RBT_multi_config(self):
        self.reset_keentuned("param", "sysctl.json, nginx_conf.json")
        cmd = 'keentune param tune -i 10 --job param3'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("param3")
        deleteDependentData("param3")
        self.reset_keentuned("param", "sysctl.json")
