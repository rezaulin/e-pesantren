// =====================================================================
// CETAK RAPORT TERPADU
// =====================================================================

window._crData = { santri: [], optSekolah: '', optDiniyyah: '', optKamar: '' };

window.loadCetakRaport = async function() {
  const isDin = user.role === 'admin_diniyyah';
  
  $('main').innerHTML = `
    <div class="page-header fade-up">
      <h2><i class="ri-printer-line"></i> Cetak Raport Terpadu</h2>
    </div>
    <div class="card fade-up">
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end;margin-bottom:1rem">
        <div class="fg" style="flex:1;min-width:150px">
          <label>Kategori Raport</label>
          <select id="crKategori" onchange="changeCrKategori()">
            ${!isDin ? '<option value="sekolah">Sekolah Formal</option>' : ''}
            <option value="diniyyah" ${isDin ? 'selected' : ''}>Madrasah Diniyyah</option>
            ${!isDin ? '<option value="kegiatan">Kegiatan Asrama</option>' : ''}
          </select>
        </div>
        <div class="fg" style="flex:1;min-width:150px" id="crFilterContainer">
          <label>Filter Kelas/Kamar</label>
          <select id="crFilterId"><option value="">-- Loading --</option></select>
        </div>
        <div class="fg" style="flex:1;min-width:120px">
          <label>Semester / Bulan</label>
          <input id="crSemester" value="2026-1" placeholder="2026-1">
        </div>
        <div class="fg" style="flex:1;min-width:150px">
          <label>Cari Nama Santri</label>
          <input id="crSearch" placeholder="Ketik nama..." oninput="renderCrList()">
        </div>
        <button class="btn btn-primary" style="margin-bottom:2px" onclick="renderCrList()"><i class="ri-search-line"></i> Tampilkan</button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
        <h3 style="margin:0;font-size:1.1rem"><i class="ri-list-check"></i> Daftar Santri</h3>
        <button class="btn btn-gold btn-sm" onclick="cetakSemuaHTML()"><i class="ri-printer-fill"></i> Cetak Semua PDF (Kelas)</button>
      </div>
      <div id="crList" style="border:1px solid var(--border);border-radius:8px;background:var(--bg)">
        <div style="padding:1.5rem;text-align:center;color:var(--t3)">Pilih filter lalu klik Tampilkan</div>
      </div>
    </div>
  `;

  try {
    const r = await api('/api/raport-terpadu/options');
    window._crData = r;
    changeCrKategori();
  } catch(e) {
    toast('Gagal memuat filter: ' + e.message);
  }
};

window.changeCrKategori = function() {
  const kat = $('crKategori').value;
  const fc = $('crFilterContainer');
  const d = window._crData;
  if(kat === 'sekolah') {
    fc.innerHTML = `<label>Filter Kelas Sekolah</label><select id="crFilterId"><option value="">-- Semua --</option>${d.optSekolah}</select>`;
  } else if(kat === 'diniyyah') {
    fc.innerHTML = `<label>Filter Kelas Diniyyah</label><select id="crFilterId"><option value="">-- Semua --</option>${d.optDiniyyah}</select>`;
  } else {
    fc.innerHTML = `<label>Filter Kamar</label><select id="crFilterId"><option value="">-- Semua --</option>${d.optKamar}</select>`;
  }
  $('crList').innerHTML = '<div style="padding:1.5rem;text-align:center;color:var(--t3)">Klik Tampilkan untuk memuat daftar santri</div>';
};

window.renderCrList = function() {
  const kat = $('crKategori').value;
  const fid = $('crFilterId').value;
  const q = ($('crSearch').value||'').toLowerCase();
  
  let list = window._crData.santri.filter(s => s.nama.toLowerCase().includes(q));
  if(fid) {
    if(kat === 'sekolah') list = list.filter(s => s.kelas_sekolah_id == fid);
    else if(kat === 'diniyyah') list = list.filter(s => s.kelas_diniyyah_id == fid);
    else if(kat === 'kegiatan') list = list.filter(s => s.kamar_id == fid);
  }

  $('crList').innerHTML = list.map(s => `
    <div class="abs-list-item" style="display:flex;align-items:center;justify-content:space-between">
      <div style="display:flex;align-items:center;gap:.6rem">
        <i class="ri-user-3-line" style="color:var(--p)"></i>
        <div>
          <div style="font-weight:600">${s.nama}</div>
          <div style="font-size:.72rem;color:var(--t3)">Kamar: ${s.kamar_nama||'-'} | Sek: ${s.kelas_sekolah||'-'} | Din: ${s.kelas_diniyyah||'-'}</div>
        </div>
      </div>
      <div style="display:flex;gap:.3rem">
        <button class="btn btn-primary btn-sm" onclick="cetakSatuHTML(${s.id})"><i class="ri-printer-line"></i> Cetak / Lihat</button>
      </div>
    </div>
  `).join('') || '<div style="padding:1.5rem;text-align:center;color:var(--t3)">Tidak ada data santri</div>';
};

