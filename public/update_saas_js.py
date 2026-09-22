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

saas_logic = """// -- Admin Dashboard (SIPP Style) --
  const hf=hasFeature;
  const totalAbsen=(d.hadir_hari_ini||0)+(d.izin_sakit||0)+(d.alfa||0);
  const pctHadir=totalAbsen>0?Math.round(((d.hadir_hari_ini||0)/totalAbsen)*100):0;
  const t_santri = d.total_santri||0;

  // Render Pill Cards for Aksi Cepat
  const pillCard = (id, ic, lb, cl) => `<div class="pill-card" onclick="nav('${id}')"><div class="pill-icon" style="background:var(--pbg);color:var(--p)"><i class="${ic}" style="color:${cl}"></i></div><div class="pill-text">${lb}</div></div>`;
  
  let pillHtml = '';
  pillHtml += pillCard('santri','ph-duotone ph-user-plus','Tambah Santri','#2563eb');
  if(hf('absen_sekolah')) pillHtml += pillCard('absen-sekolah','ph-duotone ph-check-square-offset','Input Absensi','#16a34a');
  if(hf('keuangan')) pillHtml += pillCard('pembayaran','ph-duotone ph-wallet','Pembayaran','#8b5cf6');
  pillHtml += pillCard('catatan-guru','ph-duotone ph-file-text','Catatan Guru','#f59e0b');
  if(hf('kedisiplinan')) pillHtml += pillCard('pelanggaran','ph-duotone ph-shield-warning','Pelanggaran','#ef4444');

  $('main').innerHTML=`
  <div style="margin-bottom:1.5rem">
    <h1 style="font-size:1.4rem;font-weight:800;color:var(--text);margin-bottom:.3rem">Beranda</h1>
    <p style="color:var(--t3);font-size:.85rem">Ringkasan kegiatan dan situasi pesantren hari ini.</p>
  </div>

  <div class="premium-hero au" style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:2rem">
    <div style="flex:1;min-width:300px;position:relative;z-index:1">
      <h3 style="font-size:1.6rem;font-weight:800;margin-bottom:0.8rem">Assalamu'alaikum, ${window.user?.nama||'Admin'} 👋</h3>
      <p style="font-size:0.9rem;opacity:0.9;max-width:400px;line-height:1.5;margin-bottom:1.5rem">Semoga setiap langkah kita hari ini menjadi bagian dari keberkahan dalam mendidik generasi Qur'ani.</p>
      <div style="display:flex;gap:1rem;font-size:0.8rem">
        <span style="display:flex;align-items:center;gap:0.4rem"><i class="ph-duotone ph-calendar-blank"></i> ${dateStr}</span>
        <span style="display:flex;align-items:center;gap:0.4rem"><i class="ph-duotone ph-clock"></i> <span id="dashClock"></span> WIB</span>
      </div>
    </div>
    
    <div style="flex:1;min-width:300px;background:rgba(255,255,255,0.1);padding:1.5rem;border-radius:20px;border:1px solid rgba(255,255,255,0.1);backdrop-filter:blur(10px);position:relative;z-index:1">
      <div style="font-size:0.7rem;font-weight:700;letter-spacing:0.05em;margin-bottom:1rem;opacity:0.8">SITUASI HARI INI</div>
      <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:1rem">
        <div><i class="ph-duotone ph-check-circle" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">${pctHadir}%</div><div style="font-size:0.65rem;opacity:0.7">Kehadiran Santri</div></div>
        <div><i class="ph-duotone ph-shield-warning" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">0</div><div style="font-size:0.65rem;opacity:0.7">Pelanggaran</div></div>
        <div><i class="ph-duotone ph-wallet" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">Rp 0</div><div style="font-size:0.65rem;opacity:0.7">Pembayaran</div></div>
        <div><i class="ph-duotone ph-calendar-check" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">12</div><div style="font-size:0.65rem;opacity:0.7">Kegiatan</div></div>
      </div>
    </div>
  </div>

  <div class="stat-grid au" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:1rem;margin-bottom:2rem">
    <div class="stat-card-clean"><div class="sc-icon" style="background:#eff6ff;color:#2563eb"><i class="ph-duotone ph-users-three"></i></div><div class="sc-body"><div class="sc-label">Total Santri</div><div class="sc-val">${t_santri}</div><div class="sc-sub" style="color:#2563eb">● Terdaftar aktif</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#ecfdf5;color:#16a34a"><i class="ph-duotone ph-check-circle"></i></div><div class="sc-body"><div class="sc-label">Absensi Hari Ini</div><div class="sc-val">${pctHadir}%</div><div class="sc-sub">${d.hadir_hari_ini||0} / ${t_santri} hadir</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#f5f3ff;color:#8b5cf6"><i class="ph-duotone ph-wallet"></i></div><div class="sc-body"><div class="sc-label">Pembayaran Bln Ini</div><div class="sc-val">Rp 0</div><div class="sc-sub">0% dari total tagihan</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#fffbeb;color:#d97706"><i class="ph-duotone ph-bed"></i></div><div class="sc-body"><div class="sc-label">Kamar Aktif</div><div class="sc-val">${d.total_kamar||0}</div><div class="sc-sub">100% terisi</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#fef2f2;color:#dc2626"><i class="ph-duotone ph-shield-warning"></i></div><div class="sc-body"><div class="sc-label">Pelanggaran</div><div class="sc-val">0</div><div class="sc-sub">Tidak ada kasus</div></div></div>
  </div>

  <div class="section-title au" style="margin-bottom:1rem;font-size:0.8rem;letter-spacing:0.05em">AKSI CEPAT</div>
  <div class="pill-grid au">${pillHtml}</div>

  <div class="bottom-cols au">
    <!-- Kolom 1 -->
    <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02)">
      <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Perlu Ditindaklanjuti</h3>
      <div class="list-item">
        <div class="li-icon" style="background:#eff6ff;color:#2563eb"><i class="ph-duotone ph-users"></i></div>
        <div class="li-body"><div class="li-title">Santri belum absen hari ini</div><div class="li-sub">8 santri belum melakukan absensi</div></div>
        <div class="li-right li-badge">8</div>
      </div>
      <div class="list-item">
        <div class="li-icon" style="background:#fffbeb;color:#d97706"><i class="ph-duotone ph-wallet"></i></div>
        <div class="li-body"><div class="li-title">Pembayaran tertunda</div><div class="li-sub">12 santri memiliki tunggakan</div></div>
        <div class="li-right li-badge" style="background:rgba(217,119,6,0.1);color:#d97706">12</div>
      </div>
      <div class="list-item">
        <div class="li-icon" style="background:#fef2f2;color:#dc2626"><i class="ph-duotone ph-shield-warning"></i></div>
        <div class="li-body"><div class="li-title">Pelanggaran menunggu review</div><div class="li-sub">3 kasus perlu ditinjau</div></div>
        <div class="li-right li-badge">3</div>
      </div>
      <a href="#" style="display:block;margin-top:1.5rem;font-size:0.8rem;font-weight:700;color:var(--p);text-decoration:none">Lihat semua &rarr;</a>
    </div>

    <!-- Kolom 2 -->
    <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02)">
      <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Aktivitas Terbaru</h3>
      <div class="list-item">
        <div class="li-icon" style="background:#ecfdf5;color:#16a34a"><i class="ph-duotone ph-check-circle"></i></div>
        <div class="li-body"><div class="li-title">Absensi pagi telah dimulai</div><div class="li-sub">oleh Ust. Ahmad</div></div>
        <div class="li-right" style="color:var(--t3);font-weight:500">07:15</div>
      </div>
      <div class="list-item">
        <div class="li-icon" style="background:#eff6ff;color:#2563eb"><i class="ph-duotone ph-wallet"></i></div>
        <div class="li-body"><div class="li-title">Pembayaran diterima dari Ahmad Fawaid</div><div class="li-sub" style="color:#16a34a">Rp 250.000</div></div>
        <div class="li-right" style="color:var(--t3);font-weight:500">Kemarin</div>
      </div>
      <div class="list-item">
        <div class="li-icon" style="background:#f5f3ff;color:#8b5cf6"><i class="ph-duotone ph-clock"></i></div>
        <div class="li-body"><div class="li-title">Catatan baru ditambahkan untuk kelas 8A</div><div class="li-sub">oleh Ust. Hasan</div></div>
        <div class="li-right" style="color:var(--t3);font-weight:500">Kemarin</div>
      </div>
      <div class="list-item">
        <div class="li-icon" style="background:#fffbeb;color:#d97706"><i class="ph-duotone ph-calendar"></i></div>
        <div class="li-body"><div class="li-title">Jadwal pelajaran diubah Matematika - Kelas 9B</div><div class="li-sub">oleh Ust. Rahman</div></div>
        <div class="li-right" style="color:var(--t3);font-weight:500">2 hari lalu</div>
      </div>
      <a href="#" style="display:block;margin-top:1.5rem;font-size:0.8rem;font-weight:700;color:var(--p);text-decoration:none">Lihat semua aktivitas &rarr;</a>
    </div>

    <!-- Kolom 3 -->
    <div style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02);flex:1">
        <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Kehadiran Hari Ini</h3>
        <div style="display:flex;align-items:center;gap:1.5rem">
          <div style="width:100px;height:100px;border-radius:50%;background:conic-gradient(#16a34a ${pctHadir}%, #e2e8f0 0);position:relative;display:flex;align-items:center;justify-content:center">
            <div style="width:75px;height:75px;background:#fff;border-radius:50%;display:flex;flex-direction:column;align-items:center;justify-content:center"><div style="font-size:1.1rem;font-weight:800">${pctHadir}%</div><div style="font-size:0.5rem;color:var(--t3)">hadir</div></div>
          </div>
          <div style="flex:1">
            <div style="display:flex;justify-content:space-between;margin-bottom:0.5rem;font-size:0.75rem"><span style="display:flex;align-items:center;gap:0.4rem"><span style="width:8px;height:8px;border-radius:50%;background:#16a34a"></span>Hadir</span><strong>${d.hadir_hari_ini||0}</strong></div>
            <div style="display:flex;justify-content:space-between;margin-bottom:0.5rem;font-size:0.75rem"><span style="display:flex;align-items:center;gap:0.4rem"><span style="width:8px;height:8px;border-radius:50%;background:#f59e0b"></span>Izin/Sakit</span><strong>${d.izin_sakit||0}</strong></div>
            <div style="display:flex;justify-content:space-between;font-size:0.75rem"><span style="display:flex;align-items:center;gap:0.4rem"><span style="width:8px;height:8px;border-radius:50%;background:#e2e8f0"></span>Alfa</span><strong>${d.alfa||0}</strong></div>
          </div>
        </div>
      </div>
      <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02);flex:1">
        <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Status Pembayaran Bulan Ini</h3>
        <div style="display:flex;justify-content:space-between;margin-bottom:0.5rem;font-size:0.75rem">
          <div><span style="font-size:1.2rem;font-weight:800;color:var(--text);display:block">0%</span><span style="color:var(--t3)">Rp 0 terkumpul</span></div>
          <div style="text-align:right">
            <div style="display:flex;align-items:center;gap:0.4rem;justify-content:flex-end;margin-bottom:0.3rem"><span style="width:8px;height:8px;border-radius:50%;background:#16a34a"></span>Lunas (0)</div>
            <div style="display:flex;align-items:center;gap:0.4rem;justify-content:flex-end"><span style="width:8px;height:8px;border-radius:50%;background:#ef4444"></span>Tertunggak (${t_santri})</div>
          </div>
        </div>
        <div style="width:100%;height:8px;background:#ef4444;border-radius:4px;overflow:hidden;margin-bottom:1rem"><div style="width:0%;height:100%;background:#16a34a"></div></div>
        <div style="display:flex;justify-content:space-between;font-size:0.7rem;color:var(--t3)">
          <div>Total Tagihan<br><strong style="color:var(--text)">Rp ${(t_santri*250000).toLocaleString('id')}</strong></div>
          <div style="text-align:right">Terkumpul<br><strong style="color:var(--text)">Rp 0</strong></div>
        </div>
      </div>
    </div>
  </div>

  ${jadwalHtml}
  `;
  startDashClock();
}
"""

new_js = js[:start_idx] + saas_logic + js[end_idx:]

with open('app-core.js', 'w', encoding='utf-8') as f:
    f.write(new_js)

print("SaaS JS updated perfectly.")
