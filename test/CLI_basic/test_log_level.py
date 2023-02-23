import logging
import unittest
from configparser import ConfigParser
from common import checkServerStatus

logger = logging.getLogger(__name__)


class TestLogLevel(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_log_level testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_log_level testcase finished')

    def test_log_level_FUN(self):
        conf = ConfigParser()
        conf.read("/etc/keentune/conf/keentuned.conf")
        log_level = conf['keentuned']['LOGFILE_LEVEL']
        self.assertEqual(log_level, "INFO")
