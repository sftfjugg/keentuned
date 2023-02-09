#!/bin/bash

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
if [[ $mysql_version > "8.0" ]];then
    mysqld --initialize --user=mysql --datadir=/var/lib/mysql
    systemctl restart mysqld
    raw_pswd=$(cat $mysql_log | grep 'temporary password' | tail -n 1 | awk '{print $NF}')
    set_pswd="ALTER USER 'root'@'localhost' IDENTIFIED BY '${db_password}';"
    echo "$set_pswd"
    mysql -uroot -p"${raw_pswd}" -Dmysql --connect-expired-password -e "$set_pswd"
else
    echo "mysql version error,check yum source!"
    exit 1
fi
mysql -uroot -p"${db_password}" -Dmysql < $tmp_sql
