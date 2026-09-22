// -- SANTRI --
let _santriFullList=[];
async function loadSantri(){
  const list=await api('/api/santri');
  _santriFullList=list;
  const isAdmin=['admin','superadmin'].includes(user.role);
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-graduation-cap-line"></i> Data Santri (${list.length})</h2>
    <div style="display:flex;gap:.5rem">
      ${isAdmin?`<button class="btn btn-outline btn-sm" onclick="showImportExcel()"><i class="ri-file-excel-2-line"></i> Import</button>
      <button class="btn btn-primary btn-sm" onclick="showAddSantri()"><i class="ri-add-line"></i> Tambah</button>`:''}</div></div>
    <div class="card fade-up" style="margin-bottom:.8rem"><div style="display:flex;gap:.6rem;align-items:center">
      <i class="ri-search-line" style="color:var(--t3);font-size:1.1rem"></i>
      <input type="text" id="santriSearch" placeholder="Cari nama santri, kamar, wali..." oninput="filterSantriTable(this.value)"
        style="flex:1;border:none;outline:none;font-size:.9rem;background:transparent">
      <span id="santriCount" style="font-size:.75rem;color:var(--t3)">${list.length} hasil</span>
    </div></div>
    <div id="santriTableBody"></div>`;
  renderSantriRows(list);
}
function renderSantriRows(list){
  const isAdmin=['admin','superadmin'].includes(user.role);
  if($('santriCount'))$('santriCount').textContent=list.length+' hasil';
  $('santriTableBody').innerHTML=`<div class="card"><div class="table-wrap"><table>
      <tr><th>Nama</th><th>Kamar</th><th>Alamat</th><th>Wali</th><th>No HP</th><th>Status</th>${isAdmin?'<th>Aksi</th>':''}</tr>
      ${list.length?list.map(s=>`<tr><td><strong>${s.nama}</strong></td><td>${s.kamar_nama}</td><td>${s.alamat||'-'}</td><td>${s.wali_nama||'-'}</td>
        <td style="font-size:.78rem">${s.no_hp||'-'}</td>
        <td><span class="badge-h">${s.status}</span></td>
        ${isAdmin?`<td style="white-space:nowrap"><button class="btn btn-primary btn-sm" onclick="showEditSantri(${s.id})" title="Edit"><i class="ri-edit-line"></i></button> <button class="btn btn-danger btn-sm" onclick="deleteSantri(${s.id},'${s.nama.replace(/'/g,"\\\\'")}')" title="Hapus"><i class="ri-delete-bin-line"></i></button></td>`:''}</tr>`).join('')
      :'<tr><td colspan="7" style="text-align:center;color:var(--t3);padding:2rem">Tidak ada data yang cocok</td></tr>'}
    </table></div></div>`;
}
function filterSantriTable(q){
  if(!q){renderSantriRows(_santriFullList);return;}
  const lq=q.toLowerCase();
  const filtered=_santriFullList.filter(s=>(s.nama||'').toLowerCase().includes(lq)||(s.kamar_nama||'').toLowerCase().includes(lq)||(s.wali_nama||'').toLowerCase().includes(lq)||(s.alamat||'').toLowerCase().includes(lq)||(s.no_hp||'').includes(lq));
  renderSantriRows(filtered);
}

function deleteSantri(id,nama){if(!confirm('Hapus santri "'+nama+'"?\n\nSemua data terkait akan ikut terhapus:\n- Absensi, Nilai, Catatan Guru\n- Pelanggaran, Pembayaran\n\nTIDAK BISA dibatalkan!'))return;api('/api/santri/'+id,{method:'DELETE'}).then(()=>{toast('Santri dihapus');loadSantri();});}
function showImportExcel(){
  $('modal').innerHTML=`<h3>Import Santri dari Excel</h3>
    <p style="font-size:.85rem;color:var(--text-muted);margin-bottom:.5rem">Format: Kolom Nama, Alamat, Nama Wali, No HP</p>
    <p style="font-size:.78rem;color:var(--t3);margin-bottom:1rem">Santri dengan nama yang sudah ada akan otomatis dilewati (tidak duplikat)</p>
    <input type="file" id="excelFile" accept=".xlsx,.xls" style="margin-bottom:1rem;color:var(--text)">
    <div id="importProgress" style="display:none;margin-bottom:.8rem;padding:.8rem;background:rgba(59,130,246,.08);border-radius:8px;font-size:.85rem">
      <div style="display:flex;align-items:center;gap:.5rem"><span class="spinner"></span> <span id="importStatus">Mengimport data santri... Mohon tunggu</span></div>
    </div>
    <div id="importResult" style="display:none;margin-bottom:.8rem;padding:.8rem;background:rgba(22,163,74,.08);border:1px solid rgba(22,163,74,.2);border-radius:8px;font-size:.85rem"></div>
    <div style="display:flex;gap:.5rem"><button class="btn btn-primary" id="importBtn" onclick="doImportExcel()">Upload</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function doImportExcel(){
  if(window._importingExcel)return toast('Import sedang berjalan...');
  const file=$('excelFile').files[0];if(!file)return toast('Pilih file');
  window._importingExcel=true;
  const btn=$('importBtn');btn.disabled=true;btn.textContent='Mengimport...';btn.style.opacity='0.5';btn.style.pointerEvents='none';
  const prog=$('importProgress');if(prog){prog.style.display='block';$('importStatus').textContent='Mengimport '+file.name+'... Mohon tunggu, jangan tutup halaman';}
  try{
    const fd=new FormData();fd.append('file',file);const h={};if(token)h['Authorization']='Bearer '+token;
    const r=await fetch('/api/santri/import-excel',{method:'POST',headers:h,body:fd});const res=await r.json();
    if(prog)prog.style.display='none';
    if(r.status===429){toast(res.message||'Import sedang berjalan');btn.textContent='Tunggu...';return;}
    const resBox=$('importResult');
    if(res.imported!==undefined&&resBox){
      resBox.style.display='block';
      resBox.innerHTML=`<strong>${res.imported}</strong> santri diimport<br><strong>${res.skipped||0}</strong> dilewati (sudah ada)${res.wali_created?'<br><strong>'+res.wali_created+'</strong> akun wali dibuat':''}`;
    }
    if(res.message)toast(res.message);
    btn.textContent='Selesai';
    setTimeout(()=>{window._importingExcel=false;hideModal();loadSantri();},2000);
  }catch(e){
    if(prog)prog.style.display='none';
    window._importingExcel=false;btn.disabled=false;btn.style.opacity='1';btn.style.pointerEvents='';btn.textContent='Upload';
    toast('Error: '+e.message);
  }
}
function showAddSantri(){
  $('modal').innerHTML=`<h3>Tambah Santri</h3>
    <div class="fg"><label>Nama Santri</label><input id="sNama" placeholder="Nama lengkap santri"></div>
    <div class="fg"><label>Alamat</label><input id="sAlamat" placeholder="Alamat santri"></div>
    <div class="fg"><label>Kamar</label><select id="sKamar"><option value="">-- Pilih Kamar --</option></select></div>
    <div class="fg"><label>Kelas Diniyyah</label><input id="sKelas" placeholder="Contoh: Kelas 1"></div>
    <div class="fg"><label>Nama Wali</label><input id="sWali" placeholder="Nama orang tua/wali"></div>
    <div class="fg"><label>No. Telepon Wali</label><input id="sHP" placeholder="08xxxxxxxxxx" type="tel"></div>
    <div class="fg"><label>Kategori SPP</label><select id="sKatSpp"><option value="0">-- Default/Reguler --</option></select></div>
    <p style="font-size:.72rem;color:var(--t3);margin-top:-.3rem">* Jika diisi, otomatis dibuat akun login wali (username & password = no HP)</p>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem"><button class="btn btn-primary" onclick="saveSantri()">Simpan</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  api('/api/kamar').then(k=>{$('sKamar').innerHTML='<option value="">-- Pilih --</option>'+k.map(x=>'<option value="'+x.id+'">'+x.nama+'</option>').join('');}).catch(()=>{});
  api('/api/pembayaran/kategori').then(k=>{
    var arr=Array.isArray(k)?k:[];
    $('sKatSpp').innerHTML='<option value="0">-- Default/Reguler --</option>'+arr.filter(x=>x.aktif).map(x=>'<option value="'+x.id+'">'+x.nama+' ('+fmtRp(x.nominal)+')</option>').join('');
  }).catch(()=>{});
  showModal();
}
async function saveSantri(){const b={nama:$('sNama').value,kamar_id:parseInt($('sKamar').value)||1,kelas_diniyyah:$('sKelas').value,alamat:$('sAlamat').value,nama_wali:$('sWali').value,no_hp:$('sHP').value,kategori_spp_id:parseInt($('sKatSpp').value)||0};if(!b.nama)return toast('Nama wajib');await api('/api/santri',{method:'POST',body:JSON.stringify(b)});hideModal();toast(b.no_hp?'Ditambahkan (Akun wali dibuat otomatis)':'Ditambahkan');loadSantri();}
function showEditSantri(id){
  const s = _santriFullList.find(x=>x.id===id);
  if(!s)return toast('Santri tidak ditemukan');
  $('modal').innerHTML=`<h3>Edit Santri</h3>
    <input type="hidden" id="eId" value="${s.id}">
    <div class="fg"><label>Nama Santri</label><input id="eNama" value="${s.nama}" placeholder="Nama lengkap santri"></div>
    <div class="fg"><label>Status</label><select id="eStatus"><option value="aktif">Aktif</option><option value="lulus">Lulus</option><option value="pindah">Pindah</option></select></div>
    <div class="fg"><label>Alamat</label><input id="eAlamat" value="${s.alamat||''}" placeholder="Alamat santri"></div>
    <div class="fg"><label>Kamar</label><select id="eKamar"><option value="">-- Pilih Kamar --</option></select></div>
    <div class="fg"><label>Kelas Diniyyah</label><input id="eKelas" value="${s.kelas_diniyyah||''}" placeholder="Contoh: Kelas 1"></div>
    <div class="fg"><label>Nama Wali</label><input id="eWali" value="${s.wali_nama||''}" placeholder="Nama orang tua/wali"></div>
    <div class="fg"><label>No. Telepon Wali</label><input id="eHP" value="${s.no_hp||''}" placeholder="08xxxxxxxxxx" type="tel"></div>
    <div class="fg"><label>Kategori SPP</label><select id="eKatSpp"><option value="0">-- Default/Reguler --</option></select></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem"><button class="btn btn-primary" onclick="saveEditSantri()">Simpan Perubahan</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  $('eStatus').value = s.status || 'aktif';
  api('/api/kamar').then(k=>{
    $('eKamar').innerHTML='<option value="">-- Pilih --</option>'+k.map(x=>'<option value="'+x.id+'" '+(x.id===s.kamar_id?'selected':'')+'>'+x.nama+'</option>').join('');
  }).catch(()=>{});
  api('/api/pembayaran/kategori').then(k=>{
    var arr=Array.isArray(k)?k:[];
    $('eKatSpp').innerHTML='<option value="0">-- Default/Reguler --</option>'+arr.filter(x=>x.aktif).map(x=>'<option value="'+x.id+'" '+(x.id===s.kategori_spp_id?'selected':'')+'>'+x.nama+' ('+fmtRp(x.nominal)+')</option>').join('');
  }).catch(()=>{});
  showModal();
}
async function saveEditSantri(){
  const id = $('eId').value;
  const b={nama:$('eNama').value, status:$('eStatus').value, kamar_id:parseInt($('eKamar').value)||1,kelas_diniyyah:$('eKelas').value,alamat:$('eAlamat').value,nama_wali:$('eWali').value,no_hp:$('eHP').value,kategori_spp_id:parseInt($('eKatSpp').value)||0};
  if(!b.nama)return toast('Nama wajib');
  try {
    await api('/api/santri/'+id,{method:'PUT',body:JSON.stringify(b)});
    hideModal();toast('Data diperbarui');loadSantri();
  } catch(e) {
    toast('Error: '+e.message);
  }
}
// RAPORT
let _raportSantriList=[];
async function loadRaportAbsensi(){
  if(user.role === 'wali') return loadWaliRaportDashboard();
  _raportSantriList=await api('/api/santri');
  var now=new Date(Date.now()+7*3600000);
  var monthStart=now.toISOString().slice(0,8)+'01';
  var today=now.toISOString().slice(0,10);
  $('main').innerHTML='<div class="page-header fade-up"><h2>Raport Santri</h2></div><div class="card fade-up"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end"><div class="fg" style="flex:2;min-width:200px;position:relative"><label>Cari Santri</label><input type="text" id="raportSearch" placeholder="Ketik nama santri..." oninput="searchRaportSantri(this.value)" autocomplete="off"><input type="hidden" id="raportSantri"><div id="raportSearchResults" class="search-results"></div></div><div class="fg" style="flex:1;min-width:130px"><label>Tanggal Mulai</label><input type="date" id="raportMulai" value="'+monthStart+'"></div><div class="fg" style="flex:1;min-width:130px"><label>Tanggal Akhir</label><input type="date" id="raportAkhir" value="'+today+'"></div><div class="fg" style="flex:1;min-width:120px"><label>Bulan Nilai</label><input type="month" id="raportBulanNilai" value="'+now.toISOString().slice(0,7)+'"></div><button class="btn btn-primary btn-sm" onclick="loadRaportData()">Tampilkan</button><button class="btn btn-gold btn-sm" onclick="downloadSemuaRaport()">Download Semua (ZIP)</button></div></div><div id="raportResult"></div>';
}
function searchRaportSantri(q){
  var box=$('raportSearchResults');
  if(!q||q.length<1){box.innerHTML='';return;}
  var results=_raportSantriList.filter(function(s){return s.nama.toLowerCase().includes(q.toLowerCase());}).slice(0,8);
  box.innerHTML=results.map(function(s){return '<div onclick="pickRaportSantri('+s.id+',\''+s.nama.replace(/'/g,"\\'")+'\')">'+ s.nama+'</div>';}).join('');
}
function pickRaportSantri(id,nama){$('raportSantri').value=id;$('raportSearch').value=nama;$('raportSearchResults').innerHTML='';}
async function loadRaportData(){
  var sid=$('raportSantri').value;
  if(!sid)return toast('Pilih santri dulu');
  var mulai=$('raportMulai').value,akhir=$('raportAkhir').value;
  var bulanNilai=$('raportBulanNilai')?$('raportBulanNilai').value:'';
  var qp='tgl_mulai='+mulai+'&tgl_akhir='+akhir+(bulanNilai?'&bulan_nilai='+bulanNilai:'');
  var d2=await Promise.all([api('/api/raport/'+sid+'?'+qp),api('/api/settings').catch(function(){return {};})]);
  var data=d2[0],settings=d2[1];
  var rekap=data.rekap||{},catatan=data.catatan_guru||[],pelang=data.pelanggaran||[],s=data.santri||{};
  var rm=data.rekap_malam||{H:0,I:0,S:0,A:0}, rs=data.rekap_sekolah||{H:0,I:0,S:0,A:0,jumlah_sesi:0};
  var rd=data.rekap_diniyyah||{};
  var nk=data.nilai_kegiatan||[], rrk=data.rata_rata_kegiatan||0, pkList=data.peringkat_kelompok||[];
  var lembaga=(settings&&settings.app_name)||'Pesantren';
  var rsSesi=rs.jumlah_sesi||0;
  var rsTotal=(rs.H||0)+(rs.I||0)+(rs.S||0)+(rs.A||0);
  var hasSekolah=rsTotal>0;
  var pelHtml='<p style="color:var(--t3)">Tidak ada pelanggaran</p>';
  if(pelang.length){
    pelHtml='<div class="table-wrap"><table><tr><th>Tanggal</th><th>Jenis</th><th>Deskripsi</th><th style="text-align:center">Poin</th><th style="text-align:center">Takzir</th><th style="text-align:right">Denda</th></tr>';
    pelang.forEach(function(p){
      var tb=(p.status_takzir||'belum')==='sudah'?'<span style="display:inline-block;padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:600;background:#dcfce7;color:#16a34a">V</span>':'<span style="display:inline-block;padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:600;background:#fef2f2;color:#dc2626">X</span>';
      var jt=p.jenis_takzir?'<br><span style="font-size:.68rem;color:var(--t3)">'+p.jenis_takzir+'</span>':'';
      var dd=p.denda?fmtRp(p.denda):'-';
      pelHtml+='<tr><td>'+p.tanggal+'</td><td>'+p.jenis+'</td><td>'+(p.deskripsi||'-')+'</td><td style="text-align:center"><span class="badge-a">'+(p.poin||0)+'</span></td><td style="text-align:center">'+tb+jt+'</td><td style="text-align:right;color:'+(p.denda?'#dc2626':'var(--t3)')+'">'+dd+'</td></tr>';
    });
    pelHtml+='</table></div>';
  }
  var nkHtml='';
  if(nk.length){
    nkHtml='<div class="table-wrap"><table><tr><th>Kegiatan</th><th>Kelompok</th><th style="text-align:center">Nilai</th><th>Catatan</th></tr>';
    nk.forEach(function(n){nkHtml+='<tr><td>'+n.kegiatan+'</td><td>'+n.kelompok+'</td><td style="text-align:center"><strong>'+n.nilai+'</strong></td><td>'+(n.catatan||'-')+'</td></tr>';});
    nkHtml+='<tr style="font-weight:700;background:rgba(59,130,246,.05)"><td colspan="2">Rata-rata</td><td style="text-align:center">'+rrk+'</td><td></td></tr></table></div>';
    if(pkList.length){nkHtml+='<div style="margin-top:.6rem;display:flex;gap:.5rem;flex-wrap:wrap">';pkList.forEach(function(p){nkHtml+='<div style="padding:.5rem .8rem;background:linear-gradient(135deg,#f0fdf4,#dcfce7);border-radius:10px;border:1px solid rgba(22,163,74,.15);font-size:.82rem"><i class="ri-trophy-line" style="color:#16a34a"></i> <strong>'+p.kelompok+'</strong>: Peringkat '+p.peringkat+' / '+p.total_anggota+'</div>';});nkHtml+='</div>';}
  }else{nkHtml=bulanNilai?'<p style="color:var(--t3)">Belum ada data nilai kegiatan</p>':'<p style="color:var(--t3)">Isi Bulan Nilai untuk melihat nilai kegiatan</p>';}
  var h='';
  h+='<div class="card raport-header fade-up"><h2>'+lembaga.toUpperCase()+'</h2><div class="raport-sub">RAPOR SANTRI</div></div>';
  h+='<div class="card fade-up"><table class="raport-identity"><tr><td>Nama</td><td>: '+(s.nama||'-')+'</td></tr><tr><td>Kamar</td><td>: '+(s.kamar_nama||'-')+'</td></tr><tr><td>Wali</td><td>: '+(s.wali_nama||'-')+'</td></tr><tr><td>Periode</td><td>: '+mulai+' s/d '+akhir+'</td></tr></table></div>';
  h+='<div class="card fade-up"><h3><i class="ri-bar-chart-box-line"></i> Rekap Absensi Kegiatan</h3><div class="table-wrap"><table><tr><th>Kegiatan</th><th style="text-align:center">Hadir</th><th style="text-align:center">Izin</th><th style="text-align:center">Sakit</th><th style="text-align:center">Alpa</th><th style="text-align:center">Total</th></tr>';
  Object.entries(rekap).forEach(function(e){var k=e[0],r=e[1];h+='<tr><td>'+k+'</td><td style="text-align:center"><span class="badge-h">'+r.H+'</span></td><td style="text-align:center"><span class="badge-i">'+r.I+'</span></td><td style="text-align:center"><span class="badge-s">'+r.S+'</span></td><td style="text-align:center"><span class="badge-a">'+r.A+'</span></td><td style="text-align:center;font-weight:700">'+(r.H+r.I+r.S+r.A)+'</td></tr>';});
  if(!Object.keys(rekap).length)h+='<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>';
  h+='</table></div></div>';
  h+='<div class="card fade-up"><h3><i class="ri-moon-line"></i> Rekap Absen Kamar</h3><div class="table-wrap"><table><tr><th>Hadir</th><th>Izin</th><th>Sakit</th><th>Alpa</th><th>Total</th></tr><tr><td style="text-align:center"><span class="badge-h">'+rm.H+'</span></td><td style="text-align:center"><span class="badge-i">'+rm.I+'</span></td><td style="text-align:center"><span class="badge-s">'+rm.S+'</span></td><td style="text-align:center"><span class="badge-a">'+rm.A+'</span></td><td style="text-align:center;font-weight:700">'+(rm.H+rm.I+rm.S+rm.A)+'</td></tr></table></div></div>';
  h+='<div class="card fade-up"><h3><i class="ri-school-line"></i> Rekap Absen Sekolah</h3>'+(rsSesi?'<div style="font-size:.8rem;color:var(--t2);margin-bottom:.5rem">Total sesi: <strong>'+rsSesi+' sesi</strong></div>':'')+'<div class="table-wrap"><table><tr><th>Keterangan</th><th style="text-align:center">Hadir</th><th style="text-align:center">Izin</th><th style="text-align:center">Sakit</th><th style="text-align:center">Alpa</th><th style="text-align:center">Total Sesi</th></tr>'+(hasSekolah?'<tr><td>Sekolah</td><td style="text-align:center"><span class="badge-h">'+rs.H+'</span></td><td style="text-align:center"><span class="badge-i">'+rs.I+'</span></td><td style="text-align:center"><span class="badge-s">'+rs.S+'</span></td><td style="text-align:center"><span class="badge-a">'+rs.A+'</span></td><td style="text-align:center;font-weight:700">'+rsTotal+'</td></tr>':'<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada data absen sekolah</td></tr>')+'</table></div></div>';
  h+='<div class="card fade-up"><h3><i class="ri-book-2-line"></i> Rekap Absen Madrasah Diniyyah</h3><div class="table-wrap"><table><tr><th>Kelas</th><th style="text-align:center">Hadir</th><th style="text-align:center">Izin</th><th style="text-align:center">Sakit</th><th style="text-align:center">Alpa</th><th style="text-align:center">Total</th></tr>';
  Object.entries(rd).forEach(function(e){var k=e[0],r=e[1];h+='<tr><td>'+k+'</td><td style="text-align:center"><span class="badge-h">'+r.H+'</span></td><td style="text-align:center"><span class="badge-i">'+r.I+'</span></td><td style="text-align:center"><span class="badge-s">'+r.S+'</span></td><td style="text-align:center"><span class="badge-a">'+r.A+'</span></td><td style="text-align:center;font-weight:700">'+(r.H+r.I+r.S+r.A)+'</td></tr>';});
  if(!Object.keys(rd).length)h+='<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada data absen diniyyah</td></tr>';
  h+='</table></div></div>';
  h+='<div class="card fade-up"><h3>Catatan Guru</h3>';
  if(catatan.length){catatan.forEach(function(c){h+='<div style="padding:.6rem 0;border-bottom:1px solid var(--border)"><div style="font-size:.78rem;color:var(--t3)">'+c.tanggal+'</div><div style="margin-top:.2rem">'+c.catatan+'</div></div>';});}else{h+='<p style="color:var(--t3)">Belum ada catatan</p>';}
  h+='</div>';
  h+='<div class="card fade-up"><h3>Pelanggaran</h3>'+pelHtml+'</div>';
  h+='<div class="card fade-up"><h3><i class="ri-star-line"></i> Nilai Kegiatan '+(bulanNilai?'('+bulanNilai+')':'')+'</h3>'+nkHtml+'</div>';
  h+='<div class="card fade-up" style="background:linear-gradient(135deg,#f0f9ff,#e0f2fe);border:1.5px solid rgba(59,130,246,.15)"><h3 style="margin-bottom:.8rem"><i class="ri-download-2-line"></i> Download Laporan - '+(s.nama||'Santri')+'</h3><div style="display:flex;gap:.6rem;flex-wrap:wrap">';
  h+='<button class="btn btn-gold" onclick="apiDownload(\'/api/raport/'+sid+'/pdf?tgl_mulai='+mulai+'&tgl_akhir='+akhir+'&bulan_nilai='+bulanNilai+'\',\''+(s.nama||'raport').replace(/'/g,'')+'.pdf\')" style="flex:1;min-width:140px"><i class="ri-file-pdf-2-line"></i> Download PDF</button>';
  h+='<button class="btn btn-primary" onclick="apiDownload(\'/api/raport/'+sid+'/excel?tgl_mulai='+mulai+'&tgl_akhir='+akhir+'&bulan_nilai='+bulanNilai+'\',\''+(s.nama||'raport').replace(/'/g,'')+'.xlsx\')" style="flex:1;min-width:140px"><i class="ri-file-excel-2-line"></i> Download Excel</button>';
  h+='</div></div>';
  $('raportResult').innerHTML=h;
}
async function downloadSemuaRaport(){
  var mulai=$('raportMulai').value,akhir=$('raportAkhir').value;
  if(!mulai||!akhir)return toast('Pilih periode dulu');
  toast('Mengunduh semua raport... Mohon tunggu');
  apiDownload('/api/raport-all/zip?tgl_mulai='+mulai+'&tgl_akhir='+akhir,'Raport_Semua_'+mulai+'_'+akhir+'.zip');
}
// WALI RAPORT (OLD DASHBOARD CONTENT)
async function loadWaliRaportDashboard() {
  const d=await api('/api/dashboard');
  const bulan=d.bulan||new Date().toISOString().slice(0,7);
  const anak=d.anak||[];
  let anakHtml='';
  if(!anak.length){
    anakHtml='<div class="card au"><p style="color:var(--t3)">Belum ada data anak terhubung ke akun Anda.</p></div>';
  } else {
    for(const a of anak){
      const ab=a.absensi||{};const kg=ab.kegiatan||{};const ml=ab.malam||{};const sk=ab.sekolah||{};const dn=ab.diniyyah||{};
      const kgEntries=Object.entries(kg);
      const pb=a.pembayaran||{};const cg=a.catatan_guru||[];const pl=a.pelanggaran||[];
      const stColor=pb.status==='LUNAS'?'var(--green)':pb.status==='KURANG'?'var(--amber)':'var(--red)';
      const badge=(v,c)=>`<span style="display:inline-block;min-width:24px;text-align:center;padding:2px 6px;border-radius:6px;font-weight:700;font-size:.78rem;background:${c}15;color:${c}">${v}</span>`;
      const abRow=(label,d)=>`<tr><td style="font-weight:600">${label}</td><td style="text-align:center">${badge(d.H||0,'#16a34a')}</td><td style="text-align:center">${badge(d.I||0,'#3b82f6')}</td><td style="text-align:center">${badge(d.S||0,'#f59e0b')}</td><td style="text-align:center">${badge(d.A||0,'#ef4444')}</td><td style="text-align:center;font-weight:700">${(d.H||0)+(d.I||0)+(d.S||0)+(d.A||0)}</td></tr>`;
      anakHtml+=`<div class="card au" style="margin-bottom:1rem">
        <h3 style="margin-bottom:.3rem"><i class="ri-graduation-cap-line" style="color:var(--p)"></i> ${a.nama}</h3>
        <div style="display:flex;gap:.8rem;flex-wrap:wrap;font-size:.78rem;color:var(--t3);margin-bottom:.8rem">
          <span><i class="ri-home-5-line"></i> ${a.kamar_nama||'-'}</span><span><i class="ri-book-2-line"></i> ${a.kelas_diniyyah||'-'}</span>
          <span style="color:${a.status==='aktif'?'var(--green)':'var(--red)'}">${a.status}</span>
        </div>

        <div style="background:linear-gradient(135deg,rgba(59,130,246,.04),rgba(139,92,246,.04));border:1px solid rgba(59,130,246,.1);border-radius:10px;padding:.8rem;margin-bottom:.8rem">
          <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-bar-chart-box-line"></i> Rekap Absensi - ${bulan}</h4>
          <div class="table-wrap"><table style="font-size:.78rem">
            <tr><th></th><th style="text-align:center">Hadir</th><th style="text-align:center">Izin</th><th style="text-align:center">Sakit</th><th style="text-align:center">Alpa</th><th style="text-align:center">Total</th></tr>
            ${kgEntries.length?kgEntries.map(([name,d])=>abRow(name,d)).join(''):'<tr><td style="font-weight:600">Kegiatan</td><td colspan="5" style="text-align:center;color:var(--t3);font-size:.75rem">Belum ada data</td></tr>'}${abRow('Malam',ml)}${abRow('Sekolah',sk)}${abRow('Diniyyah',dn)}
          </table></div>
        </div>

        <div style="background:rgba(245,158,11,.04);border:1px solid rgba(245,158,11,.1);border-radius:10px;padding:.8rem;margin-bottom:.8rem">
          <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-error-warning-line" style="color:var(--amber)"></i> Pelanggaran (${pl.length})${a.total_poin?' - <span style="color:var(--red)">'+a.total_poin+' poin</span>':''}</h4>
          ${pl.length?pl.map(p=>`<div style="padding:.3rem 0;border-bottom:1px solid var(--border);font-size:.78rem">
            <div style="font-weight:600;color:var(--red)">${p.jenis}</div>
            <div style="color:var(--t2)">${p.deskripsi||'-'}</div>
            <div style="font-size:.7rem;color:var(--t3)">${p.tanggal} - ${p.poin} poin</div>
          </div>`).join(''):'<p style="font-size:.78rem;color:var(--t3)">Tidak ada pelanggaran.</p>'}
        </div>

        <div style="background:rgba(59,130,246,.04);border:1px solid rgba(59,130,246,.1);border-radius:10px;padding:.8rem;margin-bottom:.8rem">
          <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-sticky-note-line" style="color:var(--p)"></i> Catatan Guru (${cg.length})</h4>
          ${cg.length?cg.map(c=>`<div style="padding:.3rem 0;border-bottom:1px solid var(--border);font-size:.78rem">
            <div style="color:var(--t1)">${c.catatan}</div>
            <div style="font-size:.7rem;color:var(--t3)">${c.tanggal} - ${c.guru||'-'}</div>
          </div>`).join(''):'<p style="font-size:.78rem;color:var(--t3)">Belum ada catatan</p>'}
        </div>

        <div style="display:flex;gap:.5rem;flex-wrap:wrap">
          <button class="btn btn-primary btn-sm" onclick="window._raportSantriId=${a.id};_raportReal()"><i class="ri-file-chart-line"></i> Raport Akademik</button>
        </div>
      </div>`;
    }
  }
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-file-chart-line"></i> Akademik & Absensi</h2></div>${anakHtml}`;
}

async function _raportReal() {
  $('main').innerHTML='<div class="card fade-up"><p style="color:var(--t3)"><i class="ri-loader-4-line"></i> Memuat raport...</p></div>';
  await loadRaportAbsensiReal();
  if($('raportSantri')) $('raportSantri').value=window._raportSantriId;
  loadRaportData();
}

async function loadRaportAbsensiReal(){
  _raportSantriList=await api('/api/santri');
  var now=new Date(Date.now()+7*3600000);
  var monthStart=now.toISOString().slice(0,8)+'01';
  var today=now.toISOString().slice(0,10);
  $('main').innerHTML='<div class="page-header fade-up"><button class="back-btn" onclick="loadWaliRaportDashboard()"><i class="ri-arrow-left-line"></i> Kembali</button><h2>Raport Santri</h2></div><div class="card fade-up"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end"><input type="hidden" id="raportSantri"><div class="fg" style="flex:1;min-width:130px"><label>Tanggal Mulai</label><input type="date" id="raportMulai" value="'+monthStart+'"></div><div class="fg" style="flex:1;min-width:130px"><label>Tanggal Akhir</label><input type="date" id="raportAkhir" value="'+today+'"></div><div class="fg" style="flex:1;min-width:120px"><label>Bulan Nilai</label><input type="month" id="raportBulanNilai" value="'+now.toISOString().slice(0,7)+'"></div><button class="btn btn-primary btn-sm" onclick="loadRaportData()">Tampilkan</button></div></div><div id="raportResult"></div>';
}

// REKAP USTADZ
function _tipeBadge(tipe){
  var m={kegiatan:{bg:'rgba(22,163,74,.1)',c:'#16a34a',icon:'ri-group-line',label:'Kegiatan'},sekolah:{bg:'rgba(59,130,246,.1)',c:'#3b82f6',icon:'ri-school-line',label:'Sekolah'},diniyyah:{bg:'rgba(139,92,246,.1)',c:'#8b5cf6',icon:'ri-book-2-line',label:'Diniyyah'},tahfidz:{bg:'rgba(13,148,136,.1)',c:'#0d9488',icon:'ri-book-read-line',label:'Tahfidz'}};
  var t=m[tipe]||m.kegiatan;
  return '<span style="display:inline-flex;align-items:center;gap:.2rem;padding:.15rem .5rem;background:'+t.bg+';color:'+t.c+';border-radius:6px;font-size:.7rem;font-weight:600"><i class="'+t.icon+'"></i> '+t.label+'</span>';
}
async function loadRekapUstadz(){
  const isAdmin=['admin','superadmin'].includes(user.role);const today=new Date(Date.now()+7*3600000).toISOString().slice(0,7);
  if(!isAdmin){const data=await api('/api/rekap-ustadz').catch(()=>[]);const todayStr=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
    $('main').innerHTML=`<div class="page-header fade-up"><h2>Rekap Saya - ${todayStr}</h2></div>
    <div class="card fade-up"><div class="table-wrap"><table><tr><th>Kelompok / Kelas</th><th>Tipe</th><th>Tanggal</th><th>Jam Absen (WIB)</th></tr>
    ${(data||[]).map(s=>`<tr><td>${s.kelompok_nama||'-'}</td><td>${_tipeBadge(s.tipe||'kegiatan')}</td><td>${s.tanggal}</td><td><span class="badge-h">${(s.recorded_at||'').slice(11,16)||'-'} WIB</span></td></tr>`).join('')}
    ${!(data||[]).length?'<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada</td></tr>':''}
    </table></div></div>`;return;}
  const now=new Date(Date.now()+7*3600000);const thisMonth=now.toISOString().slice(0,7);
  const defaultDari=$('ruDari')?$('ruDari').value:thisMonth+'-01';
  const defaultSampai=$('ruSampai')?$('ruSampai').value:now.toISOString().slice(0,10);
  let data;try{data=await api('/api/rekap-ustadz/summary?dari='+defaultDari+'&sampai='+defaultSampai);}catch(e){data={};}
  if(!data||typeof data!=='object'||Array.isArray(data))data={};
  const ustadzList=Object.entries(data);
  $('main').innerHTML=`<div class="page-header fade-up"><h2>Rekap Ustadz</h2></div>
  <div class="card fade-up" style="margin-bottom:1rem"><div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end">
    <div class="fg" style="flex:1;min-width:130px;margin:0"><label>Dari Tanggal</label><input type="date" id="ruDari" value="${defaultDari}" onchange="loadRekapUstadz()"></div>
    <div class="fg" style="flex:1;min-width:130px;margin:0"><label>Sampai Tanggal</label><input type="date" id="ruSampai" value="${defaultSampai}" onchange="loadRekapUstadz()"></div>
  </div></div>
  <div class="grid" id="ruGrid">${ustadzList.map(([nama,info])=>{
    const sesi=info.sesi||[];
    const tipeCount={kegiatan:0,sekolah:0,diniyyah:0,tahfidz:0};
    sesi.forEach(s=>{if(s.tipe)tipeCount[s.tipe]++;});
    const badges=[];
    if(tipeCount.kegiatan)badges.push('<span style="font-size:.7rem;color:#16a34a">Kegiatan: '+tipeCount.kegiatan+'x</span>');
    if(tipeCount.sekolah)badges.push('<span style="font-size:.7rem;color:#3b82f6">Sekolah: '+tipeCount.sekolah+'x</span>');
    if(tipeCount.diniyyah)badges.push('<span style="font-size:.7rem;color:#8b5cf6">Diniyyah: '+tipeCount.diniyyah+'x</span>');
    if(tipeCount.tahfidz)badges.push('<span style="font-size:.7rem;color:#0d9488">Tahfidz: '+tipeCount.tahfidz+'x</span>');
    return `<div class="grid-card fade-up" onclick="openRekapUstadzDetail('${nama.replace(/'/g,"\\'")}','${defaultDari}','${defaultSampai}')">
    <div class="title">${nama}</div>
    <div class="sub" style="display:flex;flex-direction:column;gap:.2rem;margin-top:.3rem">${badges.join('')||'<span style="font-size:.7rem;color:var(--t3)">-</span>'}</div>
    <div class="badge">${info.total||0}x absen</div></div>`;
  }).join('')}
  ${!ustadzList.length?'<div style="color:var(--t3);padding:2rem;text-align:center;grid-column:1/-1">Belum ada data</div>':''}</div><div id="ruDetail"></div>`;
}
async function openRekapUstadzDetail(nama,dari,sampai){
  let data;try{data=await api('/api/rekap-ustadz/summary?dari='+dari+'&sampai='+sampai);}catch(e){data={};}
  if(!data||typeof data!=='object')data={};
  const info=data[nama];if(!info)return toast('Data tidak ditemukan');
  const sesi=info.sesi||[];
  const kegMap={};
  sesi.forEach(s=>{
    const key=(s.tipe||'kegiatan')+'|'+s.kegiatan;
    if(!kegMap[key])kegMap[key]={count:0,tipe:s.tipe||'kegiatan',label:s.kegiatan};
    kegMap[key].count++;
  });
  const tipeColors={kegiatan:'rgba(22,163,74,.1)',sekolah:'rgba(59,130,246,.1)',diniyyah:'rgba(139,92,246,.1)'};
  const tipeBorder={kegiatan:'rgba(22,163,74,.2)',sekolah:'rgba(59,130,246,.2)',diniyyah:'rgba(139,92,246,.2)'};
  const tipeTextC={kegiatan:'var(--green)',sekolah:'#3b82f6',diniyyah:'#8b5cf6'};
  $('ruDetail').innerHTML=`<div class="card" style="margin-top:1rem"><button class="back-btn" onclick="$('ruDetail').innerHTML=''">Tutup</button>
    <h3>${nama} (${dari} s/d ${sampai})</h3>
    <div style="display:flex;gap:.6rem;flex-wrap:wrap;margin-bottom:1rem">
      ${Object.values(kegMap).map(v=>`<span style="display:inline-flex;align-items:center;gap:.3rem;padding:.3rem .7rem;background:${tipeColors[v.tipe]};border:1px solid ${tipeBorder[v.tipe]};border-radius:8px;font-size:.75rem;font-weight:600;color:${tipeTextC[v.tipe]}"><i class="ri-book-open-line"></i> ${v.label} <b>${v.count}x</b></span>`).join('')}
    </div>
    <h4 style="font-size:.85rem;margin-bottom:.5rem"><i class="ri-list-check-2"></i> Detail Pengabsenan</h4>
    <div class="table-wrap"><table>
    <tr><th>Tanggal</th><th>Jam Absen</th><th>Tipe</th><th>Kegiatan / Mapel</th><th>Kelompok / Kelas</th></tr>
    ${sesi.map(s=>`<tr>
      <td>${s.tanggal}</td>
      <td><span class="badge-h">${(s.recorded_at||'').slice(11,16)||'-'} WIB</span></td>
      <td>${_tipeBadge(s.tipe||'kegiatan')}</td>
      <td><span style="font-weight:600;color:var(--p)">${s.kegiatan}</span></td>
      <td>${s.kelompok||'-'}</td>
    </tr>`).join('')}
    ${!sesi.length?'<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>':''}
    <tr style="font-weight:700;background:rgba(59,130,246,.05)"><td>TOTAL</td><td style="text-align:center">${info.total||0}x</td><td colspan="3"></td></tr></table></div></div>`;
}

// -- PSB ADMIN --
async function loadPSBAdmin(){
  const [pendaftar,kamar]=await Promise.all([api('/api/pendaftar'),api('/api/kamar')]);
  const subdomain=user.subdomain||window.location.hostname.split('.')[0];
  const psbUrl=window.location.origin+'/psb/'+subdomain;
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-user-add-line"></i> PSB Online (${pendaftar.length} pendaftar)</h2>
    <button class="btn btn-outline btn-sm" onclick="copyPSBLink()"><i class="ri-link"></i> Salin Link PSB</button></div>
    <div class="card fade-up" style="margin-bottom:.8rem;padding:.7rem 1rem;background:rgba(22,163,74,.05);border:1px solid rgba(22,163,74,.15)">
      <div style="display:flex;align-items:center;gap:.6rem;flex-wrap:wrap">
        <i class="ri-link" style="color:var(--green)"></i>
        <code id="psbLink" style="font-size:.78rem;color:var(--p);word-break:break-all">${psbUrl}</code>
      </div>
      <p style="font-size:.72rem;color:var(--t3);margin-top:.3rem">Bagikan link ini ke calon santri/wali untuk pendaftaran online</p>
    </div>
    ${pendaftar.length?`<div class="card fade-up"><div class="table-wrap"><table>
      <tr><th>Nama</th><th>JK</th><th>Asal Sekolah</th><th>No HP</th><th>Orang Tua</th><th>Tgl Daftar</th><th>Aksi</th></tr>
      ${pendaftar.map(p=>`<tr>
        <td><strong>${p.nama}</strong>${p.catatan?'<br><small style="color:var(--t3)">'+p.catatan+'</small>':''}</td>
        <td>${p.jenis_kelamin}</td>
        <td style="font-size:.78rem">${p.asal_sekolah||'-'}</td>
        <td style="font-size:.78rem">${p.no_hp||'-'}</td>
        <td style="font-size:.78rem">${p.nama_ayah||'-'} / ${p.nama_ibu||'-'}</td>
        <td style="font-size:.75rem;color:var(--t3)">${(p.created_at||'').slice(0,10)}</td>
        <td style="white-space:nowrap">
          <button class="btn btn-primary btn-sm" onclick="showTerimaPSB(${p.id},'${p.nama.replace(/'/g,"\\\\'")}')" ><i class="ri-check-line"></i> Terima</button>
          <button class="btn btn-danger btn-sm" onclick="tolakPSB(${p.id},'${p.nama.replace(/'/g,"\\\\'")}')" ><i class="ri-close-line"></i></button>
        </td></tr>`).join('')}
    </table></div></div>`
    :'<div class="card fade-up" style="text-align:center;padding:2rem;color:var(--t3)"><i class="ri-inbox-line" style="font-size:2rem;display:block;margin-bottom:.5rem"></i>Belum ada pendaftar baru</div>'}`;
  window._psbKamarList=kamar;
}
function copyPSBLink(){const l=$('psbLink').textContent;navigator.clipboard.writeText(l);toast('Link PSB disalin!');}
function showTerimaPSB(id,nama){
  const kamar=window._psbKamarList||[];
  $('modal').innerHTML=`<h3><i class="ri-check-double-line"></i> Terima: ${nama}</h3>
    <div class="fg"><label>Pilih Kamar</label><select id="psbKamar">
      ${kamar.map(k=>`<option value="${k.id}">${k.nama}</option>`).join('')}
    </select></div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="doTerimaPSB(${id})"><i class="ri-check-line"></i> Terima & Masukkan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function doTerimaPSB(id){
  const kamar_id=parseInt($('psbKamar').value)||1;
  const r=await api('/api/pendaftar/'+id+'/terima',{method:'PUT',body:JSON.stringify({kamar_id})});
  hideModal();toast(r.message||'Diterima');loadPSBAdmin();
}
async function tolakPSB(id,nama){
  if(!confirm('Tolak pendaftar "'+nama+'"?'))return;
  await api('/api/pendaftar/'+id+'/tolak',{method:'PUT'});
  toast('Pendaftar ditolak');loadPSBAdmin();
}

// PERIZINAN
let _perizinanSantriList=null;
let _perizinanCache=null;
function _pzFilterLocal(){
  var searchQ=$('pzSearchInput')?$('pzSearchInput').value.trim():'';
  if(!_perizinanCache||!$('pzGrid'))return;
  var now=new Date();
  var isAdmin=['admin','superadmin','keamanan'].includes(user.role);
  var filtered=searchQ?_perizinanCache.filter(function(p){return (p.santri_nama||'').toLowerCase().includes(searchQ.toLowerCase());}):_perizinanCache;
  var fmtDt=function(dt){if(!dt)return'-';return dt.replace('T',' ').slice(0,16);};
  var statusBadge=function(st){
    var m={aktif:{bg:'rgba(59,130,246,.1)',c:'#3b82f6',label:'Aktif'},selesai:{bg:'rgba(22,163,74,.1)',c:'#16a34a',label:'Selesai'},terlambat:{bg:'rgba(239,68,68,.1)',c:'#ef4444',label:'Terlambat'}};
    var s=m[st]||m.aktif;
    return '<span style="display:inline-flex;align-items:center;gap:.2rem;padding:.2rem .6rem;background:'+s.bg+';color:'+s.c+';border-radius:8px;font-size:.72rem;font-weight:700">'+s.label+'</span>';
  };
  var sisaWaktu=function(selesai){
    var end=new Date(selesai);var diff=end-now;
    if(diff<=0)return '<span style="color:#ef4444;font-weight:600">Sudah lewat</span>';
    var jam=Math.floor(diff/3600000);var mnt=Math.floor((diff%3600000)/60000);
    if(jam>=24){var hr=Math.floor(jam/24);return '<span style="color:#3b82f6;font-weight:600">'+hr+' hari '+(jam%24)+' jam lagi</span>';}
    return '<span style="color:#f59e0b;font-weight:600">'+jam+' jam '+mnt+' menit lagi</span>';
  };
  pzGrid.innerHTML=filtered.map(function(p){
    var isAktif=p.status==='aktif';var deadline=new Date(p.tanggal_selesai);var lewat=deadline<now&&isAktif;
    var borderColor=isAktif?(lewat?'#f59e0b':'#3b82f6'):p.status==='selesai'?'#16a34a':'#ef4444';
    var nama=p.santri_nama||'';var namaEsc=nama.replace(/'/g,"\\'");
    return '<div class="grid-card" style="border-left:4px solid '+borderColor+'">'+
      '<div style="display:flex;justify-content:space-between;align-items:start;margin-bottom:.4rem"><div class="title" style="font-size:.95rem"><i class="ri-user-line" style="color:var(--p)"></i> '+nama+'</div>'+statusBadge(isAktif&&lewat?'terlambat':p.status)+'</div>'+
      '<div style="font-size:.82rem;color:var(--t2);margin-bottom:.3rem"><i class="ri-file-text-line"></i> '+(p.keterangan||'-')+'</div>'+
      '<div style="display:flex;gap:.8rem;flex-wrap:wrap;font-size:.75rem;color:var(--t3);margin-bottom:.5rem"><span><i class="ri-time-line"></i> '+p.durasi+' '+p.tipe_durasi+'</span><span><i class="ri-calendar-line"></i> '+fmtDt(p.tanggal_mulai)+'</span><span>s/d '+fmtDt(p.tanggal_selesai)+'</span></div>'+
      (isAktif?'<div style="font-size:.78rem;margin-bottom:.5rem"><i class="ri-hourglass-line"></i> '+sisaWaktu(p.tanggal_selesai)+'</div>':'')+
      (p.status==='terlambat'?'<div style="font-size:.78rem;color:#ef4444;margin-bottom:.5rem"><i class="ri-error-warning-line"></i> Terlambat '+p.terlambat_durasi+' '+p.terlambat_tipe+'</div>':'')+
      (p.tanggal_kembali?'<div style="font-size:.75rem;color:var(--t3)"><i class="ri-login-circle-line"></i> Kembali: '+fmtDt(p.tanggal_kembali)+'</div>':'')+
      (isAdmin&&isAktif?'<div style="display:flex;gap:.4rem;margin-top:.6rem;flex-wrap:wrap"><button class="btn btn-primary btn-sm" onclick="perizinanKembali('+p.id+',\''+namaEsc+'\')"><i class="ri-check-line"></i> Sudah Kembali</button><button class="btn btn-danger btn-sm" onclick="showPerizinanTerlambat('+p.id+',\''+namaEsc+'\')"><i class="ri-error-warning-line"></i> Terlambat</button><button class="btn btn-outline btn-sm" onclick="hapusPerizinan('+p.id+')" title="Hapus"><i class="ri-delete-bin-line"></i></button></div>':'')+
    '</div>';
  }).join('')+(!filtered.length?'<div style="text-align:center;padding:2rem;color:var(--t3);grid-column:1/-1"><i class="ri-inbox-line" style="font-size:2rem;display:block;margin-bottom:.5rem;opacity:.4"></i>'+(searchQ?'Tidak ditemukan "'+searchQ+'"':'Tidak ada data')+'</div>':'');
  if(pzCount)pzCount.textContent=filtered.length+(searchQ?' hasil':' data');
}
async function loadPerizinan(){
  var filter=$('pzFilter')?$('pzFilter').value:'';
  _perizinanCache=await api('/api/perizinan'+(filter?'?status='+filter:''));
  _renderPerizinanFull(_perizinanCache);
}
function _renderPerizinanFull(list){
  var isAdmin=['admin','superadmin','keamanan'].includes(user.role);
  var filter=$('pzFilter')?$('pzFilter').value:'';
  var aktifCount=list.filter(function(p){return p.status==='aktif';}).length;
  var selesaiCount=list.filter(function(p){return p.status==='selesai';}).length;
  var terlambatCount=list.filter(function(p){return p.status==='terlambat';}).length;
  $('main').innerHTML='<div class="page-header fade-up"><h2><i class="ri-pass-valid-line"></i> Perizinan Santri</h2>'+
    (isAdmin?'<button class="btn btn-primary btn-sm" onclick="showAddPerizinan()"><i class="ri-add-line"></i> Buat Izin</button>':'')+
    '</div>'+
    '<div class="card fade-up" style="margin-bottom:.8rem"><div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:center">'+
      '<div style="position:relative;flex:1;min-width:200px"><i class="ri-search-line" style="position:absolute;left:.7rem;top:50%;transform:translateY(-50%);color:var(--t3);font-size:.9rem"></i><input type="text" id="pzSearchInput" placeholder="Cari nama santri..." oninput="_pzFilterLocal()" style="width:100%;padding:.45rem .8rem .45rem 2rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem;background:var(--card)"></div>'+
      '<select id="pzFilter" onchange="loadPerizinan()" style="padding:.45rem .8rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem;background:var(--card)"><option value="">Semua</option><option value="aktif"'+(filter==='aktif'?' selected':'')+'>Aktif</option><option value="selesai"'+(filter==='selesai'?' selected':'')+'>Selesai</option><option value="terlambat"'+(filter==='terlambat'?' selected':'')+'>Terlambat</option></select>'+
      '<button class="btn btn-outline btn-sm" onclick="showRekapPerizinan()" title="Rekap per santri"><i class="ri-bar-chart-box-line"></i> Rekap</button>'+
    '</div></div>'+
    '<div style="display:flex;gap:.5rem;margin-bottom:.8rem;flex-wrap:wrap">'+
      '<div style="display:flex;align-items:center;gap:.3rem;padding:.35rem .7rem;background:rgba(59,130,246,.08);border-radius:8px;font-size:.72rem;font-weight:600;color:#3b82f6"><i class="ri-time-line"></i> Aktif: '+aktifCount+'</div>'+
      '<div style="display:flex;align-items:center;gap:.3rem;padding:.35rem .7rem;background:rgba(22,163,74,.08);border-radius:8px;font-size:.72rem;font-weight:600;color:#16a34a"><i class="ri-check-line"></i> Selesai: '+selesaiCount+'</div>'+
      '<div style="display:flex;align-items:center;gap:.3rem;padding:.35rem .7rem;background:rgba(239,68,68,.08);border-radius:8px;font-size:.72rem;font-weight:600;color:#ef4444"><i class="ri-error-warning-line"></i> Terlambat: '+terlambatCount+'</div>'+
      '<span id="pzCount" style="font-size:.75rem;color:var(--t3);margin-left:auto;align-self:center">'+list.length+' data</span>'+
    '</div>'+
    '<div class="grid" id="pzGrid"></div>';
  _pzFilterLocal();
}

async function showRekapPerizinan(){
  var thisMonth=new Date(Date.now()+7*3600000).toISOString().slice(0,7);
  var dari=$('rekapPzDari')?$('rekapPzDari').value:thisMonth+'-01';
  var sampai=$('rekapPzSampai')?$('rekapPzSampai').value:'';
  if(!dari)dari=thisMonth+'-01';
  const list=await api('/api/perizinan');
  // Filter by date range
  var filtered=list.filter(function(p){
    var tgl=(p.tanggal_mulai||'').slice(0,10);
    if(dari&&tgl<dari)return false;
    if(sampai&&tgl>sampai)return false;
    return true;
  });
  const map={};
  filtered.forEach(function(p){
    var key=p.santri_id;
    if(!map[key])map[key]={nama:p.santri_nama||'?',total:0,aktif:0,selesai:0,terlambat:0};
    map[key].total++;
    if(p.status==='aktif')map[key].aktif++;
    else if(p.status==='selesai')map[key].selesai++;
    else if(p.status==='terlambat')map[key].terlambat++;
  });
  var rows=Object.keys(map).map(function(k){return map[k];});
  rows.sort(function(a,b){return b.total-a.total;});
  var totalAll=rows.reduce(function(s,r){return s+r.total;},0);
  var totalTerlambat=rows.reduce(function(s,r){return s+r.terlambat;},0);
  var html='<h3><i class="ri-bar-chart-box-line"></i> Rekap Perizinan per Santri</h3>'+
    '<div style="display:flex;gap:.5rem;margin:.8rem 0;flex-wrap:wrap;align-items:center">'+
      '<div class="fg" style="margin:0;flex:1;min-width:130px"><label>Dari</label><input type="date" id="rekapPzDari" value="'+dari+'" onchange="showRekapPerizinan()"></div>'+
      '<div class="fg" style="margin:0;flex:1;min-width:130px"><label>Sampai</label><input type="date" id="rekapPzSampai" value="'+(sampai||'')+'" onchange="showRekapPerizinan()"></div>'+
    '</div>'+
    '<div style="display:flex;gap:.6rem;margin-bottom:.6rem;flex-wrap:wrap">'+
      '<div style="padding:.3rem .6rem;background:var(--pbg);border-radius:8px;font-size:.72rem;font-weight:600;color:var(--p)">Total: '+totalAll+'</div>'+
      '<div style="padding:.3rem .6rem;background:rgba(239,68,68,.08);border-radius:8px;font-size:.72rem;font-weight:600;color:#ef4444">Terlambat: '+totalTerlambat+'</div>'+
      '<span style="font-size:.72rem;color:var(--t3);margin-left:auto">'+rows.length+' santri</span>'+
    '</div>'+
    '<div style="max-height:50vh;overflow-y:auto">'+
    '<div class="table-wrap"><table><thead><tr><th>#</th><th>Nama Santri</th><th>Total</th><th>Aktif</th><th>Selesai</th><th>Terlambat</th></tr></thead><tbody>'+
    rows.map(function(r,i){
      return '<tr><td>'+(i+1)+'</td><td style="font-weight:600">'+r.nama+'</td><td style="font-weight:700">'+r.total+'</td>'+
        '<td><span style="color:#3b82f6;font-weight:600">'+r.aktif+'</span></td>'+
        '<td><span style="color:#16a34a;font-weight:600">'+r.selesai+'</span></td>'+
        '<td><span style="color:'+(r.terlambat>0?'#ef4444':'var(--t3)')+';font-weight:700">'+r.terlambat+'</span></td></tr>';
    }).join('')+
    '</tbody></table></div></div>'+
    (!rows.length?'<div style="text-align:center;padding:1.5rem;color:var(--t3)">Tidak ada data di periode ini</div>':'')+
    '<div style="margin-top:1rem"><button class="btn btn-outline" onclick="hideModal()">Tutup</button></div>';
  $('modal').innerHTML=html;showModal();
}

async function showAddPerizinan(){
  if(!_perizinanSantriList)_perizinanSantriList=await api('/api/santri');
  var today=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  $('modal').innerHTML='<h3><i class="ri-pass-valid-line"></i> Buat Perizinan Baru</h3>'+
    '<div class="fg"><label>Santri</label><input type="text" id="pzSearch" placeholder="Ketik nama santri..." oninput="searchPerizinanSantri(this.value)" autocomplete="off"><input type="hidden" id="pzSantriId"><div id="pzSearchResults" class="search-results"></div></div>'+
    '<div class="fg"><label>Keterangan Izin</label><select id="pzKet" onchange="if(this.value===\'lainnya\')$(\'pzKetCustom\').style.display=\'block\'"><option value="Pulang ke rumah">Pulang ke rumah</option><option value="Sakit (dibawa pulang)">Sakit (dibawa pulang)</option><option value="Acara keluarga">Acara keluarga</option><option value="Keperluan penting">Keperluan penting</option><option value="lainnya">Lainnya...</option></select><input type="text" id="pzKetCustom" placeholder="Tulis keterangan..." style="display:none;margin-top:.4rem"></div>'+
    '<div style="display:grid;grid-template-columns:1fr 1fr;gap:.6rem"><div class="fg"><label>Tipe Durasi</label><select id="pzTipeDurasi"><option value="hari">Hari</option><option value="jam">Jam</option></select></div><div class="fg"><label>Durasi</label><input type="number" id="pzDurasi" min="1" value="1"></div></div>'+
    '<div class="fg"><label>Tanggal Mulai</label><input type="date" id="pzTglMulai" value="'+today+'"></div>'+
    '<div style="display:flex;gap:.5rem;margin-top:1.2rem"><button class="btn btn-primary" onclick="savePerizinan()"><i class="ri-save-line"></i> Simpan</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>';
  showModal();
}
function searchPerizinanSantri(q){
  var box=$('pzSearchResults');if(!q||q.length<2){box.innerHTML='';return;}
  var results=(_perizinanSantriList||[]).filter(function(s){return s.nama.toLowerCase().includes(q.toLowerCase());}).slice(0,8);
  box.innerHTML=results.map(function(s){return '<div onclick="pickPzSantri('+s.id+',\''+s.nama.replace(/'/g,"\\'")+'\')">'+ s.nama+' <span style="font-size:.72rem;color:var(--t3)">'+(s.kamar_nama||'')+'</span></div>';}).join('');
}
function pickPzSantri(id,nama){$('pzSantriId').value=id;$('pzSearch').value=nama;$('pzSearchResults').innerHTML='';}

async function savePerizinan(){
  var sid=parseInt($('pzSantriId').value);if(!sid)return toast('Pilih santri dulu');
  var ket=$('pzKet').value;if(ket==='lainnya')ket=$('pzKetCustom').value;
  if(!ket)return toast('Keterangan wajib diisi');
  var durasi=parseInt($('pzDurasi').value);if(!durasi||durasi<1)return toast('Durasi minimal 1');
  var body={santri_id:sid,keterangan:ket,tipe_durasi:$('pzTipeDurasi').value,durasi:durasi,tanggal_mulai:$('pzTglMulai').value};
  await api('/api/perizinan',{method:'POST',body:JSON.stringify(body)});
  hideModal();toast('Perizinan berhasil dibuat');loadPerizinan();
}

async function perizinanKembali(id,nama){
  if(!confirm('Tandai '+nama+' sudah kembali tepat waktu?'))return;
  await api('/api/perizinan/'+id+'/kembali',{method:'PUT'});
  toast('Santri telah kembali');loadPerizinan();
}

function showPerizinanTerlambat(id,nama){
  $('modal').innerHTML='<h3><i class="ri-error-warning-line" style="color:var(--red)"></i> Terlambat Kembali</h3>'+
    '<p style="font-size:.85rem;margin-bottom:.8rem"><strong>'+nama+'</strong> terlambat kembali dari izin.</p>'+
    '<div style="display:grid;grid-template-columns:1fr 1fr;gap:.6rem"><div class="fg"><label>Tipe</label><select id="pzLateTipe"><option value="hari">Hari</option><option value="jam">Jam</option></select></div><div class="fg"><label>Durasi Terlambat</label><input type="number" id="pzLateDurasi" min="1" value="1"></div></div>'+
    '<div style="padding:.6rem;background:rgba(239,68,68,.06);border:1px solid rgba(239,68,68,.15);border-radius:8px;margin:.8rem 0;font-size:.8rem;color:var(--red)"><i class="ri-error-warning-line"></i> Otomatis dicatat sebagai <strong>pelanggaran (1 poin)</strong></div>'+
    '<div style="display:flex;gap:.5rem"><button class="btn btn-danger" onclick="doPerizinanTerlambat('+id+')"><i class="ri-error-warning-line"></i> Konfirmasi</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>';
  showModal();
}
async function doPerizinanTerlambat(id){
  var durasi=parseInt($('pzLateDurasi').value);if(!durasi||durasi<1)return toast('Durasi minimal 1');
  await api('/api/perizinan/'+id+'/terlambat',{method:'PUT',body:JSON.stringify({terlambat_durasi:durasi,terlambat_tipe:$('pzLateTipe').value})});
  hideModal();toast('Terlambat dicatat & pelanggaran ditambahkan');loadPerizinan();
}

async function hapusPerizinan(id){
  if(!confirm('Hapus perizinan ini?'))return;
  await api('/api/perizinan/'+id,{method:'DELETE'});
  toast('Perizinan dihapus');loadPerizinan();
}

// ═══════════════════════════════════════════════════════════
// WALI: LAPORAN TAHFIDZ QUR'AN
// ═══════════════════════════════════════════════════════════

async function loadTahfidzWali(){
  const anak = await api('/api/wali/anak');
  if(!anak.length){
    $('main').innerHTML='<div class="page-header au"><h2><i class="ri-book-read-line"></i> Laporan Tahfidz</h2></div><div class="card au"><p style="color:var(--t3)">Belum ada data anak terhubung ke akun Anda.</p></div>';
    return;
  }

  let html='<div class="page-header au"><h2><i class="ri-book-read-line"></i> Laporan Tahfidz</h2></div>';

  for(const a of anak){
    let data;
    try { data = await api('/api/wali/tahfidz/'+a.id); } catch(e){ data = null; }

    if(!data || (!data.halaqoh?.id && !data.riwayat?.length)){
      html+=`<div class="card au" style="margin-bottom:1rem">
        <h3 style="margin-bottom:.3rem"><i class="ri-graduation-cap-line" style="color:var(--p)"></i> ${a.nama}</h3>
        <p style="color:var(--t3);font-size:.85rem">Belum terdaftar di halaqoh tahfidz manapun.</p>
      </div>`;
      continue;
    }

    const s = data.summary||{};
    const h = data.halaqoh||{};
    const rw = data.riwayat||[];
    const badge = (v,c)=>`<span style="display:inline-block;min-width:26px;text-align:center;padding:3px 8px;border-radius:8px;font-weight:700;font-size:.82rem;background:${c}12;color:${c}">${v}</span>`;
    const predColor = (p)=>{
      if(p==='Mumtaz') return '#16a34a';
      if(p==='Jayyid Jiddan') return '#0d9488';
      if(p==='Jayyid') return '#3b82f6';
      if(p==='Maqbul') return '#f59e0b';
      return '#ef4444';
    };

    // Progress bar
    const pct = Math.min(s.progress_pct||0, 100);
    const pctColor = pct>=80?'#16a34a':pct>=50?'#f59e0b':'#3b82f6';

    html+=`<div class="card au" style="margin-bottom:1.2rem">
      <h3 style="margin-bottom:.3rem"><i class="ri-graduation-cap-line" style="color:var(--p)"></i> ${a.nama}</h3>
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;font-size:.78rem;color:var(--t3);margin-bottom:1rem">
        <span><i class="ri-home-5-line"></i> ${a.kamar_nama||'-'}</span>
        <span style="color:${a.status==='aktif'?'var(--green)':'var(--red)'}">${a.status}</span>
      </div>

      <!-- Halaqoh Info -->
      ${h.id?`<div style="background:linear-gradient(135deg,rgba(13,148,136,.05),rgba(13,148,136,.02));border:1px solid rgba(13,148,136,.12);border-radius:12px;padding:1rem;margin-bottom:1rem">
        <h4 style="font-size:.85rem;color:#0d9488;margin-bottom:.6rem"><i class="ri-book-read-line"></i> Halaqoh: ${h.nama}</h4>
        <div style="display:flex;gap:1rem;flex-wrap:wrap;font-size:.78rem;color:var(--t2)">
          ${h.musyrif?`<span><i class="ri-user-heart-line"></i> Musyrif: <strong>${h.musyrif}</strong></span>`:''}
          ${h.target_juz?`<span><i class="ri-flag-line"></i> Target: <strong>Juz ${h.target_juz}</strong></span>`:''}
          ${h.target_ayat?`<span><i class="ri-file-list-3-line"></i> Target Ayat: <strong>${h.target_ayat}</strong></span>`:''}
          ${h.target_deadline?`<span><i class="ri-calendar-event-line"></i> Deadline: <strong>${h.target_deadline}</strong></span>`:''}
        </div>
        ${h.target_ayat?`<div style="margin-top:.8rem">
          <div style="display:flex;justify-content:space-between;font-size:.75rem;color:var(--t3);margin-bottom:.3rem">
            <span>Capaian: <strong style="color:#0d9488">${s.capaian_ayat||0}</strong> / ${h.target_ayat} ayat</span>
            <span style="font-weight:700;color:${pctColor}">${pct}%</span>
          </div>
          <div style="background:var(--border);border-radius:6px;height:8px;overflow:hidden">
            <div style="height:100%;width:${pct}%;background:linear-gradient(90deg,${pctColor},${pctColor}cc);border-radius:6px;transition:width .5s ease"></div>
          </div>
        </div>`:''}
      </div>`:''}

      <!-- Ringkasan Absensi -->
      <div style="background:linear-gradient(135deg,rgba(59,130,246,.04),rgba(139,92,246,.04));border:1px solid rgba(59,130,246,.1);border-radius:12px;padding:1rem;margin-bottom:1rem">
        <h4 style="font-size:.85rem;margin-bottom:.6rem"><i class="ri-bar-chart-box-line"></i> Ringkasan Tahfidz</h4>
        <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(100px,1fr));gap:.5rem;margin-bottom:.8rem">
          <div style="text-align:center;padding:.6rem;background:rgba(22,163,74,.06);border-radius:10px">
            <div style="font-size:1.4rem;font-weight:800;color:#16a34a">${s.H||0}</div>
            <div style="font-size:.7rem;color:var(--t3)">Hadir</div>
          </div>
          <div style="text-align:center;padding:.6rem;background:rgba(59,130,246,.06);border-radius:10px">
            <div style="font-size:1.4rem;font-weight:800;color:#3b82f6">${s.I||0}</div>
            <div style="font-size:.7rem;color:var(--t3)">Izin</div>
          </div>
          <div style="text-align:center;padding:.6rem;background:rgba(245,158,11,.06);border-radius:10px">
            <div style="font-size:1.4rem;font-weight:800;color:#f59e0b">${s.S||0}</div>
            <div style="font-size:.7rem;color:var(--t3)">Sakit</div>
          </div>
          <div style="text-align:center;padding:.6rem;background:rgba(239,68,68,.06);border-radius:10px">
            <div style="font-size:1.4rem;font-weight:800;color:#ef4444">${s.A||0}</div>
            <div style="font-size:.7rem;color:var(--t3)">Alpa</div>
          </div>
        </div>
        <div style="display:flex;gap:1rem;flex-wrap:wrap;font-size:.82rem">
          <div style="flex:1;min-width:120px;padding:.6rem;background:var(--bg);border-radius:10px;text-align:center">
            <div style="font-size:.7rem;color:var(--t3)">Total Sesi</div>
            <div style="font-size:1.1rem;font-weight:700;color:var(--t1)">${s.total_sesi||0}</div>
          </div>
          <div style="flex:1;min-width:120px;padding:.6rem;background:var(--bg);border-radius:10px;text-align:center">
            <div style="font-size:.7rem;color:var(--t3)">Rata-rata Nilai</div>
            <div style="font-size:1.1rem;font-weight:700;color:var(--t1)">${s.avg_nilai||0}</div>
          </div>
          ${s.predikat?`<div style="flex:1;min-width:120px;padding:.6rem;background:var(--bg);border-radius:10px;text-align:center">
            <div style="font-size:.7rem;color:var(--t3)">Predikat</div>
            <div style="font-size:1rem;font-weight:700;color:${predColor(s.predikat)}">${s.predikat}</div>
          </div>`:''}
        </div>
      </div>

      <!-- Riwayat Setoran -->
      <div style="background:rgba(13,148,136,.03);border:1px solid rgba(13,148,136,.1);border-radius:12px;padding:1rem">
        <h4 style="font-size:.85rem;color:#0d9488;margin-bottom:.6rem"><i class="ri-quill-pen-line"></i> Riwayat Setoran (${rw.length} terbaru)</h4>
        ${rw.length?`<div class="table-wrap"><table style="font-size:.78rem">
          <tr><th>Tanggal</th><th>Status</th><th>Jenis</th><th>Surah</th><th>Ayat</th><th>Juz</th><th style="text-align:center">Tajwid</th><th style="text-align:center">Lancar</th><th style="text-align:center">Makhraj</th><th style="text-align:center">Rata²</th><th>Predikat</th></tr>
          ${rw.map(r=>{
            const stBadge = r.status==='H'?badge('H','#16a34a'):r.status==='I'?badge('I','#3b82f6'):r.status==='S'?badge('S','#f59e0b'):badge('A','#ef4444');
            const jenisLabel = r.jenis_setoran==='ziyadah'?'<span style="color:#0d9488;font-weight:600">Ziyadah</span>':r.jenis_setoran==='murojaah'?'<span style="color:#8b5cf6;font-weight:600">Murojaah</span>':(r.jenis_setoran||'-');
            const ayatStr = r.surah?(r.ayat_dari===r.ayat_sampai?r.ayat_dari:(r.ayat_dari+'-'+r.ayat_sampai)):'-';
            const predBadge = r.predikat?`<span style="padding:2px 8px;border-radius:6px;font-size:.7rem;font-weight:700;background:${predColor(r.predikat)}15;color:${predColor(r.predikat)}">${r.predikat}</span>`:'';
            return `<tr>
              <td style="white-space:nowrap">${r.tanggal}</td>
              <td>${stBadge}</td>
              <td>${jenisLabel}</td>
              <td style="max-width:120px;overflow:hidden;text-overflow:ellipsis">${r.surah||'-'}</td>
              <td>${ayatStr}</td>
              <td style="text-align:center">${r.juz||'-'}</td>
              <td style="text-align:center;font-weight:600">${r.status==='H'&&r.nilai_tajwid?r.nilai_tajwid:'-'}</td>
              <td style="text-align:center;font-weight:600">${r.status==='H'&&r.nilai_kelancaran?r.nilai_kelancaran:'-'}</td>
              <td style="text-align:center;font-weight:600">${r.status==='H'&&r.nilai_makhorijul?r.nilai_makhorijul:'-'}</td>
              <td style="text-align:center;font-weight:700">${r.status==='H'&&r.nilai_rata?r.nilai_rata:'-'}</td>
              <td>${predBadge}</td>
            </tr>${r.catatan?`<tr><td colspan="11" style="padding:.2rem .5rem;font-size:.72rem;color:var(--t3);border-top:none"><i class="ri-chat-3-line"></i> ${r.catatan}</td></tr>`:''}`;
          }).join('')}
        </table></div>`:'<p style="color:var(--t3);font-size:.82rem">Belum ada riwayat setoran tahfidz.</p>'}
      </div>
    </div>`;
  }

  $('main').innerHTML=html;
}