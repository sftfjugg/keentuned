import os
import re
import json
import requests
import subprocess
import time
import unittest
import logging

target_ip="localhost"
bench_ip="localhost"
brain_ip="localhost"
keentuned_ip="localhost"

target_port="9873"
bench_port="9874"
brain_port="9872"
keentuned_port="9871"

logger = logging.getLogger(__name__)


def sysCommand(cmd):
    result = subprocess.run(
        cmd,
        shell=True,
        close_fds=True,
        stderr=subprocess.PIPE,
        stdout=subprocess.PIPE
    )

    status = result.returncode
    output = result.stdout.decode('UTF-8', 'strict')
    error = result.stderr.decode('UTF-8', 'strict')

    return status, output, error


def getServerStatus(server):
    data_dict = {
        "keentuned": [keentuned_ip, keentuned_port],
        "keentune-brain": [brain_ip, brain_port],
        "keentune-target": [target_ip, target_port],
        "keentune-bench": [bench_ip, bench_port]
    }
    event = "sensitize_list" if server == "keentune-brain" else "status"
    url = "http://{}:{}/{}".format(data_dict[server][0], data_dict[server][1], event)
    res = requests.get(url, proxies={"http": None,"https": None})
    if res.status_code != 200:
        logger.error("Please check {} server is running...".format(server))
        result = 1
    else:
        result = 0
    return result


def checkServerStatus(server_list):
    result = False
    for server in server_list:
        status = getServerStatus(server)
        result = result or status
    return result


def deleteDependentData(param_name):
    cmd = "echo y | keentune param delete --job {}".format(param_name)
    sysCommand(cmd)
    cmd = 'echo y | keentune sensitize delete --job {}'.format(param_name)
    sysCommand(cmd)


def runParamTune(name, iteration=10):
    cmd = 'keentune param tune -i {} --job {}'.format(iteration, name)
    status, output, _ = sysCommand(cmd)
    assert status == 0
    path = re.search(r'\s+"(.*?)"', output).group(1)
    time.sleep(3)
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if '[BEST] Tuning improvement' in res_data or "[ERROR]" in res_data:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                 "Step5", "Step6", "[BEST] Tuning improvement"]
    res = all([word in res_data for word in word_list])
    result = 0 if res else 1
    return result


def runParamDump(name):
    cmd = 'echo y | keentune param dump -j {}'.format(name)
    sysCommand(cmd)
    path = "/var/keentune/profile/param1_group1.conf"
    res = os.path.exists(path)
    result = 0 if res else 1
    return result


def runProfileSet():
    cmd = 'keentune profile set param1_group1.conf'
    sysCommand(cmd)
    cmd = 'keentune profile list'
    _, output, _ = sysCommand(cmd)
    res = re.search(r'\[(.*?)\].+param1_group1.conf', output).group(1)
    result = 0 if res.__contains__('active') else 1
    return result


def runSensitizeCollect(name, iteration=10):
    cmd = 'keentune sensitize collect -i {} --data {}'.format(iteration, name)
    status, output, _ = sysCommand(cmd)
    assert status == 0
    path = re.search(r'\s+"(.*?)"', output).group(1)
    time.sleep(3)
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if 'Sensitization collection finished' in res_data or "[ERROR]" in res_data:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                 "Sensitization collection finished"]
    res = all([word in res_data for word in word_list])
    result = 0 if res else 1
    return result

def getSysBackupData():
    path ="/var/keentune/target/backup/Sysctl.json"
    if target_ip == "localhost":
        cmd = "sysctl -n {}"
        result = os.path.exists(path)
    else:
        cmd = "ssh %s 'sysctl -n {}'" % target_ip
        status = sysCommand("ssh {} 'ls -f {}'".format(target_ip, path))[0]
        result = True if status == 0 else False
    if result:
        status, output, _ = sysCommand('keentune param rollback')
        assert status == 0
        assert 'param rollback successfully' in output

    path = "conf/sysctl_backup.json"
    with open(path, "r", encoding='UTF-8') as f:
        backup_data = json.load(f)

    for param_name, param_info in backup_data.items():
        new_cmd = cmd.format(param_name)
        param_info["value"] = str(sysCommand(new_cmd)[1]).strip('\n')
        backup_data[param_name] = param_info

    with open(path, "w", encoding='UTF-8') as f:
        json.dump(backup_data, f, indent=4)

def checkBackupData():
    path = "conf/sysctl_backup.json"
    with open(path, "r", encoding='UTF-8') as f:
        backup_data = json.load(f)

    cmd = "sysctl -n {}" if target_ip == "localhost" else "ssh %s 'sysctl -n {}'" % target_ip
    for param_name, param_info in backup_data.items():
        new_cmd = cmd.format(param_name)
        value = str(sysCommand(new_cmd)[1]).strip('\n')
        if param_info["value"] != value:
            status = 1
            break
    else:
        status = 0
    return status

def checkProfileData(name, flag=False):
    cmd = "sysctl -n {}" if target_ip == "localhost" else "ssh %s 'sysctl -n {}'" % target_ip
    path = "/var/keentune/profile/{}.conf".format(name) if flag else "conf/{}.conf".format(name)
    with open(path, "r", encoding='UTF-8') as f:
        for line in f.readlines()[1:]:
            key = line.strip().split(": ")[0]
            value = line.strip().split(": ")[1]
            new_cmd = cmd.format(key)
            sys_val = str(sysCommand(new_cmd)[1]).strip('\n')
            if value != sys_val:
                status = 1
                break
        else:
            status = 0
    return status

def getTaskLogPath(cmd, status=0):
    res, output, _  = sysCommand(cmd)
    assert res == status
    path = re.search(r'\s+"(.*?)"', output).group(1)
    time.sleep(2)
    return path

def getTuneTaskResult(path):
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if '[BEST] Tuning improvement' in res_data or "[ERROR]" in res_data:
            break
        time.sleep(8)

    word_list = ["Step1", "Step2", "Step3", "Step4",
                    "Step5", "Step6", "[BEST] Tuning improvement"]
    result = all([word in res_data for word in word_list])
    return result

def getCollectTaskResult(path):
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if 'Sensitization collection finished' in res_data or "[ERROR]" in res_data:
            break
        time.sleep(8)
    word_list = ["Step1", "Step2", "Step3", "Step4", "Sensitization collection finished"]
    result = all([word in res_data for word in word_list])
    return result

def getTrainTaskResult(path):
    while True:
        with open(path, 'r') as f:
            res_data = f.read()
        if 'identification results successfully' in res_data or "[ERROR]" in res_data:
            break
        time.sleep(8)
    word_list = ["Step1", "Step2", "Step3", "identification results successfully"]
    result = all([word in res_data for word in word_list])
    return result

def deleteTmpFiles(file_list):
    for file in file_list:
        path = "/var/keentune/profile/{}.conf".format(file)
        if os.path.exists(path):
            os.remove(path)

def copyTmpFile(path1, path2):
    cmd = "cp {} {}".format(path1, path2)
    status, _, _  = sysCommand(cmd)
    assert status == 0
