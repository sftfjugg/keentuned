import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import checkServerStatus
from common import sysCommand
from common import getSysBackupData
from common import checkBackupData

logger = logging.getLogger(__name__)


class TestRollbackAll(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        getSysBackupData()
        logger.info('start to run test_rollback_all testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-target"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_rollback_all testcase finished')

    def test_rollback_all_FUN(self):
        cmd = 'keentune rollbackall'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        res = self.out.__contains__('Rollback all successfully') or self.out.__contains__('All Targets No Need to Rollback')
        self.assertTrue(res)
        self.assertEqual(checkBackupData(), 0)
