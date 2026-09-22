// ── ABSENSI ───────────────────────────────────────────
let absKegiatan = null,
  absKegiatanId = null,
  absKelompok = null;
async function loadAbsensi() {
  absKegiatan = null;
  absKegiatanId = null;
  absKelompok = null;
  const keg = window._cachedKegiatan || []; // We'll fetch if empty
  const today = new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  
  const renderUI = (kegList) => {
    $("main").innerHTML = `
      <div class="abs-wrapper fade-up">
        <div class="abs-wrapper-header">
          <h2><i class="ri-checkbox-circle-line"></i> Absensi Kegiatan</h2>
          <div class="fg" style="margin-bottom:0.5rem"><label>Tanggal</label><input type="date" id="absTanggal" value="${today}" onchange="if(absKegiatanId) openAbsKegiatan(absKegiatanId, absKegiatan)"></div>
        </div>
        <div id="absKegiatanTabs">
          ${kegList.map(k => `<div class="abs-list-item" id="tab-keg-${k.id}" onclick="openAbsKegiatan(${k.id},'${k.nama.replace(/'/g, "\'")}')">
            <div style="display:flex;align-items:center;gap:.6rem"><i class="ri-book-open-line" style="color:var(--p)"></i> ${k.nama}</div>
            <i class="ri-arrow-right-s-line"></i>
          </div>`).join('') || '<div style="padding:1.5rem;text-align:center;color:var(--t3)">Belum ada kegiatan</div>'}
        </div>
      </div>
      <div id="absKelompokContainer" class="au"></div>
      <div id="absContent">
        <div style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih kegiatan di atas untuk mulai absen</div>
      </div>
    `;
  };
  
  if (keg.length > 0) {
    renderUI(keg);
  } else {
    api("/api/kegiatan").then(k => {
      window._cachedKegiatan = k;
      renderUI(k);
    });
  }
}
async function openAbsKegiatan(kegId, kegNama) {
  absKegiatan = kegNama;
  absKegiatanId = kegId;
  absKelompok = null;

  document
    .querySelectorAll("#absKegiatanTabs .abs-list-item")
    .forEach((el) => el.classList.remove("active"));
  const tab = $("tab-keg-" + kegId);
  if (tab) tab.classList.add("active");

  $("absKelompokContainer").innerHTML =
    '<p style="text-align:center;color:var(--t3);font-size:.8rem;padding:.5rem"><i class="ri-loader-4-line ri-spin"></i> Memuat kelompok...</p>';
  $("absContent").innerHTML = "";

  const kel = await api("/api/kelompok");
  const filtered = kel.filter(
    (k) =>
      k.kegiatan_id === kegId ||
      k.kegiatan_nama === kegNama ||
      k.tipe === kegNama,
  );

  let html = `<div style="font-size:.75rem;color:var(--t3);margin-bottom:.4rem;padding-left:.2rem;font-weight:600">Pilih Kelompok:</div>`;
  html += `<div class="list-view" id="absKelompokTabs" style="gap:.4rem;margin-bottom:1rem">`;
  if (filtered.length) {
    filtered.forEach((k) => {
      html += `<div class="list-item abs-tab" id="tab-kel-${k.id}" onclick="openAbsKelompok(${k.id},'${k.nama.replace(/'/g, "\\'")}')" style="padding:.5rem 1rem;border-radius:8px">
        <div class="title" style="font-size:.85rem;margin:0"><i class="ri-group-line" style="margin-right:.5rem;color:var(--p)"></i> ${k.nama}</div>
      </div>`;
    });
  } else {
    html += `<p style="color:var(--t3);padding:.5rem;font-size:.85rem">Belum ada kelompok</p>`;
  }
  html += `</div>`;

  $("absKelompokContainer").innerHTML = html;
  $("absContent").innerHTML =
    '<div class="card au" style="text-align:center;padding:2rem 1rem;color:var(--t3)"><i class="ri-arrow-up-line" style="font-size:1.5rem;display:block;margin-bottom:.5rem"></i>Pilih kelompok di atas</div>';
}
async function openAbsKelompok(kelId, kelNama) {
  absKelompok = kelId;

  document
    .querySelectorAll("#absKelompokTabs .abs-tab")
    .forEach((el) => el.classList.remove("active"));
  const tab = $("tab-kel-" + kelId);
  if (tab) tab.classList.add("active");

  $("absContent").innerHTML =
    '<p style="text-align:center;color:var(--t3);padding:1rem"><i class="ri-loader-4-line ri-spin"></i> Memuat data...</p>';

  const tanggal =
    $("absTanggal")?.value ||
    new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const [members, existing] = await Promise.all([
    api("/api/kelompok/" + kelId + "/members"),
    api(
      "/api/rekap?kelompok_id=" +
        kelId +
        "&dari=" +
        tanggal +
        "&sampai=" +
        tanggal,
    ),
  ]);
  const sudah = existing.length > 0;
  const existingMap = {};
  existing.forEach((a) => (existingMap[a.santri_id] = a.status));

  if (tab) {
    if (sudah) tab.classList.add("done");
    else tab.classList.remove("done");
  }

  $("absContent").innerHTML = `
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.8rem" class="au">
      <h3 style="margin:0;font-size:1rem;color:var(--text)"><i class="ri-team-line"></i> ${kelNama}</h3>
      <span style="font-size:.75rem;color:var(--t3)">${members.length} santri</span>
    </div>
    ${sudah ? '<div class="info-box success au" style="margin-bottom:.8rem"><i class="ri-checkbox-circle-line"></i> Kelompok ini sudah diabsen pada tanggal ini</div>' : ""}
    <div class="card au" style="padding:0;overflow:hidden;border-radius:16px"><div class="table-wrap" style="border:none;box-shadow:none;border-radius:0"><table style="table-layout:fixed">
      <tr><th style="width:auto">Nama Santri</th><th style="text-align:center;width:44px">H</th><th style="text-align:center;width:44px">I</th><th style="text-align:center;width:44px">S</th><th style="text-align:center;width:44px">A</th></tr>
      ${members
        .map((m) => {
          const st = sudah ? existingMap[m.santri_id] || "H" : "H";
          const dis = sudah ? " disabled" : "";
          return `<tr><td style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${m.santri_nama}</td>
        <td style="text-align:center"><span class="abs-toggle abs-h${st === "H" ? " active" : ""}${dis}" onclick="setStatus(${m.santri_id},'H',this)">H</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-i${st === "I" ? " active" : ""}${dis}" onclick="setStatus(${m.santri_id},'I',this)">I</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-s${st === "S" ? " active" : ""}${dis}" onclick="setStatus(${m.santri_id},'S',this)">S</span></td>
        <td style="text-align:center"><span class="abs-toggle abs-a${st === "A" ? " active" : ""}${dis}" onclick="setStatus(${m.santri_id},'A',this)">A</span></td></tr>`;
        })
        .join("")}
      ${!members.length ? '<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada anggota</td></tr>' : ""}
    </table></div></div>
    <div style="height:60px"></div>
    <div style="position:fixed;bottom:var(--bn);left:0;right:0;z-index:10;padding:.5rem .55rem .4rem;background:linear-gradient(to top,var(--bg) 60%,transparent)">
    ${
      sudah
        ? '<button class="btn btn-full" disabled style="background:var(--green);color:#fff;opacity:.5"><i class="ri-checkbox-circle-line"></i> Sudah Diabsen</button>'
        : '<button class="btn btn-primary btn-full" id="btnSimpanAbs" onclick="simpanAbsensi(' +
          kelId +
          ')" style="box-shadow:0 -4px 16px rgba(0,0,0,.15)"><i class="ri-save-line"></i> Simpan Absensi</button>'
    }</div>`;
  window._absensiData = {};
  members.forEach(
    (m) =>
      (window._absensiData[m.santri_id] = sudah
        ? existingMap[m.santri_id] || "H"
        : "H"),
  );
}
function setStatus(santriId, status, el) {
  window._absensiData[santriId] = status;
  const row = el.closest("tr");
  row
    .querySelectorAll(".abs-toggle")
    .forEach((s) => s.classList.remove("active"));
  el.classList.add("active");
}
async function simpanAbsensi(kelompokId) {
  const tanggal =
    $("absTanggal")?.value ||
    new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const items = Object.entries(window._absensiData || {}).map(([sid, st]) => ({
    santri_id: parseInt(sid),
    status: st,
  }));
  if (!items.length) return toast("Tidak ada data");
  const btn = $("btnSimpanAbs");
  if (btn) {
    btn.disabled = true;
    btn.innerHTML = '<i class="ri-loader-4-line ri-spin"></i> Menyimpan...';
    btn.style.opacity = ".6";
  }
  try {
    const res = await api("/api/absensi/bulk", {
      method: "POST",
      body: JSON.stringify({
        tanggal,
        kelompok_id: kelompokId,
        kegiatan_id: absKegiatanId,
        items,
      }),
    });
    if (res.message) {
      toast("✅ " + res.message);
      if (btn) {
        btn.innerHTML = '<i class="ri-checkbox-circle-line"></i> Sudah Diabsen';
        btn.style.opacity = ".5";
        btn.style.background = "var(--green)";
      }
      document
        .querySelectorAll(".abs-toggle")
        .forEach((el) => el.classList.add("disabled"));
      const tab = $("tab-kel-" + kelompokId);
      if (tab) tab.classList.add("done");
    }
  } catch (e) {
    if (btn) {
      btn.disabled = false;
      btn.innerHTML = '<i class="ri-save-line"></i> Simpan Absensi';
      btn.style.opacity = "1";
    }
    toast("Error: " + e.message);
  }
}

