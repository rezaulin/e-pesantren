// ═══ TAHFIDZ QUR'AN MODULE ═══
const SURAH_LIST=["Al-Fatihah","Al-Baqarah","Ali 'Imran","An-Nisa'","Al-Ma'idah","Al-An'am","Al-A'raf","Al-Anfal","At-Taubah","Yunus","Hud","Yusuf","Ar-Ra'd","Ibrahim","Al-Hijr","An-Nahl","Al-Isra'","Al-Kahf","Maryam","Ta Ha","Al-Anbiya'","Al-Hajj","Al-Mu'minun","An-Nur","Al-Furqan","Asy-Syu'ara'","An-Naml","Al-Qasas","Al-'Ankabut","Ar-Rum","Luqman","As-Sajdah","Al-Ahzab","Saba'","Fatir","Ya Sin","As-Saffat","Sad","Az-Zumar","Ghafir","Fussilat","Asy-Syura","Az-Zukhruf","Ad-Dukhan","Al-Jasiyah","Al-Ahqaf","Muhammad","Al-Fath","Al-Hujurat","Qaf","Az-Zariyat","At-Tur","An-Najm","Al-Qamar","Ar-Rahman","Al-Waqi'ah","Al-Hadid","Al-Mujadalah","Al-Hasyr","Al-Mumtahanah","As-Saff","Al-Jumu'ah","Al-Munafiqun","At-Tagabun","At-Talaq","At-Tahrim","Al-Mulk","Al-Qalam","Al-Haqqah","Al-Ma'arij","Nuh","Al-Jinn","Al-Muzzammil","Al-Muddassir","Al-Qiyamah","Al-Insan","Al-Mursalat","An-Naba'","An-Nazi'at","'Abasa","At-Takwir","Al-Infitar","Al-Mutaffifin","Al-Insyiqaq","Al-Buruj","At-Tariq","Al-A'la","Al-Gasyiyah","Al-Fajr","Al-Balad","Asy-Syams","Al-Lail","Ad-Duha","Asy-Syarh","At-Tin","Al-'Alaq","Al-Qadr","Al-Bayyinah","Az-Zalzalah","Al-'Adiyat","Al-Qari'ah","At-Takasur","Al-'Asr","Al-Humazah","Al-Fil","Quraisy","Al-Ma'un","Al-Kausar","Al-Kafirun","An-Nasr","Al-Lahab","Al-Ikhlas","Al-Falaq","An-Nas"];

function _getPredikat(r){if(r>=90)return'Mumtaz';if(r>=80)return'Jayyid Jiddan';if(r>=70)return'Jayyid';if(r>=60)return'Maqbul';return"Dho'if";}
function _predColor(p){if(p==='Mumtaz')return'#059669';if(p==='Jayyid Jiddan')return'#0d9488';if(p==='Jayyid')return'#2563eb';if(p==='Maqbul')return'#d97706';return'#dc2626';}

