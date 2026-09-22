// ==========================================
// SANGU & KANTIN MODULE
// ==========================================

// 1. Admin - Kelola Merchant
async function loadSanguAdminMerchant(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const res = await api('/api/sangu/merchants');
    let html = `<div class="page-header fade-up">
      <h2><i class="ri-store-2-line"></i> Kelola Merchant</h2>
      <button class="btn btn-primary btn-sm" onclick="showAddSanguMerchant()"><i class="ri-add-line"></i> Tambah Merchant</button>
    </div>
    <div class="card au">
      <div class="table-wrap"><table>
        <tr><th>ID</th><th>Nama Merchant</th><th>Pemilik</th><th>Kontak</th><th>Saldo Terkini</th><th>Status</th><th>Aksi</th></tr>`;
    
    res.forEach(d => {
      let stBadge = d.status==='active' ? '<span class="badge-h">Aktif</span>' : '<span class="badge-i">Nonaktif</span>';
      html += `<tr>
        <td>${d.id}</td>
        <td><strong>${d.nama}</strong></td>
        <td>${d.pemilik||'-'}</td>
        <td>${d.kontak||'-'}</td>
        <td style="font-weight:600;color:var(--green)">Rp ${parseInt(d.saldo||0).toLocaleString('id')}</td>
        <td>${stBadge}</td>
        <td>
          <button class="btn btn-outline btn-sm" onclick="editSanguMerchant(${d.id},'${d.nama.replace(/'/g,"\\'")}','${d.pemilik||''}','${d.kontak||''}',${d.status==='active'})"><i class="ri-edit-line"></i> Edit</button>
        </td>
      </tr>`;
    });
    if(!res.length) html+='<tr><td colspan="7" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>';
    html += `</table></div></div>`;
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

function showAddSanguMerchant(){
  $('modal').innerHTML = `<h3>Tambah Merchant</h3>
  <div class="fg"><label>Nama Merchant</label><input type="text" id="smNama" placeholder="Kantin A"></div>
  <div class="fg"><label>Nama Pemilik</label><input type="text" id="smPemilik"></div>
  <div class="fg"><label>Kontak/WA</label><input type="text" id="smKontak"></div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doAddSanguMerchant()">Simpan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
}
async function doAddSanguMerchant(){
  const body = {
    nama: $('smNama').value,
    pemilik: $('smPemilik').value,
    kontak: $('smKontak').value
  };
  if(!body.nama) return toast("Nama wajib diisi");
  try{
    await api('/api/sangu/merchants', {method:'POST', body:JSON.stringify(body)});
    hideModal(); toast("Merchant ditambahkan"); loadSanguAdminMerchant();
  }catch(e){toast("Error: "+e.message);}
}
function editSanguMerchant(id, nama, pemilik, kontak, isActive){
  $('modal').innerHTML = `<h3>Edit Merchant</h3>
  <div class="fg"><label>Nama Merchant</label><input type="text" id="smNama" value="${nama}"></div>
  <div class="fg"><label>Nama Pemilik</label><input type="text" id="smPemilik" value="${pemilik}"></div>
  <div class="fg"><label>Kontak/WA</label><input type="text" id="smKontak" value="${kontak}"></div>
  <div style="margin-bottom:1rem"><label style="display:flex;align-items:center;gap:.5rem;cursor:pointer"><input type="checkbox" id="smActive" ${isActive?'checked':''} style="width:18px;height:18px;accent-color:var(--green)"> <span style="font-weight:600">Merchant Aktif</span></label></div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doEditSanguMerchant(${id})">Simpan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
}
async function doEditSanguMerchant(id){
  const body = {
    nama: $('smNama').value,
    pemilik: $('smPemilik').value,
    kontak: $('smKontak').value,
    status: $('smActive').checked ? 'active' : 'nonaktif'
  };
  if(!body.nama) return toast("Nama wajib diisi");
  try{
    await api('/api/sangu/merchants/'+id, {method:'PUT', body:JSON.stringify(body)});
    hideModal(); toast("Merchant diupdate"); loadSanguAdminMerchant();
  }catch(e){toast("Error: "+e.message);}
}

// 2. Admin - Kelola Penarikan
async function loadSanguAdminWd(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const res = await api('/api/sangu/withdrawals');
    let html = `<div class="page-header fade-up">
      <h2><i class="ri-money-dollar-box-line"></i> Permintaan Pencairan Dana</h2>
    </div>
    <div class="card au">
      <div class="table-wrap"><table>
        <tr><th>ID</th><th>Tanggal</th><th>Merchant</th><th>Nominal</th><th>Status</th><th>Aksi</th></tr>`;
    
    res.forEach(d => {
      let stBadge = d.status==='pending' ? '<span class="badge-w">Pending</span>' : (d.status==='sukses'?'<span class="badge-h">Sukses</span>':'<span class="badge-a">Ditolak</span>');
      html += `<tr>
        <td>${d.id}</td>
        <td>${d.created_at}</td>
        <td><strong>${d.merchant_nama}</strong></td>
        <td style="font-weight:600">Rp ${parseInt(d.nominal).toLocaleString('id')}</td>
        <td>${stBadge}</td>
        <td>
          ${d.status==='pending'?`
            <button class="btn btn-primary btn-sm" onclick="processWd(${d.id}, 'sukses')"><i class="ri-check-line"></i> Setujui</button>
            <button class="btn btn-danger btn-sm" onclick="processWd(${d.id}, 'ditolak')"><i class="ri-close-line"></i> Tolak</button>
          `:'-'}
        </td>
      </tr>`;
    });
    if(!res.length) html+='<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>';
    html += `</table></div></div>`;
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

async function processWd(id, status){
  if(!confirm("Anda yakin memproses penarikan ini menjadi "+status+"?")) return;
  try{
    await api('/api/sangu/withdrawals/'+id+'/process', {method:'PUT', body:JSON.stringify({action: status})});
    toast("Berhasil diproses"); loadSanguAdminWd();
  }catch(e){toast("Error: "+e.message);}
}

// 3. Merchant Dashboard
async function loadSanguMerchantDash(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const res = await api('/api/sangu/merchant/dashboard');
    let html = `<div class="welcome au"><h3>Dashboard Merchant: ${res.merchant.nama}</h3></div>
    <div class="stat-grid" style="grid-template-columns:repeat(2,1fr)">
      <div class="stat-card c-green au"><div class="stat-info"><div class="sn">Rp ${parseInt(res.merchant.saldo||0).toLocaleString('id')}</div><div class="sl">Saldo Saat Ini</div></div><div class="si"><i class="ri-wallet-3-line"></i></div></div>
      <div class="stat-card c-blue au"><div class="stat-info"><div class="sn">${res.produk_count||0}</div><div class="sl">Total Produk</div></div><div class="si"><i class="ri-shopping-bag-3-line"></i></div></div>
    </div>
    
    <div class="card au">
      <h3 style="margin-bottom:.8rem"><i class="ri-history-line"></i> Transaksi Terakhir</h3>
      <div class="table-wrap"><table>
        <tr><th>Trx ID</th><th>Tanggal</th><th>Santri ID</th><th>Total Pembelian</th></tr>`;
    (res.transaksi_terakhir||[]).forEach(d => {
      html += `<tr>
        <td>#${d.id}</td>
        <td>${d.created_at}</td>
        <td>Santri #${d.santri_id}</td>
        <td style="font-weight:600;color:var(--green)">Rp ${parseInt(d.total).toLocaleString('id')}</td>
      </tr>`;
    });
    if(!(res.transaksi_terakhir||[]).length) html+='<tr><td colspan="4" style="text-align:center">Belum ada transaksi</td></tr>';
    html += `</table></div>
    </div>`;
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

// 4. Merchant - Produk
async function loadSanguMerchantProd(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const res = await api('/api/sangu/merchant/products');
    let html = `<div class="card au">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
        <h3><i class="ri-shopping-bag-3-line"></i> Kelola Produk</h3>
        <div style="display:flex; gap:.5rem">
          <button class="btn btn-outline btn-sm" onclick="showAddSanguProdMassal()"><i class="ri-list-check"></i> Tambah Massal</button>
          <button class="btn btn-primary btn-sm" onclick="showAddSanguProd()"><i class="ri-add-line"></i> Tambah Produk</button>
        </div>
      </div>
      <div class="table-wrap"><table>
        <tr><th>ID</th><th>Nama Produk</th><th>Kategori</th><th>Harga</th><th>Stok</th><th>Aksi</th></tr>`;
    
    res.forEach(d => {
      html += `<tr>
        <td>${d.id}</td>
        <td><strong>${d.nama}</strong></td>
        <td>${d.kategori||'-'}</td>
        <td style="font-weight:600">Rp ${parseInt(d.harga).toLocaleString('id')}</td>
        <td>${d.stok}</td>
        <td>
          <button class="btn btn-outline btn-sm" onclick="editSanguProd(${d.id},'${d.nama.replace(/'/g,"\\'")}','${d.kategori||''}',${d.harga},${d.stok})"><i class="ri-edit-line"></i> Edit</button>
          <button class="btn btn-danger btn-sm" onclick="deleteSanguProd(${d.id})"><i class="ri-delete-bin-line"></i></button>
        </td>
      </tr>`;
    });
    if(!res.length) html+='<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>';
    html += `</table></div></div>`;
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

function showAddSanguProd(){
  $('modal').innerHTML = `<h3>Tambah Produk</h3>
  <div class="fg"><label>Nama Produk</label><input type="text" id="spNama"></div>
  <div class="fg"><label>Kategori</label><input type="text" id="spKategori"></div>
  <div class="fg"><label>Harga (Rp)</label><input type="number" id="spHarga"></div>
  <div class="fg"><label>Stok</label><input type="number" id="spStok"></div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doAddSanguProd()">Simpan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
}
async function doAddSanguProd(){
  const body = {
    nama: $('spNama').value,
    kategori: $('spKategori').value,
    harga: parseInt($('spHarga').value)||0,
    stok: parseInt($('spStok').value)||0
  };
  if(!body.nama) return toast("Nama wajib diisi");
  try{
    await api('/api/sangu/merchant/products', {method:'POST', body:JSON.stringify(body)});
    hideModal(); toast("Produk ditambahkan"); loadSanguMerchantProd();
  }catch(e){toast("Error: "+e.message);}
}

let prodMassalRows = 3;
function showAddSanguProdMassal() {
  prodMassalRows = 3;
  let html = `<h3>Tambah Produk Massal</h3>
  <p style="font-size:.8rem; color:var(--t3); margin-bottom:1rem">Isi nama produk, harga, dan stok. Baris yang kosong pada nama produk akan diabaikan.</p>
  <div style="max-height:60vh; overflow-y:auto; overflow-x:hidden; margin-bottom:1rem; padding-right:.5rem" id="massalContainer">
    ${renderMassalRows()}
  </div>
  <div style="display:flex;justify-content:space-between;margin-top:1rem">
    <button class="btn btn-outline btn-sm" onclick="addMassalRow()"><i class="ri-add-line"></i> Tambah Baris</button>
    <div style="display:flex;gap:.5rem">
      <button class="btn btn-primary" onclick="doAddSanguProdMassal(this)">Simpan Semua</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>
  </div>`;
  $('modal').innerHTML = html;
  showModal();
}

function renderMassalRows() {
  let rowsHtml = '';
  for(let i=0; i<prodMassalRows; i++) {
    rowsHtml += `
    <div style="display:flex; gap:.5rem; margin-bottom:.5rem; align-items:center" class="massal-row">
      <div style="flex:1">
        <input type="text" class="m-nama" placeholder="Nama Produk" style="width:100%">
      </div>
      <div style="width:100px">
        <input type="number" class="m-harga" placeholder="Harga" style="width:100%">
      </div>
      <div style="width:80px">
        <input type="number" class="m-stok" placeholder="Stok" value="100" style="width:100%">
      </div>
    </div>`;
  }
  return rowsHtml;
}

function addMassalRow() {
  prodMassalRows++;
  const container = $('massalContainer');
  const div = document.createElement('div');
  div.style.cssText = "display:flex; gap:.5rem; margin-bottom:.5rem; align-items:center";
  div.className = "massal-row";
  div.innerHTML = `
    <div style="flex:1"><input type="text" class="m-nama" placeholder="Nama Produk" style="width:100%"></div>
    <div style="width:100px"><input type="number" class="m-harga" placeholder="Harga" style="width:100%"></div>
    <div style="width:80px"><input type="number" class="m-stok" placeholder="Stok" value="100" style="width:100%"></div>`;
  container.appendChild(div);
  container.scrollTop = container.scrollHeight;
}

async function doAddSanguProdMassal(btn) {
  const rows = document.querySelectorAll('.massal-row');
  let payloads = [];
  
  rows.forEach(r => {
    const nama = r.querySelector('.m-nama').value.trim();
    const harga = parseInt(r.querySelector('.m-harga').value) || 0;
    const stok = parseInt(r.querySelector('.m-stok').value) || 0;
    
    if(nama && harga > 0) {
      payloads.push({ nama, kategori: '', harga, stok });
    }
  });
  
  if(payloads.length === 0) return toast("Tidak ada produk valid untuk ditambahkan!");
  
  const originalText = btn.innerHTML;
  btn.innerHTML = '<i class="ri-loader-4-line spin"></i> Menyimpan...';
  btn.disabled = true;
  
  let successCount = 0;
  try {
    for(const body of payloads) {
      await api('/api/sangu/merchant/products', {method:'POST', body:JSON.stringify(body)});
      successCount++;
    }
    hideModal(); 
    toast(`${successCount} produk berhasil ditambahkan!`);
    loadSanguMerchantProd();
  } catch(e) {
    toast("Error: " + e.message);
    btn.innerHTML = originalText;
    btn.disabled = false;
  }
}
function editSanguProd(id, nama, kategori, harga, stok){
  $('modal').innerHTML = `<h3>Edit Produk</h3>
  <div class="fg"><label>Nama Produk</label><input type="text" id="spNama" value="${nama}"></div>
  <div class="fg"><label>Kategori</label><input type="text" id="spKategori" value="${kategori}"></div>
  <div class="fg"><label>Harga (Rp)</label><input type="number" id="spHarga" value="${harga}"></div>
  <div class="fg"><label>Stok</label><input type="number" id="spStok" value="${stok}"></div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doEditSanguProd(${id})">Simpan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
}
async function doEditSanguProd(id){
  const body = {
    nama: $('spNama').value,
    kategori: $('spKategori').value,
    harga: parseInt($('spHarga').value)||0,
    stok: parseInt($('spStok').value)||0
  };
  if(!body.nama) return toast("Nama wajib diisi");
  try{
    await api('/api/sangu/merchant/products/'+id, {method:'PUT', body:JSON.stringify(body)});
    hideModal(); toast("Produk diupdate"); loadSanguMerchantProd();
  }catch(e){toast("Error: "+e.message);}
}
async function deleteSanguProd(id){
  if(!confirm("Yakin hapus produk ini?")) return;
  try{
    await api('/api/sangu/merchant/products/'+id, {method:'DELETE'});
    toast("Produk dihapus"); loadSanguMerchantProd();
  }catch(e){toast("Error: "+e.message);}
}

// 5. Merchant - Withdrawals
async function loadSanguMerchantWd(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const res = await api('/api/sangu/merchant/withdrawals');
    const dash = await api('/api/sangu/merchant/dashboard');
    
    let html = `<div class="card au">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
        <div>
          <h3><i class="ri-money-dollar-box-line"></i> Penarikan Dana</h3>
          <p style="font-size:.8rem;color:var(--t3)">Saldo Anda: <strong style="color:var(--green)">Rp ${parseInt(dash.merchant.saldo||0).toLocaleString('id')}</strong></p>
        </div>
        <button class="btn btn-primary btn-sm" onclick="showReqSanguWd(${dash.merchant.saldo||0})"><i class="ri-hand-coin-line"></i> Ajukan Penarikan</button>
      </div>
      <div class="table-wrap"><table>
        <tr><th>ID</th><th>Tanggal</th><th>Nominal</th><th>Status</th></tr>`;
    
    res.forEach(d => {
      let stBadge = d.status==='pending' ? '<span class="badge-w">Pending</span>' : (d.status==='sukses'?'<span class="badge-h">Sukses</span>':'<span class="badge-a">Ditolak</span>');
      html += `<tr>
        <td>${d.id}</td>
        <td>${d.created_at}</td>
        <td style="font-weight:600">Rp ${parseInt(d.nominal).toLocaleString('id')}</td>
        <td>${stBadge}</td>
      </tr>`;
    });
    if(!res.length) html+='<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada data</td></tr>';
    html += `</table></div></div>`;
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

function showReqSanguWd(maxSaldo){
  $('modal').innerHTML = `<h3>Ajukan Pencairan Dana</h3>
  <div class="fg">
    <label>Nominal Pencairan</label>
    <input type="number" id="wdNominal" placeholder="Contoh: 50000" max="${maxSaldo}">
    <div style="font-size:.75rem;color:var(--t3);margin-top:.2rem">Maksimal: Rp ${parseInt(maxSaldo).toLocaleString('id')}</div>
  </div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doReqSanguWd()">Ajukan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
}
async function doReqSanguWd(){
  const nominal = parseInt($('wdNominal').value)||0;
  if(nominal < 10000) return toast("Minimal pencairan Rp 10.000");
  try{
    await api('/api/sangu/merchant/withdrawals', {method:'POST', body:JSON.stringify({nominal})});
    hideModal(); toast("Pengajuan berhasil"); loadSanguMerchantWd();
  }catch(e){toast("Error: "+e.message);}
}

// 6. Kasir (POS) — RFID/NFC
let kasirCart = [];
let kasirProducts = [];
let kasirSantri = null;
let kasirSearchQuery = '';

async function loadSanguKasir(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const res = await api('/api/sangu/merchant/products');
    kasirProducts = res;
    kasirCart = [];
    kasirSantri = null;
    kasirSearchQuery = '';
    renderKasirUI();
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

function filterKasirProducts(q) {
  kasirSearchQuery = q.toLowerCase();
  renderKasirUI();
}

function toggleMobileCart() {
  const c = $('mobileCartPanel');
  if(c) c.classList.toggle('open');
}

function renderKasirUI(){
  const filtered = kasirProducts.filter(p => p.nama.toLowerCase().includes(kasirSearchQuery));
  
  let html = `
  <style>
    .kasir-layout { display:flex; gap:1.5rem; align-items:flex-start; flex-wrap:wrap; }
    .kasir-katalog { flex: 1 1 60%; min-width:300px; }
    .kasir-cart { flex: 0 0 360px; position:sticky; top:calc(var(--hh) + 1rem); }
    .cart-backdrop { display:none; }
    @media (max-width: 900px) {
      .kasir-cart.cart-receipt { position:fixed; bottom:0; left:0; right:0; top:auto; z-index:200; border-radius:28px 28px 0 0; box-shadow:0 -10px 40px rgba(0,0,0,0.2); max-height:85vh; overflow-y:auto; transition:transform 0.3s cubic-bezier(0.4, 0, 0.2, 1); transform:translateY(100%); margin:0; width:100%; border:none; padding:1.5rem 1rem calc(var(--bn, 65px) + 1.5rem) !important; }
      .kasir-cart.open { transform:translateY(0); }
      .kasir-cart.open + .cart-backdrop { display:block; position:fixed; inset:0; background:rgba(15,23,42,0.6); z-index:199; backdrop-filter:blur(3px); }
      .kasir-cart-toggle { display:flex!important; position:fixed; bottom:calc(var(--bn) + 1.5rem); right:1.5rem; z-index:99; background:linear-gradient(135deg, #2563eb, #4f46e5); color:#fff; border-radius:50%; width:64px; height:64px; box-shadow:0 12px 30px rgba(79,70,229,0.4); align-items:center; justify-content:center; font-size:1.6rem; cursor:pointer; transition:.2s;}
      .kasir-cart-toggle:active { transform:scale(0.92); }
      .kasir-cart-toggle .badge { position:absolute; top:8px; right:8px; background:var(--red); color:#fff; font-size:.7rem; padding:2px 8px; border-radius:12px; font-weight:800; border:2px solid #4f46e5; box-shadow:0 2px 8px rgba(0,0,0,.2);}
      .kasir-layout { padding-bottom:100px; }
    }
    .kasir-cart-toggle { display:none; }
    .kasir-prod-card { background:rgba(255,255,255,0.95); border:1.5px solid var(--border); border-radius:20px; padding:1.2rem; text-align:center; transition:all .3s cubic-bezier(0.4, 0, 0.2, 1); position:relative; overflow:hidden; box-shadow:0 4px 15px rgba(0,0,0,0.02); display:flex; flex-direction:column; align-items:center;}
    .kasir-prod-card:hover { border-color:#818cf8; transform:translateY(-6px); box-shadow:0 16px 32px rgba(99,102,241,0.12); background:#fff; }
    .kasir-prod-card:active { transform:translateY(-2px); }
    .k-icon-wrap { width:56px; height:56px; border-radius:16px; background:linear-gradient(135deg,rgba(99,102,241,.12),rgba(59,130,246,.18)); display:flex; align-items:center; justify-content:center; margin-bottom:.8rem; font-size:1.6rem; color:#4f46e5; transition:.3s; }
    .kasir-prod-card:hover .k-icon-wrap { transform:scale(1.1) rotate(-5deg); background:linear-gradient(135deg,rgba(99,102,241,.2),rgba(59,130,246,.3)); }
    .k-out { filter:grayscale(1); opacity:0.6; pointer-events:none; border-color:var(--border)!important; transform:none!important; box-shadow:none!important;}
    .k-out-badge { position:absolute; top:16px; right:-35px; background:var(--red); color:#fff; font-size:.65rem; font-weight:800; padding:.25rem 3.5rem; transform:rotate(45deg); z-index:2; box-shadow:0 4px 12px rgba(239,68,68,0.4); letter-spacing:1px;}
    
    .cart-item-row { display:flex; justify-content:space-between; align-items:center; padding:1rem 0; border-bottom:1px dashed var(--border); transition:.2s;}
    .cart-item-row:last-child { border-bottom:none; }
    .qty-btn { width:32px; height:32px; border-radius:10px; border:none; background:var(--bg); color:var(--text); font-weight:bold; cursor:pointer; display:flex; align-items:center; justify-content:center; transition:.2s; font-size:1rem;}
    .qty-btn:hover { background:var(--border); color:var(--p); }
    .cart-receipt { background: #fff; border-radius: 24px; box-shadow: 0 12px 48px rgba(0,0,0,0.06); border: 1px solid rgba(0,0,0,0.04); padding: 1.8rem; }
    .btn-checkout { background: linear-gradient(135deg, #2563eb, #4f46e5); color:#fff; border:none; border-radius:16px; padding:1.3rem; font-size:1.2rem; font-weight:800; width:100%; display:flex; justify-content:center; align-items:center; gap:.6rem; cursor:pointer; box-shadow: 0 8px 24px rgba(79, 70, 229, 0.35); transition:all .3s; letter-spacing:.5px;}
    .btn-checkout:hover { transform:translateY(-3px); box-shadow: 0 16px 32px rgba(79, 70, 229, 0.45); }
    .btn-checkout:active { transform:scale(0.97); }
  </style>
  
  <div class="kasir-layout">
    <div class="kasir-katalog">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1.8rem;flex-wrap:wrap;gap:1rem">
        <h2 style="font-size:1.5rem;margin:0;font-weight:800;display:flex;align-items:center;gap:.5rem"><div style="width:36px;height:36px;border-radius:10px;background:var(--p);color:#fff;display:flex;align-items:center;justify-content:center;font-size:1.1rem;box-shadow:0 4px 12px rgba(79,70,229,.3)"><i class="ri-store-2-fill"></i></div> Katalog Produk</h2>
        <div style="position:relative;flex:1;min-width:240px;max-width:340px">
          <i class="ri-search-line" style="position:absolute;left:1.2rem;top:50%;transform:translateY(-50%);color:var(--t3);font-size:1.1rem"></i>
          <input type="text" placeholder="Cari produk..." onkeyup="filterKasirProducts(this.value)" value="${kasirSearchQuery}" style="width:100%;padding:.8rem 1.2rem .8rem 2.8rem;border-radius:16px;border:1.5px solid var(--border);background:#fff;font-size:1rem;font-weight:500;outline:none;transition:all .2s;box-shadow:0 2px 8px rgba(0,0,0,.02)" onfocus="this.style.borderColor='#818cf8';this.style.boxShadow='0 0 0 4px rgba(99,102,241,0.15)'" onblur="this.style.borderColor='var(--border)';this.style.boxShadow='0 2px 8px rgba(0,0,0,.02)'">
        </div>
      </div>

      <!-- Product List -->
      <div style="display:grid;grid-template-columns:repeat(auto-fill, minmax(150px, 1fr));gap:1.2rem">
        ${filtered.length ? filtered.map(p => {
          const isHabis = p.stok <= 0;
          return `
          <div class="kasir-prod-card ${isHabis ? 'k-out' : ''}" onclick="${isHabis ? '' : 'addToCart(' + p.id + ')'}" style="cursor:${isHabis ? 'default' : 'pointer'}">
            ${isHabis ? '<div class="k-out-badge">HABIS</div>' : ''}
            <div class="k-icon-wrap">
              <i class="ri-shopping-bag-3-fill"></i>
            </div>
            <div style="font-weight:700;font-size:.95rem;line-height:1.3;margin-bottom:.4rem;color:var(--text)">${p.nama}</div>
            <div style="color:var(--green);font-size:.95rem;font-weight:800;background:rgba(22,163,74,.12);padding:.25rem .7rem;border-radius:10px;border:1px solid rgba(22,163,74,.15)">Rp ${parseInt(p.harga).toLocaleString('id')}</div>
            <div style="color:var(--t3);font-size:.78rem;margin-top:.6rem;font-weight:600;display:flex;align-items:center;gap:.3rem"><i class="ri-archive-fill" style="color:var(--t2)"></i> Stok: ${p.stok}</div>
          </div>
        `}).join('') : '<div style="grid-column:1/-1;text-align:center;padding:4rem 2rem;color:var(--t3);background:#fff;border-radius:24px;border:2px dashed var(--border)"><i class="ri-search-line" style="font-size:2.5rem;opacity:.4;margin-bottom:.8rem;display:block"></i><div style="font-size:1.1rem;font-weight:600">Produk tidak ditemukan</div></div>'}
      </div>
    </div>
    
    <!-- Cart Area -->
    <div class="kasir-cart cart-receipt" id="mobileCartPanel">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1.2rem;border-bottom:2px dashed var(--border);padding-bottom:1rem">
        <h3 style="margin:0;font-size:1.3rem;font-weight:800;display:flex;align-items:center;gap:.5rem"><i class="ri-shopping-cart-2-fill" style="color:#4f46e5"></i> Keranjang</h3>
        <button class="btn btn-outline btn-sm" style="border-radius:10px;width:36px;height:36px;padding:0;display:flex;align-items:center;justify-content:center;background:var(--bg);border:none" onclick="kasirCart=[];renderCart()" title="Kosongkan"><i class="ri-delete-bin-line" style="color:var(--red);font-size:1.1rem"></i></button>
      </div>
      <div id="cartList" style="margin:1rem 0;min-height:180px;max-height:45vh;overflow-y:auto;padding-right:.5rem">
        <!-- Cart Items -->
      </div>
      <div style="border-top:2px dashed var(--border);padding-top:1.5rem;margin-top:1rem">
        <div style="display:flex;justify-content:space-between;align-items:center;font-size:1.35rem;font-weight:800;margin-bottom:1.5rem;color:var(--text)">
          <span>Total</span>
          <span id="cartTotal" style="color:var(--green)">Rp 0</span>
        </div>
        <button class="btn-checkout" onclick="processCheckout()">
          <i class="ri-secure-payment-fill"></i> BAYAR SEKARANG
        </button>
      </div>
    </div>
    <div class="cart-backdrop" onclick="toggleMobileCart()"></div>
    
    <!-- Mobile Toggle FAB -->
    <div class="kasir-cart-toggle" onclick="toggleMobileCart()">
      <i class="ri-shopping-cart-2-fill"></i>
      <span class="badge" id="cartBadge" style="display:none">0</span>
    </div>
  </div>`;
  $('main').innerHTML = html;
  
  renderCart();
  
  // Refocus search if it was active
  const searchInput = document.querySelector('.kasir-katalog input');
  if(searchInput && kasirSearchQuery.length > 0) {
    searchInput.focus();
    // Put cursor at the end
    const val = searchInput.value;
    searchInput.value = '';
    searchInput.value = val;
  }
}

function addToCart(pid){
  const prod = kasirProducts.find(p => p.id === pid);
  if(!prod) return;
  const existing = kasirCart.find(c => c.id === pid);
  if(existing){
    if(existing.qty >= prod.stok) return toast("Stok tidak mencukupi");
    existing.qty++;
  } else {
    kasirCart.push({...prod, qty: 1});
  }
  renderCart();
}

function removeFromCart(pid){
  kasirCart = kasirCart.filter(c => c.id !== pid);
  renderCart();
}

function updateCartQty(pid, delta){
  const existing = kasirCart.find(c => c.id === pid);
  if(!existing) return;
  const prod = kasirProducts.find(p => p.id === pid);
  
  if(delta > 0 && existing.qty >= prod.stok) return toast("Stok tidak mencukupi");
  
  existing.qty += delta;
  if(existing.qty <= 0) {
    removeFromCart(pid);
  } else {
    renderCart();
  }
}

function renderCart(){
  const cl = $('cartList');
  if(!cl) return;
  let total = 0;
  let itemsCount = 0;
  let html = '';
  kasirCart.forEach(c => {
    const sub = c.harga * c.qty;
    total += sub;
    itemsCount += c.qty;
    html += `
      <div class="cart-item-row">
        <div style="flex:1">
          <div style="font-weight:700;font-size:.95rem;margin-bottom:.2rem;color:var(--text)">${c.nama}</div>
          <div style="color:var(--t3);font-size:.8rem;font-weight:600">Rp ${parseInt(c.harga).toLocaleString('id')}</div>
        </div>
        <div style="display:flex;align-items:center;gap:.6rem;margin:0 1rem">
          <button class="qty-btn" onclick="updateCartQty(${c.id}, -1)"><i class="ri-subtract-line"></i></button>
          <span style="font-weight:800;font-size:1rem;min-width:24px;text-align:center;color:var(--p)">${c.qty}</span>
          <button class="qty-btn" onclick="updateCartQty(${c.id}, 1)"><i class="ri-add-line"></i></button>
        </div>
        <div style="font-weight:800;font-size:1rem;color:var(--text);text-align:right;min-width:70px">
          ${sub.toLocaleString('id')}
        </div>
      </div>
    `;
  });
  if(!kasirCart.length) html = '<div style="text-align:center;color:var(--t3);padding:3rem 0;font-size:.95rem;font-weight:500"><div style="background:var(--bg);width:72px;height:72px;border-radius:50%;display:flex;align-items:center;justify-content:center;margin:0 auto 1.2rem;box-shadow:inset 0 4px 10px rgba(0,0,0,.03)"><i class="ri-shopping-cart-2-fill" style="font-size:2.2rem;opacity:.4"></i></div>Keranjang kosong</div>';
  cl.innerHTML = html;
  $('cartTotal').textContent = 'Rp ' + total.toLocaleString('id');
  
  const badge = $('cartBadge');
  if(badge) {
    if(itemsCount > 0) {
      badge.textContent = itemsCount;
      badge.style.display = 'block';
    } else {
      badge.style.display = 'none';
    }
  }
}

async function processCheckout(){
  if(!kasirCart.length) return toast("Keranjang kosong");
  
  const total = kasirCart.reduce((sum, c) => sum + (c.harga * c.qty), 0);
  
  $('modal').innerHTML = `<h3><i class="ri-rfid-line"></i> Selesaikan Pembayaran</h3>
  <div style="background:rgba(22,163,74,.05);padding:1.5rem;border-radius:16px;margin:1rem 0;text-align:center;border:1.5px solid rgba(22,163,74,.2)">
    <div style="font-size:.9rem;color:var(--t3)">Total Tagihan Belanja</div>
    <div style="font-size:2rem;font-weight:800;color:var(--green);margin-top:.3rem">Rp ${total.toLocaleString('id')}</div>
  </div>
  
  <div class="fg" style="margin-bottom:.5rem">
    <label style="text-align:center;display:block;margin-bottom:.5rem"><i class="ri-focus-3-line"></i> Tempelkan Kartu RFID Santri</label>
    <input type="text" id="kasirCheckoutBarcode" placeholder="Tap kartu ke reader..." autofocus
      onkeydown="if(event.key==='Enter'){event.preventDefault();executeKasirCheckout()}"
      style="width:100%;font-family:monospace;font-size:1.1rem;padding:.8rem;border-radius:12px;border:2px solid rgba(99,102,241,.4);text-align:center;background:#fff;transition:all .3s"
      onfocus="this.style.borderColor='var(--g1)';this.style.boxShadow='0 0 0 4px rgba(99,102,241,.15)'">
  </div>
  <div style="font-size:.75rem;color:var(--t3);text-align:center;margin-bottom:1.5rem">
    Reader RFID akan otomatis men-submit data saat kartu ditempelkan
  </div>
  
  <div style="display:flex;gap:.5rem;justify-content:center">
    <button class="btn btn-outline" style="min-width:120px" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
  setTimeout(() => { if($('kasirCheckoutBarcode')) $('kasirCheckoutBarcode').focus(); }, 200);
}

async function executeKasirCheckout(){
  const b = $('kasirCheckoutBarcode').value.trim();
  if(!b) return;
  
  const total = kasirCart.reduce((sum, c) => sum + (c.harga * c.qty), 0);
  
  try{
    // Step 1: Validasi Santri & Limit
    toast("Mengecek kartu...", 500);
    const santri = await api('/api/sangu/kasir/santri?barcode=' + encodeURIComponent(b));
    
    let canSpend = santri.saldo;
    if(santri.limit_harian > 0){
      const remainingLimit = santri.limit_harian - santri.terpakai_hari_ini;
      canSpend = Math.min(santri.saldo, Math.max(0, remainingLimit));
    }
    
    if(total > canSpend){
      $('kasirCheckoutBarcode').value = '';
      return toast("Gagal: Saldo/Limit tidak cukup! Maksimal: Rp " + canSpend.toLocaleString('id'));
    }
    
    // Step 2: Proses Checkout
    const body = {
      santri_id: santri.id,
      items: kasirCart.map(c => ({product_id: c.id, qty: c.qty}))
    };
    
    await api('/api/sangu/kasir/checkout', {method:'POST', body:JSON.stringify(body)});
    hideModal();
    toast("✅ Transaksi Berhasil sejumlah Rp " + total.toLocaleString('id'));
    
    // Step 3: Refresh UI & Reset
    const res = await api('/api/sangu/merchant/products');
    kasirProducts = res;
    kasirCart = [];
    renderKasirUI();
  }catch(e){
    $('kasirCheckoutBarcode').value = '';
    toast("Gagal: " + (e.message || "Kartu tidak valid"));
  }
}

let _sanguSantriList = [];
function searchRegSantriPick(q) {
  const box = $('regSearchResults'), hidden = $('regSantriId');
  if(!q || q.length < 1){ box.innerHTML = ''; hidden.value = ''; $('regStatusCard').innerHTML = ''; return; }
  const results = _sanguSantriList.filter(s => s.nama.toLowerCase().includes(q.toLowerCase())).slice(0,8);
  box.innerHTML = results.map(s => {
    const badge = (s.card_uid && s.card_uid !== '') ? ' <span class="badge-h" style="font-size:0.65rem;padding:2px 6px;margin-left:auto">✅ Terdaftar</span>' : '';
    return `<div onclick="pickRegSantri(${s.id}, '${s.nama.replace(/'/g,"\\'").replace(/"/g,"")}', '${s.card_uid||''}')" style="display:flex;align-items:center;padding:.6rem;cursor:pointer;border-bottom:1px solid #eee">${s.nama} ${badge}</div>`;
  }).join('');
}
function pickRegSantri(id, nama, cardUid) {
  $('regSantriId').value = id;
  $('regSearch').value = nama;
  $('regSearchResults').innerHTML = '';
  if(cardUid && cardUid !== 'null' && cardUid !== '') {
    $('regStatusCard').innerHTML = `<span class="badge-h" style="font-size:0.75rem;padding:2px 6px;margin-left:10px"><i class="ri-checkbox-circle-line"></i> Terdaftar</span>
      <button class="btn btn-danger btn-sm" onclick="event.preventDefault();doUnregisterCard(${id},'${nama.replace(/'/g,"\\'")}')" style="margin-left:5px;padding:.2rem .5rem;font-size:.7rem"><i class="ri-link-unlink-M"></i> Unlink</button>`;
  } else {
    $('regStatusCard').innerHTML = `<span class="badge-w" style="font-size:0.75rem;padding:2px 6px;margin-left:10px"><i class="ri-close-circle-line"></i> Belum terdaftar</span>`;
  }
  $('regCardUID').focus();
}

// 6b. Admin — Registrasi Kartu RFID/NFC
async function loadSanguRegisterCard(){
  const m=$('main');
  m.style.opacity='0.5';m.style.pointerEvents='none';
  try{
    const santriList = await api('/api/santri');
    _sanguSantriList = santriList;
    let html = `<div class="page-header fade-up">
      <h2><i class="ri-rfid-line"></i> Registrasi Kartu RFID/NFC</h2>
    </div>
    <div class="card au">
      <p style="font-size:.85rem;color:var(--t3);margin-bottom:1.5rem">Daftarkan kartu RFID ke santri untuk sistem pembayaran kantin.</p>

      <!-- Register Form -->
      <div style="background:linear-gradient(135deg,rgba(99,102,241,.06),rgba(59,130,246,.06));padding:1.5rem;border-radius:16px;margin-bottom:2rem;border:1.5px solid rgba(99,102,241,.15)">
        <div style="display:grid;grid-template-columns:1fr 1fr auto;gap:1rem;align-items:end">
          <div class="fg" style="position:relative">
            <label style="font-weight:600;display:flex;align-items:center"><i class="ri-search-line"></i> Cari Santri <span id="regStatusCard"></span></label>
            <input type="text" id="regSearch" placeholder="Ketik nama santri..." oninput="searchRegSantriPick(this.value)" autocomplete="off" style="width:100%;padding:.65rem;border-radius:10px;border:1.5px solid var(--border);background:#fff;font-size:.85rem">
            <input type="hidden" id="regSantriId">
            <div id="regSearchResults" class="search-results" style="position:absolute;top:100%;left:0;right:0;z-index:10;background:#fff;box-shadow:0 4px 12px rgba(0,0,0,0.1);border-radius:10px;max-height:200px;overflow-y:auto;margin-top:5px"></div>
          </div>
          <div class="fg">
            <label style="font-weight:600"><i class="ri-rfid-line"></i> Tap Kartu RFID</label>
            <input type="text" id="regCardUID" placeholder="Tempelkan kartu ke reader..." 
              onkeydown="if(event.key==='Enter'){event.preventDefault();doRegisterCard()}" 
              style="width:100%;padding:.65rem;border-radius:10px;border:1.5px solid rgba(99,102,241,.3);font-size:.85rem;background:#fff">
          </div>
          <button class="btn btn-primary" onclick="doRegisterCard()" style="height:42px;border-radius:10px;padding:0 1.5rem;white-space:nowrap">
            <i class="ri-save-line"></i> Daftarkan
          </button>
        </div>
      </div>

      <!-- List Santri & Card Status -->
      <h4 style="margin-bottom:.8rem"><i class="ri-list-check-2"></i> Status Kartu Santri</h4>
      <div class="table-wrap"><table>
        <tr><th>ID</th><th>Nama Santri</th><th>Kamar</th><th>Status Kartu</th><th>Card UID</th><th>Aksi</th></tr>`;
    
    santriList.forEach(s => {
      const hasCard = s.card_uid && s.card_uid !== '';
      html += `<tr>
        <td>${s.id}</td>
        <td><strong>${s.nama}</strong></td>
        <td>${s.kamar_nama||'-'}</td>
        <td>${hasCard ? '<span class="badge-h"><i class="ri-checkbox-circle-line"></i> Terdaftar</span>' : '<span class="badge-w"><i class="ri-close-circle-line"></i> Belum</span>'}</td>
        <td style="font-family:monospace;font-size:.8rem;color:var(--t2)">${s.card_uid || '-'}</td>
        <td>
          ${hasCard ? `<button class="btn btn-danger btn-sm" onclick="doUnregisterCard(${s.id},'${s.nama.replace(/'/g,"\\'")}')"><i class="ri-delete-bin-line"></i> Copot</button>` : 
            `<button class="btn btn-outline btn-sm" onclick="quickRegisterCard(${s.id},'${s.nama.replace(/'/g,"\\'")}')"><i class="ri-rfid-line"></i> Daftarkan</button>`}
        </td>
      </tr>`;
    });
    if(!santriList.length) html+='<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada data santri</td></tr>';
    html += `</table></div></div>`;
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
  finally{m.style.opacity='1';m.style.pointerEvents='auto';}
}

async function doRegisterCard(){
  const santriId = $('regSantriId').value;
  const cardUID = $('regCardUID').value.trim();
  if(!santriId) return toast("Pilih santri terlebih dahulu");
  if(!cardUID) return toast("Tap kartu RFID ke reader dulu");
  try{
    const res = await api('/api/sangu/register-card', {method:'PUT', body:JSON.stringify({santri_id: parseInt(santriId), card_uid: cardUID})});
    toast(res.message || "Kartu berhasil didaftarkan");
    loadSanguRegisterCard();
  }catch(e){toast("Error: "+e.message);}
}

async function doUnregisterCard(santriId, nama){
  if(!confirm("Copot kartu RFID dari "+nama+"?")) return;
  try{
    await api('/api/sangu/register-card/'+santriId, {method:'DELETE'});
    toast("Kartu berhasil dicopot");
    loadSanguRegisterCard();
  }catch(e){toast("Error: "+e.message);}
}

function quickRegisterCard(santriId, nama){
  $('modal').innerHTML = `<h3><i class="ri-rfid-line"></i> Daftarkan Kartu RFID</h3>
  <div style="margin-bottom:1rem;font-size:.85rem">Santri: <strong>${nama}</strong></div>
  <div class="fg">
    <label>Tap Kartu RFID ke Reader</label>
    <input type="text" id="qrCardUID" placeholder="Tempelkan kartu..." autofocus
      onkeydown="if(event.key==='Enter'){event.preventDefault();doQuickRegister(${santriId})}"
      style="font-family:monospace;font-size:1rem;padding:.7rem;border-radius:10px;border:1.5px solid rgba(99,102,241,.3)">
  </div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doQuickRegister(${santriId})"><i class="ri-save-line"></i> Simpan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
  setTimeout(() => { if($('qrCardUID')) $('qrCardUID').focus(); }, 200);
}

async function doQuickRegister(santriId){
  const cardUID = $('qrCardUID').value.trim();
  if(!cardUID) return toast("Tap kartu RFID ke reader dulu");
  try{
    const res = await api('/api/sangu/register-card', {method:'PUT', body:JSON.stringify({santri_id: santriId, card_uid: cardUID})});
    hideModal();
    toast(res.message || "Kartu berhasil didaftarkan");
    loadSanguRegisterCard();
  }catch(e){toast("Error: "+e.message);}
}

// 7. Wali - Info Sangu
async function loadSanguWali(){
  const m=$('main');m.innerHTML='<div class="card"><p>Loading...</p></div>';
  try{
    const d=await api('/api/dashboard');
    const anak = d.anak || [];
    if(!anak.length) {
      m.innerHTML = '<div class="card au"><p style="color:var(--t3)">Belum ada data anak terhubung ke akun Anda.</p></div>';
      return;
    }
    
    let html = `<div class="welcome au" style="margin-bottom:1rem;text-align:center"><h3>Informasi Sangu & Jajan Santri</h3></div>`;
    
    html += `<div style="width:100%;margin:0 auto;background:transparent;border-radius:24px;overflow:hidden;position:relative">`;

    for(const a of anak){
      let info = {saldo:0, limit_harian:0};
      let riwayat = [];
      try { info = await api('/api/sangu/wali/'+a.id); } catch(e){}
      try { riwayat = await api('/api/sangu/wali/'+a.id+'/riwayat'); } catch(e){}
      
      html += `
        <!-- E-Wallet Header (Blue Card) -->
        <div class="sangu-header" style="background:linear-gradient(135deg, #1e3a8a, #2563eb);color:#fff;padding:3.5rem 1.5rem 2.5rem;position:relative">
          <div style="position:absolute;top:1rem;right:1rem;background:rgba(255,255,255,.2);padding:.4rem .8rem;border-radius:20px;font-size:.8rem"><i class="ri-notification-3-line"></i></div>
          
          <div style="text-align:center">
            <div style="font-weight:700;font-size:1.2rem;margin-bottom:.8rem;color:#fff"><i class="ri-user-smile-line"></i> ${a.nama}</div>
            <div style="font-size:.9rem;opacity:.8;margin-bottom:.3rem">Saldo Sangu Saat Ini</div>
            <div style="font-size:2.5rem;font-weight:800;letter-spacing:-1px;margin-bottom:1.5rem">Rp ${parseInt(info.saldo||0).toLocaleString('id')}</div>
          </div>
          
          <!-- Action Buttons -->
          <div style="display:flex;justify-content:center;gap:1.5rem">
            <div style="text-align:center;cursor:pointer" onclick="showSanguTopup(${a.id}, '${a.nama.replace(/'/g,"\\'")}')">
              <div style="width:50px;height:50px;background:#fff;color:#2563eb;border-radius:16px;display:flex;align-items:center;justify-content:center;font-size:1.5rem;margin:0 auto .5rem;box-shadow:0 4px 10px rgba(0,0,0,.1);transition:transform .2s" onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">
                <i class="ri-wallet-3-line"></i>
              </div>
              <div style="font-size:.8rem;font-weight:500">Topup</div>
            </div>
            
            <div style="text-align:center;cursor:pointer" onclick="nav('pembayaran-wali')">
              <div style="width:50px;height:50px;background:#fff;color:#16a34a;border-radius:16px;display:flex;align-items:center;justify-content:center;font-size:1.5rem;margin:0 auto .5rem;box-shadow:0 4px 10px rgba(0,0,0,.1);transition:transform .2s" onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">
                <i class="ri-bank-card-line"></i>
              </div>
              <div style="font-size:.8rem;font-weight:500">Bayar SPP</div>
            </div>

            <div style="text-align:center;cursor:pointer" onclick="showSetLimit(${a.id}, ${info.limit_harian||0})">
              <div style="width:50px;height:50px;background:#fff;color:#2563eb;border-radius:16px;display:flex;align-items:center;justify-content:center;font-size:1.5rem;margin:0 auto .5rem;box-shadow:0 4px 10px rgba(0,0,0,.1);transition:transform .2s" onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">
                <i class="ri-equalizer-line"></i>
              </div>
              <div style="font-size:.8rem;font-weight:500">Atur Limit</div>
            </div>
          </div>
          <div style="text-align:center;font-size:.7rem;margin-top:1rem;opacity:.8">Limit Harian: Rp ${parseInt(info.limit_harian||0).toLocaleString('id')}</div>
        </div>
        
        <!-- Transactions List -->
        <div style="padding:1rem 1.5rem 2rem">
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
            <h4 style="margin:0;font-size:1.1rem;color:var(--t1)">Riwayat Transaksi & Topup</h4>
          </div>
          
          <div style="display:flex;flex-direction:column;gap:1rem">`;
          
      if(riwayat.length){
        riwayat.forEach(r => {
          const isTopup = r.tipe === 'topup';
          const icon = isTopup ? 'ri-arrow-left-down-line' : 'ri-shopping-bag-line';
          const iconBg = isTopup ? 'rgba(22,163,74,.1)' : 'rgba(239,68,68,.1)';
          const iconColor = isTopup ? 'var(--green)' : 'var(--red)';
          const amountSign = isTopup ? '+' : '-';
          
          html += `
            <div style="display:flex;align-items:center;background:#fff;padding:1rem;border-radius:16px;box-shadow:0 2px 8px rgba(0,0,0,.04)">
              <div style="width:44px;height:44px;border-radius:12px;background:${iconBg};color:${iconColor};display:flex;align-items:center;justify-content:center;font-size:1.2rem;margin-right:1rem">
                <i class="${icon}"></i>
              </div>
              <div style="flex:1">
                <div style="font-weight:600;font-size:.95rem;color:var(--t1);margin-bottom:.2rem">${r.keterangan}</div>
                <div style="font-size:.75rem;color:var(--t3)">${r.created_at}</div>
              </div>
              <div style="font-weight:700;font-size:.95rem;color:${iconColor}">
                ${amountSign} Rp ${parseInt(r.nominal).toLocaleString('id')}
              </div>
            </div>`;
        });
      } else {
        html += `<div style="text-align:center;color:var(--t3);font-size:.85rem;padding:2rem 0">Belum ada riwayat transaksi</div>`;
      }
      
      html += `</div></div>`;
    }
    html += `</div>`;
    
    m.innerHTML = html;
  }catch(e){m.innerHTML='<div class="card"><p style="color:var(--red)">'+e.message+'</p></div>';}
}

function showSetLimit(santriId, currentLimit){
  $('modal').innerHTML = `<h3>Atur Limit Jajan Harian</h3>
  <div class="fg">
    <label>Limit Harian (Rp)</label>
    <input type="number" id="slNominal" value="${currentLimit}" placeholder="Contoh: 20000">
    <div style="font-size:.75rem;color:var(--t3);margin-top:.3rem">Biarkan 0 jika tidak ingin dibatasi.</div>
  </div>
  <div style="display:flex;gap:.5rem;margin-top:1rem">
    <button class="btn btn-primary" onclick="doSetLimit(${santriId})">Simpan</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();
}
async function doSetLimit(santriId){
  const limit = parseInt($('slNominal').value)||0;
  try{
    await api('/api/sangu/wali/'+santriId+'/limit', {method:'PUT', body:JSON.stringify({limit_harian: limit})});
    hideModal(); toast("Limit berhasil diatur"); loadSanguWali();
  }catch(e){toast("Error: "+e.message);}
}

function showSanguTopup(santriId, namaSantri){
  $('modal').innerHTML = `<h3>Topup Saldo Sangu</h3>
  <div style="margin-bottom:.8rem;font-size:.85rem">Topup untuk santri: <strong>${namaSantri}</strong></div>
  <div class="fg">
    <label>Nominal Topup</label>
    <select id="tuNominal" style="width:100%;padding:.6rem;border-radius:8px;border:1.5px solid var(--border)">
      <option value="50000">Rp 50.000</option>
      <option value="100000">Rp 100.000</option>
      <option value="150000">Rp 150.000</option>
      <option value="200000">Rp 200.000</option>
      <option value="500000">Rp 500.000</option>
    </select>
  </div>
  <div class="fg">
    <label style="font-weight:700;font-size:.82rem;margin-bottom:.4rem;display:block">Metode Pembayaran:</label>
    <div id="tuMethodBtns" style="display:flex;flex-direction:column;gap:.3rem"></div>
  </div>
  <div style="display:flex;gap:.5rem;margin-top:.8rem">
    <button id="tuPayBtn" class="btn btn-primary" disabled style="opacity:.5" onclick="doSanguTopupSelected(${santriId})"><i class="ri-wallet-3-line"></i> Pilih metode dulu</button>
    <button class="btn btn-outline" onclick="hideModal()">Batal</button>
  </div>`;
  showModal();

  // Load settings and render methods
  api('/api/settings').then(ts => {
    window._tenantSettings = ts;
    const m = $('tuMethodBtns');
    if(!m) return;
    let html = '';
    if(ts.transfer_bank_enabled) {
      const fee = ts.transfer_bank_fee_flat || 0;
      html += `<label style="display:flex;align-items:center;gap:.5rem;padding:.5rem .7rem;border-radius:10px;cursor:pointer;border:1.5px solid var(--border);background:#f8fafc;transition:.2s" onclick="selectTopupMethod('transfer_bank',this)">
        <input type="radio" name="tuMethod" value="transfer_bank" style="accent-color:var(--green);width:16px;height:16px">
        <div style="flex:1"><div style="font-weight:700;font-size:.82rem"><i class="ri-bank-line" style="color:var(--green)"></i> Transfer Bank</div>
        <div style="font-size:.7rem;color:var(--t3)">Fee: ${fee > 0 ? 'Rp ' + fee.toLocaleString('id-ID') : 'GRATIS'}</div></div>
      </label>`;
    }
    if(ts.pg_provider) {
      html += `<label style="display:flex;align-items:center;gap:.5rem;padding:.5rem .7rem;border-radius:10px;cursor:pointer;border:1.5px solid var(--border);background:#f8fafc;transition:.2s" onclick="selectTopupMethod('midtrans',this)">
        <input type="radio" name="tuMethod" value="midtrans" style="accent-color:var(--blue);width:16px;height:16px">
        <div style="flex:1"><div style="font-weight:700;font-size:.82rem"><i class="ri-secure-payment-line" style="color:var(--blue)"></i> Payment Gateway</div>
        <div style="font-size:.7rem;color:var(--t3)">VA / Kartu Kredit</div></div>
      </label>`;
    }
    if(!html) html = '<div style="font-size:.78rem;color:var(--red)"><i class="ri-error-warning-line"></i> Belum ada metode pembayaran aktif.</div>';
    m.innerHTML = html;
  }).catch(() => {});
}

window._topupMethod = null;
function selectTopupMethod(method, el) {
  window._topupMethod = method;
  document.querySelectorAll('#tuMethodBtns label').forEach(l => {
    l.style.borderColor = 'var(--border)';
    l.style.background = '#f8fafc';
  });
  el.style.borderColor = method === 'transfer_bank' ? 'var(--green)' : 'var(--blue)';
  el.style.background = method === 'transfer_bank' ? 'rgba(22,163,74,.06)' : 'rgba(59,130,246,.06)';
  const btn = $('tuPayBtn');
  btn.disabled = false;
  btn.style.opacity = '1';
  btn.innerHTML = method === 'transfer_bank' ? '<i class="ri-bank-line"></i> Transfer Bank' : '<i class="ri-secure-payment-line"></i> Bayar via Gateway';
}

async function doSanguTopupSelected(santriId) {
  if(window._topupMethod === 'transfer_bank') {
    doSanguTopupBankTransfer(santriId);
  } else {
    doSanguTopup(santriId);
  }
}

async function doSanguTopupBankTransfer(santriId) {
  const nominal = parseInt($('tuNominal').value)||0;
  const btn = $('tuPayBtn');
  btn.disabled = true;
  btn.innerHTML = '<i class="ri-loader-4-line" style="animation:spin .8s linear infinite"></i> Memproses...';

  try {
    const res = await api('/api/wali/bank-transfer', {
      method: 'POST',
      body: JSON.stringify({ santri_id: santriId, tipe: 'topup_sangu', nominal: nominal })
    });
    const fmtR = (n) => 'Rp ' + (n || 0).toLocaleString('id-ID');
    const expiryDate = new Date(res.expiry).toLocaleString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' });

    $('modal').innerHTML = `<div style="max-width:420px;margin:0 auto">
      <h3 style="margin-bottom:.5rem;color:var(--green);display:flex;align-items:center;gap:.4rem"><i class="ri-checkbox-circle-line"></i> Instruksi Transfer</h3>
      <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">Topup Sangu — Transfer ke rekening berikut:</p>
      <div style="background:linear-gradient(135deg,#f0fdf4,#ecfdf5);border:1.5px solid rgba(22,163,74,.2);border-radius:14px;padding:1rem;margin-bottom:.8rem">
        <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.6rem"><i class="ri-bank-line" style="font-size:1.2rem;color:var(--green)"></i><strong style="font-size:.9rem">${res.rekening.bank}</strong></div>
        <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.3rem">
          <span style="font-size:1.1rem;font-weight:800;letter-spacing:1px">${res.rekening.nomor}</span>
          <button onclick="navigator.clipboard.writeText('${res.rekening.nomor}');toast('📋 Disalin!')" style="background:var(--green);color:#fff;border:none;border-radius:6px;padding:.2rem .5rem;font-size:.68rem;cursor:pointer"><i class="ri-file-copy-line"></i></button>
        </div>
        <div style="font-size:.78rem;color:var(--t3)">a.n. ${res.rekening.atas_nama}</div>
      </div>
      <div style="background:linear-gradient(135deg,rgba(59,130,246,.06),rgba(99,102,241,.06));border:1.5px solid rgba(59,130,246,.2);border-radius:14px;padding:1rem;text-align:center;margin-bottom:.8rem">
        <div style="font-size:.72rem;color:var(--t3);text-transform:uppercase;letter-spacing:.5px">Transfer Tepat Sebesar</div>
        <div style="font-size:1.6rem;font-weight:900;color:#3b82f6;letter-spacing:1px">${fmtR(res.total_transfer)}</div>
        <div style="font-size:.72rem;color:var(--t3);margin-top:.3rem">Nominal ${fmtR(res.nominal)}${res.fee > 0 ? ' + biaya Rp ' + res.fee.toLocaleString('id-ID') : ''} + kode unik <strong style="color:#3b82f6">+${res.kode_unik}</strong></div>
        <button onclick="navigator.clipboard.writeText('${res.total_transfer}');toast('📋 Nominal disalin!')" class="btn btn-sm" style="margin-top:.5rem;background:#3b82f6;color:#fff;font-size:.72rem"><i class="ri-file-copy-line"></i> Salin</button>
      </div>
      <div style="background:rgba(245,158,11,.06);border:1px solid rgba(245,158,11,.2);border-radius:10px;padding:.6rem .7rem;margin-bottom:.8rem">
        <div style="font-size:.75rem;color:#b45309"><i class="ri-error-warning-line"></i> Transfer <strong>TEPAT</strong> nominal di atas! Berlaku <strong>${res.expiry_jam} jam</strong> (s/d ${expiryDate})</div>
      </div>
      <button class="btn btn-outline" onclick="hideModal()" style="width:100%"><i class="ri-check-line"></i> OK, Saya Akan Transfer</button>
    </div>`;
  } catch(e) {
    toast('Error: ' + (e.message || 'Gagal'));
    btn.disabled = false;
    btn.innerHTML = '<i class="ri-bank-line"></i> Transfer Bank';
  }
}

async function doSanguTopup(santriId){
  const nominal = parseInt($('tuNominal').value)||0;
  const btn = $('tuPayBtn') || event?.target;
  const oldText = btn?.innerHTML || '';
  if(btn) { btn.innerHTML = '<i class="ri-loader-4-line ri-spin"></i> Memproses...'; btn.disabled = true; }
  
  try{
    const res = await api('/api/sangu/wali/topup', {method:'POST', body:JSON.stringify({santri_id: santriId, nominal})});
    hideModal();
    
    if (res.snap_token) {
      const ts = window._tenantSettings || {};
      await loadSnapJS(ts.pg_client_key, ts.pg_is_production == 1);
      
      window.snap.pay(res.snap_token, {
        onSuccess: function (result) {
          toast('✅ Pembayaran berhasil! Memproses...');
          fetch('/api/payment/notification', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(result)
          }).then(r => r.json()).then(d => {
            toast('✅ Saldo berhasil ditambahkan!');
            setTimeout(() => loadSanguWali(), 1000);
          }).catch(e => {
            toast('Topup berhasil. Memuat ulang...');
            setTimeout(() => loadSanguWali(), 2000);
          });
        },
        onPending: function (result) {
          toast('⏳ Menunggu pembayaran selesai.');
          fetch('/api/payment/notification', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(result)
          }).catch(() => {});
          setTimeout(() => loadSanguWali(), 2000);
        },
        onError: function (result) {
          toast('❌ Pembayaran gagal.');
        },
        onClose: function () {
          toast('ℹ️ Pembayaran dibatalkan/ditutup.');
        }
      });
    } else if(res.redirect_url){
      window.location.href = res.redirect_url;
    } else {
      toast("Berhasil memproses pembayaran, silakan periksa notifikasi/gateway.");
    }
  }catch(e){
    toast("Error: "+e.message);
    if(btn) { btn.innerHTML = oldText; btn.disabled = false; }
  }
}
