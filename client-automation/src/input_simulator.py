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
from datetime import datetime


import csv
import time
import json
import shutil



# https://stackoverflow.com/questions/1133857/how-accurate-is-pythons-time-sleep
def sleep(duration, get_now=time.perf_counter):
    now = get_now()
    end = now + duration
    while now < end:
        now = get_now()

cwd = os.getcwd()
sel_folder = os.path.abspath(os.path.join(cwd, "selenium"))
# https://www.geeksforgeeks.org/python-strftime-function/
stat_dir = os.getenv("STAT_DIR") if os.getenv("STAT_DIR") else os.path.abspath(os.path.join("..", "data", "client", datetime.now().strftime("%Y%m%d_%H%M%S")))
print(f"outputting to {stat_dir}")
os.makedirs(stat_dir)
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
options.add_argument("--ignore-certificate-errors")

# https://stackoverflow.com/questions/71716460/how-to-change-download-directory-location-path-in-selenium-using-chrome
prefs = {'download.default_directory': stat_dir}
options.add_experimental_option("prefs", prefs)

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
chain = ActionChains(driver).key_down(Keys.SPACE).key_up(Keys.SPACE)
body = driver.find_element(By.CSS_SELECTOR, "body")
for i, cur_time in enumerate(jump_times[:-1]):
    #print("Sending space")
    #body.send_keys(Keys.SPACE)
    #https://stackoverflow.com/questions/62501219/how-to-send-keys-to-a-canvas-element-for-longer-duration
    driver.execute_script('''
    var keydownEvt = new KeyboardEvent('keydown', {
        altKey:false,
        altKey: false,
        bubbles: true,
        cancelBubble: false,
        cancelable: true,
        charCode: 0,
        code: "Space",
        composed: true,
        ctrlKey: false,
        currentTarget: null,
        defaultPrevented: true,
        detail: 0,
        eventPhase: 0,
        isComposing: false,
        isTrusted: true,
        key: " ",
        keyCode: 32,
        location: 0,
        metaKey: false,
        repeat: false,
        returnValue: false,
        shiftKey: false,
        type: "keydown",
        which: 32,
    });
    arguments[0].dispatchEvent(keydownEvt);
    ''', body)

    # TODO how to configure this?
    pause_time = (jump_times[i+1] - cur_time) / 1000
    # sleep(pause_time - 0.012)
    get_now = time.perf_counter
    now = get_now()
    end = now + pause_time - 0.011
    to_break = False
    while now < end:
        now = get_now()
        if len(driver.find_elements(By.CSS_SELECTOR, "#past-scores > p")) > 0:
            to_break = True
    if to_break:
        break

    # To test failures
    # sleep(pause_time + 0.1)



# Now collect relevant stats: score
# gets downloaded to 

# horrible coding
while not driver.execute_script('return window.gameOverScreenShown'):
    pass


score = driver.find_element(By.CSS_SELECTOR, "#final-score")
score_num = int(score.get_attribute('innerHTML'))
auth_latency = driver.execute_script("return window.authLatency.toFixed(3);")

with open(os.path.join(stat_dir, 'extra_data.json'), 'w') as extra_f:
    json.dump({'score': score_num, 'auth_latency': auth_latency, 'seed': os.path.basename(input_csv).split(".")[0]}, extra_f)

print(auth_latency)
print(score_num)
