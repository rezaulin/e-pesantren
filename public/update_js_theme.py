import re

with open('app-core.js', 'r', encoding='utf-8') as f:
    js = f.read()

# Replace hardcoded green gradients with blue gradients in app-core.js
js = re.sub(r'linear-gradient\(135deg,#0c1f17 0%,#0f3d24 50%,#14532d 100%\)', r'linear-gradient(135deg,#0f172a 0%,#1e3a8a 50%,#1e40af 100%)', js)
js = re.sub(r'linear-gradient\(135deg,#064e3b,#0f6b3a\)', r'linear-gradient(135deg,#1e3a8a,#1e40af)', js)
js = re.sub(r'linear-gradient\(135deg,#042f1e,#064e3b 50%,#0a6e3a\)', r'linear-gradient(135deg,#0f172a,#1e3a8a 50%,#2563eb)', js)

# Replace any other hardcoded main greens
js = re.sub(r'rgba\(13,138,69,\.25\)', r'rgba(59,130,246,.25)', js) # 0d8a45

with open('app-core.js', 'w', encoding='utf-8') as f:
    f.write(js)

print("app-core.js updated successfully!")
