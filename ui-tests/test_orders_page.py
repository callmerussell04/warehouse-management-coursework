import pytest
import time
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import Select
from webdriver_manager.chrome import ChromeDriverManager
from selenium.common.exceptions import NoAlertPresentException

BASE_URL = "http://localhost"
WORKER_CREDENTIALS = {"username": "worker", "password": "3e2w1q"}

@pytest.fixture(scope="function")
def driver():
    chrome_options = Options()
    chrome_options.add_argument("--no-sandbox")
    chrome_options.add_argument("--disable-dev-shm-usage")
    chrome_options.add_argument("--window-size=1920,1080")

    service = Service(ChromeDriverManager().install())
    driver = webdriver.Chrome(service=service, options=chrome_options)
    yield driver
    driver.quit()

def click_element_js(driver, element):
    driver.execute_script("arguments[0].click();", element)

def login_as_worker(driver):
    driver.get(f"{BASE_URL}/login")
    wait = WebDriverWait(driver, 10)
    
    wait.until(EC.visibility_of_element_located((By.NAME, "username"))).send_keys(WORKER_CREDENTIALS["username"])
    driver.find_element(By.NAME, "password").send_keys(WORKER_CREDENTIALS["password"])
    driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    wait.until(EC.url_to_be(f"{BASE_URL}/"))

def ensure_dependencies_exist(driver):
    wait = WebDriverWait(driver, 5)
    
    driver.get(f"{BASE_URL}/counterparties")
    try:
        try:
            wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "tbody tr")))
            if "Нет данных" in driver.find_element(By.TAG_NAME, "body").text:
                raise Exception("Empty")
        except:
            driver.find_element(By.XPATH, "//button[contains(text(), '+ Добавить')]").click()
            modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
            modal.find_element(By.NAME, "name").send_keys("Auto Supplier")
            Select(modal.find_element(By.NAME, "type")).select_by_value("supplier")
            modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
            wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    except Exception:
        pass

    driver.get(f"{BASE_URL}/products")
    try:
        try:
            wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "tbody tr")))
            if "Нет товаров" in driver.find_element(By.TAG_NAME, "body").text:
                raise Exception("Empty")
        except:
            driver.find_element(By.XPATH, "//button[contains(text(), '+ Добавить товар')]").click()
            modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
            modal.find_element(By.NAME, "sku").send_keys("AUTO-SKU")
            modal.find_element(By.NAME, "name").send_keys("Auto Product")
            modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
            wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    except Exception:
        pass

