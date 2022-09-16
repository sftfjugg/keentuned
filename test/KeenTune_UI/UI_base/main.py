import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from UI_base.test_experts_tuning_normal import TestKeenTuneUiNormal
from UI_base.test_smart_tuning_normal import TestKeenTuneUiSmartNormal


def RunBasicCase():
    profile_suite = unittest.TestSuite()
    profile_suite.addTest(TestKeenTuneUiNormal("test_copyfile"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_creatfile"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_checkfile"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_editor"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_set_group"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_restore"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_deletefile"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_language_switch"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_refresh"))
    profile_suite.addTest(TestKeenTuneUiNormal("test_set_list"))

    param_suite = unittest.TestSuite()
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_createjob"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_createjob02"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_detail"))
    #param_suite.addTest(TestKeenTuneUiSmartNormal("test_log"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_rerun"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_delete"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_refresh"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_setting"))
    param_suite.addTest(TestKeenTuneUiSmartNormal("test_sorting"))

    suite = unittest.TestSuite([profile_suite, param_suite])
    return suite


if __name__ == '__main__':
    if sys.argv.__len__() <= 1:
        print("'web_ip' is wanted: python3 main.py 127.0.0.1")
        exit(1)
    TestKeenTuneUiNormal.web_ip = sys.argv[1]
    TestKeenTuneUiSmartNormal.web_ip = sys.argv[1]
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunBasicCase())