window.cetakSatuHTML = async function(santri_id) {
  const kat = $('crKategori').value;
  const sem = $('crSemester').value;
  toast('Menyiapkan raport... mohon tunggu');
  try {
    const res = await api(`/api/raport-terpadu/data?kategori=${kat}&semester=${sem}&bulan=${sem}&santri_id=${santri_id}`);
    _renderPrintWindow(res.data, res.settings, kat, sem);
  } catch(e) {
    toast('Gagal: ' + e.message);
  }
};

window.cetakSemuaHTML = async function() {
  const kat = $('crKategori').value;
  const sem = $('crSemester').value;
  const fid = $('crFilterId').value;
  if(!fid) return alert('Pilih kelas atau kamar terlebih dahulu untuk mencetak massal!');
  toast('Menyiapkan seluruh raport... mohon tunggu');
  try {
    const res = await api(`/api/raport-terpadu/data?kategori=${kat}&semester=${sem}&bulan=${sem}&kelas_id=${fid}`);
    _renderPrintWindow(res.data, res.settings, kat, sem);
  } catch(e) {
    toast('Gagal: ' + e.message);
  }
};

window._renderPrintWindow = function(data, settings, kategori, sem) {
  if(!data || data.length === 0) return alert('Tidak ada data raport untuk filter ini.');
  
  const sub = kategori === 'sekolah' ? 'sekolah' : kategori === 'diniyyah' ? 'diniyyah' : 'kegiatan';
  const namaLembaga = settings['nama_lembaga_'+sub] || 'PESANTREN';
  const h1 = settings['header_'+sub+'_line1'] || 'LAPORAN HASIL BELAJAR';
  const h2 = settings['header_'+sub+'_line2'] || '';
  const al = settings['header_'+sub+'_alamat'] || '';
  const logo = settings['logo_'+sub] || '';

  let html = `
    <html><head><title>Cetak Raport Terpadu</title>
    <style>
      body { font-family: 'Arial', sans-serif; font-size: 12px; margin: 0; padding: 0; color: #000; }
      .page { padding: 10mm; page-break-after: always; box-sizing: border-box; }
      .page:last-child { page-break-after: auto; }
      .header { display: flex; align-items: center; justify-content: center; text-align: center; border-bottom: 3px double #000; padding-bottom: 10px; margin-bottom: 15px; position: relative; }
      .header img { position: absolute; left: 10px; top: 0; max-height: 70px; }
      .header-text h1 { font-size: 18px; font-weight: bold; margin: 0 0 5px 0; text-transform: uppercase; }
      .header-text h2 { font-size: 14px; font-weight: normal; margin: 0 0 3px 0; }
      .header-text p { font-size: 11px; margin: 0; }
      .identity { display: grid; grid-template-columns: 100px 10px auto 100px 10px auto; margin-bottom: 15px; font-size: 12px; }
      .identity div { padding: 3px 0; }
      table { width: 100%; border-collapse: collapse; margin-bottom: 15px; font-size: 12px; }
      th, td { border: 1px solid #000; padding: 6px; text-align: center; }
      th { background-color: #f0f0f0; font-weight: bold; }
      td.text-left { text-align: left; }
      .section-title { font-weight: bold; font-size: 13px; margin: 10px 0 5px 0; text-transform: uppercase; }
      .footer { display: flex; justify-content: space-between; margin-top: 30px; }
      .ttd-box { text-align: center; width: 200px; }
      .ttd-space { height: 60px; }
      @media print {
        @page { size: A4 portrait; margin: 10mm; }
        .page { padding: 0; }
        body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
      }
    </style>
    </head><body>
  `;

  data.forEach(s => {
    html += '<div class="page">';
    html += '<div class="header">';
    if(logo) html += `<img src="${logo}">`;
    html += `<div class="header-text">
        <h1>${namaLembaga}</h1>
        ${h1 ? '<h2>'+h1+'</h2>' : ''}
        ${h2 ? '<h2>'+h2+'</h2>' : ''}
        <p>${al}</p>
      </div>
    </div>`;

    html += `<div class="identity">
      <div>Nama Santri</div><div>:</div><div style="font-weight:bold">${s.nama}</div>
      <div>Semester/Bulan</div><div>:</div><div>${sem}</div>`;
    if(kategori === 'sekolah') {
      html += `<div>Kelas</div><div>:</div><div>${s.kelas_sekolah}</div><div>Kamar</div><div>:</div><div>${s.kamar}</div>`;
    } else if(kategori === 'diniyyah') {
      html += `<div>Kelas Diniyyah</div><div>:</div><div>${s.kelas_diniyyah}</div><div>Kamar</div><div>:</div><div>${s.kamar}</div>`;
    } else {
      html += `<div>Kamar</div><div>:</div><div>${s.kamar}</div><div>Kelas Sek/Din</div><div>:</div><div>${s.kelas_sekolah} / ${s.kelas_diniyyah}</div>`;
    }
    html += '</div>';

    html += '<div class="section-title">A. HASIL BELAJAR (NILAI)</div>';
    if(s.nilai && s.nilai.length > 0) {
      if(kategori === 'kegiatan') {
        html += `<table>
          <tr><th style="width:40%">Nama Kegiatan</th><th>Kelompok</th><th>Nilai</th><th>Catatan</th></tr>
          ${s.nilai.map(n => `<tr>
            <td class="text-left">${n.kegiatan}</td>
            <td>${n.kelompok}</td>
            <td><strong>${n.nilai}</strong></td>
            <td class="text-left">${n.catatan}</td>
          </tr>`).join('')}
          <tr><th colspan="2" class="text-left">RATA-RATA</th><th colspan="2" class="text-left">${s.rata_rata}</th></tr>
        </table>`;
      } else {
        let komp = [];
        try {
          if(kategori === 'sekolah') komp = JSON.parse(settings.komponen_nilai_sekolah || '[]');
          else komp = JSON.parse(settings.komponen_nilai_diniyyah || '[]');
        } catch(e){}
        if(komp.length === 0) komp = [{nama:"Harian"}, {nama:"UTS"}, {nama:"UAS"}];
        
        let thHTML = komp.map(k => `<th>${k.nama}</th>`).join('');
        
        html += `<table>
          <tr><th style="width:40%">Mata Pelajaran</th>${thHTML}<th>Nilai Akhir</th></tr>
          ${s.nilai.map(n => {
            let tdHTML = komp.map(k => {
               // coba n.detail[nama], lalu fallback n[nama.toLowerCase()] (untuk n.harian, n.uts dsb)
               let val = n.detail && n.detail[k.nama] !== undefined ? n.detail[k.nama] : (n[k.nama.toLowerCase()] !== undefined ? n[k.nama.toLowerCase()] : '-');
               return `<td>${val}</td>`;
            }).join('');
            return `<tr>
              <td class="text-left">${n.mapel}</td>
              ${tdHTML}
              <td><strong>${n.akhir}</strong></td>
            </tr>`;
          }).join('')}
          <tr><th colspan="${1 + komp.length}" class="text-left">RATA-RATA</th><th class="text-left">${s.rata_rata}</th></tr>
        </table>`;
      }
    } else {
      html += '<p style="text-align:center;color:#666">Belum ada data nilai.</p>';
    }

    html += '<div class="section-title">B. KETIDAKHADIRAN (ABSENSI)</div>';
    if(s.absensi && s.absensi.length > 0) {
      html += `<table>
        <tr><th style="width:40%">Kategori / Kegiatan</th><th>Hadir</th><th>Izin</th><th>Sakit</th><th>Alpa</th><th>Total Absen (I+S+A)</th></tr>
        ${s.absensi.map(a => `<tr>
          <td class="text-left">${a.kegiatan}</td>
          <td>${a.H}</td><td>${a.I}</td><td>${a.S}</td><td>${a.A}</td>
          <td><strong>${a.I + a.S + a.A}</strong></td>
        </tr>`).join('')}
      </table>`;
    } else {
      html += '<p style="text-align:center;color:#666">Belum ada data ketidakhadiran.</p>';
    }

    const tgl = new Date().toLocaleDateString('id-ID', {day:'numeric',month:'long',year:'numeric'});
    const kpl = kategori==='sekolah'?'Sekolah':kategori==='diniyyah'?'Madrasah':'Asrama';
    html += `<div class="footer">
      <div class="ttd-box">
        <div>Mengetahui,</div>
        <div>Wali Kelas / Kamar</div>
        <div class="ttd-space"></div>
        <div>( ......................................... )</div>
      </div>
      <div class="ttd-box">
        <div>Ditetapkan tanggal: ${tgl}</div>
        <div>Kepala ${kpl}</div>
        <div class="ttd-space"></div>
        <div>( ......................................... )</div>
      </div>
    </div>`;

    html += '</div>'; // end page
  });

  html += `
    <script>
      window.onload = function() {
        setTimeout(function(){ window.print(); }, 500);
      }
    </script>
    </body></html>`;

  const w = window.open('', '_blank');
  w.document.write(html);
  w.document.close();
};
