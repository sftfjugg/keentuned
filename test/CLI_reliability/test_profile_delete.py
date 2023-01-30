import os
import sys
import logging
import subprocess
import unittest

logger = logging.getLogger(__name__)
sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import checkServerStatus
from common import deleteDependentData
from common import runParamTune
from common import runParamDump
from common import runProfileSet

class TestProfileDelete(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestProfileDelete begin...")
        status = runParamTune("param1")
        assert status == 0
        status = runParamDump("param1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestProfileDelete end...") 

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_delete testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_profile_delete testcase finished')

    def test_profile_delete_RBT_lose_name_param(self):
        cmd = 'keentune profile delete'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_delete_RBT_lose_name_value(self):
        cmd = 'keentune profile delete --name'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --name'))

    def test_profile_delete_RBT_name_value_null(self):
        cmd = "keentune profile delete --name ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_delete_RBT_name_value_empty(self):
        cmd = "keentune profile delete --name ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('Incomplete or Unmatched command'))

    def test_profile_delete_RBT_before_rollback(self):
        self.status = runProfileSet()
        self.assertEqual(self.status, 0)
        cmd = "echo y | keentune profile delete --name param1_group1.conf"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.out.__contains__('param1_group1.conf is active profile'))

        self.status = sysCommand("keentune profile rollback")[0]
        self.assertEqual(self.status, 0)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('delete successfully'))
