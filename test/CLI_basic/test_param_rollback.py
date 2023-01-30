import os
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import checkServerStatus
from common import sysCommand
from common import runParamTune
from common import deleteDependentData
from common import getSysBackupData
from common import checkBackupData

logger = logging.getLogger(__name__)


class TestParamRollback(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        getSysBackupData()
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        logger.info('start to run test_param_rollback testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_rollback testcase finished')

    def test_param_rollback_FUN(self):
        res = checkBackupData()
        self.assertEqual(res, 1)
        cmd = 'keentune param rollback'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('param rollback successfully'))

        res = checkBackupData()
        self.assertEqual(res, 0)
