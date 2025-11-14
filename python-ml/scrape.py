import undetected_chromedriver as uc
from selenium_stealth import stealth
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time

def scrape_website(website):
    if not website.startswith("http"):
        website = "https://" + website

    print(f"[DEBUG] Opening {website}")

    options = uc.ChromeOptions()
    options.headless = True
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--disable-gpu")
    options.add_argument("--disable-blink-features=AutomationControlled")
    options.binary_location = "/usr/bin/chromium"

    driver = uc.Chrome(options=options)

    stealth(driver,
        languages=["en-US", "en"],
        vendor="Google Inc.",
        platform="Win32",
        webgl_vendor="Intel Inc.",
        renderer="Intel Iris OpenGL Engine",
        fix_hairline=True,
    )

    html = ""
    try:
        driver.get(website)
        WebDriverWait(driver, 30).until(
            EC.presence_of_element_located((By.TAG_NAME, "body"))
        )
        time.sleep(5)
        html = driver.page_source
    except Exception as e:
        print(f"[ERROR] Scrape failed: {e}")
    finally:
        driver.quit()

    print(f"[DEBUG SCRAPE] domain={website} html_len={len(html)}")
    return html