// ═══ HALAQOH (Master) ═══
async function loadTahfidzHalaqoh(){
  const m=$('main');const isAdmin=['admin','superadmin'].includes(user.role);
  const [halaqohList]=await Promise.all([api('/api/halaqoh')]);
  let html=`<div class="card au"><div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
    <h3><i class="ri-book-read-line" style="color:#0d9488"></i> Kelola Halaqoh</h3>
    ${isAdmin?'<button class="btn btn-primary btn-sm" onclick="showCreateHalaqoh()"><i class="ri-add-line"></i> Buat Halaqoh</button>':''}
  </div>`;
  if(!halaqohList.length){html+='<p style="color:var(--t3)">Belum ada halaqoh. Buat halaqoh terlebih dahulu.</p>';}
  halaqohList.forEach(h=>{
    const pct=h.target_ayat>0?'Target: Juz '+h.target_juz+' ('+h.target_ayat+' ayat)':'Belum ada target';
    html+=`<div style="border:1.5px solid var(--border);border-radius:14px;padding:1rem;margin-bottom:.8rem;background:#fff">
      <div style="display:flex;justify-content:space-between;align-items:start">
        <div><div style="font-weight:700;font-size:.92rem">🕌 ${h.nama}</div>
        <div style="font-size:.78rem;color:var(--t3);margin-top:.2rem">Musyrif: ${h.musyrif||'-'} | ${h.member_count} Santri</div>
        <div style="font-size:.75rem;color:#0d9488;margin-top:.2rem">🎯 ${pct}</div>
        ${h.target_deadline?'<div style="font-size:.7rem;color:var(--t3)">📅 Deadline: '+h.target_deadline+'</div>':''}</div>
        <div style="display:flex;gap:.3rem">
          <button class="btn btn-outline btn-sm" onclick="showHalaqohMembers(${h.id},'${h.nama.replace(/'/g,"\\'")}')"><i class="ri-group-line"></i> Anggota</button>
          ${isAdmin?'<button class="btn btn-sm" style="background:#fee2e2;color:#dc2626" onclick="deleteHalaqoh('+h.id+')"><i class="ri-delete-bin-line"></i></button>':''}
        </div>
      </div>
    </div>`;
  });
  html+='</div>';m.innerHTML=html;
}

async function showCreateHalaqoh(){
  const users=await api('/api/users').catch(()=>[]);
  const ustadzList=users.filter(u=>u.role==='ustadz');
  let opts=ustadzList.map(u=>`<option value="${u.username}">${u.nama}</option>`).join('');
  $('modal').innerHTML=`<h3><i class="ri-add-line"></i> Buat Halaqoh Baru</h3>
    <div class="fg"><label>Nama Halaqoh</label><input id="hlqNama" placeholder="Halaqoh 1 - Juz 30"></div>
    <div class="fg"><label>Musyrif</label><input id="hlqMusyrif" placeholder="Nama musyrif"></div>
    <div class="fg"><label>Ustadz (Login)</label><select id="hlqUstadz"><option value="">-- Pilih --</option>${opts}</select></div>
    <div class="fg"><label>Target Juz</label><input id="hlqTargetJuz" placeholder="30 atau 29,30"></div>
    <div class="fg"><label>Target Ayat</label><input id="hlqTargetAyat" type="number" placeholder="564"></div>
    <div class="fg"><label>Deadline</label><input id="hlqDeadline" type="date"></div>
    <button class="btn btn-primary btn-full" onclick="doCreateHalaqoh()">Simpan</button>`;
  showModal();
}

async function doCreateHalaqoh(){
  try{await api('/api/halaqoh',{method:'POST',body:JSON.stringify({nama:$('hlqNama').value,musyrif:$('hlqMusyrif').value,ustadz_username:$('hlqUstadz').value,target_juz:$('hlqTargetJuz').value,target_ayat:parseInt($('hlqTargetAyat').value)||0,target_deadline:$('hlqDeadline').value})});
  hideModal();toast('Halaqoh dibuat');loadTahfidzHalaqoh();}catch(e){toast('Error: '+e.message);}
}

async function deleteHalaqoh(id){if(!confirm('Hapus halaqoh ini beserta semua data?'))return;try{await api('/api/halaqoh/'+id,{method:'DELETE'});toast('Dihapus');loadTahfidzHalaqoh();}catch(e){toast(e.message);}}

