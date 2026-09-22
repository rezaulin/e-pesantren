// -- Santri Search Picker --
let _allSantriList=[];
function searchSantriPick(prefix,q){
  const box=$(prefix+'SearchResults'),hidden=$(prefix+'Santri');
  if(!q||q.length<1){box.innerHTML='';hidden.value='';return;}
  const results=_allSantriList.filter(s=>s.nama.toLowerCase().includes(q.toLowerCase())).slice(0,8);
  box.innerHTML=results.map(s=>`<div onclick="pickSantri('${prefix}',${s.id},'${s.nama.replace(/'/g,"\\'")}')">${s.nama}</div>`).join('');
}
function pickSantri(prefix,id,nama){$(prefix+'Santri').value=id;$(prefix+'Search').value=nama;$(prefix+'SearchResults').innerHTML='';}

// -- CATATAN GURU --
async function loadCatatanGuru(){
  const [santri,catatan]=await Promise.all([api('/api/santri'),api('/api/catatan-guru')]);
  _allSantriList=santri;const today=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  $('main').innerHTML=`<div class="page-header fade-up"><h2> Catatan Guru</h2></div>
    <div class="card fade-up"><h3>Tambah Catatan</h3>
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
        <div class="fg" style="flex:2;min-width:200px"><label>Santri</label>
          <input type="text" id="cgSearch" placeholder="Ketik nama..." oninput="searchSantriPick('cg',this.value)" autocomplete="off">
          <input type="hidden" id="cgSantri"><div id="cgSearchResults" class="search-results"></div></div>
        <div class="fg" style="flex:1;min-width:130px"><label>Tanggal</label><input type="date" id="cgTanggal" value="${today}"></div>
        <div class="fg" style="flex:3;min-width:100%"><label>Catatan</label><textarea id="cgCatatan" rows="2" placeholder="Tulis catatan..."></textarea></div>
        <button class="btn btn-primary btn-sm" onclick="simpanCatatanGuru()"> Simpan</button></div></div>
    <div class="card fade-up"><h3>Daftar Catatan</h3><div class="table-wrap"><table>
      <tr><th>Tanggal</th><th>Santri</th><th>Catatan</th><th>Aksi</th></tr>
      ${catatan.length?catatan.map(c=>`<tr><td>${c.tanggal}</td><td>${c.santri_nama||'-'}</td><td>${c.catatan}</td>
        <td><button class="btn btn-danger btn-sm" onclick="hapusCatatanGuru(${c.id})"></button></td></tr>`).join('')
      :'<tr><td colspan="4" style="text-align:center;color:var(--text-dim)">Belum ada</td></tr>'}
    </table></div></div>`;
}
async function simpanCatatanGuru(){const sid=parseInt($('cgSantri').value)||0;if(!sid)return alert('Pilih santri terlebih dahulu');const cat=$('cgCatatan').value.trim(),tgl=$('cgTanggal').value;if(!cat)return alert('Catatan kosong');await api('/api/catatan-guru',{method:'POST',body:JSON.stringify({santri_id:sid,catatan:cat,tanggal:tgl})});loadCatatanGuru();}
async function hapusCatatanGuru(id){if(!confirm('Hapus?'))return;await api('/api/catatan-guru/'+id,{method:'DELETE'});loadCatatanGuru();}

// -- PELANGGARAN --
let _pelTab='list';
async function loadPelanggaran(){
  const [santri,pel]=await Promise.all([api('/api/santri'),api('/api/pelanggaran')]);
  _allSantriList=santri;const today=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  const tabList=_pelTab==='list'?'background:var(--p);color:#fff':'',tabRekap=_pelTab==='rekap'?'background:var(--p);color:#fff':'';
  $('main').innerHTML=`<div class="page-header fade-up"><h2>⚠️ Pelanggaran</h2></div>
    <div style="display:flex;gap:.4rem;margin-bottom:1rem" class="fade-up">
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabList}" onclick="_pelTab='list';loadPelanggaran()"><i class="ri-list-check"></i> Daftar</button>
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabRekap}" onclick="_pelTab='rekap';loadRekapTakzir()"><i class="ri-bar-chart-box-line"></i> Rekap Takzir</button>
    </div>`;
  if(_pelTab==='rekap'){loadRekapTakzir();return;}
  const totalPel=pel.length,sudahT=pel.filter(p=>p.status_takzir==='sudah').length,belumT=totalPel-sudahT;
  const totalDenda=pel.filter(p=>p.status_takzir==='sudah').reduce((s,p)=>s+(p.denda||0),0);
  $('main').innerHTML+=`
    <div class="stat-grid fade-up" style="grid-template-columns:repeat(4,1fr);margin-bottom:1rem">
      <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${totalPel}</div><div class="sl">Total</div></div><div class="si"><i class="ri-error-warning-line"></i></div></div>
      <div class="stat-card c-green"><div class="stat-info"><div class="sn">${sudahT}</div><div class="sl">Sudah Takzir</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
      <div class="stat-card c-amber"><div class="stat-info"><div class="sn">${belumT}</div><div class="sl">Belum Takzir</div></div><div class="si"><i class="ri-time-line"></i></div></div>
      <div class="stat-card c-red"><div class="stat-info"><div class="sn">${fmtRp(totalDenda)}</div><div class="sl">Total Denda</div></div><div class="si"><i class="ri-money-dollar-circle-line"></i></div></div>
    </div>
    <div class="card fade-up"><h3>Catat Pelanggaran</h3>
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
        <div class="fg" style="flex:2;min-width:200px"><label>Santri</label>
          <input type="text" id="plgSearch" placeholder="Ketik nama..." oninput="searchSantriPick('plg',this.value)" autocomplete="off">
          <input type="hidden" id="plgSantri"><div id="plgSearchResults" class="search-results"></div></div>
        <div class="fg" style="flex:1;min-width:130px"><label>Tanggal</label><input type="date" id="plgTanggal" value="${today}"></div>
        <div class="fg" style="flex:1;min-width:150px"><label>Jenis</label><input id="plgJenis" placeholder="Contoh: Terlambat"></div>
        <div class="fg" style="flex:1;min-width:80px"><label>Poin</label><input type="number" id="plgPoin" value="0" min="0"></div>
        <div class="fg" style="flex:3;min-width:100%"><label>Deskripsi</label><textarea id="plgDeskripsi" rows="2" placeholder="Detail..."></textarea></div>
        <button class="btn btn-primary btn-sm" onclick="simpanPelanggaran()"><i class="ri-save-line"></i> Simpan</button></div></div>
    <div class="card fade-up"><h3>Daftar Pelanggaran</h3><div class="table-wrap"><table>
      <tr><th>Tanggal</th><th>Santri</th><th>Jenis</th><th>Deskripsi</th><th style="text-align:center">Poin</th><th style="text-align:center">Status Takzir</th><th>Aksi</th></tr>
      ${pel.length?pel.map(p=>{
        const badge=p.status_takzir==='sudah'
          ?'<span style="display:inline-block;padding:.2rem .5rem;border-radius:8px;font-size:.68rem;font-weight:600;background:#dcfce7;color:#16a34a">✓ Sudah</span>'
          :'<span style="display:inline-block;padding:.2rem .5rem;border-radius:8px;font-size:.68rem;font-weight:600;background:#fef2f2;color:#dc2626">✗ Belum</span>';
        const dendaInfo=p.denda?'<br><span style="font-size:.68rem;color:#dc2626">Denda: '+fmtRp(p.denda)+'</span>':'';
        const takzirInfo=p.jenis_takzir?'<br><span style="font-size:.68rem;color:var(--t3)">'+p.jenis_takzir+'</span>':'';
        return `<tr><td>${p.tanggal}</td><td>${p.santri_nama||'-'}</td><td>${p.jenis}</td><td>${p.deskripsi||'-'}</td>
        <td style="text-align:center"><span class="badge-a">${p.poin||0}</span></td>
        <td style="text-align:center">${badge}${takzirInfo}${dendaInfo}</td>
        <td style="white-space:nowrap">
          <button class="btn btn-outline btn-sm" onclick="showTakzirModal(${p.id},'${(p.santri_nama||'').replace(/'/g,"\\'")}','${(p.jenis||'').replace(/'/g,"\\'")}','${p.status_takzir}','${(p.jenis_takzir||'').replace(/'/g,"\\'")}',${p.denda||0})" title="Takzir"><i class="ri-shield-check-line"></i></button>
          <button class="btn btn-danger btn-sm" onclick="hapusPelanggaran(${p.id})"><i class="ri-delete-bin-line"></i></button>
        </td></tr>`}).join('')
      :'<tr><td colspan="7" style="text-align:center;color:var(--text-dim)">Tidak ada</td></tr>'}
    </table></div></div>`;
}

function showTakzirModal(id,santriNama,jenisPel,currentStatus,currentJenis,currentDenda){
  const isSudah=currentStatus==='sudah';
  $('modal').innerHTML=`<h3><i class="ri-shield-check-line"></i> Takzir: ${santriNama}</h3>
    <p style="font-size:.82rem;color:var(--t2);margin-bottom:.8rem">Pelanggaran: <strong>${jenisPel}</strong></p>
    ${isSudah?'<div class="info-box success" style="margin-bottom:1rem"><i class="ri-checkbox-circle-line"></i> Sudah ditakzir: <strong>'+currentJenis+'</strong>'+(currentDenda?'<br>Denda: <strong>'+fmtRp(currentDenda)+'</strong>':'')+'</div>':''}
    <div class="fg"><label>Jenis Takziran</label>
      <input id="tzJenis" value="${currentJenis}" placeholder="Hafalan 1 Juz / Push up 50x / Denda dll">
    </div>
    <div class="fg">
      <label style="display:flex;align-items:center;gap:.5rem;cursor:pointer">
        <input type="checkbox" id="tzIsDenda" ${currentDenda?'checked':''} onchange="$('tzDendaWrap').style.display=this.checked?'block':'none'" style="accent-color:var(--p);width:18px;height:18px">
        <span style="font-weight:600">Termasuk Denda</span>
      </label>
    </div>
    <div id="tzDendaWrap" style="display:${currentDenda?'block':'none'}">
      <div class="fg"><label>Nominal Denda (Rp)</label>
        <input type="number" id="tzDenda" value="${currentDenda||0}" min="0" placeholder="50000">
      </div>
    </div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem;flex-wrap:wrap">
      <button class="btn btn-primary" onclick="simpanTakzir(${id},'sudah')"><i class="ri-checkbox-circle-line"></i> Tandai Sudah Takzir</button>
      ${isSudah?'<button class="btn btn-outline" onclick="simpanTakzir('+id+',\'belum\')"><i class="ri-close-circle-line"></i> Batalkan Takzir</button>':''}
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
  showModal();
}

async function simpanTakzir(id,status){
  const jenisTakzir=status==='sudah'?($('tzJenis')?.value?.trim()||''):'';
  const isDenda=status==='sudah'?($('tzIsDenda')?.checked||false):false;
  const denda=isDenda?parseInt($('tzDenda')?.value||0):0;
  if(status==='sudah'&&!jenisTakzir)return alert('Jenis takziran wajib diisi');
  await api('/api/pelanggaran/'+id+'/takzir',{method:'PUT',body:JSON.stringify({status_takzir:status,jenis_takzir:jenisTakzir,denda})});
  hideModal();toast(status==='sudah'?'✅ Takzir dicatat':'Takzir dibatalkan');loadPelanggaran();
}

async function simpanPelanggaran(){const sid=parseInt($('plgSantri').value)||0;if(!sid)return alert('Pilih santri terlebih dahulu');const jenis=$('plgJenis').value.trim(),desk=$('plgDeskripsi').value.trim(),poin=parseInt($('plgPoin').value)||0,tgl=$('plgTanggal').value;if(!jenis)return alert('Jenis wajib');await api('/api/pelanggaran',{method:'POST',body:JSON.stringify({santri_id:sid,jenis,deskripsi:desk,poin,tanggal:tgl})});loadPelanggaran();}
async function hapusPelanggaran(id){if(!confirm('Hapus?'))return;await api('/api/pelanggaran/'+id,{method:'DELETE'});loadPelanggaran();}

// -- REKAP TAKZIR --
async function loadRekapTakzir(){
  const rekap=await api('/api/pelanggaran/rekap');
  const totalDenda=rekap.reduce((s,r)=>s+(r.total_denda||0),0);
  const totalPel=rekap.reduce((s,r)=>s+(r.total||0),0);
  const totalSudah=rekap.reduce((s,r)=>s+(r.sudah_takzir||0),0);
  const tabList=_pelTab==='list'?'background:var(--p);color:#fff':'',tabRekap=_pelTab==='rekap'?'background:var(--p);color:#fff':'';
  $('main').innerHTML=`<div class="page-header fade-up"><h2>⚠️ Pelanggaran</h2></div>
    <div style="display:flex;gap:.4rem;margin-bottom:1rem" class="fade-up">
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabList}" onclick="_pelTab='list';loadPelanggaran()"><i class="ri-list-check"></i> Daftar</button>
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabRekap}" onclick="_pelTab='rekap';loadRekapTakzir()"><i class="ri-bar-chart-box-line"></i> Rekap Takzir</button>
    </div>
    <div class="stat-grid fade-up" style="grid-template-columns:repeat(3,1fr);margin-bottom:1rem">
      <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${totalPel}</div><div class="sl">Total Pelanggaran</div></div><div class="si"><i class="ri-error-warning-line"></i></div></div>
      <div class="stat-card c-green"><div class="stat-info"><div class="sn">${totalSudah}/${totalPel}</div><div class="sl">Sudah Ditakzir</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
      <div class="stat-card c-red"><div class="stat-info"><div class="sn">${fmtRp(totalDenda)}</div><div class="sl">Akumulasi Denda</div></div><div class="si"><i class="ri-money-dollar-circle-line"></i></div></div>
    </div>
    <div class="card fade-up"><h3><i class="ri-bar-chart-box-line"></i> Rekap Per Santri</h3><div class="table-wrap"><table>
      <tr><th>Santri</th><th style="text-align:center">Total</th><th style="text-align:center">Poin</th><th style="text-align:center">Sudah</th><th style="text-align:center">Belum</th><th style="text-align:right">Total Denda</th><th>Detail</th></tr>
      ${rekap.length?rekap.map(r=>{
        return `<tr>
        <td><strong>${r.santri_nama||'-'}</strong></td>
        <td style="text-align:center"><span class="badge-a">${r.total}</span></td>
        <td style="text-align:center"><span class="badge-a">${r.total_poin||0}</span></td>
        <td style="text-align:center"><span style="display:inline-block;padding:.15rem .45rem;border-radius:8px;font-size:.7rem;font-weight:600;background:#dcfce7;color:#16a34a">${r.sudah_takzir}</span></td>
        <td style="text-align:center"><span style="display:inline-block;padding:.15rem .45rem;border-radius:8px;font-size:.7rem;font-weight:600;background:${r.belum_takzir?'#fef2f2;color:#dc2626':'#f1f5f9;color:var(--t3)'}">${r.belum_takzir}</span></td>
        <td style="text-align:right;font-weight:600;color:${r.total_denda?'#dc2626':'var(--t3)'}">${r.total_denda?fmtRp(r.total_denda):'-'}</td>
        <td><button class="btn btn-outline btn-sm" onclick="showDetailTakzir(${r.santri_id},'${(r.santri_nama||'').replace(/'/g,"\\'")}')"><i class="ri-eye-line"></i></button></td>
      </tr>`}).join('')
      :'<tr><td colspan="7" style="text-align:center;color:var(--t3);padding:2rem">Belum ada data pelanggaran</td></tr>'}
    </table></div></div>`;
}

async function showDetailTakzir(santriId,santriNama){
  const data=await api('/api/pelanggaran/rekap?santri_id='+santriId);
  const detail=data.detail||[];
  $('modal').innerHTML=`<h3><i class="ri-user-line"></i> ${santriNama}</h3>
    <div style="display:flex;gap:.6rem;margin-bottom:1rem;flex-wrap:wrap">
      <div style="background:var(--greenbg);padding:.5rem .8rem;border-radius:10px;font-size:.78rem"><strong>${detail.length}</strong> pelanggaran</div>
      <div style="background:#fef2f2;padding:.5rem .8rem;border-radius:10px;font-size:.78rem;color:#dc2626"><strong>Denda: ${fmtRp(data.total_denda||0)}</strong></div>
    </div>
    <div class="table-wrap" style="max-height:400px;overflow-y:auto"><table>
      <tr><th>Tanggal</th><th>Jenis</th><th style="text-align:center">Poin</th><th>Status</th><th>Takziran</th><th style="text-align:right">Denda</th></tr>
      ${detail.map(d=>{
        const badge=d.status_takzir==='sudah'
          ?'<span style="display:inline-block;padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:600;background:#dcfce7;color:#16a34a">Sudah</span>'
          :'<span style="display:inline-block;padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:600;background:#fef2f2;color:#dc2626">Belum</span>';
        return `<tr>
          <td style="font-size:.78rem">${d.tanggal}</td>
          <td><strong>${d.jenis}</strong>${d.deskripsi?'<br><span style="font-size:.72rem;color:var(--t3)">'+d.deskripsi+'</span>':''}</td>
          <td style="text-align:center"><span class="badge-a">${d.poin||0}</span></td>
          <td>${badge}</td>
          <td style="font-size:.78rem">${d.jenis_takzir||'-'}${d.takzir_at?'<br><span style="font-size:.68rem;color:var(--t3)">'+d.takzir_at+'</span>':''}</td>
          <td style="text-align:right;font-weight:600;color:${d.denda?'#dc2626':'var(--t3)'}">${d.denda?fmtRp(d.denda):'-'}</td>
        </tr>`}).join('')}
    </table></div>
    <div style="margin-top:1rem"><button class="btn btn-outline" onclick="hideModal()">Tutup</button></div>`;
  showModal();
}

// -- SUPER ADMIN --
const FEATURE_GROUPS=[
  {section:'Kegiatan & Absensi', items:[
    {key:'absensi_harian',label:'Absensi Kegiatan',icon:'ri-checkbox-circle-line'},
    {key:'absen_malam',label:'Absen Kamar / Asrama',icon:'ri-moon-line'},
    {key:'jadwal',label:'Jadwal Kegiatan',icon:'ri-calendar-schedule-line'},
  ]},
  {section:'Sekolah & Diniyyah', items:[
    {key:'kelas_sekolah',label:'Kelas & Absen Sekolah',icon:'ri-school-line'},
    {key:'absen_sekolah',label:'Absensi Sekolah',icon:'ri-checkbox-circle-line'},
    {key:'madrasah_diniyyah',label:'Madrasah Diniyyah',icon:'ri-book-2-line'},
  ]},
  {section:'Kedisiplinan', items:[
    {key:'kedisiplinan',label:'Catatan, Pelanggaran & Perizinan',icon:'ri-shield-check-line'},
  ]},
  {section:'Keuangan', items:[
    {key:'keuangan',label:'Pembayaran SPP',icon:'ri-money-dollar-circle-line'},
    {key:'catatan_bendahara',label:'Catatan Bendahara',icon:'ri-book-3-line'},
  ]},
  {section:'Lainnya', items:[
    {key:'raport',label:'Raport Santri',icon:'ri-file-chart-line'},
    {key:'psb',label:'PSB Online',icon:'ri-user-add-line'},
  ]},
  {section:'Fitur Extra', items:[
    {key:'e_paket',label:'E-Paket (Logistik)',icon:'ri-box-3-line'},
    {key:'tahfidz',label:"Tahfidz Qur'an",icon:'ri-book-read-line'},
    {key:'sangu',label:'Sangu Santri (Kantin)',icon:'ri-wallet-3-line'},
  ]},
];
const FEATURE_LIST=FEATURE_GROUPS.flatMap(function(g){return g.items;});
function parseFeatures(s){if(!s)return[];try{return JSON.parse(s);}catch(e){return[];}}
function featureGrid(sel,pfx){
  var allOn=!sel.length;
  var html='<div class="fg"><label>Fitur Aktif</label><div style="display:flex;gap:.4rem;margin-bottom:.5rem"><button type="button" class="btn btn-outline btn-sm" onclick="document.querySelectorAll(\'input[name='+pfx+']\').forEach(function(c){c.checked=true;c.closest(\'label\').style.background=\'var(--greenbg)\';c.closest(\'label\').style.borderColor=\'var(--green)\'})" ><i class="ri-checkbox-circle-line"></i> Semua</button><button type="button" class="btn btn-outline btn-sm" onclick="document.querySelectorAll(\'input[name='+pfx+']\').forEach(function(c){c.checked=false;c.closest(\'label\').style.background=\'#f8fafc\';c.closest(\'label\').style.borderColor=\'var(--border)\'})" ><i class="ri-close-circle-line"></i> Hapus</button></div>';
  FEATURE_GROUPS.forEach(function(g){
    html+='<div style="font-size:.72rem;font-weight:700;color:var(--t3);text-transform:uppercase;letter-spacing:.05em;margin:.6rem 0 .3rem;padding-left:.1rem">'+g.section+'</div>';
    html+='<div style="display:grid;grid-template-columns:1fr 1fr;gap:.4rem;margin-bottom:.3rem">';
    g.items.forEach(function(f){
      var on=allOn||sel.includes(f.key);
      html+='<label style="display:flex;align-items:center;gap:.4rem;padding:.45rem .6rem;background:'+(on?'var(--greenbg)':'#f8fafc')+';border:1.5px solid '+(on?'var(--green)':'var(--border)')+';border-radius:10px;cursor:pointer;font-size:.78rem;font-weight:500;transition:.2s" onclick="setTimeout(function(){var c=this.querySelector(\'input\');this.style.background=c.checked?\'var(--greenbg)\':\'#f8fafc\';this.style.borderColor=c.checked?\'var(--green)\':\'var(--border)\'}.bind(this),10)"><input type="checkbox" name="'+pfx+'" value="'+f.key+'" '+(on?'checked':'')+' style="accent-color:var(--green);width:16px;height:16px"><i class="'+f.icon+'" style="font-size:.9rem;color:var(--p)"></i>'+f.label+'</label>';
    });
    html+='</div>';
  });
  html+='</div>';
  return html;
}
function getFeats(pfx){return[...document.querySelectorAll('input[name='+pfx+']:checked')].map(function(c){return c.value;});}

let _tenantList=[];
async function loadTenants(){
  _tenantList=await api('/api/super/tenants');
  const t=_tenantList;
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-building-2-line"></i> Kelola Tenant</h2>
    <div style="display:flex;gap:.5rem;flex-wrap:wrap">
      <button class="btn btn-outline btn-sm" onclick="showChangePassword()"><i class="ri-lock-line"></i> Ganti Password</button>
      <button class="btn btn-outline btn-sm" onclick="showBroadcast()"><i class="ri-megaphone-line"></i> Broadcast</button>
      <button class="btn btn-primary btn-sm" onclick="showAddTenant()"><i class="ri-add-line"></i> Tambah</button>
    </div></div>
    <div class="card au"><div class="table-wrap"><table>
      <tr><th>ID</th><th>Nama</th><th>Subdomain</th><th>Paket SaaS</th><th>Sangu?</th><th>Status</th><th>Masa Aktif</th><th>Fitur</th><th>Aksi</th></tr>
      ${t.map((x,_tidx)=>{
        const isExpired=x.masa_aktif&&new Date(x.masa_aktif)<new Date();
        const statusBadge=x.status==='active'&&!isExpired?'<span class="badge-h">Active</span>':isExpired?'<span class="badge-a">Expired</span>':'<span class="badge-i">Suspended</span>';
        const feats=parseFeatures(x.features);
        const featBadge=feats.length?`<span style="font-size:.68rem;color:var(--p);font-weight:600">${feats.length}/${FEATURE_LIST.length}</span>`:'<span style="font-size:.68rem;color:var(--green);font-weight:600">Semua</span>';
        const pktBadge=x.paket_saas==='premium'?'<span class="badge-h">Premium</span>':x.paket_saas==='standart'?'<span class="badge-p">Standart</span>':'<span class="badge-i">Trial</span>';
        const sngBadge=x.fitur_sangu?'<span class="badge-h">Aktif</span>':'<span class="badge-a">Nonaktif</span>';
        return `<tr><td>${x.id}</td><td>${x.nama}</td><td>${x.subdomain}</td>
        <td>${pktBadge}<br><small>${x.kuota_santri} santri</small></td>
        <td>${sngBadge}</td>
        <td>${statusBadge}</td>
        <td style="font-size:.78rem">${x.masa_aktif||'-'}</td>
        <td>${featBadge}</td>
        <td style="white-space:nowrap">
          <button class="btn btn-outline btn-sm" onclick="showSaaSSettingsTenant(${_tidx})" title="Atur SaaS"><i class="ri-vip-crown-line"></i></button>
          <button class="btn btn-outline btn-sm" onclick="showEditFeaturesIdx(${_tidx})" title="Fitur Modul"><i class="ri-settings-3-line"></i></button>
          <button class="btn btn-primary btn-sm" onclick="showExtendTenant(${x.id},'${x.nama.replace(/'/g,"\\'")}','${x.masa_aktif||''}')"><i class="ri-time-line"></i></button>
          <button class="btn btn-outline btn-sm" onclick="toggleTenant(${x.id})"><i class="ri-${x.status==='active'?'pause':'play'}-circle-line"></i></button>
          <button class="btn btn-danger btn-sm" onclick="delTenant(${x.id})"><i class="ri-delete-bin-line"></i></button>
        </td></tr>`}).join('')}
    </table></div></div>`;
}

