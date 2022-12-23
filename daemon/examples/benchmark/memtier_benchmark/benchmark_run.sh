#!/bin/bash
instance=2
REDIS_PORT=9400
cpu_cores=`cat /proc/cpuinfo | grep processor | wc -l`
process_cpu=$(($cpu_cores / $instance))

client_pids=""
for i in $(seq 1 $instance); do
    cpu_start=$((($i - 1) * $process_cpu))
    cpu_end=$(($i * $process_cpu - 1))
    port=$(($REDIS_PORT + i - 1))
    client_cmd="numactl -C $cpu_start-$cpu_end  memtier_benchmark  -s localhost -p $port -t 8 -c 10 -d 32 --test-time=6 --ratio=1:1 --pipeline=1  --key-maximum=100000 "
    ${client_cmd} &
    client_pids="${client_pids} $!"
done

for pid in ${client_pids}; do wait ${pid} ; done

