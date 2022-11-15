#
# migrate from tuned to keentune
#

[main]
summary=Optimize for low power consumption

[cpu]
governor=ondemand|powersave
energy_perf_bias=powersave|power

[eeepc_she]

[vm]

[audio]
timeout=10

[video]
radeon_powersave=dpm-battery, auto

[disk]
# Comma separated list of devices, all devices if commented out.
# devices=sda

[net]
# Comma separated list of devices, all devices if commented out.
# devices=eth0

[scsi_host]
alpm=min_power

[sysctl]
vm.laptop_mode=5
vm.dirty_writeback_centisecs=1500
kernel.nmi_watchdog=0

[script]
script=${i:PROFILE_DIR}/script.sh
