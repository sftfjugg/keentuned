import os
import sys
import unittest

from common import deleteDependentData
from CLI_basic.main import RunBasicCase
from CLI_reliability.main import RunReliabilityCase
from MT_restful.main import RunModelCase


def RunAllCase():
    basic_suite = RunBasicCase()
    reliability_suite = RunReliabilityCase()
    model_suite = RunModelCase()
    suite = unittest.TestSuite([basic_suite, reliability_suite, model_suite])
    return suite


if __name__ == '__main__':
    print("--------------- start to run test cases ---------------")
    deleteDependentData("param1")
    deleteDependentData("sensitize1")
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(RunAllCase())
    print("--------------- run test cases end ---------------")

