import pytest
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import StaleElementReferenceException
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

def login(driver):
    driver.get(f"{BASE_URL}/login")
    wait = WebDriverWait(driver, 5)
    
    wait.until(EC.visibility_of_element_located((By.NAME, "username"))).send_keys(WORKER_CREDENTIALS["username"])
    driver.find_element(By.NAME, "password").send_keys(WORKER_CREDENTIALS["password"])
    driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
    
    wait.until(EC.url_to_be(f"{BASE_URL}/"))

def test_dashboard_basic_elements(driver):
    login(driver)
    wait = WebDriverWait(driver, 10)

    header = wait.until(EC.visibility_of_element_located((By.XPATH, "//h2[contains(text(), 'Обзор склада')]")))
    assert header.is_displayed()

    welcome_msg = driver.find_element(By.XPATH, "//div[contains(text(), 'Добро пожаловать')]")
    assert welcome_msg.is_displayed()

    chart_title = driver.find_element(By.XPATH, "//h5[contains(text(), 'Динамика заказов')]")
    assert chart_title.is_displayed()

    status_card = driver.find_element(By.XPATH, "//span[contains(text(), 'Статус системы')]")
    assert status_card.is_displayed()
    
    online_badge = driver.find_element(By.XPATH, "//span[contains(text(), 'Online')]")
    assert "text-success" in online_badge.get_attribute("class")

def test_kpi_cards_load_data(driver):
    login(driver)
    wait = WebDriverWait(driver, 10)

    card_titles = ["Всего товаров", "Всего заказов", "Контрагенты"]

    for title in card_titles:
        wait_stale = WebDriverWait(driver, 10, ignored_exceptions=(StaleElementReferenceException,))
        
        def data_is_loaded(d):
            title_el = d.find_element(By.XPATH, f"//div[contains(text(), '{title}')]")
            value_container = title_el.find_element(By.XPATH, "./following-sibling::div")
            
            txt = value_container.text.strip()
            html = value_container.get_attribute("innerHTML")
            
            if txt != "" and "spinner" not in html:
                return value_container
            return False

        value_element = wait_stale.until(data_is_loaded)
        
        value = value_element.text
        assert value.replace(',', '').isdigit() or value == "0", f"В карточке {title} некорректное значение: {value}"

def test_kpi_cards_navigation(driver):
    login(driver)
    wait = WebDriverWait(driver, 10)

    kpi_map = [
        ("Всего товаров", "/products"),
        ("Всего заказов", "/orders"),
        ("Контрагенты", "/counterparties"),
        ("Отчеты", "/reports")
    ]

    for title, expected_route in kpi_map:
        driver.get(f"{BASE_URL}/")
        
        def card_is_ready_and_clickable(d):
            try:
                card = d.find_element(By.XPATH, f"//div[contains(text(), '{title}')]/ancestor::div[contains(@class, 'card')]")
                
                if "spinner" in card.get_attribute("innerHTML"):
                    return False
                
                return card
            except StaleElementReferenceException:
                return False

        card = wait.until(card_is_ready_and_clickable)
        
        card.click()
        
        wait.until(EC.url_contains(expected_route))
        assert driver.current_url == f"{BASE_URL}{expected_route}"

def test_quick_actions_navigation(driver):
    login(driver)
    wait = WebDriverWait(driver, 10)

    actions_map = [
        ("Создать заказ", "/orders"),
        ("Добавить товар", "/products"),
        ("Скачать отчет", "/reports")
    ]

    for btn_text, expected_route in actions_map:
        driver.get(f"{BASE_URL}/")
        
        def btn_is_ready(d):
            try:
                btn = d.find_element(By.XPATH, f"//div[contains(text(), '{btn_text}')]/ancestor::button")
                return btn
            except StaleElementReferenceException:
                return False

        button = wait.until(btn_is_ready)
        button.click()
        
        wait.until(EC.url_contains(expected_route))
        assert driver.current_url == f"{BASE_URL}{expected_route}"