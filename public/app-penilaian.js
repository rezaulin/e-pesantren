// ═══════════════════════════════════════════════════════════
// MADRASAH DINIYYAH — Kelas + Anggota + Mata Pelajaran
// ═══════════════════════════════════════════════════════════
let _dinKelasId=null,_dinKelasNama=null;
async function loadKelasDiniyyah(){
  _dinKelasId=null;
  const list=await api('/api/kelas-diniyyah');
  const isAdm=['admin','superadmin'].includes(user.role);
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-book-2-line"></i> Madrasah Diniyyah</h2>
    ${isAdm?'<button class="btn btn-primary btn-sm" onclick="showAddKelasDin()"><i class="ri-add-line"></i> Tambah Kelas</button>':''}</div>
    <div class="grid">${list.map(k=>`<div class="grid-card fade-up" style="cursor:pointer;position:relative" onclick="openKelasDin(${k.id},'${k.nama.replace(/'/g,"\\'")}')">
      <div class="title"><i class="ri-book-2-line" style="color:var(--p)"></i> ${k.nama}</div>
      <div class="sub">${k.jumlah_santri} santri · ${k.jumlah_mapel} mapel</div>
      <div style="font-size:.7rem;color:var(--t3);margin-top:.3rem">Kelola anggota & mapel <i class="ri-arrow-right-s-line"></i></div>
      ${isAdm?`<div class="card-actions" style="position:absolute;top:.5rem;right:.5rem;display:flex;gap:.3rem">
        <button class="btn btn-outline btn-sm" onclick="event.stopPropagation();showEditKelasDin(${k.id},'${k.nama.replace(/'/g,"\\'")}')">✏️</button>
        <button class="btn btn-danger btn-sm" onclick="event.stopPropagation();delKelasDin(${k.id})">🗑️</button></div>`:''}</div>`).join('')}
    ${!list.length?'<div class="card" style="text-align:center;color:var(--t3)">Belum ada kelas diniyyah</div>':''}
    </div>`;
}
function showAddKelasDin(){
  $('modal').innerHTML=`<h3><i class="ri-book-2-line"></i> Tambah Kelas Diniyyah</h3>
    <div class="fg"><label>Nama Kelas</label><input id="kdNama" placeholder="Contoh: Kelas 1 Ibtidaiyyah"></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem"><button class="btn btn-primary" onclick="saveKelasDin()">Simpan</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveKelasDin(){
  if(!$('kdNama').value)return toast('Nama wajib');
  await api('/api/kelas-diniyyah',{method:'POST',body:JSON.stringify({nama:$('kdNama').value})});
  hideModal();toast('✅ Ditambahkan');loadKelasDiniyyah();
}
function showEditKelasDin(id,nama){
  $('modal').innerHTML=`<h3>✏️ Edit Kelas</h3><div class="fg"><label>Nama</label><input id="kdNama" value="${nama}"></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem"><button class="btn btn-primary" onclick="saveEditKelasDin(${id})">Simpan</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveEditKelasDin(id){await api('/api/kelas-diniyyah/'+id,{method:'PUT',body:JSON.stringify({nama:$('kdNama').value})});hideModal();toast('✅ Diupdate');loadKelasDiniyyah();}
async function delKelasDin(id){if(!confirm('Hapus kelas ini beserta semua data?'))return;await api('/api/kelas-diniyyah/'+id,{method:'DELETE'});toast('Dihapus');loadKelasDiniyyah();}

async function openKelasDin(id,nama){
  _dinKelasId=id;_dinKelasNama=nama;
  const [members,mapel,allSantri,jadwal,users]=await Promise.all([
    api('/api/kelas-diniyyah/'+id+'/members'),
    api('/api/mata-pelajaran?kelas_diniyyah_id='+id),
    api('/api/santri'),
    api('/api/jadwal-diniyyah?kelas_diniyyah_id='+id).catch(()=>[]),
    api('/api/users').catch(()=>[])
  ]);
  const isAdm=['admin','superadmin'].includes(user.role);
  const canInput=['admin','superadmin','ustadz'].includes(user.role);
  window._dinAllSantri=allSantri;
  const ustadzList=users.filter(u=>u.role==='ustadz');
  const hariOpts=['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu','Minggu'].map(h=>`<option value="${h}">${h}</option>`).join('');
  const kelasJadwal=(jadwal||[]);
  $('main').innerHTML=`<button class="back-btn" onclick="loadKelasDiniyyah()"><i class="ri-arrow-left-line"></i> Kembali</button>
    <div class="page-header"><h2><i class="ri-book-2-line"></i> ${nama}</h2><span class="badge-h">${members.length} santri</span></div>

    <div class="card au"><h3><i class="ri-team-line"></i> Anggota (${members.length})</h3>
    ${isAdm?`<button class="btn btn-primary btn-sm" style="margin-bottom:.6rem" onclick="showBulkAddSantri({title:'Tambah Santri ke ${nama.replace(/'/g,"\\'")}',apiUrl:'/api/kelas-diniyyah/${id}/members/bulk',existingIds:[${members.map(m=>m.id).join(',')}],onDone:function(){openKelasDin(${id},'${nama.replace(/'/g,"\\'")}');}})"><i class="ri-user-add-line"></i> Tambah Santri</button>`:``}
    <div class="table-wrap"><table><tr><th>Nama</th><th>Alamat</th>${isAdm?'<th></th>':''}</tr>
      ${members.map(m=>`<tr><td>${m.nama}</td><td>${m.alamat||'-'}</td>
        ${isAdm?`<td><button class="btn btn-danger btn-sm" onclick="removeDinMember(${id},${m.id})"><i class="ri-close-line"></i></button></td>`:''}</tr>`).join('')}
      ${!members.length?'<tr><td colspan="3" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>':''}
    </table></div></div>

    <div class="card au"><h3><i class="ri-calendar-schedule-line"></i> Jadwal Pelajaran</h3>
    ${isAdm?`<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:.8rem;align-items:end">
      <div class="fg" style="flex:1;min-width:120px"><label>Mata Pelajaran</label><input id="jdMapel" placeholder="Fiqh, Nahwu..."></div>
      <div class="fg" style="flex:1;min-width:100px"><label>Ustadz</label>${searchSelect({id:'jdUstadz',items:ustadzList.map(u=>({value:u.username,label:u.nama})),placeholder:'Ketik nama ustadz...'})}</div>
      <div class="fg" style="flex:3;min-width:100%"><label>Hari (centang beberapa)</label>
        <div style="display:flex;gap:.35rem;flex-wrap:wrap">
          ${['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu','Minggu'].map(h=>`<label style="display:flex;align-items:center;gap:.25rem;padding:.35rem .55rem;background:#f8fafc;border:1.5px solid var(--border);border-radius:10px;cursor:pointer;font-size:.76rem;font-weight:500;transition:.2s" onclick="setTimeout(()=>{const c=this.querySelector('input');this.style.background=c.checked?'var(--greenbg)':'#f8fafc';this.style.borderColor=c.checked?'var(--green)':'var(--border)'},10)"><input type="checkbox" name="jdHariCb" value="${h}" style="accent-color:var(--green);width:14px;height:14px"> ${h}</label>`).join('')}
        </div>
      </div>
      <div style="display:flex;gap:.5rem;flex:2;min-width:100%">
        <div class="fg" style="flex:1"><label>Mulai</label><input type="time" id="jdMulai" value="06:00"></div>
        <div class="fg" style="flex:1"><label>Selesai</label><input type="time" id="jdSelesai" value="07:00"></div>
        <button class="btn btn-primary btn-sm" style="align-self:end;margin-bottom:.8rem" onclick="tambahJadwalDin(${id},'${nama.replace(/'/g,"\\'")}')"><i class="ri-add-line"></i> Tambah</button>
      </div>
    </div>`:''}
    <div class="table-wrap"><table><tr><th>Hari</th><th>Pelajaran</th><th>Ustadz</th><th>Jam</th>${isAdm?'<th></th>':''}</tr>
    ${kelasJadwal.sort((a,b)=>{const d=['Senin','Selasa','Rabu','Kamis','Jumat','Sabtu','Minggu'];return d.indexOf(a.hari)-d.indexOf(b.hari)||a.jam_mulai.localeCompare(b.jam_mulai);}).map(j=>`<tr>
      <td>${j.hari}</td><td>${j.mata_pelajaran}</td><td>${j.ustadz_username}</td><td>${j.jam_mulai} - ${j.jam_selesai}</td>
      ${isAdm?`<td><button class="btn btn-danger btn-sm" onclick="hapusJadwalDin(${j.id},${id},'${nama.replace(/'/g,"\\'")}')"><i class="ri-delete-bin-line"></i></button></td>`:''}</tr>`).join('')}
    ${!kelasJadwal.length?'<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada jadwal</td></tr>':''}
    </table></div></div>

    <div class="card au"><h3><i class="ri-book-open-line"></i> Mata Pelajaran Nilai (${mapel.length})</h3>
    ${isAdm?`<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:.8rem;align-items:end">
      <div class="fg" style="flex:2;min-width:150px"><label>Nama Mapel</label><input id="mpNama" placeholder="Fiqh, Nahwu, Akidah..."></div>
      <button class="btn btn-primary btn-sm" style="margin-bottom:.8rem" onclick="saveMapelDin(${id})"><i class="ri-add-line"></i> Tambah</button>
    </div>`:''}
    <div class="table-wrap"><table><tr><th>Nama</th>${isAdm?'<th></th>':''}</tr>
      ${mapel.map(m=>`<tr><td>${m.nama}</td>${isAdm?`<td><button class="btn btn-danger btn-sm" onclick="delMapelDin(${m.id},${id})"><i class="ri-delete-bin-line"></i></button></td>`:''}</tr>`).join('')}
      ${!mapel.length?'<tr><td colspan="2" style="text-align:center;color:var(--t3)">Belum ada mapel</td></tr>':''}
    </table></div>
    ${mapel.length?`<button class="btn btn-gold" style="margin-top:.8rem" onclick="window._inputNilaiKD=${id};window._inputNilaiKDNama='${nama.replace(/'/g,"\\'")}';nav('input-nilai-diniyyah')"><i class="ri-edit-line"></i> Input Nilai Semester</button>`:''}</div>`;
}
function showAddDinMember(kdId){
  const avail=(window._dinAllSantri||[]);
  $('modal').innerHTML=`<h3><i class="ri-user-add-line"></i> Tambah Santri</h3>
    <div class="fg"><label>Cari Santri</label><input id="dinMemSearch" placeholder="Ketik nama..." oninput="searchDinMember(${kdId})"></div>
    <div id="dinMemResults" style="max-height:300px;overflow-y:auto"></div>`;
  showModal();
}
function searchDinMember(kdId){
  const q=$('dinMemSearch').value.toLowerCase();
  const results=(window._dinAllSantri||[]).filter(s=>s.nama.toLowerCase().includes(q)).slice(0,10);
  $('dinMemResults').innerHTML=results.map(s=>`<div style="padding:.5rem;cursor:pointer;border-bottom:1px solid var(--border)" onclick="addDinMember(${kdId},${s.id})">${s.nama} <span style="font-size:.75rem;color:var(--t3)">${s.kamar_nama||''}</span></div>`).join('');
}
function searchDinMemberInline(q,kdId){
  const box=$('dinMbResults');if(!box)return;if(!q){box.innerHTML='';return;}
  const results=(window._dinAllSantri||[]).filter(s=>s.nama.toLowerCase().includes(q.toLowerCase())).slice(0,6);
  box.innerHTML=results.map(s=>`<div onclick="addDinMemberInline(${kdId},${s.id})">${s.nama}</div>`).join('');
}
async function addDinMemberInline(kdId,sid){
  try{await api('/api/kelas-diniyyah/'+kdId+'/members',{method:'POST',body:JSON.stringify({santri_id:sid})});toast('✅ Ditambahkan');openKelasDin(kdId,_dinKelasNama);}catch(e){toast('Error: '+e.message);}
}
async function addDinMember(kdId,sid){await api('/api/kelas-diniyyah/'+kdId+'/members',{method:'POST',body:JSON.stringify({santri_id:sid})});hideModal();toast('✅ Ditambahkan');openKelasDin(kdId,_dinKelasNama);}
async function removeDinMember(kdId,sid){if(!confirm('Hapus dari kelas?'))return;await api('/api/kelas-diniyyah/'+kdId+'/members/'+sid,{method:'DELETE'});toast('Dihapus');openKelasDin(kdId,_dinKelasNama);}
async function tambahJadwalDin(kdId,kdNama){
  const mp=$('jdMapel')?.value?.trim();if(!mp)return toast('Mata pelajaran wajib');
  const hariList=[...document.querySelectorAll('input[name=jdHariCb]:checked')].map(cb=>cb.value);
  if(!hariList.length)return toast('Pilih minimal 1 hari');
  try{await api('/api/jadwal-diniyyah',{method:'POST',body:JSON.stringify({kelas_diniyyah_id:kdId,mata_pelajaran:mp,ustadz_username:$('jdUstadz_val')?$('jdUstadz_val').value:$('jdUstadz').value,hari_list:hariList,jam_mulai:$('jdMulai').value,jam_selesai:$('jdSelesai').value})});
  toast('Jadwal ditambahkan untuk '+hariList.length+' hari');openKelasDin(kdId,kdNama);}catch(e){toast('Error: '+e.message);}
}
async function hapusJadwalDin(id,kdId,kdNama){
  if(!confirm('Hapus jadwal ini?'))return;
  try{await api('/api/jadwal-diniyyah/'+id,{method:'DELETE'});toast('Dihapus');openKelasDin(kdId,kdNama);}catch(e){toast('Error: '+e.message);}
}
function showAddMapelDin(kdId){
  $('modal').innerHTML=`<h3><i class="ri-book-open-line"></i> Tambah Mata Pelajaran</h3>
    <div class="fg"><label>Nama Mapel</label><input id="mpNama" placeholder="Contoh: Fiqh, Nahwu, Akidah"></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem"><button class="btn btn-primary" onclick="saveMapelDin(${kdId})">Simpan</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveMapelDin(kdId){if(!$('mpNama').value)return toast('Nama wajib');await api('/api/mata-pelajaran',{method:'POST',body:JSON.stringify({kelas_diniyyah_id:kdId,nama:$('mpNama').value})});hideModal();toast('✅ Ditambahkan');openKelasDin(kdId,_dinKelasNama);}
async function delMapelDin(mpId,kdId){if(!confirm('Hapus mapel?'))return;await api('/api/mata-pelajaran/'+mpId,{method:'DELETE'});toast('Dihapus');openKelasDin(kdId,_dinKelasNama);}

// ═══════════════════════════════════════════════════════════
// INPUT NILAI DINIYYAH
// ═══════════════════════════════════════════════════════════
async function loadInputNilaiDiniyyah(){
  const [list,settings]=await Promise.all([api('/api/kelas-diniyyah'),api('/api/settings').catch(()=>({}))]);
  const defSmt=settings.active_semester||'2026-1';
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-award-line"></i> Penilaian Madrasah Diniyyah</h2>
    <p style="color:var(--t3);margin-top:.3rem">Pilih kelas, mapel, dan semester untuk input nilai</p></div>
    <div class="card fade-up"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:2;min-width:150px"><label>Kelas</label>
        <select id="nilaiKelas" onchange="updateMapelDiniyyah(this.value)">
          <option value="">- Pilih Kelas -</option>
          ${list.map(k=>`<option value="${k.id}">${k.nama}</option>`).join('')}
        </select>
      </div>
      <div class="fg" style="flex:2;min-width:150px"><label>Mata Pelajaran</label>
        <select id="nilaiMapel"><option value="">- Pilih Kelas Dulu -</option></select>
      </div>
      <div class="fg" style="flex:1;min-width:100px"><label>Semester</label>
        <select id="nilaiSmt" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%">
          <option value="${settings.tahun_ajaran_aktif || '2026/2027'} - ${settings.semester_aktif || 'Ganjil'}">${settings.tahun_ajaran_aktif || '2026/2027'} - ${settings.semester_aktif || 'Ganjil'}</option>
        </select>
      </div>
      <button class="btn btn-primary btn-sm" onclick="loadNilaiTableDiniyyah()"><i class="ri-search-line"></i> Tampilkan</button>
    </div></div>
    <div id="nilaiTableArea"></div>`;
}
async function updateMapelDiniyyah(kdId){
  const el=$('nilaiMapel');
  if(!kdId){el.innerHTML='<option value="">- Pilih Kelas Dulu -</option>';return;}
  el.innerHTML='<option value="">Loading...</option>';
  try{
    const mapel=await api('/api/mata-pelajaran?kelas_diniyyah_id='+kdId);
    el.innerHTML=mapel.length?mapel.map(m=>`<option value="${m.id}">${m.nama}</option>`).join(''):'<option value="">Belum ada mapel</option>';
  }catch(e){el.innerHTML='<option value="">Error</option>';}
}
async function loadNilaiTableDiniyyah(){
  const kdId=$('nilaiKelas').value, mpId=$('nilaiMapel').value, smt=$('nilaiSmt').value;
  if(!kdId||!mpId||!smt)return toast('Pilih kelas, mapel dan semester');
  const [members,existing,settings]=await Promise.all([api('/api/kelas-diniyyah/'+kdId+'/members'),api('/api/nilai-pelajaran?kelas_diniyyah_id='+kdId+'&semester='+smt+'&mata_pelajaran_id='+mpId),api('/api/settings').catch(()=>({}))]);
  const existMap={};existing.forEach(n=>{existMap[n.santri_id]=n;});
  let comps = [];
  try { comps = JSON.parse(settings.komponen_nilai_diniyyah||'[]'); } catch(e){}
  if(!comps.length) comps = [{nama:'Harian',bobot:30},{nama:'UTS',bobot:30},{nama:'UAS',bobot:40}];
  window._nilaiCompsDin = comps;
  let ths = comps.map(c=>`<th style="width:80px" title="Bobot: ${c.bobot}%">${c.nama}</th>`).join('');
  $('nilaiTableArea').innerHTML=`<div class="card fade-up"><h3><i class="ri-table-line"></i> Tabel Nilai</h3>
    <div class="table-wrap"><table><tr><th>Santri</th>${ths}<th style="width:80px">Akhir</th><th style="width:80px">KKM</th></tr>
    ${members.map(m=>{
      const e=existMap[m.id]||{};
      let detail = {};
      try{ detail=JSON.parse(e.nilai_detail||'{}'); }catch(e){}
      let tds = comps.map((c, idx) => {
        let val = detail[c.nama] !== undefined ? detail[c.nama] : '';
        if(val==='' && c.nama.toLowerCase()=='harian') val=e.nilai_harian||'';
        if(val==='' && c.nama.toLowerCase()=='uts') val=e.nilai_uts||'';
        if(val==='' && c.nama.toLowerCase()=='uas') val=e.nilai_uas||'';
        return `<td><input type="number" class="ni" id="nd_${m.id}_${idx}" value="${val}" min="0" max="100" style="width:70px" oninput="calcNADin(${m.id})"></td>`;
      }).join('');
      return `<tr><td>${m.nama}</td>
      ${tds}
      <td><strong id="nf_${m.id}">${e.nilai_akhir||'-'}</strong></td>
      <td><input type="number" class="ni" id="nkkm_${m.id}" value="${e.kkm||75}" min="0" max="100" style="width:70px"></td></tr>`;
    }).join('')}
    ${!members.length?`<tr><td colspan="${comps.length+2}" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>`:''}
    </table></div>
    <button class="btn btn-primary" style="margin-top:1rem" onclick="saveNilaiBulk(${kdId},'diniyyah')"><i class="ri-save-line"></i> Simpan Semua</button></div>`;
  window._nilaiMembers=members;
}
function calcNADin(sid){
  const comps = window._nilaiCompsDin || [];
  let total = 0;
  comps.forEach((c, idx) => {
    const val = parseFloat($(`nd_${sid}_${idx}`)?.value) || 0;
    total += val * (parseFloat(c.bobot)/100);
  });
  const na = Math.round(total * 100) / 100;
  const el = $('nf_'+sid); if(el) el.textContent = na;
}
async function saveNilaiBulk(kdId,tipe){
  const mpId=$('nilaiMapel').value,smt=$('nilaiSmt').value;
  const comps = window._nilaiCompsDin || [];
  const data=(window._nilaiMembers||[]).map(m=>{
    let detail = {};
    let hasVal = false;
    let oldFields = { harian:0, uts:0, uas:0 };
    comps.forEach((c, idx) => {
       const val = parseFloat($(`nd_${m.id}_${idx}`)?.value) || 0;
       detail[c.nama] = val;
       if (val > 0) hasVal = true;
       if (c.nama.toLowerCase()=='harian') oldFields.harian = val;
       if (c.nama.toLowerCase()=='uts') oldFields.uts = val;
       if (c.nama.toLowerCase()=='uas') oldFields.uas = val;
    });
    const na = parseFloat($('nf_'+m.id)?.textContent) || 0;
    const kkm = parseFloat($('nkkm_'+m.id)?.value) || 75;
    return {
      santri_id: m.id,
      nilai_harian: oldFields.harian,
      nilai_uts: oldFields.uts,
      nilai_uas: oldFields.uas,
      nilai_akhir: na,
      kkm: kkm,
      nilai_detail: JSON.stringify(detail)
    };
  }).filter(d=>d.nilai_akhir > 0 || d.nilai_detail.length > 5);
  if(!data.length)return toast('Tidak ada data untuk disimpan');
  const url=tipe==='diniyyah'?'/api/nilai-pelajaran/bulk':'/api/nilai-sekolah/bulk';
  const body=tipe==='diniyyah'?{kelas_diniyyah_id:kdId,mata_pelajaran_id:parseInt(mpId),semester:smt,data}:{kelas_id:kdId,mata_pelajaran_sekolah_id:parseInt(mpId),semester:smt,data};
  const r=await api(url,{method:'POST',body:JSON.stringify(body)});
  toast('✅ '+(r.message||'Tersimpan'));
}

// ═══════════════════════════════════════════════════════════
// INPUT NILAI SEKOLAH
// ═══════════════════════════════════════════════════════════
async function loadInputNilaiSekolah(){
  const [list,settings]=await Promise.all([api('/api/kelas-sekolah'),api('/api/settings').catch(()=>({}))]);
  const defSmt=settings.active_semester||'2026-1';
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-award-line"></i> Penilaian Sekolah Formal</h2>
    <p style="color:var(--t3);margin-top:.3rem">Pilih kelas, mapel, dan semester untuk input nilai</p></div>
    <div class="card fade-up"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:2;min-width:150px"><label>Kelas</label>
        <select id="nilaiKelas" onchange="updateMapelSekolah(this.value)">
          <option value="">- Pilih Kelas -</option>
          ${list.map(k=>`<option value="${k.id}">${k.nama}</option>`).join('')}
        </select>
      </div>
      <div class="fg" style="flex:2;min-width:150px"><label>Mata Pelajaran</label>
        <select id="nilaiMapel"><option value="">- Pilih Kelas Dulu -</option></select>
      </div>
      <div class="fg" style="flex:1;min-width:100px"><label>Semester</label>
        <select id="nilaiSmt" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%">
          <option value="${settings.tahun_ajaran_aktif || '2026/2027'} - ${settings.semester_aktif || 'Ganjil'}">${settings.tahun_ajaran_aktif || '2026/2027'} - ${settings.semester_aktif || 'Ganjil'}</option>
        </select>
      </div>
      <button class="btn btn-primary btn-sm" onclick="loadNilaiTableSekolah()"><i class="ri-search-line"></i> Tampilkan</button>
    </div></div>
    <div id="nilaiTableArea"></div>`;
}
async function updateMapelSekolah(kid){
  const el=$('nilaiMapel');
  if(!kid){el.innerHTML='<option value="">- Pilih Kelas Dulu -</option>';return;}
  el.innerHTML='<option value="">Loading...</option>';
  try{
    const mapel=await api('/api/mata-pelajaran-sekolah?kelas_id='+kid);
    el.innerHTML=mapel.length?mapel.map(m=>`<option value="${m.id}">${m.nama}</option>`).join(''):'<option value="">Belum ada mapel</option>';
  }catch(e){el.innerHTML='<option value="">Error</option>';}
}
async function loadNilaiTableSekolah(){
  const kid=$('nilaiKelas').value, mpId=$('nilaiMapel').value, smt=$('nilaiSmt').value;
  if(!kid||!mpId||!smt)return toast('Pilih kelas, mapel dan semester');
  const [members,existing,settings]=await Promise.all([api('/api/kelas-sekolah/'+kid+'/members'),api('/api/nilai-sekolah?kelas_id='+kid+'&semester='+smt),api('/api/settings').catch(()=>({}))]);
  const existMap={};existing.forEach(n=>{if(n.mata_pelajaran_sekolah_id==mpId)existMap[n.santri_id]=n;});
  let comps = [];
  try { comps = JSON.parse(settings.komponen_nilai_sekolah||'[]'); } catch(e){}
  if(!comps.length) comps = [{nama:'Harian',bobot:30},{nama:'UTS',bobot:30},{nama:'UAS',bobot:40}];
  window._nilaiCompsSekolah = comps;
  let ths = comps.map(c=>`<th style="width:80px" title="Bobot: ${c.bobot}%">${c.nama}</th>`).join('');

  $('nilaiTableArea').innerHTML=`<div class="card fade-up"><h3><i class="ri-table-line"></i> Tabel Nilai</h3>
    <div class="table-wrap"><table><tr><th>Santri</th>${ths}<th style="width:80px">Akhir</th><th style="width:80px">KKM</th></tr>
    ${members.map(m=>{
      const sid=m.santri_id||m.id;const nm=m.santri_nama||m.nama;const e=existMap[sid]||{};
      let detail = {};
      try{ detail=JSON.parse(e.nilai_detail||'{}'); }catch(e){}
      let tds = comps.map((c, idx) => {
        let val = detail[c.nama] !== undefined ? detail[c.nama] : '';
        if(val==='' && c.nama.toLowerCase()=='harian') val=e.nilai_harian||'';
        if(val==='' && c.nama.toLowerCase()=='uts') val=e.nilai_uts||'';
        if(val==='' && c.nama.toLowerCase()=='uas') val=e.nilai_uas||'';
        return `<td><input type="number" class="ni" id="ns_${sid}_${idx}" value="${val}" min="0" max="100" style="width:70px" oninput="calcNASekolah(${sid})"></td>`;
      }).join('');
      return `<tr><td>${nm}</td>
      ${tds}
      <td><strong id="nf_${sid}">${e.nilai_akhir||'-'}</strong></td>
      <td><input type="number" class="ni" id="ns_kkm_${sid}" value="${e.kkm||75}" min="0" max="100" style="width:70px"></td></tr>`;
    }).join('')}
    </table></div>
    <button class="btn btn-primary" style="margin-top:1rem" onclick="saveNilaiBulkSekolah()"><i class="ri-save-line"></i> Simpan Semua</button></div>`;
  window._nilaiMembers=members.map(m=>({id:m.santri_id||m.id}));
}
function calcNASekolah(sid){
  const comps = window._nilaiCompsSekolah || [];
  let total = 0;
  comps.forEach((c, idx) => {
    const val = parseFloat($(`ns_${sid}_${idx}`)?.value) || 0;
    total += val * (parseFloat(c.bobot)/100);
  });
  const na = Math.round(total * 100) / 100;
  const el = $('nf_'+sid); if(el) el.textContent = na;
}
async function saveNilaiBulkSekolah(){
  const kid=$('nilaiKelas').value,mpId=$('nilaiMapel').value,smt=$('nilaiSmt').value;
  const comps = window._nilaiCompsSekolah || [];
  const data=(window._nilaiMembers||[]).map(m=>{
    let detail = {};
    let hasVal = false;
    let oldFields = { harian:0, uts:0, uas:0 };
    comps.forEach((c, idx) => {
       const val = parseFloat($(`ns_${m.id}_${idx}`)?.value) || 0;
       detail[c.nama] = val;
       if (val > 0) hasVal = true;
       if (c.nama.toLowerCase()=='harian') oldFields.harian = val;
       if (c.nama.toLowerCase()=='uts') oldFields.uts = val;
       if (c.nama.toLowerCase()=='uas') oldFields.uas = val;
    });
    const na = parseFloat($('nf_'+m.id)?.textContent) || 0;
    const kkm = parseFloat($('ns_kkm_'+m.id)?.value) || 75;
    return {
      santri_id: m.id,
      nilai_harian: oldFields.harian,
      nilai_uts: oldFields.uts,
      nilai_uas: oldFields.uas,
      nilai_akhir: na,
      kkm: kkm,
      nilai_detail: JSON.stringify(detail)
    };
  }).filter(d=>d.nilai_akhir > 0 || d.nilai_detail.length > 5);
  if(!data.length)return toast('Tidak ada data');
  const r=await api('/api/nilai-sekolah/bulk',{method:'POST',body:JSON.stringify({kelas_id:kid,mata_pelajaran_sekolah_id:parseInt(mpId),semester:smt,data})});
  toast('✅ '+(r.message||'Tersimpan'));
}

// ═══════════════════════════════════════════════════════════
// RAPORT PENILAIAN
// ═══════════════════════════════════════════════════════════
let _rpSantriList=[],_rpData=null,_rpSettings={},_rpKelasSekolah=[],_rpKelasDiniyyah=[];
async function loadRaportPenilaian(kategori = 'semua'){
  window._rpKategori = kategori;
  const [santri,kelasS,kelasD]=await Promise.all([api('/api/santri'),api('/api/kelas-sekolah').catch(()=>[]),api('/api/kelas-diniyyah').catch(()=>[])]);
  _rpSantriList=santri;_rpData=null;_rpKelasSekolah=kelasS;_rpKelasDiniyyah=kelasD;
  const now=new Date();const yr=now.getFullYear();const sem=now.getMonth()<6?'1':'2';
  const bulanNow=now.toISOString().slice(0,7);
  
  let title = 'Raport Penilaian';
  if(kategori==='sekolah') title='Raport Sekolah Formal';
  if(kategori==='diniyyah') title='Raport Madrasah Diniyyah';
  if(kategori==='kegiatan') title='Raport Kegiatan';
  
  let downBtns='';
  if(kategori==='semua') downBtns=`<button class="btn btn-gold btn-sm" onclick="downloadRaportKategori('semua')"><i class="ri-file-list-3-line"></i> Semua Nilai</button>
        <button class="btn btn-outline btn-sm" onclick="showDownloadKelas('sekolah')" style="border-color:#3b82f6;color:#3b82f6"><i class="ri-school-line"></i> Sekolah</button>
        <button class="btn btn-outline btn-sm" onclick="showDownloadKelas('diniyyah')" style="border-color:#8b5cf6;color:#8b5cf6"><i class="ri-book-2-line"></i> Madrasah Diniyyah</button>
        <button class="btn btn-outline btn-sm" onclick="downloadRaportKategori('kegiatan')" style="border-color:#16a34a;color:#16a34a"><i class="ri-star-line"></i> Nilai Kegiatan</button>`;
  else if(kategori==='sekolah') downBtns=`<button class="btn btn-primary btn-sm" onclick="showDownloadKelas('sekolah')"><i class="ri-school-line"></i> Download Raport Sekolah</button>`;
  else if(kategori==='diniyyah') downBtns=`<button class="btn btn-primary btn-sm" onclick="showDownloadKelas('diniyyah')" style="background:#8b5cf6;border-color:#8b5cf6"><i class="ri-book-2-line"></i> Download Raport Diniyyah</button>`;
  else if(kategori==='kegiatan') downBtns=`<button class="btn btn-primary btn-sm" onclick="downloadRaportKategori('kegiatan')" style="background:#16a34a;border-color:#16a34a"><i class="ri-star-line"></i> Download Raport Kegiatan</button>`;

  $('main').innerHTML=`<div class="page-header fade-up">
      <h2><i class="ri-award-line"></i> ${title}</h2>
      ${kategori!=='semua'?`<button class="btn btn-outline btn-sm" onclick="showPengaturanRaport('${kategori}')"><i class="ri-settings-3-line"></i> Pengaturan</button>`:''}
    </div>
    <div class="card fade-up"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:2;min-width:200px;position:relative"><label>Cari Santri</label>
        <input type="text" id="rpSearch" placeholder="Ketik nama santri..." oninput="searchRPSantri(this.value)" autocomplete="off">
        <input type="hidden" id="rpSantriId">
        <div id="rpSearchResults" class="search-results"></div></div>
      ${kategori!=='kegiatan'?`<div class="fg" style="flex:1;min-width:120px"><label>Semester</label>
        <select id="rpSemester" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%">
          <option value="\${window._rpSettings?.tahun_ajaran_aktif || '2026/2027'} - \${window._rpSettings?.semester_aktif || 'Ganjil'}">\${window._rpSettings?.tahun_ajaran_aktif || '2026/2027'} - \${window._rpSettings?.semester_aktif || 'Ganjil'}</option>
        </select>
      </div>`:''}
      ${kategori==='kegiatan'||kategori==='semua'?`<div class="fg" style="flex:1;min-width:140px"><label>Bulan Kegiatan</label><input type="month" id="rpBulanKegiatan" value="${bulanNow}"></div>`:''}
      <button class="btn btn-primary" onclick="loadRaportPenilaianData()"><i class="ri-search-line"></i> Tampilkan</button>
    </div></div>
    <div class="card fade-up" style="background:linear-gradient(135deg,#f0f9ff,#e0f2fe);border:1.5px solid rgba(59,130,246,.15)">
      <h3 style="margin-bottom:.6rem"><i class="ri-download-cloud-2-line"></i> Download Semua Raport (ZIP)</h3>
      <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Download raport penilaian <strong>semua santri</strong> sekaligus.</p>
      <div style="display:flex;gap:.5rem;flex-wrap:wrap">
        ${downBtns}
      </div>
    </div>
    <div id="rpResult"></div>`;
}
function searchRPSantri(q){
  const box=$('rpSearchResults');if(!q){box.innerHTML='';return;}
  const results=_rpSantriList.filter(s=>s.nama.toLowerCase().includes(q.toLowerCase())).slice(0,8);
  box.innerHTML=results.map(s=>`<div onclick="pickRPSantri(${s.id},'${s.nama.replace(/'/g,"\\\\'")}')">${s.nama}</div>`).join('');
}
function pickRPSantri(id,nama){$('rpSantriId').value=id;$('rpSearch').value=nama;$('rpSearchResults').innerHTML='';}
async function loadRaportPenilaianData(){
  const sid=$('rpSantriId').value;
  const smt=$('rpSemester')?$('rpSemester').value:'2026-1';
  const bulanKeg=$('rpBulanKegiatan')?$('rpBulanKegiatan').value:'';
  if(!sid)return toast('Pilih santri dulu');
  if(window._rpKategori!=='kegiatan' && !smt)return toast('Isi semester');
  if(window._rpKategori==='kegiatan' && !bulanKeg)return toast('Isi bulan kegiatan');
  let qp='semester='+smt+(bulanKeg?'&bulan_nilai='+bulanKeg:'');
  if(smt) {
    let yr = parseInt(smt.split('-')[0]);
    let tr = smt.split('-')[1];
    if(yr && tr) {
       if(tr === '1') qp += `&tgl_mulai=${yr}-07-01&tgl_akhir=${yr}-12-31`;
       else qp += `&tgl_mulai=${yr+1}-01-01&tgl_akhir=${yr+1}-06-30`;
    }
  }
  const [settings,data]=await Promise.all([api('/api/settings').catch(()=>({})),api('/api/raport/'+sid+'?'+qp)]);
  _rpData=data;_rpSettings=settings;
  renderRaportPenilaian(window._rpKategori||'semua');
}
function renderRaportPenilaian(filter){
  if(!_rpData)return;
  const data=_rpData,settings=_rpSettings;
  const sid=$('rpSantriId').value,smt=$('rpSemester').value;
  const bulanKeg=$('rpBulanKegiatan')?.value||'';
  const s=data.santri||{},lembaga=settings?.app_name||'Pesantren';
  const npList=data.nilai_pelajaran||[],nsList=data.nilai_sekolah||[];
  const rpd=data.rata_rata_pelajaran||0,rsd=data.rata_rata_sekolah||0;
  const pkd=data.peringkat_kelas_diniyyah||null,pks=data.peringkat_kelas_sekolah||null;
  const nk=data.nilai_kegiatan||[],rrk=data.rata_rata_kegiatan||0,pkList=data.peringkat_kelompok||[];

  const mkNilai=(list,title,rata,peringkat,compsStr)=>{
    if(!list.length)return `<div class="card fade-up"><h3>${title}</h3><p style="color:var(--t3)">Belum ada data nilai</p></div>`;
    let comps=[];
    try{comps=JSON.parse(compsStr||'[]');}catch(e){}
    if(!comps.length)comps=[{nama:'Harian',bobot:30},{nama:'UTS',bobot:30},{nama:'UAS',bobot:40}];
    const ths = comps.map(c=>`<th style="text-align:center">${c.nama}</th>`).join('');

    return `<div class="card fade-up"><h3>${title}</h3>
    <div class="table-wrap"><table><tr><th>Mata Pelajaran</th>${ths}<th style="text-align:center;font-weight:700">Akhir</th></tr>
    ${list.map(n=>{
      let detail={};try{detail=JSON.parse(n.nilai_detail||'{}');}catch(e){}
      let tds=comps.map(c=>{
        let val=detail[c.nama];
        if(val===undefined&&c.nama.toLowerCase()=='harian')val=n.nilai_harian;
        if(val===undefined&&c.nama.toLowerCase()=='uts')val=n.nilai_uts;
        if(val===undefined&&c.nama.toLowerCase()=='uas')val=n.nilai_uas;
        return `<td style="text-align:center">${val!==undefined?val:'-'}</td>`;
      }).join('');
      return `<tr><td>${n.mata_pelajaran}</td>${tds}<td style="text-align:center;font-weight:700">${n.nilai_akhir}</td></tr>`;
    }).join('')}
    <tr style="font-weight:700;background:rgba(59,130,246,.05)"><td>Rata-rata</td><td colspan="${comps.length}"></td><td style="text-align:center">${rata}</td></tr>
    </table></div>
    ${peringkat?`<div style="margin-top:.8rem;padding:.6rem 1rem;background:linear-gradient(135deg,#f0fdf4,#dcfce7);border-radius:10px;border:1px solid rgba(22,163,74,.15)"><i class="ri-trophy-line" style="color:#16a34a"></i> <strong>Peringkat ${peringkat.peringkat}</strong> dari ${peringkat.total_siswa} siswa</div>`:''}</div>`;
  };

  const mkKegiatan=()=>{
    if(!bulanKeg)return `<div class="card fade-up"><h3><i class="ri-star-line"></i> Nilai Kegiatan</h3><p style="color:var(--t3)">Isi "Bulan Kegiatan" untuk melihat nilai kegiatan</p></div>`;
    if(!nk.length)return `<div class="card fade-up"><h3><i class="ri-star-line"></i> Nilai Kegiatan (${bulanKeg})</h3><p style="color:var(--t3)">Belum ada data nilai kegiatan</p></div>`;
    return `<div class="card fade-up"><h3><i class="ri-star-line"></i> Nilai Kegiatan (${bulanKeg})</h3>
    <div class="table-wrap"><table>
      <tr><th>Kegiatan</th><th>Kelompok</th><th style="text-align:center">Nilai</th><th>Catatan</th></tr>
      ${nk.map(n=>`<tr><td>${n.kegiatan}</td><td>${n.kelompok}</td><td style="text-align:center"><strong>${n.nilai}</strong></td><td>${n.catatan||'-'}</td></tr>`).join('')}
      <tr style="font-weight:700;background:rgba(59,130,246,.05)"><td colspan="2">Rata-rata</td><td style="text-align:center">${rrk}</td><td></td></tr>
    </table></div>
    ${pkList.length?'<div style="margin-top:.6rem;display:flex;gap:.5rem;flex-wrap:wrap">'+pkList.map(p=>'<div style="padding:.5rem .8rem;background:linear-gradient(135deg,#f0fdf4,#dcfce7);border-radius:10px;border:1px solid rgba(22,163,74,.15);font-size:.82rem"><i class="ri-trophy-line" style="color:#16a34a"></i> <strong>'+p.kelompok+'</strong>: Peringkat '+p.peringkat+' / '+p.total_anggota+'</div>').join('')+'</div>':''}</div>`;
  };

  const tabs=[
    {key:'semua',label:'Semua',icon:'ri-list-check-2'},
    {key:'sekolah',label:'Sekolah',icon:'ri-school-line'},
    {key:'diniyyah',label:'Madrasah',icon:'ri-book-2-line'},
    {key:'kegiatan',label:'Kegiatan',icon:'ri-star-line'}
  ];
  let tabHtml='';
  if(window._rpKategori==='semua'){
    tabHtml=`<div style="display:flex;gap:.4rem;flex-wrap:wrap;margin:.8rem 0">
      ${tabs.map(t=>`<button class="btn ${t.key===filter?'btn-primary':'btn-outline'} btn-sm" onclick="renderRaportPenilaian('${t.key}')"><i class="${t.icon}"></i> ${t.label}</button>`).join('')}
    </div>`;
  }

  let contentHtml='';
  if(filter==='semua'){
    contentHtml=mkNilai(nsList,'<i class="ri-school-line"></i> Nilai Sekolah',rsd,pks,settings.komponen_nilai_sekolah)+mkNilai(npList,'<i class="ri-book-2-line"></i> Nilai Madrasah Diniyyah',rpd,pkd,settings.komponen_nilai_diniyyah)+mkKegiatan();
  }else if(filter==='sekolah'){
    contentHtml=mkNilai(nsList,'<i class="ri-school-line"></i> Nilai Sekolah',rsd,pks,settings.komponen_nilai_sekolah);
  }else if(filter==='diniyyah'){
    contentHtml=mkNilai(npList,'<i class="ri-book-2-line"></i> Nilai Madrasah Diniyyah',rpd,pkd,settings.komponen_nilai_diniyyah);
  }else if(filter==='kegiatan'){
    contentHtml=mkKegiatan();
  }

  $('rpResult').innerHTML=`
    <div class="card raport-header fade-up"><h2>${lembaga.toUpperCase()}</h2><div class="raport-sub">RAPORT PENILAIAN ${filter!=='kegiatan'?'— SEMESTER '+($('rpSemester')?$('rpSemester').value:''):'— KEGIATAN'}</div></div>
    <div class="card fade-up"><table class="raport-identity">
      <tr><td>Nama</td><td>: ${s.nama||'-'}</td></tr><tr><td>Kamar</td><td>: ${s.kamar_nama||'-'}</td></tr>
      <tr><td>Kelas</td><td>: ${filter==='sekolah'?s.kelas_sekolah||'-':s.kelas_diniyyah||'-'}</td></tr>${filter!=='kegiatan'?`<tr><td>Semester</td><td>: ${$('rpSemester')?$('rpSemester').value:''}</td></tr>`:''}</table></div>
    ${tabHtml}
    ${contentHtml}
    <div class="card fade-up" style="background:linear-gradient(135deg,#f0f9ff,#e0f2fe);border:1.5px solid rgba(59,130,246,.15)">
      <button class="btn btn-gold" onclick="apiDownload('/api/raport-penilaian/${sid}/excel?semester=${smt}','Raport_${(s.nama||'raport').replace(/'/g,'')}.xlsx')"><i class="ri-file-excel-2-line"></i> Download Excel</button>
      <button class="btn btn-primary" onclick="cetakRaportPDF('${filter}')"><i class="ri-printer-line"></i> Cetak Raport (PDF)</button>
    </div>`;
}
function cetakRaportPDF(filter){
  if(!_rpData)return toast('Data belum dimuat');
  const data=_rpData,settings=_rpSettings;
  const smt=$('rpSemester')?$('rpSemester').value:'2026-1';
  const s=data.santri||{};
  
  let logo = settings?.logo_base64||'';
  let headerHtml = '';
  let identitasHtml = '';

  if (filter === 'diniyyah') {
    logo = settings?.logo_diniyyah || logo;
    const lembagaName = settings?.nama_lembaga_diniyyah || settings?.app_name || 'MADRASAH DINIYYAH';
    const alamat = settings?.header_diniyyah_alamat || (settings?.alamat_lembaga ? `${settings.alamat_lembaga} ${settings.nama_kota ? ' - '+settings.nama_kota : ''}` : '');
    
    headerHtml = `
      <div style="display:flex;align-items:center;border-bottom:3px solid #000;padding-bottom:10px;margin-bottom:2px">
        ${logo ? `<img src="${logo}" style="height:90px;margin-right:20px">` : '<div style="width:110px"></div>'}
        <div style="flex:1;text-align:center">
          <h1 style="margin:5px 0;font-size:26px;text-transform:uppercase;letter-spacing:1px">${lembagaName}</h1>
          <h2 style="margin:0;font-size:18px;text-transform:uppercase;font-weight:bold">LAPORAN HASIL BELAJAR SANTRI</h2>
          ${alamat ? `<p style="margin:5px 0 0;font-size:12px">${alamat}</p>` : ''}
        </div>
        <div style="width:110px"></div>
      </div>
      <div style="border-top:1px solid #000;margin-bottom:20px"></div>
    `;

    identitasHtml = `
      <table style="width:100%;margin-bottom:20px;font-size:13px">
        <tr>
          <td style="width:140px">Nama Santri</td><td style="width:10px">:</td><td style="width:40%"><strong>${s.nama||'-'}</strong></td>
          <td style="width:120px">Kelas</td><td style="width:10px">:</td><td>${s.kelas_diniyyah||'-'}</td>
        </tr>
        <tr>
          <td>NIS</td><td>:</td><td>${s.nis||'-'}</td>
          <td>Semester</td><td>:</td><td>${smt.split('-')[1]==='1'?'Ganjil (1)':'Genap (2)'}</td>
        </tr>
        <tr>
          <td>Peringkat</td><td>:</td><td><strong>_PERINGKAT_</strong></td>
          <td>Tahun Pelajaran</td><td>:</td><td>${smt.split('-')[0]}/${parseInt(smt.split('-')[0])+1}</td>
        </tr>
      </table>
    `;
  } else if (filter === 'kegiatan') {
    logo = settings?.logo_kegiatan || logo;
    const lembagaName = settings?.nama_lembaga_kegiatan || settings?.app_name || 'PONDOK PESANTREN';
    const alamat = settings?.header_kegiatan_alamat || (settings?.alamat_lembaga ? `${settings.alamat_lembaga} ${settings.nama_kota ? ' - '+settings.nama_kota : ''}` : '');
    const bulan = $('rpBulanKegiatan')?.value || '';
    
    headerHtml = `
      <div style="display:flex;align-items:center;border-bottom:3px solid #000;padding-bottom:10px;margin-bottom:2px">
        ${logo ? `<img src="${logo}" style="height:90px;margin-right:20px">` : '<div style="width:110px"></div>'}
        <div style="flex:1;text-align:center">
          <h1 style="margin:5px 0;font-size:26px;text-transform:uppercase;letter-spacing:1px">${lembagaName}</h1>
          <h2 style="margin:0;font-size:18px;text-transform:uppercase;font-weight:bold">LAPORAN HASIL BELAJAR KEGIATAN</h2>
          ${alamat ? `<p style="margin:5px 0 0;font-size:12px">${alamat}</p>` : ''}
        </div>
        <div style="width:110px"></div>
      </div>
      <div style="border-top:1px solid #000;margin-bottom:20px"></div>
    `;

    identitasHtml = `
      <table style="width:100%;margin-bottom:20px;font-size:13px">
        <tr>
          <td style="width:140px">Nama Santri</td><td style="width:10px">:</td><td style="width:40%"><strong>${s.nama||'-'}</strong></td>
          <td style="width:120px">Bulan</td><td style="width:10px">:</td><td>${bulan}</td>
        </tr>
        <tr>
          <td>NIS</td><td>:</td><td>${s.nis||'-'}</td>
          <td>Tahun</td><td>:</td><td>${bulan.split('-')[0]||'-'}</td>
        </tr>
      </table>
    `;
  } else {
    // Sekolah
    logo = settings?.logo_sekolah || logo;
    const headerL1 = settings?.header_sekolah_line1 || 'PEMERINTAH DAERAH';
    const headerL2 = settings?.header_sekolah_line2 || 'DINAS PENDIDIKAN';
    const lembagaName = settings?.nama_lembaga_sekolah || settings?.app_name || 'SEKOLAH';
    const alamat = settings?.header_sekolah_alamat || (settings?.alamat_lembaga ? `${settings.alamat_lembaga} ${settings.nama_kota ? ' - '+settings.nama_kota : ''}` : '');

    headerHtml = `
      <div style="display:flex;align-items:center;border-bottom:3px solid #000;padding-bottom:10px;margin-bottom:2px">
        ${logo ? `<img src="${logo}" style="height:90px;margin-right:20px">` : '<div style="width:110px"></div>'}
        <div style="flex:1;text-align:center">
          <h2 style="margin:0;font-size:16px;text-transform:uppercase;font-weight:normal">${headerL1}<br>${headerL2}</h2>
          <h1 style="margin:5px 0;font-size:24px;text-transform:uppercase;letter-spacing:1px">${lembagaName}</h1>
          ${alamat ? `<p style="margin:0;font-size:12px">${alamat}</p>` : ''}
        </div>
        <div style="width:110px"></div>
      </div>
      <div style="border-top:1px solid #000;margin-bottom:20px"></div>
      <div style="text-align:center;margin-bottom:20px">
        <h3 style="margin:0;font-size:16px;text-transform:uppercase;font-weight:bold">LAPORAN HASIL BELAJAR PESERTA DIDIK</h3>
        <p style="margin:5px 0 0;font-size:14px">Semester ${smt.split('-')[1]==='1'?'Ganjil (1)':'Genap (2)'} - Tahun Pelajaran ${smt.split('-')[0]}/${parseInt(smt.split('-')[0])+1}</p>
      </div>
    `;

    identitasHtml = `
      <table style="width:100%;margin-bottom:20px;font-size:13px">
        <tr>
          <td style="width:140px">Nama Peserta Didik</td><td style="width:10px">:</td><td style="width:40%"><strong>${s.nama||'-'}</strong></td>
          <td style="width:120px">Kelas</td><td style="width:10px">:</td><td>${s.kelas_sekolah||'-'}</td>
        </tr>
        <tr>
          <td>NIS / NISN</td><td>:</td><td>${s.nis||'-'} / -</td>
          <td>Semester</td><td>:</td><td>${smt.split('-')[1]==='1'?'Ganjil (1)':'Genap (2)'}</td>
        </tr>
        <tr>
          <td>NPSN Sekolah</td><td>:</td><td>-</td>
          <td>Peringkat</td><td>:</td><td><strong>_PERINGKAT_</strong></td>
        </tr>
      </table>
    `;
  }

  const npList=data.nilai_pelajaran||[],nsList=data.nilai_sekolah||[];
  const rpd=data.rata_rata_pelajaran||0,rsd=data.rata_rata_sekolah||0;
  const pkd=data.peringkat_kelas_diniyyah||null,pks=data.peringkat_kelas_sekolah||null;
  const nk=data.nilai_kegiatan||[],rrk=data.rata_rata_kegiatan||0,pkList=data.peringkat_kelompok||[];
  
  const mkTabel=(list,rata,peringkat)=>{
    if(!list.length)return '';
    
    // Inject peringkat to header
    let peringkatStr = (peringkat && peringkat.total_siswa > 0) ? `${peringkat.peringkat} dari ${peringkat.total_siswa}` : '-';
    identitasHtml = identitasHtml.replace('_PERINGKAT_', peringkatStr);

    let html = `<h3 style="margin:15px 0 10px;font-size:14px;font-weight:bold">A. Nilai Akademik</h3>`;
    html += `<table style="width:100%;border-collapse:collapse;margin-bottom:20px;font-size:13px" border="1">
      <thead>
        <tr>
          <th style="padding:6px;width:40px">No</th>
          <th style="padding:6px;text-align:left">Mata Pelajaran</th>
          <th style="padding:6px;width:60px">KKM</th>
          <th style="padding:6px;width:60px">Nilai</th>
          <th style="padding:6px;width:70px">Predikat</th>
          <th style="padding:6px;text-align:left">Guru Pengampu</th>
        </tr>
      </thead><tbody>`;
    list.forEach((n,i)=>{
      let kkm = n.kkm || 75; // dynamic KKM
      let nilai = n.nilai_akhir;
      let predikat = 'D';
      if(nilai >= 90) predikat = 'A';
      else if(nilai >= 80) predikat = 'B';
      else if(nilai >= kkm) predikat = 'C';
      
      html += `<tr>
        <td style="padding:6px;text-align:center">${i+1}</td>
        <td style="padding:6px 8px">${n.mata_pelajaran}</td>
        <td style="padding:6px;text-align:center">${kkm}</td>
        <td style="padding:6px;text-align:center;font-weight:bold">${nilai}</td>
        <td style="padding:6px;text-align:center">${predikat}</td>
        <td style="padding:6px 8px">-</td>
      </tr>`;
    });
    html += `<tr><td colspan="3" style="padding:8px;font-weight:bold;text-align:center">Rata-rata</td><td style="padding:8px;text-align:center;font-weight:bold">${rata}</td><td colspan="2"></td></tr></tbody></table>`;
    return html;
  };
  
  const mkKeg=(list,rata,peringkatList)=>{
    if(!list.length)return '';
    identitasHtml = identitasHtml.replace('_PERINGKAT_', '-'); // Kegiatan might not have a single rank
    let html = `<h3 style="margin:15px 0 10px;font-size:14px;font-weight:bold">A. Nilai Kegiatan</h3>`;
    html += `<table style="width:100%;border-collapse:collapse;margin-bottom:20px;font-size:13px" border="1">
      <thead><tr><th style="padding:6px;width:40px">No</th><th style="padding:6px;text-align:left">Kegiatan</th><th style="padding:6px;text-align:left">Kelompok</th><th style="padding:6px">Nilai</th><th style="padding:6px;text-align:left">Catatan</th></tr></thead><tbody>`;
    list.forEach((n,i)=>{
      html += `<tr><td style="padding:6px;text-align:center">${i+1}</td><td style="padding:6px 8px">${n.kegiatan}</td><td style="padding:6px 8px">${n.kelompok}</td><td style="padding:6px;text-align:center;font-weight:bold">${n.nilai}</td><td style="padding:6px 8px">${n.catatan||'-'}</td></tr>`;
    });
    html += `<tr><td colspan="3" style="padding:8px;font-weight:bold;text-align:center">Rata-rata</td><td style="padding:8px;text-align:center;font-weight:bold">${rata}</td><td></td></tr></tbody></table>`;
    return html;
  };

  let contentHtml='';
  let rekapAbsen = null;
  if(filter==='semua' || filter==='sekolah'){
    contentHtml=mkTabel(nsList,rsd,pks);
    rekapAbsen = data.rekap_sekolah;
  }else if(filter==='diniyyah'){
    contentHtml=mkTabel(npList,rpd,pkd);
  }else if(filter==='kegiatan'){
    contentHtml=mkKeg(nk,rrk,pkList);
  }
  
  // Cleanup replacing _PERINGKAT_ if it wasn't replaced
  identitasHtml = identitasHtml.replace('_PERINGKAT_', '-');

  // Ketidakhadiran
  if(filter!=='kegiatan') {
    let s_cnt = rekapAbsen ? rekapAbsen.S : 0;
    let i_cnt = rekapAbsen ? rekapAbsen.I : 0;
    let a_cnt = rekapAbsen ? rekapAbsen.A : 0;
    contentHtml += `<h3 style="margin:15px 0 10px;font-size:14px;font-weight:bold">B. Ketidakhadiran</h3>
    <table style="width:300px;border-collapse:collapse;margin-bottom:20px;font-size:13px" border="1">
      <tr><td style="padding:6px 8px">Sakit</td><td style="padding:6px;text-align:center;width:40px">${s_cnt}</td><td style="padding:6px;text-align:center;width:40px">hari</td></tr>
      <tr><td style="padding:6px 8px">Izin</td><td style="padding:6px;text-align:center">${i_cnt}</td><td style="padding:6px;text-align:center">hari</td></tr>
      <tr><td style="padding:6px 8px">Tanpa Keterangan (Alfa)</td><td style="padding:6px;text-align:center">${a_cnt}</td><td style="padding:6px;text-align:center">hari</td></tr>
      <tr><td style="padding:6px 8px">Terlambat</td><td style="padding:6px;text-align:center">0</td><td style="padding:6px;text-align:center">hari</td></tr>
    </table>`;
    
    // Catatan Wali Kelas
    contentHtml += `<h3 style="margin:15px 0 10px;font-size:14px;font-weight:bold">C. Catatan Wali Kelas</h3>
    <table style="width:100%;border-collapse:collapse;margin-bottom:20px;font-size:13px" border="1">
      <tr><td style="padding:15px 8px 30px;vertical-align:top">Terus tingkatkan prestasi belajar.</td></tr>
    </table>`;
  }

  let ttdJabatan = 'Kepala / Pengasuh';
  let ttdNama = '';
  let ttdImg = '';
  if(filter === 'sekolah'){
    ttdJabatan = 'Kepala Sekolah';
    ttdNama = settings.ttd_sekolah_nama || '';
    ttdImg = settings.ttd_sekolah_img || '';
  } else if(filter === 'diniyyah'){
    ttdJabatan = 'Kepala Madrasah Diniyyah';
    ttdNama = settings.ttd_diniyyah_nama || '';
    ttdImg = settings.ttd_diniyyah_img || '';
  } else if(filter === 'kegiatan'){
    ttdJabatan = 'Pengasuh / Pembimbing';
    ttdNama = settings.ttd_kegiatan_nama || '';
    ttdImg = settings.ttd_kegiatan_img || '';
  }

  let today = new Date().toLocaleDateString('id-ID', {day:'numeric', month:'long', year:'numeric'});
  let kota = settings.nama_kota || '.............';

  let ttdHtml = '';
  if (filter === 'kegiatan') {
    // 2 columns
    ttdHtml = `
      <table style="width:100%;margin-top:30px;font-size:13px;text-align:center">
        <tr>
          <td style="width:50%;vertical-align:top">Mengetahui,<br>Orang Tua / Wali<br><br><br><br><br><br><br><strong>( ....................................... )</strong></td>
          <td style="width:50%;vertical-align:top">${kota}, ${today}<br>${ttdJabatan}<br>
          ${ttdImg ? `<img src="${ttdImg}" style="height:70px;display:block;margin:5px auto">` : '<br><br><br><br><br>'}
          <strong>${ttdNama ? ttdNama : '( ....................................... )'}</strong>
          </td>
        </tr>
      </table>
    `;
  } else {
    // 3 columns
    ttdHtml = `
      <table style="width:100%;margin-top:30px;font-size:13px;text-align:center">
        <tr>
          <td style="width:33%;vertical-align:top">Mengetahui,<br>Orang Tua / Wali<br><br><br><br><br><br><br><strong>( ....................................... )</strong></td>
          <td style="width:33%;vertical-align:top"><br>Wali Kelas<br><br><br><br><br><br><br><strong>( ....................................... )</strong></td>
          <td style="width:33%;vertical-align:top">${kota}, ${today}<br>${ttdJabatan}<br>
          ${ttdImg ? `<img src="${ttdImg}" style="height:70px;display:block;margin:5px auto">` : '<br><br><br><br><br>'}
          <strong>${ttdNama ? ttdNama : '( ....................................... )'}</strong>
          </td>
        </tr>
      </table>
    `;
  }

  let win = window.open('','_blank');
  win.document.write(`<html><head><title>Cetak Raport - ${s.nama||''}</title><style>
    body { font-family: 'Times New Roman', Times, serif; color: #000; padding: 20px; line-height:1.4; }
    @media print { body { padding: 0; } @page { margin: 1.5cm; } }
  </style></head><body>${headerHtml}${identitasHtml}${contentHtml}${ttdHtml}
  <script>setTimeout(()=>{window.print();window.close();},500);</script>
  </body></html>`);
  win.document.close();
}
function showDownloadKelas(kategori){
  const smt=$('rpSemester')?.value;
  if(!smt)return toast('Isi semester dulu');
  const isSekolah=kategori==='sekolah';
  const kelasList=isSekolah?_rpKelasSekolah:_rpKelasDiniyyah;
  const title=isSekolah?'Raport Sekolah':'Raport Madrasah Diniyyah';
  const icon=isSekolah?'ri-school-line':'ri-book-2-line';
  const color=isSekolah?'#3b82f6':'#8b5cf6';
  $('modal').innerHTML=`<h3 style="display:flex;align-items:center;gap:.5rem"><i class="${icon}" style="color:${color}"></i> Download ${title}</h3>
    <p style="font-size:.82rem;color:var(--t3);margin-bottom:1rem">Pilih kelas untuk download raport <strong>${title}</strong> semester <strong>${smt}</strong>.</p>
    <div style="display:flex;flex-direction:column;gap:.5rem">
      <button class="btn btn-primary" onclick="doDownloadKelas('${kategori}','')" style="text-align:left;justify-content:flex-start">
        <i class="ri-group-line"></i> Semua Kelas (${kelasList.reduce((a,k)=>a+(k.jumlah_santri||0),0)} santri)
      </button>
      ${kelasList.map(k=>`<button class="btn btn-outline" onclick="doDownloadKelas('${kategori}','${k.id}')" style="text-align:left;justify-content:flex-start;border-color:${color};color:${color}">
        <i class="${icon}"></i> ${k.nama} <span style="margin-left:auto;font-size:.75rem;opacity:.7">${k.jumlah_santri||0} santri</span>
      </button>`).join('')}
    </div>
    <button class="btn btn-outline" onclick="hideModal()" style="margin-top:1rem;width:100%">Batal</button>`;
  showModal();
}
function doDownloadKelas(kategori,kelasId){
  hideModal();
  downloadRaportKategori(kategori,kelasId);
}

// ═══════════════════════════════════════════════════════════
// PENGATURAN RAPORT (DINAMIS)
// ═══════════════════════════════════════════════════════════
window.showPengaturanRaport = async function(kategori){
  const s = await api('/api/settings').catch(()=>({}));
  let title = 'Pengaturan Raport';
  let innerHtml = '';
  
  if(kategori === 'kegiatan'){
    title = 'Pengaturan Raport Kegiatan';
    innerHtml = `<div class="fg"><label>Bulan Kegiatan Aktif (Bawaan)</label>
      <input type="month" id="prActiveMonth" value="${s.active_month||''}">
      <p style="font-size:.7rem;color:var(--t3);margin-top:.3rem">Bulan ini akan otomatis terpilih saat membuka Raport Kegiatan atau Nilai Kegiatan.</p>
    </div>
    <div style="margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)">
      <h4 style="font-size:.85rem;margin-bottom:.5rem"><i class="ri-layout-top-line"></i> Header Kop Raport</h4>
      <div class="fg"><label>Nama Lembaga (Baris 1)</label><input id="prNamaLembaga" value="${s.nama_lembaga_kegiatan||''}" placeholder="Contoh: PONDOK PESANTREN AL-FALAH"></div>
      <div class="fg"><label>Alamat (Baris 2)</label><input id="prAlamatHeader" value="${s.header_kegiatan_alamat||''}" placeholder="Contoh: Jl. Raya No. 1, Desa..."></div>
      <div class="fg"><label>Upload Logo (Kiri)</label>
        ${s.logo_kegiatan?`<img src="${s.logo_kegiatan}" style="max-height:60px;margin-bottom:.5rem;display:block">`:''}
        <input type="file" id="prLogoImg" accept="image/png, image/jpeg">
      </div>
    </div>
    <div style="margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)">
      <h4 style="font-size:.85rem;margin-bottom:.5rem"><i class="ri-pen-nib-line"></i> Tanda Tangan Pejabat (Pengasuh)</h4>
      <div class="fg"><label>Nama Pejabat</label><input id="prTtdNama" value="${s.ttd_kegiatan_nama||''}" placeholder="Nama Lengkap & Gelar"></div>
      <div class="fg"><label>Upload Tanda Tangan (Opsional)</label>
        ${s.ttd_kegiatan_img?`<img src="${s.ttd_kegiatan_img}" style="max-height:80px;margin-bottom:.5rem;display:block">`:''}
        <input type="file" id="prTtdImg" accept="image/png, image/jpeg">
        <p style="font-size:.7rem;color:var(--t3);margin-top:.3rem">Biarkan kosong jika tidak ingin mengubah. Gunakan gambar berlatar putih/transparan.</p>
      </div>
    </div>`;
  } else {
    title = kategori === 'sekolah' ? 'Pengaturan Raport Sekolah' : 'Pengaturan Raport Diniyyah';
    const jsonStr = kategori === 'sekolah' ? s.komponen_nilai_sekolah : s.komponen_nilai_diniyyah;
    let comps = [];
    try{ comps = JSON.parse(jsonStr||'[]'); }catch(e){}
    if(!comps.length) comps = [{nama:'Harian', bobot:50}, {nama:'Ujian', bobot:50}];
    
    innerHtml = `
      <div class="fg"><label>Semester Aktif (Bawaan)</label>
        <input id="prActiveSmt" value="${s.active_semester||''}" placeholder="Contoh: 2026-1">
        <p style="font-size:.7rem;color:var(--t3);margin-top:.3rem">Semester ini akan otomatis terpilih saat membuka aplikasi.</p>
      </div>
      <div style="margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)">
        <h4 style="font-size:.85rem;margin-bottom:.5rem"><i class="ri-layout-top-line"></i> Header Kop Raport</h4>
        ${kategori === 'sekolah' ? `
          <div class="fg"><label>Pemerintah Daerah (Baris 1)</label><input id="prHeaderL1" value="${s.header_sekolah_line1||'PEMERINTAH DAERAH'}" placeholder="Contoh: PEMERINTAH DAERAH"></div>
          <div class="fg"><label>Dinas (Baris 2)</label><input id="prHeaderL2" value="${s.header_sekolah_line2||'DINAS PENDIDIKAN'}" placeholder="Contoh: DINAS PENDIDIKAN PROVINSI/KABUPATEN"></div>
          <div class="fg"><label>Nama Lembaga (Baris 3)</label><input id="prNamaLembaga" value="${s.nama_lembaga_sekolah||''}" placeholder="Contoh: SMP ISLAM AL-FALAH"></div>
          <div class="fg"><label>Alamat (Baris 4)</label><input id="prAlamatHeader" value="${s.header_sekolah_alamat||''}" placeholder="Contoh: Jl. Raya No. 1, Desa..."></div>
          <div class="fg"><label>Upload Logo (Kiri)</label>
            ${s.logo_sekolah?`<img src="${s.logo_sekolah}" style="max-height:60px;margin-bottom:.5rem;display:block">`:''}
            <input type="file" id="prLogoImg" accept="image/png, image/jpeg">
          </div>
        ` : `
          <div class="fg"><label>Nama Lembaga Diniyyah (Baris 1)</label><input id="prNamaLembaga" value="${s.nama_lembaga_diniyyah||''}" placeholder="Contoh: MADRASAH DINIYYAH TAKMILIYAH"></div>
          <p style="font-size:.75rem;color:var(--t3);margin-bottom:.5rem">Baris 2 akan otomatis berisi "LAPORAN HASIL BELAJAR SANTRI"</p>
          <div class="fg"><label>Alamat (Baris 3)</label><input id="prAlamatHeader" value="${s.header_diniyyah_alamat||''}" placeholder="Contoh: Jl. Raya No. 1, Desa..."></div>
          <div class="fg"><label>Upload Logo (Kiri)</label>
            ${s.logo_diniyyah?`<img src="${s.logo_diniyyah}" style="max-height:60px;margin-bottom:.5rem;display:block">`:''}
            <input type="file" id="prLogoImg" accept="image/png, image/jpeg">
          </div>
        `}
      </div>
      <div style="margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)">
        <h4 style="font-size:.85rem;margin-bottom:.5rem"><i class="ri-medal-line"></i> Komposisi Penilaian</h4>
        <p style="font-size:.75rem;color:var(--t3);margin-bottom:1rem">Atur komponen nilai dan bobot persentasenya. Pastikan total bobot = 100%.</p>
        <div id="prKompList" style="display:flex;flex-direction:column;gap:.5rem">
          ${comps.map(c => `
            <div class="pr-komp-row" style="display:flex;gap:.5rem;align-items:center">
              <input type="text" class="pr-komp-nama" value="${c.nama}" placeholder="Nama Komponen (cth: Harian)" style="flex:2">
              <input type="number" class="pr-komp-bobot" value="${c.bobot}" placeholder="Bobot %" style="flex:1" min="1" max="100">
              <button class="btn btn-outline btn-sm" onclick="hapusBarisKomponen(this)" style="color:var(--red);border-color:var(--red);padding:.4rem .6rem"><i class="ri-delete-bin-line"></i></button>
            </div>
          `).join('')}
        </div>
        <button class="btn btn-outline btn-sm" onclick="tambahBarisKomponen()" style="margin-top:.8rem;border-style:dashed"><i class="ri-add-line"></i> Tambah Komponen</button>
      </div>
      <div style="margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)">
        <h4 style="font-size:.85rem;margin-bottom:.5rem"><i class="ri-pen-nib-line"></i> Tanda Tangan Pejabat (${kategori === 'sekolah' ? 'Kepala Sekolah' : 'Kepala Diniyyah'})</h4>
        <div class="fg"><label>Nama Pejabat</label><input id="prTtdNama" value="${kategori==='sekolah'?s.ttd_sekolah_nama||'':s.ttd_diniyyah_nama||''}" placeholder="Nama Lengkap & Gelar"></div>
        <div class="fg"><label>Upload Tanda Tangan (Opsional)</label>
          ${(kategori==='sekolah'?s.ttd_sekolah_img:s.ttd_diniyyah_img)?`<img src="${kategori==='sekolah'?s.ttd_sekolah_img:s.ttd_diniyyah_img}" style="max-height:80px;margin-bottom:.5rem;display:block">`:''}
          <input type="file" id="prTtdImg" accept="image/png, image/jpeg">
          <p style="font-size:.7rem;color:var(--t3);margin-top:.3rem">Biarkan kosong jika tidak ingin mengubah. Gunakan gambar berlatar putih/transparan.</p>
        </div>
      </div>
    `;
  }

  $('main').innerHTML = `
    <div class="page-header fade-up">
      <h2><i class="ri-settings-3-line"></i> ${title}</h2>
    </div>
    <div class="card fade-up">
      ${innerHtml}
      <div style="display:flex;gap:.5rem;margin-top:1.5rem">
        <button class="btn btn-primary" onclick="savePengaturanRaport('${kategori}')"><i class="ri-save-line"></i> Simpan Pengaturan</button>
      </div>
    </div>
  `;
}

function tambahBarisKomponen(){
  const div = document.createElement('div');
  div.className = 'pr-komp-row';
  div.style = 'display:flex;gap:.5rem;align-items:center';
  div.innerHTML = `
    <input type="text" class="pr-komp-nama" placeholder="Nama Komponen (cth: Tugas)" style="flex:2">
    <input type="number" class="pr-komp-bobot" placeholder="Bobot %" style="flex:1" min="1" max="100">
    <button class="btn btn-outline btn-sm" onclick="hapusBarisKomponen(this)" style="color:var(--red);border-color:var(--red);padding:.4rem .6rem"><i class="ri-delete-bin-line"></i></button>
  `;
  $('prKompList').appendChild(div);
}

function hapusBarisKomponen(btn){
  btn.parentElement.remove();
}

window.savePengaturanRaport = async function(kategori){
  const body = {};
  
  if(kategori === 'kegiatan'){
    body.active_month = $('prActiveMonth').value;
    body.ttd_kegiatan_nama = $('prTtdNama').value;
    body.nama_lembaga_kegiatan = $('prNamaLembaga')?.value;
    body.header_kegiatan_alamat = $('prAlamatHeader')?.value;
  } else {
    body.active_semester = $('prActiveSmt').value;
    
    // Parse the dynamic rows
    const rows = document.querySelectorAll('.pr-komp-row');
    let comps = [];
    let totalBobot = 0;
    rows.forEach(row => {
      const nama = row.querySelector('.pr-komp-nama').value.trim();
      const bobot = parseInt(row.querySelector('.pr-komp-bobot').value) || 0;
      if(nama && bobot > 0){
        comps.push({nama, bobot});
        totalBobot += bobot;
      }
    });
    
    if(comps.length === 0){
      return toast('Minimal harus ada 1 komponen nilai!');
    }
    if(totalBobot !== 100){
      if(!confirm(`Total bobot saat ini ${totalBobot}%. Biasanya total harus 100%. Lanjutkan simpan?`)){
        return;
      }
    }
        if(kategori === 'sekolah'){
        body.komponen_nilai_sekolah = JSON.stringify(comps);
        body.ttd_sekolah_nama = $('prTtdNama').value;
        body.header_sekolah_line1 = $('prHeaderL1')?.value;
        body.header_sekolah_line2 = $('prHeaderL2')?.value;
        body.nama_lembaga_sekolah = $('prNamaLembaga')?.value;
        body.header_sekolah_alamat = $('prAlamatHeader')?.value;
      } else {
        body.komponen_nilai_diniyyah = JSON.stringify(comps);
        body.ttd_diniyyah_nama = $('prTtdNama').value;
        body.nama_lembaga_diniyyah = $('prNamaLembaga')?.value;
        body.header_diniyyah_alamat = $('prAlamatHeader')?.value;
      }
    }
    
    // Process logo if uploaded
    const fileLogo = $('prLogoImg')?.files[0];
    if(fileLogo){
      const readerL = new FileReader();
      readerL.onload = function(e){
        if(kategori === 'kegiatan') body.logo_kegiatan = e.target.result;
        else if(kategori === 'sekolah') body.logo_sekolah = e.target.result;
        else body.logo_diniyyah = e.target.result;
        uploadTtd(kategori, body);
      };
      readerL.readAsDataURL(fileLogo);
    } else {
      uploadTtd(kategori, body);
    }
}

async function uploadTtd(kategori, body) {
    
    // Process image if uploaded
    const file = $('prTtdImg')?.files[0];
    if(file){
      const reader = new FileReader();
      reader.onload = async function(e){
        if(kategori === 'kegiatan') body.ttd_kegiatan_img = e.target.result;
        else if(kategori === 'sekolah') body.ttd_sekolah_img = e.target.result;
        else body.ttd_diniyyah_img = e.target.result;
        
        try {
          await api('/api/settings', {method:'PUT', body:JSON.stringify(body)});
          toast('✅ Pengaturan disimpan');
          // Reload settings internally
          window._rpSettings = await api('/api/settings');
        } catch(err) { toast('Gagal menyimpan: '+err.message); }
      };
      reader.readAsDataURL(file);
      return;
    }
    
    try {
      await api('/api/settings', {method:'PUT', body:JSON.stringify(body)});
      toast('Pengaturan raport disimpan');
      window._rpSettings = await api('/api/settings');
      // Refresh the view if it matches the current category
      if(window._rpKategori === kategori){
        loadRaportPenilaian(kategori);
      }
    }catch(e){
      toast('Error: ' + e.message);
    }
}

async function downloadRaportKategori(kategori,kelasId){
  const smt=$('rpSemester')?$('rpSemester').value:'';
  const bulanKeg=$('rpBulanKegiatan')?.value||'';
  if(kategori!=='kegiatan' && !smt)return toast('Isi semester dulu');
  if(kategori==='kegiatan' && !bulanKeg)return toast('Isi bulan kegiatan dulu');
  const labels={semua:'Semua Nilai',sekolah:'Sekolah',diniyyah:'Madrasah Diniyyah',kegiatan:'Nilai Kegiatan'};
  toast('⏳ Mengunduh raport '+labels[kategori]+'...');
  let qp='kategori='+kategori;
  if(smt)qp+='&semester='+smt;
  if(bulanKeg)qp+='&bulan='+bulanKeg;
  if(kelasId)qp+='&kelas_id='+kelasId;
  const fname=kategori==='kegiatan'?'Raport_'+kategori+'_'+bulanKeg+'.zip':'Raport_'+kategori+'_'+smt+'.zip';
  apiDownload('/api/raport-penilaian-all/zip?'+qp,fname);
}

// ═══════════════════════════════════════════════════════════
// ABSENSI DINIYYAH
// ═══════════════════════════════════════════════════════════
let _absDinKelasId=null,_absDinJadwalId=0,_absDinMapel='',_absDinKelasNama='';
async function loadAbsenDiniyyah(){
  _absDinKelasId=null;_absDinJadwalId=0;_absDinMapel='';_absDinKelasNama='';
  const kelas=await api('/api/kelas-diniyyah');
  const today=new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  $('main').innerHTML=`<div class="page-header au"><h2><i class="ri-book-2-line"></i> Absen Diniyyah</h2></div>
    <div class="fg"><label>Tanggal</label><input type="date" id="absDinTgl" value="${today}" onchange="if(_absDinKelasId) openAbsenDinKelas(_absDinKelasId, window._absDinKelasNama)"></div>
    <div class="abs-tabs au" id="absDinTabs">
      ${kelas.map(k=>`<div class="abs-tab" id="tab-din-kelas-${k.id}" onclick="openAbsenDinKelas(${k.id},'${k.nama.replace(/'/g,"\\'")}')">${k.nama}</div>`).join('')||'<p style="color:var(--t3)">Belum ada kelas</p>'}
    </div>
    <div id="absDinJadwalContainer" class="au"></div>
    <div id="absDinContent">
      <div class="card au" style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih kelas diniyyah di atas</div>
    </div>`;
}

async function openAbsenDinKelas(kelasId,kelasNama){
  _absDinKelasId=kelasId;
  window._absDinKelasNama=kelasNama;
  _absDinJadwalId=0;
  _absDinMapel='';
  
  document.querySelectorAll('#absDinTabs .abs-tab').forEach(el=>el.classList.remove('active'));
  const tab=$('tab-din-kelas-'+kelasId);
  if(tab) tab.classList.add('active');
  
  $('absDinJadwalContainer').innerHTML='<p style="text-align:center;color:var(--t3);font-size:.8rem;padding:.5rem"><i class="ri-loader-4-line ri-spin"></i> Memuat jadwal...</p>';
  $('absDinContent').innerHTML='';
  
  var tanggal=$('absDinTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  var hariMap={0:'Minggu',1:'Senin',2:'Selasa',3:'Rabu',4:'Kamis',5:'Jumat',6:'Sabtu'};
  var hariIni=hariMap[new Date().getDay()];
  var jadwal=await api('/api/jadwal-diniyyah?kelas_diniyyah_id='+kelasId).catch(function(){return[];});
  var todayJ=(jadwal||[]).filter(function(j){return j.hari===hariIni;}).sort(function(a,b){return a.jam_mulai.localeCompare(b.jam_mulai);});
  
  let html=`<div style="font-size:.75rem;color:var(--t3);margin-bottom:.4rem;padding-left:.2rem;font-weight:600">Pilih Mata Pelajaran (${hariIni}):</div>`;
  html += `<div class="abs-tabs" id="absDinJadwalTabs" style="margin-bottom:1rem">`;
  if(todayJ.length){
    todayJ.forEach(j=>{
      html += `<div class="abs-tab" id="tab-din-jadwal-${j.id}" onclick="absDinPilihJadwal(${kelasId},'${kelasNama.replace(/'/g,"\\'")}',${j.id},'${j.mata_pelajaran.replace(/'/g,"\\'")}')" style="display:flex;flex-direction:column;gap:.1rem;align-items:center;padding:.3rem .8rem">
        <span style="font-size:.8rem">${j.mata_pelajaran}</span>
        <span style="font-size:.65rem;font-weight:500;opacity:.8">${j.jam_mulai}-${j.jam_selesai}</span>
      </div>`;
    });
  }
  html += `<div class="abs-tab" id="tab-din-jadwal-0" onclick="absDinPilihJadwal(${kelasId},'${kelasNama.replace(/'/g,"\\'")}',0,'Manual')" style="display:flex;flex-direction:column;gap:.1rem;align-items:center;background:var(--pbg);color:var(--p);border-color:var(--pbg);padding:.3rem .8rem">
      <span style="font-size:.8rem">Absen Manual</span>
      <span style="font-size:.65rem;font-weight:500;opacity:.8">Tanpa Jadwal</span>
    </div>`;
  html += `</div>`;
  
  $('absDinJadwalContainer').innerHTML = html;
  $('absDinContent').innerHTML = '<div class="card au" style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih mata pelajaran di atas</div>';
}

async function absDinPilihJadwal(kelasId,kelasNama,jadwalId,mapel){
  _absDinKelasId=kelasId;_absDinJadwalId=jadwalId;_absDinMapel=mapel;
  
  document.querySelectorAll('#absDinJadwalTabs .abs-tab').forEach(el=>el.classList.remove('active'));
  const tab=$('tab-din-jadwal-'+jadwalId);
  if(tab) tab.classList.add('active');
  
  $('absDinContent').innerHTML='<p style="text-align:center;color:var(--t3);padding:1rem"><i class="ri-loader-4-line ri-spin"></i> Memuat data...</p>';
  
  var tanggal=$('absDinTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  var [members,existing]=await Promise.all([
    api('/api/kelas-diniyyah/'+kelasId+'/members'),
    api('/api/absen-diniyyah?tanggal='+tanggal+'&kelas_diniyyah_id='+kelasId+'&jadwal_diniyyah_id='+jadwalId)
  ]);
  var sudah=existing.length>0;
  if(tab){if(sudah)tab.classList.add('done');else tab.classList.remove('done');}
  
  var existMap={};existing.forEach(function(a){existMap[a.santri_id]=a.status;});
  $('absDinContent').innerHTML=`
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.8rem" class="au">
      <h3 style="margin:0;font-size:1rem;color:var(--text)"><i class="ri-book-2-line"></i> ${kelasNama} — ${mapel}</h3>
      <span style="font-size:.75rem;color:var(--t3)">${members.length} santri</span>
    </div>
    ${sudah?'<div class="info-box success au" style="margin-bottom:.8rem"><i class="ri-checkbox-circle-line"></i> Sudah diabsen diniyyah untuk jadwal ini</div>':''}
    <div class="card au" style="padding:0;overflow:hidden;border-radius:16px"><div class="table-wrap" style="border:none;box-shadow:none;border-radius:0"><table style="table-layout:fixed">
      <tr><th style="width:auto">Nama Santri</th><th style="text-align:center;width:44px">H</th><th style="text-align:center;width:44px">I</th><th style="text-align:center;width:44px">S</th><th style="text-align:center;width:44px">A</th></tr>
      ${members.map(m=>{const st=sudah?(existMap[m.id]||'H'):'H';const dis=sudah?' disabled':'';
        return `<tr><td style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${m.nama}</td>
        <td style="text-align:center"><span class="abs-toggle abs-h${st==='H'?' active':''}${dis}" onclick="setAbsDin(${m.id},'H',this)">H</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-i${st==='I'?' active':''}${dis}" onclick="setAbsDin(${m.id},'I',this)">I</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-s${st==='S'?' active':''}${dis}" onclick="setAbsDin(${m.id},'S',this)">S</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-a${st==='A'?' active':''}${dis}" onclick="setAbsDin(${m.id},'A',this)">A</span></td></tr>`;}).join('')}
      ${!members.length?'<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>':''}
    </table></div></div>
    <div style="height:60px"></div>
    <div style="position:fixed;bottom:var(--bn);left:0;right:0;z-index:10;padding:.5rem .55rem .4rem;background:linear-gradient(to top,var(--bg) 60%,transparent)">
    ${sudah?'<button class="btn btn-full" disabled style="background:var(--green);color:#fff;opacity:.5"><i class="ri-checkbox-circle-line"></i> Sudah Diabsen</button>'
    :'<button class="btn btn-primary btn-full" id="btnAbsDin" onclick="simpanAbsenDin('+kelasId+')" style="box-shadow:0 -4px 16px rgba(0,0,0,.15)"><i class="ri-save-line"></i> Simpan Absen Diniyyah</button>'}</div>`;
  window._absDinData={};members.forEach(m=>window._absDinData[m.id]=sudah?(existMap[m.id]||'H'):'H');
}
function setAbsDin(santriId,status,el){
  window._absDinData[santriId]=status;const row=el.closest('tr');
  row.querySelectorAll('.abs-toggle').forEach(s=>s.classList.remove('active'));el.classList.add('active');
}
async function simpanAbsenDin(kelasId){
  const tanggal=$('absDinTgl')?.value||new Date(Date.now()+7*3600000).toISOString().slice(0,10);
  const items=Object.entries(window._absDinData||{}).map(([sid,st])=>({santri_id:parseInt(sid),status:st}));
  if(!items.length)return toast('Tidak ada data');
  const btn=$('btnAbsDin');
  if(btn){btn.disabled=true;btn.innerHTML='<i class="ri-loader-4-line ri-spin"></i> Menyimpan...';btn.style.opacity='.6';}
  try{
    const res=await api('/api/absen-diniyyah/bulk',{method:'POST',body:JSON.stringify({tanggal,kelas_diniyyah_id:kelasId,jadwal_diniyyah_id:_absDinJadwalId,mata_pelajaran:_absDinMapel,items})});
    toast('✅ '+res.message);
    if(btn){btn.innerHTML='<i class="ri-checkbox-circle-line"></i> Sudah Diabsen';btn.style.opacity='.5';btn.style.background='var(--green)';}
    document.querySelectorAll('.abs-toggle').forEach(el=>el.classList.add('disabled'));
    const tab=$('tab-din-jadwal-'+_absDinJadwalId);
    if(tab)tab.classList.add('done');
  }catch(e){if(btn){btn.disabled=false;btn.innerHTML='<i class="ri-save-line"></i> Simpan Absen Diniyyah';btn.style.opacity='1';}toast('Error: '+e.message);}
}
// =====================================================================
// CETAK RAPORT TERPADU
// =====================================================================

window._crData = { santri: [], optSekolah: '', optDiniyyah: '', optKamar: '' };

window.loadCetakRaport = async function() {
  const isDin = user.role === 'admin_diniyyah';
  try {
    window._rpSettings = await api('/api/settings').catch(()=>({}));
  } catch(e) {}
  
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
          <select id="crSemester" style="padding:.65rem;border-radius:10px;border:1.5px solid var(--border);width:100%">
            <option value="${window._rpSettings?.tahun_ajaran_aktif || '2026/2027'} - ${window._rpSettings?.semester_aktif || 'Ganjil'}">${window._rpSettings?.tahun_ajaran_aktif || '2026/2027'} - ${window._rpSettings?.semester_aktif || 'Ganjil'}</option>
          </select>
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
  
  let list = (window._crData.santri || []).filter(s => s.nama.toLowerCase().includes(q));
  if(fid) {
    if(kat === 'sekolah') list = list.filter(s => s.kelas_sekolah_id == fid);
    else if(kat === 'diniyyah') list = list.filter(s => s.kelas_diniyyah_id == fid);
    else if(kat === 'kegiatan') list = list.filter(s => s.kamar_id == fid);
  }

  $('crList').innerHTML = list.slice(0, 100).map(s => `
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

