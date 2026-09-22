with open('style.css', 'a', encoding='utf-8') as f:
    f.write("""
/* Premium Overhaul */
@media(min-width: 900px){ 
  .bottom-nav { display: none !important; } 
  .main { padding-bottom: 2rem !important; } 
}
.premium-hero { background: linear-gradient(135deg, #1e3a8a, #3b82f6); border-radius: 24px; padding: 2rem; color: #fff; position: relative; overflow: hidden; margin-bottom: 1.5rem; box-shadow: 0 12px 30px rgba(37,99,235,0.2); }
.premium-hero::before { content:''; position:absolute; inset:0; background: radial-gradient(circle at top right, rgba(255,255,255,0.2), transparent 50%); pointer-events:none; }
.premium-hero h3 { font-size: 1.5rem; font-weight: 800; margin-bottom: 0.5rem; position:relative; z-index:1; }
.premium-hero p { font-size: 0.9rem; opacity: 0.9; margin-bottom: 1.2rem; position:relative; z-index:1; }
.quick-action-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
.qa-card { background: var(--card-solid); border-radius: 20px; padding: 1.2rem 1rem; text-align: center; border: 1px solid var(--border); box-shadow: var(--sh); transition: 0.3s; cursor: pointer; display: flex; flex-direction: column; align-items: center; }
.qa-card:hover { transform: translateY(-3px); box-shadow: 0 12px 24px rgba(0,0,0,0.06); border-color: rgba(37,99,235,0.3); }
.qa-icon-wrap { width: 56px; height: 56px; border-radius: 16px; display: flex; align-items: center; justify-content: center; font-size: 1.8rem; margin: 0 auto 0.8rem; }
.qa-title { font-size: 0.8rem; font-weight: 700; color: var(--text); }
.qa-sub { font-size: 0.65rem; color: var(--t3); margin-top: 0.2rem; }
.qa-stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
@media(max-width: 768px) { .qa-stat-grid { grid-template-columns: repeat(2, 1fr); } }
""")
print("Premium CSS appended.")
