import re

with open('app-core.js', 'r', encoding='utf-8') as f:
    js = f.read()

start_marker = "// -- Admin Dashboard (SIPP Style) --"
end_marker = "// Waktu page (server time + jadwal)"

start_idx = js.find(start_marker)
end_idx = js.find(end_marker)

if start_idx == -1 or end_idx == -1:
    print("Could not find markers!")
    exit(1)

premium_logic = """// -- Admin Dashboard (SIPP Style) --
  const hf=hasFeature;
  const totalAbsen=(d.hadir_hari_ini||0)+(d.izin_sakit||0)+(d.alfa||0);
  const pctHadir=totalAbsen>0?Math.round(((d.hadir_hari_ini||0)/totalAbsen)*100):0;

  const qaCard = (id, ic, lb, sub, bg, fg) => `<div class="qa-card" onclick="nav('${id}')"><div class="qa-icon-wrap" style="background:${bg};color:${fg}"><i class="${ic}"></i></div><div class="qa-title">${lb}</div><div class="qa-sub">${sub}</div></div>`;

  let qaHtml = '';
  qaHtml += qaCard('santri','ph-duotone ph-users-three','Data Santri','Kelola profil','#eff6ff','#1e40af');
  if(hf('absen_sekolah')) qaHtml += qaCard('absen-sekolah','ph-duotone ph-check-circle','Absensi','Sekolah','#ecfdf5','#065f46');
  if(hf('absen_malam')) qaHtml += qaCard('kamar','ph-duotone ph-bed','Kamar','Asrama','#eef2ff','#3730a3');
  if(hf('keuangan')) qaHtml += qaCard('pembayaran','ph-duotone ph-wallet','Pembayaran','Keuangan','#ecfdf5','#065f46');
  if(hf('kedisiplinan')) qaHtml += qaCard('pelanggaran','ph-duotone ph-warning-circle','Pelanggaran','Kedisiplinan','#fef2f2','#991b1b');
  qaHtml += qaCard('psb','ph-duotone ph-user-plus','PSB Online','Penerimaan','#ecfdf5','#065f46');
  qaHtml += qaCard('raport-penilaian','ph-duotone ph-medal','Raport Nilai','Cetak Nilai','#fdf2f8','#9d174d');
  if(isAdmin) qaHtml += qaCard('settings','ph-duotone ph-gear','Pengaturan','Sistem','#f1f5f9','#334155');

  $('main').innerHTML=`
  <div class="premium-hero au">
    <h3>Assalamu'alaikum, ${user.nama}</h3>
    <p>Mendidik generasi Qur'ani dengan ilmu dan akhlaq</p>
    <div style="display:flex;gap:0.8rem;position:relative;z-index:1">
      <span style="background:rgba(255,255,255,0.2);padding:0.4rem 0.8rem;border-radius:12px;font-size:0.75rem;backdrop-filter:blur(10px)"><i class="ph-duotone ph-calendar-blank"></i> ${dateStr}</span>
      <span style="background:rgba(255,255,255,0.2);padding:0.4rem 0.8rem;border-radius:12px;font-size:0.75rem;backdrop-filter:blur(10px)"><i class="ph-duotone ph-clock"></i> <span id="dashClock"></span> WIB</span>
    </div>
  </div>

  <div class="qa-stat-grid au">
    <div class="stat-card c-green"><div class="stat-info"><div class="sn">${d.total_santri||0}</div><div class="sl">TOTAL SANTRI</div></div><div class="si"><i class="ph-duotone ph-users"></i></div></div>
    <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${d.hadir_hari_ini||0}</div><div class="sl">HADIR</div></div><div class="si"><i class="ph-duotone ph-check-fat"></i></div></div>
    <div class="stat-card c-amber"><div class="stat-info"><div class="sn">${d.izin_sakit||0}</div><div class="sl">IZIN/SAKIT</div></div><div class="si"><i class="ph-duotone ph-first-aid-kit"></i></div></div>
    <div class="stat-card c-red"><div class="stat-info"><div class="sn">${pctHadir}%</div><div class="sl">KEHADIRAN</div></div><div class="si"><i class="ph-duotone ph-chart-pie-slice"></i></div></div>
  </div>

  <div class="card au" style="margin-bottom:2rem;position:relative;z-index:10;padding:1rem;border-radius:24px;border:none;box-shadow:0 8px 30px rgba(0,0,0,.04)">
    <div style="display:flex;align-items:center;gap:.5rem">
      <div style="position:relative;flex:1">
        <i class="ph-duotone ph-magnifying-glass" style="position:absolute;left:1.2rem;top:50%;transform:translateY(-50%);color:var(--t3);font-size:1.2rem"></i>
        <input type="text" id="dashSantriSearch" placeholder="Cari data santri..." oninput="searchSantriProfile(this.value)" autocomplete="off" style="width:100%;padding:1.1rem 1rem 1.1rem 3.2rem;border-radius:20px;border:1.5px solid var(--border);font-size:.9rem;background:#f8fafc">
        <div id="dashSantriResults" class="search-results"></div>
      </div>
    </div>
    <div id="santriProfileArea"></div>
  </div>

  <div class="section-title au" style="margin-bottom:1rem">Akses Cepat</div>
  <div class="quick-action-grid au">${qaHtml}</div>

  ${jadwalHtml}
  `;
  startDashClock();
}
"""

new_js = js[:start_idx] + premium_logic + js[end_idx:]

with open('app-core.js', 'w', encoding='utf-8') as f:
    f.write(new_js)

print("Premium dashboard JS updated.")
