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
ADMIN_CREDENTIALS = {"username": "admin", "password": "admin"}

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

def login_as_admin(driver):
    driver.get(f"{BASE_URL}/login")
    wait = WebDriverWait(driver, 5)
    
    wait.until(EC.visibility_of_element_located((By.NAME, "username"))).send_keys(ADMIN_CREDENTIALS["username"])
    driver.find_element(By.NAME, "password").send_keys(ADMIN_CREDENTIALS["password"])
    driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    wait.until(EC.url_to_be(f"{BASE_URL}/"))

def test_user_lifecycle(driver):
    login_as_admin(driver)
    wait = WebDriverWait(driver, 10)

    users_link = wait.until(EC.element_to_be_clickable((By.LINK_TEXT, "Пользователи")))
    users_link.click()
    wait.until(EC.url_to_be(f"{BASE_URL}/users"))

    unique_id = str(uuid.uuid4())[:8]
    new_username = f"test_user_{unique_id}"
    new_email = f"test_{unique_id}@example.com"
    new_fullname = f"Test User {unique_id}"

    create_btn = wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Создать пользователя')]")))
    create_btn.click()

    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.NAME, "username").send_keys(new_username)
    modal.find_element(By.NAME, "full_name").send_keys(new_fullname)
    modal.find_element(By.NAME, "email").send_keys(new_email)
    
    role_select = Select(modal.find_element(By.NAME, "role"))
    role_select.select_by_value("worker")

    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))

    try:
        user_row = wait.until(EC.presence_of_element_located((By.XPATH, f"//tr[td[text()='{new_username}']]")))
        assert new_email in user_row.text
        assert "Сотрудник" in user_row.text 
    except:
        pytest.fail(f"Созданный пользователь {new_username} не найден в таблице")

    delete_btn = user_row.find_element(By.CSS_SELECTOR, "button .bi-trash")
    delete_btn.click()

    delete_modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-dialog")))
    
    confirm_delete_btn = delete_modal.find_element(By.XPATH, "//button[text()='Удалить']")
    confirm_delete_btn.click()

    wait.until(EC.invisibility_of_element_located((By.CLASS_NAME, "modal-backdrop")))
    wait.until(EC.invisibility_of_element_located((By.XPATH, f"//tr[td[text()='{new_username}']]")))

    rows = driver.find_elements(By.XPATH, f"//tr[td[text()='{new_username}']]")
    assert len(rows) == 0, "Пользователь не был удален из таблицы"

def test_create_user_validation(driver):
    login_as_admin(driver)
    wait = WebDriverWait(driver, 5)

    driver.get(f"{BASE_URL}/users")
    
    wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(), '+ Создать пользователя')]"))).click()
    
    modal = wait.until(EC.visibility_of_element_located((By.CLASS_NAME, "modal-content")))
    
    modal.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    form = modal.find_element(By.TAG_NAME, "form")
    assert "was-validated" in form.get_attribute("class"), "Класс валидации не применился к форме"
    
    assert modal.is_displayed(), "Модальное окно закрылось, хотя данные не были введены"