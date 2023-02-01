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


class TestKeenTuneUiSmartNormal(unittest.TestCase):
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
        self.driver.get("http://{}:8082/list/tuning-task/".format(self.web_ip))

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.get("http://{}:8082/list/tuning-task".format(self.web_ip))
        del_list = ["auto_test_TPE","auto_test_HORD","auto_test_Random","auto_test2_TPE","auto_test2_HORD","auto_test2_Random"]
        for i in range(9):
            first_text = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
            if first_text in del_list:
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[6]/span[1]'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
                sleep(5)
            else:
                break
        self.driver.quit()

    def test_createjob(self):
        Algorithm_list = ["TPE","HORD","Random"]
        #遍历创建不同算法任务
        for i in Algorithm_list:
            self.wait.until(EC.element_to_be_clickable((By.XPATH,'//button[@class="ant-btn ant-btn-default"]'))).click()
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test_" + i)
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(i)
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
            self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(10)
            self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
            #等待任务执行完成，任务完成重新创建下一个，超时则结束
            for j in range(1,9):
                sleep(35)
                self.driver.refresh()
                Total_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[11]'))).text
                if Total_Time != "-":
                    break
                elif j == 8:
                    self.assertNotIn("-", Total_Time)
        name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.assertIn("auto_test_"+Algorithm_list[-1], name)

    def test_createjob02(self):
        Algorithm_list = ["TPE","HORD","Random"]
        #分别创建不同参数任务
        for i in Algorithm_list:
            self.wait.until(EC.element_to_be_clickable((By.XPATH,'//button[@class="ant-btn ant-btn-default"]'))).click()
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test2_" + i)
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(i)
            self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
            self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.CONTROL, "a")
            self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.BACKSPACE)
            self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(10)
            if i == "TPE":
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(2)
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(2)
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(2)
            elif i == "HORD":
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(2)
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(2)
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(1)
            else:
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "baseline_bench_round"))).send_keys(2)
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "tuning_bench_round"))).send_keys(1)
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(Keys.CONTROL, "a")
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(Keys.BACKSPACE)
                self.wait.until(EC.visibility_of_element_located((By.ID, "recheck_bench_round"))).send_keys(1)
            self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
            for j in range(1,9):
                sleep(35)
                self.driver.refresh()
                Total_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[11]'))).text
                if Total_Time != "-":
                    break
                elif j == 8:
                    self.assertNotIn("-",Total_Time)
        name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.assertIn("auto_test2_"+Algorithm_list[-1], name)

    def test_detail(self):
        #获取任务页面参数
        Name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        Algorithm = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[3]'))).text
        Status = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[4]'))).text
        Iteration = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[5]'))).text
        Start_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[9]'))).text
        End_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[10]'))).text
        #点击详情页，获取详情页参数进行对比
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[12]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[1]/span[1]'))).click()
        Job_Name  = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-row tag_row___Mi3-g"]/div[1]//div[@class="ellipsis___2bpK7"]'))).text
        detail_Ststus = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-row tag_row___Mi3-g"]/div[2]//div[@class="ellipsis___2bpK7"]'))).text
        detail_Algorithm = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-row tag_row___Mi3-g"]/div[4]//div[@class="ellipsis___2bpK7"]'))).text
        detail_Iteration = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-row tag_row___Mi3-g"]/div[5]//div[@class="ellipsis___2bpK7"]'))).text
        detail_Start_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-row tag_row___Mi3-g"]/div[7]//div[@class="ellipsis___2bpK7"]'))).text
        detail_End_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-row tag_row___Mi3-g"]/div[8]//div[@class="ellipsis___2bpK7"]'))).text
        self.driver.back()
        self.assertIn(Job_Name, Name)
        self.assertIn(detail_Ststus, Status)
        self.assertIn(detail_Algorithm, Algorithm)
        self.assertIn(detail_Iteration, Iteration)
        self.assertIn(detail_Start_Time, Start_Time)
        self.assertIn(detail_End_Time, End_Time)

    def test_log(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[2]/span[1]'))).click()
        log_checkfile_1 = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="CodeMirror-code"]/div[3]/pre/span')))
        log_checkfile_2 = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="CodeMirror-code"]/div[9]/pre/span/span')))
        self.assertIn("Run benchmark as baseline:", log_checkfile_1.text)
        self.assertIn("10", log_checkfile_2.text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-footer"]/button[@class="ant-btn ant-btn-primary"]'))).click()
        sleep(5)

    def test_rollback(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[3]/span[1]'))).click()
        alert_text = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@style="text-align: left;"]/pre'))).text
        self.assertIn('ok', alert_text)
        sleep(5)

    def test_dump(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                    '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[4]/span[1]'))).click()
        alert_text1 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-message-custom-content ant-message-success"]/span[2]/div/div[1]/pre'))).text
        alert_text2 = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-message-custom-content ant-message-success"]/span[2]/div/div[2]/pre'))).text
        name = alert_text2.split('/')[-1]
        self.assertIn('ok', alert_text1)

        self.driver.find_element(By.XPATH, '//div[@class="ant-pro-top-nav-header-logo"]//img').click()
        sleep(5)
        self.driver.find_element(By.XPATH, '//div[@class="list___y_nmN"]/div[1]//img').click()
        sleep(1)
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//tr[@data-row-key="1"]/td[1]//span'))).click()
        name_profile = self.driver.find_element(By.XPATH,'//tr[@data-row-key="1-1"]//td[1]').text
        self.assertEqual(name, name_profile)
        self.driver.find_element(By.XPATH, '//div[@class="ant-pro-top-nav-header-logo"]//img').click()
        sleep(3)
        self.driver.find_element(By.XPATH, '//div[@class="list___y_nmN"]/div[2]//img').click()
        sleep(3)

    def test_rerun(self):
        #点击重跑按钮
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-dropdown ant-dropdown-placement-bottomLeft "]/ul/li[5]/span[1]'))).click()
        rerun_Name = self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).get_attribute("value")
        rerun_Algorithm = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-form-item-control-input"]//div[@class="ant-select-selector"]/span[@class="ant-select-selection-item"]'))).text
        rerun_Iteration = self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).get_attribute("value")
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-footer"]//button[@class="ant-btn ant-btn-primary"]/span'))).click()
        for j in range(1,9):
            sleep(35)
            self.driver.refresh()
            Total_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[11]'))).text
            if Total_Time != "-":
                break
            elif j == 8 :
                self.assertNotIn("-", Total_Time)
        web_Name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        web_Algorithm = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[3]'))).text
        web_Iteration = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[5]'))).text
        self.assertEqual(rerun_Name, web_Name)
        self.assertEqual(rerun_Algorithm, web_Algorithm)
        self.assertEqual(rerun_Iteration, web_Iteration)
        sleep(5)

    def test_delete(self):
        first_name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[6]/span[1]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
        sleep(5)
        new_first_name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.assertNotEqual(first_name, new_first_name)

    def test_refresh(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//span[@class="anticon anticon-reload"]'))).click()

    def test_setting(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                            '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-tree-list-holder-inner"]/div[1]//span[4]'))).click()
        sleep(1)
        ele = self.driver.find_element(By.XPATH, '//thead[@class="ant-table-thead"]')
        self.assertNotIn("Name", ele.text)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//a[@class="ant-pro-table-column-setting-action-rest-button"]'))).click()
        self.assertIn("Name", ele.text)
        sleep(1)

    def test_sorting(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//thead[@class="ant-table-thead"]//th[9]'))).click()
        times = self.wait.until(EC.visibility_of_all_elements_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr/td[9]/span')))
        web_time_list = []
        for time in times:
            web_time_list.append(time.text)
        sort_time = sorted(web_time_list)
        self.assertEqual(web_time_list, sort_time)
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//thead[@class="ant-table-thead"]//th[9]'))).click()
        sleep(5)

    def test_language_switch(self):
        lan_dict = {"en": "Auto-Tuning Job List", "cn": "智能参数调优任务记录"}
        start_value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img'))).click()
        end_value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        sleep(1)
        language = "en" if "Auto-Tuning" in end_value else "cn"
        self.assertNotEqual(end_value, start_value)
        self.assertIn(end_value, lan_dict[language])
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img').click()