// ── REKAP ─────────────────────────────────────────────
let _rekapTab = "kegiatan";
async function loadRekap() {
  const today = new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const bulanIni = today.slice(0, 7) + "-01";
  _rekapTab = "kegiatan";
  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-file-list-3-line"></i> Rekap Absensi</h2></div>
    <div id="rekapTabs" style="display:flex;gap:.4rem;margin-bottom:.8rem;flex-wrap:wrap">
      <button class="btn btn-primary btn-sm" onclick="switchRekapTab('kegiatan')"><i class="ri-checkbox-circle-line"></i> Kegiatan</button>
      <button class="btn btn-outline btn-sm" onclick="switchRekapTab('sekolah')"><i class="ri-school-line"></i> Sekolah</button>
      <button class="btn btn-outline btn-sm" onclick="switchRekapTab('diniyyah')"><i class="ri-book-2-line"></i> Diniyyah</button>
    </div>
    <div id="rekapFilterArea"></div>
    <div id="rekapSummary"></div>
    <div class="card" id="rekapTable"><p style="color:var(--t3)">Memuat data...</p></div>`;
  await loadRekapKegiatan();
}

function switchRekapTab(tab) {
  _rekapTab = tab;
  const labels = {
    kegiatan: "Kegiatan",
    sekolah: "Sekolah",
    diniyyah: "Diniyyah",
  };
  document.querySelectorAll("#rekapTabs button").forEach((b) => {
    b.className = b.className.replace("btn-primary", "btn-outline");
    Object.values(labels).forEach((l) => {
      if (b.textContent.trim().includes(l) && l === labels[tab])
        b.className = b.className.replace("btn-outline", "btn-primary");
    });
  });
  if (tab === "kegiatan") loadRekapKegiatan();
  else if (tab === "sekolah") loadRekapSekolah();
  else loadRekapDiniyyah();
}

async function loadRekapKegiatan() {
  const today = new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const bulanIni = today.slice(0, 7) + "-01";
  const [keg, kel] = await Promise.all([
    api("/api/kegiatan"),
    api("/api/kelompok"),
  ]);
  $("rekapFilterArea").innerHTML = `<div class="card au">
    <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:1;min-width:150px"><label>Kegiatan</label>${searchSelect(
        {
          id: "rekapKeg",
          items: [
            { value: "", label: "Semua" },
            ...keg.map((k) => ({ value: k.nama, label: k.nama })),
          ],
          placeholder: "Semua / ketik...",
          onSelect: function () {
            updateRekapKelompok();
          },
        },
      )}</div>
      <div class="fg" style="flex:1;min-width:150px"><label>Kelompok</label>${searchSelect({ id: "rekapKel", items: [{ value: "", label: "Semua" }, ...kel.map((k) => ({ value: k.id.toString(), label: k.nama }))], placeholder: "Semua / ketik..." })}</div>
      <div class="fg" style="flex:1;min-width:130px"><label>Dari</label><input type="date" id="rekapDari" value="${bulanIni}"></div>
      <div class="fg" style="flex:1;min-width:130px"><label>Sampai</label><input type="date" id="rekapSampai" value="${today}"></div>
      <button class="btn btn-primary btn-sm" onclick="filterRekap()"><i class="ri-filter-3-line"></i> Filter</button>
    </div></div>`;
  window._allKelompok = kel;
  filterRekap();
}

function updateRekapKelompok() {
  const kn = $("rekapKeg")?.value || "";
  const kel = $("rekapKel");
  if (!kel) return;
  const allK = window._allKelompok || [];
  const f = kn
    ? allK.filter((k) => k.kegiatan_nama === kn || k.tipe === kn)
    : allK;
  kel.innerHTML =
    '<option value="">Semua</option>' +
    f.map((k) => `<option value="${k.id}">${k.nama}</option>`).join("");
}
async function filterRekap() {
  const kn = $("rekapKeg_val")?.value || "",
    ki = $("rekapKel_val")?.value || "",
    d = $("rekapDari")?.value || "",
    s = $("rekapSampai")?.value || "";
  let url = "/api/rekap/summary?";
  if (d) url += "dari=" + d + "&";
  if (s) url += "sampai=" + s + "&";
  if (kn) url += "kegiatan_nama=" + encodeURIComponent(kn) + "&";
  if (ki) url += "kelompok_id=" + ki + "&";
  const res = await api(url);
  const data = res.data || [];
  const jp = res.jumlah_pertemuan || 0;
  const ja = res.jumlah_absensi || 0;
  renderRekapSummary(jp, ja, data.length);
  let exportUrl = "/api/rekap/export-excel?";
  if (d) exportUrl += "dari=" + d + "&";
  if (s) exportUrl += "sampai=" + s + "&";
  if (kn) exportUrl += "kegiatan_nama=" + encodeURIComponent(kn) + "&";
  if (ki) exportUrl += "kelompok_id=" + ki + "&";
  renderRekapTable(data, exportUrl);
}

async function loadRekapSekolah() {
  const today = new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const bulanIni = today.slice(0, 7) + "-01";
  const kelas = await api("/api/kelas-sekolah");
  $("rekapFilterArea").innerHTML = `<div class="card au">
    <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:1;min-width:150px"><label>Kelas</label>${searchSelect({ id: "rekapKelasSekolah", items: [{ value: "", label: "Semua" }, ...kelas.map((k) => ({ value: k.id.toString(), label: k.nama }))], placeholder: "Semua / ketik..." })}</div>
      <div class="fg" style="flex:1;min-width:130px"><label>Dari</label><input type="date" id="rekapDari" value="${bulanIni}"></div>
      <div class="fg" style="flex:1;min-width:130px"><label>Sampai</label><input type="date" id="rekapSampai" value="${today}"></div>
      <button class="btn btn-primary btn-sm" onclick="filterRekapSekolah()"><i class="ri-filter-3-line"></i> Filter</button>
    </div></div>`;
  filterRekapSekolah();
}
async function filterRekapSekolah() {
  const ki = $("rekapKelasSekolah_val")?.value || "",
    d = $("rekapDari")?.value || "",
    s = $("rekapSampai")?.value || "";
  let url = "/api/rekap-absen-sekolah?";
  if (d) url += "dari=" + d + "&";
  if (s) url += "sampai=" + s + "&";
  if (ki) url += "kelas_id=" + ki + "&";
  const res = await api(url);
  const data = res.data || [];
  const js = res.jumlah_sesi || 0;
  renderRekapSummary(
    js,
    data.reduce((t, r) => t + r.total, 0),
    data.length,
  );
  let exportUrl = "/api/rekap-absen-sekolah/export-excel?";
  if (d) exportUrl += "dari=" + d + "&";
  if (s) exportUrl += "sampai=" + s + "&";
  if (ki) exportUrl += "kelas_id=" + ki + "&";
  renderRekapTable(data, exportUrl, "kelas");
}

async function loadRekapDiniyyah() {
  const today = new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const bulanIni = today.slice(0, 7) + "-01";
  const kelas = await api("/api/kelas-diniyyah");
  $("rekapFilterArea").innerHTML = `<div class="card au">
    <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:1;min-width:150px"><label>Kelas Diniyyah</label>${searchSelect({ id: "rekapKelasDiniyyah", items: [{ value: "", label: "Semua" }, ...kelas.map((k) => ({ value: k.id.toString(), label: k.nama }))], placeholder: "Semua / ketik..." })}</div>
      <div class="fg" style="flex:1;min-width:130px"><label>Dari</label><input type="date" id="rekapDari" value="${bulanIni}"></div>
      <div class="fg" style="flex:1;min-width:130px"><label>Sampai</label><input type="date" id="rekapSampai" value="${today}"></div>
      <button class="btn btn-primary btn-sm" onclick="filterRekapDiniyyah()"><i class="ri-filter-3-line"></i> Filter</button>
    </div></div>`;
  filterRekapDiniyyah();
}
async function filterRekapDiniyyah() {
  const ki = $("rekapKelasDiniyyah_val")?.value || "",
    d = $("rekapDari")?.value || "",
    s = $("rekapSampai")?.value || "";
  let url = "/api/rekap-absen-diniyyah?";
  if (d) url += "dari=" + d + "&";
  if (s) url += "sampai=" + s + "&";
  if (ki) url += "kelas_diniyyah_id=" + ki + "&";
  const res = await api(url);
  const data = res.data || [];
  const js = res.jumlah_sesi || 0;
  renderRekapSummary(
    js,
    data.reduce((t, r) => t + r.total, 0),
    data.length,
  );
  let exportUrl = "/api/rekap-absen-diniyyah/export-excel?";
  if (d) exportUrl += "dari=" + d + "&";
  if (s) exportUrl += "sampai=" + s + "&";
  if (ki) exportUrl += "kelas_diniyyah_id=" + ki + "&";
  renderRekapTable(data, exportUrl, "kelas");
}

function renderRekapSummary(pertemuan, absensi, santri) {
  $("rekapSummary").innerHTML =
    `<div style="display:flex;gap:1rem;flex-wrap:wrap;margin:.8rem 0">
    <div class="card au" style="flex:1;min-width:200px;text-align:center;background:linear-gradient(135deg,rgba(16,185,129,.1),rgba(16,185,129,.05));border:1px solid rgba(16,185,129,.2)">
      <div style="font-size:2rem;font-weight:800;color:var(--green)">${pertemuan}</div>
      <div style="font-size:.85rem;color:var(--t2);margin-top:.2rem"><i class="ri-calendar-check-line"></i> Jumlah Pertemuan</div>
    </div>
    <div class="card au" style="flex:1;min-width:200px;text-align:center;background:linear-gradient(135deg,rgba(59,130,246,.1),rgba(59,130,246,.05));border:1px solid rgba(59,130,246,.2)">
      <div style="font-size:2rem;font-weight:800;color:var(--accent)">${absensi}</div>
      <div style="font-size:.85rem;color:var(--t2);margin-top:.2rem"><i class="ri-user-follow-line"></i> Jumlah Pengabsenan</div>
    </div>
    <div class="card au" style="flex:1;min-width:200px;text-align:center;background:linear-gradient(135deg,rgba(168,85,247,.1),rgba(168,85,247,.05));border:1px solid rgba(168,85,247,.2)">
      <div style="font-size:2rem;font-weight:800;color:#a855f7">${santri}</div>
      <div style="font-size:.85rem;color:var(--t2);margin-top:.2rem"><i class="ri-team-line"></i> Jumlah Santri</div>
    </div>
  </div>`;
}

function renderRekapTable(data, exportUrl, extraCol) {
  let tH = 0,
    tI = 0,
    tS = 0,
    tA = 0;
  data.forEach((r) => {
    tH += r.H;
    tI += r.I;
    tS += r.S;
    tA += r.A;
  });
  const hasExtra = extraCol && data.some((r) => r[extraCol]);
  $("rekapTable").innerHTML = `
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.8rem;flex-wrap:wrap;gap:.5rem">
      <h3 style="margin:0"><i class="ri-list-check-3"></i> Detail Per Santri</h3>
      ${exportUrl ? `<button class="btn btn-primary btn-sm" onclick="apiDownload('${exportUrl}','Rekap_Absensi.xlsx')"><i class="ri-file-excel-2-line"></i> Export Excel</button>` : ""}
    </div>
    <div class="table-wrap"><table>
    <tr><th style="width:40px;text-align:center">No</th><th>Nama Santri</th>${hasExtra ? `<th>${extraCol === "kelas" ? "Kelas" : extraCol}</th>` : ""}
    <th style="text-align:center">Hadir</th><th style="text-align:center">Izin</th><th style="text-align:center">Sakit</th><th style="text-align:center">Alpha</th><th style="text-align:center">Total</th></tr>
    ${data
      .map(
        (
          r,
          i,
        ) => `<tr><td style="text-align:center">${i + 1}</td><td>${r.santri_nama || "-"}</td>${hasExtra ? `<td>${r[extraCol] || "-"}</td>` : ""}
      <td style="text-align:center"><span class="badge-h">${r.H}</span></td>
      <td style="text-align:center"><span class="badge-i">${r.I}</span></td>
      <td style="text-align:center"><span class="badge-s">${r.S}</span></td>
      <td style="text-align:center"><span class="badge-a">${r.A}</span></td>
      <td style="text-align:center;font-weight:700">${r.total}</td></tr>`,
      )
      .join("")}
    ${!data.length ? `<tr><td colspan="${hasExtra ? 8 : 7}" style="text-align:center;color:var(--t3)">Tidak ada data</td></tr>` : ""}
    ${
      data.length
        ? `<tr style="font-weight:700;background:rgba(59,130,246,.05)"><td colspan="${hasExtra ? 3 : 2}" style="text-align:center">TOTAL</td>
      <td style="text-align:center">${tH}</td><td style="text-align:center">${tI}</td><td style="text-align:center">${tS}</td><td style="text-align:center">${tA}</td>
      <td style="text-align:center">${tH + tI + tS + tA}</td></tr>`
        : ""
    }
  </table></div>`;
}
// -- STANDALONE REKAP PAGES (per module) --
async function loadRekapKegiatanPage() {
  _rekapTab = "kegiatan";
  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-checkbox-circle-line"></i> Rekap Absensi Kegiatan</h2></div>
    <div id="rekapFilterArea"></div><div id="rekapSummary"></div>
    <div class="card" id="rekapTable"><p style="color:var(--t3)">Memuat data...</p></div>`;
  await loadRekapKegiatan();
}
async function loadRekapSekolahPage() {
  _rekapTab = "sekolah";
  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-school-line"></i> Rekap Absensi Sekolah</h2></div>
    <div id="rekapFilterArea"></div><div id="rekapSummary"></div>
    <div class="card" id="rekapTable"><p style="color:var(--t3)">Memuat data...</p></div>`;
  await loadRekapSekolah();
}
async function loadRekapDiniyyahPage() {
  _rekapTab = "diniyyah";
  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-book-2-line"></i> Rekap Absensi Diniyyah</h2></div>
    <div id="rekapFilterArea"></div><div id="rekapSummary"></div>
    <div class="card" id="rekapTable"><p style="color:var(--t3)">Memuat data...</p></div>`;
  await loadRekapDiniyyah();
}

