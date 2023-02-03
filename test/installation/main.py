import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))
os.chdir(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from installation.test_install_source import TestInstallSource
from installation.test_install_yum import TestInstallYum


def RunLongStableCase():
    suite = unittest.TestSuite()
    suite.addTest(TestInstallSource("test_install_source_FUN"))
    suite.addTest(TestInstallYum("test_install_yum_FUN"))
    suite.addTest(TestInstallYum("test_instatll_yum_RBT_start"))
    suite.addTest(TestInstallYum("test_instatll_yum_RBT_stop"))
    return suite


if __name__ == '__main__':
    print("--------------- start to run install test cases ---------------")
    deleteDependentData("param1")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunLongStableCase())
    print("--------------- run install test cases end ---------------")
