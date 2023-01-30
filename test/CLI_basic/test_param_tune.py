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

logger = logging.getLogger(__name__)


class TestParamTune(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_param_tune testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_param_tune testcase finished')

    def check_result(self):
        cmd = 'keentune param jobs'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("param1\ttpe\t10\tfinish", self.out)

        path = "/var/keentune/tuning_workspace/param1/param1_group1_best.json"
        res = os.path.exists(path)
        self.assertTrue(res)

    def test_param_tune_FUN(self):
        cmd = 'keentune param tune -i 10 --job param1'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if '[BEST] Tuning improvement' in res_data or "[ERROR]" in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4",
                     "Step5", "Step6", "[BEST] Tuning improvement"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

        self.check_result()
