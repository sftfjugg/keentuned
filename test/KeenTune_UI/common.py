from time import sleep
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC


def keentuneInit(obj):
    obj.driver.get("http://{}:8082/settings".format(obj.web_ip))
    sleep(5)
    brain_config = "[brain]\nBRAIN_IP = localhost"
    bench_config = "[bench-group-1]\nBENCH_SRC_IP = localhost\nBENCH_DEST_IP = localhost\nBENCH_SRC_PORT = 9874\nBENCH_CONFIG = wrk_http_long.json"
    target_config = "[target-group-1]\nTARGET_IP = localhost\nTARGET_PORT = 9873\nPARAMETER = sysctl.json"

    obj.wait.until(EC.visibility_of_element_located((By.ID, "brain"))).send_keys(Keys.CONTROL, "a")
    obj.wait.until(EC.visibility_of_element_located((By.ID, "brain"))).send_keys(Keys.BACKSPACE)
    obj.wait.until(EC.visibility_of_element_located((By.ID, "brain"))).send_keys(brain_config)
    sleep(3)
    obj.wait.until(EC.visibility_of_element_located((By.ID, "benchGroup"))).send_keys(Keys.CONTROL, "a")
    obj.wait.until(EC.visibility_of_element_located((By.ID, "benchGroup"))).send_keys(Keys.BACKSPACE)
    obj.wait.until(EC.visibility_of_element_located((By.ID, "benchGroup"))).send_keys(bench_config)
    sleep(3)
    obj.wait.until(EC.visibility_of_element_located((By.ID, "targetGroup"))).send_keys(Keys.CONTROL, "a")
    obj.wait.until(EC.visibility_of_element_located((By.ID, "targetGroup"))).send_keys(Keys.BACKSPACE)
    obj.wait.until(EC.visibility_of_element_located((By.ID, "targetGroup"))).send_keys(target_config)
    sleep(3)

    obj.wait.until(EC.visibility_of_element_located((By.ID, "Check"))).click()
    sleep(5)
