import re
import time

with open("index.html", "r", encoding="utf-8") as f:
    content = f.read()

new_v = str(int(time.time()))
content = re.sub(r'\?v=\d+', f'?v={new_v}', content)

with open("index.html", "w", encoding="utf-8") as f:
    f.write(content)

print(f"Updated cache buster to ?v={new_v}")
