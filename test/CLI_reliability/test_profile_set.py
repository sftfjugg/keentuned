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

class TestProfileSet(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        logger.info("TestProfileSet begin...")
        status = runParamTune("param1")
        assert status == 0
        status = runParamDump("param1")
        assert status == 0

    @classmethod
    def tearDownClass(self) -> None:
        deleteDependentData("param1")
        logger.info("TestProfileSet end...") 

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_set testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_profile_set testcase finished')

    def test_profile_set_RBT_lose_name_param(self):
        cmd = 'keentune profile set'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('found group is null'))

    def test_profile_set_RBT_lose_name_value(self):
        cmd = 'keentune profile set --group1'
        self.status, _, self.error = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertTrue(self.error.__contains__('flag needs an argument: --group1'))

    def test_profile_set_RBT_name_value_null(self):
        cmd = "keentune profile set --group1 ''"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('found group is null'))

    def test_profile_set_RBT_name_value_empty(self):
        cmd = "keentune profile set --group1 ' '"
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('found group is null'))

    def test_profile_set_RBT_full_path(self):
        path = "/var/keentune/profile/param1_group1.conf"
        cmd = "keentune profile set --group1 {}".format(path)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Set /var/keentune/profile/param1_group1.conf successfully'))
