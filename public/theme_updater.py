import re

with open('style.css', 'r', encoding='utf-8') as f:
    css = f.read()

# Replace variables
css = re.sub(r'--p:#0a6e3a', r'--p:#2563eb', css)
css = re.sub(r'--p2:#085c30', r'--p2:#1d4ed8', css)
css = re.sub(r'--p3:#13925e', r'--p3:#3b82f6', css)
css = re.sub(r'--pbg:rgba\(10,110,58,\.06\)', r'--pbg:rgba(37,99,235,.08)', css)

css = re.sub(r'--bg:#f5f6f3', r'--bg:#f8fafc', css)
css = re.sub(r'--card:rgba\(255,255,255,\.82\)', r'--card:rgba(255,255,255,.90)', css)
css = re.sub(r'--text:#1a2b23', r'--text:#0f172a', css)
css = re.sub(r'--t2:#3d5a4a', r'--t2:#334155', css)
css = re.sub(r'--t3:#8fa39a', r'--t3:#64748b', css)

css = re.sub(r'--border:rgba\(10,60,30,\.06\)', r'--border:rgba(15,23,42,.06)', css)
css = re.sub(r'--sh:0 2px 12px rgba\(10,60,30,\.05\),0 1px 3px rgba\(0,0,0,\.04\)', r'--sh:0 2px 12px rgba(15,23,42,.04),0 1px 3px rgba(0,0,0,.03)', css)
css = re.sub(r'--sh2:0 4px 20px rgba\(10,60,30,\.08\),0 1px 4px rgba\(0,0,0,\.04\)', r'--sh2:0 4px 20px rgba(15,23,42,.08),0 1px 4px rgba(0,0,0,.04)', css)

css = re.sub(r'--green:#0d8a45', r'--green:#10b981', css)
css = re.sub(r'--g1:linear-gradient\(135deg,#053b28,#0a6e3a\)', r'--g1:linear-gradient(135deg,#1e3a8a,#2563eb)', css)
css = re.sub(r'--g3:linear-gradient\(135deg,#0d8a45,#2dd4a0\)', r'--g3:linear-gradient(135deg,#0ea5e9,#38bdf8)', css)
css = re.sub(r'--sb-bg:linear-gradient\(180deg,#0c1a10 0%,#112618 60%,#0e1a11 100%\)', r'--sb-bg:linear-gradient(180deg,#0f172a 0%,#1e293b 60%,#0f172a 100%)', css)

# Increase border radius to make it softer
css = re.sub(r'--r:16px', r'--r:20px', css)

# Replace hardcoded welcome banner gradients (green to blue)
css = re.sub(r'linear-gradient\(145deg,#042f1e 0%,#064e3b 30%,#0a6e3a 70%,#0d8a45 100%\)', r'linear-gradient(145deg,#0f172a 0%,#1e3a8a 40%,#2563eb 80%,#3b82f6 100%)', css)
css = re.sub(r'radial-gradient\(ellipse at 80% 20%,rgba\(13,138,69,\.3\),transparent 60%\),radial-gradient\(ellipse at 20% 80%,rgba\(4,47,30,\.4\),transparent 50%\)', r'radial-gradient(ellipse at 80% 20%,rgba(59,130,246,.3),transparent 60%),radial-gradient(ellipse at 20% 80%,rgba(30,58,138,.4),transparent 50%)', css)
css = re.sub(r'linear-gradient\(135deg,#042f1e,#064e3b 50%,#0a6e3a\)', r'linear-gradient(135deg,#0f172a,#1e3a8a 50%,#2563eb)', css)
css = re.sub(r'rgba\(13,138,69,\.25\)', r'rgba(59,130,246,.25)', css)
css = re.sub(r'#0f6b3a', r'#1e40af', css)
css = re.sub(r'#064e3b', r'#1e3a8a', css)

# Also update body background gradient
css = re.sub(r'linear-gradient\(170deg,#e6ede8 0%,#f0f3ee 35%,#f5f6f3 65%,#edf1ea 100%\)', r'linear-gradient(170deg,#e2e8f0 0%,#f1f5f9 35%,#f8fafc 65%,#e2e8f0 100%)', css)
css = re.sub(r'linear-gradient\(180deg,#d4edda 0%,#e8f0ea 50%,#f3f5f1 100%\)', r'linear-gradient(180deg,#bfdbfe 0%,#e0f2fe 50%,#f8fafc 100%)', css)

# Update some borders
css = re.sub(r'border-radius:18px', r'border-radius:24px', css)

with open('style.css', 'w', encoding='utf-8') as f:
    f.write(css)

print("CSS updated successfully!")
