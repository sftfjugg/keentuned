#!/bin/bash

keentune_code_dir=/tmp/keentune-cluster-tmp
test_conf_dir=$keentune_code_dir/acops-new/test/conf
target_server=$1
bench_server=$2
scene_cmd=$3

keentuned_conf_path=/etc/keentune/conf/keentuned.conf
local_keentuned_conf=$test_conf_dir/keentuned.conf


clear_keentune_env()
{
    ps -ef|grep -w 'keentuned'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
    sleep 5
}

restart_keentuned()
{
    echo y | cp $local_keentuned_conf $keentuned_conf_path
    sed -i "s/BENCH_SRC_IP.*=.*/BENCH_SRC_IP = ${bench_server}/g" $keentuned_conf_path
    sed -i "s/BENCH_DEST_IP.*=.*/BENCH_DEST_IP = ${target_server}/g" $keentuned_conf_path
    echo -e $scene_cmd >> $keentuned_conf_path

    #keentuned > keentuned-init.log 2>&1 &
    sleep 5
}

#clear_keentune_env
restart_keentuned
echo "init keentune successfully!"