async function showHalaqohMembers(hid,nama){
  const members=await api('/api/halaqoh/'+hid+'/members');
  let rows=members.map((m,i)=>`<tr><td>${i+1}</td><td>${m.santri_nama}</td><td>${m.kamar_nama||'-'}</td><td><button class="btn btn-sm" style="background:#fee2e2;color:#dc2626;padding:2px 8px" onclick="removeHlqMember(${hid},${m.santri_id})"><i class="ri-close-line"></i></button></td></tr>`).join('');
  $('modal').innerHTML=`<h3>Anggota: ${nama}</h3>
    <button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="hideModal();showBulkAddSantri({title:'Tambah ke ${nama.replace(/'/g,"\\'")}',apiUrl:'/api/halaqoh/${hid}/members/bulk',existingIds:[${members.map(m=>m.santri_id)}],onDone:()=>loadTahfidzHalaqoh()})">+ Tambah Santri</button>
    <div class="table-wrap"><table><tr><th>#</th><th>Nama</th><th>Kamar</th><th></th></tr>${rows||'<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>'}</table></div>`;
  showModal();
}

async function removeHlqMember(hid,sid){try{await api('/api/halaqoh/'+hid+'/members/'+sid,{method:'DELETE'});toast('Dihapus');showHalaqohMembers(hid,'');}catch(e){toast(e.message);}}

// ═══ ABSENSI & PENILAIAN HARIAN ═══
async function loadTahfidzAbsensi(){
  const m=$('main');const today=new Date().toISOString().slice(0,10);
  const halaqohList=await api('/api/halaqoh');
  let opts=halaqohList.map(h=>`<option value="${h.id}">${h.nama}</option>`).join('');
  const surahOpts=SURAH_LIST.map((s,i)=>`<option value="${s}">${i+1}. ${s}</option>`).join('');

  m.innerHTML=`<div class="card au"><h3><i class="ri-quill-pen-line" style="color:#059669"></i> Absensi & Penilaian Tahfidz</h3>
    <div style="display:flex;gap:.5rem;flex-wrap:wrap;margin:.8rem 0">
      <input type="date" id="thfTgl" value="${today}" style="padding:.4rem .6rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem">
      <select id="thfHalaqoh" onchange="loadTahfidzForm()" style="padding:.4rem .6rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem;flex:1;min-width:150px"><option value="">-- Pilih Halaqoh --</option>${opts}</select>
    </div>
    <div id="thfFormArea"></div>
  </div>`;
  window._surahOpts=surahOpts;
}

