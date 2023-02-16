import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from CLI_basic.test_help import TestHelp
from CLI_basic.test_init import TestInit
from CLI_basic.test_version import TestVersion
from CLI_basic.test_rollback_all import TestRollbackAll
from CLI_basic.test_param_tune import TestParamTune
from CLI_basic.test_param_list import TestParamList
from CLI_basic.test_param_dump import TestParamDump
from CLI_basic.test_param_rollback import TestParamRollback
from CLI_basic.test_param_delete import TestParamDelete
from CLI_basic.test_profile_delete import TestProfileDelete
from CLI_basic.test_profile_generate import TestProfileGenerate
from CLI_basic.test_profile_info import TestProfileInfo
from CLI_basic.test_profile_list import TestProfileList
from CLI_basic.test_profile_rollback import TestProfileRollback
from CLI_basic.test_profile_set import TestProfileSet
from CLI_basic.test_sensitize_train import TestSensitizeTrain
from CLI_basic.test_sensitize_jobs import TestSensitizeJobs
from CLI_basic.test_sensitize_delete import TestSensitizeDelete
from CLI_basic.test_tuned_profile import TestTunedProfile


def RunBasicCase():
    param_suite = unittest.TestSuite()
    param_suite.addTest(TestRollbackAll('test_rollback_all_FUN'))
    param_suite.addTest(TestHelp('test_help_FUN'))
    param_suite.addTest(TestInit('test_init_FUN'))
    param_suite.addTest(TestVersion('test_version_FUN'))
    param_suite.addTest(TestParamTune('test_param_tune_FUN'))
    param_suite.addTest(TestParamList('test_param_list_FUN'))
    param_suite.addTest(TestParamDump('test_param_dump_FUN'))
    param_suite.addTest(TestParamRollback('test_param_rollback_FUN'))
    param_suite.addTest(TestParamDelete('test_param_delete_FUN'))

    profile_suite = unittest.TestSuite()
    profile_suite.addTest(TestProfileList('test_profile_list_FUN'))
    profile_suite.addTest(TestProfileInfo('test_profile_info_FUN'))
    profile_suite.addTest(TestProfileSet('test_profile_set_FUN'))
    profile_suite.addTest(TestProfileGenerate('test_profile_generate_FUN'))
    profile_suite.addTest(TestProfileRollback('test_profile_rollback_FUN'))
    profile_suite.addTest(TestProfileDelete('test_profile_delete_FUN'))

    sensitize_suite = unittest.TestSuite()
    sensitize_suite.addTest(TestSensitizeTrain('test_sensitize_train_FUN'))
    sensitize_suite.addTest(TestSensitizeJobs('test_sensitize_jobs_FUN'))
    sensitize_suite.addTest(TestSensitizeDelete('test_sensitize_delete_FUN'))
    
    tuned_suite = unittest.TestSuite()
    tuned_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestTunedProfile))

    suite = unittest.TestSuite([param_suite, profile_suite, sensitize_suite, tuned_suite])
    return suite


if __name__ == '__main__':
    os.chdir(os.path.abspath(os.path.join(os.getcwd(), "..")))
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunBasicCase())

