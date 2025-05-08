# https://www.geeksforgeeks.org/how-do-i-pass-options-to-the-selenium-chrome-driver-using-python/
# https://stackoverflow.com/questions/5137497/find-the-current-directory-and-files-directory 
# https://stackoverflow.com/questions/5664808/difference-between-webdriver-get-and-webdriver-navigate
# https://selenium-python.readthedocs.io/locating-elements.html
# https://selenium-python.readthedocs.io/navigating.html
# https://stackoverflow.com/questions/32098110/selenium-webdriver-java-need-to-send-space-keypress-to-the-website-as-whol
# https://stackoverflow.com/questions/46361494/how-to-get-the-localstorage-with-python-and-selenium-webdriver
# https://stackoverflow.com/questions/26566799/wait-until-page-is-loaded-with-selenium-webdriver-for-python - wait (copied)

from selenium import webdriver
import sys
import os
from time import sleep
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

cwd = os.getcwd()
sel_folder = os.path.abspath(os.path.join(cwd, "selenium"))
game_url = os.getenv("GAME_URL")

print(f"sel_folder is {sel_folder}")

options = webdriver.ChromeOptions()
for it in sys.argv[1:]:
    print(f"Adding argument {it}")
    options.add_argument(it)



options.add_argument(f"--user-data-dir={sel_folder}")

driver = webdriver.Chrome(options=options)
driver.get(game_url)

# Clear local storage to test auth latencies
#  https://stackoverflow.com/questions/54571696/how-to-hard-refresh-using-selenium/54571878#54571878
#WebDriverWait(driver, 100).until(lambda driver: driver.execute_script('return document.readyState') == 'complete')
driver.execute_script("window.localStorage.clear(); location.reload(true);")

username = driver.find_element(By.CSS_SELECTOR, "#username")
password = driver.find_element(By.CSS_SELECTOR, "#password")
submit = driver.find_element(By.CSS_SELECTOR, ".form-group")

username.send_keys("admin")
password.send_keys("password")
submit.submit()
