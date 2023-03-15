import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand
from common import getSysBackupData
from common import checkBackupData
from common import runParamTune

logger = logging.getLogger(__name__)


class TestParamTuneRollback(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        print("server status is: %s" % status)
        self.assertEqual(status, 0)
        cmd = "ps -ef|grep keentune"
        self.status, self.out, _  = sysCommand(cmd)
        print("keentune status is: %s" % self.status)
        print("keentune output is: %s" % self.out)
        print("keentune error is: %s" % _)
        logger.info('start to run test_param_tune_rollback testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        cmd = "ps -ef|grep keentune"
        self.status, self.out, _  = sysCommand(cmd)
        print("tearDown status is: %s" % self.status)
        print("tearDown output is: %s" % self.out)
        print("tearDown error is: %s" % _)
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_tune_rollback testcase finished')

    def test_param_tune_RBT_stop_rollback(self):
        getSysBackupData()
        cmd = "ps -ef|grep keentune"
        self.status, self.out, _  = sysCommand(cmd)
        print("keentune status is: %s" % self.status)
        print("keentune output is: %s" % self.out)
        print("keentune error is: %s" % _)
        cmd = 'keentune param tune -i 10 --job param1'
        self.status, self.out, _  = sysCommand(cmd)
        print("status is: %s" % self.status)
        print("output is: %s" % self.out)
        print("error is: %s" % _)
        self.assertEqual(self.status, 0)
        time.sleep(2)
        cmd = 'keentune param stop'
        print("------------------------------------------------")
        self.status, self.out, _ = sysCommand(cmd)
        print("status is: %s" % self.status)
        print("output is: %s" % self.out)
        print("error is: %s" % _)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Abort parameter optimization job'))
        time.sleep(5)
        res = checkBackupData()
        self.assertEqual(res, 0)
        deleteDependentData("param1")

        status = runParamTune("param1")
        self.assertEqual(status, 0)
        res = checkBackupData()
        self.assertEqual(res, 1)

        self.status, self.out, _ = sysCommand('keentune param rollback')
        print("status is: %s" % self.status)
        print("output is: %s" % self.out)
        print("error is: %s" % _)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('param rollback successfully'))
        res = checkBackupData()
        self.assertEqual(res, 0)