def test_orders_page_structure(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/orders")
    wait = WebDriverWait(driver, 10)

    header = wait.until(EC.visibility_of_element_located((By.XPATH, "//h2[contains(text(), 'Заказы')]")))
    assert header.is_displayed()
    assert driver.find_element(By.XPATH, "//button[contains(text(), '+ Новый заказ')]").is_displayed()

def test_create_inbound_order(driver):
    login_as_worker(driver)
    ensure_dependencies_exist(driver)
    driver.get(f"{BASE_URL}/orders")
    wait = WebDriverWait(driver, 10)

    old_row_text = ""
    try:
        old_row_text = driver.find_element(By.CSS_SELECTOR, "tbody tr").text
    except:
        pass

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Новый заказ')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))

    Select(modal.find_element(By.TAG_NAME, "select")).select_by_value("inbound")

    modal.find_element(By.XPATH, "//button[contains(text(), 'Выбрать контрагента')]").click()
    sel_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Выберите')]]")))
    click_element_js(driver, sel_modal.find_element(By.XPATH, ".//button[contains(text(), 'Выбрать')]"))
    wait.until(EC.invisibility_of_element_located(sel_modal))

    modal.find_element(By.XPATH, "//button[contains(text(), '+ Добавить товар')]").click()
    prod_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Выберите товар')]]")))
    click_element_js(driver, prod_modal.find_element(By.XPATH, ".//button[contains(text(), 'Выбрать')]"))
    wait.until(EC.invisibility_of_element_located(prod_modal))

    modal.find_element(By.CSS_SELECTOR, "input[type='number']").send_keys("10")
    click_element_js(driver, modal.find_element(By.CSS_SELECTOR, "button[type='submit']"))
    
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    def row_updated(d):
        try:
            row = d.find_element(By.CSS_SELECTOR, "tbody tr")
            return row if row.text != old_row_text else False
        except:
            return False

    new_row = wait.until(row_updated)
    assert "Поступление" in new_row.text

def test_create_outbound_validation(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/orders")
    wait = WebDriverWait(driver, 10)

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Новый заказ')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))

    Select(modal.find_element(By.TAG_NAME, "select")).select_by_value("outbound")
    wait.until(EC.presence_of_element_located((By.ID, "destination")))
    
    click_element_js(driver, modal.find_element(By.CSS_SELECTOR, "button[type='submit']"))
    assert "was-validated" in modal.find_element(By.TAG_NAME, "form").get_attribute("class")

def test_order_status_lifecycle(driver):
    login_as_worker(driver)
    ensure_dependencies_exist(driver)
    
    driver.get(f"{BASE_URL}/orders")
    wait = WebDriverWait(driver, 10)

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Новый заказ')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.XPATH, "//button[contains(text(), 'Выбрать контрагента')]").click()
    sel_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Выберите')]]")))
    click_element_js(driver, sel_modal.find_element(By.XPATH, ".//button[contains(text(), 'Выбрать')]"))
    time.sleep(0.5)

    modal.find_element(By.XPATH, "//button[contains(text(), '+ Добавить товар')]").click()
    prod_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Выберите товар')]]")))
    click_element_js(driver, prod_modal.find_element(By.XPATH, ".//button[contains(text(), 'Выбрать')]"))
    time.sleep(0.5)

    click_element_js(driver, modal.find_element(By.CSS_SELECTOR, "button[type='submit']"))
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    
    driver.refresh()
    time.sleep(1)
    
    first_row = wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "tbody tr")))
    click_element_js(driver, first_row.find_element(By.XPATH, ".//button[contains(text(), 'Просмотр')]"))

    view_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Детали заказа')]]")))
    
    click_element_js(driver, wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), 'В обработку')]"))))
    wait.until(EC.presence_of_element_located((By.XPATH, "//span[contains(text(), 'В обработке')]")))

    click_element_js(driver, wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), 'Выполнить')]"))))
    wait.until(EC.presence_of_element_located((By.XPATH, "//span[contains(text(), 'Выполнен')]")))

    click_element_js(driver, view_modal.find_element(By.XPATH, "//button[contains(text(), 'Закрыть')]"))
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    
    wait.until(EC.text_to_be_present_in_element((By.CSS_SELECTOR, "tbody tr"), "Выполнен"))
    
    status_badge = driver.find_element(By.XPATH, "//tbody//tr//span[contains(@class, 'badge') and contains(text(), 'Выполнен')]")
    badge_class = status_badge.get_attribute("class")
    
    assert "success" in badge_class or "bg-success" in badge_class, f"Ожидался класс успеха, получен: {badge_class}"

def test_delete_order(driver):
    login_as_worker(driver)
    ensure_dependencies_exist(driver)
    
    driver.get(f"{BASE_URL}/orders")
    wait = WebDriverWait(driver, 10)

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Новый заказ')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.XPATH, "//button[contains(text(), 'Выбрать контрагента')]").click()
    sel_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Выберите')]]")))
    click_element_js(driver, sel_modal.find_element(By.XPATH, ".//button[contains(text(), 'Выбрать')]"))
    time.sleep(0.5)

    modal.find_element(By.XPATH, "//button[contains(text(), '+ Добавить товар')]").click()
    prod_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Выберите товар')]]")))
    click_element_js(driver, prod_modal.find_element(By.XPATH, ".//button[contains(text(), 'Выбрать')]"))
    time.sleep(0.5)
    
    click_element_js(driver, modal.find_element(By.CSS_SELECTOR, "button[type='submit']"))
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    time.sleep(1)

    first_row = wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "tbody tr")))
    click_element_js(driver, first_row.find_element(By.XPATH, ".//button[contains(text(), 'Просмотр')]"))

    wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')][.//div[contains(text(), 'Детали заказа')]]")))
    
    delete_btn = wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), 'Удалить заказ')]")))
    click_element_js(driver, delete_btn)

    try:
        WebDriverWait(driver, 3).until(EC.alert_is_present())
        driver.switch_to.alert.accept()
    except NoAlertPresentException:
        pytest.fail("Alert не появился")

    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    assert len(driver.find_elements(By.XPATH, "//div[contains(text(), 'Детали заказа')]")) == 0
