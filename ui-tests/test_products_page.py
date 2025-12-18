import pytest
import uuid
import time
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager

BASE_URL = "http://localhost:5000"
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

def login_as_worker(driver):
    driver.get(f"{BASE_URL}/login")
    wait = WebDriverWait(driver, 10)
    
    wait.until(EC.visibility_of_element_located((By.NAME, "username"))).send_keys(WORKER_CREDENTIALS["username"])
    driver.find_element(By.NAME, "password").send_keys(WORKER_CREDENTIALS["password"])
    driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    wait.until(EC.url_to_be(f"{BASE_URL}/"))

def test_products_page_structure(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/products")
    wait = WebDriverWait(driver, 10)

    header = wait.until(EC.visibility_of_element_located((By.XPATH, "//h2[contains(text(), 'Товары')]")))
    assert header.is_displayed()

    add_btn = driver.find_element(By.XPATH, "//button[contains(text(), '+ Добавить товар')]")
    assert add_btn.is_displayed()
    assert add_btn.is_enabled()

    table_headers = ["Артикул (SKU)", "Название", "Остаток", "Обновлено", "Действия"]
    for h in table_headers:
        assert driver.find_element(By.XPATH, f"//th[contains(text(), '{h}')]").is_displayed()

def test_product_crud_lifecycle(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/products")
    wait = WebDriverWait(driver, 10)

    unique_id = str(uuid.uuid4())[:8]
    sku = f"TEST-SKU-{unique_id}"
    name = f"Test Product {unique_id}"
    new_name = f"Edited Product {unique_id}"

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Добавить товар')]"))).click()
    
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.NAME, "sku").send_keys(sku)
    modal.find_element(By.NAME, "name").send_keys(name)
    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    row = wait.until(EC.visibility_of_element_located((By.XPATH, f"//tr[td[contains(text(), '{sku}')]]")))
    assert name in row.text
    assert "0 шт." in row.text

    edit_btn = row.find_element(By.CSS_SELECTOR, "button .bi-pencil")
    edit_btn.click()

    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    name_input = modal.find_element(By.NAME, "name")
    assert name_input.get_attribute("value") == name
    
    name_input.clear()
    name_input.send_keys(new_name)
    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    wait.until(EC.text_to_be_present_in_element(
        (By.XPATH, f"//tr[td[contains(text(), '{sku}')]]"), 
        new_name
    ))

    row = driver.find_element(By.XPATH, f"//tr[td[contains(text(), '{sku}')]]")
    delete_btn = row.find_element(By.CSS_SELECTOR, "button .bi-trash")
    delete_btn.click()

    confirm_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')]//div[contains(text(), 'Удаление товара')]")))
    
    confirm_btn = confirm_modal.find_element(By.XPATH, "//button[text()='Удалить']")
    confirm_btn.click()

    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    wait.until(EC.invisibility_of_element_located((By.XPATH, f"//tr[td[contains(text(), '{sku}')]]")))

def test_create_product_validation(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/products")
    wait = WebDriverWait(driver, 10)

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Добавить товар')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))

    submit_btn = modal.find_element(By.CSS_SELECTOR, "button[type='submit']")
    submit_btn.click()

    form = modal.find_element(By.TAG_NAME, "form")
    assert "was-validated" in form.get_attribute("class")

    assert driver.find_element(By.XPATH, "//div[contains(text(), 'Укажите артикул')]").is_displayed()
    assert driver.find_element(By.XPATH, "//div[contains(text(), 'Укажите название')]").is_displayed()

    assert modal.is_displayed()

def test_history_navigation(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/products")
    wait = WebDriverWait(driver, 10)

    try:
        first_row = wait.until(EC.visibility_of_element_located((By.CSS_SELECTOR, "tbody tr")))
        if "Нет товаров" in first_row.text:
             driver.find_element(By.XPATH, "//button[contains(text(), '+ Добавить товар')]").click()
             modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
             modal.find_element(By.NAME, "sku").send_keys(f"HIST-{uuid.uuid4()}")
             modal.find_element(By.NAME, "name").send_keys("History Test Product")
             modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
             wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
             first_row = driver.find_element(By.CSS_SELECTOR, "tbody tr")
    except:
        pytest.fail("Не удалось найти строку товара для теста истории")

    history_btn = first_row.find_element(By.CSS_SELECTOR, "button .bi-clock-history")
    history_btn.click()

    wait.until(EC.url_contains("/history"))
    
    wait.until(EC.visibility_of_element_located((By.XPATH, "//h2[contains(text(), 'История')]")))
    
    assert driver.find_element(By.ID, "history-tabs-tab-table").is_displayed()
    assert driver.find_element(By.ID, "history-tabs-tab-chart").is_displayed()

    back_btn = driver.find_element(By.XPATH, "//button[contains(text(), 'Назад')]")
    back_btn.click()
    
    wait.until(EC.url_to_be(f"{BASE_URL}/products"))