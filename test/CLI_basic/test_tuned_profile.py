import os
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

    def test_ecs_guest_FUN(self):
        cmd = 'keentune profile set ecs-guest.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)

    def test_ecs_performance_FUN(self):
        cmd = 'keentune profile set ecs-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)

    def test_accelerator_performance_FUN(self):
        cmd = 'keentune profile set accelerator-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)
        self.assertIn("[scheduler]\t2 Succeeded", self.out)

    def test_atomic_guest_FUN(self):
        cmd = 'keentune profile set atomic-guest.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t5 Succeeded", self.out)

    def test_atomic_host_FUN(self):
        cmd = 'keentune profile set atomic-host.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t5 Succeeded", self.out)

    def test_balanced_FUN(self):
        cmd = 'keentune profile set balanced.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[cpu]\t0 Succeeded, 1 Failed", self.out)

    def test_cpu_partitioning_FUN(self):
        cmd = 'keentune profile set cpu-partitioning.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[env]\t1 Succeeded", self.out)
        self.assertIn("[sysfs]\t4 Succeeded", self.out)
        self.assertIn("[vm]\t1 Succeeded", self.out)
        self.assertIn("[systemd]\t1 Succeeded", self.out)
        self.assertIn("[irqbalance]\t1 Succeeded", self.out)

    def test_desktop_powersave_FUN(self):
        cmd = 'keentune profile set desktop-powersave.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("All of the domain backup failed", self.out)

    def test_desktop_FUN(self):
        cmd = 'keentune profile set desktop.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t1 Succeeded", self.out)

    def test_hpc_compute_FUN(self):
        cmd = 'keentune profile set hpc-compute.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm]\t1 Succeeded", self.out)
        self.assertIn("[sysctl]\t9 Succeeded", self.out)

    def test_intel_sst_FUN(self):
        cmd = 'keentune profile set intel-sst.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("No valid domain can be used", self.out)

    def test_latency_performance_FUN(self):
        cmd = 'keentune profile set latency-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)

    def test_mssql_FUN(self):
        cmd = 'keentune profile set mssql.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm]\t1 Succeeded", self.out)
        self.assertIn("[scheduler]\t4 Succeeded", self.out)

    def test_network_latency_FUN(self):
        cmd = 'keentune profile set network-latency.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm]\t1 Succeeded", self.out)

    def test_network_throughput_FUN(self):
        cmd = 'keentune profile set network-throughput.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t5 Succeeded", self.out)

    def test_optimize_serial_FUN(self):
        cmd = 'keentune profile set optimize-serial-console.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t1 Succeeded", self.out)
    
    def test_oracle_FUN(self):
        cmd = 'keentune profile set oracle.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm]\t1 Succeeded", self.out)
        self.assertIn("[sysctl]\t18 Succeeded", self.out)

    def test_postgresql_FUN(self):
        cmd = 'keentune profile set postgresql.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[vm]\t1 Succeeded", self.out)
        self.assertIn("[scheduler]\t2 Succeeded", self.out)
    
    def test_powersave_FUN(self):
        cmd = 'keentune profile set powersave.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("All of the domain backup failed", self.out)
    
    def test_realtime_FUN(self):
        cmd = 'keentune profile set realtime.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[env]\t1 Succeeded", self.out)
        self.assertIn("[sysfs]\t4 Succeeded", self.out)
        self.assertIn("[irqbalance]\t1 Succeeded", self.out)

    def test_server_powersave_FUN(self):
        cmd = 'keentune profile set server-powersave.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("All of the domain backup failed", self.out)

    def test_spindown_disk_FUN(self):
        cmd = 'keentune profile set spindown-disk.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[disk]\t2 Succeeded", self.out)
        self.assertIn("[env]\t1 Succeeded", self.out)

    def test_throughput_performance_FUN(self):
        cmd = 'keentune profile set throughput-performance.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)
    
    def test_virtual_guest_FUN(self):
        cmd = 'keentune profile set virtual-guest.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)

    def test_virtual_host_FUN(self):
        cmd = 'keentune profile set virtual-host.conf'
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("[sysctl]\t3 Succeeded", self.out)

