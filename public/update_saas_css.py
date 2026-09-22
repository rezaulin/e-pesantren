import re

with open('style.css', 'r', encoding='utf-8') as f:
    css = f.read()

# Change sidebar background back to white/light
css = re.sub(
    r'--sb-bg:linear-gradient\(180deg,#1e3a8a 0%,#1e40af 100%\);',
    r'--sb-bg:#ffffff;',
    css
)

# Revert the text colors in sidebar to dark (because it's white background now)
css = css.replace('.sb-menu a{padding:.6rem 1.2rem;display:flex;align-items:center;gap:.7rem;color:rgba(255,255,255,0.7);text-decoration:none;font-size:.85rem;font-weight:500;transition:.2s;border-left:3px solid transparent}', '.sb-menu a{padding:.6rem 1.2rem;display:flex;align-items:center;gap:.7rem;color:var(--t2);text-decoration:none;font-size:.85rem;font-weight:600;transition:.2s;border-left:3px solid transparent}')
css = css.replace('.sb-menu a i{font-size:1.1rem;color:rgba(255,255,255,0.5)}', '.sb-menu a i{font-size:1.2rem;color:var(--t3)}')
css = css.replace('.sb-section{font-size:.65rem;font-weight:700;color:rgba(255,255,255,0.4);text-transform:uppercase;letter-spacing:.05em;margin:1.2rem 1.2rem .4rem;display:block}', '.sb-section{font-size:.65rem;font-weight:800;color:var(--t3);text-transform:uppercase;letter-spacing:.05em;margin:1.2rem 1.2rem .4rem;display:block}')

# Sidebar hover & active for SaaS look (blue background for active)
css = css.replace('.sb-menu a:hover{background:rgba(255,255,255,0.1);color:#fff}', '.sb-menu a:hover{background:var(--bg);color:var(--text)}')
css = css.replace('.sb-menu a:hover i{color:#fff}', '.sb-menu a:hover i{color:var(--p)}')
css = css.replace('.sb-menu a.active{background:rgba(56,189,248,.15);color:#38bdf8;font-weight:700;border-left-color:#38bdf8;}', '.sb-menu a.active{background:var(--p);color:#fff;font-weight:700;border-left-color:transparent;border-radius:12px;margin:0 0.8rem;padding:.6rem .8rem;}')
css = css.replace('.sb-menu a.active i{color:var(--p)}', '.sb-menu a.active i{color:#fff}')

# Adjust sidebar header (blue background)
css = re.sub(
    r'\.sb-head\{padding:1\.4rem 1\.2rem;background:transparent;border-bottom:1px solid var\(--border\);.*?\}',
    r'.sb-head{padding:1.4rem 1.2rem;background:linear-gradient(135deg, #1e3a8a, #2563eb);color:#fff;border-radius:0 0 20px 20px;margin-bottom:1rem;display:flex;align-items:center;gap:0.8rem}',
    css,
    flags=re.DOTALL
)
css = css.replace('.sb-head .sb-name{font-size:.95rem;font-weight:700;color:var(--text);letter-spacing:-.01em}', '.sb-head .sb-name{font-size:.95rem;font-weight:800;color:#fff;letter-spacing:-.01em}')
css = css.replace('.sb-head .sb-sub{font-size:.65rem;color:var(--t3);font-weight:500}', '.sb-head .sb-sub{font-size:.65rem;color:rgba(255,255,255,0.7);font-weight:500}')

# Add promo banner CSS
promo_css = """
/* Sidebar Promo Banner */
.sb-promo { background: #f8fafc; border-radius: 16px; padding: 1.2rem; margin: 1.5rem 1rem 1rem; border: 1px solid var(--border); text-align: center; }
.sbp-icon { width: 40px; height: 40px; background: rgba(37,99,235,0.1); color: var(--p); border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 1.4rem; margin: 0 auto 0.8rem; }
.sbp-title { font-size: 0.75rem; font-weight: 700; color: var(--text); line-height: 1.3; margin-bottom: 0.3rem; }
.sbp-sub { font-size: 0.65rem; color: var(--t3); line-height: 1.4; margin-bottom: 0.8rem; }
.sbp-btn { background: #fff; border: 1px solid var(--border); border-radius: 8px; padding: 0.4rem 0.8rem; font-size: 0.7rem; font-weight: 600; color: var(--text); cursor: pointer; transition: 0.2s; box-shadow: 0 2px 5px rgba(0,0,0,0.02); }
.sbp-btn:hover { border-color: var(--p); color: var(--p); }

/* Header adjustments */
.header { background: #fff; border-bottom: 1px solid var(--border); }
@media(min-width: 900px) {
  .header { box-shadow: 0 4px 20px rgba(0,0,0,0.02); margin-left: calc(var(--sw) + 14px); left: 0; width: calc(100% - var(--sw) - 14px); top: 0; border-radius: 0; }
  .main { margin-top: 70px; margin-left: calc(var(--sw) + 14px); }
  .dash-hero { margin-top: 0 !important; border-radius: 24px !important; margin-left: 0 !important; margin-right: 0 !important; }
}

/* Quick Action Pills */
.pill-grid { display: flex; flex-wrap: wrap; gap: 0.8rem; margin-bottom: 2rem; }
.pill-card { background: #fff; border: 1px solid var(--border); border-radius: 16px; padding: 0.8rem 1.2rem; display: flex; align-items: center; gap: 0.8rem; cursor: pointer; transition: 0.3s; box-shadow: 0 4px 12px rgba(0,0,0,0.02); }
.pill-card:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.05); border-color: var(--pbg); }
.pill-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 1.2rem; }
.pill-text { font-size: 0.82rem; font-weight: 700; color: var(--text); }

/* 3 Column Bottom Layout */
.bottom-cols { display: grid; grid-template-columns: 1fr 1fr 1.2fr; gap: 1.5rem; }
@media(max-width: 1100px) { .bottom-cols { grid-template-columns: 1fr 1fr; } }
@media(max-width: 768px) { .bottom-cols { grid-template-columns: 1fr; } }

.list-item { display: flex; align-items: center; gap: 1rem; padding: 1rem 0; border-bottom: 1px solid var(--border); }
.list-item:last-child { border-bottom: none; }
.li-icon { width: 40px; height: 40px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 1.2rem; flex-shrink: 0; }
.li-body { flex: 1; }
.li-title { font-size: 0.82rem; font-weight: 700; color: var(--text); margin-bottom: 0.1rem; }
.li-sub { font-size: 0.7rem; color: var(--t3); }
.li-right { font-size: 0.7rem; font-weight: 700; }
.li-badge { padding: 0.2rem 0.5rem; border-radius: 6px; background: rgba(239,68,68,0.1); color: #ef4444; }

/* Stat Cards PC Refined */
.stat-card-clean { background: #fff; border: 1px solid var(--border); border-radius: 20px; padding: 1.2rem; display: flex; align-items: center; gap: 1rem; box-shadow: 0 4px 15px rgba(0,0,0,0.02); }
.sc-icon { width: 48px; height: 48px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 1.4rem; }
.sc-body { flex: 1; }
.sc-label { font-size: 0.65rem; font-weight: 800; color: var(--t3); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.2rem; }
.sc-val { font-size: 1.4rem; font-weight: 800; color: var(--text); line-height: 1.1; }
.sc-sub { font-size: 0.65rem; color: var(--t3); margin-top: 0.3rem; }
"""
if 'Sidebar Promo Banner' not in css:
    css += promo_css

with open('style.css', 'w', encoding='utf-8') as f:
    f.write(css)

print("style.css updated for Light SaaS structure.")