async function loadTahfidzForm(){
  const hid=$('thfHalaqoh').value;const tgl=$('thfTgl').value;const area=$('thfFormArea');
  if(!hid){area.innerHTML='';return;}
  const members=await api('/api/halaqoh/'+hid+'/members');
  const existing=await api('/api/tahfidz?halaqoh_id='+hid+'&tanggal='+tgl);
  const exMap={};existing.forEach(e=>{exMap[e.santri_id]=e;});

  let html='';
  members.forEach((mb,idx)=>{
    const ex=exMap[mb.santri_id]||{};const st=ex.status||'H';const isH=st==='H';
    const id='thf_'+mb.santri_id;
    html+=`<div style="border:1.5px solid var(--border);border-radius:14px;padding:.8rem;margin-bottom:.7rem;background:linear-gradient(135deg,#f0fdfa,#f8fafc);overflow:hidden">
      <div style="font-weight:700;font-size:.85rem;margin-bottom:.5rem">${idx+1}. ${mb.santri_nama}</div>
      <div style="display:flex;gap:.6rem;flex-wrap:wrap;margin-bottom:.5rem;font-size:.8rem">
        ${['H','I','S','A'].map(s=>`<label style="display:flex;align-items:center;gap:.2rem;cursor:pointer"><input type="radio" name="${id}_st" value="${s}" ${st===s?'checked':''} onchange="document.getElementById('${id}_detail').style.display=this.value==='H'?'block':'none'"> ${s}</label>`).join('')}
      </div>
      <div id="${id}_detail" style="display:${isH?'block':'none'}">
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:.4rem;margin-bottom:.4rem">
          <select id="${id}_jenis" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem"><option value="ziyadah" ${ex.jenis_setoran==='ziyadah'?'selected':''}>Ziyadah</option><option value="murojaah" ${ex.jenis_setoran==='murojaah'?'selected':''}>Murojaah</option></select>
          <select id="${id}_surah" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem"><option value="">Surah</option>${window._surahOpts}</select>
        </div>
        <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:.4rem;margin-bottom:.4rem">
          <input id="${id}_dari" type="number" placeholder="Ayat dari" value="${ex.ayat_dari||''}" min="1" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem;min-width:0;box-sizing:border-box">
          <input id="${id}_sampai" type="number" placeholder="Ayat sampai" value="${ex.ayat_sampai||''}" min="1" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem;min-width:0;box-sizing:border-box">
          <input id="${id}_juz" type="number" placeholder="Juz" value="${ex.juz||''}" min="1" max="30" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem;min-width:0;box-sizing:border-box">
        </div>
        <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:.4rem;margin-bottom:.4rem">
          <div><div style="font-size:.65rem;color:var(--t3);margin-bottom:2px">Tajwid</div><input id="${id}_tj" type="number" placeholder="0-100" value="${ex.nilai_tajwid||''}" min="0" max="100" oninput="calcThfNilai('${id}')" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem;width:100%"></div>
          <div><div style="font-size:.65rem;color:var(--t3);margin-bottom:2px">Kelancaran</div><input id="${id}_kl" type="number" placeholder="0-100" value="${ex.nilai_kelancaran||''}" min="0" max="100" oninput="calcThfNilai('${id}')" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem;width:100%"></div>
          <div><div style="font-size:.65rem;color:var(--t3);margin-bottom:2px">Makhraj</div><input id="${id}_mk" type="number" placeholder="0-100" value="${ex.nilai_makhorijul||''}" min="0" max="100" oninput="calcThfNilai('${id}')" style="padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem;width:100%"></div>
        </div>
        <div id="${id}_result" style="font-size:.78rem;font-weight:700;color:#0d9488;margin-bottom:.3rem"></div>
        <input id="${id}_cat" placeholder="Catatan..." value="${ex.catatan||''}" style="width:100%;padding:.35rem;border:1px solid var(--border);border-radius:8px;font-size:.78rem">
      </div>
    </div>`;
    // Set surah after render
    setTimeout(()=>{const sel=document.getElementById(id+'_surah');if(sel&&ex.surah)sel.value=ex.surah;calcThfNilai(id);},100);
  });
  html+=`<button class="btn btn-primary btn-full" onclick="saveTahfidzBulk()" style="margin-top:.5rem"><i class="ri-save-line"></i> Simpan Semua</button>`;
  area.innerHTML=html;
  // Store member IDs
  window._thfMembers=members.map(m=>m.santri_id);
}

function calcThfNilai(id){
  const tj=parseInt(document.getElementById(id+'_tj')?.value)||0;
  const kl=parseInt(document.getElementById(id+'_kl')?.value)||0;
  const mk=parseInt(document.getElementById(id+'_mk')?.value)||0;
  const el=document.getElementById(id+'_result');if(!el)return;
  if(tj===0&&kl===0&&mk===0){el.textContent='';return;}
  const rata=Math.round((tj+kl+mk)/3*100)/100;
  const pred=_getPredikat(rata);
  el.innerHTML=`Rata: <span style="color:${_predColor(pred)}">${rata} — ${pred}</span>`;
}

