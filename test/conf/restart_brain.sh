#!/bin/bash
algorithm=$1
flag=$2

keentuned_conf=/etc/keentune/conf/keentuned.conf

clear_keentune_env()
{
    ps -ef|grep -w 'keentuned'|grep -v grep|awk '{print $2}'| xargs -I {} kill -9 {}
    sleep 5
}

restart_brain()
{
    if [[ $flag == "train" ]];then
        sed -i "s/SENSITIZE_ALGORITHM.*=.*/SENSITIZE_ALGORITHM = ${algorithm}/g" $keentuned_conf
    else
        sed -i "s/AUTO_TUNING_ALGORITHM.*=.*/AUTO_TUNING_ALGORITHM = ${algorithm}/g" $keentuned_conf
    fi
    #keentuned > keentune_algorithm.log 2>&1 &
    sleep 5
}

#clear_keentune_env
restart_brain
echo "restart brain server successfully!"

