import pytest
import time
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager

BASE_URL = "http://localhost"
ADMIN_USERNAME = "admin"
ADMIN_PASSWORD = "admin"

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

def test_admin_login_success(driver):
    driver.get(f"{BASE_URL}/login")

    wait = WebDriverWait(driver, 10)

    username_field = wait.until(EC.visibility_of_element_located((By.NAME, "username")))
    password_field = driver.find_element(By.NAME, "password")
    submit_button = driver.find_element(By.CSS_SELECTOR, "button[type='submit']")

    username_field.clear()
    username_field.send_keys(ADMIN_USERNAME)

    password_field.clear()
    password_field.send_keys(ADMIN_PASSWORD)

    submit_button.click()

    try:
        wait.until(EC.url_to_be(f"{BASE_URL}/"))
    except:
        pytest.fail(f"Не произошел редирект на главную. Текущий URL: {driver.current_url}")

    try:
        profile_link = wait.until(EC.presence_of_element_located((By.LINK_TEXT, "Профиль")))
        assert profile_link.is_displayed(), "Ссылка на профиль не отображается"
        
        users_link = driver.find_element(By.LINK_TEXT, "Пользователи")
        assert users_link.is_displayed(), "Ссылка 'Пользователи' (для админа) отсутствует"
        
    except Exception as e:
        pytest.fail(f"Элементы интерфейса после входа не найдены: {str(e)}")

def test_login_invalid_credentials(driver):
    driver.get(f"{BASE_URL}/login")
    wait = WebDriverWait(driver, 5)

    wait.until(EC.visibility_of_element_located((By.NAME, "username"))).send_keys("admin")
    driver.find_element(By.NAME, "password").send_keys("wrong_password")
    driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()

    time.sleep(1)
    assert driver.current_url == f"{BASE_URL}/login", "Система пропустила с неверным паролем!"
    
    try:
        toast = wait.until(EC.presence_of_element_located((By.XPATH, "//*[contains(text(), 'Error') or contains(text(), 'ошибка') or contains(text(), 'Неверные')]")))
        assert toast.is_displayed()
    except:
        pass
