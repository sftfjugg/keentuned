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
from common import runSensitizeCollect
from common import getTaskLogPath
from common import getTrainTaskResult


class TestSensitizeTrain(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestSensitizeTrain begin...")
        status = runSensitizeCollect("sensitize1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("sensitize1")
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

    def check_train_job(self):
        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        result = getTrainTaskResult(path)
        self.assertTrue(result)

        self.path = "/var/keentune/sensitize/sensi-sensitize1.json"
        res = os.path.exists(self.path)
        self.assertTrue(res)

    def test_sensitize_train_RBT_lose_data_param(self):
        cmd = 'echo y | keentune sensitize train --output sensitize1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_train_RBT_lose_data_value(self):
        cmd = 'echo y | keentune sensitize train --data --output sensitize1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('--data file [--output] does not exist'))

    def test_sensitize_train_RBT_data_value_null(self):
        cmd = "echo y | keentune sensitize train --data '' --output sensitize1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_train_RBT_data_value_empty(self):
        cmd = "echo y | keentune sensitize train --data ' ' --output sensitize1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_sensitize_train_RBT_lose_output_param(self):
        cmd = 'echo y | keentune sensitize train --data sensitize1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job()

    def test_sensitize_train_RBT_lose_output_value(self):
        cmd = 'echo y | keentune sensitize train --data sensitize1 --output'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --output'))

    def test_sensitize_train_RBT_output_value_null(self):
        cmd = "echo y | keentune sensitize train --data sensitize1 --output ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job()

    def test_sensitize_train_RBT_output_value_empty(self):
        cmd = "echo y | keentune sensitize train --data sensitize1 --output ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Running Sensitize Train Success'))
        self.check_train_job()

    def test_sensitize_train_RBT_lose_trials_param(self):
        cmd = 'echo y | keentune sensitize train --data sensitize1 --output sensitize1'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        pattern = re.compile(r'trials:\s*(\d+)')
        trials = re.search(pattern, self.out).group(1)
        self.assertEqual(trials, "1")
        self.check_train_job()

    def test_sensitize_train_RBT_lose_trials_value(self):
        cmd = 'echo y | keentune sensitize train --data sensitize1 --output sensitize1 --trials'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --trials'))

    def test_sensitize_train_RBT_trials_value_error(self):
        cmd = 'echo y | keentune sensitize train --data sensitize1 --output sensitize1 --trials -o'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))

    def test_sensitize_train_RBT_trials_value_null(self):
        cmd = "echo y | keentune sensitize train --data sensitize1 --output sensitize1 --trials ''"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))
    
    def test_sensitize_train_RBT_trials_value_empty(self):
        cmd = "echo y | keentune sensitize train --data sensitize1 --output sensitize1 --trials ' '"
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('invalid argument'))
