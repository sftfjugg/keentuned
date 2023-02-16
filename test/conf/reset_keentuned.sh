#!/bin/bash
scene=$1
json_name=$2


keentuned_conf_path=/etc/keentune/conf/keentuned.conf

clear_keentune_env()
{
    ps -ef|grep -w 'keentuned'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
    sleep 5
}

restart_keentuned()
{
    if [ "$scene" == "param" ];then
        sed -i "s/PARAMETER.*=.*/PARAMETER = ${json_name}/g" $keentuned_conf_path
    else
        sed -i "s/BENCH_CONFIG.*=.*/BENCH_CONFIG = ${json_name}/g" $keentuned_conf_path
    fi
    
    #keentuned > keentuned-multi_target.log 2>&1 &
    sleep 5
}

#clear_keentune_env
restart_keentuned
echo "restart keentuned server successfully!"
