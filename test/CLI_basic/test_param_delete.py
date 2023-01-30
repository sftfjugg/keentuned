import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import runParamTune
from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestParamDelete(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_param_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_delete testcase finished')

    def test_param_delete_FUN(self):
        cmd = "echo y | keentune param delete --job param1"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))

        cmd = "keentune param jobs"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertFalse(self.out.__contains__('param1'))

        path = "/var/keentune/tuning_workspace/param1"
        res = os.path.exists(path)
        self.assertFalse(res)
