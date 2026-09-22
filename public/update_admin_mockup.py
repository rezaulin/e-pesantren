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

admin_logic = """// -- Admin Dashboard (SIPP Style) --
  const hf=hasFeature;
  const totalAbsen=(d.hadir_hari_ini||0)+(d.izin_sakit||0)+(d.alfa||0);
  const pctHadir=totalAbsen>0?Math.round(((d.hadir_hari_ini||0)/totalAbsen)*100):0;

  const subCard = (id, ic, lb, bg, fg) => `<div class="admin-sub-card" onclick="nav('${id}')" style="background:${bg};color:${fg}"><i class="${ic}"></i><div class="admin-sub-label">${lb}</div></div>`;

  let groupsHtml = '';

  // 1. Data Santri
  let sItems = [];
  sItems.push(subCard('santri','ri-group-line','Data Santri','#eff6ff','#1e40af'));
  sItems.push(subCard('psb','ri-user-add-line','PSB Online','#ecfdf5','#065f46'));
  groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Data Santri</div><div class="admin-sub-grid">${sItems.join('')}</div></div>`;

  // 2. Sekolah Formal
  if(hf('absen_sekolah')||hf('kelas_sekolah')){
    let items = [];
    items.push(subCard('kelas-sekolah','ri-school-line','Sekolah','#fffbeb','#92400e'));
    items.push(subCard('absen-sekolah','ri-checkbox-circle-line','Absensi','#ecfdf5','#065f46'));
    items.push(subCard('rekap-sekolah','ri-file-list-3-line','Rekap','#eff6ff','#1e40af'));
    if(isAdmin)items.push(subCard('input-nilai-sekolah','ri-award-line','Penilaian','#fdf2f8','#9d174d'));
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Sekolah Formal</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 3. Madrasah Diniyyah
  {
    let items = [];
    if(isAdmin)items.push(subCard('kelas-diniyyah','ri-book-2-line','Diniyyah','#f3e8ff','#6b21a8'));
    items.push(subCard('absen-diniyyah','ri-checkbox-circle-line','Absensi','#ecfdf5','#065f46'));
    items.push(subCard('rekap-diniyyah','ri-file-list-3-line','Rekap','#eff6ff','#1e40af'));
    if(isAdmin)items.push(subCard('input-nilai-diniyyah','ri-award-line','Penilaian','#fdf2f8','#9d174d'));
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Madrasah Diniyyah</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 4. Asrama
  if(hf('absen_malam')){
    let items = [];
    if(isAdmin)items.push(subCard('kamar','ri-home-5-line','Kamar','#eef2ff','#3730a3'));
    items.push(subCard('absen-malam','ri-moon-line','Absen Malam','#e0e7ff','#3730a3'));
    items.push(subCard('rekap-kamar','ri-file-list-3-line','Rekap','#eff6ff','#1e40af'));
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Asrama & Kamar</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 5. Kegiatan Harian
  if(hf('absensi_harian')||hf('jadwal')){
    let items = [];
    if(isAdmin)items.push(subCard('kegiatan','ri-book-open-line','Kegiatan','#ecfdf5','#065f46'));
    if(hf('jadwal'))items.push(subCard('jadwal','ri-calendar-schedule-line','Jadwal','#fffbeb','#92400e'));
    items.push(subCard('absensi','ri-checkbox-circle-line','Absensi','#ecfdf5','#065f46'));
    items.push(subCard('rekap-kegiatan','ri-file-list-3-line','Rekap','#eff6ff','#1e40af'));
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Kegiatan Harian</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 6. Keuangan
  if(hf('keuangan')){
    let items = [subCard('pembayaran','ri-money-dollar-circle-line','Pembayaran','#ecfdf5','#065f46')];
    if(hf('catatan_bendahara'))items.push(subCard('catatan-bendahara','ri-book-3-line','Catatan','#dcfce7','#166534'));
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Keuangan</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 7. Kedisiplinan
  if(hf('kedisiplinan')){
    let items = [
      subCard('pelanggaran','ri-error-warning-line','Pelanggaran','#fef2f2','#991b1b'),
      subCard('perizinan','ri-pass-valid-line','Perizinan','#fffbeb','#92400e'),
      subCard('catatan-guru','ri-sticky-note-line','Catatan Guru','#fdf2f8','#9d174d')
    ];
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Kedisiplinan</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 8. Laporan
  {
    let items = [
      subCard('raport-absensi','ri-file-chart-line','Raport Absen','#eff6ff','#1e40af'),
      subCard('raport-penilaian','ri-award-line','Raport Nilai','#fdf2f8','#9d174d'),
      subCard('sensus','ri-bar-chart-box-line','Sensus','#f1f5f9','#334155')
    ];
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Laporan & Statistik</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 9. Tahfidz
  if(hf('tahfidz')){
    let items = [];
    if(isAdmin)items.push(subCard('tahfidz-halaqoh','ri-book-read-line','Halaqoh','#ccfbf1','#0f766e'));
    items.push(subCard('tahfidz-absensi','ri-quill-pen-line','Setoran','#d1fae5','#047857'));
    items.push(subCard('tahfidz-rekap','ri-bar-chart-grouped-line','Rekap','#cffafe','#0e7490'));
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Tahfidz Qur'an</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  // 10. Pengaturan & Extra
  {
    let items = [];
    if(hf('e_paket')) items.push(subCard('e-paket','ri-box-3-line','E-Paket','#f3e8ff','#6b21a8'));
    if(isAdmin) {
      items.push(subCard('users','ri-user-settings-line','Pengguna','#f1f5f9','#334155'));
      items.push(subCard('pengumuman','ri-megaphone-line','Pengumuman','#ffedd5','#c2410c'));
      items.push(subCard('settings','ri-settings-3-line','Pengaturan','#f1f5f9','#334155'));
    }
    groupsHtml += `<div class="admin-card-group au"><div class="admin-card-title">Pengaturan & Extra</div><div class="admin-sub-grid">${items.join('')}</div></div>`;
  }

  $('main').innerHTML=`
  <div class="welcome-banner au">
    <h3>Assalamu'alaikum, ${user.nama}</h3>
    <div class="wb-motto">Mendidik generasi Qur'ani dengan ilmu dan akhlaq</div>
    <div class="wb-date"><span><i class="ri-calendar-line"></i> ${dateStr}</span><span><i class="ri-time-line"></i> <span id="dashClock"></span> WIB</span></div>
  </div>

  <div class="stat-grid au" style="grid-template-columns:repeat(4,1fr);margin-bottom:1.5rem">
    <div class="stat-card c-green"><div class="stat-info"><div class="sn">${d.total_santri||0}</div><div class="sl">TOTAL SANTRI</div></div><div class="si"><i class="ri-group-line"></i></div></div>
    <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${d.hadir_hari_ini||0}</div><div class="sl">HADIR</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
    <div class="stat-card c-amber"><div class="stat-info"><div class="sn">${d.izin_sakit||0}</div><div class="sl">IZIN/SAKIT</div></div><div class="si"><i class="ri-pass-valid-line"></i></div></div>
    <div class="stat-card c-red"><div class="stat-info"><div class="sn">${pctHadir}%</div><div class="sl">KEHADIRAN</div></div><div class="si"><i class="ri-pie-chart-line"></i></div></div>
  </div>

  <div class="card au" style="margin-bottom:2rem;position:relative;z-index:10;padding:1rem;border-radius:24px;border:none;box-shadow:0 8px 30px rgba(0,0,0,.04)">
    <div style="display:flex;align-items:center;gap:.5rem">
      <div style="position:relative;flex:1">
        <i class="ri-search-line" style="position:absolute;left:1.2rem;top:50%;transform:translateY(-50%);color:var(--t3);font-size:1.2rem"></i>
        <input type="text" id="dashSantriSearch" placeholder="Cari data santri, kelas, guru, atau laporan..." oninput="searchSantriProfile(this.value)" autocomplete="off" style="width:100%;padding:1.1rem 1rem 1.1rem 3.2rem;border-radius:20px;border:1.5px solid var(--border);font-size:.9rem;background:#f8fafc">
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

print("Admin dashboard logic updated for mockup design successfully!")
