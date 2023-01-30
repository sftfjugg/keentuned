import os
import re
import sys
import logging
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import checkServerStatus
from common import sysCommand

logger = logging.getLogger(__name__)


class TestTunedProfile(unittest.TestCase):
    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_tuned_profile testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('the test_tuned_profile testcase finished')

    def get_server_ip(self):
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        target = re.search(r"target_ip=\"(.*)\"", data).group(1)
        return target

    def get_cmd_res(self, cmd):
        target_ip = self.get_server_ip()
        if target_ip != "localhost":
            cmd = "ssh {} '{}'".format(target_ip, cmd)
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        return self.out.strip('\n').replace("\t", " ")

    def set_tuned_profile(self, profile_name):
        cmd = 'keentune profile set {}'.format(profile_name)
        self.status, _, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)

    def test_ecs_guest_RBT(self):
        self.set_tuned_profile("ecs-guest.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "40")

    def test_ecs_performance_RBT(self):
        self.set_tuned_profile("ecs-performance.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "40")

    def test_accelerator_performance_RBT(self):
        self.set_tuned_profile("accelerator-performance.conf")
        res = self.get_cmd_res("cat /proc/sys/kernel/sched_min_granularity_ns")
        self.assertEqual(res, "10000000")

    def test_atomic_guest_RBT(self):
        self.set_tuned_profile("atomic-guest.conf")
        res = self.get_cmd_res("sysctl -n kernel.pid_max")
        self.assertEqual(res, "131072")

    def test_atomic_host_RBT(self):
        self.set_tuned_profile("atomic-host.conf")
        res = self.get_cmd_res("sysctl -n fs.inotify.max_user_watches")
        self.assertEqual(res, "65536")

    def test_cpu_partitioning_RBT(self):
        self.set_tuned_profile("cpu-partitioning.conf")
        res = self.get_cmd_res("cat /etc/sysconfig/irqbalance")
        self.assertIn("banned_cpus = 5", res)
        res = self.get_cmd_res("cat /etc/systemd/system.conf")
        self.assertIn("CPUAffinity = 0,1,2,3,4,6,7", res)
        res = self.get_cmd_res("cat /sys/bus/workqueue/devices/writeback/cpumask")
        self.assertEqual(res, "df")

    def test_desktop_RBT(self):
        self.set_tuned_profile("desktop.conf")
        res = self.get_cmd_res("sysctl -n kernel.sched_autogroup_enabled")
        self.assertEqual(res, "1")

    def test_hpc_compute_RBT(self):
        self.set_tuned_profile("hpc-compute.conf")
        res = self.get_cmd_res("sysctl -n vm.min_free_kbytes")
        self.assertEqual(res, "135168")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[always]", res)

    def test_latency_performance_RBT(self):
        self.set_tuned_profile("latency-performance.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "10")

    def test_mssql_RBT(self):
        self.set_tuned_profile("mssql.conf")
        res = self.get_cmd_res("cat /proc/sys/kernel/sched_latency_ns")
        self.assertEqual(res, "60000000")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[always]", res)

    def test_network_latency_RBT(self):
        self.set_tuned_profile("network-latency.conf")
        res = self.get_cmd_res("sysctl -n net.core.busy_poll")
        self.assertEqual(res, "50")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[never]", res)

    def test_network_throughput_RBT(self):
        self.set_tuned_profile("network-throughput.conf")
        res = self.get_cmd_res("sysctl -n net.ipv4.tcp_rmem").replace("\t", " ")
        self.assertEqual(res, "4096 87380 16777216")

    def test_optimize_serial_RBT(self):
        self.set_tuned_profile("optimize-serial-console.conf")
        res = self.get_cmd_res("sysctl -n kernel.printk").replace("\t", " ")
        self.assertEqual(res, "4 4 1 7")
    
    def test_oracle_RBT(self):
        self.set_tuned_profile("oracle.conf")
        res = self.get_cmd_res("sysctl -n kernel.shmmni")
        self.assertEqual(res, "4096")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[never]", res)

    def test_postgresql_RBT(self):
        self.set_tuned_profile("postgresql.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_bytes")
        self.assertEqual(res, "536870912")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[never]", res)
        res = self.get_cmd_res("cat /proc/sys/kernel/sched_min_granularity_ns")
        self.assertEqual(res, "10000000")
    
    def test_realtime_RBT(self):
        self.set_tuned_profile("realtime.conf")
        res = self.get_cmd_res("cat /etc/sysconfig/irqbalance")
        self.assertIn("banned_cpus = 5", res)
        res = self.get_cmd_res("cat /sys/bus/workqueue/devices/writeback/cpumask")
        self.assertEqual(res, "df")

    def test_spindown_disk_RBT(self):
        self.set_tuned_profile("spindown-disk.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_expire_centisecs")
        self.assertEqual(res, "9000")

    def test_throughput_performance_RBT(self):
        self.set_tuned_profile("throughput-performance.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "40")
    
    def test_virtual_guest_RBT(self):
        self.set_tuned_profile("virtual-guest.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "30")

    def test_virtual_host_RBT(self):
        self.set_tuned_profile("virtual-host.conf")
        res = self.get_cmd_res("sysctl -n vm.dirty_background_ratio")
        self.assertEqual(res, "5")

