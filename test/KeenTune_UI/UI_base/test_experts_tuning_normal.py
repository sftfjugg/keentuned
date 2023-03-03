import os
import sys
import unittest
from time import sleep
from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver import ActionChains

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
            self.driver.implicitly_wait(10)
            self.wait = WebDriverWait(self.driver, 60, 0.5)
            self.AC = ActionChains(self.driver)

        else:
            if no_ui:
                option = webdriver.ChromeOptions()
                option.add_argument('headless')
                option.add_argument('--start-maximized')
                self.driver = webdriver.Chrome(chrome_options=option)
                self.wait = WebDriverWait(self.driver, 30, 0.5)
                self.AC = ActionChains(self.driver)
            else:
                self.driver = webdriver.Chrome()
                self.driver.maximize_window()
                self.wait = WebDriverWait(self.driver, 30, 0.5)
                self.AC = ActionChains(self.driver)
        keentuneInit(self)
        self.driver.get("http://{}:8082/list/profile".format(self.web_ip))
        sleep(5)

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.get("http://{}:8082/list/profile".format(self.web_ip))
        isactive = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]/td[3]'))).text
        count = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        if isactive == 'Yes':
            for number in range(1, count + 1):
                if self.wait.until(EC.visibility_of_element_located(
                        (By.XPATH, f'//tr[@data-row-key="1-{number}"]/td[2]'))).text == 'active':
                    self.wait.until(
                        EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]/div/div[4]'))).click()
                break

        for number in range(count):
            self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
            self.wait.until(
                EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]//td[4]//div[1]//div[1]/span'))).click()
            self.wait.until(
                EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-popover-buttons"]/button[2]/span'))).click()
            sleep(5)
        self.driver.quit()

    def tearDown(self) -> None:
        self.driver.refresh()
        sleep(5)

    def test_copyfile(self):
        rows = len(self.driver.find_elements(By.XPATH, '//td[@class="ant-table-cell ant-table-row-expand-icon-cell"]'))

        for number in range(2, rows + 1):
            count = int(
                self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
            self.wait.until(
                EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="{number}"]/td[1]//span'))).click()
            self.wait.until(
                EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="{number}-1"]/td[4]//div[2]'))).click()
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(f"{number}")
            content = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//textarea[@id="info"]'))).text
            self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
            sleep(5)
            count2 = int(
                self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
            self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
            ele_copy = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, f'//tr[@data-row-key="1-{number - 1}"]//td[1]'))).text

            self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, f'//tr[@data-row-key="1-{number - 1}"]//td[4]//div[1]//div[3]/span'))).click()
            sleep(1)
            copy_content = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//textarea[@id="info"]'))).text
            self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
            self.assertIn(f'{number}.conf', ele_copy)
            self.assertIn(content, copy_content)
            self.assertEqual(int(count) + 1, int(count2))

    def test_copyfile_check_name(self):
        count = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="2"]/td[1]//span'))).click()
        file_name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, f'//tr[@data-row-key="2-1"]/td[1]'))).text.split('.')[0]
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="2-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(5)
        count2 = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_copy = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, f'//tr[@data-row-key="1-{count2}"]//td[1]'))).text
        self.assertEqual(f'{file_name}_copy.conf', ele_copy)
        self.assertEqual(int(count) + 1, int(count2))

    def test_checkfile(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="5"]/td[1]//span'))).click()
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="5-1"]/td[4]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()

        sleep(2)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[1]//span'))).click()
        sleep(1)
        ele_checkfile = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//div[@class="CodeMirror-code"]')))
        self.assertIn("vm.dirty_background_ratio", ele_checkfile.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/button'))).click()

    def test_creatfile(self):
        count = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("zzzcreatfile")
        self.wait.until(EC.visibility_of_element_located((By.ID, "info"))).send_keys(
            "[sysctl]\nkernel.sched_migration_cost_ns: 5000")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(5)
        count2 = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_creat = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, f'//tr[@data-row-key="1-{count2}"]//td[1]')))
        sleep(1)
        self.assertEqual("zzzcreatfile.conf", ele_creat.text)
        self.assertEqual(int(count) + 1, int(count2))

    def test_editor(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        sleep(1)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[3]/span'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("111")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        sleep(5)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        file_name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1-1"]/td[1]'))).text
        sleep(1)
        self.assertEqual("111.conf", file_name)

    def test_set_group(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]//td[4]//div[5]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//label[@class="ant-radio-wrapper"]//span[2]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//div[3]/div/div[2]/button/span'))).click()
        sleep(5)
        ele_set = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1-1"]/td[3]'))).text
        status = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1-1"]/td[2]'))).text
        isactive = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]/td[3]'))).text
        self.assertIn("[target-group-1]", ele_set)
        self.assertEqual('active', status)
        self.assertEqual('Yes', isactive)

    def test_rollback(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]//div[4]'))).click()
        sleep(5)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        ele_set = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1-1"]/td[3]'))).text
        status = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1-1"]/td[2]'))).text
        isactive = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]/td[3]'))).text
        self.assertIn("-", ele_set)
        self.assertEqual('available', status)
        self.assertEqual('No', isactive)

    def test_deletefile(self):
        isactive = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]/td[3]'))).text
        count = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        if isactive == 'Yes':
            for number in range(1, count + 1):
                if self.wait.until(EC.visibility_of_element_located(
                        (By.XPATH, f'//tr[@data-row-key="1-{number}"]/td[2]'))).text == 'active':
                    self.wait.until(
                        EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1-1"]/td[4]/div/div[4]'))).click()
                break
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        sleep(1)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//tr[@data-row-key="1-1"]//td[4]//div[1]//div[1]/span'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-popover-buttons"]/button[2]/span'))).click()
        sleep(5)
        count2 = int(
            self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tr[@data-row-key="1"]//td[4]'))).text)
        self.assertEqual(count - 1, count2)

    def test_language_switch(self):
        lan_dict = {"en": "List of Expert Knowledge Based Tuning Profiles", "cn": "调优专家知识库列表"}
        start_value = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]'))).text
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img'))).click()
        end_value = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]'))).text
        sleep(1)
        language = "en" if "Tuning Profiles" in end_value else "cn"
        self.assertNotEqual(end_value, start_value)
        self.assertIn(lan_dict[language], end_value)
        self.wait.until(EC.visibility_of_element_located((By.XPATH,
                                                          '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img'))).click()

    def test_refresh(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, f'//tr[@data-row-key="1"]/td[1]//span'))).click()
        tabel_name = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//tbody[@class="ant-table-tbody"]//thead[@class="ant-table-thead"]/tr/th[1]'))).text
        page_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-pro-table ProTable___1C14L"]'))).text
        self.assertIn(tabel_name, page_info)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        sleep(5)
        page_info2 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-pro-table ProTable___1C14L"]'))).text
        self.assertNotIn(tabel_name, page_info2)

    def test_set_list(self):
        # step1 取消一列显示
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[1]//span[4]'))).click()
        sleep(1)
        ele_title = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead')))
        self.assertNotIn("Profile Set", ele_title.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.assertIn("Profile Set", ele_title.text)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        # step2 取消所有列
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-popover-title"]//label/span'))).click()
        sleep(1)
        ele_title = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead')))
        self.assertNotIn("Profile Set", ele_title.text)
        self.assertNotIn("Contains Active?", ele_title.text)
        self.assertNotIn("Profile Count", ele_title.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.assertIn("Profile Set", ele_title.text)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        # step3 固定在列首
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        head_ele = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[2]//span[4]')))
        self.AC.move_to_element(head_ele).perform()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@draggable="true"][2]/span[5]/span/span/span/span[1]'))).click()
        # step4 固定在列尾
        tail_ele = self.wait.until(EC.visibility_of_element_located((By.XPATH,
                                                                     '//div[@class="ant-popover-inner-content"]/div/div[2]//div[@class="ant-tree-list-holder-inner"]/div[1]/span[4]')))
        self.AC.move_to_element(tail_ele).perform()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-popover-inner-content"]/div/div[2]//div[@class="ant-tree-list-holder-inner"]/div[1]//span[@aria-label="vertical-align-bottom"]'))).click()
        ele_title = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-table-content"]//thead')))
        set_title = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-popover-inner-content"]')))
        self.assertEqual('Contains Active? Profile Count Profile Set', ele_title.text)
        setting = ['固定在左侧', 'Contains Active?', '不固定', 'Profile Count', '固定在右侧', 'Profile Set']
        self.assertEqual(setting, set_title.text.split('\n'))
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        sleep(5)
        # step5 取消固定
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        head_ele = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[2]//span[4]')))
        self.AC.move_to_element(head_ele).perform()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@draggable="true"][2]/span[5]/span/span/span/span[1]'))).click()
        ele_title = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-table-content"]//thead')))
        self.assertEqual('Contains Active? Profile Set Profile Count', ele_title.text)
        cancel_ele = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-popover-inner-content"]/div/div[1]')))
        self.AC.move_to_element(cancel_ele).perform()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-popover-inner-content"]/div/div[1]//span[@aria-label="vertical-align-middle"]'))).click()
        ele_title = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead')))
        self.assertEqual('Profile Set Contains Active? Profile Count', ele_title.text)