// -- REKAP KAMAR (Asrama) --
async function loadRekapKamarPage() {
  const today = new Date(Date.now() + 7 * 3600000).toISOString().slice(0, 10);
  const bulanIni = today.slice(0, 7) + "-01";
  const kamar = await api("/api/kamar");
  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-hotel-bed-line"></i> Rekap Absensi Kamar</h2></div>
    <div class="card au"><div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
      <div class="fg" style="flex:1;min-width:150px"><label>Kamar</label>${searchSelect({ id: "rekapKamar", items: [{ value: "", label: "Semua" }, ...kamar.map((k) => ({ value: k.id.toString(), label: k.nama }))], placeholder: "Semua / ketik..." })}</div>
      <div class="fg" style="flex:1;min-width:130px"><label>Dari</label><input type="date" id="rekapDari" value="${bulanIni}"></div>
      <div class="fg" style="flex:1;min-width:130px"><label>Sampai</label><input type="date" id="rekapSampai" value="${today}"></div>
      <button class="btn btn-primary btn-sm" onclick="filterRekapKamar()"><i class="ri-filter-3-line"></i> Filter</button>
    </div></div>
    <div id="rekapSummary"></div>
    <div class="card" id="rekapTable"><p style="color:var(--t3)">Memuat data...</p></div>`;
  filterRekapKamar();
}
async function filterRekapKamar() {
  const ki = $("rekapKamar_val")?.value || "",
    d = $("rekapDari")?.value || "",
    s = $("rekapSampai")?.value || "";
  let url = "/api/rekap-absen-malam?";
  if (d) url += "dari=" + d + "&";
  if (s) url += "sampai=" + s + "&";
  if (ki) url += "kamar_id=" + ki + "&";
  const res = await api(url);
  const data = res.data || [];
  const js = res.jumlah_sesi || 0;
  renderRekapSummary(
    js,
    data.reduce((t, r) => t + r.total, 0),
    data.length,
  );
  let exportUrl = "/api/rekap-absen-malam/export-excel?";
  if (d) exportUrl += "dari=" + d + "&";
  if (s) exportUrl += "sampai=" + s + "&";
  if (ki) exportUrl += "kamar_id=" + ki + "&";
  renderRekapTable(data, exportUrl, "kamar");
}
