import re

with open('app-core.js', 'r', encoding='utf-8') as f:
    js = f.read()

# We want to replace the Admin Dashboard logic from `// -- Admin Dashboard (SIPP Style) --`
# to the end of the `loadDashboard` function.
# Let's find the start and end indices.

start_marker = "// -- Admin Dashboard (SIPP Style) --"
end_marker = "// Waktu page (server time + jadwal)"

start_idx = js.find(start_marker)
end_idx = js.find(end_marker)

if start_idx == -1 or end_idx == -1:
    print("Could not find markers!")
    exit(1)

admin_logic = """// -- Admin Dashboard (SIPP Style) --
  const hf=hasFeature;
  const totalAbsen=(d.hadir_hari_ini||0)+(d.izin_sakit||0)+(d.alfa||0);
  const pctHadir=totalAbsen>0?Math.round(((d.hadir_hari_ini||0)/totalAbsen)*100):0;

  const umc = (id, ic, lb, sub, bg) => `<div class="ust-menu-card au" onclick="nav('${id}')"><div class="umc-icon" style="background:${bg}"><i class="${ic}"></i></div><div class="umc-label">${lb}</div><div class="umc-sub">${sub}</div></div>`;

  let groupsHtml = '';

  // 1. Data Santri
  let sItems = [];
  sItems.push(umc('santri','ri-group-line','Data Santri','Kelola profil','linear-gradient(135deg,#0ea5e9,#38bdf8)'));
  sItems.push(umc('psb','ri-user-add-line','PSB Online','Penerimaan','linear-gradient(135deg,#10b981,#34d399)'));
  groupsHtml += `<div class="section-title" style="margin-top:1rem">DATA SANTRI</div><div class="ust-menu-grid au">${sItems.join('')}</div>`;

  // 2. Sekolah Formal
  if(hf('absen_sekolah')||hf('kelas_sekolah')){
    let items = [];
    items.push(umc('kelas-sekolah','ri-school-line','Sekolah','Kelas & Jadwal','linear-gradient(135deg,#f59e0b,#fbbf24)'));
    items.push(umc('absen-sekolah','ri-checkbox-circle-line','Absensi','Kelola Absensi','linear-gradient(135deg,#10b981,#34d399)'));
    items.push(umc('rekap-sekolah','ri-file-list-3-line','Rekap','Rekap Absen','linear-gradient(135deg,#3b82f6,#60a5fa)'));
    if(isAdmin)items.push(umc('input-nilai-sekolah','ri-award-line','Penilaian','Nilai Siswa','linear-gradient(135deg,#ec4899,#f472b6)'));
    groupsHtml += `<div class="section-title" style="margin-top:1rem">SEKOLAH FORMAL</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 3. Madrasah Diniyyah
  {
    let items = [];
    if(isAdmin)items.push(umc('kelas-diniyyah','ri-book-2-line','Diniyyah','Kelas & Jadwal','linear-gradient(135deg,#8b5cf6,#a78bfa)'));
    items.push(umc('absen-diniyyah','ri-checkbox-circle-line','Absensi','Kelola Absen','linear-gradient(135deg,#10b981,#34d399)'));
    items.push(umc('rekap-diniyyah','ri-file-list-3-line','Rekap','Rekap Absen','linear-gradient(135deg,#3b82f6,#60a5fa)'));
    if(isAdmin)items.push(umc('input-nilai-diniyyah','ri-award-line','Penilaian','Nilai Santri','linear-gradient(135deg,#ec4899,#f472b6)'));
    groupsHtml += `<div class="section-title" style="margin-top:1rem">MADRASAH DINIYYAH</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 4. Asrama
  if(hf('absen_malam')){
    let items = [];
    if(isAdmin)items.push(umc('kamar','ri-home-5-line','Kamar','Kelola Kamar','linear-gradient(135deg,#4f46e5,#818cf8)'));
    items.push(umc('absen-malam','ri-moon-line','Absen Malam','Asrama','linear-gradient(135deg,#6366f1,#818cf8)'));
    items.push(umc('rekap-kamar','ri-file-list-3-line','Rekap','Rekap Absen','linear-gradient(135deg,#3b82f6,#60a5fa)'));
    groupsHtml += `<div class="section-title" style="margin-top:1rem">ASRAMA & KAMAR</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 5. Kegiatan Harian
  if(hf('absensi_harian')||hf('jadwal')){
    let items = [];
    if(isAdmin)items.push(umc('kegiatan','ri-book-open-line','Kegiatan','Kelola Kegiatan','linear-gradient(135deg,#059669,#34d399)'));
    if(hf('jadwal'))items.push(umc('jadwal','ri-calendar-schedule-line','Jadwal','Jadwal Kegiatan','linear-gradient(135deg,#d97706,#fbbf24)'));
    items.push(umc('absensi','ri-checkbox-circle-line','Absensi','Kegiatan Harian','linear-gradient(135deg,#10b981,#34d399)'));
    items.push(umc('rekap-kegiatan','ri-file-list-3-line','Rekap','Rekap Kegiatan','linear-gradient(135deg,#3b82f6,#60a5fa)'));
    groupsHtml += `<div class="section-title" style="margin-top:1rem">KEGIATAN HARIAN</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 6. Keuangan
  if(hf('keuangan')){
    let items = [umc('pembayaran','ri-money-dollar-circle-line','Pembayaran','Tagihan & Kas','linear-gradient(135deg,#10b981,#34d399)')];
    if(hf('catatan_bendahara'))items.push(umc('catatan-bendahara','ri-book-3-line','Catatan','Buku Keuangan','linear-gradient(135deg,#059669,#34d399)'));
    groupsHtml += `<div class="section-title" style="margin-top:1rem">KEUANGAN</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 7. Kedisiplinan
  if(hf('kedisiplinan')){
    let items = [
      umc('pelanggaran','ri-error-warning-line','Pelanggaran','Catat Kasus','linear-gradient(135deg,#ef4444,#f87171)'),
      umc('perizinan','ri-pass-valid-line','Perizinan','Kelola Izin','linear-gradient(135deg,#f59e0b,#fbbf24)'),
      umc('catatan-guru','ri-sticky-note-line','Catatan Guru','Perkembangan','linear-gradient(135deg,#db2777,#f472b6)')
    ];
    groupsHtml += `<div class="section-title" style="margin-top:1rem">KEDISIPLINAN</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 8. Laporan
  {
    let items = [
      umc('raport-absensi','ri-file-chart-line','Raport Absen','Cetak Dokumen','linear-gradient(135deg,#2563eb,#60a5fa)'),
      umc('raport-penilaian','ri-award-line','Raport Nilai','Cetak Nilai','linear-gradient(135deg,#db2777,#f472b6)'),
      umc('sensus','ri-bar-chart-box-line','Sensus','Data Statistik','linear-gradient(135deg,#475569,#94a3b8)')
    ];
    groupsHtml += `<div class="section-title" style="margin-top:1rem">LAPORAN & STATISTIK</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 9. Tahfidz
  if(hf('tahfidz')){
    let items = [];
    if(isAdmin)items.push(umc('tahfidz-halaqoh','ri-book-read-line','Halaqoh','Kelompok','linear-gradient(135deg,#0d9488,#14b8a6)'));
    items.push(umc('tahfidz-absensi','ri-quill-pen-line','Setoran','Absensi & Nilai','linear-gradient(135deg,#059669,#10b981)'));
    items.push(umc('tahfidz-rekap','ri-bar-chart-grouped-line','Rekap','Progress Hafalan','linear-gradient(135deg,#0891b2,#06b6d4)'));
    groupsHtml += `<div class="section-title" style="margin-top:1rem">TAHFIDZ QUR'AN</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  // 10. Pengaturan & Extra
  {
    let items = [];
    if(hf('e_paket')) items.push(umc('e-paket','ri-box-3-line','E-Paket','Logistik','linear-gradient(135deg,#7c3aed,#a78bfa)'));
    if(isAdmin) {
      items.push(umc('users','ri-user-settings-line','Pengguna','Kelola Akun','linear-gradient(135deg,#475569,#94a3b8)'));
      items.push(umc('pengumuman','ri-megaphone-line','Pengumuman','Buat Info','linear-gradient(135deg,#ea580c,#fb923c)'));
      items.push(umc('settings','ri-settings-3-line','Pengaturan','Konfigurasi','linear-gradient(135deg,#64748b,#94a3b8)'));
    }
    groupsHtml += `<div class="section-title" style="margin-top:1rem">PENGATURAN & EXTRA</div><div class="ust-menu-grid au">${items.join('')}</div>`;
  }

  $('main').innerHTML=`
  <div class="welcome-banner au">
    <h3>Assalamu'alaikum, ${user.nama}</h3>
    <div class="wb-motto">Mendidik generasi Qur'ani dengan ilmu dan akhlaq</div>
    <div class="wb-date"><span><i class="ri-calendar-line"></i> ${dateStr}</span><span><i class="ri-time-line"></i> <span id="dashClock"></span> WIB</span></div>
  </div>

  <div class="stat-grid au" style="grid-template-columns:repeat(4,1fr);margin-bottom:.8rem">
    <div class="stat-card c-green"><div class="stat-info"><div class="sn">${d.total_santri||0}</div><div class="sl">TOTAL SANTRI</div></div><div class="si"><i class="ri-group-line"></i></div></div>
    <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${d.hadir_hari_ini||0}</div><div class="sl">HADIR</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
    <div class="stat-card c-amber"><div class="stat-info"><div class="sn">${d.izin_sakit||0}</div><div class="sl">IZIN/SAKIT</div></div><div class="si"><i class="ri-pass-valid-line"></i></div></div>
    <div class="stat-card c-red"><div class="stat-info"><div class="sn">${pctHadir}%</div><div class="sl">KEHADIRAN</div></div><div class="si"><i class="ri-pie-chart-line"></i></div></div>
  </div>

  <div class="card au" style="margin-bottom:1.5rem;position:relative;z-index:10;padding:1rem;border-radius:24px">
    <div style="display:flex;align-items:center;gap:.5rem">
      <div style="position:relative;flex:1">
        <i class="ri-search-line" style="position:absolute;left:1rem;top:50%;transform:translateY(-50%);color:var(--t3);font-size:1.1rem"></i>
        <input type="text" id="dashSantriSearch" placeholder="Cari data santri, kelas, guru, atau laporan..." oninput="searchSantriProfile(this.value)" autocomplete="off" style="width:100%;padding:1rem 1rem 1rem 2.8rem;border-radius:18px;border:1.5px solid var(--border);font-size:.88rem;background:#f8fafc">
        <div id="dashSantriResults" class="search-results"></div>
      </div>
    </div>
    <div id="santriProfileArea"></div>
  </div>

  ${jadwalHtml}

  ${groupsHtml}
  `;
  startDashClock();
}
"""

new_js = js[:start_idx] + admin_logic + js[end_idx:]

with open('app-core.js', 'w', encoding='utf-8') as f:
    f.write(new_js)

print("Admin dashboard logic replaced successfully!")
