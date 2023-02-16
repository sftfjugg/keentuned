import os
import sys
import unittest
from time import sleep
from selenium import webdriver
from selenium.common.exceptions import NoSuchElementException
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import keentuneInit


class TestKeenTuneUiAbnormal(unittest.TestCase):
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
        value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        if "Tuning Profiles" in value:
            self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img').click()

        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.driver.find_element(By.ID, "name").send_keys("1")
        self.driver.find_element(By.ID, "info").send_keys("[sysctl]\nkernel.sched_migration_cost_ns: 5000")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
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

    def setUp(self) -> None:
        sleep(3)

    def tearDown(self) -> None:
        sleep(1)
        try:
            self.driver.find_element(By.XPATH,
                                     '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]')
        except NoSuchElementException:
            pass
        else:
            self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_group_empty(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]//td[4]//div[5]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//div[3]/div/div[2]/button/span'))).click()
        ele_group_error = self.driver.find_element(By.XPATH, '//div[@class="ant-message-notice-content"]//span[2]')
        sleep(1)
        self.assertIn("请选择一个配置，再提交", ele_group_error.text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//div[3]/div/div[1]/button/span'))).click()

    def test_copyfile_name_exsit(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("cpu_high_load")
        sleep(1)
        ele_nameexit = self.driver.find_element(By.XPATH, '//div[@class="ant-form-item-explain-error"]')
        self.assertIn("Profile Name名字重复!", ele_nameexit.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_creatfile_name_exsit(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("cpu_high_load")
        sleep(1)
        ele_nameexit = self.driver.find_element(By.XPATH,
                                                '//div[@class="ant-form-item-explain-error"]')
        self.assertIn("Profile Name名字重复!", ele_nameexit.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_editorfile_exsit_name(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[3]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("cpu_high_load")
        sleep(2)
        ele_deletecontent = self.driver.find_element(By.XPATH,
                                                     '//div[@class="ant-form-item-explain-error"]')
        sleep(2)
        self.assertIn("Profile Name名字重复!", ele_deletecontent.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_copyfile_name_empty(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        sleep(1)
        ele_emptyname = self.driver.find_element(By.XPATH,
                                                 '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        self.assertIn("请输入", ele_emptyname.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_copyfile_context_empty(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.BACKSPACE)
        sleep(1)
        ele_copy_context_empty = self.driver.find_element(By.XPATH,
                                                          '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        sleep(0.5)
        self.assertIn("请输入", ele_copy_context_empty.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_copyfile_context_error(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys("error_file_context")
        sleep(1)
        ele_copy_context_error = self.driver.find_element(By.XPATH,
                                                          '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        self.assertIn("第1行 数据格式不对!", ele_copy_context_error.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_creatfile_name_empty(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys("[sysctl]")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(1)
        ele_nameempty = self.driver.find_element(By.XPATH, '//div[@class="ant-form-item-explain-error"]')
        self.assertIn("请输入", ele_nameempty.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_creatfile_content_empty(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("content_empty")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        sleep(1)
        ele_contentempty = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        self.assertIn("请输入", ele_contentempty.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_creatfile_content_error(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("content_error")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys("content_error")
        sleep(1)
        ele_contenterror = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain-error"]')
        self.assertIn("第1行 数据格式不对!", ele_contenterror.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_editorfile_delete_name(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[3]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        ele_deletename = self.driver.find_element(By.XPATH,
                                                  '//div[@class="ant-form-item-explain-error"]')
        sleep(1)
        self.assertIn("请输入", ele_deletename.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_editorfile_delete_content(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[3]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.BACKSPACE)
        sleep(1)
        ele_deletecontent = self.driver.find_element(By.XPATH,
                                                     '//div[@class="ant-form-item-explain-error"]')
        self.assertIn("请输入", ele_deletecontent.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_editorfile_error_content(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[3]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys("error_content")
        sleep(1)
        ele_errorcontent = self.driver.find_element(By.XPATH, '//div[@class="ant-form-item-explain-error"]')
        sleep(1)
        self.assertIn("第1行 数据格式不对!", ele_errorcontent.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()
