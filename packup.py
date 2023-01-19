import os
import re  
from datetime import datetime

""" warpping the KeenTune module keentuned

This script will
1. Check the version number in *.spec and Makefile
2. Check the date of changelog in *.spec
3. Pick necessary file to a folder named as keentuned-{version}
4. Pack the folder to .tar.gz
5. Get a copy of *.spec.

You can run this script in any position as 

python3 /path/in/your/environment/packup.py
"""
source_dir = os.path.split(os.path.realpath(__file__))[0]

def dateCheck(spec):
    date_items = re.findall(r"\* (\w+) (\w+) (\d+) (\d+) .*",spec)
    for date in date_items:
        _date = datetime.strptime("{} {} {}".format(date[3], date[1], date[2]),"%Y %b %d")
        if not _date.strftime("%a") == date[0]:
            raise Exception("week error:'{}', should be '{}'".format(date, _date.strftime("%a")))


def warppingCheck():
    with open(os.path.join(source_dir,"keentuned.spec"),'r') as f:
        spec = f.read()
        version_in_spec = re.search("Version:        ([\d.]+)\n",spec).group(1)
        release_in_spec = re.search("define anolis_release (\d)\n",spec).group(1)
        print("Get version: {}-{}".format(version_in_spec, release_in_spec))
        
        dateCheck(spec)
        if int(release_in_spec) <= 0:
            print("[Failed] release number must be greater than zero!")
            return 

        if re.search(" - {}".format(version_in_spec), spec):
            print("[OK] check version in changelog at keentuned.spec")
        else:
            print("[Failed] wrong version number in changelog of keentuned.spec")
            return
        
    with open(os.path.join(source_dir,"Makefile"), 'r') as f:
        makefile = f.read()

        if re.search(version_in_spec, makefile):
            print("[OK] check version in changelog at Makefile")
        else:
            print("[Failed] wrong version number in Makefile")
            return

    print("Start wrap up of keentune-{}".format(version_in_spec))
    return version_in_spec


if __name__ == "__main__":
    version_in_spec = warppingCheck()

    if os.path.exists("keentuned-{}".format(version_in_spec)):
        os.system("rm -rf keentuned-{}".format(version_in_spec))
    
    os.system("mkdir keentuned-{}".format(version_in_spec))

    os.system("cp -r {} keentuned-{}".format(os.path.join(source_dir,"cli"), version_in_spec))
    os.system("cp -r {} keentuned-{}".format(os.path.join(source_dir,"daemon"), version_in_spec))
    os.system("cp -r {} keentuned-{}".format(os.path.join(source_dir,"docs"), version_in_spec))
    os.system("cp -r {} keentuned-{}".format(os.path.join(source_dir,"vendor"), version_in_spec))
    os.system("cp -r {} keentuned-{}".format(os.path.join(source_dir,"man"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"go.mod"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"go.sum"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"keentuned.conf"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"keentuned.service"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"LICENSE"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"Makefile"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"README.md"), version_in_spec))
    os.system("cp {} keentuned-{}".format(os.path.join(source_dir,"keentune.bash"), version_in_spec))
    
    if os.path.exists(os.path.join("keentuned-{}".format(version_in_spec),"vendor")):
        os.system("tar -cvzf keentuned-{}.tar.gz keentuned-{}".format(version_in_spec, version_in_spec))
    else:
        print("[ERROR] run 'go mod vendor'")
        
    if os.path.exists("keentuned-{}".format(version_in_spec)):
        os.system("rm -rf keentuned-{}".format(version_in_spec))

    os.system("cp {} ./".format(os.path.join(source_dir, "keentuned.spec")))
