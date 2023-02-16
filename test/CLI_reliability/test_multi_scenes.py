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
from common import getTuneTaskResult
from common import getTaskLogPath
from common import runParamTune

logger = logging.getLogger(__name__)


class TestMultiScenes(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        self.target, self.bench = self.get_server_ip()
        self.keentune_path = "/etc/keentune/conf/keentuned.conf"
        if self.target != "localhost":
            status = sysCommand("scp conf/init_mysql.sh {}:/opt".format(self.target))[0]
            assert status == 0
            status = sysCommand("ssh {} 'sh /opt/init_mysql.sh'".format(self.target))[0]
            assert status == 0
            status = sysCommand("ssh {} 'nohup iperf3 -s > /dev/null 2>&1 &'".format(self.target))[0]
            assert status == 0
        else:
            status = sysCommand("nohup iperf3 -s > /dev/null 2>&1 &")[0]
            assert status == 0
        
    @classmethod
    def tearDownClass(self) -> None:
        cmd = "ps -ef|grep -E 'iperf3 -s'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}"
        if self.target != "localhost":
            status = sysCommand("ssh {} 'rm -rf /opt/init_mysql.sh'".format(self.target))[0]
            assert status == 0
            status = sysCommand("ssh {} '{}'".format(self.target, cmd))[0]
            assert status == 0
        else:
            status = sysCommand(cmd)[0]
            assert status == 0

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_multiple_scenes testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        logger.info('the test_multiple_scenes testcase finished')

    @staticmethod
    def get_server_ip():
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        target = re.search(r"target_ip=\"(.*)\"", data).group(1)
        bench = re.search(r"bench_ip=\"(.*)\"", data).group(1)
        return target, bench

    def check_param_tune_job(self, name):
        cmd = 'keentune param jobs'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__(name))

    def run_sensitize_train(self, name):
        cmd = "echo y | keentune sensitize train --data {0} --job {0}".format(name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if "identification results successfully" in res_data or "[ERROR]" in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "identification results successfully"]
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

        self.path = "/var/keentune/sensitize_workspace/{}/knobs.json".format(name)
        res = os.path.exists(self.path)
        self.assertTrue(res)

    def restart_brain_server(self, algorithm, flag):
        cmd = "sh conf/restart_brain.sh {} {}".format(algorithm, flag)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('restart brain server successfully!'))

    def reset_keentuned(self, config, file):
        cmd = "sh conf/reset_keentuned.sh {} {}".format(config, file)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("restart keentuned server successfully!", self.out)

    def run_param_tune(self):
        cmd = 'keentune param tune -i 10 --job param1'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("param1")

    def reset_default_conf(self):
        self.run_param_tune()
        time.sleep(10)
        self.reset_keentuned("param", "sysctl.json")
        self.reset_keentuned("bench", "wrk_http_long.json")

    def set_yitian_profile(self, profile_name):
        cmd = 'keentune profile set {}'.format(profile_name)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)

    def get_cmd_res(self, cmd):
        if self.target != "localhost":
            cmd = "ssh {} '{}'".format(self.target, cmd)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.out = self.out.strip('\n').replace("\t", " ")

    def change_target_ip(self, server):
        cmd = r'sed -i "s/TARGET_IP.*=.*/TARGET_IP = {}/g" {}'.format(server, self.keentune_path)
        self.assertEqual(sysCommand(cmd)[0], 0)

    def test_param_domain_FUN_sysctl(self):
        self.reset_keentuned("param", "sysctl.json")
        self.run_param_tune()
    
    def test_param_domain_FUN_nginx(self):
        self.reset_keentuned("param", "nginx_conf.json")
        self.run_param_tune()
        self.reset_keentuned("param", "sysctl.json")

    def test_param_domain_FUN_mysql(self):
        self.reset_keentuned("param", "my_cnf.json")
        self.reset_keentuned("bench", "sysbench_mysql_read_write.json")
        self.reset_default_conf()

    def test_param_domain_FUN_sysbench(self):
        self.change_target_ip(self.bench)
        self.reset_keentuned("param", "sysbench.json")
        self.reset_keentuned("bench", "sysbench_mysql_read_write.json")
        self.reset_default_conf()
        self.change_target_ip(self.target)

    def test_param_domain_FUN_iperf(self):
        self.change_target_ip(self.bench)
        cmd = r'sed -i "s/\(.*\)8388608\(.*\)/\1300000\2/" /etc/keentune/parameter/iperf.json'
        self.assertEqual(sysCommand(cmd)[0], 0)
        self.reset_keentuned("param", "iperf.json")
        self.reset_keentuned("bench", "iperf_bench.json")
        self.reset_default_conf()
        self.change_target_ip(self.target)

    def test_param_domain_FUN_fio(self):
        self.change_target_ip(self.bench)
        self.reset_keentuned("param", "fio.json")
        self.reset_keentuned("bench", "bench_fio_disk_IOPS.json")
        self.reset_default_conf()
        self.change_target_ip(self.target)

    def test_param_domain_FUN_wrk(self):
        self.change_target_ip(self.bench)
        json_path = "/etc/keentune/parameter/wrk.json"
        cmd = r'sed -i "s/\(.*\)cpu_core#\*300\(.*\)/\1cpu_core#\*100\2/" {}'.format(json_path)
        self.assertEqual(sysCommand(cmd)[0], 0)
        cmd = r'sed -i "s/\(.*\)cpu_core#\*3\(.*\)/\1cpu_core#\2/" {}'.format(json_path)
        self.assertEqual(sysCommand(cmd)[0], 0)
        self.reset_keentuned("param", "wrk.json")
        self.reset_keentuned("bench", "wrk_parameter_tuning.json")
        self.reset_default_conf()
        self.change_target_ip(self.target)

    def test_profile_yitian_FUN_nginx(self):
        self.set_yitian_profile("nginx.conf")
        self.assertIn("[limits] 4 Succeeded", self.out)
        self.get_cmd_res("cat /proc/sys/fs/file-max")
        self.assertEqual(self.out, "10485760")

    def test_profile_yitian_FUN_mysql(self):
        self.set_yitian_profile("mysql.conf")
        self.assertIn("[net] 3 Succeeded", self.out)
        self.get_cmd_res("cat /sys/class/net/eth0/queues/tx-0/xps_cpus")
        self.assertNotEqual(int(self.out), 0)

    def test_profile_yitian_FUN_pgsql(self):
        self.set_yitian_profile("pgsql.conf")
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(self.out, "40")

    def test_profile_yitian_FUN_redis(self):
        self.set_yitian_profile("redis.conf")
        self.assertIn("[net] 3 Succeeded", self.out)
        self.get_cmd_res("cat /sys/class/net/eth0/queues/tx-0/xps_cpus")
        self.assertNotEqual(int(self.out), 0)

    def test_param_tune_FUN_tpe(self):
        self.restart_brain_server("tpe", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_hord(self):
        self.restart_brain_server("hord", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_random(self):
        self.restart_brain_server("random", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_lamcts(self):
        self.restart_brain_server("lamcts", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)

    def test_param_tune_FUN_bgcs(self):
        self.restart_brain_server("bgcs", "tune")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
    
    def test_sensitize_train_FUN_lasso(self):
        self.restart_brain_server("lasso", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_univariate(self):
        self.restart_brain_server("univariate", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_gp(self):
        self.restart_brain_server("gp", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_shap(self):
        self.restart_brain_server("shap", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")

    def test_sensitize_train_FUN_explain(self):
        self.restart_brain_server("explain", "train")
        status = runParamTune("param1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("param1")



