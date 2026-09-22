import re
import os

with open('style.css', 'r', encoding='utf-8') as f:
    css = f.read()

# Fix .sb-head
css = re.sub(r'\.sb-head\{.*?\}', r'.sb-head{padding:1.4rem 1.2rem;background:linear-gradient(135deg, #1e3a8a, #2563eb);display:flex;align-items:center;gap:.8rem;min-height:var(--hh);color:#fff;border-radius:0;margin:0;}', css, flags=re.DOTALL)

# Fix .sidebar (Mobile)
css = re.sub(r'\.sidebar\{.*?\}', r'.sidebar{position:fixed;left:0;top:0;width:var(--sw);height:100vh;z-index:100;background:var(--sb-bg);border-right:1px solid var(--border);display:flex;flex-direction:column;overflow-y:auto;transform:translateX(-100%);transition:transform .3s cubic-bezier(.4,0,.2,1);box-shadow:4px 0 24px rgba(0,0,0,.03)}', css, flags=re.DOTALL, count=1)

# Fix desktop media query
def repl_media(m):
    block = m.group(0)
    block = re.sub(r'\.sidebar\{.*?\}', r'.sidebar{transform:none !important; left:0; top:0; bottom:0; height:100vh; border-radius:0; width:var(--sw); border-right:1px solid var(--border); box-shadow:4px 0 24px rgba(0,0,0,.03)}', block, flags=re.DOTALL)
    block = re.sub(r'\.header\s*\{.*?\}', r'.header{left:var(--sw); width:calc(100% - var(--sw)); border-bottom:1px solid var(--border); box-shadow:none; background:#fff; margin:0; top:0; border-radius:0;}', block, flags=re.DOTALL)
    return block

css = re.sub(r'@media\s*\(min-width:\s*900px\)\s*\{.*?\n\}', repl_media, css, flags=re.DOTALL)

with open('style.css', 'w', encoding='utf-8') as f:
    f.write(css)

# Fix JS
with open('app-core.js', 'r', encoding='utf-8') as f:
    js_core = f.read()
js_core = re.sub(r'\$\(\'sidebarName\'\)\.textContent\s*=\s*appName;', r'$(\'sidebarName\').textContent = \'Pesantren Digital\';', js_core)
js_core = re.sub(r'\$\(\'hTitle\'\)\.textContent\s*=\s*appName;', r'$(\'hTitle\').textContent = \'Pesantren Digital\';', js_core)
# Fix bad replacement from multi_replace
js_core = re.sub(r'\$\(\'sidebarName\'\)\.textContent=\'Pesantren Digital\';\s*\}', r'$(\'sidebarRole\').textContent=user.role.toUpperCase();\n}', js_core)
# Fix bad replacement on pgSubtotal
js_core = re.sub(r'\$\(\'hTitle\'\)\.textContent=\'Pesantren Digital\';\s*\$\(\'pgFee\'\)', r'$(\'pgFee\')', js_core)

with open('app-core.js', 'w', encoding='utf-8') as f:
    f.write(js_core)

with open('app-extra.js', 'r', encoding='utf-8') as f:
    js_ext = f.read()
js_ext = re.sub(r'\$\(\'sidebarName\'\)\.textContent\s*=\s*verify\.app_name\s*\|\|\s*\'Pesantren\';', r'$(\'sidebarName\').textContent = \'Pesantren Digital\';', js_ext)
js_ext = re.sub(r'\$\(\'hTitle\'\)\.textContent\s*=\s*verify\.app_name\s*\|\|\s*\'Pesantren\';', r'$(\'hTitle\').textContent = \'Pesantren Digital\';', js_ext)
with open('app-extra.js', 'w', encoding='utf-8') as f:
    f.write(js_ext)

print('Fixed CSS and JS via script')
