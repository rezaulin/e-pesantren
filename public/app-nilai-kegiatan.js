// ── NILAI KEGIATAN ────────────────────────────────────────
// Flow: Pilih Kegiatan → Pilih Kelompok → Input Nilai per Santri

async function loadNilaiKegiatan(){
  const [keg,settings]=await Promise.all([api('/api/kegiatan'),api('/api/settings').catch(()=>({}))]);
  const now=new Date(Date.now()+7*3600000);
  const defBulan=settings.active_month||now.toISOString().slice(0,7);
  $('main').innerHTML=`<div class="page-header fade-up"><h2><i class="ri-star-line"></i> Nilai Kegiatan</h2>
    <p style="color:var(--t3);margin-top:.3rem">Pilih kegiatan, kelompok, dan bulan untuk input nilai</p></div>
    <div class="card fade-up"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:2;min-width:150px"><label>Kegiatan</label>
        <select id="nkKegiatan" onchange="nkUpdateKelompok(this)">
          <option value="">- Pilih Kegiatan -</option>
          ${keg.map(k=>`<option value="${k.id}" data-nama="${k.nama.replace(/"/g,'&quot;')}">${k.nama}</option>`).join('')}
        </select>
      </div>
      <div class="fg" style="flex:2;min-width:150px"><label>Kelompok</label>
        <select id="nkKelompok"><option value="">- Pilih Kegiatan Dulu -</option></select>
      </div>
      <div class="fg" style="flex:1;min-width:100px"><label>Bulan</label><input type="month" id="nkBulan" value="${defBulan}"></div>
      <button class="btn btn-primary btn-sm" onclick="nkLoadTable()"><i class="ri-search-line"></i> Tampilkan</button>
    </div></div>
    <div id="nkTableArea"></div>`;
}
async function nkUpdateKelompok(sel){
  const el=$('nkKelompok');
  if(!sel.value){el.innerHTML='<option value="">- Pilih Kegiatan Dulu -</option>';return;}
  const knama=sel.options[sel.selectedIndex].getAttribute('data-nama');
  el.innerHTML='<option value="">Loading...</option>';
  try{
    const kel=await api('/api/kelompok');
    const filtered=kel.filter(k=>k.kegiatan_nama===knama||k.tipe===knama);
    el.innerHTML=filtered.length?filtered.map(k=>`<option value="${k.id}" data-nama="${k.nama.replace(/"/g,'&quot;')}">${k.nama}</option>`).join(''):'<option value="">Belum ada kelompok</option>';
  }catch(e){el.innerHTML='<option value="">Error</option>';}
}

async function nkLoadTable(){
  const kegId=$('nkKegiatan').value, kelId=$('nkKelompok').value, bulan=$('nkBulan').value;
  if(!kegId||!kelId||!bulan)return toast('Pilih kegiatan, kelompok dan bulan');
  const kegNama=$('nkKegiatan').options[$('nkKegiatan').selectedIndex].text;
  const kelNama=$('nkKelompok').options[$('nkKelompok').selectedIndex].text;

  var [members,existing]=await Promise.all([
    api('/api/kelompok/'+kelId+'/members'),
    api('/api/nilai-kegiatan?kelompok_id='+kelId+'&bulan='+bulan+'&kegiatan_id='+kegId)
  ]);
  var existMap={};
  (existing||[]).forEach(function(n){existMap[n.santri_id]={nilai:n.nilai,catatan:n.catatan||''};});
  var sudah=existing&&existing.length>0;

  var rows=members.map(function(m){
    var ex=existMap[m.santri_id]||{nilai:'',catatan:''};
    return '<tr data-sid="'+m.santri_id+'">'+
      '<td>'+m.santri_nama+'</td>'+
      '<td style="width:90px"><input type="number" class="nk-nilai" min="0" max="100" step="0.1" value="'+(ex.nilai||'')+'" placeholder="0-100" style="width:100%;padding:.4rem .5rem;border-radius:8px;border:1.5px solid var(--border);font-size:.82rem;text-align:center"></td>'+
      '<td><input type="text" class="nk-catatan" value="'+ex.catatan.replace(/"/g,'&quot;')+'" placeholder="Opsional..." style="width:100%;padding:.4rem .5rem;border-radius:8px;border:1.5px solid var(--border);font-size:.82rem"></td>'+
      '</tr>';
  }).join('');

  $('nkTableArea').innerHTML=`
    <div class="card au">
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.6rem;flex-wrap:wrap;gap:.5rem">
    <h3 style="margin:0"><i class="ri-edit-line"></i> Input Nilai (${members.length} santri)</h3>
    <div style="display:flex;gap:.3rem">
    <button class="btn btn-outline btn-sm" onclick="nkSetAll(100)"><i class="ri-checkbox-circle-line"></i> Semua 100</button>
    <button class="btn btn-outline btn-sm" onclick="nkSetAll(0)"><i class="ri-close-circle-line"></i> Reset</button>
    </div></div>
    ${sudah?'<div class="info-box success" style="margin-bottom:.8rem"><i class="ri-checkbox-circle-line"></i> Nilai sudah diinput untuk bulan ini. Anda bisa memperbarui.</div>':''}
    <div class="table-wrap"><table>
    <tr><th>Nama Santri</th><th style="text-align:center;width:90px">Nilai</th><th>Catatan</th></tr>
    ${rows}
    ${members.length?'':'<tr><td colspan="3" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>'}
    </table></div>
    <button class="btn btn-primary" id="btnNkSave" onclick="nkSimpan(${kegId},${kelId})" style="margin-top:1rem"><i class="ri-save-line"></i> Simpan Nilai</button></div>`;

  window._nkState={kegId:kegId,kegNama:kegNama,kelId:kelId,kelNama:kelNama,bulan:bulan};
}

function nkSetAll(val){
  document.querySelectorAll('.nk-nilai').forEach(function(el){el.value=val;});
}

async function nkSimpan(kegId,kelId){
  var bulan=window._nkState?.bulan||$('nkBulan')?.value;
  var rows=document.querySelectorAll('tr[data-sid]');
  var data=[];
  rows.forEach(function(row){
    var sid=parseInt(row.getAttribute('data-sid'));
    var nilai=parseFloat(row.querySelector('.nk-nilai').value)||0;
    var catatan=row.querySelector('.nk-catatan').value||'';
    data.push({santri_id:sid,nilai:nilai,catatan:catatan});
  });
  if(!data.length)return toast('Tidak ada data');
  var btn=$('btnNkSave');
  if(btn){btn.disabled=true;btn.innerHTML='<i class="ri-loader-4-line"></i> Menyimpan...';btn.style.opacity='.6';}
  try{
    var res=await api('/api/nilai-kegiatan/bulk',{method:'POST',body:JSON.stringify({
      kegiatan_id:kegId,kelompok_id:kelId,bulan:bulan,data:data
    })});
    toast('✅ '+(res.message||'Nilai disimpan'));
    if(btn){btn.innerHTML='<i class="ri-checkbox-circle-line"></i> Tersimpan';btn.style.background='var(--green)';btn.style.opacity='.7';}
  }catch(e){
    if(btn){btn.disabled=false;btn.innerHTML='<i class="ri-save-line"></i> Simpan Nilai';btn.style.opacity='1';}
    toast('Error: '+e.message);
  }
}
