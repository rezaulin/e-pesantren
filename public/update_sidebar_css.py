import re

with open('style.css', 'r', encoding='utf-8') as f:
    css = f.read()

# Change sidebar background
css = re.sub(
    r'--sb-bg:linear-gradient\(180deg,#0f172a 0%,#1e293b 60%,#0f172a 100%\);',
    r'--sb-bg:linear-gradient(180deg,#1e3a8a 0%,#1e40af 100%);',
    css
)

# Change sidebar active state from green to light blue
css = css.replace('rgba(16,163,74,.12)', 'rgba(56,189,248,.15)')
css = css.replace('color:#4ade80', 'color:#38bdf8')
css = css.replace('border-left-color:#4ade80', 'border-left-color:#38bdf8')
css = css.replace('rgba(74,222,128,.08)', 'rgba(56,189,248,.1)')

# Change sidebar badge from green to light blue
css = css.replace('background:rgba(16,163,74,.2);color:#4ade80;border-radius:10px;font-size:.62rem;font-weight:700;letter-spacing:.05em;border:1px solid rgba(74,222,128,.1)', 'background:rgba(56,189,248,.2);color:#38bdf8;border-radius:10px;font-size:.62rem;font-weight:700;letter-spacing:.05em;border:1px solid rgba(56,189,248,.1)')

# Add responsive classes
if '.hide-on-pc' not in css:
    css += '\n/* Responsive Display Classes */\n@media(min-width: 900px) { .hide-on-pc { display: none !important; } }\n@media(max-width: 899px) { .hide-on-mobile { display: none !important; } }\n'

with open('style.css', 'w', encoding='utf-8') as f:
    f.write(css)

print("style.css updated for elegant blue sidebar and responsive classes.")
