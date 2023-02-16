import os
import re
import sys
import time
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import sysCommand
from common import runParamTune
from common import getTaskLogPath
from common import checkServerStatus
from common import getTrainTaskResult

logger = logging.getLogger(__name__)


class TestLongStability(unittest.TestCase):
    job_name = "param_tune"
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_long_stability testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain", "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_long_stability testcase finished')

    def profile_list(self, state, profile_name):
        cmd = 'keentune profile list'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.result = re.search(r'\[(.*?)\].+{}'.format(profile_name), self.out).group(1)
        self.assertIn(state, self.result)

    def profile_dump(self, job_name):
        cmd = 'echo y | keentune param dump -j {}'.format(job_name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('dump successfully'))

    def profile_set(self,profile_name):
        cmd = 'keentune profile set {}'.format(profile_name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.profile_list("active", profile_name)

    def profile_rollback(self):
        cmd = 'keentune profile rollback'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('profile rollback successfully'))

    def test_long_stability_RBT(self):
        start_time = time.time()
        time_diff = 0
        while time_diff < self.time_limit:
            init_cmd = "keentune init"
            self.status, self.out, _ = sysCommand(init_cmd)
            self.assertEqual(self.status, 0)
            self.assertTrue(self.out.__contains__('KeenTune Init success'))

            all_cmd = "keentune rollbackall"
            self.status, self.out, _ = sysCommand(all_cmd)
            self.assertEqual(self.status, 0)
            self.assertTrue(self.out.__contains__('Rollback all successfully') or self.out.__contains__('All Targets No Need to Rollback'))

            for tune_algorithm in ("tpe", "hord", "random", "bgcs", "lamcts"):
                cmd = "echo y | keentune param delete --job {}".format(self.job_name)
                sysCommand(cmd)

                sed_cmd = 'sed -i "s/AUTO_TUNING_ALGORITHM\(.*\)=.*/AUTO_TUNING_ALGORITHM\\1= {}/" /etc/keentune/conf/keentuned.conf'.format(tune_algorithm)
                sysCommand(sed_cmd)
                result = runParamTune(self.job_name, iteration=500)
                self.assertEqual(result, 0)
                time.sleep(5)
            
            for algorithm in ('lasso', 'univariate', 'shap', 'explain', 'gp'):
                sed_cmd = 'sed -i "s/SENSITIZE_ALGORITHM\(.*\)=.*/SENSITIZE_ALGORITHM\1= {}/" /etc/keentune/conf/keentuned.conf'.format(algorithm)
                sysCommand(sed_cmd)

                cmd = "echo y | keentune sensitize train --data {0} --job {0} -t 10".format(self.job_name)
                path = getTaskLogPath(cmd)
                result = getTrainTaskResult(path)
                self.assertTrue(result)
                cmd = "echo y | keentune sensitize delete --job {}".format(self.job_name)
                sysCommand(cmd)
                time.sleep(2)

            self.profile_dump(self.job_name)
            profile_list = ["{}_group1.conf".format(self.job_name), "cpu_high_load.conf"]
            for profile_name in profile_list:
                for i in range(100):
                    self.profile_set(profile_name)
                    time.sleep(2)
                    self.profile_rollback()
                    time.sleep(2)

            cmd = "echo y | keentune param delete --job {}".format(self.job_name)
            sysCommand(cmd)
            cmd = 'echo y | keentune profile delete --name {}_group1.conf'.format(self.job_name)
            self.status, self.out, _ = sysCommand(cmd)
            self.assertEqual(self.status, 0)
            self.assertTrue(self.out.__contains__('delete successfully'))

            logger.info("current round testcase finished")
            time_diff = (time.time() - start_time) / 3600


