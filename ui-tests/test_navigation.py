import pytest
import time
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager

BASE_URL = "http://localhost:5000"

ADMIN_CREDENTIALS = {"username": "admin", "password": "admin"}
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


def login(driver, username, password):
    driver.get(f"{BASE_URL}/login")
    wait = WebDriverWait(driver, 5)
    
    wait.until(EC.visibility_of_element_located((By.NAME, "username"))).send_keys(username)
    driver.find_element(By.NAME, "password").send_keys(password)
    driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    wait.until(EC.url_to_be(f"{BASE_URL}/"))

def is_element_present(driver, by, value):
    try:
        driver.find_element(by, value)
        return True
    except:
        return False

def test_navigation_guest(driver):
    driver.get(BASE_URL)
    
    assert is_element_present(driver, By.CLASS_NAME, "navbar-brand"), "Нет логотипа/бренда"

    login_link = driver.find_element(By.LINK_TEXT, "Вход")
    assert login_link.is_displayed()
    
    assert not is_element_present(driver, By.LINK_TEXT, "Товары"), "Гость видит 'Товары'!"
    assert not is_element_present(driver, By.LINK_TEXT, "Профиль"), "Гость видит 'Профиль'!"

def test_navigation_worker(driver):
    login(driver, WORKER_CREDENTIALS["username"], WORKER_CREDENTIALS["password"])
    
    wait = WebDriverWait(driver, 5)
    wait.until(EC.visibility_of_element_located((By.LINK_TEXT, "Профиль")))

    menu_items = ["Главная", "Товары", "Контрагенты", "Заказы", "Отчеты", "Профиль"]
    for item in menu_items:
        assert is_element_present(driver, By.LINK_TEXT, item), f"Ссылка '{item}' не найдена для сотрудника"

    assert not is_element_present(driver, By.LINK_TEXT, "Пользователи"), "Сотрудник видит админский раздел 'Пользователи'!"
    
    assert not is_element_present(driver, By.LINK_TEXT, "Вход"), "Авторизованный юзер видит ссылку 'Вход'"

def test_navigation_admin(driver):
    login(driver, ADMIN_CREDENTIALS["username"], ADMIN_CREDENTIALS["password"])
    
    wait = WebDriverWait(driver, 5)
    wait.until(EC.visibility_of_element_located((By.LINK_TEXT, "Профиль")))

    try:
        users_link = driver.find_element(By.LINK_TEXT, "Пользователи")
        assert users_link.is_displayed(), "Админ не видит ссылку 'Пользователи'"
    except:
        pytest.fail("Ссылка 'Пользователи' не найдена в навигации админа")

def test_navigation_click_behavior(driver):
    login(driver, ADMIN_CREDENTIALS["username"], ADMIN_CREDENTIALS["password"])
    
    wait = WebDriverWait(driver, 5)

    nav_items = [
        ("Пользователи", "/users", "Пользователи"),
        ("Товары", "/products", "Товары"),
        ("Контрагенты", "/counterparties", "Контрагенты"),
        ("Заказы", "/orders", "Заказы"),
        ("Отчеты", "/reports", "Отчеты"),
        ("Профиль", "/profile", "Профиль пользователя"),
        ("Главная", "/", "Обзор склада")
    ]

    for link_text, url_suffix, header_text in nav_items:
        link = wait.until(EC.element_to_be_clickable((By.LINK_TEXT, link_text)))
        link.click()
        
        wait.until(EC.url_to_be(f"{BASE_URL}{url_suffix}"))
        
        header = wait.until(EC.visibility_of_element_located((By.TAG_NAME, "h2")))
        assert header_text in header.text