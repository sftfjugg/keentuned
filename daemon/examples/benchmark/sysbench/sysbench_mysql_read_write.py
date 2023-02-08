# -*- coding: utf-8 -*-
import re
import sys
import subprocess
import time
import logging

"""
sysbench benchmark
"""

logger = logging.getLogger(__name__)

# const
MYSQL_PORT = 3306
MYSQL_USER = 'sysbench'
MYSQL_PASSWORD = "password"
DATABASE_NAME = "sysbenchdb"
REPORT_INTERVAL = 0
TIME = 20
PORT = 22
TEST_TYPE = "oltp_read_write"

DEFAULT = "--thread-stack-size=32768 --table-size=100000 --tables=3 --threads=1"

script_file = "setupmysql.sh"

mysqlshell = """#!/bin/sh
db_name=sysbenchdb
db_password=password
db_port=3306
db_user=sysbench
lib_dir=/var/lib/mysql/
log_dir=/var/log/
mysql_log=$log_dir/mysql/mysqld.log
tmp_sql=/tmp/mysql_setup.sql

systemctl stop mysqld

rm -rf $lib_dir
mkdir -p $log_dir/mysql
chmod 777 -R $log_dir
touch $mysql_log
chown -R mysql:mysql $mysql_log

#set the user mysql, permission and time zone
cat > $tmp_sql << EOF
use mysql;
update user set user.Host='%' where user.User='root';
FLUSH PRIVILEGES;
CREATE DATABASE IF NOT EXISTS ${db_name};
create user ${db_user}@'%' identified by '${db_password}';
grant all privileges on ${db_name}.* to ${db_user}@'%';
EOF

#check mysql version
mysql_version=$(mysql -V|awk '{print $3}')
if [[ $mysql_version > \"8.0\" ]];then
    mysqld --initialize --user=mysql --datadir=/var/lib/mysql
    systemctl restart mysqld
    raw_pswd=$(cat $mysql_log | grep 'temporary password' | tail -n 1 | awk '{print $NF}')
    set_pswd=\"ALTER USER 'root'@'localhost' IDENTIFIED BY '${db_password}';\"
    echo \"$set_pswd\"
    mysql -uroot -p\"${raw_pswd}\" -Dmysql --connect-expired-password -e \"$set_pswd\"
else
    echo \"mysql version error,check yum source!\"
    exit 1
fi
mysql -uroot -p\"${db_password}\" -Dmysql < $tmp_sql
"""


class Benchmark:
    def __init__(self, url):
        """Init benchmark

        Args:
            url (string): url
            password (string): MySQL database password.
            db_name (int, optional):  table name.
            tbl_size (int, optional): The maximum size of table. Defaults to 10000.
            tbl_nums (int, optional): The maximum numbers of table. Defaults to 10.
            rep_interval (int, optional): The rep_interval of table. Defaults to 10.
            threads (int, optional): The numbers of thread. Defaults to 40.
            times (int, optional): The time. Defaults to 10.
        """
        CMD_LOGIN = 'sysbench {} --mysql-host={} --mysql-port={} --mysql-user={} --mysql-password={} --db-driver=mysql --mysql-db={} --report-interval={} --time={}'.format(TEST_TYPE, str(url), MYSQL_PORT, MYSQL_USER, MYSQL_PASSWORD, DATABASE_NAME, REPORT_INTERVAL, TIME)
        CMD = ' '.join((CMD_LOGIN, DEFAULT))

        self.CMD_RUN = ' '.join((CMD,'run'))
        self.CMD_CLEAN = ' '.join((CMD,'cleanup'))
        self.CMD_PREPARE = ' '.join((CMD,'prepare'))

    def prepare(self):
        with open(script_file, "w", encoding='UTF-8') as f:
            f.write(mysqlshell)
        result = subprocess.run("sh {}".format(script_file), shell=True, close_fds=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        if result.returncode != 0:
            logger.error("setup database failed")
            return False

        cmd = self.CMD_CLEAN
        result = subprocess.run(cmd, shell=True, close_fds=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        if result.returncode == 0:
            cmd = self.CMD_PREPARE
            result = subprocess.run(cmd, shell=True, close_fds=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
            if result.returncode == 0:
                logger.info("prepare database successfully")
                return True
            else:
                logger.error("prepare database failed")
                return False
        else:
            logger.error("cleanup database failed")
            return False

    def run(self):
        """Run benchmark and parse output

        Return True and score list if running benchmark successfully, otherwise return False and empty list.
        """
        if not self.prepare():
            logger.error("prepare database failed")
            return False, []

        cmd = self.CMD_RUN
        result = subprocess.run(
            cmd,
            shell=True,
            close_fds=True,
            stderr=subprocess.PIPE,
            stdout=subprocess.PIPE
        )
        self.out = result.stdout.decode('UTF-8', 'strict')
        self.error = result.stderr.decode('UTF-8', 'strict')
        if result.returncode == 0:
            logger.info(self.out)
            pattern_static_trans = re.compile(r'transactions:\s+[\d.]+\s+\(([\d.]+)')
            pattern_static_queries = re.compile(r'queries:\s+[\d.]+\s+\(([\d.]+)')
            pattern_through_eps = re.compile(r'events\s\(avg/stddev\):\s+([\d.]+)')
            pattern_latency_avg = re.compile(r'avg:\s+([\d.]+)')
            pattern_latency_pct = re.compile(r'95th percentile:\s+([\d.]+)')

            if not re.search(pattern_static_trans, self.out) \
                    or not re.search(pattern_static_queries, self.out) \
                    or not re.search(pattern_through_eps, self.out) \
                    or not re.search(pattern_latency_avg, self.out) \
                    or not re.search(pattern_latency_pct, self.out):
                logger.error("can not parse output: {}".format(self.out))
                return False, []

            transactions = float(re.search(pattern_static_trans, self.out).group(1))
            queries = float(re.search(pattern_static_queries, self.out).group(1))
            events = float(re.search(pattern_through_eps, self.out).group(1))
            avg = float(re.search(pattern_latency_avg, self.out).group(1))
            percentile = float(re.search(pattern_latency_pct, self.out).group(1))

            result = {
                "TPS": transactions,
                "QPS": queries,
                "EPS": events,
                "Latency_avg": avg,
                "Latency_pct": percentile
            }

            result_str = ", ".join(["{} = {}".format(k, v) for k, v in result.items()])
            print(result_str)
            #clean up
            cmd = self.CMD_CLEAN
            result = subprocess.run(cmd, shell=True, close_fds=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)

            return True, result_str
        else:
            logger.error(self.error)
            return False, []


if __name__ == "__main__":
    if sys.argv.__len__() <= 1:
        print("'Target ip' is wanted: python3 sysbench_mysql_read_write.py [Target ip]")
        exit(1)
    bench = Benchmark(sys.argv[1])
    suc, res = bench.run()



