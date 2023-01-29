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

    def test_ecs_guest_FUN(self):
        cmd = 'keentune profile set ecs-guest.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "40")

    def test_ecs_performance_FUN(self):
        cmd = 'keentune profile set ecs-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "40")

    def test_accelerator_performance_FUN(self):
        cmd = 'keentune profile set accelerator-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        self.assertIn("[scheduler] 2 Succeeded", self.out)
        res = self.get_cmd_res("cat /proc/sys/kernel/sched_min_granularity_ns")
        self.assertEqual(res, "10000000")

    def test_atomic_guest_FUN(self):
        cmd = 'keentune profile set atomic-guest.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 5 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n kernel.pid_max")
        self.assertEqual(res, "131072")

    def test_atomic_host_FUN(self):
        cmd = 'keentune profile set atomic-host.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 5 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n fs.inotify.max_user_watches")
        self.assertEqual(res, "65536")

    def test_balanced_FUN(self):
        cmd = 'keentune profile set balanced.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertIn("[cpu] 0 Succeeded, 1 Failed", self.out)

    def test_cpu_partitioning_FUN(self):
        cmd = 'keentune profile set cpu-partitioning.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysfs] 4 Succeeded", self.out)
        self.assertIn("[vm] 1 Succeeded", self.out)
        self.assertIn("[systemd] 1 Succeeded", self.out)
        self.assertIn("[irqbalance] 1 Succeeded", self.out)
        res = self.get_cmd_res("cat /etc/sysconfig/irqbalance")
        self.assertIn("banned_cpus = 5", res)
        res = self.get_cmd_res("cat /etc/systemd/system.conf")
        self.assertIn("CPUAffinity = 0,1,2,3,4,6,7", res)
        res = self.get_cmd_res("cat /sys/bus/workqueue/devices/writeback/cpumask")
        self.assertEqual(res, "df")

    def test_desktop_powersave_FUN(self):
        cmd = 'keentune profile set desktop-powersave.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertIn("All of the domain backup failed", self.out)

    def test_desktop_FUN(self):
        cmd = 'keentune profile set desktop.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 1 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n kernel.sched_autogroup_enabled")
        self.assertEqual(res, "1")

    def test_hpc_compute_FUN(self):
        cmd = 'keentune profile set hpc-compute.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm] 1 Succeeded", self.out)
        self.assertIn("[sysctl] 9 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.min_free_kbytes")
        self.assertEqual(res, "135168")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[always]", res)

    def test_intel_sst_FUN(self):
        cmd = 'keentune profile set intel-sst.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertIn("No valid domain can be used", self.out)

    def test_latency_performance_FUN(self):
        cmd = 'keentune profile set latency-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "10")

    def test_mssql_FUN(self):
        cmd = 'keentune profile set mssql.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm] 1 Succeeded", self.out)
        self.assertIn("[scheduler] 4 Succeeded", self.out)
        res = self.get_cmd_res("cat /proc/sys/kernel/sched_latency_ns")
        self.assertEqual(res, "60000000")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[always]", res)

    def test_network_latency_FUN(self):
        cmd = 'keentune profile set network-latency.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm] 1 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n net.core.busy_poll")
        self.assertEqual(res, "50")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[never]", res)

    def test_network_throughput_FUN(self):
        cmd = 'keentune profile set network-throughput.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 5 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n net.ipv4.tcp_rmem").replace("\t", " ")
        self.assertEqual(res, "4096 87380 16777216")

    def test_optimize_serial_FUN(self):
        cmd = 'keentune profile set optimize-serial-console.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 1 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n kernel.printk").replace("\t", " ")
        self.assertEqual(res, "4 4 1 7")
    
    def test_oracle_FUN(self):
        cmd = 'keentune profile set oracle.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm] 1 Succeeded", self.out)
        self.assertIn("[sysctl] 18 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n kernel.shmmni")
        self.assertEqual(res, "4096")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[never]", res)

    def test_postgresql_FUN(self):
        cmd = 'keentune profile set postgresql.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm] 1 Succeeded", self.out)
        self.assertIn("[scheduler] 2 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_bytes")
        self.assertEqual(res, "536870912")
        res = self.get_cmd_res("cat /sys/kernel/mm/transparent_hugepage/enabled")
        self.assertIn("[never]", res)
        res = self.get_cmd_res("cat /proc/sys/kernel/sched_min_granularity_ns")
        self.assertEqual(res, "10000000")
    
    def test_powersave_FUN(self):
        cmd = 'keentune profile set powersave.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertIn("All of the domain backup failed", self.out)
    
    def test_realtime_FUN(self):
        cmd = 'keentune profile set realtime.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[env] 1 Succeeded", self.out)
        self.assertIn("[sysfs] 4 Succeeded", self.out)
        self.assertIn("[irqbalance] 1 Succeeded", self.out)
        res = self.get_cmd_res("cat /etc/sysconfig/irqbalance")
        self.assertIn("banned_cpus = 5", res)
        res = self.get_cmd_res("cat /sys/bus/workqueue/devices/writeback/cpumask")
        self.assertEqual(res, "df")

    def test_server_powersave_FUN(self):
        cmd = 'keentune profile set server-powersave.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 1)
        self.assertIn("All of the domain backup failed", self.out)

    def test_spindown_disk_FUN(self):
        cmd = 'keentune profile set spindown-disk.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 5 Succeeded", self.out)
        self.assertIn("[env] 1 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_expire_centisecs")
        self.assertEqual(res, "9000")

    def test_throughput_performance_FUN(self):
        cmd = 'keentune profile set throughput-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "40")
    
    def test_virtual_guest_FUN(self):
        cmd = 'keentune profile set virtual-guest.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_ratio")
        self.assertEqual(res, "30")

    def test_virtual_host_FUN(self):
        cmd = 'keentune profile set virtual-host.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl] 3 Succeeded", self.out)
        res = self.get_cmd_res("sysctl -n vm.dirty_background_ratio")
        self.assertEqual(res, "5")
