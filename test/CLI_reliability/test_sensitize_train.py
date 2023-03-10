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
from common import getTaskLogPath
from common import getTrainTaskResult


class TestSensitizeTrain(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestSensitizeTrain begin...")
        status = runParamTune("param1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestSensitizeTrain end...")

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_sensitize_train testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_sensitize_train testcase finished')

    def delete_train_job(self, job):
        cmd = 'echo y | keentune sensitize delete --job {}'.format(job)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))

    def check_train_job(self, job):
        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        result = getTrainTaskResult(path)
        self.assertTrue(result)
        
        self.path = "/var/keentune/sensitize_workspace/{}/knobs.json".format(job)
        res = os.path.exists(self.path)
        self.assertTrue(res)

    def test_sensitize_train_RBT_lose_data_param(self):
        cmd = 'echo y | keentune sensitize train --job param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_train_RBT_lose_data_value(self):
        cmd = 'echo y | keentune sensitize train --data --job param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('find unexpected characters'))

    def test_sensitize_train_RBT_data_value_null(self):
        cmd = "echo y | keentune sensitize train --data '' --job param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_train_RBT_data_value_empty(self):
        cmd = "echo y | keentune sensitize train --data ' ' --job param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_train_RBT_lose_job_param(self):
        cmd = 'echo y | keentune sensitize train --data param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job("param1")
        self.delete_train_job("param1")

    def test_sensitize_train_RBT_lose_job_value(self):
        cmd = 'echo y | keentune sensitize train --data param1 --job'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --job'))

    def test_sensitize_train_RBT_job_value_null(self):
        cmd = "echo y | keentune sensitize train --data param1 --job ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job("param1")
        self.delete_train_job("param1")

    def test_sensitize_train_RBT_job_value_empty(self):
        cmd = "echo y | keentune sensitize train --data param1 --job ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job("param1")
        self.delete_train_job("param1")

    def test_sensitize_train_RBT_job_value_repeat(self):
        cmd = "echo y | keentune sensitize train --data param1 --job param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job("param1")
        cmd = "echo y | keentune sensitize train --data param1 --job param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__("name 'param1' already exists"))
        self.delete_train_job("param1")

    def test_sensitize_train_RBT_lose_trials_param(self):
        cmd = 'echo y | keentune sensitize train --data param1 --job param1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        pattern = re.compile(r'trials:\s*(\d+)')
        trials = re.search(pattern, self.out).group(1)
        self.assertEqual(trials, "3")
        self.check_train_job("param1")
        self.delete_train_job("param1")

    def test_sensitize_train_RBT_lose_trials_value(self):
        cmd = 'echo y | keentune sensitize train --data param1 --job param1 --trials'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --trials'))

    def test_sensitize_train_RBT_trials_value_error(self):
        cmd = 'echo y | keentune sensitize train --data param1 --job param1 --trials -o'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

    def test_sensitize_train_RBT_trials_value_null(self):
        cmd = "echo y | keentune sensitize train --data param1 --job param1 --trials ''"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))
    
    def test_sensitize_train_RBT_trials_value_empty(self):
        cmd = "echo y | keentune sensitize train --data param1 --job param1 --trials ' '"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

