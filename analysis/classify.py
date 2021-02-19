from selenium import webdriver
from selenium.webdriver.support.wait import WebDriverWait
import selenium.webdriver.support.expected_conditions as EC
from selenium.webdriver.common.by import By
import re
import os
import time

DIR = 'bytecode'
PWD = os.getcwd()

filepath = "/home/quentin/p/scrapingwasm/new/analysis/bytecode/01fc1bcf0957d51633f99345f821111f57ea37486f7f7dfbb9b346b52ee5e46f.wasm"
url = "http://localhost:4000"
driver = webdriver.Firefox()

for f in os.listdir(DIR):
    filepath = PWD + '/' + DIR + '/' + f
    driver.get(url)

    wasmfile = driver.find_element_by_name("wasm-file")
    wasmfile.send_keys(filepath)
    button = driver.find_element_by_class_name("btn-primary")
    button.click()

    time.sleep(1) # TODO: wait until element is present
    text = WebDriverWait(driver, 20).until(EC.visibility_of_element_located((By.CSS_SELECTOR, ".jumbotron > dl > dd > div > h4"))).text

    # text = driver.find_element_by_css_selector(".jumbotron > dl > dd > div > h4").text

    results = re.findall('(.*) \(([0-9.]*)%\)', text)
    for (guess, percentage) in results:
        rounded = round(float(percentage))
        print(f'{f} {guess} {rounded}')