async function saveTahfidzBulk(){
  const hid=$('thfHalaqoh').value;const tgl=$('thfTgl').value;
  if(!hid||!tgl)return toast('Pilih halaqoh dan tanggal');
  const items=(window._thfMembers||[]).map(sid=>{
    const id='thf_'+sid;
    const stEl=document.querySelector(`input[name="${id}_st"]:checked`);
    return{santri_id:sid,status:stEl?stEl.value:'H',jenis_setoran:document.getElementById(id+'_jenis')?.value||'ziyadah',surah:document.getElementById(id+'_surah')?.value||'',ayat_dari:parseInt(document.getElementById(id+'_dari')?.value)||0,ayat_sampai:parseInt(document.getElementById(id+'_sampai')?.value)||0,juz:parseInt(document.getElementById(id+'_juz')?.value)||0,nilai_tajwid:parseInt(document.getElementById(id+'_tj')?.value)||0,nilai_kelancaran:parseInt(document.getElementById(id+'_kl')?.value)||0,nilai_makhorijul:parseInt(document.getElementById(id+'_mk')?.value)||0,catatan:document.getElementById(id+'_cat')?.value||''};
  });
  try{const r=await api('/api/tahfidz/bulk',{method:'POST',body:JSON.stringify({tanggal:tgl,halaqoh_id:parseInt(hid),items})});toast(r.message);}catch(e){toast('Error: '+e.message);}
}

// ═══ REKAP TAHFIDZ ═══
async function loadTahfidzRekap(){
  const m=$('main');const halaqohList=await api('/api/halaqoh');
  let opts=halaqohList.map(h=>`<option value="${h.id}">${h.nama}</option>`).join('');
  const now=new Date();const dari=now.toISOString().slice(0,8)+'01';const sampai=now.toISOString().slice(0,10);
  m.innerHTML=`<div class="card au"><h3><i class="ri-bar-chart-grouped-line" style="color:#0891b2"></i> Rekap Tahfidz</h3>
    <div style="display:flex;gap:.5rem;flex-wrap:wrap;margin:.8rem 0">
      <input type="date" id="thfRDari" value="${dari}"><input type="date" id="thfRSampai" value="${sampai}">
      <select id="thfRHalaqoh"><option value="">Semua</option>${opts}</select>
      <button class="btn btn-primary btn-sm" onclick="doRekapTahfidz()"><i class="ri-search-line"></i> Cari</button>
      <button class="btn btn-outline btn-sm" onclick="apiDownload('/api/tahfidz/export-excel?dari='+$('thfRDari').value+'&sampai='+$('thfRSampai').value+'&halaqoh_id='+($('thfRHalaqoh').value||''),'Rekap_Tahfidz.xlsx')"><i class="ri-download-line"></i> Excel</button>
    </div>
    <div id="thfRekapArea"></div></div>`;
  doRekapTahfidz();
}

async function doRekapTahfidz(){
  const dari=$('thfRDari').value,sampai=$('thfRSampai').value,hid=$('thfRHalaqoh').value;
  let url='/api/tahfidz/rekap?dari='+dari+'&sampai='+sampai;if(hid)url+='&halaqoh_id='+hid;
  const r=await api(url);const data=r.data||[];
  let info=`<div style="font-size:.78rem;color:var(--t3);margin-bottom:.6rem">Sesi: <strong>${r.jumlah_sesi}</strong>`;
  if(r.target_ayat>0)info+=` | Target: Juz ${r.target_juz} (${r.target_ayat} ayat)`;
  if(r.target_deadline)info+=` | Deadline: ${r.target_deadline}`;
  info+='</div>';
  let rows=data.map((d,i)=>{
    const pc=d.progress_pct>0?`<div style="width:60px;height:6px;background:#e5e7eb;border-radius:3px;display:inline-block;vertical-align:middle"><div style="width:${Math.min(100,d.progress_pct)}%;height:100%;background:#0d9488;border-radius:3px"></div></div> ${d.progress_pct}%`:'';
    return`<tr onclick="window._thfProgressSid=${d.santri_id};nav('tahfidz-progress')" style="cursor:pointer"><td>${i+1}</td><td><strong>${d.santri_nama}</strong></td><td style="text-align:center">${d.H}</td><td style="text-align:center">${d.I}</td><td style="text-align:center">${d.S}</td><td style="text-align:center">${d.A}</td><td style="text-align:center">${d.capaian_ayat||0}</td><td>${pc}</td><td style="text-align:center">${d.avg_nilai||0}</td><td style="color:${_predColor(d.predikat)};font-weight:700;font-size:.78rem">${d.predikat||'-'}</td></tr>`;
  }).join('');
  $('thfRekapArea').innerHTML=info+`<div class="table-wrap"><table style="font-size:.8rem"><tr><th>#</th><th>Nama</th><th>H</th><th>I</th><th>S</th><th>A</th><th>Ayat</th><th>Progress</th><th>Avg</th><th>Predikat</th></tr>${rows||'<tr><td colspan="10" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>'}</table></div>`;
}

