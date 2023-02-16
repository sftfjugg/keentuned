import os
import sys
import re
import unittest
import logging

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import runParamDump
from common import runParamTune
from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand
from common import getSysBackupData
from common import checkBackupData

logger = logging.getLogger(__name__)


class TestProfileSet(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        status = runParamDump("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_profile_set testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_profile_set testcase finished')

    def test_profile_set_FUN(self):
        getSysBackupData()
        cmd = 'keentune profile set --group1 param1_group1.conf'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('Succeeded'))

        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.result = re.search(r'\[(.*?)\].+param1_group1.conf', self.out).group(1)
        self.assertTrue(self.result.__contains__('active'))

        res = checkBackupData()
        self.assertEqual(res, 1)
