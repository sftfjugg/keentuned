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


class TestKeenTuneUiSensitiveNormal(unittest.TestCase):
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

        # 在智能参数调优页面创建任务
        keentuneInit(self)
        self.driver.get("http://{}:8082/list/tuning-task".format(self.web_ip))
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys("TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(10)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        # 等待任务执行完成，任务完成重新创建下一个，超时则结束
        for j in range(1, 9):
            sleep(35)
            self.driver.refresh()
            Total_Time = self.wait.until(
                EC.visibility_of_element_located(
                    (By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[11]'))).text
            if Total_Time != "-":
                break
            elif j == 8:
                self.assertNotIn("-", Total_Time)
        self.driver.find_element(By.XPATH, '//div[@class="ant-pro-top-nav-header-logo"]//img').click()
        sleep(5)
        self.driver.find_element(By.XPATH, '//div[@class="list___y_nmN"]/div[3]//img').click()

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.get("http://{}:8082/list/tuning-task".format(self.web_ip))
        del_list = "auto_test_TPE"
        for i in range(9):
            first_text = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
            if first_text == del_list:
                self.wait.until(EC.element_to_be_clickable(
                    (By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[6]/span[1]'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
                sleep(1)
            else:
                break
        self.driver.find_element(By.XPATH, '//div[@class="ant-pro-top-nav-header-logo"]//img').click()
        sleep(5)
        self.driver.find_element(By.XPATH, '//div[@class="list___y_nmN"]/div[3]//img').click()
        del_list = ["auto_test_TPE", "auto_test_TPE2", "rerun_test"]
        for i in range(9):
            first_text = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
            if first_text in del_list:
                self.wait.until(EC.element_to_be_clickable(
                    (By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[4]/span[1]'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
                sleep(1)
            else:
                break
        self.driver.quit()

    def tearDown(self) -> None:
        self.driver.refresh()
        sleep(5)

    def test_create_job(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-right"]//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys("lasso")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys("1")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        # 等待任务执行完成，任务完成重新创建下一个，超时则结束
        for j in range(1, 5):
            sleep(20)
            self.driver.refresh()
            Total_Time = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[9]'))).text
            if Total_Time != "-":
                break
            elif j == 8:
                self.assertNotIn("-", Total_Time)

    def test_create_job2(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-right"]//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test_TPE2")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys("lasso")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys("3")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        # 等待任务执行完成，任务完成重新创建下一个，超时则结束
        for j in range(1, 5):
            sleep(20)
            self.driver.refresh()
            Total_Time = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[9]'))).text
            if Total_Time != "-":
                break
            elif j == 8:
                self.assertNotIn("-", Total_Time)

        name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.assertEqual("auto_test_TPE2", name)

    def test_detail(self):
        # 获取任务页面参数
        Name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        Data = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[3]'))).text
        Algorithm = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[4]'))).text
        Status = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[5]'))).text
        Trial = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[6]'))).text
        Start_Time = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[7]'))).text
        End_Time = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[8]'))).text
        Total_time = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[9]'))).text

        # 点击详情页，获取详情页参数进行对比
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[1]/span[1]'))).click()
        Job_Name = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[1]//div[@class="tag_value___37nPH"]'))).text
        detal_Data = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[2]//div[@class="tag_value___37nPH"]'))).text
        detail_Status = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[3]//div[@class="tag_value___37nPH"]'))).text
        detail_Algorithm = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[4]//div[@class="tag_value___37nPH"]'))).text
        detail_Trial = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[5]//div[@class="tag_value___37nPH"]'))).text
        detail_Start_Time = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[7]//div[@class="tag_value___37nPH"]'))).text
        detail_End_Time = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[8]//div[@class="tag_value___37nPH"]'))).text
        detail_Total_time = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-row tag_row___35DRX"]/div[9]//div[@class="tag_value___37nPH"]'))).text
        self.driver.back()
        self.assertEqual(Name, Job_Name)
        self.assertEqual(Data, detal_Data)
        self.assertEqual(Algorithm, detail_Algorithm)
        self.assertEqual(Status, detail_Status)
        self.assertEqual(Trial, detail_Trial)
        self.assertEqual(Start_Time, detail_Start_Time)
        self.assertEqual(End_Time, detail_End_Time)
        self.assertEqual(Total_time, detail_Total_time)

    def test_log(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-dropdown ant-dropdown-placement-bottomLeft "]/ul/li[2]/span[1]'))).click()
        log_checkfile_1 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="CodeMirror-code"]/div[3]/pre/span')))
        log_checkfile_2 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="CodeMirror-code"]/div[4]/pre/span')))
        self.assertIn("Initiate sensitization success.", log_checkfile_1.text)
        self.assertIn("identification results successfully", log_checkfile_2.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-footer"]/button[@class="ant-btn ant-btn-primary"]'))).click()
        sleep(2)

    def test_rerun(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-dropdown ant-dropdown-placement-bottomLeft "]/ul/li[3]/span[1]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("rerun_test")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        # 等待任务执行完成，任务完成重新创建下一个，超时则结束
        for j in range(1, 5):
            sleep(20)
            self.driver.refresh()
            Total_Time = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[9]'))).text
            if Total_Time != "-":
                break
            elif j == 8:
                self.assertNotIn("-", Total_Time)
        first_name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.assertEqual("rerun_test", first_name)

    def test_delete(self):
        first_name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[4]/span[1]'))).click()

        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
        sleep(3)
        new_first_name = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.assertNotEqual(first_name, new_first_name)

    def test_refresh(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        text = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//label[@class="ant-checkbox-wrapper ant-checkbox-wrapper-checked"]'))).text
        self.assertEqual('列展示', text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//span[@class="anticon anticon-reload"]'))).click()
        sleep(5)
        page_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="static_page___39bhG"]'))).text
        self.assertNotIn('列展示', page_info)

    def test_setting(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[1]//span[4]'))).click()
        sleep(1)
        ele = self.driver.find_element(By.XPATH, '//thead[@class="ant-table-thead"]')
        self.assertNotIn("Name", ele.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.assertIn("Name", ele.text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()

    def test_sorting(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//thead[@class="ant-table-thead"]//th[7]'))).click()
        times = self.wait.until(
            EC.visibility_of_all_elements_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr/td[7]/span')))
        web_time_list = []
        for time in times:
            web_time_list.append(time.text)
        sort_time = sorted(web_time_list)
        self.assertEqual(web_time_list, sort_time)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//thead[@class="ant-table-thead"]//th[7]'))).click()

    def test_language_switch(self):
        lan_dict = {"en": "Sensitivity Identification Job List", "cn": "敏感参数识别任务记录"}
        start_value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img'))).click()
        end_value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        sleep(1)
        language = "en" if "Sensitivity" in end_value else "cn"
        self.assertNotEqual(end_value, start_value)
        self.assertIn(end_value, lan_dict[language])
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img').click()

    def test_set_list(self):
        # step1 取消一列显示
        default_list = ['', 'Name', 'Data', 'Algorithm', 'Status', 'Trial', 'Start Time', 'End Time', 'Total Time',
                        'Operations']
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        table1 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[1]'))).text
        table2 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[2]'))).text

        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[1]//span[4]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[2]//span[4]'))).click()
        ele_title = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead')))
        self.assertNotIn(table1, ele_title.text)
        self.assertNotIn(table2, ele_title.text)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.assertIn(table1, default_list)
        self.assertIn(table2, default_list)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        # step2 取消所有列
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-popover-title"]//label/span'))).click()
        sleep(1)
        ele_title_cancle = self.wait.until(EC.visibility_of_element_located(
            (By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead'))).text

        tables = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-tree-list-holder-inner"]'))).text.split('\n')
        tables.append('Operations')
        for table in tables:
            self.assertNotIn(table, ele_title_cancle)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        titles_list = []
        titles = len(
            self.driver.find_elements(By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead/tr/th'))
        for i in range(1, titles + 1):
            title = self.wait.until(EC.visibility_of_element_located(
                (By.XPATH, f'//div[@class="ant-table ant-table-small"]/div/div/table/thead/tr/th[{i}]'))).text
            titles_list.append(title)
        for table in tables:
            self.assertIn(table, titles_list)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        # step3 固定在列首
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        head_ele = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-tree-list-holder-inner"]/div[3]')))
        set_head = head_ele.text
        self.AC.move_to_element(head_ele).perform()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@draggable="true"][3]/span[5]/span/span/span/span[1]'))).click()
        # step4 固定在列尾
        tail_ele = self.wait.until(EC.visibility_of_element_located((By.XPATH,
                                                                     '//div[@class="ant-popover-inner-content"]/div/div[2]//div[@class="ant-tree-list-holder-inner"]/div[1]')))
        set_tail = tail_ele.text
        self.AC.move_to_element(tail_ele).perform()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-popover-inner-content"]/div/div[2]//div[@class="ant-tree-list-holder-inner"]/div[1]/span[5]/span/span/span/span[2]'))).click()
        titles_list = []
        titles = len(
            self.driver.find_elements(By.XPATH, '//div[@class="ant-table-content"]//th'))
        for i in range(1, titles + 1):
            title = self.wait.until(EC.visibility_of_element_located(
                (By.XPATH, f'//div[@class="ant-table-content"]//th[{i}]'))).text
            titles_list.append(title)
        self.assertEqual(1, titles_list.index(set_head))
        self.assertEqual(titles - 1, titles_list.index(set_tail))
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                  'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()
        sleep(5)
        # step5 取消固定
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        set_result = self.wait.until(EC.visibility_of_element_located((By.XPATH,
                                                                       '//div[@class="ant-pro-table-column-setting-list ant-pro-table-column-setting-list-group"]'))).text
        head_ele = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-popover-inner-content"]/div/div[1]/div[3]')))
        self.AC.move_to_element(head_ele).perform()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH,
             '//div[@class="ant-popover-inner-content"]/div/div[1]/div[3]//span[@aria-label="vertical-align-middle"]'))).click()

        tail_ele = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-popover-inner-content"]/div/div[2]')))
        self.AC.move_to_element(tail_ele).perform()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//div[@class="ant-popover-inner-content"]/div/div[2]//span[@aria-label="vertical-align-middle"]'))).click()
        cancle_set = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-popover-inner"]'))).text
        titles_list = []
        titles = len(
            self.driver.find_elements(By.XPATH, '//div[@class="ant-table ant-table-small"]/div/div/table/thead/tr/th'))
        for i in range(1, titles + 1):
            title = self.wait.until(EC.visibility_of_element_located(
                (By.XPATH, f'//div[@class="ant-table ant-table-small"]/div/div/table/thead/tr/th[{i}]'))).text
            titles_list.append(title)
        self.assertIn('固定在左侧', set_result)
        self.assertIn('不固定', set_result)
        self.assertIn('固定在右侧', set_result)
        self.assertNotIn('固定在左侧', cancle_set)
        self.assertNotIn('不固定', cancle_set)
        self.assertNotIn('固定在右侧', cancle_set)
        self.assertEqual(default_list, titles_list)

    def test_compare(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]//label/span'))).click()
        sleep(10)
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[2]//label/span'))).click()

        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-primary"]'))).click()
        page_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="list_root___2lNK0"]//ol/li[2]/span[1]'))).text
        head_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="page_title___2bkEd"]'))).text

        self.assertEqual('敏感参数-Compare', page_info)
        self.assertEqual('敏感参数箱型图', head_info)

        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//div[@class="list_root___2lNK0"]//ol/li[1]/span[1]'))).click()
        sleep(3)
        back_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-pro-table-list-toolbar-left"]'))).text
        self.assertEqual('敏感参数识别任务记录', back_info)