function showSaaSSettingsTenant(idx){
  const x = _tenantList[idx];
  $('modal').innerHTML = `<h3><i class="ri-vip-crown-line"></i> Atur SaaS: ${x.nama}</h3>
    <div class="fg"><label>Paket SaaS</label><select id="stPaket">
      <option value="trial" ${x.paket_saas==='trial'?'selected':''}>Trial</option>
      <option value="standart" ${x.paket_saas==='standart'?'selected':''}>Standart</option>
      <option value="premium" ${x.paket_saas==='premium'?'selected':''}>Premium</option>
    </select></div>
    <div class="fg"><label>Kuota Santri</label><input type="number" id="stKuota" value="${x.kuota_santri}"></div>
    <div class="fg">
      <label><input type="checkbox" id="stSangu" ${x.fitur_sangu?'checked':''}> Aktifkan Fitur Sangu Santri</label>
    </div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="saveSaaSSettingsTenant(${x.id})"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
  showModal();
}

async function saveSaaSSettingsTenant(id) {
  const body = {
    paket_saas: $('stPaket').value,
    kuota_santri: parseInt($('stKuota').value)||100,
    fitur_sangu: $('stSangu').checked
  };
  await api('/api/saas/tenants/'+id+'/fitur', {method:'PUT', body:JSON.stringify(body)});
  hideModal(); toast('Berhasil update SaaS tenant'); loadTenants();
}
function showAddTenant(){
  $('modal').innerHTML=`<h3><i class="ri-building-2-line"></i> Tambah Tenant</h3>
    <div class="fg"><label>Nama Pesantren</label><input id="tNama" placeholder="Pesantren Al-Falah"></div>
    <div class="fg"><label>Subdomain</label><input id="tSub" placeholder="al-falah"></div>
    <div class="fg"><label>Alamat</label><input id="tAlm" placeholder="Jl. Contoh No. 1"></div>
    ${featureGrid([],'tf')}
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="saveTenant()"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveTenant(){
  const features=getFeats('tf');
  const b={nama:$('tNama').value,subdomain:$('tSub').value,alamat:$('tAlm')?.value||'',features};
  if(!b.nama||!b.subdomain)return toast('Nama & subdomain wajib');
  const r=await api('/api/super/tenants',{method:'POST',body:JSON.stringify(b)});
  hideModal();toast(r.message||'Ditambahkan');loadTenants();
}
async function delTenant(id){if(!confirm('Hapus tenant? SEMUA DATA terhapus!'))return;await api('/api/super/tenants/'+id,{method:'DELETE'});toast('Dihapus');loadTenants();}

function showEditFeaturesIdx(idx){const x=_tenantList[idx];if(!x)return;showEditFeatures(x.id,x.nama,x.features||'');}
function showEditFeatures(id,nama,featStr){
  const feats=parseFeatures(featStr);
  $('modal').innerHTML=`<h3><i class="ri-settings-3-line"></i> Fitur: ${nama}</h3>
    <p style="font-size:.82rem;color:var(--t2);margin-bottom:1rem">Pilih fitur yang diaktifkan untuk tenant ini.</p>
    ${featureGrid(feats,'ef')}
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="saveEditFeatures(${id})"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveEditFeatures(id){
  const features=getFeats('ef');
  await api('/api/super/tenants/'+id,{method:'PUT',body:JSON.stringify({features})});
  hideModal();toast('Fitur diperbarui');loadTenants();
}

// Extend tenant
function showExtendTenant(id,nama,currentExp){
  $('modal').innerHTML=`<h3><i class="ri-time-line"></i> Perpanjang: ${nama}</h3>
    <p style="font-size:.82rem;color:var(--t2);margin-bottom:1rem">Expired saat ini: <strong>${currentExp||'Tidak diset'}</strong></p>
    <div class="fg"><label>Tambah Durasi</label>
      <select id="extMonths">
        <option value="1">1 Bulan</option><option value="3">3 Bulan</option>
        <option value="6" selected>6 Bulan</option><option value="12">12 Bulan</option>
      </select></div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="doExtendTenant(${id})"><i class="ri-time-line"></i> Perpanjang</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function doExtendTenant(id){
  const months=parseInt($('extMonths').value);
  const r=await api('/api/super/tenants/'+id+'/extend',{method:'PUT',body:JSON.stringify({months})});
  hideModal();toast(r.message||'Diperpanjang');loadTenants();
}

// Toggle active/suspend
async function toggleTenant(id){
  const r=await api('/api/super/tenants/'+id+'/toggle',{method:'PUT'});
  toast(r.message||'Status diubah');loadTenants();
}

// Change password
function showChangePassword(){
  $('modal').innerHTML=`<h3><i class="ri-lock-line"></i> Ganti Password</h3>
    <div class="fg"><label>Password Lama</label><input type="password" id="cpOld" placeholder="Password lama"></div>
    <div class="fg"><label>Password Baru</label><input type="password" id="cpNew" placeholder="Min 6 karakter"></div>
    <div class="fg"><label>Konfirmasi</label><input type="password" id="cpConf" placeholder="Ulangi password baru"></div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="doChangePassword()"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function doChangePassword(){
  const old_password=$('cpOld').value,new_password=$('cpNew').value,conf=$('cpConf').value;
  if(!old_password||!new_password)return toast('Isi semua field');
  if(new_password!==conf)return toast('Konfirmasi tidak cocok');
  if(new_password.length<6)return toast('Min 6 karakter');
  try{const r=await api('/api/super/change-password',{method:'PUT',body:JSON.stringify({old_password,new_password})});
    hideModal();toast(r.message||'Password diubah');}catch(e){toast(e.message);}
}

// Broadcast
function showBroadcast(){
  $('modal').innerHTML=`<h3><i class="ri-megaphone-line"></i> Broadcast Pengumuman</h3>
    <p style="font-size:.82rem;color:var(--t2);margin-bottom:1rem">Pengumuman akan dikirim ke <strong>semua tenant</strong> aktif.</p>
    <div class="fg"><label>Judul</label><input id="bcJudul" placeholder="Judul pengumuman"></div>
    <div class="fg"><label>Isi Pengumuman</label><textarea id="bcIsi" rows="4" placeholder="Tulis isi pengumuman..."></textarea></div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="doBroadcast()"><i class="ri-send-plane-line"></i> Kirim ke Semua</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function doBroadcast(){
  const judul=$('bcJudul').value.trim(),isi=$('bcIsi').value.trim();
  if(!judul||!isi)return toast('Judul & isi wajib');
  if(!confirm('Kirim pengumuman ke SEMUA tenant?'))return;
  const r=await api('/api/super/broadcast',{method:'POST',body:JSON.stringify({judul,isi})});
  hideModal();toast(r.message||'Terkirim');
}

async function loadSuperStats() {
  const m = $('main');
  m.innerHTML = '<div class="card"><p>Memuat statistik...</p></div>';
  try {
    const [s, srv] = await Promise.all([
      api('/api/super/stats'),
      api('/api/super/server-stats').catch(() => null)
    ]);
    
    let html = `<div class="page-header au"><h2><i class="ri-bar-chart-box-line"></i> Statistik & Server</h2></div>`;
    
    if(srv) {
      const gbRamUsed = (srv.ram_used / 1073741824).toFixed(1);
      const gbRamTotal = (srv.ram_total / 1073741824).toFixed(1);
      const gbDiskUsed = (srv.disk_used / 1073741824).toFixed(1);
      const gbDiskTotal = (srv.disk_total / 1073741824).toFixed(1);
      
      html += `
      <div class="qa-stat-grid au" style="margin-bottom:1.5rem">
        <div class="card" style="padding:1rem">
          <div style="font-size:.75rem;color:var(--t3);margin-bottom:.3rem"><i class="ri-cpu-line"></i> CPU Usage</div>
          <div style="font-size:1.5rem;font-weight:800;color:var(--p)">${srv.cpu_usage}%</div>
          <div style="width:100%;background:var(--border);height:6px;border-radius:3px;margin-top:.5rem;overflow:hidden">
            <div style="height:100%;background:var(--p);width:${srv.cpu_usage}%"></div>
          </div>
        </div>
        <div class="card" style="padding:1rem">
          <div style="font-size:.75rem;color:var(--t3);margin-bottom:.3rem"><i class="ri-memory-line"></i> RAM Usage</div>
          <div style="font-size:1.5rem;font-weight:800;color:var(--amber)">${srv.ram_percent}%</div>
          <div style="font-size:.7rem;color:var(--t3);margin-top:.2rem">${gbRamUsed} GB / ${gbRamTotal} GB</div>
          <div style="width:100%;background:var(--border);height:6px;border-radius:3px;margin-top:.5rem;overflow:hidden">
            <div style="height:100%;background:var(--amber);width:${srv.ram_percent}%"></div>
          </div>
        </div>
        <div class="card" style="padding:1rem">
          <div style="font-size:.75rem;color:var(--t3);margin-bottom:.3rem"><i class="ri-hard-drive-2-line"></i> Storage (/)</div>
          <div style="font-size:1.5rem;font-weight:800;color:var(--green)">${srv.disk_percent}%</div>
          <div style="font-size:.7rem;color:var(--t3);margin-top:.2rem">${gbDiskUsed} GB / ${gbDiskTotal} GB</div>
          <div style="width:100%;background:var(--border);height:6px;border-radius:3px;margin-top:.5rem;overflow:hidden">
            <div style="height:100%;background:var(--green);width:${srv.disk_percent}%"></div>
          </div>
        </div>
      </div>`;
    }
    
    html += `<div class="card au"><div class="table-wrap"><table><tr><th>Tenant</th><th>Santri</th><th>Users</th><th>Expired</th><th>Status</th></tr>`;
    html += s.map(x => {
      const isExp = x.expired_at && new Date(x.expired_at) < new Date();
      return `<tr><td>${x.nama}</td><td>${x.total_santri}</td><td>${x.total_users}</td><td style="font-size:.78rem">${x.expired_at||'-'}</td><td>${x.status==='active'&&!isExp?'<span class="badge-h">Active</span>':isExp?'<span class="badge-a">Expired</span>':'<span class="badge-i">Suspended</span>'}</td></tr>`;
    }).join('');
    html += `</table></div></div>`;
    m.innerHTML = html;
  } catch(e) {
    m.innerHTML = '<div class="card"><p style="color:red">Error loading stats</p></div>';
  }
}

// -- USERS & SETTINGS --
let _usersData=[];let _usersFilter='semua';
async function loadUsers(){
  _usersData=await api('/api/users');
  renderUsers();
}
function renderUsers(){
  const list=_usersData;
  const search=($('userSearch')?.value||'').toLowerCase();
  const roleOrder=['admin','ustadz','bendahara','keamanan','wali','merchant_admin','kasir'];
  const roleLabels={admin:'Admin',ustadz:'Ustadz',bendahara:'Bendahara',keamanan:'Keamanan',wali:'Wali Santri',merchant_admin:'Admin Kantin',kasir:'Kasir Kantin'};
  const roleIcons={admin:'ri-shield-star-line',ustadz:'ri-user-heart-line',bendahara:'ri-money-dollar-circle-line',keamanan:'ri-shield-check-line',wali:'ri-parent-line',merchant_admin:'ri-store-2-line',kasir:'ri-shopping-cart-2-line'};
  const roleColors={admin:'#3b82f6',ustadz:'#10b981',bendahara:'#16a34a',keamanan:'#f59e0b',wali:'#8b5cf6',merchant_admin:'#0ea5e9',kasir:'#ec4899'};
  const filtered=list.filter(u=>{
    if(_usersFilter!=='semua'&&u.role!==_usersFilter)return false;
    if(search&&!u.nama.toLowerCase().includes(search)&&!u.username.toLowerCase().includes(search))return false;
    return true;
  });
  const counts={};roleOrder.forEach(r=>counts[r]=list.filter(u=>u.role===r).length);

  const tabs=`<div style="display:flex;gap:.3rem;overflow-x:auto;padding:.2rem 0;margin-bottom:.8rem;-webkit-overflow-scrolling:touch">
    <button onclick="_usersFilter='semua';renderUsers()" class="btn btn-sm ${_usersFilter==='semua'?'btn-primary':'btn-outline'}" style="white-space:nowrap;flex-shrink:0">Semua (${list.length})</button>
    ${roleOrder.map(r=>`<button onclick="_usersFilter='${r}';renderUsers()" class="btn btn-sm ${_usersFilter===r?'btn-primary':'btn-outline'}" style="white-space:nowrap;flex-shrink:0">${roleLabels[r]} (${counts[r]||0})</button>`).join('')}
  </div>`;

  // Group filtered users by role
  const grouped={};roleOrder.forEach(r=>grouped[r]=[]);
  filtered.forEach(u=>{if(grouped[u.role])grouped[u.role].push(u);});

  let sections='';
  const rolesToShow=_usersFilter==='semua'?roleOrder:[_usersFilter];
  for(const r of rolesToShow){
    const users=grouped[r]||[];
    if(!users.length)continue;
    const isWali=r==='wali';
    sections+=`<div class="card au" style="margin-bottom:.6rem">
      <div style="display:flex;align-items:center;gap:.6rem;margin-bottom:.7rem">
        <div style="width:36px;height:36px;border-radius:10px;background:${roleColors[r]};display:flex;align-items:center;justify-content:center"><i class="${roleIcons[r]}" style="color:#fff;font-size:1.1rem"></i></div>
        <div style="flex:1"><div style="font-weight:700;font-size:.9rem">${roleLabels[r]}</div><div style="font-size:.68rem;color:var(--t3)">${users.length} pengguna</div></div>
        ${isWali?`<div style="display:flex;gap:.3rem">
          <button class="btn btn-sm" style="background:#10b981;color:#fff;font-size:.65rem" onclick="bulkToggleWali(1)"><i class="ri-checkbox-circle-line"></i> Aktifkan</button>
          <button class="btn btn-sm btn-danger" style="font-size:.65rem" onclick="bulkToggleWali(0)"><i class="ri-close-circle-line"></i> Nonaktifkan</button>
        </div>`:''}
      </div>
      ${isWali?`<div style="margin-bottom:.5rem;display:flex;align-items:center;gap:.5rem">
        <label style="font-size:.75rem;color:var(--t3);display:flex;align-items:center;gap:.3rem;cursor:pointer">
          <input type="checkbox" id="selectAllWali" onchange="toggleSelectAllWali(this.checked)" style="width:16px;height:16px;accent-color:var(--p)"> Pilih Semua
        </label>
        <span id="waliSelectedCount" style="font-size:.7rem;color:var(--p);font-weight:600"></span>
      </div>`:''}
      <div style="display:flex;flex-direction:column;gap:.4rem">
        ${users.map(u=>{
          const active=u.is_active!==0;
          return `<div style="display:flex;align-items:center;gap:.6rem;padding:.55rem .7rem;background:${active?'#f8fafc':'#fef2f2'};border-radius:12px;border:1px solid ${active?'var(--border)':'rgba(239,68,68,.15)'}">
            ${isWali?`<input type="checkbox" class="wali-cb" value="${u.id}" onchange="updateWaliCount()" style="width:16px;height:16px;accent-color:var(--p);flex-shrink:0">`:''}
            <div style="flex:1;min-width:0">
              <div style="font-weight:700;font-size:.82rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${u.nama}</div>
              <div style="font-size:.68rem;color:var(--t3)">@${u.username}</div>
            </div>
            ${isWali?`<span style="padding:.15rem .45rem;border-radius:6px;font-size:.6rem;font-weight:700;${active?'background:var(--greenbg);color:var(--green)':'background:var(--redbg);color:var(--red)'}">${active?'Aktif':'Nonaktif'}</span>`:''}
            <div style="display:flex;gap:.2rem;flex-shrink:0">
              <button class="btn btn-outline btn-sm" onclick="showResetPw(${u.id},'${u.username}')" title="Reset Password"><i class="ri-key-line"></i></button>
              <button class="btn btn-danger btn-sm" onclick="delUser(${u.id})" title="Hapus"><i class="ri-delete-bin-line"></i></button>
            </div>
          </div>`;
        }).join('')}
      </div>
    </div>`;
  }

  $('main').innerHTML=`<div class="page-header au">
    <h2><i class="ri-group-line"></i> Kelola Pengguna</h2>
    <button class="btn btn-primary btn-sm" onclick="showAddUser()"><i class="ri-user-add-line"></i> Tambah</button>
  </div>
  <div class="fg" style="margin-bottom:.5rem;position:relative">
    <i class="ri-search-line" style="position:absolute;left:.8rem;top:50%;transform:translateY(-50%);color:var(--t3)"></i>
    <input id="userSearch" class="search-box" placeholder="Cari nama atau username..." oninput="renderUsers()" style="padding-left:2.3rem">
  </div>
  ${tabs}
  ${sections||'<div class="card" style="text-align:center;padding:2rem;color:var(--t3)"><i class="ri-user-unfollow-line" style="font-size:2rem;display:block;margin-bottom:.5rem"></i>Tidak ada pengguna ditemukan</div>'}
  <div class="card au" style="background:#f8fafc;border:1px dashed var(--border)">
    <div style="display:flex;align-items:center;gap:.6rem">
      <i class="ri-information-line" style="font-size:1.2rem;color:var(--p)"></i>
      <div style="font-size:.78rem;color:var(--t3)"><strong>Admin</strong> = akses penuh, <strong>Ustadz</strong> = absensi & catatan, <strong>Bendahara</strong> = keuangan, <strong>Keamanan</strong> = perizinan & paket, <strong>Wali</strong> = lihat raport anak</div>
    </div>
  </div>`;
}
function toggleSelectAllWali(checked){
  document.querySelectorAll('.wali-cb').forEach(cb=>cb.checked=checked);
  updateWaliCount();
}
function updateWaliCount(){
  const checked=document.querySelectorAll('.wali-cb:checked').length;
  const el=$('waliSelectedCount');
  if(el)el.textContent=checked?checked+' dipilih':'';
}
async function bulkToggleWali(isActive){
  const ids=[...document.querySelectorAll('.wali-cb:checked')].map(cb=>parseInt(cb.value));
  if(!ids.length)return toast('Pilih minimal 1 akun wali');
  const action=isActive?'mengaktifkan':'menonaktifkan';
  if(!confirm(`Yakin ${action} ${ids.length} akun wali?`))return;
  try{
    await api('/api/users/bulk-toggle',{method:'PUT',body:JSON.stringify({ids,is_active:isActive})});
    toast(`[OK] ${ids.length} akun wali ${isActive?'diaktifkan':'dinonaktifkan'}`);
    loadUsers();
  }catch(e){toast('[Error] '+e.message);}
}
async function showAddUser(){
  let merchants = [];
  try { merchants = await api('/api/sangu/merchants'); } catch(e){}
  
  $('modal').innerHTML=`<h3><i class="ri-user-add-line"></i> Tambah Pengguna</h3>
    <div class="fg"><label>Username</label><input id="uU" placeholder="username_baru"></div>
    <div class="fg"><label>Password</label><input id="uP" type="password" placeholder="Minimal 6 karakter"></div>
    <div class="fg"><label>Nama Lengkap</label><input id="uN" placeholder="Nama lengkap pengguna"></div>
    <div class="fg"><label>Role</label>
      <select id="uR" onchange="document.getElementById('uMerchantArea').style.display=(this.value==='merchant_admin'||this.value==='kasir')?'block':'none'" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%;font-size:.85rem">
        <option value="ustadz"> Ustadz</option>
        <option value="bendahara"> Bendahara</option>
        <option value="keamanan"> Keamanan</option>
        <option value="wali"> Wali Santri</option>
        <option value="merchant_admin"> Admin Merchant</option>
        <option value="kasir"> Kasir Kantin</option>
        <option value="admin"> Admin</option>
      </select>
    </div>
    <div class="fg" id="uMerchantArea" style="display:none">
      <label>Pilih Merchant</label>
      <select id="uMerchantId" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%;font-size:.85rem">
        <option value="">- Pilih Merchant -</option>
        ${merchants.map(m=>`<option value="${m.id}">${m.nama}</option>`).join('')}
      </select>
    </div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="saveUser()"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
  showModal();
}
async function saveUser(){
  const b={username:$('uU').value,password:$('uP').value,nama:$('uN').value,role:$('uR').value};
  if($('uMerchantId') && $('uMerchantId').value) {
    b.merchant_id = parseInt($('uMerchantId').value);
  }
  if(b.role === 'merchant_admin' || b.role === 'kasir') {
    if(!b.merchant_id) return toast('Pilih merchant terlebih dahulu!');
  }
  if(!b.username||!b.password||!b.nama)return toast('Semua field wajib diisi');
  if(b.password.length<6)return toast('Password minimal 6 karakter');
  try{
    await api('/api/users',{method:'POST',body:JSON.stringify(b)});
    hideModal();toast('[OK]  Pengguna ditambahkan');loadUsers();
  }catch(e){toast('[Error]  Gagal: '+e.message);}
}
async function showEditRole(id,username,currentRole){
  let merchants = [];
  try { merchants = await api('/api/sangu/merchants'); } catch(e){}
  
  $('modal').innerHTML=`<h3><i class="ri-shield-user-line"></i> Ubah Role</h3>
    <p style="margin-bottom:1rem;color:var(--t3)">Pengguna: <strong>${username}</strong></p>
    <div class="fg"><label>Role Baru</label>
      <select id="newRole" onchange="document.getElementById('editMerchantArea').style.display=(this.value==='merchant_admin'||this.value==='kasir')?'block':'none'" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%;font-size:.85rem">
        <option value="ustadz" ${currentRole==='ustadz'?'selected':''}> Ustadz</option>
        <option value="bendahara" ${currentRole==='bendahara'?'selected':''}> Bendahara</option>
        <option value="keamanan" ${currentRole==='keamanan'?'selected':''}> Keamanan</option>
        <option value="wali" ${currentRole==='wali'?'selected':''}> Wali Santri</option>
        <option value="merchant_admin" ${currentRole==='merchant_admin'?'selected':''}> Admin Merchant</option>
        <option value="kasir" ${currentRole==='kasir'?'selected':''}> Kasir Kantin</option>
        <option value="admin" ${currentRole==='admin'?'selected':''}> Admin</option>
      </select>
    </div>
    <div class="fg" id="editMerchantArea" style="display:${(currentRole==='merchant_admin'||currentRole==='kasir')?'block':'none'}">
      <label>Pilih Merchant (Jika Role = Admin Kantin / Kasir)</label>
      <select id="newMerchantId" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%;font-size:.85rem">
        <option value="">- Pilih Merchant -</option>
        ${merchants.map(m=>`<option value="${m.id}">${m.nama}</option>`).join('')}
      </select>
    </div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="updateRole(${id})"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
  showModal();
}
async function updateRole(id){
  const role=$('newRole').value;
  const body = {role};
  if(role === 'merchant_admin' || role === 'kasir') {
    const mid = parseInt($('newMerchantId').value);
    if(!mid) return toast('Pilih merchant terlebih dahulu!');
    body.merchant_id = mid;
  }
  try{
    await api('/api/users/'+id,{method:'PUT',body:JSON.stringify(body)});
    hideModal();toast('[OK]  Role diperbarui');loadUsers();
  }catch(e){toast('[Error]  Gagal: '+e.message);}
}
function showResetPw(id,username){
  $('modal').innerHTML=`<h3><i class="ri-key-line"></i> Reset Password</h3>
    <p style="margin-bottom:1rem;color:var(--t3)">Pengguna: <strong>${username}</strong></p>
    <div class="fg"><label>Password Baru</label><input id="newPw" type="password" placeholder="Minimal 6 karakter"></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="resetPw(${id})"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
  showModal();
}
async function resetPw(id){
  const pw=$('newPw').value;
  if(!pw||pw.length<6)return toast('Password minimal 6 karakter');
  try{
    await api('/api/users/'+id,{method:'PUT',body:JSON.stringify({password:pw})});
    hideModal();toast('[OK]  Password direset');
  }catch(e){toast('[Error]  Gagal: '+e.message);}
}
async function delUser(id){if(!confirm('Yakin hapus pengguna ini?'))return;await api('/api/users/'+id,{method:'DELETE'});toast('[OK]  Dihapus');loadUsers();}

async function loadSettings(){const s=await api('/api/settings');
  const pgProviders=[{v:'',l:'Tidak Aktif'},{v:'midtrans',l:'Midtrans'}];
  const pgProviderOpts=pgProviders.map(p=>`<option value="${p.v}" ${(s.pg_provider||'')===p.v?'selected':''}>${p.l}</option>`).join('');
  $('main').innerHTML=`<div class="page-header fade-up"><h2> Pengaturan</h2></div><div class="card fade-up"><h3><i class="ri-building-line"></i> Informasi Lembaga</h3><div class="fg"><label>Nama Aplikasi</label><input id="stN" value="${s.app_name||''}"></div><div class="fg"><label>Alamat</label><input id="stA" value="${s.alamat_lembaga||''}"></div><div class="fg"><label>Kepala</label><input id="stK" value="${s.kepala_nama||''}"></div><div class="fg"><label>Kota</label><input id="stKo" value="${s.nama_kota||''}"></div></div><div class="card fade-up"><h3><i class="ri-image-line"></i> Logo Lembaga</h3>
<p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Logo ini akan digunakan pada kop surat Raport.</p>
<div class="fg">
  <input type="file" id="stLogoFile" accept="image/*" onchange="handleLogoUpload(this)">
  <input type="hidden" id="stLogoBase64" value="${s.logo_base64||''}">
  <div style="margin-top:10px">
    <img id="stLogoPreview" src="${s.logo_base64||''}" style="max-height:80px; ${s.logo_base64?'':'display:none'}">
  </div>
</div>
</div>
<div class="card fade-up"><h3><i class="ri-medal-line"></i> Komposisi Penilaian</h3>
<p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Atur komponen nilai dan bobot persentasenya. Format JSON: [{"nama":"Harian","bobot":30}, ...]</p>
<div class="fg"><label>Sekolah Formal</label><textarea id="stKompSekolah" rows="3" style="width:100%;font-family:monospace">${s.komponen_nilai_sekolah||''}</textarea></div>
<div class="fg"><label>Madrasah Diniyyah</label><textarea id="stKompDiniyyah" rows="3" style="width:100%;font-family:monospace">${s.komponen_nilai_diniyyah||''}</textarea></div>
</div>
<div class="card fade-up"><h3><i class="ri-bank-card-line"></i> Info Pembayaran (Tampil ke Wali)</h3><p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Data ini akan ditampilkan ke wali santri saat pembayaran belum lunas</p><div class="fg"><label>Nama Bank</label><input id="stRekBank" value="${s.rekening_bank||''}" placeholder="BRI / BSI / Mandiri / dll"></div><div class="fg"><label>Nomor Rekening</label><input id="stRekNomor" value="${s.rekening_nomor||''}" placeholder="1234567890"></div><div class="fg"><label>Atas Nama</label><input id="stRekNama" value="${s.rekening_atas_nama||''}" placeholder="Nama pemilik rekening"></div><div class="fg"><label>No. HP/WA Bendahara</label><input id="stBendTelp" value="${s.bendahara_telp||''}" placeholder="08xxxxxxxxxx"></div></div>
  <div class="card fade-up"><h3><i class="ri-secure-payment-line" style="color:var(--green)"></i> Payment Gateway</h3>
    <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Atur payment gateway agar wali santri bisa bayar tagihan online langsung dari dashboard</p>
    <div class="fg"><label>Provider</label><select id="stPgProvider" onchange="togglePGFields()" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%;font-size:.85rem">${pgProviderOpts}</select></div>
    <div id="pgFieldsArea" style="display:${s.pg_provider?'block':'none'}">
      <div style="display:flex;gap:.5rem;margin-bottom:.6rem">
        <label style="display:flex;align-items:center;gap:.4rem;padding:.4rem .7rem;border-radius:10px;cursor:pointer;font-size:.78rem;font-weight:600;${!s.pg_is_production?'background:var(--greenbg);border:1.5px solid var(--green)':'background:#f8fafc;border:1.5px solid var(--border)'}">
          <input type="radio" name="pgMode" value="0" ${!s.pg_is_production?'checked':''} style="accent-color:var(--green)"> Sandbox
        </label>
        <label style="display:flex;align-items:center;gap:.4rem;padding:.4rem .7rem;border-radius:10px;cursor:pointer;font-size:.78rem;font-weight:600;${s.pg_is_production?'background:rgba(59,130,246,.08);border:1.5px solid var(--blue)':'background:#f8fafc;border:1.5px solid var(--border)'}">
          <input type="radio" name="pgMode" value="1" ${s.pg_is_production?'checked':''} style="accent-color:var(--blue)"> Production
        </label>
      </div>
      <div class="fg"><label>Server Key</label><input id="stPgServerKey" type="password" value="${s.pg_server_key||''}" placeholder="SB-Mid-server-xxxxx atau Midtrans Server Key"></div>
      <div class="fg"><label>Client Key</label><input id="stPgClientKey" value="${s.pg_client_key||''}" placeholder="SB-Mid-client-xxxxx atau Midtrans Client Key"></div>
      <div style="border-top:1px solid var(--border);margin:.8rem 0;padding-top:.8rem">
        <div style="font-size:.82rem;font-weight:700;margin-bottom:.5rem"><i class="ri-price-tag-3-line"></i> Biaya Admin (ditanggung wali)</div>
        <div style="display:flex;gap:.5rem">
          <div class="fg" style="flex:1"><label>Fee Persen (%)</label><input type="number" id="stPgFeePercent" value="${s.pg_fee_percent||0}" min="0" max="100" step="0.1" placeholder="2.5"></div>
          <div class="fg" style="flex:1"><label>Fee Flat (Rp)</label><input type="number" id="stPgFeeFlat" value="${s.pg_fee_flat||0}" min="0" placeholder="0"></div>
        </div>
        <div style="font-size:.68rem;color:var(--t3);margin-top:.2rem"><i class="ri-information-line"></i> Total fee = (tagihan × persen/100) + flat. Fee ditambahkan ke nominal yang dibayar wali.</div>
      </div>
    </div>
  </div>
  <div class="card fade-up"><h3><i class="ri-bank-line" style="color:var(--green)"></i> Transfer Bank</h3>
    <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Wali bayar langsung ke rekening pesantren dengan kode unik. Lebih murah dari payment gateway.</p>
    <div class="fg"><label style="display:flex;align-items:center;gap:.5rem;cursor:pointer">
      <input type="checkbox" id="stTbEnabled" ${s.transfer_bank_enabled?'checked':''} style="accent-color:var(--green);width:18px;height:18px" onchange="$('tbFieldsArea').style.display=this.checked?'block':'none'">
      <span style="font-weight:700;font-size:.85rem">Aktifkan Pembayaran via Transfer Bank</span>
    </label></div>
    <div id="tbFieldsArea" style="display:${s.transfer_bank_enabled?'block':'none'}">
      <div style="border-top:1px solid var(--border);margin:.6rem 0;padding-top:.8rem">
        <div style="font-size:.82rem;font-weight:700;margin-bottom:.5rem"><i class="ri-price-tag-3-line"></i> Biaya & Batas Waktu</div>
        <div style="display:flex;gap:.5rem">
          <div class="fg" style="flex:1"><label>Fee per Transaksi (Rp)</label><input type="number" id="stTbFee" value="${s.transfer_bank_fee_flat||0}" min="0" placeholder="0 = gratis"></div>
          <div class="fg" style="flex:1"><label>Jam Kadaluarsa</label><input type="number" id="stTbExpiry" value="${s.transfer_bank_expiry_hours||24}" min="1" placeholder="24"></div>
        </div>
        <div style="font-size:.68rem;color:var(--t3);margin-top:.2rem"><i class="ri-information-line"></i> Fee untuk cover biaya operasional/Moota. Set 0 jika gratis. Contoh: Moota Rp 100rb/bulan ÷ 50 trx = Rp 2.000/trx.</div>
      </div>
      <div style="border-top:1px solid var(--border);margin:.6rem 0;padding-top:.8rem">
        <div class="fg"><label style="display:flex;align-items:center;gap:.5rem;cursor:pointer">
          <input type="checkbox" id="stMootaEnabled" ${s.moota_enabled?'checked':''} style="accent-color:var(--green);width:18px;height:18px" onchange="$('mootaFieldsArea').style.display=this.checked?'block':'none'">
          <span style="font-weight:700;font-size:.82rem"><i class="ri-flashlight-line" style="color:#f59e0b"></i> Moota Auto-Confirm</span>
        </label></div>
        <div id="mootaFieldsArea" style="display:${s.moota_enabled?'block':'none'}">
          <div class="fg"><label>API Token Moota</label><input type="password" id="stMootaToken" value="${s.moota_api_token||''}" placeholder="Dari dashboard moota.co"></div>
          <div class="fg"><label>Secret Token Webhook</label><input type="password" id="stMootaSecret" value="${s.moota_secret_token||''}" placeholder="Secret untuk validasi webhook"></div>
          <div style="background:rgba(59,130,246,.05);border:1px solid rgba(59,130,246,.15);border-radius:10px;padding:.5rem .6rem;margin-top:.4rem">
            <div style="font-size:.7rem;color:var(--t2)"><i class="ri-information-line" style="color:var(--blue)"></i> Daftar di <strong>moota.co</strong>, input iBanking pesantren, lalu set webhook URL: <code style="background:rgba(0,0,0,.06);padding:.1rem .3rem;border-radius:4px;font-size:.68rem">https://domain-anda/api/moota/webhook</code></div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div class="card fade-up"><h3><i class="ri-user-add-line"></i> Kuota PSB</h3><p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Batas jumlah pendaftar PSB (0 = tanpa batas)</p><div class="fg"><label>Kuota Pendaftaran</label><input type="number" id="stPsbKuota" value="${s.psb_kuota||0}" min="0" placeholder="0 = unlimited"></div></div><div class="fade-up" style="margin-top:.5rem"><button class="btn btn-primary" onclick="saveSettings()"><i class="ri-save-line"></i> Simpan Pengaturan</button></div><div style="height:80px"></div>`;
}
function handleLogoUpload(input){
  const file = input.files[0];
  if(file){
    const reader = new FileReader();
    reader.onload = e => {
      $('stLogoBase64').value = e.target.result;
      $('stLogoPreview').src = e.target.result;
      $('stLogoPreview').style.display = 'block';
    };
    reader.readAsDataURL(file);
  }
}
function togglePGFields(){const v=$('stPgProvider').value;$('pgFieldsArea').style.display=v?'block':'none';}
async function saveSettings(){
  const appName=$('stN').value,alamat=$('stA').value,kepala=$('stK').value,kota=$('stKo').value;
  const rekBank=$('stRekBank')?.value||'',rekNomor=$('stRekNomor')?.value||'',rekNama=$('stRekNama')?.value||'',bendTelp=$('stBendTelp')?.value||'';
  const pgProvider=$('stPgProvider')?.value||'';
  const pgServerKey=$('stPgServerKey')?.value||'';
  const pgClientKey=$('stPgClientKey')?.value||'';
  const pgMode=document.querySelector('input[name=pgMode]:checked');
  const pgIsProduction=pgMode?parseInt(pgMode.value):0;
  const pgFeePercent=parseFloat($('stPgFeePercent')?.value)||0;
  const pgFeeFlat=parseInt($('stPgFeeFlat')?.value)||0;
  const logoBase64=$('stLogoBase64')?.value||'';
  const kompSekolah=$('stKompSekolah')?.value||'';
  const kompDiniyyah=$('stKompDiniyyah')?.value||'';
  try{
    const body={app_name:appName,alamat_lembaga:alamat,kepala_nama:kepala,nama_kota:kota,rekening_bank:rekBank,rekening_nomor:rekNomor,rekening_atas_nama:rekNama,bendahara_telp:bendTelp,psb_kuota:parseInt(stPsbKuota?.value)||0,
      pg_provider:pgProvider,pg_server_key:pgServerKey,pg_client_key:pgClientKey,pg_is_production:pgIsProduction,pg_fee_percent:pgFeePercent,pg_fee_flat:pgFeeFlat,
      transfer_bank_enabled:$('stTbEnabled')?.checked?1:0,
      transfer_bank_fee_flat:parseInt($('stTbFee')?.value)||0,
      transfer_bank_expiry_hours:parseInt($('stTbExpiry')?.value)||24,
      moota_enabled:$('stMootaEnabled')?.checked?1:0,
      moota_api_token:$('stMootaToken')?.value||'',
      moota_secret_token:$('stMootaSecret')?.value||'',
      logo_base64:logoBase64,
      komponen_nilai_sekolah:kompSekolah,
      komponen_nilai_diniyyah:kompDiniyyah
    };
    console.log('Saving settings:', body);
    const res=await api('/api/settings',{method:'PUT',body:JSON.stringify(body)});
    console.log('Save response:', res);
    if(res.message&&res.message.includes('Gagal')){toast('[Error]  '+res.message);alert('Error: '+res.message);return;}
    // Verifikasi: langsung fetch ulang dari server
    const verify=await api('/api/settings');
    console.log('Verify GET:', verify);
    if(verify.app_name!==appName){
      toast(' Nama tidak tersimpan! Cek server restart.');
      alert('Settings tidak tersimpan di database!\n\nYang dikirim: '+appName+'\nYang di DB: '+(verify.app_name||'kosong')+'\n\nPastikan server Node.js sudah di-restart setelah deploy.');
      return;
    }
    const brandName = appName || user.tenant_nama || 'E-Pesantren';
    $('sidebarName').textContent = brandName;
    $('hTitle').textContent = brandName;
    document.title=(verify.app_name||'Pesantren')+'  -  Dashboard';
    document.querySelector('.sb-sub').textContent=verify.alamat_lembaga||'Management System';
    toast('[OK]  Disimpan & Terverifikasi');
  }catch(e){console.error('Save error:',e);toast('[Error]  Gagal: '+e.message);alert('Error saving: '+e.message);}
}

// -- JADWAL KEGIATAN --
async function loadJadwalUmum(){
  const [jadwal,kelompok,users,kegiatan]=await Promise.all([api('/api/jadwal-umum'),api('/api/kelompok'),api('/api/users').catch(()=>[]),api('/api/kegiatan').catch(()=>[])]);
  const ustadzList=users.filter(u=>u.role==='ustadz');
  const hariList=['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu','Minggu'];
  const grouped={};hariList.forEach(h=>grouped[h]=[]);
  jadwal.forEach(j=>{
    if(j.hari==='Setiap Hari'){hariList.forEach(h=>grouped[h].push(j));}
    else if(grouped[j.hari])grouped[j.hari].push(j);
  });
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-calendar-schedule-line"></i> Jadwal Kegiatan</h2>
    ${['admin','superadmin'].includes(user.role)?`<button class="btn btn-primary btn-sm" onclick="showAddJadwal()"><i class="ri-add-line"></i> Tambah</button>`:''}</div>
    ${hariList.map(h=>{const items=grouped[h];if(!items.length)return '';return `<div class="card au" style="margin-bottom:.5rem">
      <h3 style="margin-bottom:.5rem"><i class="ri-calendar-line"></i> ${h}</h3>
      <div class="table-wrap"><table><tr><th>Jam</th><th>Kelompok</th><th>Ustadz</th><th>Hari</th>${['admin','superadmin'].includes(user.role)?'<th>Aksi</th>':''}</tr>
      ${items.sort((a,b)=>a.jam_mulai.localeCompare(b.jam_mulai)).map(j=>`<tr>
        <td><span class="badge-h">${j.jam_mulai}</span>  <span class="badge-s">${j.jam_selesai}</span></td>
        <td>${kelompok.find(k=>k.id===j.kelompok_id)?.nama||'#'+j.kelompok_id}</td>
        <td><strong>${j.ustadz_username}</strong></td>
        <td><span style="font-size:.72rem;padding:.15rem .4rem;border-radius:6px;background:${j.hari==='Setiap Hari'?'var(--greenbg)':'#f1f5f9'};color:${j.hari==='Setiap Hari'?'var(--green)':'var(--t2)'};font-weight:600">${j.hari}</span></td>
        ${['admin','superadmin'].includes(user.role)?`<td><button class="btn btn-danger btn-sm" onclick="hapusJadwal(${j.id})"><i class="ri-delete-bin-line"></i></button></td>`:''}</tr>`).join('')}
      </table></div></div>`}).join('')}
    ${!jadwal.length?'<div class="card au" style="text-align:center;color:var(--t3);padding:2rem">Belum ada jadwal. Klik + Tambah untuk membuat.</div>':''}`;
  window._kelompokList=kelompok;window._ustadzList=ustadzList;window._kegiatanList=kegiatan;
}
function showAddJadwal(){
  const kel=window._kelompokList||[];const ust=window._ustadzList||[];const keg=window._kegiatanList||[];
  const kegList=keg.length?keg:[...new Map(kel.filter(k=>k.kegiatan_nama).map(k=>[k.kegiatan_id||k.kegiatan_nama,{id:k.kegiatan_id,nama:k.kegiatan_nama}])).values()];
  const allHari=['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu','Minggu'];
  $('modal').innerHTML=`<h3><i class="ri-calendar-schedule-line"></i> Tambah Jadwal</h3>
    <div class="fg"><label>Pilih Hari</label>
      <div style="margin-bottom:.4rem">
        <label style="display:inline-flex;align-items:center;gap:.4rem;padding:.4rem .7rem;background:var(--greenbg);border:1.5px solid var(--green);border-radius:10px;cursor:pointer;font-size:.8rem;font-weight:600;transition:.2s"
          onclick="setTimeout(()=>{const c=this.querySelector('input');document.querySelectorAll('input[name=jHariCb]').forEach(cb=>{cb.checked=c.checked;cb.closest('label').style.background=c.checked?'var(--greenbg)':'#f8fafc';cb.closest('label').style.borderColor=c.checked?'var(--green)':'var(--border)'});this.style.background=c.checked?'var(--greenbg)':'#f8fafc';this.style.borderColor=c.checked?'var(--green)':'var(--border)'},10)">
          <input type="checkbox" id="jHariSemua" style="accent-color:var(--green);width:16px;height:16px">
          <i class="ri-calendar-check-line" style="color:var(--green)"></i> Setiap Hari
        </label>
      </div>
      <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:.4rem">
        ${allHari.map(h=>`<label style="display:flex;align-items:center;gap:.35rem;padding:.4rem .6rem;background:#f8fafc;border:1.5px solid var(--border);border-radius:10px;cursor:pointer;font-size:.78rem;font-weight:500;transition:.2s"
          onclick="setTimeout(()=>{const c=this.querySelector('input');this.style.background=c.checked?'var(--greenbg)':'#f8fafc';this.style.borderColor=c.checked?'var(--green)':'var(--border)'},10)">
          <input type="checkbox" name="jHariCb" value="${h}" style="accent-color:var(--green);width:15px;height:15px"> ${h}
        </label>`).join('')}
      </div>
    </div>
    <div style="display:flex;gap:.5rem">
      <div class="fg" style="flex:1"><label>Jam Mulai</label><input type="time" id="jMulai" value="06:00"></div>
      <div class="fg" style="flex:1"><label>Jam Selesai</label><input type="time" id="jSelesai" value="07:00"></div>
    </div>
    <div class="fg"><label>Kegiatan</label><select id="jKegiatan" onchange="filterJadwalKelompok()">
      <option value=""> Pilih Kegiatan </option>
      ${kegList.map(k=>`<option value="${k.nama}">${k.nama}</option>`).join('')}
    </select></div>
    <div class="fg"><label>Kelompok</label><select id="jKelompok">
      <option value=""> Pilih kegiatan dulu </option>
    </select></div>
    <div class="fg"><label>Ustadz</label><select id="jUstadz">
      ${ust.length?ust.map(u=>`<option value="${u.username}">${u.nama} (${u.username})</option>`).join('')
       :'<option value="">Belum ada ustadz</option>'}
    </select></div>
    <div class="info-box warn"><i class="ri-information-line"></i> Jika jadwal diatur, hanya ustadz terjadwal yang bisa absen kelompok ini (jam mulai s/d selesai + 30 menit toleransi). Kelompok tanpa jadwal tetap bebas.</div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="simpanJadwal()"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
function filterJadwalKelompok(){
  const sel=$('jKegiatan').value;const kel=window._kelompokList||[];
  const filtered=sel?kel.filter(k=>k.kegiatan_nama===sel||k.tipe===sel):[];
  $('jKelompok').innerHTML=filtered.length
    ?filtered.map(k=>`<option value="${k.id}">${k.nama}</option>`).join('')
    :'<option value=""> Tidak ada kelompok </option>';
}
async function simpanJadwal(){
  // Collect checked days
  const semuaHari=$('jHariSemua')?.checked;
  let hariList=[];
  if(semuaHari){
    hariList=['Setiap Hari'];
  }else{
    document.querySelectorAll('input[name=jHariCb]:checked').forEach(cb=>hariList.push(cb.value));
  }
  if(!hariList.length)return toast('Pilih minimal 1 hari');
  const b={hari_list:hariList,jam_mulai:$('jMulai').value,jam_selesai:$('jSelesai').value,
    kelompok_id:parseInt($('jKelompok').value),ustadz_username:$('jUstadz').value};
  if(!b.ustadz_username)return toast('Pilih ustadz');
  if(!b.kelompok_id)return toast('Pilih kelompok');
  await api('/api/jadwal-umum',{method:'POST',body:JSON.stringify(b)});
  hideModal();toast('[OK]  Jadwal ditambahkan');loadJadwalUmum();
}
async function hapusJadwal(id){if(!confirm('Hapus jadwal?'))return;await api('/api/jadwal-umum/'+id,{method:'DELETE'});toast('Dihapus');loadJadwalUmum();}

// -- Utils --
function $(id){return document.getElementById(id)}
function showModal(){$('modalBg').classList.add('active');history.pushState({modal:true},'',' ')}
function hideModal(){
  if(!$('modalBg').classList.contains('active'))return;
  window._modalClosing=true;
  $('modalBg').classList.remove('active');
  if(history.state&&history.state.modal)history.back();
  setTimeout(function(){window._modalClosing=false;},100);
}
// Close modal on back button (mobile gesture)
window.addEventListener('popstate',function(e){
  if($('modalBg').classList.contains('active')){
    window._modalClosing=true;
    $('modalBg').classList.remove('active');
    e.stopImmediatePropagation();
    setTimeout(function(){window._modalClosing=false;},100);
  }
});
function toast(msg){const t=$('toast');t.textContent=msg;t.style.display='block';clearTimeout(t._t);t._t=setTimeout(()=>t.style.display='none',3000);}

// -- Auto-login --
// -- ABSEN MALAM --
let _absMalamKamarId=null;
async function loadAbsenMalam(){
  _absMalamKamarId=null;
  const kamar=await api('/api/kamar');
  const today=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-moon-line"></i> Absen Kamar</h2></div>
    <div class="fg"><label>Tanggal</label><input type="date" id="absMalamTgl" value="${today}" onchange="if(_absMalamKamarId) openAbsenMalamKamar(_absMalamKamarId, window._absMalamKamarNama)"></div>
    <div class="abs-tabs au" id="absMalamTabs">
      ${kamar.map(k=>`<div class="abs-tab" id="tab-kamar-${k.id}" onclick="openAbsenMalamKamar(${k.id},'${k.nama.replace(/'/g,"\\'")}')">${k.nama}</div>`).join('')||'<p style="color:var(--t3)">Belum ada kamar</p>'}
    </div>
    <div id="absMalamContent">
      <div class="card au" style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih kamar di atas untuk mulai absen</div>
    </div>`;
}

async function openAbsenMalamKamar(kamarId,kamarNama){
  _absMalamKamarId=kamarId;
  window._absMalamKamarNama=kamarNama;
  document.querySelectorAll('#absMalamTabs .abs-tab').forEach(el=>el.classList.remove('active'));
  const tab=$('tab-kamar-'+kamarId);
  if(tab) tab.classList.add('active');
  $('absMalamContent').innerHTML='<p style="text-align:center;color:var(--t3);padding:1rem"><i class="ri-loader-4-line ri-spin"></i> Memuat data...</p>';
  
  const tanggal=$('absMalamTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  const [members,existing]=await Promise.all([
    api('/api/kamar/'+kamarId+'/members'),
    api('/api/absen-malam?tanggal='+tanggal+'&kamar_id='+kamarId)
  ]);
  const sudah=existing.length>0;
  if(tab){if(sudah)tab.classList.add('done');else tab.classList.remove('done');}
  
  const existingMap={};existing.forEach(a=>existingMap[a.santri_id]=a.status);
  $('absMalamContent').innerHTML=`
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.8rem" class="au">
      <h3 style="margin:0;font-size:1rem;color:var(--text)"><i class="ri-hotel-bed-line"></i> ${kamarNama}</h3>
      <span style="font-size:.75rem;color:var(--t3)">${members.length} santri</span>
    </div>
    ${sudah?'<div class="info-box success au" style="margin-bottom:.8rem"><i class="ri-checkbox-circle-line"></i> Kamar ini sudah diabsen pada tanggal ini</div>':''}
    <div class="card au" style="padding:0;overflow:hidden;border-radius:16px"><div class="table-wrap" style="border:none;box-shadow:none;border-radius:0"><table style="table-layout:fixed">
      <tr><th style="width:auto">Nama Santri</th><th style="text-align:center;width:44px">H</th><th style="text-align:center;width:44px">I</th><th style="text-align:center;width:44px">S</th><th style="text-align:center;width:44px">A</th></tr>
      ${members.map(m=>{const st=sudah?(existingMap[m.id]||'H'):'H';const dis=sudah?' disabled':'';
        return `<tr><td style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${m.nama}</td>
        <td style="text-align:center"><span class="abs-toggle abs-h${st==='H'?' active':''}${dis}" onclick="setAbsMalam(${m.id},'H',this)">H</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-i${st==='I'?' active':''}${dis}" onclick="setAbsMalam(${m.id},'I',this)">I</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-s${st==='S'?' active':''}${dis}" onclick="setAbsMalam(${m.id},'S',this)">S</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-a${st==='A'?' active':''}${dis}" onclick="setAbsMalam(${m.id},'A',this)">A</span></td></tr>`;}).join('')}
      ${!members.length?'<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>':''}
    </table></div></div>
    <div style="height:60px"></div>
    <div style="position:fixed;bottom:var(--bn);left:0;right:0;z-index:10;padding:.5rem .55rem .4rem;background:linear-gradient(to top,var(--bg) 60%,transparent)">
    ${sudah?'<button class="btn btn-full" disabled style="background:var(--green);color:#fff;opacity:.5"><i class="ri-checkbox-circle-line"></i> Sudah Diabsen</button>'
    :'<button class="btn btn-primary btn-full" id="btnAbsMalam" onclick="simpanAbsenMalam('+kamarId+')" style="box-shadow:0 -4px 16px rgba(0,0,0,.15)"><i class="ri-save-line"></i> Simpan Absen Kamar</button>'}</div>`;
  window._absMalamData={};members.forEach(m=>window._absMalamData[m.id]=sudah?(existingMap[m.id]||'H'):'H');
}

function setAbsMalam(santriId,status,el){
  window._absMalamData[santriId]=status;const row=el.closest('tr');
  row.querySelectorAll('.abs-toggle').forEach(s=>s.classList.remove('active'));el.classList.add('active');
}

async function simpanAbsenMalam(kamarId){
  const tanggal=$('absMalamTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  const items=Object.entries(window._absMalamData||{}).map(([sid,st])=>({santri_id:parseInt(sid),status:st}));
  if(!items.length)return toast('Tidak ada data');
  const btn=$('btnAbsMalam');
  if(btn){btn.disabled=true;btn.innerHTML='<i class="ri-loader-4-line"></i> Menyimpan...';btn.style.opacity='.6';}
  try{
    const res=await api('/api/absen-malam/bulk',{method:'POST',body:JSON.stringify({tanggal,kamar_id:kamarId,items})});
    toast('[OK]  '+res.message);
    if(btn){btn.innerHTML='<i class="ri-checkbox-circle-line"></i> Sudah Diabsen';btn.style.opacity='.5';btn.style.background='var(--green)';}
    document.querySelectorAll('.abs-toggle').forEach(el=>el.classList.add('disabled'));
    const tab=$('tab-kamar-'+kamarId);
    if(tab)tab.classList.add('done');
  }catch(e){if(btn){btn.disabled=false;btn.innerHTML='<i class="ri-save-line"></i> Simpan Absen Kamar';btn.style.opacity='1';}toast('Error: '+e.message);}
}

// -- KELOLA KELAS SEKOLAH --
async function loadKelasSekolah(){
  const kelas=await api('/api/kelas-sekolah');
  const isAdmin=['admin','superadmin'].includes(user.role);
  $('main').innerHTML=`<div class="page-header"><h2><i class="ri-school-line"></i> Kelola Kelas (${kelas.length})</h2>
    ${isAdmin?'<button class="btn btn-primary" onclick="showAddKelas()"><i class="ri-add-line"></i> Tambah</button>':''}</div>
    <div class="grid">${kelas.map(k=>`<div class="grid-card au" onclick="openKelasDetail(${k.id},'${k.nama.replace(/'/g,"\\'")}')"><div class="title"><i class="ri-school-line"></i> ${k.nama}</div>
      <div class="sub">${k.jumlah_santri||0} santri</div>
      ${isAdmin?`<div class="card-actions"><button class="btn btn-outline btn-sm" onclick="event.stopPropagation();showEditKelas(${k.id},'${k.nama.replace(/'/g,"\\'")}')"><i class="ri-edit-line"></i></button><button class="btn btn-danger btn-sm" onclick="event.stopPropagation();hapusKelas(${k.id})"><i class="ri-delete-bin-line"></i></button></div>`:''}</div>`).join('')||'<p style="color:var(--t3)">Belum ada kelas</p>'}
    </div>`;
}
function showAddKelas(){
  $('modal').innerHTML=`<h3><i class="ri-school-line"></i> Tambah Kelas</h3>
    <div class="fg"><label>Nama Kelas</label><input id="mdNama" placeholder="Kelas 7A"></div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="saveAddKelas()"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveAddKelas(){
  const n=$('mdNama').value.trim();if(!n)return toast('Nama wajib');
  await api('/api/kelas-sekolah',{method:'POST',body:JSON.stringify({nama:n})});
  hideModal();loadKelasSekolah();
}
function showEditKelas(id,namaLama){
  $('modal').innerHTML=`<h3><i class="ri-edit-line"></i> Edit Kelas</h3>
    <div class="fg"><label>Nama Kelas</label><input id="mdEditNama" value="${namaLama}"></div>
    <div style="display:flex;gap:.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="saveEditKelas(${id})"><i class="ri-save-line"></i> Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveEditKelas(id){
  const n=$('mdEditNama').value.trim();if(!n)return toast('Nama wajib');
  await api('/api/kelas-sekolah/'+id,{method:'PUT',body:JSON.stringify({nama:n})});
  hideModal();toast('[OK]  Kelas diperbarui');loadKelasSekolah();
}
async function hapusKelas(id){if(!confirm('Hapus kelas ini?'))return;await api('/api/kelas-sekolah/'+id,{method:'DELETE'});loadKelasSekolah();}

async function openKelasDetail(kelasId,kelasNama){
  kelasId=parseInt(kelasId);
  $('main').innerHTML='<div class="card"><p style="color:var(--t3)">Loading...</p></div>';
  try{
  var kne=kelasNama.replace(/'/g,"\\'");
  const [members,jadwal,santri,users,mapelSekolah]=await Promise.all([
    api('/api/kelas-sekolah/'+kelasId+'/members').catch(()=>[]),
    api('/api/jadwal-sekolah?kelas_id='+kelasId).catch(()=>[]),
    api('/api/santri').catch(()=>[]),
    api('/api/users').catch(()=>[]),
    api('/api/mata-pelajaran-sekolah?kelas_id='+kelasId).catch(()=>[])
  ]);
  const kelasJadwal=(jadwal||[]);
  const isAdmin=['admin','superadmin'].includes(user.role);
  const canNilai=['admin','superadmin','ustadz'].includes(user.role);
  const ustadzList=users.filter(function(u){return u.role==='ustadz';});
  var allHari=['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu'];
  var hariCbs=allHari.map(function(h){return '<label style="display:flex;align-items:center;gap:.25rem;padding:.35rem .55rem;background:#f8fafc;border:1.5px solid var(--border);border-radius:10px;cursor:pointer;font-size:.76rem;font-weight:500;transition:.2s" onclick="setTimeout(function(){var c=this.querySelector(\x27input\x27);this.style.background=c.checked?\x27var(--greenbg)\x27:\x27#f8fafc\x27;this.style.borderColor=c.checked?\x27var(--green)\x27:\x27var(--border)\x27}.bind(this),10)"><input type="checkbox" name="jsHariCb" value="'+h+'" style="accent-color:var(--green);width:14px;height:14px"> '+h+'</label>';}).join('');
  var html='<button class="back-btn" onclick="loadKelasSekolah()"><i class="ri-arrow-left-line"></i> Kembali</button>';
  html+='<div class="page-header"><h2><i class="ri-school-line"></i> '+kelasNama+'</h2></div>';
  html+='<div class="card au"><h3><i class="ri-group-line"></i> Anggota ('+members.length+')</h3>';
  if(isAdmin)html+='<button class="btn btn-primary btn-sm" style="margin-bottom:.6rem" onclick="showBulkAddSantri({title:\'Tambah Santri ke '+kelasNama.replace(/'/g,"\\'")+'\',apiUrl:\'/api/kelas-sekolah/'+kelasId+'/members/bulk\',existingIds:['+members.map(function(m){return m.id;}).join(',')+'],onDone:function(){openKelasDetail('+kelasId+',\''+kne+'\');}})"><i class="ri-user-add-line"></i> Tambah Santri</button>';
  html+='<div class="table-wrap"><table><tr><th>Nama</th><th>Kelas Diniyyah</th>'+(isAdmin?'<th></th>':'')+'</tr>';
  members.forEach(function(m){html+='<tr><td>'+m.nama+'</td><td>'+(m.kelas_diniyyah||'-')+'</td>'+(isAdmin?'<td><button class="btn btn-danger btn-sm" onclick="removeKelasMember('+kelasId+','+m.id+',\''+kne+'\')"><i class="ri-close-line"></i></button></td>':'')+'</tr>';});
  if(!members.length)html+='<tr><td colspan="3" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>';
  html+='</table></div></div>';
  html+='<div class="card au"><h3><i class="ri-calendar-schedule-line"></i> Jadwal Pelajaran</h3>';
  if(isAdmin){
    html+='<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:.8rem;align-items:end">';
    html+='<div class="fg" style="flex:2;min-width:120px"><label>Mata Pelajaran</label><input id="jsMapel" placeholder="Matematika"></div>';
    html+='<div class="fg" style="flex:2;min-width:100px"><label>Ustadz/Guru</label>'+searchSelect({id:'jsUstadz',items:ustadzList.map(function(u){return {value:u.username,label:u.nama};}),placeholder:'Ketik nama ustadz...'})+'</div>';
    html+='<div class="fg" style="flex:3;min-width:100%"><label>Hari (centang beberapa)</label><div style="display:flex;gap:.35rem;flex-wrap:wrap">'+hariCbs+'</div></div>';
    html+='<div style="display:flex;gap:.5rem;flex:2;min-width:100%"><div class="fg" style="flex:1"><label>Mulai</label><input type="time" id="jsMulai" value="07:00"></div><div class="fg" style="flex:1"><label>Selesai</label><input type="time" id="jsSelesai" value="08:00"></div>';
    html+='<button class="btn btn-primary btn-sm" style="align-self:end;margin-bottom:.8rem" onclick="tambahJadwalSekolah('+kelasId+',\''+kne+'\')"><i class="ri-add-line"></i> Tambah</button></div></div>';
  }
  html+='<div class="table-wrap"><table><tr><th>Hari</th><th>Pelajaran</th><th>Guru</th><th>Jam</th>'+(isAdmin?'<th></th>':'')+'</tr>';
  kelasJadwal.sort(function(a,b){var d=['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu'];return d.indexOf(a.hari)-d.indexOf(b.hari)||a.jam_mulai.localeCompare(b.jam_mulai);}).forEach(function(j){
    html+='<tr><td>'+j.hari+'</td><td>'+j.mata_pelajaran+'</td><td>'+j.ustadz_username+'</td><td>'+j.jam_mulai+' - '+j.jam_selesai+'</td>'+(isAdmin?'<td><button class="btn btn-danger btn-sm" onclick="hapusJadwalSekolah('+j.id+','+kelasId+',\''+kne+'\')"><i class="ri-delete-bin-line"></i></button></td>':'')+'</tr>';
  });
  if(!kelasJadwal.length)html+='<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada jadwal</td></tr>';
  html+='</table></div></div>';
  if(canNilai){
    html+='<div class="card au"><h3><i class="ri-book-open-line"></i> Mata Pelajaran Nilai ('+mapelSekolah.length+')</h3>';
    if(isAdmin)html+='<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:.8rem;align-items:end"><div class="fg" style="flex:2;min-width:150px"><label>Nama Mapel</label><input id="mpsNama" placeholder="Matematika, IPA..."></div><button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="tambahMapelSekolah('+kelasId+',\''+kne+'\')"><i class="ri-add-line"></i> Tambah</button></div>';
    html+='<div class="table-wrap"><table><tr><th>Nama</th>'+(isAdmin?'<th></th>':'')+'</tr>';
    mapelSekolah.forEach(function(m){html+='<tr><td>'+m.nama+'</td>'+(isAdmin?'<td><button class="btn btn-danger btn-sm" onclick="hapusMapelSekolah('+m.id+','+kelasId+',\''+kne+'\')"><i class="ri-delete-bin-line"></i></button></td>':'')+'</tr>';});
    if(!mapelSekolah.length)html+='<tr><td colspan="2" style="text-align:center;color:var(--t3)">Belum ada mapel nilai</td></tr>';
    html+='</table></div>';
    if(mapelSekolah.length)html+='<button class="btn btn-gold" style="margin-top:.8rem" onclick="window._inputNilaiKS='+kelasId+';window._inputNilaiKSNama=\''+kne+'\';nav(\'input-nilai-sekolah\')"><i class="ri-edit-line"></i> Input Nilai Semester</button>';
    html+='</div>';
  }
  $('main').innerHTML=html;
  window._kelasAllSantri=santri||[];
  }catch(e){
    $('main').innerHTML='<button class="back-btn" onclick="loadKelasSekolah()"><i class="ri-arrow-left-line"></i> Kembali</button><div class="card"><p style="color:var(--red)">Error: '+e.message+'</p></div>';
  }
}
function searchKelasAdd(q,kelasId){
  kelasId=parseInt(kelasId);
  const box=$('ksMbResults');if(!q){box.innerHTML='';return;}
  const r=(window._kelasAllSantri||[]).filter(s=>s.nama.toLowerCase().includes(q.toLowerCase())).slice(0,6);
  box.innerHTML=r.map(s=>`<div onclick="addKelasMember(${kelasId},${s.id})">${s.nama}</div>`).join('');
}
async function addKelasMember(kelasId,santriId){
  kelasId=parseInt(kelasId);santriId=parseInt(santriId);
  try{await api('/api/kelas-sekolah/'+kelasId+'/members',{method:'POST',body:JSON.stringify({santri_id:santriId})});toast('[OK]  Ditambahkan');
    const kelas=await api('/api/kelas-sekolah');const k=kelas.find(x=>x.id===kelasId);openKelasDetail(kelasId,k?.nama||'Kelas');
  }catch(e){toast('Error: '+e.message);}
}
async function removeKelasMember(kelasId,santriId,kelasNama){
  if(!confirm('Hapus santri dari kelas?'))return;
  try{await api('/api/kelas-sekolah/'+kelasId+'/members/'+santriId,{method:'DELETE'});toast('Dihapus');openKelasDetail(kelasId,kelasNama);}catch(e){toast('Error: '+e.message);}
}
async function tambahJadwalSekolah(kelasId,kelasNama){
  kelasId=parseInt(kelasId);
  const mp=$('jsMapel').value.trim(),u=$('jsUstadz_val')?$('jsUstadz_val').value:$('jsUstadz').value,jm=$('jsMulai').value,js=$('jsSelesai').value;
  const hariList=[...document.querySelectorAll('input[name=jsHariCb]:checked')].map(cb=>cb.value);
  if(!mp)return toast('Mata pelajaran wajib');
  if(!hariList.length)return toast('Pilih minimal 1 hari');
  try{await api('/api/jadwal-sekolah',{method:'POST',body:JSON.stringify({kelas_id:kelasId,mata_pelajaran:mp,ustadz_username:u,hari_list:hariList,jam_mulai:jm,jam_selesai:js})});
  toast('Jadwal ditambahkan untuk '+hariList.length+' hari');openKelasDetail(kelasId,kelasNama);
  }catch(e){toast('Error: '+e.message);}
}
async function hapusJadwalSekolah(id,kelasId,kelasNama){
  if(!confirm('Hapus jadwal ini?'))return;
  try{await api('/api/jadwal-sekolah/'+id,{method:'DELETE'});toast('Dihapus');openKelasDetail(kelasId,kelasNama);}catch(e){toast('Error: '+e.message);}
}
async function tambahMapelSekolah(kelasId,kelasNama){
  const n=$('mpsNama')?.value?.trim();if(!n)return toast('Nama mapel wajib');
  try{await api('/api/mata-pelajaran-sekolah',{method:'POST',body:JSON.stringify({kelas_id:kelasId,nama:n})});
  toast('Mapel ditambahkan');openKelasDetail(kelasId,kelasNama);}catch(e){toast('Error: '+e.message);}
}
async function hapusMapelSekolah(mpId,kelasId,kelasNama){
  if(!confirm('Hapus mata pelajaran ini? Semua nilai terkait akan terhapus.'))return;
  try{await api('/api/mata-pelajaran-sekolah/'+mpId,{method:'DELETE'});toast('Dihapus');openKelasDetail(kelasId,kelasNama);}catch(e){toast('Error: '+e.message);}
}


// -- ABSEN SEKOLAH --
let _absSekolahState={kelasId:null,kelasNama:null,jadwalId:null,mapel:null};
async function loadAbsenSekolah(){
  _absSekolahState={kelasId:null,kelasNama:null,jadwalId:null,mapel:null};
  const kelas=await api('/api/kelas-sekolah');
  const today=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-school-line"></i> Absen Sekolah</h2></div>
    <div class="fg"><label>Tanggal</label><input type="date" id="absSekolahTgl" value="${today}" onchange="if(_absSekolahState.kelasId) absSekolahPilihKelas(_absSekolahState.kelasId, _absSekolahState.kelasNama)"></div>
    <div class="abs-tabs au" id="absSekolahTabs">
      ${kelas.map(k=>`<div class="abs-tab" id="tab-kelas-${k.id}" onclick="absSekolahPilihKelas(${k.id},'${k.nama.replace(/'/g,"\\'")}')">${k.nama}</div>`).join('')||'<p style="color:var(--t3)">Belum ada kelas</p>'}
    </div>
    <div id="absSekolahJadwalContainer" class="au"></div>
    <div id="absSekolahContent">
      <div class="card au" style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih kelas di atas untuk mulai absen</div>
    </div>`;
}

async function absSekolahPilihKelas(kelasId,kelasNama){
  _absSekolahState.kelasId=kelasId;
  _absSekolahState.kelasNama=kelasNama;
  _absSekolahState.jadwalId=null;
  _absSekolahState.mapel=null;
  
  document.querySelectorAll('#absSekolahTabs .abs-tab').forEach(el=>el.classList.remove('active'));
  const tab=$('tab-kelas-'+kelasId);
  if(tab) tab.classList.add('active');
  
  $('absSekolahJadwalContainer').innerHTML='<p style="text-align:center;color:var(--t3);font-size:.8rem;padding:.5rem"><i class="ri-loader-4-line ri-spin"></i> Memuat jadwal...</p>';
  $('absSekolahContent').innerHTML='';
  
  const hariMap={0:'Minggu',1:'Senin',2:'Selasa',3:'Rabu',4:'Kamis',5:'Jumat',6:'Sabtu'};
  const hariIni=hariMap[new Date().getDay()];
  const jadwal=await api('/api/jadwal-sekolah?kelas_id='+kelasId);
  const todayJ=jadwal.filter(j=>j.hari===hariIni).sort((a,b)=>a.jam_mulai.localeCompare(b.jam_mulai));
  
  let html=`<div style="font-size:.75rem;color:var(--t3);margin-bottom:.4rem;padding-left:.2rem;font-weight:600">Pilih Mata Pelajaran (${hariIni}):</div>`;
  html += `<div class="abs-tabs" id="absSekolahJadwalTabs" style="margin-bottom:1rem">`;
  if(todayJ.length){
    todayJ.forEach(j=>{
      html += `<div class="abs-tab" id="tab-jadwal-${j.id}" onclick="absSekolahPilihJadwal(${j.id},'${j.mata_pelajaran.replace(/'/g,"\\'")}')" style="display:flex;flex-direction:column;gap:.1rem;align-items:center;padding:.3rem .8rem">
        <span style="font-size:.8rem">${j.mata_pelajaran}</span>
        <span style="font-size:.65rem;font-weight:500;opacity:.8">${j.jam_mulai}-${j.jam_selesai}</span>
      </div>`;
    });
  }
  html += `<div class="abs-tab" id="tab-jadwal-0" onclick="absSekolahPilihJadwal(0,'Manual')" style="display:flex;flex-direction:column;gap:.1rem;align-items:center;background:var(--pbg);color:var(--p);border-color:var(--pbg);padding:.3rem .8rem">
      <span style="font-size:.8rem">Absen Manual</span>
      <span style="font-size:.65rem;font-weight:500;opacity:.8">Tanpa Jadwal</span>
    </div>`;
  html += `</div>`;
  
  $('absSekolahJadwalContainer').innerHTML = html;
  $('absSekolahContent').innerHTML = '<div class="card au" style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih mata pelajaran di atas</div>';
}

async function absSekolahPilihJadwal(jadwalId,mapel){
  _absSekolahState.jadwalId=jadwalId;
  _absSekolahState.mapel=mapel;
  const {kelasId,kelasNama}=_absSekolahState;
  
  document.querySelectorAll('#absSekolahJadwalTabs .abs-tab').forEach(el=>el.classList.remove('active'));
  const tab=$('tab-jadwal-'+jadwalId);
  if(tab) tab.classList.add('active');
  
  $('absSekolahContent').innerHTML='<p style="text-align:center;color:var(--t3);padding:1rem"><i class="ri-loader-4-line ri-spin"></i> Memuat data...</p>';
  
  const tanggal=$('absSekolahTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  const [members,existing]=await Promise.all([
    api('/api/kelas-sekolah/'+kelasId+'/members'),
    api('/api/absen-sekolah?tanggal='+tanggal+'&kelas_id='+kelasId+'&jadwal_sekolah_id='+jadwalId)
  ]);
  const sudah=existing.length>0;
  if(tab){if(sudah)tab.classList.add('done');else tab.classList.remove('done');}
  
  const existingMap={};existing.forEach(a=>existingMap[a.santri_id]=a.status);
  $('absSekolahContent').innerHTML=`
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.8rem" class="au">
      <h3 style="margin:0;font-size:1rem;color:var(--text)"><i class="ri-book-2-line"></i> ${kelasNama} — ${mapel}</h3>
      <span style="font-size:.75rem;color:var(--t3)">${members.length} santri</span>
    </div>
    ${sudah?'<div class="info-box success au" style="margin-bottom:.8rem"><i class="ri-checkbox-circle-line"></i> Sudah diabsen untuk pelajaran ini</div>':''}
    <div class="card au" style="padding:0;overflow:hidden;border-radius:16px"><div class="table-wrap" style="border:none;box-shadow:none;border-radius:0"><table style="table-layout:fixed">
      <tr><th style="width:auto">Nama Santri</th><th style="text-align:center;width:44px">H</th><th style="text-align:center;width:44px">I</th><th style="text-align:center;width:44px">S</th><th style="text-align:center;width:44px">A</th></tr>
      ${members.map(m=>{const st=sudah?(existingMap[m.id]||'H'):'H';const dis=sudah?' disabled':'';
        return `<tr><td style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${m.nama}</td>
        <td style="text-align:center"><span class="abs-toggle abs-h${st==='H'?' active':''}${dis}" onclick="setAbsSekolah(${m.id},'H',this)">H</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-i${st==='I'?' active':''}${dis}" onclick="setAbsSekolah(${m.id},'I',this)">I</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-s${st==='S'?' active':''}${dis}" onclick="setAbsSekolah(${m.id},'S',this)">S</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-a${st==='A'?' active':''}${dis}" onclick="setAbsSekolah(${m.id},'A',this)">A</span></td></tr>`;}).join('')}
      ${!members.length?'<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada anggota.<br>Tambahkan via Kelola Kelas.</td></tr>':''}
    </table></div></div>
    <div style="height:60px"></div>
    <div style="position:fixed;bottom:var(--bn);left:0;right:0;z-index:10;padding:.5rem .55rem .4rem;background:linear-gradient(to top,var(--bg) 60%,transparent)">
    ${sudah?'<button class="btn btn-full" disabled style="background:var(--green);color:#fff;opacity:.5"><i class="ri-checkbox-circle-line"></i> Sudah Diabsen</button>'
    :'<button class="btn btn-primary btn-full" id="btnAbsSekolah" onclick="simpanAbsenSekolah()" style="box-shadow:0 -4px 16px rgba(0,0,0,.15)"><i class="ri-save-line"></i> Simpan Absen Sekolah</button>'}</div>`;
  window._absSekolahData={};members.forEach(m=>window._absSekolahData[m.id]=sudah?(existingMap[m.id]||'H'):'H');
}

function setAbsSekolah(santriId,status,el){
  window._absSekolahData[santriId]=status;const row=el.closest('tr');
  row.querySelectorAll('.abs-toggle').forEach(s=>s.classList.remove('active'));el.classList.add('active');
}

async function simpanAbsenSekolah(){
  const {kelasId,jadwalId,mapel,kelasNama}=_absSekolahState;
  const tanggal=$('absSekolahTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  const items=Object.entries(window._absSekolahData||{}).map(([sid,st])=>({santri_id:parseInt(sid),status:st}));
  if(!items.length)return toast('Tidak ada data');
  const btn=$('btnAbsSekolah');
  if(btn){btn.disabled=true;btn.innerHTML='<i class="ri-loader-4-line ri-spin"></i> Menyimpan...';btn.style.opacity='.6';}
  try{
    const res=await api('/api/absen-sekolah/bulk',{method:'POST',body:JSON.stringify({tanggal,kelas_id:kelasId,jadwal_sekolah_id:jadwalId,mata_pelajaran:mapel,items})});
    toast('[OK]  '+res.message);
    if(btn){btn.innerHTML='<i class="ri-checkbox-circle-line"></i> Sudah Diabsen';btn.style.opacity='.5';btn.style.background='var(--green)';}
    document.querySelectorAll('.abs-toggle').forEach(el=>el.classList.add('disabled'));
    const tab=$('tab-jadwal-'+jadwalId);
    if(tab)tab.classList.add('done');
  }catch(e){if(btn){btn.disabled=false;btn.innerHTML='<i class="ri-save-line"></i> Simpan Absen Sekolah';btn.style.opacity='1';}toast('Error: '+e.message);}
}

// -- PEMBAYARAN SPP --
let _pembayaranBulan='';
function fmtRp(n){return 'Rp '+Number(n||0).toLocaleString('id-ID');}

async function loadPembayaran(){
  try{
  const now=new Date(Date.now()+7*3600000);
  _pembayaranBulan=_pembayaranBulan||now.toISOString().slice(0,7);
  const isAdmin=['admin','superadmin'].includes(user.role);
  let rekap={};
  try{rekap=await api('/api/pembayaran/rekap?bulan='+_pembayaranBulan);}catch(e){rekap={};}
  const persen=rekap.persen||0;
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-money-dollar-circle-line"></i> Pembayaran SPP</h2>
    <div class="fg" style="margin:0;min-width:150px"><input type="month" id="pbBulan" value="${_pembayaranBulan}" onchange="_pembayaranBulan=this.value;loadPembayaran()"></div></div>
  <div class="stat-grid" style="grid-template-columns:repeat(3,1fr)">
    <div class="stat-card c-blue au"><div class="stat-info"><div class="sn" style="font-size:1.1rem">${fmtRp(rekap.total_tagihan)}</div><div class="sl">Total Tagihan</div></div><div class="si"><i class="ri-bill-line"></i></div></div>
    <div class="stat-card c-green au"><div class="stat-info"><div class="sn" style="font-size:1.1rem">${fmtRp(rekap.total_bayar)}</div><div class="sl">Terbayar</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
    <div class="stat-card c-amber au"><div class="stat-info"><div class="sn" style="font-size:1.3rem">${persen}%</div><div class="sl">Progress</div></div><div class="si"><i class="ri-pie-chart-line"></i></div></div>
  </div>
  <div class="card au" style="padding:.6rem .9rem">
    <div style="display:flex;justify-content:space-between;font-size:.75rem;margin-bottom:.3rem"><span>Lunas: <b style="color:var(--green)">${rekap.lunas||0}</b></span><span>Kurang: <b style="color:var(--amber)">${rekap.kurang||0}</b></span><span>Belum: <b style="color:var(--red)">${rekap.belum||0}</b></span></div>
    <div style="background:#e2e8f0;border-radius:8px;height:10px;overflow:hidden"><div style="background:linear-gradient(90deg,#16a34a,#22c55e);height:100%;width:${persen}%;border-radius:8px;transition:width .6s ease"></div></div>
  </div>
  <div id="pbTabs" style="display:flex;gap:.4rem;flex-wrap:wrap;margin:.8rem 0 .5rem">
    <button class="btn btn-primary btn-sm" onclick="pbTab('daftar')"><i class="ri-list-check-2"></i> Daftar</button>
    <button class="btn btn-outline btn-sm" onclick="pbTab('kategori')"><i class="ri-price-tag-3-line"></i> Kategori</button>
    <button class="btn btn-outline btn-sm" onclick="pbTab('tarif')"><i class="ri-settings-3-line"></i> Tarif</button>
    <button class="btn btn-outline btn-sm" onclick="pbTab('potongan')"><i class="ri-scissors-cut-line"></i> Potongan</button>
    <button class="btn btn-outline btn-sm" onclick="pbTab('bayar')"><i class="ri-hand-coin-line"></i> Bayar</button>
    <button class="btn btn-outline btn-sm" onclick="pbTab('riwayat')"><i class="ri-history-line"></i> Riwayat</button>
    <button class="btn btn-gold btn-sm" onclick="showExportSppModal()"><i class="ri-file-excel-2-line"></i> Export</button>
    <button class="btn btn-primary btn-sm" onclick="showBulkBayarModal()" style="margin-left:auto"><i class="ri-hand-coin-fill"></i> Bayar Massal</button>
  </div>
  <div id="pbContent"></div>`;
  pbTab('daftar');
  }catch(e){toast('Error: '+e.message);console.error(e);}
}

function pbTab(tab){
  document.querySelectorAll('#pbTabs button').forEach(b=>{b.className=b.className.replace('btn-primary','btn-outline');});
  const labels={daftar:'Daftar',kategori:'Kategori',tarif:'Tarif',potongan:'Potongan',bayar:'Bayar',riwayat:'Riwayat'};
  document.querySelectorAll('#pbTabs button').forEach(b=>{if(b.textContent.trim().includes(labels[tab]||'xxx'))b.className=b.className.replace('btn-outline','btn-primary');});
  if(tab==='daftar')showPembayaranList();
  else if(tab==='kategori')showKategoriPanel();
  else if(tab==='tarif')showTarifPanel();
  else if(tab==='potongan')showPotonganPanel();
  else if(tab==='bayar')showBayarPanel();
  else if(tab==='riwayat')showRiwayatPanel();
}

window.showExportSppModal = async function() {
  let kamars = [], kelas = [];
  try {
    kamars = await api('/api/kamar');
    kelas = await api('/api/kelas-diniyyah');
  } catch(e) {}
  let html = `<h3 style="margin-bottom:1rem"><i class="ri-file-excel-2-line"></i> Export Rekap SPP</h3>
    <div style="display:flex;gap:.5rem;flex-wrap:wrap">
      <div class="fg" style="flex:1;min-width:140px"><label>Bulan Mulai</label><input type="month" id="exportBulanMulai" value="${_pembayaranBulan}"></div>
      <div class="fg" style="flex:1;min-width:140px"><label>Bulan Akhir</label><input type="month" id="exportBulanAkhir" value="${_pembayaranBulan}"></div>
    </div>
    <div class="fg"><label>Filter Kamar (Opsional)</label><select id="exportKamar"><option value="">-- Semua Kamar --</option>
    ${kamars.map(k=>`<option value="${k.id}">${k.nama}</option>`).join('')}</select></div>
    <div class="fg"><label>Filter Kelas Diniyyah (Opsional)</label><select id="exportKelas"><option value="">-- Semua Kelas --</option>
    ${kelas.map(k=>`<option value="${k.id}">${k.nama}</option>`).join('')}</select></div>
    <p style="font-size:.72rem;color:var(--t3);margin-top:.3rem"><i class="ri-information-line"></i> Jika memilih lebih dari 1 bulan, setiap bulan akan mendapat kolom Tagihan, Bayar, dan Status terpisah.</p>
    <div style="display:flex;gap:.5rem;margin-top:1.5rem">
      <button class="btn btn-primary" onclick="doExportSpp()"><i class="ri-download-2-line"></i> Download Excel</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
  $('modal').innerHTML=html;showModal();
};

window.doExportSpp = function() {
  const bm = $('exportBulanMulai').value;
  const ba = $('exportBulanAkhir').value;
  const k = $('exportKamar').value;
  const c = $('exportKelas').value;
  if(!bm||!ba) return toast('Pilih bulan mulai dan akhir');
  if(bm>ba) return toast('Bulan mulai harus <= bulan akhir');
  let url = '/api/pembayaran/export-excel?bulan_mulai='+bm+'&bulan_akhir='+ba;
  if(k) url += '&kamar_id='+k;
  if(c) url += '&kelas_id='+c;
  apiDownload(url, 'Rekap_SPP_'+bm+(bm!==ba?'_sd_'+ba:'')+'.xlsx');
  hideModal();
};

async function showPembayaranList(){
  try{
  const data=await api('/api/pembayaran?bulan='+_pembayaranBulan);
  const arr=Array.isArray(data)?data:[];
  $('pbContent').innerHTML=`<div class="card au"><h3><i class="ri-list-check-2"></i> Status Pembayaran - ${_pembayaranBulan}</h3>
    <div class="table-wrap"><table>
    <tr><th>Nama</th><th>Kamar</th><th>Kategori</th><th style="text-align:right">Tarif</th><th style="text-align:right">Potongan</th><th style="text-align:right">Tagihan</th><th style="text-align:right">Dibayar</th><th style="text-align:center">Status</th></tr>
    ${arr.map(s=>{
      const badge=s.status==='LUNAS'?'badge-h':s.status==='KURANG'?'badge-i':s.status==='GRATIS'?'badge-s':'badge-a';
      const st=s.status==='KURANG'?'Kurang '+fmtRp(s.kekurangan):s.status;
      const kat = s.kategori_nama && s.kategori_nama !== 'Default' ? s.kategori_nama : '-';
      return '<tr><td>'+s.nama+'</td><td>'+(s.kamar||'-')+'</td><td><span style="font-size:.65rem;background:#f3f4f6;padding:.1rem .3rem;border-radius:4px">'+kat+'</span></td><td style="text-align:right">'+fmtRp(s.tarif)+'</td><td style="text-align:right;color:var(--red)">'+(s.potongan?'-'+fmtRp(s.potongan):'-')+'</td><td style="text-align:right;font-weight:700">'+fmtRp(s.tagihan)+'</td><td style="text-align:right">'+fmtRp(s.total_bayar)+'</td><td style="text-align:center"><span class="'+badge+'">'+st+'</span></td></tr>';
    }).join('')}
    ${!arr.length?'<tr><td colspan="8" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>':''}
    </table></div></div>`;
  }catch(e){$('pbContent').innerHTML='<div class="card au" style="text-align:center;padding:2rem;color:var(--t3)">Error: '+e.message+'</div>';}
}

async function showKategoriPanel(){
  try{
  const kat=await api('/api/pembayaran/kategori');
  const arr=Array.isArray(kat)?kat:[];
  $('pbContent').innerHTML=`<div class="card au"><h3><i class="ri-price-tag-3-line"></i> Kategori SPP Santri</h3>
    <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Atur kategori pembayaran untuk dikenakan ke santri. Kategori ini nantinya dapat diubah pada menu Data Santri. Total nominal kategori akan ditambahkan pada tagihan.</p>
    <div class="table-wrap"><table>
    <tr><th>Nama Kategori</th><th style="text-align:right">Nominal</th><th>Status</th><th>Aksi</th></tr>
    ${arr.map(t=>'<tr><td>'+t.nama+'</td><td style="text-align:right;font-weight:700">'+fmtRp(t.nominal)+'</td><td><span class="'+(t.aktif?'badge-h':'badge-a')+'">'+(t.aktif?'Aktif':'Nonaktif')+'</span></td><td style="white-space:nowrap"><button class="btn btn-outline btn-sm" onclick="showKategoriPeriode('+t.id+',\''+t.nama.replace(/'/g,"\\'").replace(/"/g,"&quot;")+'\')" title="Atur Periode"><i class="ri-calendar-line"></i></button> <button class="btn btn-outline btn-sm" onclick="showBulkSetKategori('+t.id+',\''+t.nama.replace(/'/g,"\\'").replace(/"/g,"&quot;")+'\')" title="Tetapkan Massal"><i class="ri-group-line"></i></button> <button class="btn btn-danger btn-sm" onclick="hapusKategori('+t.id+')"><i class="ri-delete-bin-line"></i></button></td></tr>').join('')}
    ${!arr.length?'<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada kategori</td></tr>':''}
    </table></div>
    <div style="margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--border)">
      <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-add-circle-line"></i> Tambah Kategori</h4>
      <div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end">
        <div class="fg" style="flex:2;min-width:150px"><label>Nama</label><input id="katNama" placeholder="Reguler / KIP / Khusus"></div>
        <div class="fg" style="flex:1;min-width:120px"><label>Nominal (Rp)</label><input id="katNominal" type="number" placeholder="500000"></div>
        <button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="simpanKategori()"><i class="ri-add-line"></i> Tambah</button>
      </div>
    </div></div>`;
  }catch(e){$('pbContent').innerHTML='<div class="card au" style="text-align:center;padding:2rem;color:var(--t3)">Error: '+e.message+'</div>';}
}
async function simpanKategori(){
  try{const n=$('katNama').value.trim(),v=parseInt($('katNominal').value)||0;
  if(!n)return toast('Nama wajib');
  await api('/api/pembayaran/kategori',{method:'POST',body:JSON.stringify({nama:n,nominal:v})});
  toast('Kategori ditambahkan');showKategoriPanel();}catch(e){toast('Error: '+e.message);}
}
async function hapusKategori(id){if(!confirm('Hapus kategori ini? Santri yang memakainya akan direset.'))return;
  try{await api('/api/pembayaran/kategori/'+id,{method:'DELETE'});toast('Dihapus');showKategoriPanel();}catch(e){toast('Error: '+e.message);}
}

async function showBulkSetKategori(kategoriId, nama) {
  const allSantri=await api('/api/santri').catch(()=>[]);
  let selected=new Set();
  function render(q){
    const filt=q?allSantri.filter(s=>s.nama.toLowerCase().includes(q.toLowerCase()) || (s.kamar_nama||'').toLowerCase().includes(q.toLowerCase()) || (s.kelas_diniyyah||'').toLowerCase().includes(q.toLowerCase()) || (s.kelas_sekolah||'').toLowerCase().includes(q.toLowerCase())):allSantri;
    var rows='';
    filt.forEach(function(s){
      var chk=selected.has(s.id)?'checked':'';
      var badge = s.kategori_spp_id === kategoriId ? ' <span style="font-size:.65rem;color:var(--green)">(Sudah)</span>' : '';
      var kelas = s.kelas_diniyyah ? '<span style="font-size:.65rem;background:#f3f4f6;padding:.1rem .3rem;border-radius:4px;margin-left:.3rem">'+s.kelas_diniyyah+'</span>' : '';
      rows+='<label style="display:flex;align-items:center;gap:.5rem;padding:.45rem .6rem;border-bottom:1px solid var(--border);cursor:pointer;transition:.15s" onmouseenter="this.style.background=\'rgba(22,163,74,.04)\'" onmouseleave="this.style.background=\'transparent\'">';
      rows+='<input type="checkbox" '+chk+' onchange="this.checked?window._bulkKatSel.add('+s.id+'):window._bulkKatSel.delete('+s.id+');document.getElementById(\'bulkCount\').textContent=window._bulkKatSel.size" style="accent-color:var(--green);width:16px;height:16px;flex-shrink:0">';
      rows+='<span style="flex:1;font-size:.82rem">'+s.nama+kelas+badge+'</span>';
      rows+='<span style="font-size:.7rem;color:var(--t3)">'+(s.kamar_nama||'')+'</span></label>';
    });
    if(!filt.length)rows='<div style="text-align:center;padding:1.5rem;color:var(--t3)">Tidak ditemukan</div>';
    var hdr='<span style="font-size:.78rem;color:var(--t3)">'+filt.length+' santri</span>';
    document.getElementById('bulkList').innerHTML=rows;
    document.getElementById('bulkInfo').innerHTML=hdr;
  }
  window._bulkKatSel=selected;
  var html='<h3 style="margin-bottom:.8rem"><i class="ri-price-tag-3-line"></i> Tetapkan Kategori: '+nama+'</h3>';
  html+='<div style="position:relative;margin-bottom:.6rem"><i class="ri-search-line" style="position:absolute;left:.7rem;top:50%;transform:translateY(-50%);color:var(--t3)"></i><input type="text" id="bulkSearch" placeholder="Cari nama, kelas, atau kamar..." oninput="window._bulkKatRender(this.value)" style="width:100%;padding:.5rem .8rem .5rem 2rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem;background:var(--card)"></div>';
  html+='<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.4rem"><label style="display:flex;align-items:center;gap:.3rem;cursor:pointer;font-size:.78rem;font-weight:600"><input type="checkbox" id="bulkAll" onchange="var cs=document.querySelectorAll(\'#bulkList input[type=checkbox]\');cs.forEach(function(c){c.checked=this.checked;if(this.checked)window._bulkKatSel.add(parseInt(c.closest(\'label\').querySelector(\'input\').onchange.toString().match(/add\\((\\d+)\\)/)[1]));else window._bulkKatSel.clear();}.bind(this));document.getElementById(\'bulkCount\').textContent=window._bulkKatSel.size" style="accent-color:var(--green)"> Pilih Semua</label><span id="bulkInfo"></span></div>';
  html+='<div id="bulkList" style="max-height:50vh;overflow-y:auto;border:1.5px solid var(--border);border-radius:12px;background:#fff"></div>';
  html+='<div style="display:flex;gap:.5rem;margin-top:1rem;align-items:center"><button class="btn btn-primary" onclick="doBulkSetKategori('+kategoriId+')"><i class="ri-save-line"></i> Simpan (<span id="bulkCount">0</span> Santri)</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>';
  $('modal').innerHTML=html;showModal();
  window._bulkKatRender=render;
  render('');
}
async function doBulkSetKategori(kategoriId){
  var ids=[...window._bulkKatSel];
  if(!ids.length)return toast('Pilih santri dulu');
  try{
    var r=await api('/api/santri/bulk-kategori',{method:'PUT',body:JSON.stringify({santri_ids:ids, kategori_spp_id: kategoriId})});
    hideModal();toast(r.message||'Kategori diterapkan ke '+ids.length+' santri');
    showKategoriPanel();
  }catch(e){toast('Error: '+e.message);}
}

async function showTarifPanel(){
  try{
  const tarif=await api('/api/pembayaran/tarif');
  const arr=Array.isArray(tarif)?tarif:[];
  const total=arr.filter(t=>t.aktif).reduce((s,t)=>s+(t.nominal||0),0);
  $('pbContent').innerHTML=`<div class="card au"><h3><i class="ri-settings-3-line"></i> Rincian Tarif Syahriyah</h3>
    <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Masukkan rincian biaya bulanan. Klik <b>Periode</b> untuk atur nominal berbeda per bulan.</p>
    <div class="table-wrap"><table>
    <tr><th>Nama Komponen</th><th style="text-align:right">Nominal Default</th><th>Status</th><th>Aksi</th></tr>
    ${arr.map(t=>'<tr><td>'+t.nama+'</td><td style="text-align:right;font-weight:700">'+fmtRp(t.nominal)+'</td><td><span class="'+(t.aktif?'badge-h':'badge-a')+'">'+(t.aktif?'Aktif':'Nonaktif')+'</span></td><td style="white-space:nowrap"><button class="btn btn-outline btn-sm" onclick="showTarifPeriode('+t.id+',\''+t.nama.replace(/'/g,"\\\\'")+'\')" title="Atur Periode"><i class="ri-calendar-line"></i></button> <button class="btn btn-danger btn-sm" onclick="hapusTarif('+t.id+')"><i class="ri-delete-bin-line"></i></button></td></tr>').join('')}
    ${!arr.length?'<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada tarif</td></tr>':''}
    <tr style="background:var(--greenbg);font-weight:700"><td>TOTAL / BULAN (Default)</td><td style="text-align:right;color:var(--green)">${fmtRp(total)}</td><td colspan="2"></td></tr>
    </table></div>
    <div style="margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--border)">
      <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-add-circle-line"></i> Tambah Komponen</h4>
      <div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end">
        <div class="fg" style="flex:2;min-width:150px"><label>Nama</label><input id="tarifNama" placeholder="SPP / Makan / Laundry"></div>
        <div class="fg" style="flex:1;min-width:120px"><label>Nominal (Rp)</label><input id="tarifNominal" type="number" placeholder="300000"></div>
        <button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="simpanTarif()"><i class="ri-add-line"></i> Tambah</button>
      </div>
    </div></div>`;
  }catch(e){$('pbContent').innerHTML='<div class="card au" style="text-align:center;padding:2rem;color:var(--t3)">Error: '+e.message+'</div>';}
}
async function simpanTarif(){
  try{const n=$('tarifNama').value.trim(),v=parseInt($('tarifNominal').value)||0;
  if(!n||v<=0)return toast('Nama dan nominal wajib');
  await api('/api/pembayaran/tarif',{method:'POST',body:JSON.stringify({nama:n,nominal:v})});
  toast('Tarif ditambahkan');showTarifPanel();}catch(e){toast('Error: '+e.message);}
}
async function hapusTarif(id){if(!confirm('Hapus tarif ini?'))return;
  try{await api('/api/pembayaran/tarif/'+id,{method:'DELETE'});toast('Dihapus');showTarifPanel();}catch(e){toast('Error: '+e.message);}
}

// ── Tarif Periode ──
async function showTarifPeriode(tarifId, nama){
  try{
  const periodeList=await api('/api/pembayaran/tarif/'+tarifId+'/periode');
  const arr=Array.isArray(periodeList)?periodeList:[];
  let rows=arr.map(p=>'<tr><td>'+p.bulan_mulai+'</td><td>'+p.bulan_akhir+'</td><td style="text-align:right;font-weight:700">'+fmtRp(p.nominal)+'</td><td><button class="btn btn-danger btn-sm" onclick="hapusTarifPeriode('+p.id+','+tarifId+',\''+nama.replace(/'/g,"\\'")+'\')"><i class="ri-delete-bin-line"></i></button></td></tr>').join('');
  if(!arr.length) rows='<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada periode khusus. Semua bulan pakai nominal default.</td></tr>';
  let html='<h3 style="margin-bottom:.5rem"><i class="ri-calendar-line"></i> Periode Tarif: '+nama+'</h3>';
  html+='<p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Bulan yang tidak tercakup periode menggunakan nominal default. Periode tidak boleh tumpang tindih.</p>';
  html+='<div class="table-wrap"><table><tr><th>Bulan Mulai</th><th>Bulan Akhir</th><th style="text-align:right">Nominal</th><th>Aksi</th></tr>'+rows+'</table></div>';
  html+='<div style="margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--border)">';
  html+='<h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-add-circle-line"></i> Tambah Periode</h4>';
  html+='<div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end">';
  html+='<div class="fg" style="flex:1;min-width:130px"><label>Bulan Mulai</label><input type="month" id="tpMulai"></div>';
  html+='<div class="fg" style="flex:1;min-width:130px"><label>Bulan Akhir</label><input type="month" id="tpAkhir"></div>';
  html+='<div class="fg" style="flex:1;min-width:120px"><label>Nominal (Rp)</label><input type="number" id="tpNom" placeholder="300000"></div>';
  html+='<button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="simpanTarifPeriode('+tarifId+',\''+nama.replace(/'/g,"\\'")+'\')"><i class="ri-add-line"></i> Simpan</button>';
  html+='</div></div>';
  html+='<div style="margin-top:1rem"><button class="btn btn-outline btn-sm" onclick="hideModal()">Tutup</button></div>';
  $('modal').innerHTML=html;showModal();
  }catch(e){toast('Error: '+e.message);}
}
async function simpanTarifPeriode(tarifId, nama){
  try{
  const mulai=$('tpMulai').value, akhir=$('tpAkhir').value, nom=parseInt($('tpNom').value)||0;
  if(!mulai||!akhir||nom<=0) return toast('Semua field wajib diisi');
  if(mulai>akhir) return toast('Bulan mulai harus <= bulan akhir');
  await api('/api/pembayaran/tarif/'+tarifId+'/periode',{method:'POST',body:JSON.stringify({bulan_mulai:mulai,bulan_akhir:akhir,nominal:nom})});
  toast('Periode ditambahkan');showTarifPeriode(tarifId,nama);
  }catch(e){toast('Error: '+e.message);}
}
async function hapusTarifPeriode(periodeId, tarifId, nama){
  if(!confirm('Hapus periode ini?'))return;
  try{await api('/api/pembayaran/tarif/periode/'+periodeId,{method:'DELETE'});toast('Dihapus');showTarifPeriode(tarifId,nama);}catch(e){toast('Error: '+e.message);}
}

// ── Kategori Periode ──
async function showKategoriPeriode(katId, nama){
  try{
  const periodeList=await api('/api/pembayaran/kategori/'+katId+'/periode');
  const arr=Array.isArray(periodeList)?periodeList:[];
  let rows=arr.map(p=>'<tr><td>'+p.bulan_mulai+'</td><td>'+p.bulan_akhir+'</td><td style="text-align:right;font-weight:700">'+fmtRp(p.nominal)+'</td><td><button class="btn btn-danger btn-sm" onclick="hapusKategoriPeriode('+p.id+','+katId+',\''+nama.replace(/'/g,"\\'")+'\')"><i class="ri-delete-bin-line"></i></button></td></tr>').join('');
  if(!arr.length) rows='<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada periode khusus. Semua bulan pakai nominal default.</td></tr>';
  let html='<h3 style="margin-bottom:.5rem"><i class="ri-calendar-line"></i> Periode Kategori: '+nama+'</h3>';
  html+='<p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Bulan yang tidak tercakup periode menggunakan nominal default. Periode tidak boleh tumpang tindih.</p>';
  html+='<div class="table-wrap"><table><tr><th>Bulan Mulai</th><th>Bulan Akhir</th><th style="text-align:right">Nominal</th><th>Aksi</th></tr>'+rows+'</table></div>';
  html+='<div style="margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--border)">';
  html+='<h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-add-circle-line"></i> Tambah Periode</h4>';
  html+='<div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end">';
  html+='<div class="fg" style="flex:1;min-width:130px"><label>Bulan Mulai</label><input type="month" id="kpMulai"></div>';
  html+='<div class="fg" style="flex:1;min-width:130px"><label>Bulan Akhir</label><input type="month" id="kpAkhir"></div>';
  html+='<div class="fg" style="flex:1;min-width:120px"><label>Nominal (Rp)</label><input type="number" id="kpNom" placeholder="200000"></div>';
  html+='<button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="simpanKategoriPeriode('+katId+',\''+nama.replace(/'/g,"\\'")+'\')"><i class="ri-add-line"></i> Simpan</button>';
  html+='</div></div>';
  html+='<div style="margin-top:1rem"><button class="btn btn-outline btn-sm" onclick="hideModal()">Tutup</button></div>';
  $('modal').innerHTML=html;showModal();
  }catch(e){toast('Error: '+e.message);}
}
async function simpanKategoriPeriode(katId, nama){
  try{
  const mulai=$('kpMulai').value, akhir=$('kpAkhir').value, nom=parseInt($('kpNom').value)||0;
  if(!mulai||!akhir||nom<=0) return toast('Semua field wajib diisi');
  if(mulai>akhir) return toast('Bulan mulai harus <= bulan akhir');
  await api('/api/pembayaran/kategori/'+katId+'/periode',{method:'POST',body:JSON.stringify({bulan_mulai:mulai,bulan_akhir:akhir,nominal:nom})});
  toast('Periode ditambahkan');showKategoriPeriode(katId,nama);
  }catch(e){toast('Error: '+e.message);}
}
async function hapusKategoriPeriode(periodeId, katId, nama){
  if(!confirm('Hapus periode ini?'))return;
  try{await api('/api/pembayaran/kategori/periode/'+periodeId,{method:'DELETE'});toast('Dihapus');showKategoriPeriode(katId,nama);}catch(e){toast('Error: '+e.message);}
}

async function showPotonganPanel(){
  try{
  const [potongan,santri]=await Promise.all([api('/api/pembayaran/potongan'),api('/api/santri')]);
  window._potSantriList=Array.isArray(santri)?santri:[];
  window._potSelected=[];
  const arr=Array.isArray(potongan)?potongan:[];
  const groups={};
  arr.forEach(p=>{if(!groups[p.nama])groups[p.nama]={nominal:p.nominal,items:[]};groups[p.nama].items.push(p);});
  let gh='';
  if(Object.keys(groups).length){
    gh=Object.entries(groups).map(function(e){const nama=e[0],g=e[1];
      return '<div style="margin-bottom:.5rem;background:var(--greenbg);border-radius:12px;padding:.6rem .8rem"><div style="display:flex;justify-content:space-between;margin-bottom:.3rem"><b style="font-size:.82rem">'+nama+'</b><b style="color:var(--green);font-size:.82rem">-'+fmtRp(g.nominal)+'</b></div><div style="display:flex;flex-wrap:wrap;gap:.3rem">'+g.items.map(function(p){return '<span style="display:inline-flex;align-items:center;gap:.2rem;padding:.2rem .5rem;background:#fff;border-radius:6px;font-size:.72rem">'+p.santri_nama+' <i class="ri-close-circle-line" style="cursor:pointer;color:var(--red)" onclick="hapusPotongan('+p.id+')"></i></span>';}).join('')+'</div></div>';
    }).join('');
  }else{gh='<div style="text-align:center;color:var(--t3);padding:1rem">Belum ada potongan</div>';}
  $('pbContent').innerHTML='<div class="card au"><h3><i class="ri-scissors-cut-line"></i> Kelola Potongan</h3><p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Buat jenis potongan, lalu pilih santri penerima.</p>'+gh+'<div style="margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--border)"><h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-add-circle-line"></i> Tambah Potongan</h4><div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end;margin-bottom:.5rem"><div class="fg" style="flex:2;min-width:150px"><label>Jenis Potongan</label><input id="potNama" placeholder="Yatim / Hafidz / Berprestasi"></div><div class="fg" style="flex:1;min-width:120px"><label>Nominal (Rp)</label><input id="potNominal" type="number" placeholder="100000"></div></div><div class="fg" style="position:relative"><label>Cari Santri</label><input id="potSearch" placeholder="Ketik nama..." oninput="searchPotSantri(this.value)" autocomplete="off"><div id="potSR" class="search-results"></div></div><div id="potSel" style="display:flex;flex-wrap:wrap;gap:.3rem;margin-bottom:.6rem"></div><button class="btn btn-primary btn-sm" onclick="simpanPotongan()"><i class="ri-save-line"></i> Simpan</button></div></div>';
  }catch(e){$('pbContent').innerHTML='<div class="card au" style="text-align:center;padding:2rem;color:var(--t3)">Error: '+e.message+'</div>';}
}
function searchPotSantri(q){
  var box=$('potSR');if(!box)return;
  if(!q){box.innerHTML='';return;}
  var r=(window._potSantriList||[]).filter(function(s){return s.nama.toLowerCase().indexOf(q.toLowerCase())>=0&&!(window._potSelected||[]).find(function(x){return x.id===s.id});}).slice(0,8);
  box.innerHTML=r.map(function(s){return '<div onclick="addPotS('+s.id+',\''+s.nama.replace(/'/g,"\\'")+'\')">'+s.nama+' <span style="font-size:.7rem;color:var(--t3)">('+( s.kamar_nama||'-')+')</span></div>';}).join('');
}
function addPotS(id,nama){
  if((window._potSelected||[]).find(function(x){return x.id===id}))return;
  window._potSelected.push({id:id,nama:nama});
  $('potSearch').value='';$('potSR').innerHTML='';renderPotSel();
}
function rmPotS(id){window._potSelected=window._potSelected.filter(function(x){return x.id!==id});renderPotSel();}
function renderPotSel(){
  var el=$('potSel');if(!el)return;
  el.innerHTML=(window._potSelected||[]).map(function(s){return '<span style="display:inline-flex;align-items:center;gap:.3rem;padding:.25rem .6rem;background:var(--pbg);color:var(--p);border-radius:8px;font-size:.75rem;font-weight:600">'+s.nama+' <i class="ri-close-line" style="cursor:pointer" onclick="rmPotS('+s.id+')"></i></span>';}).join('');
}
async function simpanPotongan(){
  try{var ids=(window._potSelected||[]).map(function(s){return s.id});
  var n=$('potNama').value.trim(),v=parseInt($('potNominal').value)||0;
  if(!ids.length)return toast('Pilih santri dulu');
  if(!n||v<=0)return toast('Jenis dan nominal wajib');
  await api('/api/pembayaran/potongan',{method:'POST',body:JSON.stringify({santri_ids:ids,nama:n,nominal:v})});
  toast('Potongan ditambahkan');showPotonganPanel();}catch(e){toast('Error: '+e.message);}
}
async function hapusPotongan(id){if(!confirm('Hapus?'))return;
  try{await api('/api/pembayaran/potongan/'+id,{method:'DELETE'});toast('Dihapus');showPotonganPanel();}catch(e){toast('Error: '+e.message);}
}

async function showBayarPanel(){
  try{
  var santri=await api('/api/santri');
  window._bayarSantriList=Array.isArray(santri)?santri:[];
  $('pbContent').innerHTML='<div class="card au"><h3><i class="ri-hand-coin-line"></i> Input Pembayaran</h3><div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end"><div class="fg" style="flex:2;min-width:200px;position:relative"><label>Santri</label><input id="bayarSearch" placeholder="Ketik nama santri..." oninput="searchBayarS(this.value)" autocomplete="off"><input type="hidden" id="bayarSID"><div id="bayarSR" class="search-results"></div></div><div class="fg" style="flex:1;min-width:130px"><label>Bulan</label><input type="month" id="bayarBulan" value="'+_pembayaranBulan+'"></div></div><div id="bayarInfo" style="display:none;background:var(--bluebg);border:1px solid rgba(59,130,246,.15);border-radius:12px;padding:.7rem .9rem;margin-bottom:.8rem;font-size:.82rem"></div><div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end"><div class="fg" style="flex:1;min-width:130px"><label>Nominal Bayar (Rp)</label><input id="bayarNom" type="number" placeholder="500000"></div><div class="fg" style="flex:1;min-width:120px"><label>Metode</label><select id="bayarMet"><option value="tunai">Tunai</option><option value="transfer">Transfer</option></select></div></div><div class="fg"><label>Keterangan</label><input id="bayarKet" placeholder="Opsional..."></div><button class="btn btn-primary" onclick="simpanBayar()"><i class="ri-save-line"></i> Simpan Pembayaran</button></div>';
  }catch(e){$('pbContent').innerHTML='<div class="card au" style="text-align:center;padding:2rem;color:var(--t3)">Error: '+e.message+'</div>';}
}
function searchBayarS(q){
  var box=$('bayarSR');if(!box)return;
  if(!q){box.innerHTML='';return;}
  var r=(window._bayarSantriList||[]).filter(function(s){return s.nama.toLowerCase().indexOf(q.toLowerCase())>=0;}).slice(0,8);
  box.innerHTML=r.map(function(s){return '<div onclick="pickBayarS('+s.id+',\''+s.nama.replace(/'/g,"\\'")+'\')">'+s.nama+' <span style="font-size:.7rem;color:var(--t3)">('+( s.kamar_nama||'-')+')</span></div>';}).join('');
}
async function pickBayarS(id,nama){
  $('bayarSID').value=id;$('bayarSearch').value=nama;$('bayarSR').innerHTML='';
  try{var bulan=$('bayarBulan').value||_pembayaranBulan;
  var data=await api('/api/pembayaran?bulan='+bulan);
  var s=(Array.isArray(data)?data:[]).find(function(x){return x.santri_id===id});
  var box=$('bayarInfo');
  if(s){box.style.display='block';
    var b=s.status==='LUNAS'?'badge-h':s.status==='KURANG'?'badge-i':s.status==='GRATIS'?'badge-s':'badge-a';
    box.innerHTML='<b><i class="ri-information-line"></i> Info Tagihan - '+nama+'</b><div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:.3rem;margin-top:.4rem;font-size:.78rem"><div>Tarif: <b>'+fmtRp(s.tarif)+'</b></div><div>Potongan: <b style="color:var(--red)">'+( s.potongan?'-'+fmtRp(s.potongan):'-')+'</b></div><div>Tagihan: <b style="color:var(--p)">'+fmtRp(s.tagihan)+'</b></div><div>Dibayar: <b>'+fmtRp(s.total_bayar)+'</b></div><div>Sisa: <b style="color:'+(s.kekurangan>0?'var(--red)':'var(--green)')+'">'+fmtRp(s.kekurangan)+'</b></div><div>Status: <span class="'+b+'">'+s.status+'</span></div></div>';
    if(s.kekurangan>0)$('bayarNom').value=s.kekurangan;
  }else{box.style.display='none';}
  }catch(e){console.error(e);}
}
async function simpanBayar(){
  try{var sid=parseInt($('bayarSID').value);
  if(!sid)return toast('Pilih santri dulu');
  var bulan=$('bayarBulan').value,nom=parseInt($('bayarNom').value)||0;
  if(!bulan||nom<=0)return toast('Bulan dan nominal wajib');
  var r=await api('/api/pembayaran',{method:'POST',body:JSON.stringify({santri_id:sid,bulan:bulan,nominal:nom,metode:$('bayarMet').value,keterangan:$('bayarKet').value})});
  toast('Pembayaran dicatat - Status: '+(r.status||''));
  await pickBayarS(sid,$('bayarSearch').value);
  $('bayarNom').value='';$('bayarKet').value='';
  // Refresh rekap stats
  refreshRekapStats();
  }catch(e){toast('Error: '+e.message);}
}

async function refreshRekapStats(){
  try{
  var rekap=await api('/api/pembayaran/rekap?bulan='+_pembayaranBulan);
  var persen=rekap.persen||0;
  // Update stat cards
  var cards=document.querySelectorAll('.stat-card .sn');
  if(cards[0])cards[0].textContent=fmtRp(rekap.total_tagihan);
  if(cards[1])cards[1].textContent=fmtRp(rekap.total_bayar);
  if(cards[2])cards[2].textContent=persen+'%';
  // Update progress bar
  var lunas=document.querySelector('.stat-card.c-blue ~ .stat-card ~ .stat-card ~ .card b[style*="green"]');
  // Update the Lunas/Kurang/Belum row
  var spans=document.querySelectorAll('.card [style*="justify-content:space-between"] b');
  if(spans[0])spans[0].textContent=rekap.lunas||0;
  if(spans[1])spans[1].textContent=rekap.kurang||0;
  if(spans[2])spans[2].textContent=rekap.belum||0;
  // Update progress bar width
  var bar=document.querySelector('[style*="background:linear-gradient(90deg,#16a34a"]');
  if(bar)bar.style.width=persen+'%';
  }catch(e){console.error('Rekap refresh error:',e);}
}

async function showRiwayatPanel(){
  try{
  var data=await api('/api/pembayaran/riwayat');
  var arr=Array.isArray(data)?data:[];
  $('pbContent').innerHTML='<div class="card au"><h3><i class="ri-history-line"></i> 200 Riwayat Pembayaran Terakhir</h3><div class="table-wrap"><table><tr><th>Tanggal Transaksi</th><th>Santri</th><th>Bulan SPP</th><th style="text-align:right">Nominal</th><th>Metode</th><th>Keterangan</th><th>Aksi</th></tr>'+arr.map(function(p){var time='00:00',date='';if(p.created_at){var d=new Date(p.created_at);time=(d.getHours()<10?'0':'')+d.getHours()+':'+(d.getMinutes()<10?'0':'')+d.getMinutes();date=d.getFullYear()+'-'+((d.getMonth()+1)<10?'0':'')+(d.getMonth()+1)+'-'+(d.getDate()<10?'0':'')+d.getDate();}return '<tr><td style="font-size:.75rem;white-space:nowrap">'+date+' '+time+'</td><td>'+(p.santri_nama||'-')+'</td><td><span class="badge-h">'+(p.bulan||'-')+'</span></td><td style="text-align:right;font-weight:700">'+fmtRp(p.nominal)+'</td><td><span class="badge-s">'+(p.metode||'tunai')+'</span></td><td style="font-size:.78rem;color:var(--t3)">'+(p.keterangan||'-')+'</td><td><button class="btn btn-danger btn-sm" onclick="hapusBayar('+p.id+')"><i class="ri-delete-bin-line"></i></button></td></tr>';}).join('')+(arr.length?'':'<tr><td colspan="7" style="text-align:center;color:var(--t3)">Belum ada pembayaran</td></tr>')+'</table></div></div>';
  }catch(e){$('pbContent').innerHTML='<div class="card au" style="text-align:center;padding:2rem;color:var(--t3)">Error: '+e.message+'</div>';}
}
async function hapusBayar(id){if(!confirm('Hapus?'))return;
  try{await api('/api/pembayaran/'+id,{method:'DELETE'});toast('Dihapus');showRiwayatPanel();refreshRekapStats();}catch(e){toast('Error: '+e.message);}
}

// -- CATATAN BENDAHARA (Uang Masuk/Keluar) --
let _keuTglMulai='',_keuTglAkhir='';
async function loadCatatanBendahara(){
  try{
  var now=new Date(Date.now()+7*3600000);
  if(!_keuTglMulai)_keuTglMulai=now.toISOString().slice(0,8)+'01';
  if(!_keuTglAkhir)_keuTglAkhir=now.toISOString().slice(0,10);
  var qs='tgl_mulai='+_keuTglMulai+'&tgl_akhir='+_keuTglAkhir;
  var saldo={};
  try{saldo=await api('/api/keuangan/saldo?'+qs);}catch(e){saldo={};}
  var data=[];
  try{data=await api('/api/keuangan?'+qs);}catch(e){data=[];}
  var arr=Array.isArray(data)?data:[];

  $('main').innerHTML='<div class="page-header au"><h2><i class="ri-book-3-line"></i> Catatan Bendahara</h2></div>'+
  '<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:.8rem;align-items:end">'+
    '<div class="fg" style="flex:1;min-width:130px;margin:0"><label>Dari</label><input type="date" id="keuTgl1" value="'+_keuTglMulai+'" onchange="_keuTglMulai=this.value;loadCatatanBendahara()"></div>'+
    '<div class="fg" style="flex:1;min-width:130px;margin:0"><label>Sampai</label><input type="date" id="keuTgl2" value="'+_keuTglAkhir+'" onchange="_keuTglAkhir=this.value;loadCatatanBendahara()"></div>'+
    '<button class="btn btn-gold btn-sm" onclick="apiDownload(\'/api/keuangan/export-excel?tgl_mulai=\'+_keuTglMulai+\'&tgl_akhir=\'+_keuTglAkhir,\'Laporan_Keuangan.xlsx\')"><i class="ri-file-excel-2-line"></i> Export</button>'+
  '</div>'+
  '<div class="stat-grid" style="grid-template-columns:repeat(auto-fit,minmax(140px,1fr))">'+
    '<div class="stat-card c-green au"><div class="stat-info"><div class="sn" style="font-size:.9rem">'+fmtRp(saldo.total_spp)+'</div><div class="sl">SPP</div></div><div class="si"><i class="ri-money-dollar-circle-line"></i></div></div>'+
    '<div class="stat-card c-teal au"><div class="stat-info"><div class="sn" style="font-size:.9rem">'+fmtRp(saldo.total_insidental)+'</div><div class="sl">Insidental</div></div><div class="si"><i class="ri-file-list-line"></i></div></div>'+
    '<div class="stat-card c-blue au"><div class="stat-info"><div class="sn" style="font-size:.9rem">'+fmtRp(saldo.total_masuk_lain)+'</div><div class="sl">Masuk Lain</div></div><div class="si"><i class="ri-add-circle-line"></i></div></div>'+
    '<div class="stat-card c-red au"><div class="stat-info"><div class="sn" style="font-size:.9rem">'+fmtRp(saldo.total_keluar)+'</div><div class="sl">Keluar</div></div><div class="si"><i class="ri-subtract-line"></i></div></div>'+
    '<div class="stat-card c-amber au"><div class="stat-info"><div class="sn" style="font-size:1.05rem">'+fmtRp(saldo.saldo)+'</div><div class="sl">SALDO</div></div><div class="si"><i class="ri-wallet-3-line"></i></div></div>'+
  '</div>'+
  '<div class="card au"><h3><i class="ri-add-circle-line"></i> Tambah Catatan</h3>'+
    '<div style="display:flex;gap:.5rem;flex-wrap:wrap;align-items:end">'+
      '<div class="fg" style="flex:1;min-width:120px"><label>Tipe</label><select id="keuTipe"><option value="masuk">Uang Masuk</option><option value="keluar">Uang Keluar</option></select></div>'+
      '<div class="fg" style="flex:1;min-width:130px"><label>Nominal (Rp)</label><input id="keuNom" type="number" placeholder="500000"></div>'+
      '<div class="fg" style="flex:1;min-width:130px"><label>Tanggal</label><input type="date" id="keuTgl" value="'+now.toISOString().slice(0,10)+'"></div>'+
      '<div class="fg" style="flex:2;min-width:200px"><label>Keterangan</label><input id="keuKet" placeholder="Sumber dana / keperluan pengeluaran"></div>'+
      '<button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="simpanCatatanKeu()"><i class="ri-save-line"></i> Simpan</button>'+
    '</div>'+
  '</div>'+
  '<div class="card au"><h3><i class="ri-file-list-3-line"></i> Riwayat Transaksi</h3>'+
    '<div class="table-wrap"><table>'+
      '<tr><th>Tanggal</th><th>Tipe</th><th style="text-align:right">Nominal</th><th>Keterangan</th><th>Aksi</th></tr>'+
      arr.map(function(r){
        var isIn=r.tipe==='masuk';
        var time='00:00';if(r.created_at){var d=new Date(r.created_at);time=(d.getHours()<10?'0':'')+d.getHours()+':'+(d.getMinutes()<10?'0':'')+d.getMinutes();}
        return '<tr><td style="font-size:.78rem;white-space:nowrap">'+(r.tanggal||'').slice(0,10)+' '+time+'</td>'+
          '<td><span class="'+(isIn?'badge-h':'badge-a')+'">'+(isIn?'MASUK':'KELUAR')+'</span></td>'+
          '<td style="text-align:right;font-weight:700;color:'+(isIn?'var(--green)':'var(--red)')+'">'+( isIn?'+':'-')+fmtRp(r.nominal)+'</td>'+
          '<td style="font-size:.78rem">'+( r.keterangan||'-')+'</td>'+
          '<td><button class="btn btn-danger btn-sm" onclick="hapusCatatanKeu('+r.id+')"><i class="ri-delete-bin-line"></i></button></td></tr>';
      }).join('')+
      (arr.length?'':'<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada catatan</td></tr>')+
    '</table></div></div>';
  }catch(e){toast('Error: '+e.message);console.error(e);}
}
async function simpanCatatanKeu(){
  try{var tipe=$('keuTipe').value,nom=parseInt($('keuNom').value)||0,tgl=$('keuTgl').value,ket=$('keuKet').value;
  if(nom<=0)return toast('Nominal wajib diisi');
  if(!tgl)return toast('Tanggal wajib');
  await api('/api/keuangan',{method:'POST',body:JSON.stringify({tipe:tipe,nominal:nom,tanggal:tgl,keterangan:ket})});
  toast('Catatan ditambahkan');loadCatatanBendahara();
  }catch(e){toast('Error: '+e.message);}
}
async function hapusCatatanKeu(id){if(!confirm('Hapus catatan ini?'))return;
  try{await api('/api/keuangan/'+id,{method:'DELETE'});toast('Dihapus');loadCatatanBendahara();}catch(e){toast('Error: '+e.message);}
}

// ══════════════════════════════════════════════════════════
// E-PAKET (Logistik Paket Santri)
// ══════════════════════════════════════════════════════════
let _epkTab='gerbang',_epkSantriList=[],_epkBulkIds=[];

async function loadEPaket(){
  const [stats,santri]=await Promise.all([api('/api/paket/stats'),api('/api/santri')]);
  _epkSantriList=santri;
  const tabGerbang=_epkTab==='gerbang'?'background:var(--p);color:#fff':'';
  const tabAsrama=_epkTab==='asrama'?'background:var(--p);color:#fff':'';
  const tabLacak=_epkTab==='lacak'?'background:var(--p);color:#fff':'';
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-box-3-line"></i> E-Paket</h2></div>
    <div class="stat-grid fade-up" style="grid-template-columns:repeat(4,1fr);margin-bottom:1rem">
      <div class="stat-card c-amber"><div class="stat-info"><div class="sn">${stats.di_gerbang||0}</div><div class="sl">Di Gerbang</div></div><div class="si"><i class="ri-door-open-line"></i></div></div>
      <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${stats.di_asrama||0}</div><div class="sl">Di Asrama</div></div><div class="si"><i class="ri-home-5-line"></i></div></div>
      <div class="stat-card c-green"><div class="stat-info"><div class="sn">${stats.selesai_today||0}</div><div class="sl">Selesai Hari Ini</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
      <div class="stat-card c-red"><div class="stat-info"><div class="sn">${stats.total_today||0}</div><div class="sl">Masuk Hari Ini</div></div><div class="si"><i class="ri-inbox-archive-line"></i></div></div>
    </div>
    <div style="display:flex;gap:.4rem;margin-bottom:1rem" class="fade-up">
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabGerbang}" onclick="_epkTab='gerbang';loadEPaket()"><i class="ri-door-open-line"></i> Pos Gerbang</button>
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabAsrama}" onclick="_epkTab='asrama';loadEPaket()"><i class="ri-home-5-line"></i> Distribusi Asrama</button>
      <button class="btn btn-outline btn-sm" style="border-radius:10px;${tabLacak}" onclick="_epkTab='lacak';loadEPaket()"><i class="ri-search-line"></i> Lacak Paket</button>
    </div>`;
  if(_epkTab==='gerbang') loadEPaketGerbang();
  else if(_epkTab==='asrama') loadEPaketAsrama();
  else loadEPaketLacak();
}

// ── Tab 1: Form Penerimaan (Pos Gerbang) ─────────────────
async function loadEPaketGerbang(){
  const paketList=await api('/api/paket?status=DI_GERBANG');
  $('main').innerHTML+= `
    <div class="card fade-up" style="border-left:4px solid var(--amber)">
      <h3 style="margin-bottom:.8rem"><i class="ri-inbox-archive-line" style="color:var(--amber)"></i> Form Penerimaan Paket</h3>
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
        <div class="fg" style="flex:2;min-width:220px"><label><span style="background:var(--p);color:#fff;width:22px;height:22px;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;font-size:.7rem;margin-right:.3rem">1</span> Cari Nama Santri</label>
          <input type="text" id="epkSearch" placeholder="Ketik nama santri..." oninput="epkSearchSantri(this.value)" autocomplete="off" autofocus>
          <input type="hidden" id="epkSantri">
          <div id="epkSearchResults" class="search-results"></div>
        </div>
        <div class="fg" style="flex:2;min-width:180px"><label><span style="background:var(--p);color:#fff;width:22px;height:22px;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;font-size:.7rem;margin-right:.3rem">2</span> Pengirim / Ekspedisi</label>
          <input id="epkPengirim" placeholder="Ibu Fatimah / Shopee / JNE">
        </div>
        <div class="fg" style="flex:3;min-width:100%"><label><span style="background:var(--p);color:#fff;width:22px;height:22px;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;font-size:.7rem;margin-right:.3rem">3</span> Keterangan (Opsional)</label>
          <input id="epkKeterangan" placeholder="Kardus coklat, jangan dibanting">
        </div>
        <button class="btn btn-primary" onclick="simpanPaketGerbang()" style="margin-bottom:.8rem"><i class="ri-save-line"></i> Simpan Paket (Enter)</button>
      </div>
    </div>
    <div class="card fade-up">
      <h3><i class="ri-door-open-line" style="color:var(--amber)"></i> Paket di Gerbang (${paketList.length})</h3>
      <div class="table-wrap"><table>
        <tr><th style="width:30px"><input type="checkbox" id="epkCheckAll" onchange="epkToggleAll(this.checked)" style="accent-color:var(--p);width:16px;height:16px"></th><th>Santri</th><th>Kamar</th><th>Pengirim</th><th>Ket</th><th>Jam Masuk</th><th>Aksi</th></tr>
        ${paketList.length?paketList.map(p=>`<tr>
          <td><input type="checkbox" class="epkCb" value="${p.id}" onchange="epkUpdateBulk()" style="accent-color:var(--p);width:16px;height:16px"></td>
          <td><strong>${p.santri_nama}</strong></td>
          <td style="font-size:.78rem">${p.kamar_nama||'-'}</td>
          <td style="font-size:.78rem">${p.pengirim||'-'}</td>
          <td style="font-size:.78rem">${p.keterangan||'-'}</td>
          <td style="font-size:.78rem;white-space:nowrap">${p.waktu_gerbang}</td>
          <td><button class="btn btn-danger btn-sm" onclick="hapusPaket(${p.id})"><i class="ri-delete-bin-line"></i></button></td>
        </tr>`).join('')
        :'<tr><td colspan="7" style="text-align:center;color:var(--t3);padding:1.5rem"><i class="ri-inbox-line" style="font-size:2rem;display:block;margin-bottom:.3rem;opacity:.3"></i>Tidak ada paket di gerbang</td></tr>'}
      </table></div>
      ${paketList.length?`<div style="display:flex;gap:.5rem;margin-top:.8rem;align-items:center;flex-wrap:wrap">
        <span id="epkBulkCount" style="font-size:.78rem;color:var(--t3)">0 dipilih</span>
        <div class="fg" style="flex:1;min-width:120px;margin-bottom:0"><input id="epkBulkRak" placeholder="Posisi Rak (A-01)" style="font-size:.82rem"></div>
        <button id="epkBulkBtn" class="btn btn-primary btn-sm" onclick="epkBulkTransfer()" disabled><i class="ri-truck-line"></i> Pindahkan ke Asrama</button>
      </div>`:''}
    </div>`;
  // Auto-focus search
  setTimeout(()=>{const el=$('epkSearch');if(el)el.focus();},100);
}

function epkSearchSantri(q){
  const box=$('epkSearchResults'),hidden=$('epkSantri');
  if(!q||q.length<1){box.innerHTML='';hidden.value='';return;}
  const results=_epkSantriList.filter(s=>s.nama.toLowerCase().includes(q.toLowerCase())).slice(0,8);
  box.innerHTML=results.map(s=>`<div onclick="epkPickSantri(${s.id},'${s.nama.replace(/'/g,"\\'")}')"><strong>${s.nama}</strong> <span style="font-size:.72rem;color:var(--t3)">${s.kamar_nama||''}</span></div>`).join('');
}
function epkPickSantri(id,nama){$('epkSantri').value=id;$('epkSearch').value=nama;$('epkSearchResults').innerHTML='';$('epkPengirim').focus();}

async function simpanPaketGerbang(){
  const sid=parseInt($('epkSantri').value)||0;
  if(!sid)return toast('Pilih santri terlebih dahulu');
  const pengirim=$('epkPengirim').value.trim();
  const ket=$('epkKeterangan').value.trim();
  await api('/api/paket',{method:'POST',body:JSON.stringify({santri_id:sid,pengirim:pengirim,keterangan:ket})});
  toast('📦 Paket tercatat!');
  // Reset form & reload
  _epkTab='gerbang';loadEPaket();
}

function epkToggleAll(checked){
  document.querySelectorAll('.epkCb').forEach(cb=>cb.checked=checked);
  epkUpdateBulk();
}
function epkUpdateBulk(){
  _epkBulkIds=[...document.querySelectorAll('.epkCb:checked')].map(cb=>parseInt(cb.value));
  const cnt=$('epkBulkCount');if(cnt)cnt.textContent=_epkBulkIds.length+' dipilih';
  const btn=$('epkBulkBtn');if(btn)btn.disabled=_epkBulkIds.length===0;
}
async function epkBulkTransfer(){
  if(!_epkBulkIds.length)return toast('Pilih paket terlebih dahulu');
  const rak=$('epkBulkRak')?.value?.trim()||'';
  if(!confirm('Pindahkan '+_epkBulkIds.length+' paket ke asrama?'))return;
  const r=await api('/api/paket/bulk-transfer',{method:'PUT',body:JSON.stringify({ids:_epkBulkIds,posisi_rak:rak})});
  toast(r.message);_epkTab='gerbang';loadEPaket();
}

async function hapusPaket(id){
  if(!confirm('Hapus paket ini?'))return;
  await api('/api/paket/'+id,{method:'DELETE'});toast('Dihapus');loadEPaket();
}

// ── Tab 2: Distribusi Asrama ─────────────────────────────
async function loadEPaketAsrama(){
  const paketList=await api('/api/paket?status=DI_ASRAMA');
  // Group by kamar
  const kamarMap={};
  paketList.forEach(p=>{
    const k=p.kamar_nama||'Tanpa Kamar';
    if(!kamarMap[k])kamarMap[k]=[];
    kamarMap[k].push(p);
  });
  const kamarKeys=Object.keys(kamarMap).sort();
  let filterHtml=`<div class="fg" style="max-width:250px;margin-bottom:.8rem"><input type="text" id="epkAsramaSearch" placeholder="Cari nama santri..." oninput="epkFilterAsrama(this.value)" autocomplete="off"></div>`;
  let tableHtml='';
  if(!paketList.length){
    tableHtml=`<div style="text-align:center;color:var(--t3);padding:2rem"><i class="ri-inbox-line" style="font-size:2.5rem;display:block;margin-bottom:.5rem;opacity:.3"></i>Tidak ada paket menunggu distribusi</div>`;
  } else {
    kamarKeys.forEach(kamar=>{
      const items=kamarMap[kamar];
      tableHtml+=`<div class="card fade-up" style="margin-bottom:.5rem;border-left:4px solid var(--blue)">
        <h3 style="margin-bottom:.5rem"><i class="ri-home-5-line" style="color:var(--blue)"></i> ${kamar} <span style="font-size:.72rem;font-weight:400;color:var(--t3)">(${items.length} paket)</span></h3>
        <div class="table-wrap"><table class="epkAsramaTable">
          <tr><th>Santri</th><th>Pengirim</th><th>Posisi Rak</th><th>Jam Masuk Asrama</th><th style="text-align:center">Aksi</th></tr>
          ${items.map(p=>`<tr data-nama="${p.santri_nama.toLowerCase()}">
            <td><strong>${p.santri_nama}</strong></td>
            <td style="font-size:.78rem">${p.pengirim||'-'}</td>
            <td><span style="display:inline-block;padding:.2rem .5rem;border-radius:6px;font-size:.75rem;font-weight:600;background:#eef2ff;color:#4f46e5">${p.posisi_rak||'-'}</span></td>
            <td style="font-size:.78rem;white-space:nowrap">${p.waktu_asrama}</td>
            <td style="text-align:center">
              <button class="btn btn-primary btn-sm" onclick="epkSerahkan(${p.id},'${p.santri_nama.replace(/'/g,"\\\'")}')" style="background:var(--green)"><i class="ri-hand-heart-line"></i> Serahkan</button>
            </td>
          </tr>`).join('')}
        </table></div>
      </div>`;
    });
  }
  $('main').innerHTML+=`${filterHtml}${tableHtml}`;
}

function epkFilterAsrama(q){
  const rows=document.querySelectorAll('.epkAsramaTable tr[data-nama]');
  rows.forEach(row=>{
    row.style.display=!q||row.dataset.nama.includes(q.toLowerCase())?'':'none';
  });
}

async function epkSerahkan(id,nama){
  if(!confirm('Serahkan paket ke '+nama+'?'))return;
  const r=await api('/api/paket/'+id+'/serahkan',{method:'PUT'});
  toast(r.message);_epkTab='asrama';loadEPaket();
}

// ── Tab 3: Lacak Paket ───────────────────────────────────
let _epkLacakStatus='';
async function loadEPaketLacak(){
  const statusOpts=[{v:'',l:'Semua'},{v:'DI_GERBANG',l:'Di Gerbang'},{v:'DI_ASRAMA',l:'Di Asrama'},{v:'SELESAI',l:'Selesai'}];
  let url='/api/paket';
  if(_epkLacakStatus)url+='?status='+_epkLacakStatus;
  const paketList=await api(url);
  const statusBadge=st=>{
    if(st==='DI_GERBANG')return '<span style="display:inline-block;padding:.2rem .5rem;border-radius:6px;font-size:.7rem;font-weight:600;background:#fef3c7;color:#d97706"><i class="ri-door-open-line"></i> Gerbang</span>';
    if(st==='DI_ASRAMA')return '<span style="display:inline-block;padding:.2rem .5rem;border-radius:6px;font-size:.7rem;font-weight:600;background:#dbeafe;color:#2563eb"><i class="ri-home-5-line"></i> Asrama</span>';
    return '<span style="display:inline-block;padding:.2rem .5rem;border-radius:6px;font-size:.7rem;font-weight:600;background:#dcfce7;color:#16a34a"><i class="ri-checkbox-circle-line"></i> Selesai</span>';
  };
  $('main').innerHTML+=`
    <div class="card fade-up">
      <div style="display:flex;gap:.5rem;margin-bottom:.8rem;align-items:center;flex-wrap:wrap">
        <h3 style="flex:1;min-width:150px"><i class="ri-search-line"></i> Lacak Paket</h3>
        <div style="display:flex;gap:.3rem">
          ${statusOpts.map(o=>`<button class="btn btn-outline btn-sm" style="border-radius:8px;font-size:.72rem;${_epkLacakStatus===o.v?'background:var(--p);color:#fff':''}" onclick="_epkLacakStatus='${o.v}';_epkTab='lacak';loadEPaket()">${o.l}</button>`).join('')}
        </div>
      </div>
      <div class="fg" style="margin-bottom:.8rem"><input type="text" id="epkLacakSearch" placeholder="Cari nama santri atau pengirim..." oninput="epkFilterLacak(this.value)" autocomplete="off"></div>
      <div class="table-wrap"><table id="epkLacakTable">
        <tr><th>Santri</th><th>Kamar</th><th>Pengirim</th><th>Status</th><th>Waktu Gerbang</th><th>Waktu Asrama</th><th>Waktu Diterima</th><th>Aksi</th></tr>
        ${paketList.length?paketList.map(p=>`<tr data-search="${(p.santri_nama+' '+p.pengirim).toLowerCase()}">
          <td><strong>${p.santri_nama}</strong></td>
          <td style="font-size:.78rem">${p.kamar_nama||'-'}</td>
          <td style="font-size:.78rem">${p.pengirim||'-'}</td>
          <td>${statusBadge(p.status)}</td>
          <td style="font-size:.75rem;white-space:nowrap">${p.waktu_gerbang||'-'}</td>
          <td style="font-size:.75rem;white-space:nowrap">${p.waktu_asrama||'-'}</td>
          <td style="font-size:.75rem;white-space:nowrap">${p.waktu_diterima||'-'}</td>
          <td><button class="btn btn-danger btn-sm" onclick="hapusPaket(${p.id})"><i class="ri-delete-bin-line"></i></button></td>
        </tr>`).join('')
        :'<tr><td colspan="8" style="text-align:center;color:var(--t3);padding:1.5rem">Tidak ada data</td></tr>'}
      </table></div>
    </div>`;
}

function epkFilterLacak(q){
  const rows=document.querySelectorAll('#epkLacakTable tr[data-search]');
  rows.forEach(row=>{
    row.style.display=!q||row.dataset.search.includes(q.toLowerCase())?'':'none';
  });
}

// -- AUTO LOGIN --
if(token){
  // Try instant restore from sessionStorage (per-tab, avoids cross-tab swap)
  var cached=sessionStorage.getItem('mt_user');
  if(cached){try{user=JSON.parse(cached);if(user.features&&typeof user.features==='string'){try{user._features=JSON.parse(user.features);}catch(e){user._features=[];}}else{user._features=[];}showApp();}catch(e){user=null;}}
  // Always verify with server (updates user data, catches expired tokens)
  api('/api/me').then(u=>{
    user=u;
    if(user.features&&typeof user.features==='string'){try{user._features=JSON.parse(user.features);}catch(e){user._features=[];}}else{user._features=[];}
    sessionStorage.setItem('mt_user',JSON.stringify(user));
    if(!cached)showApp(); // only call showApp if not already shown from cache
  }).catch(()=>{logout();});
}else{
  const ls=document.getElementById('loadingScreen');if(ls)ls.style.display='none';
  document.getElementById('loginPage').style.display='flex';
}

window.showBulkBayarModal = async function() {
  const allSantri = await api('/api/santri').catch(()=>[]);
  const year = _pembayaranBulan.slice(0,4);
  const months = ['01','02','03','04','05','06','07','08','09','10','11','12'];
  
  let selectedSantri = new Set();
  let selectedMonths = new Set();
  selectedMonths.add(_pembayaranBulan); // default to current view
  
  window._bulkBayarRender = function(q) {
    const filt = q ? allSantri.filter(s=>s.nama.toLowerCase().includes(q.toLowerCase()) || (s.kamar_nama||'').toLowerCase().includes(q.toLowerCase()) || (s.kelas_diniyyah||'').toLowerCase().includes(q.toLowerCase())) : allSantri;
    let rows = '';
    filt.forEach(s => {
      let chk = selectedSantri.has(s.id) ? 'checked' : '';
      let kelas = s.kelas_diniyyah ? '<span style="font-size:.65rem;background:#f3f4f6;padding:.1rem .3rem;border-radius:4px;margin-left:.3rem">'+s.kelas_diniyyah+'</span>' : '';
      rows += `<label style="display:flex;align-items:center;gap:.5rem;padding:.45rem .6rem;border-bottom:1px solid var(--border);cursor:pointer" onmouseenter="this.style.background='rgba(22,163,74,.04)'" onmouseleave="this.style.background='transparent'">
        <input type="checkbox" ${chk} onchange="this.checked?window._bulkBayarSel.add(${s.id}):window._bulkBayarSel.delete(${s.id});document.getElementById('bbCount').textContent=window._bulkBayarSel.size" style="accent-color:var(--green)">
        <span style="flex:1;font-size:.82rem">${s.nama}${kelas}</span>
        <span style="font-size:.7rem;color:var(--t3)">${s.kamar_nama||''}</span>
      </label>`;
    });
    if(!filt.length) rows = '<div style="text-align:center;padding:1.5rem;color:var(--t3)">Tidak ditemukan</div>';
    document.getElementById('bbList').innerHTML = rows;
    document.getElementById('bbInfo').innerHTML = `<span style="font-size:.78rem;color:var(--t3)">${filt.length} santri</span>`;
  };
  
  window._bulkBayarSel = selectedSantri;
  window._bulkBayarMonths = selectedMonths;
  
  let html = `<h3 style="margin-bottom:.8rem"><i class="ri-hand-coin-fill"></i> Pembayaran Massal</h3>
    <p style="font-size:.8rem;color:var(--t3);margin-bottom:.8rem">Santri yang dicentang akan dilunaskan otomatis sesuai tagihan masing-masing untuk bulan-bulan yang dipilih.</p>
    
    <div style="margin-bottom:1rem">
      <label style="font-size:.8rem;font-weight:600;margin-bottom:.3rem;display:block">Pilih Bulan (Tahun ${year}):</label>
      <div style="display:flex;gap:.4rem;flex-wrap:wrap">
        ${months.map(m => {
          let val = `${year}-${m}`;
          let c = selectedMonths.has(val) ? 'checked' : '';
          return `<label style="font-size:.75rem;padding:.2rem .4rem;border:1px solid var(--border);border-radius:4px;cursor:pointer"><input type="checkbox" ${c} onchange="this.checked?window._bulkBayarMonths.add('${val}'):window._bulkBayarMonths.delete('${val}')"> ${val}</label>`;
        }).join('')}
      </div>
    </div>

    <div style="position:relative;margin-bottom:.6rem">
      <i class="ri-search-line" style="position:absolute;left:.7rem;top:50%;transform:translateY(-50%);color:var(--t3)"></i>
      <input type="text" placeholder="Cari nama, kelas, atau kamar..." oninput="window._bulkBayarRender(this.value)" style="width:100%;padding:.5rem .8rem .5rem 2rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem">
    </div>
    
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.4rem">
      <label style="display:flex;align-items:center;gap:.3rem;cursor:pointer;font-size:.78rem;font-weight:600">
        <input type="checkbox" onchange="var cs=document.querySelectorAll('#bbList input[type=checkbox]');cs.forEach(c=>{c.checked=this.checked;let id=parseInt(c.closest('label').querySelector('input').onchange.toString().match(/add\\((\\d+)\\)/)[1]);if(this.checked)window._bulkBayarSel.add(id);else window._bulkBayarSel.clear();});document.getElementById('bbCount').textContent=window._bulkBayarSel.size" style="accent-color:var(--green)"> Pilih Semua
      </label>
      <span id="bbInfo"></span>
    </div>
    
    <div id="bbList" style="max-height:40vh;overflow-y:auto;border:1.5px solid var(--border);border-radius:12px;background:#fff"></div>
    
    <div style="display:flex;gap:.5rem;margin-top:1rem;align-items:center">
      <button class="btn btn-primary" onclick="doBulkBayar()"><i class="ri-save-line"></i> Bayar Lunas (<span id="bbCount">0</span> Santri)</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;
    
  $('modal').innerHTML=html;showModal();
  window._bulkBayarRender('');
};

window.doBulkBayar = async function() {
  if(window._bulkBayarSel.size === 0) return toast('Pilih minimal 1 santri');
  if(window._bulkBayarMonths.size === 0) return toast('Pilih minimal 1 bulan');
  if(!confirm('Anda yakin ingin melunaskan pembayaran untuk santri & bulan yang dipilih?')) return;
  
  const payload = {
    santri_ids: Array.from(window._bulkBayarSel),
    bulans: Array.from(window._bulkBayarMonths)
  };
  
  try {
    await api('/api/pembayaran/bulk', {method: 'POST', body: JSON.stringify(payload)});
    toast('Pembayaran massal berhasil!');
    hideModal();
    loadPembayaran();
  } catch(e) {
    toast('Error: ' + e.message);
  }
};
