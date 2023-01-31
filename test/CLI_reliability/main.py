import os
import sys
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from CLI_reliability.test_param_tune import TestParamTune
from CLI_reliability.test_param_dump import TestParamDump
from CLI_reliability.test_param_delete import TestParamDelete
from CLI_reliability.test_profile_info import TestProfileInfo
from CLI_reliability.test_profile_set import TestProfileSet
from CLI_reliability.test_profile_delete import TestProfileDelete
from CLI_reliability.test_profile_generate import TestProfileGenerate
from CLI_reliability.test_sensitize_train import TestSensitizeTrain
from CLI_reliability.test_sensitize_delete import TestSensitizeDelete
from CLI_reliability.test_param_tune_rollback import TestParamTuneRollback
from CLI_reliability.test_param_tune_delete import TestParamTuneDelete
from CLI_reliability.test_param_tune_dump import TestParamTuneDump
from CLI_reliability.test_profile_set_rollback import TestProfileSetRollback
from CLI_reliability.test_param_profile_rollback import TestParamProfileRollback
from CLI_reliability.test_sensitize_param import TestSensitizeParam
from CLI_reliability.test_sensitize_profile import TestSensitizeProfile
from CLI_reliability.test_param_tune_stop import TestParamTuneStop
from CLI_reliability.test_multi_scenes import TestMultiScenes
from CLI_reliability.test_multi_target import TestMultiTarget
from CLI_reliability.test_keentune_init import TestKeentuneInit
from CLI_reliability.test_tuned_profile import TestTunedProfile


def RunReliabilityCase():
    param_suite = unittest.TestSuite()
    param_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamTune))
    param_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamDump))
    param_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamDelete))

    profile_suite = unittest.TestSuite()
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileInfo))
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileSet))
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileDelete))
    profile_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileGenerate))

    sensitize_suite = unittest.TestSuite()
    sensitize_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestSensitizeTrain))
    sensitize_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestSensitizeDelete))

    combination_suite = unittest.TestSuite()
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamTuneRollback))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamTuneDelete))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamTuneDump))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestProfileSetRollback))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamProfileRollback))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestSensitizeParam))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestSensitizeProfile))
    combination_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestParamTuneStop))

    multi_scenes = unittest.TestSuite()
    multi_scenes.addTests(unittest.TestLoader().loadTestsFromTestCase(TestMultiScenes))

    multi_target = unittest.TestSuite()
    multi_target.addTests(unittest.TestLoader().loadTestsFromTestCase(TestMultiTarget))

    multi_init = unittest.TestSuite()
    multi_init.addTests(unittest.TestLoader().loadTestsFromTestCase(TestKeentuneInit))

    tuned_suite = unittest.TestSuite()
    tuned_suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestTunedProfile))

    suite = unittest.TestSuite([param_suite, profile_suite, sensitize_suite, combination_suite, tuned_suite, multi_scenes, multi_target, multi_init])
    return suite


if __name__ == '__main__':
    os.chdir(os.path.abspath(os.path.join(os.getcwd(), "..")))
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunReliabilityCase())

