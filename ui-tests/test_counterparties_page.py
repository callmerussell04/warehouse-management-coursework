import pytest
import uuid
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import Select
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

def test_counterparties_page_structure(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/counterparties")
    wait = WebDriverWait(driver, 10)

    header = wait.until(EC.visibility_of_element_located((By.XPATH, "//h2[contains(text(), 'Контрагенты')]")))
    assert header.is_displayed()

    add_btn = driver.find_element(By.XPATH, "//button[contains(text(), '+ Добавить')]")
    assert add_btn.is_displayed()

    table_headers = ["Имя / Название", "Тип", "Email", "Телефон", "Действия"]
    for h in table_headers:
        assert driver.find_element(By.XPATH, f"//th[contains(text(), '{h}')]").is_displayed()

def test_counterparty_crud_lifecycle_client(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/counterparties")
    wait = WebDriverWait(driver, 10)

    unique_id = str(uuid.uuid4())[:8]
    name = f"Client {unique_id}"
    email = f"client_{unique_id}@test.com"
    phone = "+7 (999) 111-22-33"
    new_name = f"Edited Client {unique_id}"

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Добавить')]"))).click()
    
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.NAME, "name").send_keys(name)
    modal.find_element(By.NAME, "email").send_keys(email)
    modal.find_element(By.NAME, "phone_number").send_keys(phone)
    
    type_select = Select(modal.find_element(By.NAME, "type"))
    type_select.select_by_value("client")

    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    row = wait.until(EC.visibility_of_element_located((By.XPATH, f"//tr[td[contains(text(), '{name}')]]")))
    assert email in row.text
    assert phone in row.text
    assert "Клиент" in row.text

    edit_btn = row.find_element(By.CSS_SELECTOR, "button .bi-pencil")
    edit_btn.click()

    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    type_input = modal.find_element(By.NAME, "type")
    assert not type_input.is_enabled()

    name_input = modal.find_element(By.NAME, "name")
    name_input.clear()
    name_input.send_keys(new_name)
    
    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    wait.until(EC.visibility_of_element_located((By.XPATH, f"//tr[td[contains(text(), '{new_name}')]]")))

    row = driver.find_element(By.XPATH, f"//tr[td[contains(text(), '{new_name}')]]")
    delete_btn = row.find_element(By.CSS_SELECTOR, "button .bi-trash")
    delete_btn.click()

    confirm_modal = wait.until(EC.visibility_of_element_located((By.XPATH, "//div[contains(@class, 'modal-content')]//div[contains(text(), 'Удаление контрагента')]")))
    confirm_modal.find_element(By.XPATH, "//button[text()='Удалить']").click()

    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    wait.until(EC.invisibility_of_element_located((By.XPATH, f"//tr[td[contains(text(), '{new_name}')]]")))

def test_create_supplier_badge_check(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/counterparties")
    wait = WebDriverWait(driver, 10)

    unique_id = str(uuid.uuid4())[:8]
    name = f"Supplier {unique_id}"

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Добавить')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.NAME, "name").send_keys(name)
    Select(modal.find_element(By.NAME, "type")).select_by_value("supplier")
    
    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    row = wait.until(EC.visibility_of_element_located((By.XPATH, f"//tr[td[contains(text(), '{name}')]]")))
    
    badge = row.find_element(By.XPATH, ".//span[contains(@class, 'badge')]")
    assert "Поставщик" in badge.text
    assert "bg-warning" in badge.get_attribute("class")

    row.find_element(By.CSS_SELECTOR, "button .bi-trash").click()
    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[text()='Удалить']"))).click()

def test_counterparty_validation(driver):
    login_as_worker(driver)
    driver.get(f"{BASE_URL}/counterparties")
    wait = WebDriverWait(driver, 10)

    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Добавить')]"))).click()
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))

    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

    form = modal.find_element(By.TAG_NAME, "form")
    assert "was-validated" in form.get_attribute("class")

    name_input = modal.find_element(By.NAME, "name")
    assert name_input.get_attribute("validity") != "" 

    assert modal.is_displayed()