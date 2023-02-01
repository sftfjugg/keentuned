import os
import sys
import unittest
from time import sleep
from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import keentuneInit


class TestKeenTuneUiNormal(unittest.TestCase):
    @classmethod
    def setUpClass(self, no_ui=False) -> None:
        if 'linux' in sys.platform:
            option = webdriver.ChromeOptions()
            option.add_argument('headless')
            option.add_argument('no-sandbox')
            option.add_argument('--start-maximized')
            option.add_argument('--disable-gpu')
            option.add_argument('lang=zh_CN.UTF-8')
            option.add_argument('--window-size=1920,1080')
            self.driver = webdriver.Chrome(options=option)
            self.driver.implicitly_wait(3)
            self.wait = WebDriverWait(self.driver, 30, 0.5)

        else:
            if no_ui:
                option = webdriver.ChromeOptions()
                option.add_argument('headless')
                option.add_argument('--start-maximized')
                self.driver = webdriver.Chrome(chrome_options=option)
                self.wait = WebDriverWait(self.driver, 30, 0.5)
            else:
                self.driver = webdriver.Chrome()
                self.driver.maximize_window()
                self.wait = WebDriverWait(self.driver, 30, 0.5)

        keentuneInit(self)
        self.driver.get("http://{}:8082/list/profile".format(self.web_ip))
        sleep(5)

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.get("http://{}:8082/list/profile".format(self.web_ip))
        for i in range(9):
            count_text = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[4]')
            sleep(1)
            if count_text.text != "0":
                self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1-1"]//td[4]//div[1]//div[1]/span'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-popover-buttons"]/button[2]/span'))).click()
                sleep(5)
            else:
                break
        self.driver.quit()

    def test_copyfile(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="3"]/td[1]//span'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="3-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("1")
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(5)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_copy = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1-1"]//td[1]')
        sleep(1)
        self.assertIn("1.conf", ele_copy.text)

    def test_creatfile(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("11")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys("[sysctl]\nkernel.sched_migration_cost_ns: 5000")
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(5)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_creat = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1-2"]//td[1]')
        sleep(1)
        self.assertIn("11.conf", ele_creat.text)

    def test_checkfile(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1-2"]/td[1]//span'))).click()
        sleep(1)
        ele_checkfile = self.driver.find_element(By.XPATH,'//div[@class="CodeMirror-code"]')
        self.assertIn("sysctl]", ele_checkfile.text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/button'))).click()
        sleep(1)

    def test_editor(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1-2"]/td[4]//div[3]/span'))).click()        
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("111")
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(5)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_editor = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1-2"]//td[1]')
        sleep(1)
        self.assertIn("111.conf", ele_editor.text)

    def test_set_group(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1-1"]//td[4]//div[5]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//label[@class="ant-radio-wrapper"]//span[2]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[3]/div/div[2]/button/span'))).click()
        sleep(5)
        ele_set = self.driver.find_element(By.XPATH,'//tr[@data-row-key="1-1"]/td[3]')
        self.assertIn("[target-group-1]", ele_set.text)

    def test_restore(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1-1"]/td[4]//div[4]'))).click()
        sleep(5)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_set = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1-1"]/td[3]')
        self.assertNotIn("[target-group-1]\nTARGET_IP = localhost", ele_set.text)

    def test_deletefile(self):
        for i in range(9):
            count_text = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[4]')
            sleep(1)
            if count_text.text != "0":
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1-1"]//td[4]//div[1]//div[1]/span'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-popover-buttons"]/button[2]/span'))).click()
                sleep(5)
                self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
            else:
                break
        ele_copy = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[4]')
        self.assertIn("0", ele_copy.text)

    def test_language_switch(self):
        lan_dict = {"en": "List of Expert Knowledge Based Tuning Profiles", "cn": "调优专家知识库列表"}
        start_value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img'))).click()
        end_value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        sleep(1)
        language = "en" if "Tuning Profiles" in end_value else "cn"
        self.assertNotEqual(end_value, start_value)
        self.assertIn(lan_dict[language], end_value)
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img').click()

    def test_refresh(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                     'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()

    def test_set_list(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                            '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-tree-list-holder-inner"]/div[1]//span[4]'))).click()
        sleep(1)
        ele = self.driver.find_element(By.XPATH, '//thead[@class="ant-table-thead"]')
        self.assertNotIn("Profile Set", ele.text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.assertIn("Profile Set", ele.text)

