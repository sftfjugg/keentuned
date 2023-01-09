#!/bin/bash

keentune_code_dir=/tmp/keentune-cluster-tmp
test_conf_dir=$keentune_code_dir/acops-new/test/conf
target_server=$1
bench_server=$2
scene_cmd=$3

keentuned_path=/etc/keentune
keentuned_conf_path=/etc/keentune/conf/keentuned.conf
sysctl_path=$test_conf_dir/sysctl_target.json
local_keentuned_conf=$test_conf_dir/keentuned.conf
benchmark_script=$test_conf_dir/nginx_http_long_multi_target.py
json_path=$test_conf_dir/bench_wrk_nginx_long_multi_target.json

clear_keentune_env()
{
    ps -ef|grep -w 'keentuned'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
    sleep 5
}

restart_keentuned()
{
    cp $sysctl_path $keentuned_path/parameter
    cp $json_path $keentuned_path/benchmark/wrk
    cp $benchmark_script $keentuned_path/benchmark/wrk

    echo y | cp $local_keentuned_conf $keentuned_conf_path
    sed -i "s/BENCH_SRC_IP.*=.*/BENCH_SRC_IP = ${bench_server}/g" $keentuned_conf_path
    sed -i "s/BENCH_DEST_IP.*=.*/BENCH_DEST_IP = ${target_server}/g" $keentuned_conf_path
    sed -i "s/BENCH_CONFIG.*=.*/BENCH_CONFIG = bench_wrk_nginx_long_multi_target.json/g" $keentuned_conf_path
    echo -e $scene_cmd >> $keentuned_conf_path

    #keentuned > keentuned-multi_target.log 2>&1 &
    sleep 5
}

#clear_keentune_env
restart_keentuned
echo "restart keentuned server successfully!"
