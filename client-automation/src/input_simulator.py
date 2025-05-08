# https://www.geeksforgeeks.org/how-do-i-pass-options-to-the-selenium-chrome-driver-using-python/
# https://stackoverflow.com/questions/5137497/find-the-current-directory-and-files-directory 
# https://stackoverflow.com/questions/5664808/difference-between-webdriver-get-and-webdriver-navigate
# https://selenium-python.readthedocs.io/locating-elements.html
# https://selenium-python.readthedocs.io/navigating.html
# https://stackoverflow.com/questions/32098110/selenium-webdriver-java-need-to-send-space-keypress-to-the-website-as-whol
# https://stackoverflow.com/questions/46361494/how-to-get-the-localstorage-with-python-and-selenium-webdriver
# https://stackoverflow.com/questions/26566799/wait-until-page-is-loaded-with-selenium-webdriver-for-python - wait (copied)
# https://stackoverflow.com/questions/14257373/how-to-skip-the-headers-when-processing-a-csv-file-using-python
# https://smithay.github.io/smithay/smithay/backend/libinput/struct.LibinputInputBackend.html
# https://stackoverflow.com/questions/62501219/how-to-send-keys-to-a-canvas-element-for-longer-duration

from selenium import webdriver
import sys
import os
from time import sleep
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.support import expected_conditions as EC

import csv

cwd = os.getcwd()
sel_folder = os.path.abspath(os.path.join(cwd, "selenium"))
game_url = os.getenv("GAME_URL")
input_csv = os.getenv("INPUT_CSV")
jump_times = []

print(f"sel_folder is {sel_folder}")
print(f"input_csv is {input_csv}")

with open(input_csv, newline='') as csvfile:
    reader = csv.reader(csvfile)
    next(reader, None)
    # skip header row
    for row in reader:
        jump_times.append(int(row[1]))

# transform jump times
jump_times = [i - jump_times[0] for i in jump_times]
print(jump_times)

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

# https://stackoverflow.com/questions/59130200/selenium-wait-until-element-is-present-visible-and-interactable
WebDriverWait(driver, 20).until(EC.presence_of_element_located((By.CSS_SELECTOR, "#jump-instruction")))

# turn the input into an action chain.
chain = ActionChains(driver)
for i, time in enumerate(jump_times[:-1]):
    chain.key_down(Keys.SPACE)
    chain.key_up(Keys.SPACE)

    pause_time = (jump_times[i+1] - time) / 1000
    print(f"pausing for {pause_time}")

    # TODO how to configure this?
    chain.pause(pause_time)

chain.perform()

# Now collect relevant stats: score