// ═══ PROGRESS INDIVIDUAL ═══
async function loadTahfidzProgress(){
  const m=$('main');const sid=window._thfProgressSid;
  if(!sid){m.innerHTML='<div class="card"><p>Pilih santri dari rekap terlebih dahulu.</p></div>';return;}
  const data=await api('/api/tahfidz/progress/'+sid);
  let totalAyat=0;data.forEach(d=>{if(d.jenis_setoran==='ziyadah'&&d.status==='H')totalAyat+=Math.max(0,(d.ayat_sampai||0)-(d.ayat_dari||0)+1);});
  const hadir=data.filter(d=>d.status==='H').length;const total=data.length;
  const avgVals=data.filter(d=>d.nilai_rata>0);const avg=avgVals.length?Math.round(avgVals.reduce((s,d)=>s+d.nilai_rata,0)/avgVals.length*100)/100:0;
  const pred=avg>0?_getPredikat(avg):'-';

  let rows=data.map(d=>{
    const color=d.status==='H'?'':'color:var(--t3)';
    return`<tr style="${color}"><td>${d.tanggal?.slice(0,10)||''}</td><td>${d.status}</td><td>${d.jenis_setoran||'-'}</td><td>${d.surah||'-'}</td><td>${d.ayat_dari||''}-${d.ayat_sampai||''}</td><td style="text-align:center">${d.nilai_rata||'-'}</td><td style="color:${_predColor(d.predikat)};font-weight:600">${d.predikat||'-'}</td></tr>`;
  }).join('');

  m.innerHTML=`<div class="card au">
    <button class="btn btn-outline btn-sm" onclick="nav('tahfidz-rekap')" style="margin-bottom:.8rem"><i class="ri-arrow-left-line"></i> Kembali</button>
    <h3><i class="ri-book-read-line" style="color:#0d9488"></i> Progress Tahfidz</h3>
    <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:.6rem;margin:.8rem 0">
      <div style="background:#f0fdfa;border:1px solid #99f6e4;border-radius:12px;padding:.8rem;text-align:center"><div style="font-size:1.3rem;font-weight:800;color:#0d9488">${totalAyat}</div><div style="font-size:.7rem;color:var(--t3)">Total Ayat Ziyadah</div></div>
      <div style="background:#ecfdf5;border:1px solid #a7f3d0;border-radius:12px;padding:.8rem;text-align:center"><div style="font-size:1.3rem;font-weight:800;color:#059669">${total>0?Math.round(hadir/total*100):0}%</div><div style="font-size:.7rem;color:var(--t3)">Kehadiran (${hadir}/${total})</div></div>
      <div style="background:#f0f9ff;border:1px solid #bae6fd;border-radius:12px;padding:.8rem;text-align:center"><div style="font-size:1.3rem;font-weight:800;color:${_predColor(pred)}">${avg}</div><div style="font-size:.7rem;color:var(--t3)">${pred}</div></div>
    </div>
    <h4 style="margin:.8rem 0 .4rem;font-size:.85rem"><i class="ri-history-line"></i> Riwayat Setoran</h4>
    <div class="table-wrap"><table style="font-size:.78rem"><tr><th>Tanggal</th><th>St</th><th>Jenis</th><th>Surah</th><th>Ayat</th><th>Nilai</th><th>Predikat</th></tr>${rows||'<tr><td colspan="7" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>'}</table></div>
  </div>`;
}
