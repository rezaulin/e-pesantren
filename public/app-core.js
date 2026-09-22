window.$ = window.$ || (id => document.getElementById(id));
let token =
    sessionStorage.getItem("mt_token") || localStorage.getItem("mt_token"),
  user = null,
  view = "dashboard",
  navStack = [];
const ALL_FEATURES = [
  "absensi_harian",
  "absen_malam",
  "absen_sekolah",
  "keuangan",
  "catatan_bendahara",
  "jadwal",
  "kedisiplinan",
  "raport",
  "kelas_sekolah",
  "madrasah_diniyyah",
  "psb",
  "e_paket",
  "tahfidz",
  "sangu",
];
function hasFeature(key) {
  if (key === "sangu")
    return (
      (user && user._features && user._features.includes("sangu")) ||
      (user && user.fitur_sangu)
    );
  if (!user || !user._features || !user._features.length) return true;
  if (user._features.includes(key)) return true;
  if (key === "kedisiplinan")
    return (
      user._features.includes("catatan_guru") ||
      user._features.includes("pelanggaran")
    );
  return false;
}

// -- API --
async function api(path, opts = {}) {
  const h = { "Content-Type": "application/json" };
  if (token) h["Authorization"] = "Bearer " + token;
  const r = await fetch(path, {
    method: opts.method || 'GET',
    ...opts,
    headers: { ...h, ...(opts.headers || {}) },
  });
  if (r.status === 401) {
    logout();
    throw new Error("Unauthorized");
  }
  const ct = r.headers.get("content-type") || "";
  if (ct.includes("json")) {
    const data = await r.json();
    if (!r.ok) {
      throw new Error(data.message || "Server error " + r.status);
    }
    return data;
  }
  if (!r.ok) throw new Error("Server error " + r.status);
  return r;
}
async function apiDownload(path, filename) {
  const h = {};
  if (token) h["Authorization"] = "Bearer " + token;
  const r = await fetch(path, { headers: h });
  const b = await r.blob();
  const u = URL.createObjectURL(b);
  const a = document.createElement("a");
  a.href = u;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(u);
}

// -- Auth --
async function doLogin() {
  const u = $("loginUser").value,
    p = $("loginPass").value;
  if (!u || !p) return toast("Isi username & password");
  // Extract subdomain from hostname (e.g. "pesantren1.domain.com" + "pesantren1")
  const host = window.location.hostname;
  const parts = host.split(".");
  const subdomain =
    parts.length >= 3 ? parts[0] : host === "localhost" ? "localhost" : "app";
  try {
    const r = await api("/api/login", {
      method: "POST",
      body: JSON.stringify({ username: u, password: p, subdomain }),
    });
    if (r.message) return toast(r.message);
    token = r.token;
    user = r.user;
    // Parse features JSON string into array
    if (user.features && typeof user.features === "string") {
      try {
        user._features = JSON.parse(user.features);
      } catch (e) {
        user._features = [];
      }
    } else {
      user._features = [];
    }
    localStorage.setItem("mt_token", token);
    sessionStorage.setItem("mt_token", token);
    sessionStorage.setItem("mt_user", JSON.stringify(user));
    showApp();
  } catch (e) {
    toast("Login gagal");
  }
}
function logout() {
  token = null;
  user = null;
  localStorage.removeItem("mt_token");
  sessionStorage.removeItem("mt_token");
  sessionStorage.removeItem("mt_user");
  const ls = document.getElementById("loadingScreen");
  if (ls) ls.style.display = "none";
  $("loginPage").style.display = "flex";
  $("app").style.display = "none";
}

// -- App Init --
async function showApp() {
  const ls = document.getElementById("loadingScreen");
  if (ls) ls.style.display = "none";
  $("loginPage").style.display = "none";
  $("app").style.display = "block";
  $("avatarInit").textContent = user.nama.charAt(0).toUpperCase();
  // Dynamic bottom nav per role
  if (user.role === "superadmin") {
    $("bottomNav").style.display = "none";
    document.body.classList.add("sa-mode");
  } else if (user.role === "wali") {
    $("bottomNav").innerHTML = `
      <a onclick="nav('dashboard')" id="bnav-dashboard"><i class="ri-home-5-line"></i>Beranda</a>
      <a onclick="nav('pembayaran-wali')" id="bnav-pembayaran-wali"><i class="ri-bank-card-line"></i>Pembayaran</a>
      <a onclick="nav('raport-absensi')" id="bnav-raport-absensi"><i class="ri-file-chart-line"></i>Raport</a>
      <a onclick="showUserMenu()" id="bnav-profil"><i class="ri-user-3-line"></i>Profil</a>`;
    $("bottomNav").style.display = "flex";
    document.body.classList.remove("sa-mode");
    document.body.classList.add("wali-mode");
  } else if (user.role === "bendahara") {
    $("bottomNav").innerHTML = `
      <a onclick="nav('dashboard')" id="bnav-dashboard"><i class="ri-home-5-line"></i>Beranda</a>
      <a onclick="nav('pembayaran')" id="bnav-pembayaran"><i class="ri-money-dollar-circle-line"></i>SPP</a>
      <a onclick="nav('rekap-tunggakan')" id="bnav-rekap-tunggakan"><i class="ri-file-list-3-line"></i>Tunggakan</a>
      <a onclick="openMoreMenu()" id="bnav-more"><i class="ri-menu-2-line"></i>Lainnya</a>`;
    $("bottomNav").style.display = "flex";
    document.body.classList.remove("sa-mode");
  } else if (user.role === "merchant_admin") {
    $("bottomNav").innerHTML = `
      <a onclick="nav('sangu-merchant-dash')" id="bnav-sangu-merchant-dash"><i class="ri-home-5-line"></i>Beranda</a>
      <a onclick="nav('sangu-merchant-prod')" id="bnav-sangu-merchant-prod"><i class="ri-shopping-bag-3-line"></i>Produk</a>
      <a onclick="nav('sangu-merchant-wd')" id="bnav-sangu-merchant-wd"><i class="ri-money-dollar-box-line"></i>Pencairan</a>
      <a onclick="openMoreMenu()" id="bnav-more"><i class="ri-menu-2-line"></i>Lainnya</a>`;
    $("bottomNav").style.display = "flex";
    document.body.classList.remove("sa-mode");
  } else if (user.role === "kasir") {
    $("bottomNav").innerHTML = `
      <a onclick="nav('sangu-kasir')" id="bnav-sangu-kasir"><i class="ri-shopping-cart-2-line"></i>Kasir POS</a>
      <a onclick="openMoreMenu()" id="bnav-more"><i class="ri-menu-2-line"></i>Lainnya</a>`;
    $("bottomNav").style.display = "flex";
    document.body.classList.remove("sa-mode");
  } else {
    $("bottomNav").innerHTML = `
      <a onclick="nav('dashboard')" id="bnav-dashboard"><i class="ri-home-5-line"></i>Beranda</a>
      <a onclick="nav('absensi')" id="bnav-absensi"><i class="ri-checkbox-circle-line"></i>Absen</a>
      <a onclick="nav('rekap')" id="bnav-rekap"><i class="ri-file-list-3-line"></i>Rekap</a>
      <a onclick="openMoreMenu()" id="bnav-more"><i class="ri-menu-2-line"></i>Lainnya</a>`;
    $("bottomNav").style.display = "flex";
    document.body.classList.remove("sa-mode");
  }
  // Load tenant name from settings
  try {
    const settings = await api("/api/settings");
    console.log("showApp settings loaded:", JSON.stringify(settings));
    const appName = settings.app_name || "Pesantren";
    window._tenantSettings = settings;
    const brandName = appName || user.tenant_nama || "E-Pesantren";
    $("sidebarName").textContent = brandName;
    $("hTitle").textContent = brandName;
    document.title = appName + " - Dashboard";
    document.querySelector(".sb-sub").textContent =
      settings.alamat_lembaga || "Management System";
  } catch (e) {
    console.error("Settings load error:", e);
  }
  buildSidebar();
  let defaultView = "dashboard";
  if (user.role === "merchant_admin" || user.role === "kasir")
    defaultView = "sangu-kasir";
  nav(defaultView);
  // Show PWA install prompt after login (with delay so dashboard loads first)
  setTimeout(() => {
    if (typeof showPwaInstallPrompt === "function") showPwaInstallPrompt();
  }, 2000);
}

// Auto-detect scrollable tables and show fade hint
function initTableFades() {
  document.querySelectorAll(".table-wrap").forEach((w) => {
    const check = () => {
      w.classList.toggle(
        "show-fade",
        w.scrollWidth > w.clientWidth + 4 &&
          w.scrollLeft < w.scrollWidth - w.clientWidth - 4,
      );
    };
    check();
    w.addEventListener("scroll", check, { passive: true });
  });
}
const _tableObs = new MutationObserver(() => setTimeout(initTableFades, 100));
if (document.getElementById("main"))
  _tableObs.observe(document.getElementById("main"), {
    childList: true,
    subtree: true,
  });

// ═══ BULK ADD SANTRI MODAL ═══
async function showBulkAddSantri(opts) {
  // opts: { title, apiUrl, existingIds[], onDone() }
  const allSantri = await api("/api/santri").catch(() => []);
  const exSet = new Set((opts.existingIds || []).map(Number));
  const available = allSantri.filter((s) => !exSet.has(s.id));
  let selected = new Set();
  function render(q) {
    const filt = q
      ? available.filter((s) => s.nama.toLowerCase().includes(q.toLowerCase()))
      : available;
    var rows = "";
    filt.forEach(function (s) {
      var chk = selected.has(s.id) ? "checked" : "";
      rows +=
        '<label style="display:flex;align-items:center;gap:.5rem;padding:.45rem .6rem;border-bottom:1px solid var(--border);cursor:pointer;transition:.15s" onmouseenter="this.style.background=\'rgba(22,163,74,.04)\'" onmouseleave="this.style.background=\'transparent\'">';
      rows +=
        '<input type="checkbox" ' +
        chk +
        ' onchange="this.checked?window._bulkSel.add(' +
        s.id +
        "):window._bulkSel.delete(" +
        s.id +
        ');document.getElementById(\'bulkCount\').textContent=window._bulkSel.size" style="accent-color:var(--green);width:16px;height:16px;flex-shrink:0">';
      rows += '<span style="flex:1;font-size:.82rem">' + s.nama + "</span>";
      rows +=
        '<span style="font-size:.7rem;color:var(--t3)">' +
        (s.kamar_nama || "") +
        "</span></label>";
    });
    if (!filt.length)
      rows =
        '<div style="text-align:center;padding:1.5rem;color:var(--t3)">Tidak ditemukan</div>';
    var hdr =
      '<span style="font-size:.78rem;color:var(--t3)">' +
      available.length +
      " santri tersedia" +
      (exSet.size ? " · " + exSet.size + " sudah anggota" : "") +
      "</span>";
    document.getElementById("bulkList").innerHTML = rows;
    document.getElementById("bulkInfo").innerHTML = hdr;
  }
  window._bulkSel = selected;
  var html =
    '<h3 style="margin-bottom:.8rem"><i class="ri-user-add-line"></i> ' +
    (opts.title || "Tambah Santri") +
    "</h3>";
  html +=
    '<div style="position:relative;margin-bottom:.6rem"><i class="ri-search-line" style="position:absolute;left:.7rem;top:50%;transform:translateY(-50%);color:var(--t3)"></i><input type="text" id="bulkSearch" placeholder="Cari nama santri..." oninput="window._bulkRender(this.value)" style="width:100%;padding:.5rem .8rem .5rem 2rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem;background:var(--card)"></div>';
  html +=
    '<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.4rem"><label style="display:flex;align-items:center;gap:.3rem;cursor:pointer;font-size:.78rem;font-weight:600"><input type="checkbox" id="bulkAll" onchange="var cs=document.querySelectorAll(\'#bulkList input[type=checkbox]\');cs.forEach(function(c){c.checked=this.checked;if(this.checked)window._bulkSel.add(parseInt(c.closest(\'label\').querySelector(\'input\').onchange.toString().match(/add\\((\\d+)\\)/)[1]));else window._bulkSel.clear();}.bind(this));document.getElementById(\'bulkCount\').textContent=window._bulkSel.size" style="accent-color:var(--green)"> Pilih Semua</label><span id="bulkInfo"></span></div>';
  html +=
    '<div id="bulkList" style="max-height:50vh;overflow-y:auto;border:1.5px solid var(--border);border-radius:12px;background:#fff"></div>';
  html +=
    '<div style="display:flex;gap:.5rem;margin-top:1rem;align-items:center"><button class="btn btn-primary" onclick="doBulkAdd()"><i class="ri-add-line"></i> Tambah <span id="bulkCount">0</span> Santri</button><button class="btn btn-outline" onclick="hideModal()">Batal</button></div>';
  $("modal").innerHTML = html;
  showModal();
  window._bulkRender = render;
  window._bulkApiUrl = opts.apiUrl;
  window._bulkOnDone = opts.onDone;
  render("");
}
async function doBulkAdd() {
  var ids = [...window._bulkSel];
  if (!ids.length) return toast("Pilih santri dulu");
  try {
    var r = await api(window._bulkApiUrl, {
      method: "POST",
      body: JSON.stringify({ santri_ids: ids }),
    });
    hideModal();
    toast(r.message || ids.length + " santri ditambahkan");
    if (window._bulkOnDone) window._bulkOnDone();
  } catch (e) {
    toast("Error: " + e.message);
  }
}

// ═══ SEARCHABLE SELECT (replaces dropdown) ═══
// Usage: searchSelect({id, items:[{value,label}], placeholder, onSelect(value,label)})
// Returns HTML string
function searchSelect(opts) {
  var uid = opts.id || "ss_" + Math.random().toString(36).slice(2, 8);
  window["_ss_" + uid] = {
    items: opts.items || [],
    onSelect: opts.onSelect || null,
    value: "",
  };
  var h = '<div style="position:relative">';
  h +=
    '<input type="text" id="' +
    uid +
    '" placeholder="' +
    (opts.placeholder || "Ketik untuk mencari...") +
    '" autocomplete="off" ';
  h +=
    "oninput=\"ssFilter('" +
    uid +
    "',this.value)\" onfocus=\"ssFilter('" +
    uid +
    "',this.value)\" ";
  h +=
    'style="width:100%;padding:.5rem .7rem;border:1.5px solid var(--border);border-radius:10px;font-size:.82rem;background:var(--card)">';
  h += '<input type="hidden" id="' + uid + '_val">';
  h +=
    '<div id="' +
    uid +
    '_dd" class="ss-dropdown" style="display:none;position:absolute;left:0;right:0;top:100%;z-index:50;background:#fff;border:1.5px solid var(--border);border-radius:0 0 12px 12px;max-height:180px;overflow-y:auto;box-shadow:0 8px 24px rgba(0,0,0,.1)"></div>';
  h += "</div>";
  return h;
}
function ssFilter(uid, q) {
  var dd = document.getElementById(uid + "_dd");
  if (!dd) return;
  var data = window["_ss_" + uid];
  if (!data) return;
  var items = data.items || [];
  var lq = (q || "").toLowerCase();
  var filt = lq
    ? items.filter(function (it) {
        return (it.label || "").toLowerCase().indexOf(lq) !== -1;
      })
    : items;
  filt = filt.slice(0, 12);
  if (!filt.length) {
    dd.style.display = "none";
    return;
  }
  dd.style.display = "block";
  dd.innerHTML = filt
    .map(function (it) {
      return (
        '<div style="padding:.45rem .7rem;cursor:pointer;font-size:.82rem;border-bottom:1px solid rgba(0,0,0,.04);transition:.15s" onmousedown="ssPick(\'' +
        uid +
        "','" +
        it.value.toString().replace(/'/g, "\\'") +
        "','" +
        it.label.replace(/'/g, "\\'") +
        "');event.preventDefault()\" onmouseenter=\"this.style.background='rgba(22,163,74,.06)'\" onmouseleave=\"this.style.background='transparent'\">" +
        it.label +
        "</div>"
      );
    })
    .join("");
}
function ssPick(uid, value, label) {
  var inp = document.getElementById(uid);
  if (inp) {
    inp.value = label;
  }
  var hid = document.getElementById(uid + "_val");
  if (hid) {
    hid.value = value;
  }
  var dd = document.getElementById(uid + "_dd");
  if (dd) dd.style.display = "none";
  var data = window["_ss_" + uid];
  if (data) {
    data.value = value;
    if (data.onSelect) data.onSelect(value, label);
  }
}
// Close dropdown on outside click
document.addEventListener("click", function (e) {
  document.querySelectorAll(".ss-dropdown").forEach(function (dd) {
    if (!dd.parentElement.contains(e.target)) dd.style.display = "none";
  });
});

// -- Header Search --
let _searchTimer = null,
  _searchCache = [];
async function doHeaderSearch(q) {
  clearTimeout(_searchTimer);
  const box = $("headerSearchResults");
  if (!q || q.length < 1) {
    box.classList.remove("show");
    box.innerHTML = "";
    return;
  }
  _searchTimer = setTimeout(async () => {
    try {
      if (!_searchCache.length) _searchCache = await api("/api/santri");
      const results = _searchCache
        .filter((s) => s.nama.toLowerCase().includes(q.toLowerCase()))
        .slice(0, 8);
      if (!results.length) {
        box.innerHTML = '<div class="no-result">Tidak ditemukan</div>';
        box.classList.add("show");
        return;
      }
      box.innerHTML = results
        .map(
          (s) =>
            `<a onclick="pickSearchResult(${s.id},'${s.nama.replace(/'/g, "\\'")}')"><i class="ri-graduation-cap-line"></i><div><strong>${s.nama}</strong><br><span style="font-size:.7rem;color:var(--t3)">${s.kamar_nama || "-"}</span></div></a>`,
        )
        .join("");
      box.classList.add("show");
    } catch (e) {
      box.innerHTML = '<div class="no-result">Error</div>';
      box.classList.add("show");
    }
  }, 200);
}
function pickSearchResult(id, nama) {
  $("headerSearch").value = "";
  $("headerSearchResults").classList.remove("show");
  // Navigate to raport for this santri
  nav("raport");
  setTimeout(() => {
    if ($("raportSantri")) {
      $("raportSantri").value = id;
    }
  }, 300);
}
// Close search on click outside
document.addEventListener("click", (e) => {
  if (!e.target.closest(".h-search"))
    $("headerSearchResults")?.classList.remove("show");
});

function buildSidebar() {
  const isAdmin = ["admin", "superadmin"].includes(user.role);
  const isBendahara = user.role === "bendahara";
  const hf = hasFeature;
  let html = "";
  const itm = (id, ic, lb) =>
    `<a onclick="nav('${id}')" id="nav-${id}"><i class="${ic}"></i>${lb}</a>`;
  const sec = (lb) => `<div class="sb-section">${lb}</div>`;
  if (user.role === "superadmin") {
    html += itm("dashboard", "ph-duotone ph-house", "Dashboard");
    html += sec("MASTER");
    html += itm("tenants", "ph-duotone ph-buildings", "Tenants");
    html += itm("super-stats", "ph-duotone ph-chart-bar", "Statistik");
    html += itm("users", "ph-duotone ph-user-gear", "Pengguna SA");
    html += itm("settings", "ph-duotone ph-gear", "Pengaturan");
  } else if (isBendahara) {
    html += sec("UTAMA");
    html += itm("dashboard", "ph-duotone ph-house", "Beranda");
    html += itm("santri", "ph-duotone ph-graduation-cap", "Data Santri");
    if (hf("keuangan") || hf("catatan_bendahara")) {
      html += sec("KEUANGAN");
      if (hf("keuangan"))
        html += itm(
          "pembayaran",
          "ph-duotone ph-currency-circle-dollar",
          "Pembayaran SPP",
        );
      if (hf("keuangan"))
        html += itm(
          "insidental",
          "ph-duotone ph-receipt",
          "Tagihan Insidental",
        );
      if (hf("keuangan"))
        html += itm(
          "rekap-tunggakan",
          "ph-duotone ph-warning-circle",
          "Rekap Tunggakan",
        );
      if (hf("keuangan"))
        html += itm(
          "bank-transfers",
          "ph-duotone ph-bank",
          "Konfirmasi Transfer",
        );
      if (hf("catatan_bendahara"))
        html += itm(
          "catatan-bendahara",
          "ph-duotone ph-book-bookmark",
          "Catatan Bendahara",
        );
    }
    html += sec("LAINNYA");
    html += itm("pengumuman", "ph-duotone ph-megaphone", "Pengumuman");
  } else if (user.role === "keamanan") {
    html += sec("UTAMA");
    html += itm("dashboard", "ph-duotone ph-house", "Beranda");
    html += sec("KEDISIPLINAN");
    html += itm("catatan-guru", "ph-duotone ph-notepad", "Catatan Guru");
    html += itm("pelanggaran", "ph-duotone ph-warning", "Pelanggaran");
    html += itm("perizinan", "ph-duotone ph-check-square-offset", "Perizinan");
    if (hf("e_paket")) {
      html += sec("FITUR EXTRA");
      html += itm("e-paket", "ph-duotone ph-package", "E-Paket");
    }
    html += sec("LAINNYA");
    html += itm("pengumuman", "ph-duotone ph-megaphone", "Pengumuman");
  } else if (user.role === "wali") {
    html += sec("UTAMA");
    html += itm("dashboard", "ph-duotone ph-house", "Beranda");
    html += itm(
      "pembayaran-wali",
      "ph-duotone ph-credit-card",
      "Pembayaran SPP",
    );
    html += sec("LAPORAN ANAK");
    if (hf("raport"))
      html += itm(
        "raport-absensi",
        "ph-duotone ph-file-text",
        "Raport Absensi",
      );
    if (hf("raport"))
      html += itm(
        "raport-penilaian",
        "ph-duotone ph-medal",
        "Raport Penilaian",
      );
    if (hf("raport"))
      html += itm("cetak-raport", "ph-duotone ph-printer", "Cetak Raport");
    if (hf("tahfidz"))
      html += itm(
        "tahfidz-wali",
        "ph-duotone ph-book-open-text",
        "Laporan Tahfidz",
      );
    if (hf("sangu"))
      html += itm("sangu-wali", "ph-duotone ph-wallet", "Sangu & Jajan");
    html += sec("LAINNYA");
    html += itm("pengumuman", "ph-duotone ph-megaphone", "Pengumuman");
  } else if (user.role === "merchant_admin") {
    html += sec("MERCHANT");
    html += itm("sangu-kasir", "ph-duotone ph-shopping-cart", "Kasir (POS)");
    html += itm("sangu-merchant-dash", "ph-duotone ph-storefront", "Dashboard");
    html += itm("sangu-merchant-prod", "ph-duotone ph-shopping-bag", "Produk");
    html += itm("sangu-merchant-wd", "ph-duotone ph-money", "Penarikan");
  } else if (user.role === "kasir") {
    html += sec("KASIR");
    html += itm("sangu-kasir", "ph-duotone ph-shopping-cart", "Kasir (POS)");
  } else {
    html += sec("UTAMA");
    html += itm("dashboard", "ph-duotone ph-house", "Beranda");
    html += itm("santri", "ph-duotone ph-graduation-cap", "Data Santri");
    html += itm("psb", "ph-duotone ph-user-plus", "PSB Online");
    if (hf("kedisiplinan")) {
      html += sec("KEDISIPLINAN");
      html += itm("catatan-guru", "ph-duotone ph-notepad", "Catatan Guru");
      html += itm("pelanggaran", "ph-duotone ph-warning", "Pelanggaran");
      if (isAdmin)
        html += itm(
          "perizinan",
          "ph-duotone ph-check-square-offset",
          "Perizinan",
        );
    }
    if (hf("kelas_sekolah") || hf("absen_sekolah")) {
      html += sec("SEKOLAH FORMAL");
      if (isAdmin) {
        html += itm("kelas-sekolah", "ph-duotone ph-student", "Kelas");
        html += itm(
          "jadwal-sekolah",
          "ph-duotone ph-calendar-blank",
          "Jadwal Pelajaran",
        );
      }
      if (hf("absen_sekolah"))
        html += itm("absen-sekolah", "ph-duotone ph-check-circle", "Absensi");
      html += itm(
        "rekap-sekolah",
        "ph-duotone ph-list-dashes",
        "Rekap Absensi",
      );
      if (isAdmin)
        html += itm("input-nilai-sekolah", "ph-duotone ph-medal", "Penilaian");
      if (isAdmin)
        html += itm(
          "pengaturan-raport-sekolah",
          "ph-duotone ph-gear",
          "Pengaturan Raport",
        );
    }
    // Madrasah Diniyyah (always show)
    {
      html += sec("MADRASAH DINIYYAH");
      if (isAdmin) {
        html += itm("kelas-diniyyah", "ph-duotone ph-books", "Kelas");
        html += itm(
          "jadwal-diniyyah",
          "ph-duotone ph-calendar-blank",
          "Jadwal Pelajaran",
        );
      }
      html += itm("absen-diniyyah", "ph-duotone ph-check-circle", "Absensi");
      html += itm(
        "rekap-diniyyah",
        "ph-duotone ph-list-dashes",
        "Rekap Absensi",
      );
      if (isAdmin)
        html += itm("input-nilai-diniyyah", "ph-duotone ph-medal", "Penilaian");
      if (isAdmin)
        html += itm(
          "pengaturan-raport-diniyyah",
          "ph-duotone ph-gear",
          "Pengaturan Raport",
        );
    }
    if (hf("tahfidz")) {
      html += sec("TAHFIDZ QUR'AN");
      if (isAdmin)
        html += itm(
          "tahfidz-halaqoh",
          "ph-duotone ph-book-open-text",
          "Halaqoh",
        );
      html += itm(
        "tahfidz-absensi",
        "ph-duotone ph-pen-nib",
        "Absensi & Nilai",
      );
      html += itm(
        "tahfidz-rekap",
        "ph-duotone ph-chart-line-up",
        "Rekap Tahfidz",
      );
    }
    if (hf("absen_malam")) {
      html += sec("ASRAMA");
      if (isAdmin)
        html += itm("kamar", "ph-duotone ph-buildings", "Data Kamar");
      html += itm("absen-malam", "ph-duotone ph-moon-stars", "Absen Malam");
      html += itm("rekap-malam", "ph-duotone ph-list-dashes", "Rekap Absen");
    }
    if (hf("absensi_harian") || hf("jadwal")) {
      html += sec("KEGIATAN");
      if (isAdmin)
        html += itm(
          "kegiatan-master",
          "ph-duotone ph-book-open",
          "Kelola Kegiatan",
        );
      if (hf("jadwal"))
        html += itm("jadwal", "ph-duotone ph-calendar", "Jadwal");
      if (hf("absensi_harian"))
        html += itm("absensi", "ph-duotone ph-check-circle", "Absensi");
      html += itm(
        "rekap-kegiatan",
        "ph-duotone ph-list-dashes",
        "Rekap Absensi",
      );
      html += itm("nilai-kegiatan", "ph-duotone ph-star", "Nilai Kegiatan");
      if (isAdmin)
        html += itm(
          "pengaturan-raport-kegiatan",
          "ph-duotone ph-gear",
          "Pengaturan Raport",
        );
    }
    // Laporan
    {
      html += sec("LAPORAN");
      html += itm("cetak-raport", "ph-duotone ph-printer", "Cetak Raport");
      html += itm("rekap-ustadz", "ph-duotone ph-user-focus", "Rekap Ustadz");
      html += itm("sensus", "ph-duotone ph-chart-bar", "Sensus Santri");
      html += itm("statistik", "ph-duotone ph-presentation-chart", "Statistik");
    }
    if (isAdmin || isBendahara) {
      if (hf("keuangan") || hf("catatan_bendahara")) {
        html += sec("KEUANGAN");
        if (hf("keuangan"))
          html += itm(
            "pembayaran",
            "ph-duotone ph-currency-circle-dollar",
            "Pembayaran SPP",
          );
        if (hf("keuangan"))
          html += itm(
            "insidental",
            "ph-duotone ph-receipt",
            "Tagihan Insidental",
          );
        if (hf("keuangan"))
          html += itm(
            "rekap-tunggakan",
            "ph-duotone ph-warning-circle",
            "Rekap Tunggakan",
          );
        if (hf("keuangan"))
          html += itm(
            "bank-transfers",
            "ph-duotone ph-bank",
            "Konfirmasi Transfer",
          );
        if (hf("catatan_bendahara"))
          html += itm(
            "catatan-bendahara",
            "ph-duotone ph-book-bookmark",
            "Catatan Bendahara",
          );
      }
    }
    if (hf("e_paket")) {
      html += sec("FITUR EXTRA");
      html += itm("e-paket", "ph-duotone ph-package", "E-Paket");
    }
    if (hf("sangu")) {
      html += sec("SANGU & KANTIN");
      if (isAdmin) {
        html += itm(
          "sangu-admin-merchant",
          "ph-duotone ph-storefront",
          "Kelola Merchant",
        );
        html += itm("sangu-admin-wd", "ph-duotone ph-money", "Pencairan Dana");
        html += itm(
          "sangu-register-card",
          "ph-duotone ph-identification-card",
          "Registrasi Kartu",
        );
      }
    }
    if (isAdmin) {
      html += sec("PENGATURAN");
      html += itm("users", "ph-duotone ph-users", "Pengguna");
      html += itm("pengumuman", "ph-duotone ph-megaphone", "Pengumuman");
      html += itm("settings", "ph-duotone ph-gear", "Pengaturan");
    }
  }
  $("sidebarMenu").innerHTML = html;
  $("sidebarRole").textContent = user.role.toUpperCase();
}

function nav(id, pushStack = true) {
  if (pushStack && view !== id) navStack.push(view);
  view = id;
  document
    .querySelectorAll(".sb-menu a")
    .forEach((a) => a.classList.remove("active"));
  const el = $("nav-" + id);
  if (el) el.classList.add("active");
  closeSidebar();
  // Bottom nav highlight
  document
    .querySelectorAll(".bottom-nav a")
    .forEach((a) => a.classList.remove("active"));
  const bn = $("bnav-" + id);
  if (bn) bn.classList.add("active");
  // Push browser history so back button works
  if (pushStack) history.pushState({ view: id }, "", " ");
  loadView(id);
}

function goBack() {
  if (navStack.length) {
    const prev = navStack.pop();
    view = prev;
    history.pushState({ view: prev }, "", " ");
    loadView(prev);
  }
}

// Handle browser back button
window.addEventListener("popstate", function (e) {
  if (window._modalClosing) return;
  if (navStack.length) {
    const prev = navStack.pop();
    view = prev;
    loadView(prev);
  } else if (user) {
    view = "dashboard";
    loadView("dashboard");
  }
});

async function loadView(id) {
  const m = $("main");
  // Optional: m.style.opacity = '0.7'; to give slight feedback without clearing
  m.style.opacity = "0.5";
  m.style.pointerEvents = "none";
  try {
    let res;
    switch (id) {
      case "dashboard":
        return await loadDashboard();
      case "tenants":
        return await loadTenants();
      case "super-stats":
        return await loadSuperStats();
      case "kamar":
        return await loadKamar();
      case "kegiatan":
        return await loadKegiatan();
      case "kelompok":
      case "kegiatan-master":
        return await loadKelompok();
      case "absensi":
        return await loadAbsensi();
      case "rekap":
        return await loadRekap();
      case "santri":
        return await loadSantri();
      case "raport":
        return await loadRaportAbsensi();
      case "raport-absensi":
        return await loadRaportAbsensi();
      case "raport-penilaian":
        return await loadRaportPenilaian("semua");
      case "cetak-raport":
        return await loadCetakRaport();
      case "pengaturan-raport-sekolah":
        return await showPengaturanRaport("sekolah");
      case "pengaturan-raport-diniyyah":
        return await showPengaturanRaport("diniyyah");
      case "pengaturan-raport-kegiatan":
        return await showPengaturanRaport("kegiatan");
      case "raport-sekolah":
        return await loadRaportPenilaian("sekolah");
      case "raport-diniyyah":
        return await loadRaportPenilaian("diniyyah");
      case "raport-kegiatan":
        return await loadRaportPenilaian("kegiatan");
      case "catatan-guru":
        return await loadCatatanGuru();
      case "pelanggaran":
        return await loadPelanggaran();
      case "rekap-ustadz":
        return await loadRekapUstadz();
      case "users":
        return await loadUsers();
      case "settings":
        return await loadSettings();
      case "jadwal":
        return await loadJadwalUmum();
      case "kelas":
        return await loadKelasSekolah();
      case "kelas-sekolah":
        return await loadKelasSekolah();
      case "jadwal-sekolah":
        return await loadJadwalSekolahMaster();
      case "kelas-diniyyah":
        return await loadKelasDiniyyah();
      case "jadwal-diniyyah":
        return await loadJadwalDiniyyahMaster();
      case "absen-malam":
        return await loadAbsenMalam();
      case "absen-sekolah":
        return await loadAbsenSekolah();
      case "absen-diniyyah":
        return await loadAbsenDiniyyah();
      case "pembayaran":
        return await loadPembayaran();
      case "insidental":
        return await loadInsidental();
      case "catatan-bendahara":
        return await loadCatatanBendahara();
      case "input-nilai-diniyyah":
        return await loadInputNilaiDiniyyah();
      case "input-nilai-sekolah":
        return await loadInputNilaiSekolah();
      case "nilai-kegiatan":
        return await loadNilaiKegiatan();
      case "rekap-kegiatan":
        return await loadRekapKegiatanPage();
      case "rekap-sekolah":
        return await loadRekapSekolahPage();
      case "rekap-diniyyah":
        return await loadRekapDiniyyahPage();
      case "rekap-malam":
        return await loadRekapKamarPage();
      case "psb":
        return await loadPSBAdmin();
      case "perizinan":
        return await loadPerizinan();
      case "e-paket":
        return await loadEPaket();
      case "tahfidz-halaqoh":
        return await loadTahfidzHalaqoh();
      case "tahfidz-absensi":
        return await loadTahfidzAbsensi();
      case "tahfidz-rekap":
        return await loadTahfidzRekap();
      case "tahfidz-progress":
        return await loadTahfidzProgress();
      case "waktu":
        return await loadWaktu();
      case "pembayaran-wali":
        return await loadPembayaranWali();
      case "rekap-tunggakan":
        return await loadRekapTunggakan();
      case "bank-transfers":
        return await loadBankTransfers();
      case "sensus":
        return await loadSensus();

      // Sangu
      case "sangu-admin-merchant":
        return await loadSanguAdminMerchant();
      case "sangu-admin-wd":
        return await loadSanguAdminWd();
      case "sangu-merchant-dash":
        return await loadSanguMerchantDash();
      case "sangu-merchant-prod":
        return await loadSanguMerchantProd();
      case "sangu-merchant-wd":
        return await loadSanguMerchantWd();
      case "sangu-kasir":
        return await loadSanguKasir();
      case "sangu-register-card":
        return await loadSanguRegisterCard();
      case "sangu-wali":
        return await loadSanguWali();
      case "tahfidz-wali":
        return await loadTahfidzWali();

      default:
        m.innerHTML = '<div class="card"><h3>Coming soon</h3></div>';
    }
  } catch (e) {
    m.innerHTML =
      '<div class="card"><p style="color:var(--red)">' +
      e.message +
      "</p></div>";
  } finally {
    m.style.opacity = "1";
    m.style.pointerEvents = "auto";
  }
}

// -- Dashboard --
async function loadDashboard() {
  if (user.role === "wali") {
    if (typeof loadSanguWali === "function") {
      await loadSanguWali();
      // Highlight dashboard nav instead of sangu-wali
      document
        .querySelectorAll(".bottom-nav a")
        .forEach((a) => a.classList.remove("active"));
      const bn = document.getElementById("bnav-dashboard");
      if (bn) bn.classList.add("active");
      return;
    }
  }
  const d = await api("/api/dashboard");
  const now = new Date(Date.now() + 7 * 3600000);
  const dateStr = new Date().toLocaleDateString("id-ID", {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
  const hariMap = {
    0: "Minggu",
    1: "Senin",
    2: "Selasa",
    3: "Rabu",
    4: "Kamis",
    5: "Jumat",
    6: "Sabtu",
  };
  const hariIni = hariMap[new Date().getDay()];
  // Fetch today's schedule
  let jadwalHtml = "";
  try {
    const jadwal = await api("/api/jadwal-umum");
    const todayJ = jadwal
      .filter((j) => {
        if (j.hari !== hariIni) return false;
        if (user.role === "ustadz") return j.ustadz_username === user.username;
        return true;
      })
      .sort((a, b) => a.jam_mulai.localeCompare(b.jam_mulai));
    if (todayJ.length) {
      jadwalHtml = `<div class="card au" style="margin-top:.8rem">
        <h3><i class="ri-calendar-schedule-line"></i> Jadwal ${hariIni}</h3>
        ${todayJ
          .map(
            (
              j,
            ) => `<div style="display:flex;align-items:center;gap:.7rem;padding:.5rem 0;border-bottom:1px solid var(--border)">
          <div style="width:4px;height:36px;border-radius:4px;background:var(--g1);flex-shrink:0"></div>
          <div style="flex:1"><div style="font-weight:600;font-size:.85rem">${j.kelompok_nama || "Kelompok #" + j.kelompok_id}</div>
          <div style="font-size:.72rem;color:var(--t3)">Ustadz: ${j.ustadz_username}</div></div>
          <div style="text-align:right"><span class="badge-h">${j.jam_mulai}</span><span style="color:var(--t3);font-size:.72rem"> - </span><span class="badge-s">${j.jam_selesai}</span></div>
        </div>`,
          )
          .join("")}
      </div>`;
    }
  } catch (e) {}

  if (d.role === "wali") {
    const bulan = d.bulan || new Date().toISOString().slice(0, 7);
    const anak = d.anak || [];
    let anakHtml = "";
    if (!anak.length) {
      anakHtml =
        '<div class="card au"><p style="color:var(--t3)">Belum ada data anak terhubung ke akun Anda.</p></div>';
    } else {
      for (const a of anak) {
        const ab = a.absensi || {};
        const kg = ab.kegiatan || {};
        const ml = ab.malam || {};
        const sk = ab.sekolah || {};
        const dn = ab.diniyyah || {};
        const kgEntries = Object.entries(kg);
        const pb = a.pembayaran || {};
        const cg = a.catatan_guru || [];
        const pl = a.pelanggaran || [];
        const stColor =
          pb.status === "LUNAS"
            ? "var(--green)"
            : pb.status === "KURANG"
              ? "var(--amber)"
              : "var(--red)";
        const badge = (v, c) =>
          `<span style="display:inline-block;min-width:24px;text-align:center;padding:2px 6px;border-radius:6px;font-weight:700;font-size:.78rem;background:${c}15;color:${c}">${v}</span>`;
        const abRow = (label, d) =>
          `<tr><td style="font-weight:600">${label}</td><td style="text-align:center">${badge(d.H || 0, "#16a34a")}</td><td style="text-align:center">${badge(d.I || 0, "#3b82f6")}</td><td style="text-align:center">${badge(d.S || 0, "#f59e0b")}</td><td style="text-align:center">${badge(d.A || 0, "#ef4444")}</td><td style="text-align:center;font-weight:700">${(d.H || 0) + (d.I || 0) + (d.S || 0) + (d.A || 0)}</td></tr>`;
        anakHtml += `<div class="card au" style="margin-bottom:1rem">
          <h3 style="margin-bottom:.3rem"><i class="ri-graduation-cap-line" style="color:var(--p)"></i> ${a.nama}</h3>
          <div style="display:flex;gap:.8rem;flex-wrap:wrap;font-size:.78rem;color:var(--t3);margin-bottom:.8rem">
            <span><i class="ri-home-5-line"></i> ${a.kamar_nama || "-"}</span><span><i class="ri-book-2-line"></i> ${a.kelas_diniyyah || "-"}</span>
            <span style="color:${a.status === "aktif" ? "var(--green)" : "var(--red)"}">${a.status}</span>
          </div>

          <div style="background:linear-gradient(135deg,rgba(59,130,246,.04),rgba(139,92,246,.04));border:1px solid rgba(59,130,246,.1);border-radius:10px;padding:.8rem;margin-bottom:.8rem">
            <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-bar-chart-box-line"></i> Rekap Absensi - ${bulan}</h4>
            <div class="table-wrap"><table style="font-size:.78rem">
              <tr><th></th><th style="text-align:center">Hadir</th><th style="text-align:center">Izin</th><th style="text-align:center">Sakit</th><th style="text-align:center">Alpa</th><th style="text-align:center">Total</th></tr>
              ${kgEntries.length ? kgEntries.map(([name, d]) => abRow(name, d)).join("") : '<tr><td style="font-weight:600">Kegiatan</td><td colspan="5" style="text-align:center;color:var(--t3);font-size:.75rem">Belum ada data</td></tr>'}${abRow("Malam", ml)}${abRow("Sekolah", sk)}${abRow("Diniyyah", dn)}
            </table></div>
          </div>

          <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:.8rem;margin-bottom:.8rem">
            <div style="background:rgba(22,163,74,.04);border:1px solid rgba(22,163,74,.1);border-radius:10px;padding:.8rem">
              <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-money-dollar-circle-line" style="color:var(--green)"></i> Pembayaran ${bulan}</h4>
              <div style="font-size:.82rem;display:grid;gap:.3rem">
                <div>Tagihan: <strong>Rp ${(pb.tagihan || 0).toLocaleString("id")}</strong></div>
                <div>Dibayar: <strong style="color:var(--green)">Rp ${(pb.total_bayar || 0).toLocaleString("id")}</strong></div>
                <div>Status: <strong style="color:${stColor}">${pb.status || "-"}</strong></div>
                ${pb.kekurangan > 0 ? `<div>Kekurangan: <strong style="color:var(--red)">Rp ${pb.kekurangan.toLocaleString("id")}</strong></div>` : ""}
              </div>
              ${(() => {
                const ts = window._tenantSettings || {};
                // Show "Bayar Sekarang" button if gateway is configured and not lunas
                let payBtn = "";
                if (
                  pb.status !== "LUNAS" &&
                  pb.status !== "GRATIS" &&
                  ts.pg_provider &&
                  ts.pg_client_key
                ) {
                  payBtn = `<div style="margin-top:.6rem"><button class="btn btn-primary btn-sm" onclick="showTunggakanModal(${a.id},'${a.nama.replace(/'/g, "\\'")}')"><i class="ri-bank-card-2-line"></i> Bayar Sekarang</button></div>`;
                }
                // Also show transfer info as alternative
                let transferInfo = "";
                if (pb.status !== "LUNAS" && ts.rekening_nomor) {
                  const waLink = ts.bendahara_telp
                    ? "https://wa.me/" + ts.bendahara_telp.replace(/^0/, "62")
                    : "";
                  transferInfo = `<div style="margin-top:.6rem;padding:.6rem;background:rgba(59,130,246,.06);border:1px solid rgba(59,130,246,.15);border-radius:8px">
                    <div style="font-size:.75rem;font-weight:600;color:var(--p);margin-bottom:.3rem"><i class="ri-bank-card-line"></i> Transfer Manual:</div>
                    <div style="font-size:.8rem;line-height:1.5">
                      <div>${ts.rekening_bank || "Bank"} - <strong>${ts.rekening_nomor}</strong></div>
                      <div>a.n. <strong>${ts.rekening_atas_nama || "-"}</strong></div>
                      ${ts.bendahara_telp ? '<div style="margin-top:.3rem"><a href="' + waLink + '" target="_blank" style="color:var(--green);font-weight:600;text-decoration:none"><i class="ri-whatsapp-line"></i> Hubungi Bendahara (' + ts.bendahara_telp + ")</a></div>" : ""}
                    </div>
                  </div>`;
                }
                return payBtn + transferInfo;
              })()}
            </div>

            <div style="background:rgba(245,158,11,.04);border:1px solid rgba(245,158,11,.1);border-radius:10px;padding:.8rem">
              <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-error-warning-line" style="color:var(--amber)"></i> Pelanggaran (${pl.length})${a.total_poin ? ' - <span style="color:var(--red)">' + a.total_poin + " poin</span>" : ""}</h4>
              ${
                pl.length
                  ? pl
                      .map(
                        (
                          p,
                        ) => `<div style="padding:.3rem 0;border-bottom:1px solid var(--border);font-size:.78rem">
                <div style="font-weight:600;color:var(--red)">${p.jenis}</div>
                <div style="color:var(--t2)">${p.deskripsi || "-"}</div>
                <div style="font-size:.7rem;color:var(--t3)">${p.tanggal} - ${p.poin} poin</div>
              </div>`,
                      )
                      .join("")
                  : '<p style="font-size:.78rem;color:var(--t3)">Tidak ada pelanggaran.</p>'
              }
            </div>
          </div>

          <div style="background:rgba(59,130,246,.04);border:1px solid rgba(59,130,246,.1);border-radius:10px;padding:.8rem;margin-bottom:.8rem">
            <h4 style="font-size:.82rem;margin-bottom:.5rem"><i class="ri-sticky-note-line" style="color:var(--p)"></i> Catatan Guru (${cg.length})</h4>
            ${
              cg.length
                ? cg
                    .map(
                      (
                        c,
                      ) => `<div style="padding:.3rem 0;border-bottom:1px solid var(--border);font-size:.78rem">
              <div style="color:var(--t1)">${c.catatan}</div>
              <div style="font-size:.7rem;color:var(--t3)">${c.tanggal} - ${c.guru || "-"}</div>
            </div>`,
                    )
                    .join("")
                : '<p style="font-size:.78rem;color:var(--t3)">Belum ada catatan</p>'
            }
          </div>

          <div style="display:flex;gap:.5rem;flex-wrap:wrap">
            <button class="btn btn-primary btn-sm" onclick="window._raportSantriId=${a.id};nav('raport-absensi')"><i class="ri-file-chart-line"></i> Raport Absensi</button>
            <button class="btn btn-gold btn-sm" onclick="$('rpSantriId')&&($('rpSantriId').value=${a.id});nav('raport-penilaian')"><i class="ri-award-line"></i> Raport Penilaian</button>
          </div>
        </div>`;
      }
    }
    $("main").innerHTML =
      `<div class="welcome au"><h3>Assalamu'alaikum, ${user.nama}</h3><p>${dateStr} | <span id="dashClock"></span> WIB</p></div>${anakHtml}`;
    startDashClock();
    return;
  }
  const isAdmin = ["admin", "superadmin"].includes(user.role);

  // -- SuperAdmin Dashboard --
  if (user.role === "superadmin") {
    let stats = [];
    try {
      stats = await api("/api/super/stats");
    } catch (e) {}
    const totalTenant = stats.length;
    const totalSantri = stats.reduce((s, x) => s + (x.total_santri || 0), 0);
    const totalUsers = stats.reduce((s, x) => s + (x.total_users || 0), 0);
    const active = stats.filter(
      (x) =>
        x.status === "active" &&
        !(x.expired_at && new Date(x.expired_at) < new Date()),
    ).length;
    const expired = stats.filter(
      (x) => x.expired_at && new Date(x.expired_at) < new Date(),
    ).length;
    const suspended = stats.filter((x) => x.status === "suspended").length;
    $("main").innerHTML =
      `<div class="welcome au"><h3>Selamat datang, ${user.nama}</h3><p/h3><p>Kelola semua tenant dan pantau sistem e-Pesantren.</p></div>
    <div class="stat-grid" style="grid-template-columns:repeat(4,1fr)">
      <div class="stat-card c-green au"><div class="stat-info"><div class="sn">${totalTenant}</div><div class="sl">Total Tenant</div></div><div class="si"><i class="ri-building-2-line"></i></div></div>
      <div class="stat-card c-blue au"><div class="stat-info"><div class="sn">${totalSantri.toLocaleString("id-ID")}</div><div class="sl">Total Santri</div></div><div class="si"><i class="ri-group-line"></i></div></div>
      <div class="stat-card c-amber au"><div class="stat-info"><div class="sn">${totalUsers}</div><div class="sl">Total Pengguna</div></div><div class="si"><i class="ri-user-line"></i></div></div>
      <div class="stat-card c-red au"><div class="stat-info"><div class="sn">${expired}</div><div class="sl">Expired</div></div><div class="si"><i class="ri-error-warning-line"></i></div></div>
    </div>
    <div style="display:flex;gap:.8rem;flex-wrap:wrap;margin-bottom:1rem">
      <button class="btn btn-primary btn-sm" onclick="nav('tenants')"><i class="ri-building-2-line"></i> Kelola Tenant</button>
      <button class="btn btn-outline btn-sm" onclick="showChangePassword()"><i class="ri-lock-line"></i> Ganti Password</button>
      <button class="btn btn-outline btn-sm" onclick="showBroadcast()"><i class="ri-megaphone-line"></i> Broadcast</button>
    </div>
    <div class="card au"><h3 style="margin-bottom:.8rem"><i class="ri-building-2-line"></i> Daftar Tenant</h3>
    <div class="table-wrap"><table>
      <tr><th>#</th><th>Nama Pesantren</th><th>Subdomain</th><th>Santri</th><th>Status</th><th>Expired</th></tr>
      ${stats
        .slice(0, 10)
        .map((x, i) => {
          const isExp = x.expired_at && new Date(x.expired_at) < new Date();
          const badge =
            x.status === "active" && !isExp
              ? '<span class="badge-h">Aktif</span>'
              : isExp
                ? '<span class="badge-a">Expired</span>'
                : '<span class="badge-i">Suspended</span>';
          return `<tr><td>${i + 1}</td><td><strong>${x.nama}</strong></td><td style="font-size:.78rem;color:var(--t2)">${x.subdomain || "-"}</td><td>${x.total_santri}</td><td>${badge}</td><td style="font-size:.78rem">${x.expired_at || "-"}</td></tr>`;
        })
        .join("")}
      ${!stats.length ? '<tr><td colspan="6" style="text-align:center;color:var(--t3)">Belum ada tenant</td></tr>' : ""}
    </table></div>
    ${stats.length > 10 ? `<p style="text-align:center;margin-top:.5rem"><a onclick="nav('tenants')" style="color:var(--p);cursor:pointer;font-size:.82rem;font-weight:600">Lihat semua ${totalTenant} tenant</a></p>` : ""}
    </div>`;
    startDashClock();
    return;
  }

  // -- Ustadz Dashboard (premium) --
  if (user.role === "ustadz") {
    const _totalAbsen =
      (d.hadir_hari_ini || 0) + (d.izin_sakit || 0) + (d.alfa || 0);
    const pctHadir =
      _totalAbsen > 0
        ? Math.round(((d.hadir_hari_ini || 0) / _totalAbsen) * 100)
        : 0;
    const thfCard = hasFeature("tahfidz")
      ? `<div class="ust-menu-card au" onclick="nav('tahfidz-absensi')"><div class="umc-icon" style="background:rgba(13,148,136,0.12);color:#0d9488"><i class="ph-duotone ph-book-open-text"></i></div><div class="umc-label">Tahfidz</div><div class="umc-sub">Setoran</div></div>`
      : "";
    $("main").innerHTML = `
    <div class="welcome-banner au">
      <h3>Assalamu'alaikum, ${user.nama}</h3>
      <div class="wb-motto">Mendidik generasi Qur'ani dengan ilmu dan akhlaq</div>
      <div class="wb-date"><span><i class="ri-calendar-line"></i> ${dateStr}</span><span><i class="ri-time-line"></i> <span id="dashClock"></span> WIB</span></div>
    </div>

    <div class="stat-grid au" style="grid-template-columns:repeat(4,1fr);margin-bottom:.8rem">
      <div class="stat-card c-green"><div class="stat-info"><div class="sn">${d.total_santri || 0}</div><div class="sl">Total Santri</div></div><div class="si"><i class="ri-group-line"></i></div></div>
      <div class="stat-card c-blue"><div class="stat-info"><div class="sn">${d.hadir_hari_ini || 0}</div><div class="sl">Hadir</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
      <div class="stat-card c-amber"><div class="stat-info"><div class="sn">${d.izin_sakit || 0}</div><div class="sl">Izin/Sakit</div></div><div class="si"><i class="ri-pass-valid-line"></i></div></div>
      <div class="stat-card c-red"><div class="stat-info"><div class="sn">${pctHadir}%</div><div class="sl">Kehadiran</div></div><div class="si"><i class="ri-pie-chart-line"></i></div></div>
    </div>

    <div class="card au" style="margin-bottom:.8rem;position:relative;z-index:10;padding:1rem">
      <div style="display:flex;align-items:center;gap:.5rem">
        <div style="position:relative;flex:1">
          <i class="ri-search-line" style="position:absolute;left:.75rem;top:50%;transform:translateY(-50%);color:var(--t3);font-size:1rem"></i>
          <input type="text" id="dashSantriSearch" placeholder="Cari data santri..." oninput="searchSantriProfile(this.value)" autocomplete="off" style="width:100%;padding:.7rem 1rem .7rem 2.4rem;border-radius:14px;border:1.5px solid var(--border);font-size:.82rem;background:#f6f9f7">
          <div id="dashSantriResults" class="search-results"></div>
        </div>
      </div>
      <div id="santriProfileArea"></div>
    </div>

    ${jadwalHtml}

    <div class="section-title" style="margin-top:.6rem">MENU UTAMA</div>
    <div class="ust-menu-grid au">
      <div class="ust-menu-card au" onclick="nav('absensi')"><div class="umc-icon" style="background:rgba(5,150,105,0.12);color:#059669"><i class="ph-duotone ph-check-circle"></i></div><div class="umc-label">Absensi</div><div class="umc-sub">Harian</div></div>
      <div class="ust-menu-card au" onclick="nav('absen-sekolah')"><div class="umc-icon" style="background:rgba(217,119,6,0.12);color:#d97706"><i class="ph-duotone ph-student"></i></div><div class="umc-label">Sekolah</div><div class="umc-sub">Formal</div></div>
      <div class="ust-menu-card au" onclick="nav('absen-diniyyah')"><div class="umc-icon" style="background:rgba(124,58,237,0.12);color:#7c3aed"><i class="ph-duotone ph-books"></i></div><div class="umc-label">Diniyyah</div><div class="umc-sub">Madrasah</div></div>
      <div class="ust-menu-card au" onclick="nav('absen-malam')"><div class="umc-icon" style="background:rgba(79,70,229,0.12);color:#4f46e5"><i class="ph-duotone ph-moon-stars"></i></div><div class="umc-label">Kamar</div><div class="umc-sub">Asrama</div></div>
      ${thfCard}
      <div class="ust-menu-card au" onclick="nav('santri')"><div class="umc-icon" style="background:rgba(37,99,235,0.12);color:#2563eb"><i class="ph-duotone ph-users"></i></div><div class="umc-label">Data Santri</div><div class="umc-sub">Profil Santri</div></div>
      <div class="ust-menu-card au" onclick="nav('catatan-guru')"><div class="umc-icon" style="background:rgba(234,88,12,0.12);color:#ea580c"><i class="ph-duotone ph-notepad"></i></div><div class="umc-label">Catatan</div><div class="umc-sub">Guru</div></div>
      <div class="ust-menu-card au" onclick="nav('pelanggaran')"><div class="umc-icon" style="background:rgba(220,38,38,0.12);color:#dc2626"><i class="ph-duotone ph-warning-octagon"></i></div><div class="umc-label">Pelanggaran</div><div class="umc-sub">Santri</div></div>
    </div>`;
    startDashClock();
    return;
  }

  // -- Bendahara Dashboard --
  if (user.role === "bendahara") {
    $("main").innerHTML =
      `<div class="welcome au"><h3>Assalamu'alaikum, ${user.nama}</h3><p>${dateStr} | <span id="dashClock"></span> WIB</p></div>
    <div class="stat-grid" style="grid-template-columns:repeat(2,1fr)">
      <div class="stat-card c-green au"><div class="stat-info"><div class="sn">${d.total_santri || 0}</div><div class="sl">Total Santri</div></div><div class="si"><i class="ri-group-line"></i></div></div>
      <div class="stat-card c-blue au"><div class="stat-info"><div class="sn">${d.hadir_hari_ini || 0}</div><div class="sl">Hadir</div></div><div class="si"><i class="ri-checkbox-circle-line"></i></div></div>
    </div>
    <div class="section-title">KEUANGAN</div>
    <div class="dash-quick-grid au">
      <div class="dash-quick-card" onclick="nav('pembayaran')"><div class="dqc-icon" style="background:linear-gradient(135deg,#16a34a,#22c55e)"><i class="ri-money-dollar-circle-line"></i></div><div class="dqc-text"><div class="dqc-title">Pembayaran SPP</div><div class="dqc-sub">Kelola tagihan santri</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
      <div class="dash-quick-card" onclick="nav('catatan-bendahara')"><div class="dqc-icon" style="background:linear-gradient(135deg,#3b82f6,#60a5fa)"><i class="ri-book-3-line"></i></div><div class="dqc-text"><div class="dqc-title">Catatan Bendahara</div><div class="dqc-sub">Uang masuk & keluar</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
      <div class="dash-quick-card" onclick="nav('santri')"><div class="dqc-icon" style="background:var(--g4)"><i class="ri-graduation-cap-line"></i></div><div class="dqc-text"><div class="dqc-title">Data Santri</div><div class="dqc-sub">Lihat data santri</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
    </div>`;
    startDashClock();
    return;
  }
  // -- Keamanan Dashboard --
  if (user.role === "keamanan") {
    $("main").innerHTML =
      `<div class="welcome au"><h3>Assalamu'alaikum, ${user.nama}</h3><p>${dateStr} | <span id="dashClock"></span> WIB</p></div>
    <div class="section-title">KEDISIPLINAN</div>
    <div class="dash-quick-grid au">
      <div class="dash-quick-card" onclick="nav('catatan-guru')"><div class="dqc-icon" style="background:linear-gradient(135deg,#ec4899,#f472b6)"><i class="ri-sticky-note-line"></i></div><div class="dqc-text"><div class="dqc-title">Catatan Guru</div><div class="dqc-sub">Catatan perkembangan santri</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
      <div class="dash-quick-card" onclick="nav('pelanggaran')"><div class="dqc-icon" style="background:linear-gradient(135deg,#ef4444,#f87171)"><i class="ri-error-warning-line"></i></div><div class="dqc-text"><div class="dqc-title">Pelanggaran</div><div class="dqc-sub">Catat pelanggaran santri</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
      <div class="dash-quick-card" onclick="nav('perizinan')"><div class="dqc-icon" style="background:linear-gradient(135deg,#f59e0b,#fbbf24)"><i class="ri-pass-valid-line"></i></div><div class="dqc-text"><div class="dqc-title">Perizinan</div><div class="dqc-sub">Kelola izin santri</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
    </div>
    ${
      hasFeature("e_paket")
        ? `<div class="section-title">FITUR EXTRA</div>
    <div class="dash-quick-grid au">
      <div class="dash-quick-card" onclick="nav('e-paket')"><div class="dqc-icon" style="background:linear-gradient(135deg,#8b5cf6,#a78bfa)"><i class="ri-box-3-line"></i></div><div class="dqc-text"><div class="dqc-title">E-Paket</div><div class="dqc-sub">Terima & distribusi paket</div></div><i class="ri-arrow-right-s-line dqc-arrow"></i></div>
    </div>`
        : ""
    }`;
    startDashClock();
    return;
  }

  // -- Admin Dashboard (SIPP Style) --
  const hf = hasFeature;
  const totalAbsen =
    (d.hadir_hari_ini || 0) + (d.izin_sakit || 0) + (d.alfa || 0);
  const pctHadir =
    totalAbsen > 0
      ? Math.round(((d.hadir_hari_ini || 0) / totalAbsen) * 100)
      : 0;
  const t_santri = d.total_santri || 0;

  // Render Pill Cards for Aksi Cepat
  const pillCard = (id, ic, lb, cl) =>
    `<div class="pill-card" onclick="nav('${id}')"><div class="pill-icon" style="background:var(--pbg);color:var(--p)"><i class="${ic}" style="color:${cl}"></i></div><div class="pill-text">${lb}</div></div>`;

  let pillHtml = "";
  pillHtml += pillCard(
    "santri",
    "ri-user-add-line",
    "Tambah Santri",
    "#2563eb",
  );
  if (hf("absen_sekolah"))
    pillHtml += pillCard(
      "absen-sekolah",
      "ri-checkbox-circle-line",
      "Input Absensi",
      "#16a34a",
    );
  if (hf("keuangan"))
    pillHtml += pillCard(
      "pembayaran",
      "ri-wallet-3-line",
      "Pembayaran",
      "#8b5cf6",
    );
  pillHtml += pillCard(
    "catatan-guru",
    "ri-file-text-line",
    "Catatan Guru",
    "#f59e0b",
  );
  if (hf("kedisiplinan"))
    pillHtml += pillCard(
      "pelanggaran",
      "ri-shield-star-line",
      "Pelanggaran",
      "#ef4444",
    );

  // Stats for Pembayaran
  let t_tagihan = d.tagihan_bulan_ini || 0;
  let t_terkumpul = d.terkumpul_bulan_ini || 0;
  let pctBayar =
    t_tagihan > 0 ? Math.round((t_terkumpul / t_tagihan) * 100) : 0;
  if (pctBayar > 100) pctBayar = 100;

  let pctKamar =
    d.total_kamar > 0
      ? Math.round(((d.kamar_terisi || 0) / d.total_kamar) * 100)
      : 0;

  $("main").innerHTML = `
  <div style="margin-bottom:1.5rem">
    <h1 style="font-size:1.4rem;font-weight:800;color:var(--text);margin-bottom:.3rem">Beranda</h1>
    <p style="color:var(--t3);font-size:.85rem">Ringkasan kegiatan dan situasi pesantren hari ini.</p>
  </div>

  <div class="premium-hero au" style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:2rem">
    <div style="flex:1;min-width:300px;position:relative;z-index:1">
      <h3 style="font-size:1.6rem;font-weight:800;margin-bottom:0.8rem">Assalamu'alaikum, ${window.user?.nama || "Admin"} 👋</h3>
      <p style="font-size:0.9rem;opacity:0.9;max-width:400px;line-height:1.5;margin-bottom:1.5rem">Semoga setiap langkah kita hari ini menjadi bagian dari keberkahan dalam mendidik generasi Qur'ani.</p>
      <div style="display:flex;gap:1rem;font-size:0.8rem">
        <span style="display:flex;align-items:center;gap:0.4rem"><i class="ri-calendar-line"></i> ${dateStr}</span>
        <span style="display:flex;align-items:center;gap:0.4rem"><i class="ri-time-line"></i> <span id="dashClock"></span> WIB</span>
      </div>
    </div>
    
    <div style="flex:1;min-width:300px;background:rgba(255,255,255,0.1);padding:1.5rem;border-radius:20px;border:1px solid rgba(255,255,255,0.1);backdrop-filter:blur(10px);position:relative;z-index:1">
      <div style="font-size:0.7rem;font-weight:700;letter-spacing:0.05em;margin-bottom:1rem;opacity:0.8">SITUASI HARI INI</div>
      <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:1rem">
        <div><i class="ri-checkbox-circle-line" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">${pctHadir}%</div><div style="font-size:0.65rem;opacity:0.7">Kehadiran Santri</div></div>
        <div onclick="nav('pelanggaran')" style="cursor:pointer"><i class="ri-error-warning-line" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">${d.pelanggaran_pending || 0}</div><div style="font-size:0.65rem;opacity:0.7">Kasus Tertunda</div></div>
        <div onclick="nav('pembayaran')" style="cursor:pointer"><i class="ri-wallet-3-line" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">${pctBayar}%</div><div style="font-size:0.65rem;opacity:0.7">SPP Terkumpul</div></div>
        <div><i class="ri-calendar-check-line" style="font-size:1.5rem;opacity:0.8;margin-bottom:0.5rem;display:block"></i><div style="font-size:1.1rem;font-weight:700">12</div><div style="font-size:0.65rem;opacity:0.7">Kegiatan</div></div>
      </div>
    </div>
  </div>

  <div class="stat-grid au" style="display:grid;gap:1rem;margin-bottom:2rem">
    <div class="stat-card-clean"><div class="sc-icon" style="background:#eff6ff;color:#2563eb"><i class="ri-group-line"></i></div><div class="sc-body"><div class="sc-label">Total Santri</div><div class="sc-val">${t_santri}</div><div class="sc-sub" style="color:#2563eb">● Terdaftar aktif</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#ecfdf5;color:#16a34a"><i class="ri-checkbox-circle-line"></i></div><div class="sc-body"><div class="sc-label">Absensi Hari Ini</div><div class="sc-val">${pctHadir}%</div><div class="sc-sub">${d.hadir_hari_ini || 0} / ${t_santri} hadir</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#f5f3ff;color:#8b5cf6"><i class="ri-wallet-3-line"></i></div><div class="sc-body"><div class="sc-label">Pembayaran Bln Ini</div><div class="sc-val">Rp ${t_terkumpul.toLocaleString("id")}</div><div class="sc-sub">${pctBayar}% dari total tagihan</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#fffbeb;color:#d97706"><i class="ri-hotel-bed-line"></i></div><div class="sc-body"><div class="sc-label">Kamar Aktif</div><div class="sc-val">${d.total_kamar || 0}</div><div class="sc-sub">${pctKamar}% terisi</div></div></div>
    <div class="stat-card-clean"><div class="sc-icon" style="background:#fef2f2;color:#dc2626"><i class="ri-error-warning-line"></i></div><div class="sc-body"><div class="sc-label">Pelanggaran</div><div class="sc-val">${d.pelanggaran_pending || 0}</div><div class="sc-sub">${d.pelanggaran_pending > 0 ? "Perlu review" : "Tidak ada kasus"}</div></div></div>
  </div>

  <div class="section-title au" style="margin-bottom:1rem;font-size:0.8rem;letter-spacing:0.05em">AKSI CEPAT</div>
  <div class="pill-grid au">${pillHtml}</div>

  <div class="bottom-cols au">
    <!-- Kolom 1 -->
    <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02)">
      <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Perlu Ditindaklanjuti</h3>
      <div class="list-item" onclick="nav('absen-sekolah')" style="cursor:pointer">
        <div class="li-icon" style="background:#eff6ff;color:#2563eb"><i class="ri-group-line"></i></div>
        <div class="li-body"><div class="li-title">Santri belum absen hari ini</div><div class="li-sub">${d.belum_absen_hari_ini || 0} santri belum melakukan absensi</div></div>
        <div class="li-right li-badge">${d.belum_absen_hari_ini || 0}</div>
      </div>
      <div class="list-item" onclick="nav('pembayaran')" style="cursor:pointer">
        <div class="li-icon" style="background:#fffbeb;color:#d97706"><i class="ri-wallet-3-line"></i></div>
        <div class="li-body"><div class="li-title">Pembayaran tertunda</div><div class="li-sub">${d.tertunggak_count || 0} santri memiliki tunggakan bulan ini</div></div>
        <div class="li-right li-badge" style="background:rgba(217,119,6,0.1);color:#d97706">${d.tertunggak_count || 0}</div>
      </div>
      <div class="list-item" onclick="nav('pelanggaran')" style="cursor:pointer">
        <div class="li-icon" style="background:#fef2f2;color:#dc2626"><i class="ri-error-warning-line"></i></div>
        <div class="li-body"><div class="li-title">Pelanggaran menunggu review</div><div class="li-sub">${d.pelanggaran_pending || 0} kasus perlu ditinjau</div></div>
        <div class="li-right li-badge">${d.pelanggaran_pending || 0}</div>
      </div>
      <a href="#" onclick="nav('pelanggaran');return false;" style="display:block;margin-top:1.5rem;font-size:0.8rem;font-weight:700;color:var(--p);text-decoration:none">Lihat semua &rarr;</a>
    </div>

    <!-- Kolom 2 -->
    <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02)">
      <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Aktivitas Terbaru</h3>
      <div id="dashAktivitasList">
        ${
          (d.aktivitas_terbaru || [])
            .map((a) => {
              let ic = "ri-time-line",
                clBg = "#f5f3ff",
                clT = "#8b5cf6",
                path = "";
              if (a.type === "catatan") {
                ic = "ri-file-text-line";
                path = "catatan-guru";
              }
              if (a.type === "pelanggaran") {
                ic = "ri-error-warning-line";
                clBg = "#fef2f2";
                clT = "#dc2626";
                path = "pelanggaran";
              }
              if (a.type === "pembayaran") {
                ic = "ri-wallet-3-line";
                clBg = "#eff6ff";
                clT = "#2563eb";
                path = "pembayaran";
              }
              return `<div class="list-item" onclick="nav('${path}')" style="cursor:pointer">
            <div class="li-icon" style="background:${clBg};color:${clT}"><i class="${ic}"></i></div>
            <div class="li-body"><div class="li-title">${a.teks}</div><div class="li-sub" style="${a.type === "pembayaran" ? "color:#16a34a" : ""}">${a.sub}</div></div>
            <div class="li-right" style="color:var(--t3);font-weight:500;font-size:0.75rem">${a.waktu.substring(0, 10)}</div>
          </div>`;
            })
            .join("") ||
          '<div style="color:var(--t3);font-size:0.85rem;text-align:center;padding:1rem 0">Belum ada aktivitas</div>'
        }
      </div>
    </div>

    <!-- Kolom 3 -->
    <div style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02);flex:1">
        <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Kehadiran Hari Ini</h3>
        <div style="display:flex;align-items:center;gap:1.5rem">
          <div style="width:100px;height:100px;border-radius:50%;background:conic-gradient(#16a34a ${pctHadir}%, #e2e8f0 0);position:relative;display:flex;align-items:center;justify-content:center">
            <div style="width:75px;height:75px;background:#fff;border-radius:50%;display:flex;flex-direction:column;align-items:center;justify-content:center"><div style="font-size:1.1rem;font-weight:800">${pctHadir}%</div><div style="font-size:0.5rem;color:var(--t3)">hadir</div></div>
          </div>
          <div style="flex:1">
            <div style="display:flex;justify-content:space-between;margin-bottom:0.5rem;font-size:0.75rem"><span style="display:flex;align-items:center;gap:0.4rem"><span style="width:8px;height:8px;border-radius:50%;background:#16a34a"></span>Hadir</span><strong>${d.hadir_hari_ini || 0}</strong></div>
            <div style="display:flex;justify-content:space-between;margin-bottom:0.5rem;font-size:0.75rem"><span style="display:flex;align-items:center;gap:0.4rem"><span style="width:8px;height:8px;border-radius:50%;background:#f59e0b"></span>Izin/Sakit</span><strong>${d.izin_sakit || 0}</strong></div>
            <div style="display:flex;justify-content:space-between;font-size:0.75rem"><span style="display:flex;align-items:center;gap:0.4rem"><span style="width:8px;height:8px;border-radius:50%;background:#e2e8f0"></span>Alfa</span><strong>${d.alfa || 0}</strong></div>
          </div>
        </div>
      </div>
      <div class="card" style="padding:1.5rem;border-radius:24px;box-shadow:0 4px 20px rgba(0,0,0,0.02);flex:1">
        <h3 style="font-size:1.1rem;font-weight:800;margin-bottom:1.5rem">Status Pembayaran Bulan Ini</h3>
        <div style="display:flex;justify-content:space-between;margin-bottom:0.5rem;font-size:0.75rem">
          <div><span style="font-size:1.2rem;font-weight:800;color:var(--text);display:block">${pctBayar}%</span><span style="color:var(--t3)">Rp ${(t_terkumpul || 0).toLocaleString("id")} terkumpul</span></div>
          <div style="text-align:right">
            <div style="display:flex;align-items:center;gap:0.4rem;justify-content:flex-end;margin-bottom:0.3rem"><span style="width:8px;height:8px;border-radius:50%;background:#16a34a"></span>Lunas (${d.lunas_count || 0})</div>
            <div style="display:flex;align-items:center;gap:0.4rem;justify-content:flex-end"><span style="width:8px;height:8px;border-radius:50%;background:#ef4444"></span>Tertunggak (${d.tertunggak_count || 0})</div>
          </div>
        </div>
        <div style="width:100%;height:8px;background:#ef4444;border-radius:4px;overflow:hidden;margin-bottom:1rem"><div style="width:${pctBayar}%;height:100%;background:#16a34a"></div></div>
        <div style="display:flex;justify-content:space-between;font-size:0.7rem;color:var(--t3)">
          <div>Total Tagihan<br><strong style="color:var(--text)">Rp ${(t_tagihan || 0).toLocaleString("id")}</strong></div>
          <div style="text-align:right">Terkumpul<br><strong style="color:var(--text)">Rp ${(t_terkumpul || 0).toLocaleString("id")}</strong></div>
        </div>
      </div>
    </div>
  </div>

  ${jadwalHtml}
  `;
  startDashClock();
}
// Waktu page (server time + jadwal)
async function loadWaktu() {
  const now = new Date(Date.now() + 7 * 3600000);
  const timeStr = now.toISOString().slice(11, 16);
  const dateStr = now.toLocaleDateString("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });
  let jadwalHtml = "";
  try {
    const jadwal = await api("/api/jadwal-umum");
    const hariNow = now.toLocaleDateString("id-ID", { weekday: "long" });
    const todayJ = jadwal.filter((j) => j.hari === hariNow);
    jadwalHtml = todayJ
      .map(
        (
          j,
        ) => `<div style="display:flex;align-items:center;gap:.8rem;padding:.7rem 0;border-bottom:1px solid var(--border)">
      <div style="width:44px;height:44px;border-radius:13px;background:var(--greenbg);display:flex;align-items:center;justify-content:center"><i class="ri-time-line" style="color:var(--green);font-size:1.1rem"></i></div>
      <div style="flex:1"><div style="font-size:.85rem;font-weight:700">${j.nama}</div><div style="font-size:.7rem;color:var(--t3)">${j.jam_mulai} - ${j.jam_selesai}</div></div>
    </div>`,
      )
      .join("");
    if (!todayJ.length)
      jadwalHtml =
        '<div style="text-align:center;padding:1.5rem;color:var(--t3);font-size:.82rem"><i class="ri-calendar-close-line" style="font-size:1.5rem;display:block;margin-bottom:.3rem"></i>Tidak ada jadwal hari ini</div>';
  } catch (e) {
    jadwalHtml =
      '<div style="color:var(--t3);font-size:.82rem;padding:1rem;text-align:center">Gagal memuat jadwal</div>';
  }
  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-time-line"></i> Waktu</h2></div>
  <div class="card au" style="text-align:center;padding:2rem">
    <div class="time-wib" id="waktuClock">${timeStr}</div>
    <div class="time-wib-label">Waktu Indonesia Barat</div>
    <div style="margin-top:.5rem;font-size:.82rem;color:var(--t2)">${dateStr}</div>
  </div>
  <div class="card au"><h3><i class="ri-calendar-schedule-line"></i> Jadwal Hari Ini</h3>${jadwalHtml}</div>`;
  // Live clock
  clearInterval(_dashClockInterval);
  _dashClockInterval = setInterval(() => {
    const el = $("waktuClock");
    if (!el) return clearInterval(_dashClockInterval);
    el.textContent = new Date(Date.now() + 7 * 3600000)
      .toISOString()
      .slice(11, 16);
  }, 1000);
}
let _dashClockInterval = null;
function startDashClock() {
  clearInterval(_dashClockInterval);
  function updateClock() {
    const el = $("dashClock"),
      el2 = $("dashClockD");
    if (!el && !el2) return clearInterval(_dashClockInterval);
    const now = new Date(Date.now() + 7 * 3600000);
    const t = now.toISOString().slice(11, 16);
    if (el) el.textContent = t;
    if (el2) el2.textContent = t;
  }
  updateClock();
  _dashClockInterval = setInterval(updateClock, 1000);
}

// -- DESKTOP DASHBOARD: Santri Butuh Perhatian --
async function loadDashPerhatian() {
  const area = $("dashPerhatianArea");
  if (!area) return;
  try {
    const pelList = await api("/api/pelanggaran").catch(() => []);
    if (!Array.isArray(pelList)) {
      area.innerHTML = "";
      return;
    }
    // Group pelanggaran by santri
    const map = {};
    pelList.forEach((p) => {
      const k = p.santri_nama || "Unknown";
      if (!map[k])
        map[k] = { nama: k, count: 0, poin: 0, items: [], lastDate: "" };
      map[k].count++;
      map[k].poin += p.poin || 0;
      map[k].items.push(p);
      if (p.tanggal > map[k].lastDate) map[k].lastDate = p.tanggal;
    });
    const sorted = Object.values(map)
      .sort((a, b) => b.poin - a.poin)
      .slice(0, 10);
    if (!sorted.length) {
      area.innerHTML = "";
      return;
    }
    area.innerHTML = `<div class="card au" style="margin-bottom:.8rem">
      <h3 style="margin-bottom:.6rem"><i class="ri-alert-line" style="color:var(--amber)"></i> Santri Butuh Perhatian Khusus</h3>
      <p style="font-size:.75rem;color:var(--t3);margin-bottom:.8rem">Santri dengan pelanggaran terbanyak dan sering tidak hadir</p>
      <div class="table-wrap"><table>
        <tr><th>#</th><th>Nama Santri</th><th style="text-align:center">Pelanggaran</th><th style="text-align:center">Total Poin</th><th>Terakhir</th><th>Jenis Terakhir</th></tr>
        ${sorted
          .map((s, i) => {
            const last = s.items[s.items.length - 1] || {};
            const poinColor =
              s.poin >= 10
                ? "var(--red)"
                : s.poin >= 5
                  ? "var(--amber)"
                  : "var(--green)";
            return `<tr>
            <td>${i + 1}</td>
            <td><strong>${s.nama}</strong></td>
            <td style="text-align:center"><span class="badge-a">${s.count}x</span></td>
            <td style="text-align:center"><span style="font-weight:700;color:${poinColor}">${s.poin}</span></td>
            <td style="font-size:.78rem;color:var(--t3)">${s.lastDate || "-"}</td>
            <td style="font-size:.78rem">${last.jenis || "-"}</td>
          </tr>`;
          })
          .join("")}
      </table></div>
    </div>`;
  } catch (e) {
    area.innerHTML = "";
  }
}
// -- SANTRI PROFILE SEARCH (Admin Dashboard) --
let _dashSantriList = null;
async function searchSantriProfile(q) {
  const box = $("dashSantriResults");
  if (!box) return;
  if (!q || q.length < 2) {
    box.innerHTML = "";
    return;
  }
  if (!_dashSantriList)
    _dashSantriList = await api("/api/santri").catch(() => []);
  const results = _dashSantriList
    .filter((s) => s.nama.toLowerCase().includes(q.toLowerCase()))
    .slice(0, 8);
  box.innerHTML = results
    .map(
      (s) =>
        `<div onclick="openSantriProfile(${s.id})">${s.nama} <span style="font-size:.72rem;color:var(--t3)">${s.kamar_nama || ""}</span></div>`,
    )
    .join("");
}
async function openSantriProfile(sid) {
  $("dashSantriResults").innerHTML = "";
  $("dashSantriSearch").value = "";
  const area = $("santriProfileArea");
  area.innerHTML =
    '<p style="color:var(--t3);margin-top:.8rem"><i class="ri-loader-4-line"></i> Memuat data...</p>';
  try {
    const d = await api("/api/santri/" + sid + "/profile");
    const s = d.santri,
      pb = d.pembayaran,
      cg = d.catatan_guru || [],
      pl = d.pelanggaran || [];
    const fmtDate = (dt) => {
      if (!dt) return "-";
      const d = new Date(dt);
      return d.toLocaleDateString("id", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      });
    };
    area.innerHTML = `
    <div style="margin-top:1rem">
      <!-- Header Card -->
      <div style="background:linear-gradient(135deg,#1e40af,#7c3aed);border-radius:16px;padding:1.5rem;color:#fff;position:relative;overflow:hidden">
        <div style="position:absolute;top:-20px;right:-20px;width:120px;height:120px;background:rgba(255,255,255,.08);border-radius:50%"></div>
        <div style="position:absolute;bottom:-30px;right:40px;width:80px;height:80px;background:rgba(255,255,255,.05);border-radius:50%"></div>
        <div style="display:flex;align-items:center;gap:1rem;position:relative;z-index:1">
          <div style="width:56px;height:56px;background:rgba(255,255,255,.2);border-radius:14px;display:flex;align-items:center;justify-content:center;font-size:1.4rem;font-weight:700;backdrop-filter:blur(10px)">${s.nama.charAt(0).toUpperCase()}</div>
          <div style="flex:1">
            <h3 style="font-size:1.15rem;margin-bottom:.2rem">${s.nama}</h3>
            <div style="display:flex;gap:.5rem;flex-wrap:wrap;font-size:.75rem;opacity:.85">
              <span style="background:rgba(255,255,255,.15);padding:.15rem .5rem;border-radius:6px"><i class="ri-hotel-bed-line"></i> ${s.kamar_nama || "-"}</span>
              <span style="background:rgba(255,255,255,.15);padding:.15rem .5rem;border-radius:6px"><i class="ri-school-line"></i> ${s.kelas_sekolah || "-"}</span>
              <span style="background:rgba(255,255,255,.15);padding:.15rem .5rem;border-radius:6px"><i class="ri-book-2-line"></i> ${s.kelas_diniyyah_nama || s.kelas_diniyyah || "-"}</span>
            </div>
          </div>
          <span style="background:${s.status === "aktif" ? "rgba(34,197,94,.9)" : "rgba(239,68,68,.9)"};padding:.25rem .7rem;border-radius:8px;font-size:.72rem;font-weight:600">${s.status?.toUpperCase() || "-"}</span>
        </div>
      </div>

      <!-- Info Grid -->
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:.6rem;margin-top:.8rem">
        <div style="background:var(--card);border-radius:12px;padding:.8rem;border:1px solid var(--border);display:flex;align-items:center;gap:.6rem">
          <div style="width:36px;height:36px;background:rgba(59,130,246,.1);border-radius:10px;display:flex;align-items:center;justify-content:center"><i class="ri-parent-line" style="color:#3b82f6;font-size:1rem"></i></div>
          <div><div style="font-size:.68rem;color:var(--t3);text-transform:uppercase;letter-spacing:.5px">Wali</div><div style="font-size:.85rem;font-weight:600">${s.nama_wali || "-"}</div></div>
        </div>
        <div style="background:var(--card);border-radius:12px;padding:.8rem;border:1px solid var(--border);display:flex;align-items:center;gap:.6rem">
          <div style="width:36px;height:36px;background:rgba(22,163,74,.1);border-radius:10px;display:flex;align-items:center;justify-content:center"><i class="ri-phone-line" style="color:#16a34a;font-size:1rem"></i></div>
          <div><div style="font-size:.68rem;color:var(--t3);text-transform:uppercase;letter-spacing:.5px">No HP</div><div style="font-size:.85rem;font-weight:600">${s.no_hp || "-"}</div></div>
        </div>
        <div style="background:var(--card);border-radius:12px;padding:.8rem;border:1px solid var(--border);display:flex;align-items:center;gap:.6rem;grid-column:1/-1">
          <div style="width:36px;height:36px;background:rgba(139,92,246,.1);border-radius:10px;display:flex;align-items:center;justify-content:center"><i class="ri-map-pin-line" style="color:#8b5cf6;font-size:1rem"></i></div>
          <div><div style="font-size:.68rem;color:var(--t3);text-transform:uppercase;letter-spacing:.5px">Alamat</div><div style="font-size:.85rem;font-weight:600">${s.alamat || "-"}</div></div>
        </div>
      </div>

      <!-- Pembayaran -->
      <div style="background:var(--card);border-radius:12px;padding:1rem;border:1px solid var(--border);margin-top:.8rem">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:.6rem">
          <h4 style="display:flex;align-items:center;gap:.4rem;font-size:.9rem"><i class="ri-money-dollar-circle-line" style="color:var(--green)"></i> Pembayaran</h4>
          <span style="font-size:.78rem;color:var(--t3)">Tarif: <strong>Rp ${(pb.tarif_bulanan || 0).toLocaleString("id")}</strong>/bln</span>
        </div>
        ${
          pb.riwayat.length
            ? `<div class="table-wrap"><table style="font-size:.78rem">
          <tr><th>Bulan</th><th style="text-align:right">Dibayar</th><th style="text-align:right">Kekurangan</th><th style="text-align:center">Status</th></tr>
          ${pb.riwayat
            .map((r) => {
              const sc =
                r.status === "LUNAS"
                  ? "#16a34a"
                  : r.status === "KURANG"
                    ? "#f59e0b"
                    : "#ef4444";
              const bg =
                r.status === "LUNAS"
                  ? "rgba(22,163,74,.08)"
                  : r.status === "KURANG"
                    ? "rgba(245,158,11,.08)"
                    : "rgba(239,68,68,.08)";
              return `<tr><td><strong>${r.bulan}</strong></td><td style="text-align:right">Rp ${r.total_bayar.toLocaleString("id")}</td><td style="text-align:right">${r.kekurangan > 0 ? '<span style="color:#ef4444">Rp ' + r.kekurangan.toLocaleString("id") + "</span>" : "-"}</td><td style="text-align:center"><span style="display:inline-block;padding:.15rem .5rem;border-radius:6px;font-size:.7rem;font-weight:700;background:${bg};color:${sc}">${r.status}</span></td></tr>`;
            })
            .join("")}
        </table></div>`
            : '<div style="text-align:center;padding:1rem;color:var(--t3);font-size:.82rem"><i class="ri-inbox-line" style="font-size:1.5rem;display:block;margin-bottom:.3rem;opacity:.4"></i>Belum ada riwayat pembayaran</div>'
        }
      </div>

      <!-- Pelanggaran & Catatan -->
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:.8rem;margin-top:.8rem">
        <div style="background:var(--card);border-radius:12px;padding:1rem;border:1px solid var(--border)">
          <h4 style="display:flex;align-items:center;gap:.4rem;font-size:.9rem;margin-bottom:.6rem"><i class="ri-error-warning-line" style="color:var(--amber)"></i> Pelanggaran <span style="background:rgba(245,158,11,.1);color:var(--amber);padding:.1rem .45rem;border-radius:6px;font-size:.7rem;font-weight:700">${pl.length}</span></h4>
          ${
            pl.length
              ? '<div style="max-height:200px;overflow-y:auto">' +
                pl
                  .map(
                    (
                      p,
                    ) => `<div style="padding:.5rem 0;border-bottom:1px solid var(--border)">
            <div style="display:flex;justify-content:space-between;align-items:center"><span style="font-weight:600;color:var(--red);font-size:.82rem">${p.jenis}</span><span style="font-size:.65rem;color:var(--t3)">${fmtDate(p.created_at)}</span></div>
            <div style="font-size:.78rem;color:var(--t2);margin-top:.15rem">${p.keterangan || "-"}</div>
          </div>`,
                  )
                  .join("") +
                "</div>"
              : '<div style="text-align:center;padding:.8rem;color:var(--t3);font-size:.82rem"><i class="ri-shield-check-line" style="color:var(--green);font-size:1.3rem;display:block;margin-bottom:.3rem"></i>Tidak ada pelanggaran</div>'
          }
        </div>
        <div style="background:var(--card);border-radius:12px;padding:1rem;border:1px solid var(--border)">
          <h4 style="display:flex;align-items:center;gap:.4rem;font-size:.9rem;margin-bottom:.6rem"><i class="ri-sticky-note-line" style="color:var(--p)"></i> Catatan Guru <span style="background:rgba(59,130,246,.1);color:var(--p);padding:.1rem .45rem;border-radius:6px;font-size:.7rem;font-weight:700">${cg.length}</span></h4>
          ${
            cg.length
              ? '<div style="max-height:200px;overflow-y:auto">' +
                cg
                  .map(
                    (
                      c,
                    ) => `<div style="padding:.5rem 0;border-bottom:1px solid var(--border)">
            <div style="font-size:.82rem;color:var(--t1)">${c.catatan}</div>
            <div style="font-size:.65rem;color:var(--t3);margin-top:.15rem">${fmtDate(c.created_at)} - ${c.guru || "-"}</div>
          </div>`,
                  )
                  .join("") +
                "</div>"
              : '<div style="text-align:center;padding:.8rem;color:var(--t3);font-size:.82rem"><i class="ri-chat-check-line" style="color:var(--p);font-size:1.3rem;display:block;margin-bottom:.3rem;opacity:.5"></i>Belum ada catatan</div>'
          }
        </div>
      </div>

      <!-- Action Buttons -->
      <div style="display:flex;gap:.5rem;margin-top:.8rem;flex-wrap:wrap">
        <button class="btn btn-primary btn-sm" onclick="window._raportSantriId=${sid};nav('raport-absensi')"><i class="ri-file-chart-line"></i> Raport Absensi</button>
        <button class="btn btn-gold btn-sm" onclick="nav('raport-penilaian')"><i class="ri-award-line"></i> Raport Penilaian</button>
      </div>
    </div>`;
  } catch (e) {
    area.innerHTML =
      '<p style="color:var(--red);margin-top:.5rem">Error: ' +
      e.message +
      "</p>";
  }
}

// -- KAMAR --
let kamarMembersId = null;
async function loadKamar() {
  if (kamarMembersId) return loadKamarMembers(kamarMembersId);
  api("/api/kamar").then(list => {
    $("main").innerHTML = `
      <div class="page-header fade-up">
        <h2><i class="ri-home-5-line"></i> Kelola Kamar</h2>
        ${user.role !== "wali" ? '<button class="btn btn-primary btn-sm" onclick="showAddKamar()"><i class="ri-add-line"></i> Tambah</button>' : ""}
      </div>
      <div class="grid">
        ${list.map(k => `
        <div class="grid-card fade-up" style="padding:1.2rem; cursor:pointer;" onclick="openKamar(${k.id},'${k.nama.replace(/'/g, "\'")}')">
          <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:.8rem">
            <div class="modern-grid-icon" style="color:#059669; background:rgba(5,150,105,0.1)"><i class="ri-home-5-line"></i></div>
            ${user.role !== "wali" ? `<button class="btn btn-outline btn-sm" style="border:none; background:transparent; padding:4px" onclick="event.stopPropagation();showEditKamar(${k.id},'${k.nama.replace(/'/g, "\'")}',${k.kapasitas || 10})" title="Edit"><i class="ri-edit-line"></i></button>` : ''}
          </div>
          <div style="font-weight:700; font-size:1.05rem; color:var(--text); margin-bottom:.3rem">${k.nama}</div>
          <div style="font-size:.8rem; color:var(--t3); display:flex; gap:.6rem; align-items:center">
            <span><i class="ri-group-line"></i> ${k.jumlah_santri} / ${k.kapasitas} Santri</span>
          </div>
        </div>`).join('') || '<div style="grid-column:1/-1; text-align:center; padding:2rem; color:var(--t3)">Belum ada kamar</div>'}
      </div>
    `;
  });
}

async function openKamar(id, nama) {
  kamarMembersId = id;
  navStack.push("kamar");
  await loadKamarMembers(id);
}

async function loadKamarMembers(id) {
  const [members, kamarList] = await Promise.all([
    api("/api/kamar/" + id + "/members"),
    api("/api/santri"),
  ]);
  $("main").innerHTML =
    `<button class="back-btn" onclick="kamarMembersId=null;goBack()"><i class="ri-arrow-left-line"></i> Kembali</button>
    <div class="page-header"><h2><i class="ri-home-5-line"></i> Anggota Kamar</h2></div>
    ${user.role !== "wali" ? `<div style="margin-bottom:1rem"><button class="btn btn-primary btn-sm" onclick="showBulkAddSantri({title:'Tambah Santri ke Kamar',apiUrl:'/api/kamar/${id}/members/bulk',existingIds:[${members.map((m) => m.id).join(",")}],onDone:function(){loadKamarMembers(${id});}})"><i class="ri-user-add-line"></i> Tambah Santri</button></div>` : ""} 
    <div class="card"><div class="table-wrap"><table>
      <tr><th>Nama</th><th>Kelas Diniyyah</th>${user.role !== "wali" ? "<th>Aksi</th>" : ""}</tr>
      ${members
        .map(
          (m) => `<tr><td>${m.nama}</td><td>${m.kelas_diniyyah || "-"}</td>
        ${user.role !== "wali" ? `<td><button class="btn btn-danger btn-sm" onclick="removeKamarMember(${id},${m.id})">Pindah</button></td>` : ""}</tr>`,
        )
        .join("")}
      ${!members.length ? '<tr><td colspan="3" style="text-align:center;color:var(--text-dim)">Belum ada anggota</td></tr>' : ""}
    </table></div></div>`;
  window._allSantri = kamarList;
}

async function searchKamarSantri(kamarId) {
  const q = $("kamarSearch").value.toLowerCase().trim();
  const box = $("kamarSearchResults");
  if (!q) {
    box.innerHTML = "";
    return;
  }
  const results = (window._allSantri || [])
    .filter((s) => s.nama.toLowerCase().includes(q) && s.kamar_id !== kamarId)
    .slice(0, 8);
  box.innerHTML = results
    .map(
      (
        s,
      ) => `<div style="padding:.5rem .8rem;cursor:pointer;border:1px solid var(--border);border-radius:8px;margin-bottom:.3rem;background:var(--card);color:var(--text-muted);transition:.15s" onmouseover="this.style.borderColor='var(--primary)'" onmouseout="this.style.borderColor='var(--border)'" onclick="addKamarMember(${kamarId},${s.id})">
    ${s.nama} <span style="font-size:.8rem;color:var(--text-dim)">(${s.kamar_nama || "belum ada kamar"})</span></div>`,
    )
    .join("");
}

async function addKamarMember(kamarId, santriId) {
  await api("/api/santri/" + santriId, {
    method: "PUT",
    body: JSON.stringify({ kamar_id: kamarId }),
  });
  toast("Santri ditambahkan ke kamar");
  loadKamarMembers(kamarId);
}
async function removeKamarMember(kamarId, santriId) {
  await api("/api/santri/" + santriId, {
    method: "PUT",
    body: JSON.stringify({ kamar_id: null }),
  });
  toast("Santri dipindahkan");
  loadKamarMembers(kamarId);
}

function showAddKamar() {
  $("modal").innerHTML = `<h3><i class="ri-home-5-line"></i> Tambah Kamar</h3>
    <div class="fg"><label>Nama</label><input id="kNama" placeholder="Nama kamar"></div>
    <div class="fg"><label>Kapasitas</label><input id="kKap" type="number" value="10"></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="saveKamar()">Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveKamar() {
  const b = { nama: $("kNama").value, kapasitas: parseInt($("kKap").value) };
  if (!b.nama) return toast("Nama wajib");
  await api("/api/kamar", { method: "POST", body: JSON.stringify(b) });
  hideModal();
  toast("Ditambahkan");
  loadKamar();
}
function showEditKamar(id, nama, kapasitas) {
  $("modal").innerHTML = `<h3><i class="ri-edit-line"></i> Edit Kamar</h3>
    <div class="fg"><label>Nama</label><input id="kNama" value="${nama}"></div>
    <div class="fg"><label>Kapasitas</label><input id="kKap" type="number" value="${kapasitas}"></div>
    <div style="display:flex;gap:.5rem;margin-top:1.2rem">
      <button class="btn btn-primary" onclick="saveEditKamar(${id})">Simpan</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button></div>`;
  showModal();
}
async function saveEditKamar(id) {
  const b = { nama: $("kNama").value, kapasitas: parseInt($("kKap").value) };
  if (!b.nama) return toast("Nama wajib");
  await api("/api/kamar/" + id, { method: "PUT", body: JSON.stringify(b) });
  hideModal();
  toast("Diupdate");
  loadKamar();
}

// -- KEGIATAN & KELOMPOK (3-level drill-down) --
let _kegLevel = 1,
  _kegNama = null,
  _kegKelId = null,
  _kegKelNama = null;
async function loadKegiatan() {
  _kegLevel = 1; _kegNama = null; _kegKelId = null;
  api("/api/kegiatan").then(list => {
    var isAdm = ["admin", "superadmin"].includes(user.role);
    $("main").innerHTML = `
      <div class="page-header fade-up">
        <h2><i class="ri-book-open-line"></i> Kelola Kegiatan</h2>
        ${isAdm ? '<button class="btn btn-primary btn-sm" onclick="showAddKeg()"><i class="ri-add-line"></i> Tambah</button>' : ""}
      </div>
      <div class="grid">
        ${list.map(k => `
        <div class="grid-card fade-up" style="padding:1.2rem; cursor:pointer;" onclick="openKelKegiatan(${k.id},'${k.nama.replace(/'/g, "\'")}')">
          <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:.8rem">
            <div class="modern-grid-icon" style="color:#d97706; background:rgba(217,119,6,0.1)"><i class="ri-book-open-line"></i></div>
            ${isAdm ? `<button class="btn btn-outline btn-sm" style="border:none; background:transparent; padding:4px" onclick="event.stopPropagation();showEditKeg(${k.id},'${k.nama.replace(/'/g, "\'")}')" title="Edit"><i class="ri-edit-line"></i></button>` : ''}
          </div>
          <div style="font-weight:700; font-size:1.05rem; color:var(--text); margin-bottom:.3rem">${k.nama}</div>
          <div style="font-size:.8rem; color:var(--t3); display:flex; gap:.6rem; align-items:center">
            <span><i class="ri-folder-add-line"></i> ${k.jumlah_kelompok || 0} Kelompok</span>
          </div>
        </div>`).join('') || '<div style="grid-column:1/-1; text-align:center; padding:2rem; color:var(--t3)">Belum ada kegiatan</div>'}
      </div>
    `;
  });
}

async function showTunggakanModal(santriId, santriNama) {
  $("modal").innerHTML = `<div style="text-align:center;padding:2rem"><i class="ri-loader-4-line" style="animation:spin .8s linear infinite;font-size:2rem;color:var(--t3)"></i><p style="color:var(--t3)">Memuat data tagihan...</p></div>`;
  showModal();

  try {
    const data = await api("/api/wali/tunggakan/" + santriId);
    const tunggakan = data.tunggakan || [];
    window._pgSantriId = santriId;
    window._pgSantriNama = santriNama;
    window._pgTunggakan = tunggakan;

    if (tunggakan.length === 0) {
      $("modal").innerHTML = `<div style="text-align:center;padding:2rem"><i class="ri-checkbox-circle-line" style="font-size:3rem;color:var(--green)"></i><h3 style="margin-top:1rem">Tidak Ada Tagihan</h3><p style="color:var(--t3)">Alhamdulillah, semua tagihan untuk ${santriNama} sudah lunas.</p><button class="btn btn-outline" style="margin-top:1rem" onclick="hideModal()">Tutup</button></div>`;
      return;
    }

    let html = `<h3 style="margin-bottom:1rem;font-size:1.1rem"><i class="ri-bank-card-line"></i> Pembayaran Tagihan</h3>`;
    html += `<p style="font-size:.85rem;color:var(--t3);margin-bottom:1rem">Pilih metode pembayaran tagihan untuk <strong>${santriNama}</strong>:</p>`;

    html += `<div style="display:flex;gap:.5rem;margin-bottom:1rem;background:#f8fafc;padding:.5rem;border-radius:10px;border:1px solid var(--border)">
      <label style="flex:1;display:flex;align-items:center;gap:.5rem;cursor:pointer"><input type="radio" name="sppPayMode" value="bulan" checked onchange="toggleSPPMode()" style="accent-color:var(--green)"> Pilih Bulan</label>
      <label style="flex:1;display:flex;align-items:center;gap:.5rem;cursor:pointer"><input type="radio" name="sppPayMode" value="cicil" onchange="toggleSPPMode()" style="accent-color:var(--green)"> Nominal Bebas (Cicil)</label>
    </div>`;

    html += `<div id="pgBulanContainer" style="max-height:40vh;overflow-y:auto;border:1px solid var(--border);border-radius:10px;padding:.5rem;margin-bottom:1rem;background:#f8fafc">`;
    let maxCicil = 0;
    tunggakan.forEach((t, i) => {
      maxCicil += t.sisa;
      const bl = new Date(t.bulan + "-01").toLocaleDateString("id-ID", { month: "long", year: "numeric" });
      html += `<label style="display:flex;justify-content:space-between;align-items:center;padding:.7rem;border-bottom:1px solid var(--border);cursor:pointer;background:#fff;border-radius:8px;margin-bottom:.3rem">
        <div style="display:flex;align-items:center;gap:.7rem">
          <input type="checkbox" class="pg-chk" data-idx="${i}" onchange="updatePGTotal()" style="width:18px;height:18px;accent-color:var(--green)">
          <div>
            <div style="font-weight:700;font-size:.85rem">SPP ${bl}</div>
            <div style="font-size:.7rem;color:var(--t3)">Tagihan: Rp ${(t.tagihan || t.nominal || 0).toLocaleString("id-ID")}</div>
          </div>
        </div>
        <div style="font-weight:800;color:var(--red)">Rp ${t.sisa.toLocaleString("id-ID")}</div>
      </label>`;
    });
    html += `</div>`;

    html += `<div id="pgCicilContainer" style="display:none;margin-bottom:1rem;background:#f8fafc;padding:.8rem;border-radius:10px;border:1px solid var(--border)">
      <label style="font-size:.85rem;color:var(--t3);display:block;margin-bottom:.5rem">Masukkan Nominal (Max Rp ${maxCicil.toLocaleString("id-ID")})</label>
      <input type="number" id="pgCicilInput" class="input" style="width:100%" placeholder="Contoh: 50000" oninput="updatePGTotal()" max="${maxCicil}">
    </div>`;

    html += `<div id="pgSummary" style="display:none;background:rgba(59,130,246,.05);border:1px solid rgba(59,130,246,.2);padding:.8rem;border-radius:10px;margin-bottom:1rem">
      <div style="display:flex;justify-content:space-between;margin-bottom:.3rem;font-size:.8rem;color:var(--t3)"><span>Subtotal:</span><strong id="pgSubtotal">Rp 0</strong></div>
      <div style="display:flex;justify-content:space-between;margin-bottom:.3rem;font-size:.8rem;color:var(--t3)"><span>Biaya Admin:</span><strong id="pgFee">Gratis</strong></div>
      <div style="display:flex;justify-content:space-between;margin-top:.5rem;padding-top:.5rem;border-top:1px dashed var(--border);font-size:.95rem;font-weight:800"><span>Total Bayar:</span><strong id="pgTotal" style="color:var(--blue)">Rp 0</strong></div>
    </div>`;

    html += `<div id="pgMethodBtns" style="display:flex;flex-direction:column;gap:.4rem;margin-bottom:1.5rem"></div>`;

    html += `<div style="display:flex;gap:.5rem">
      <button id="pgPayBtn" class="btn btn-primary" style="flex:1;opacity:.5" disabled>Lengkapi Data Dulu</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;

    $("modal").innerHTML = html;
    updatePGTotal();
  } catch (e) {
    $("modal").innerHTML = `<div style="color:var(--red);padding:1rem;text-align:center"><i class="ri-error-warning-line" style="font-size:2rem;display:block;margin-bottom:.5rem"></i>Gagal memuat tagihan: ${e.message}<br><button class="btn btn-outline" style="margin-top:1rem" onclick="hideModal()">Tutup</button></div>`;
  }
}

function toggleSPPMode() {
  const mode = document.querySelector('input[name="sppPayMode"]:checked')?.value || "bulan";
  if (mode === "bulan") {
    $("pgBulanContainer").style.display = "block";
    $("pgCicilContainer").style.display = "none";
  } else {
    $("pgBulanContainer").style.display = "none";
    $("pgCicilContainer").style.display = "block";
  }
  updatePGTotal();
}

// Update total when checkboxes or inputs change
function updatePGTotal() {
  const mode = document.querySelector('input[name="sppPayMode"]:checked')?.value || "bulan";
  const data = window._pgTunggakan || [];
  let subtotal = 0, totalFee = 0;
  const selectedBulan = [];

  if (mode === "bulan") {
    const checks = document.querySelectorAll(".pg-chk:checked");
    checks.forEach((chk) => {
      const idx = parseInt(chk.dataset.idx);
      const t = data[idx];
      if (t) {
        subtotal += t.sisa;
        totalFee += t.fee || 0;
        selectedBulan.push(t.bulan);
      }
    });
  } else {
    const maxCicil = data.reduce((sum, t) => sum + t.sisa, 0);
    const inputVal = parseInt($("pgCicilInput")?.value) || 0;
    subtotal = Math.min(inputVal, maxCicil);
    if (inputVal > maxCicil) {
      $("pgCicilInput").value = maxCicil;
    }
  }

  const total = subtotal + totalFee;
  window._pgSelectedBulan = selectedBulan;
  window._pgSubtotal = subtotal;
  window._pgTotalFee = totalFee;
  window._pgGrandTotal = total;
  window._pgSelectedMethod = null;

  const summary = $("pgSummary");
  const btn = $("pgPayBtn");
  if (subtotal > 0) {
    summary.style.display = "block";
    $("pgSubtotal").textContent = "Rp " + subtotal.toLocaleString("id-ID");
    $("pgFee").textContent =
      totalFee > 0 ? "Rp " + totalFee.toLocaleString("id-ID") : "Gratis";
    $("pgTotal").textContent = "Rp " + total.toLocaleString("id-ID");

    // Show method buttons
    const ts = window._tenantSettings || {};
    const methods = $("pgMethodBtns");
    if (methods) {
      let mHtml = "";
      if (ts.transfer_bank_enabled) {
        const tbFee = ts.transfer_bank_fee_flat || 0;
        const tbFeeLabel =
          tbFee > 0 ? "Fee: Rp " + tbFee.toLocaleString("id-ID") : "GRATIS!";
        mHtml += `<label style="display:flex;align-items:center;gap:.5rem;padding:.55rem .7rem;border-radius:10px;cursor:pointer;border:1.5px solid var(--border);background:#f8fafc;transition:.2s" onclick="selectPayMethod('transfer_bank',this)">
          <input type="radio" name="pgMethod" value="transfer_bank" style="accent-color:var(--green);width:16px;height:16px">
          <div style="flex:1"><div style="font-weight:700;font-size:.82rem"><i class="ri-bank-line" style="color:var(--green)"></i> Transfer Bank</div>
          <div style="font-size:.7rem;color:var(--t3)">Ke rekening pesantren · ${tbFeeLabel}</div></div>
        </label>`;
      }
      if (ts.pg_provider) {
        const pgFeeTotal =
          Math.round(subtotal * ((ts.pg_fee_percent || 0) / 100)) +
          (ts.pg_fee_flat || 0);
        mHtml += `<label style="display:flex;align-items:center;gap:.5rem;padding:.55rem .7rem;border-radius:10px;cursor:pointer;border:1.5px solid var(--border);background:#f8fafc;transition:.2s" onclick="selectPayMethod('midtrans',this)">
          <input type="radio" name="pgMethod" value="midtrans" style="accent-color:var(--blue);width:16px;height:16px">
          <div style="flex:1"><div style="font-weight:700;font-size:.82rem"><i class="ri-secure-payment-line" style="color:var(--blue)"></i> Payment Gateway</div>
          <div style="font-size:.7rem;color:var(--t3)">Kartu kredit / VA · Fee: Rp ${pgFeeTotal.toLocaleString("id-ID")}</div></div>
        </label>`;
      }
      if (!mHtml) {
        mHtml = `<div style="font-size:.78rem;color:var(--red)"><i class="ri-error-warning-line"></i> Belum ada metode pembayaran aktif. Hubungi admin pesantren.</div>`;
      }
      methods.innerHTML = mHtml;
    }

    btn.disabled = true;
    btn.style.opacity = ".5";
  } else {
    summary.style.display = "none";
    btn.disabled = true;
    btn.style.opacity = ".5";
  }
}

function selectPayMethod(method, el) {
  window._pgSelectedMethod = method;
  // Highlight selected
  document.querySelectorAll("#pgMethodBtns label").forEach((l) => {
    l.style.borderColor = "var(--border)";
    l.style.background = "#f8fafc";
  });
  el.style.borderColor =
    method === "transfer_bank" ? "var(--green)" : "var(--blue)";
  el.style.background =
    method === "transfer_bank" ? "rgba(22,163,74,.06)" : "rgba(59,130,246,.06)";

  // Update fee display based on method
  const ts = window._tenantSettings || {};
  const subtotal = window._pgSubtotal || 0;
  let fee = 0;
  if (method === "transfer_bank") {
    fee = ts.transfer_bank_fee_flat || 0;
  } else {
    fee =
      Math.round(subtotal * ((ts.pg_fee_percent || 0) / 100)) +
      (ts.pg_fee_flat || 0);
  }
  $("pgFee").textContent =
    fee > 0 ? "Rp " + fee.toLocaleString("id-ID") : "Gratis";
  $("pgTotal").textContent = "Rp " + (subtotal + fee).toLocaleString("id-ID");

  const btn = $("pgPayBtn");
  btn.disabled = false;
  btn.style.opacity = "1";
  if (method === "transfer_bank") {
    btn.innerHTML = '<i class="ri-bank-line"></i> Transfer Bank';
    btn.onclick = () => processPaymentBankTransfer();
  } else {
    btn.innerHTML = '<i class="ri-secure-payment-line"></i> Bayar via Gateway';
    btn.onclick = () => processPayment();
  }
}

// ============================================
// WALI - PEMBAYARAN INSIDENTAL
// ============================================

async function showInsidentalWaliModal(santriId, santriNama) {
  $("modal").innerHTML = `<div style="text-align:center;padding:2rem"><i class="ri-loader-4-line" style="animation:spin .8s linear infinite;font-size:2rem;color:var(--t3)"></i><p style="color:var(--t3)">Memuat tagihan insidental...</p></div>`;
  showModal();

  try {
    const data = await api("/api/insidental/santri/" + santriId);
    const pendingIns = (data||[]).filter(t => t.status !== 'LUNAS');
    
    if (pendingIns.length === 0) {
      $("modal").innerHTML = `<div style="text-align:center;padding:2rem"><i class="ri-checkbox-circle-line" style="font-size:3rem;color:var(--green)"></i><h3 style="margin-top:1rem">Tidak Ada Tagihan</h3><p style="color:var(--t3)">Tagihan insidental untuk ${santriNama} sudah lunas.</p><button class="btn btn-outline" style="margin-top:1rem" onclick="hideModal()">Tutup</button></div>`;
      return;
    }

    window._pgSantriId = santriId;
    window._pgSantriNama = santriNama;
    window._pgInsidentalData = pendingIns;

    let html = `<h3 style="margin-bottom:1rem;font-size:1.1rem"><i class="ri-receipt-line"></i> Bayar Tagihan Insidental</h3>`;
    html += `<p style="font-size:.85rem;color:var(--t3);margin-bottom:1rem">Pilih tagihan yang ingin dibayar untuk <strong>${santriNama}</strong>:</p>`;

    html += `<div style="margin-bottom:1rem">
      <select id="pgInsSelect" class="input" style="width:100%" onchange="updatePGInsTotal()">
        <option value="">-- Pilih Tagihan --</option>`;
    pendingIns.forEach((t, i) => {
      html += `<option value="${i}">${t.nama} (Sisa: Rp ${t.sisa_tagihan.toLocaleString("id-ID")})</option>`;
    });
    html += `</select></div>`;

    html += `<div id="pgInsInputContainer" style="display:none;margin-bottom:1rem;background:#f8fafc;padding:.8rem;border-radius:10px;border:1px solid var(--border)">
      <label style="font-size:.85rem;color:var(--t3);display:block;margin-bottom:.5rem">Masukkan Nominal (Max Rp <span id="pgInsMaxCicil"></span>)</label>
      <input type="number" id="pgInsCicilInput" class="input" style="width:100%" placeholder="Contoh: 50000" oninput="updatePGInsTotal()">
    </div>`;

    html += `<div id="pgInsSummary" style="display:none;background:rgba(59,130,246,.05);border:1px solid rgba(59,130,246,.2);padding:.8rem;border-radius:10px;margin-bottom:1rem">
      <div style="display:flex;justify-content:space-between;margin-bottom:.3rem;font-size:.8rem;color:var(--t3)"><span>Subtotal:</span><strong id="pgInsSubtotal">Rp 0</strong></div>
      <div style="display:flex;justify-content:space-between;margin-bottom:.3rem;font-size:.8rem;color:var(--t3)"><span>Biaya Admin:</span><strong id="pgInsFee">Gratis</strong></div>
      <div style="display:flex;justify-content:space-between;margin-top:.5rem;padding-top:.5rem;border-top:1px dashed var(--border);font-size:.95rem;font-weight:800"><span>Total Bayar:</span><strong id="pgInsTotal" style="color:var(--blue)">Rp 0</strong></div>
    </div>`;

    html += `<div id="pgInsMethodBtns" style="display:flex;flex-direction:column;gap:.4rem;margin-bottom:1.5rem"></div>`;

    html += `<div style="display:flex;gap:.5rem">
      <button id="pgInsPayBtn" class="btn btn-primary" style="flex:1;opacity:.5" disabled>Lengkapi Data Dulu</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>`;

    $("modal").innerHTML = html;
  } catch (e) {
    $("modal").innerHTML = `<div style="color:var(--red);padding:1rem;text-align:center">Gagal memuat: ${e.message}<br><button class="btn btn-outline" style="margin-top:1rem" onclick="hideModal()">Tutup</button></div>`;
  }
}

function updatePGInsTotal() {
  const selIdx = $("pgInsSelect").value;
  const data = window._pgInsidentalData || [];
  
  if (selIdx === "") {
    $("pgInsInputContainer").style.display = "none";
    $("pgInsSummary").style.display = "none";
    $("pgInsPayBtn").disabled = true;
    $("pgInsPayBtn").style.opacity = ".5";
    $("pgInsMethodBtns").innerHTML = "";
    return;
  }

  const t = data[selIdx];
  $("pgInsInputContainer").style.display = "block";
  $("pgInsMaxCicil").textContent = t.sisa_tagihan.toLocaleString("id-ID");
  
  const inputEl = $("pgInsCicilInput");
  let inputVal = parseInt(inputEl.value) || 0;
  
  // Set default to max if empty
  if (inputEl.value === "") {
     inputVal = t.sisa_tagihan;
     inputEl.value = inputVal;
  }
  
  if (inputVal > t.sisa_tagihan) {
    inputVal = t.sisa_tagihan;
    inputEl.value = inputVal;
  }

  let subtotal = inputVal;
  
  window._pgInsSelectedId = t.id;
  window._pgInsSubtotal = subtotal;

  const ts = window._tenantSettings || {};
  let fee = 0;
  
  // Hanya support payment gateway
  if (ts.pg_provider) {
    fee = Math.round(subtotal * ((ts.pg_fee_percent || 0) / 100)) + (ts.pg_fee_flat || 0);
  }

  window._pgInsTotalFee = fee;
  const total = subtotal + fee;

  const summary = $("pgInsSummary");
  const btn = $("pgInsPayBtn");
  
  if (subtotal > 0) {
    summary.style.display = "block";
    $("pgInsSubtotal").textContent = "Rp " + subtotal.toLocaleString("id-ID");
    $("pgInsFee").textContent = fee > 0 ? "Rp " + fee.toLocaleString("id-ID") : "Gratis";
    $("pgInsTotal").textContent = "Rp " + total.toLocaleString("id-ID");

    const methods = $("pgInsMethodBtns");
    let mHtml = "";
    if (ts.transfer_bank_enabled) {
      const tbFee = ts.transfer_bank_fee_flat || 0;
      const tbFeeLabel = tbFee > 0 ? "Fee: Rp " + tbFee.toLocaleString("id-ID") : "GRATIS!";
      mHtml += `<label style="display:flex;align-items:center;gap:.5rem;padding:.55rem .7rem;border-radius:10px;cursor:pointer;border:1.5px solid var(--border);background:#f8fafc;transition:.2s" onclick="selectPayMethodInsidental('transfer_bank',this)">
        <input type="radio" name="pgInsMethod" value="transfer_bank" style="accent-color:var(--green);width:16px;height:16px">
        <div style="flex:1"><div style="font-weight:700;font-size:.82rem"><i class="ri-bank-line" style="color:var(--green)"></i> Transfer Bank</div>
        <div style="font-size:.7rem;color:var(--t3)">Ke rekening pesantren · ${tbFeeLabel}</div></div>
      </label>`;
    }
    if (ts.pg_provider) {
      mHtml += `<label style="display:flex;align-items:center;gap:.5rem;padding:.55rem .7rem;border-radius:10px;cursor:pointer;border:1.5px solid var(--border);background:#f8fafc;transition:.2s" onclick="selectPayMethodInsidental('midtrans',this)">
        <input type="radio" name="pgInsMethod" value="midtrans" style="accent-color:var(--blue);width:16px;height:16px">
        <div style="flex:1"><div style="font-weight:700;font-size:.82rem"><i class="ri-secure-payment-line" style="color:var(--blue)"></i> Payment Gateway</div>
        <div style="font-size:.7rem;color:var(--t3)">Kartu kredit / VA · Fee: Rp ${fee.toLocaleString("id-ID")}</div></div>
      </label>`;
    }
    if (!mHtml) {
      mHtml = `<div style="font-size:.78rem;color:var(--red)"><i class="ri-error-warning-line"></i> Belum ada metode pembayaran aktif.</div>`;
    }
    methods.innerHTML = mHtml;
  } else {
    summary.style.display = "none";
    btn.disabled = true;
    btn.style.opacity = ".5";
  }
}

async function processPaymentInsidental() {
  const santriId = window._pgSantriId;
  const tagihanId = window._pgInsSelectedId;
  const nominal = window._pgInsSubtotal;

  if (!tagihanId || nominal <= 0) return toast("Pilih tagihan dan masukkan nominal");

  const btn = $("pgInsPayBtn");
  btn.disabled = true;
  btn.innerHTML = '<i class="ri-loader-4-line" style="animation:spin .8s linear infinite"></i> Memproses...';

  try {
    const ts = window._tenantSettings || {};

    const res = await api("/api/wali/pay-insidental", {
      method: "POST",
      body: JSON.stringify({ 
        santri_id: santriId, 
        tagihan_santri_id: tagihanId, 
        nominal_bebas: nominal 
      }),
    });

    if (res.snap_token) {
      await loadSnapJS(ts.pg_client_key, ts.pg_is_production == 1);
      hideModal();
      const orderIdForVerify = res.order_id;
      window.snap.pay(res.snap_token, {
        onSuccess: function (result) {
          toast("✅ Pembayaran berhasil diproses");
          fetch("/api/payment/notification", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(result),
          }).catch(() => {});
          setTimeout(() => loadDashboard(), 1500);
        },
        onPending: function (result) {
          toast("⏳ Menunggu pembayaran selesai");
          fetch("/api/payment/notification", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(result),
          }).catch(() => {});
          setTimeout(() => loadDashboard(), 1500);
        },
        onError: function (result) {
          toast("❌ Pembayaran gagal");
        },
        onClose: function () {
          verifyAndRefresh(orderIdForVerify);
        },
      });
    } else if (res.redirect_url) {
      // Tripay redirect
      window.location.href = res.redirect_url;
    }
  } catch (e) {
    btn.disabled = false;
    btn.innerHTML = '<i class="ri-secure-payment-line"></i> Bayar via Gateway';
    toast("Error: " + e.message);
  }
}

// Process payment — call backend and open Snap popup (Midtrans)
async function processPayment() {
  const santriId = window._pgSantriId;
  const mode = document.querySelector('input[name="sppPayMode"]:checked')?.value || "bulan";
  const selectedBulan = window._pgSelectedBulan || [];
  
  if (mode === "bulan" && !selectedBulan.length) return toast("Pilih minimal 1 bulan");
  
  const subtotal = window._pgSubtotal || 0;
  if (mode === "cicil" && subtotal <= 0) return toast("Masukkan nominal yang valid");

  const btn = $("pgPayBtn");
  btn.disabled = true;
  btn.innerHTML =
    '<i class="ri-loader-4-line" style="animation:spin .8s linear infinite"></i> Memproses...';

  try {
    const ts = window._tenantSettings || {};

    const payload = { santri_id: santriId };
    if (mode === "bulan") {
      payload.bulan_list = selectedBulan;
    } else {
      payload.nominal_bebas = subtotal;
    }

    // Send transaction
    const res = await api("/api/wali/pay", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    if (res.snap_token) {
      // Load Snap.js if not loaded
      await loadSnapJS(ts.pg_client_key, ts.pg_is_production == 1);

      // Open Snap popup
      hideModal();
      const orderIdForVerify = res.order_id;
      window.snap.pay(res.snap_token, {
        onSuccess: function (result) {
          toast("✅ Pembayaran berhasil! Memproses...");
          fetch("/api/payment/notification", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(result),
          })
            .then((r) => r.json())
            .then((d) => {
              toast("✅ Pembayaran LUNAS! Jazakallah khairan");
              setTimeout(() => loadDashboard(), 1000);
            })
            .catch((e) => {
              toast("Pembayaran berhasil. Memuat ulang...");
              setTimeout(() => loadDashboard(), 2000);
            });
        },
        onPending: function (result) {
          toast("⏳ Menunggu pembayaran selesai.");
          fetch("/api/payment/notification", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(result),
          }).catch(() => {});
          setTimeout(() => loadDashboard(), 2000);
        },
        onError: function (result) {
          toast("❌ Pembayaran gagal. Silakan coba lagi.");
        },
        onClose: function () {
          verifyAndRefresh(orderIdForVerify);
        },
      });
    }
  } catch (e) {
    toast("Error: " + (e.message || "Gagal memproses pembayaran"));
    btn.disabled = false;
    btn.innerHTML = '<i class="ri-secure-payment-line"></i> Bayar Sekarang';
  }
}

// Process payment via Bank Transfer with unique code
async function processPaymentBankTransfer() {
  const santriId = window._pgSantriId;
  const santriNama = window._pgSantriNama;
  const selectedBulan = window._pgSelectedBulan || [];
  const subtotal = window._pgSubtotal || 0;
  if (!selectedBulan.length) return toast("Pilih minimal 1 bulan");

  const btn = $("pgPayBtn");
  btn.disabled = true;
  btn.innerHTML =
    '<i class="ri-loader-4-line" style="animation:spin .8s linear infinite"></i> Memproses...';

  try {
    const res = await api("/api/wali/bank-transfer", {
      method: "POST",
      body: JSON.stringify({
        santri_id: santriId,
        tipe: "spp",
        bulan_list: selectedBulan,
        nominal: subtotal,
      }),
    });

    const fmtR = (n) => "Rp " + (n || 0).toLocaleString("id-ID");
    const expiryDate = new Date(res.expiry).toLocaleString("id-ID", {
      day: "numeric",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

    $("modal").innerHTML = `<div style="max-width:420px;margin:0 auto">
      <h3 style="margin-bottom:.5rem;color:var(--green);display:flex;align-items:center;gap:.4rem"><i class="ri-checkbox-circle-line"></i> Instruksi Transfer</h3>
      <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">${santriNama} — Transfer ke rekening berikut:</p>

      <div style="background:linear-gradient(135deg,#f0fdf4,#ecfdf5);border:1.5px solid rgba(22,163,74,.2);border-radius:14px;padding:1rem;margin-bottom:.8rem">
        <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.6rem">
          <i class="ri-bank-line" style="font-size:1.2rem;color:var(--green)"></i>
          <strong style="font-size:.9rem">${res.rekening.bank}</strong>
        </div>
        <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.3rem">
          <span style="font-size:1.1rem;font-weight:800;letter-spacing:1px">${res.rekening.nomor}</span>
          <button onclick="navigator.clipboard.writeText('${res.rekening.nomor}');toast('📋 Nomor rekening disalin!')" style="background:var(--green);color:#fff;border:none;border-radius:6px;padding:.2rem .5rem;font-size:.68rem;cursor:pointer"><i class="ri-file-copy-line"></i> Salin</button>
        </div>
        <div style="font-size:.78rem;color:var(--t3)">a.n. ${res.rekening.atas_nama}</div>
      </div>

      <div style="background:linear-gradient(135deg,rgba(59,130,246,.06),rgba(99,102,241,.06));border:1.5px solid rgba(59,130,246,.2);border-radius:14px;padding:1rem;text-align:center;margin-bottom:.8rem">
        <div style="font-size:.72rem;color:var(--t3);margin-bottom:.3rem;text-transform:uppercase;letter-spacing:.5px">Transfer Tepat Sebesar</div>
        <div style="font-size:1.6rem;font-weight:900;color:#3b82f6;letter-spacing:1px">${fmtR(res.total_transfer)}</div>
        <div style="font-size:.72rem;color:var(--t3);margin-top:.3rem">Tagihan ${fmtR(res.nominal)}${res.fee > 0 ? " + biaya Rp " + res.fee.toLocaleString("id-ID") : ""} + kode unik <strong style="color:#3b82f6">+${res.kode_unik}</strong></div>
        <button onclick="navigator.clipboard.writeText('${res.total_transfer}');toast('📋 Nominal disalin!')" class="btn btn-sm" style="margin-top:.5rem;background:#3b82f6;color:#fff;font-size:.72rem"><i class="ri-file-copy-line"></i> Salin Nominal</button>
      </div>

      <div style="background:rgba(245,158,11,.06);border:1px solid rgba(245,158,11,.2);border-radius:10px;padding:.6rem .7rem;margin-bottom:.8rem">
        <div style="font-size:.75rem;color:#b45309;display:flex;align-items:start;gap:.3rem">
          <i class="ri-error-warning-line" style="flex-shrink:0;margin-top:1px"></i>
          <div>Transfer <strong>TEPAT</strong> nominal di atas! Berlaku <strong>${res.expiry_jam} jam</strong> (s/d ${expiryDate})</div>
        </div>
      </div>

      <div style="display:flex;gap:.5rem">
        <button class="btn btn-outline" onclick="hideModal()" style="flex:1"><i class="ri-check-line"></i> OK, Saya Akan Transfer</button>
      </div>
    </div>`;
  } catch (e) {
    toast("Error: " + (e.message || "Gagal membuat transfer"));
    btn.disabled = false;
    btn.innerHTML = '<i class="ri-bank-line"></i> Transfer Bank';
  }
}

// Fallback verify for when popup is closed without clear result
async function verifyAndRefresh(orderId) {
  try {
    const res = await api("/api/wali/verify-payment", {
      method: "POST",
      body: JSON.stringify({ order_id: orderId }),
    });
    if (res.transaction_status === "settlement") {
      toast("✅ Pembayaran LUNAS! Jazakallah khairan");
    } else if (res.transaction_status === "pending") {
      toast("⏳ Menunggu pembayaran selesai.");
    } else {
      toast("Status: " + (res.transaction_status || "mengecek..."));
    }
  } catch (e) {
    console.error("Verify error:", e);
  }
  setTimeout(() => loadDashboard(), 1000);
}

// ═══════════════════════════════════════════════════════════
// PEMBAYARAN WALI — Standalone page listing all children's SPP
// ═══════════════════════════════════════════════════════════
async function loadPembayaranWali() {
  const [anak, btData] = await Promise.all([
    api("/api/wali/anak"),
    api("/api/wali/bank-transfers").catch(() => ({ transfers: [] })),
  ]);
  const fmtR = (n) => "Rp " + (n || 0).toLocaleString("id-ID");
  const pendingTransfers = (btData.transfers || []).filter(
    (t) => t.status === "pending",
  );

  let cards = "";

  // Show pending bank transfers at top
  if (pendingTransfers.length > 0) {
    cards += `<div class="card au" style="margin-bottom:.8rem;background:linear-gradient(135deg,rgba(245,158,11,.05),rgba(251,191,36,.05));border:1.5px solid rgba(245,158,11,.2)">
      <h3 style="font-size:.88rem;margin-bottom:.5rem;display:flex;align-items:center;gap:.3rem"><i class="ri-time-line" style="color:#f59e0b"></i> Transfer Menunggu Konfirmasi</h3>
      ${pendingTransfers
        .map((t) => {
          const rek = btData.rekening || {};
          return `<div style="padding:.5rem .7rem;background:#fff;border-radius:10px;border:1px solid var(--border);margin-bottom:.4rem">
          <div style="display:flex;justify-content:space-between;align-items:center">
            <div>
              <div style="font-weight:700;font-size:.82rem">${t.santri_nama}</div>
              <div style="font-size:.72rem;color:var(--t3)">${t.tipe === "spp" ? "SPP" : "Topup Sangu"} · ${new Date(t.created_at).toLocaleDateString("id-ID")}</div>
            </div>
            <div style="text-align:right">
              <div style="font-weight:800;font-size:.9rem;color:#3b82f6">${fmtR(t.total_transfer)}</div>
              <span style="background:rgba(245,158,11,.12);color:#b45309;padding:.1rem .4rem;border-radius:5px;font-size:.62rem;font-weight:700">PENDING</span>
            </div>
          </div>
          <div style="font-size:.7rem;color:var(--t3);margin-top:.3rem">${rek.bank || ""} ${rek.nomor || ""} · a.n. ${rek.atas_nama || ""} · kode unik +${t.kode_unik}</div>
        </div>`;
        })
        .join("")}
    </div>`;
  }

  for (const a of anak) {
    try {
      const data = await api("/api/wali/tunggakan/" + a.id);
      const tunggakan = data.tunggakan || [];
      const totalTunggakan = tunggakan.reduce((s, t) => s + t.sisa, 0);

      cards += `<div class="card au" style="margin-bottom:.8rem">
        <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:.6rem">
          <div>
            <div style="font-weight:700;font-size:.95rem">${a.nama}</div>
            <div style="font-size:.75rem;color:var(--t3)">${a.kamar_nama || "-"} · ${a.kelas_diniyyah || "-"}</div>
          </div>
          ${
            totalTunggakan > 0
              ? `<span style="background:var(--red);color:#fff;padding:.15rem .5rem;border-radius:8px;font-size:.72rem;font-weight:700">${tunggakan.length} bulan</span>`
              : '<span style="background:var(--green);color:#fff;padding:.15rem .5rem;border-radius:8px;font-size:.72rem;font-weight:700">LUNAS</span>'
          }
        </div>
        ${
          totalTunggakan > 0
            ? `
          <div style="font-size:.82rem;color:var(--red);font-weight:600;margin-bottom:.5rem">
            Total Tunggakan SPP: ${fmtR(totalTunggakan)}
          </div>
          <div style="display:flex;flex-wrap:wrap;gap:.3rem;margin-bottom:.6rem">
            ${tunggakan
              .map((t) => {
                const bl = new Date(t.bulan + "-01").toLocaleDateString(
                  "id-ID",
                  { month: "short", year: "2-digit" },
                );
                return `<span style="padding:.15rem .4rem;border-radius:6px;font-size:.68rem;font-weight:600;background:rgba(239,68,68,.1);color:var(--red);border:1px solid rgba(239,68,68,.2)">${bl}</span>`;
              })
              .join("")}
          </div>
          <button class="btn btn-primary btn-sm" style="margin-bottom:1rem" onclick="showTunggakanModal(${a.id},'${a.nama.replace(/'/g, "\\\\'")}')">
            <i class="ri-bank-card-2-line"></i> Bayar SPP Sekarang
          </button>`
            : `
          <div style="text-align:center;padding:.5rem;color:var(--green);font-size:.85rem;margin-bottom:1rem">
            <i class="ri-checkbox-circle-line" style="font-size:1.5rem;display:block;margin-bottom:.2rem"></i>
            Alhamdulillah, semua tagihan SPP lunas!
          </div>`
        }`;

      // Tagihan Insidental
      try {
        const insidental = await api("/api/insidental/santri/" + a.id);
        if (insidental && insidental.length > 0) {
           cards += `<div style="border-top:1px dashed var(--border);padding-top:.8rem;margin-top:.5rem">
              <div style="font-size:.82rem;color:var(--amber);font-weight:600;margin-bottom:.5rem"><i class="ri-receipt-line"></i> Tagihan Insidental:</div>
              ${insidental.map(t => {
                 const isLunas = t.status === 'LUNAS';
                 const bg = isLunas ? '#dcfce7' : '#fff9c4';
                 const val = isLunas ? '<span style="color:var(--green);font-weight:700">LUNAS</span>' : `<strong style="color:var(--red)">${fmtR(t.sisa_tagihan)}</strong>`;
                 return `<div style="display:flex;justify-content:space-between;align-items:center;background:${bg};padding:.4rem .6rem;border-radius:6px;margin-bottom:.4rem;font-size:.8rem">
                   <span>${t.nama}</span>
                   ${val}
                 </div>`;
              }).join('')}`;
           
           const pendingIns = insidental.filter(t => t.status !== 'LUNAS');
           if (pendingIns.length > 0) {
             cards += `<button class="btn btn-gold btn-sm" style="margin-top:.5rem" onclick="showInsidentalWaliModal(${a.id}, '${a.nama.replace(/'/g, "\\\\'")}')">Bayar Insidental</button>`;
           }
           cards += `</div>`;
        }
      } catch (err) {}
      
      cards += `</div>`;
    } catch (e) {
      cards += `<div class="card" style="margin-bottom:.8rem"><p style="color:var(--red)">${a.nama}: ${e.message}</p></div>`;
    }
  }

  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-bank-card-line"></i> Pembayaran SPP</h2></div>
    ${!anak.length ? '<div class="card"><p style="color:var(--t3);text-align:center">Belum ada data anak.</p></div>' : cards}`;
}

// ═══════════════════════════════════════════════════════════
// BANK TRANSFERS — Admin/Bendahara confirmation page
// ═══════════════════════════════════════════════════════════
let _btFilter = "pending";
async function loadBankTransfers() {
  const [transfers, countRes] = await Promise.all([
    api("/api/bank-transfers?status=" + _btFilter),
    api("/api/bank-transfers/count-pending").catch(() => ({ count: 0 })),
  ]);
  const fmtR = (n) => "Rp " + (n || 0).toLocaleString("id-ID");
  const pending = countRes.count || 0;

  const tabs = ["pending", "confirmed", "expired", "rejected", "all"]
    .map((s) => {
      const labels = {
        pending: `Pending (${pending})`,
        confirmed: "Dikonfirmasi",
        expired: "Kadaluarsa",
        rejected: "Ditolak",
        all: "Semua",
      };
      return `<button class="btn btn-sm ${_btFilter === s ? "btn-primary" : "btn-outline"}" onclick="_btFilter='${s}';loadBankTransfers()" style="white-space:nowrap;flex-shrink:0">${labels[s]}</button>`;
    })
    .join("");

  let rows = "";
  if (!transfers.length) {
    rows = `<div class="card" style="text-align:center;padding:2rem;color:var(--t3)"><i class="ri-inbox-line" style="font-size:2rem;display:block;margin-bottom:.5rem"></i>Tidak ada transfer ${_btFilter}</div>`;
  } else {
    rows = transfers
      .map((t) => {
        const isPending = t.status === "pending";
        const tipeLabel = t.tipe === "spp" ? "SPP" : (t.tipe === "insidental" ? "Insidental" : "Topup Sangu");
        const tipeColor = t.tipe === "spp" ? "var(--green)" : (t.tipe === "insidental" ? "var(--amber)" : "var(--blue)");
        const statusBadge =
          {
            pending:
              '<span style="background:rgba(245,158,11,.12);color:#b45309;padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:700">⏳ Pending</span>',
            confirmed:
              '<span style="background:var(--greenbg);color:var(--green);padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:700">✅ Dikonfirmasi</span>',
            expired:
              '<span style="background:rgba(107,114,128,.1);color:#6b7280;padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:700">⏰ Expired</span>',
            rejected:
              '<span style="background:var(--redbg);color:var(--red);padding:.15rem .4rem;border-radius:6px;font-size:.65rem;font-weight:700">❌ Ditolak</span>',
          }[t.status] || t.status;

        let bulanInfo = "";
        if (t.bulan_list && t.bulan_list !== "[]") {
          try {
            const bulanArr = JSON.parse(t.bulan_list);
            bulanInfo = bulanArr
              .map((b) => {
                const d = new Date(b + "-01");
                return d.toLocaleDateString("id-ID", {
                  month: "short",
                  year: "2-digit",
                });
              })
              .join(", ");
          } catch (e) {
            bulanInfo = t.bulan_list;
          }
        }

        const confirmedInfo = t.confirmed_at
          ? `<div style="font-size:.68rem;color:var(--green);margin-top:.2rem">Dikonfirmasi: ${new Date(t.confirmed_at).toLocaleString("id-ID")}${t.confirmed_by === 0 ? " (Moota Auto)" : ""}</div>`
          : "";

        return `<div class="card au" style="margin-bottom:.5rem">
        <div style="display:flex;justify-content:space-between;align-items:start">
          <div style="flex:1">
            <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.3rem">
              <strong style="font-size:.88rem">${t.santri_nama}</strong>
              <span style="background:${tipeColor};color:#fff;padding:.1rem .35rem;border-radius:5px;font-size:.6rem;font-weight:700">${tipeLabel}</span>
              ${statusBadge}
            </div>
            ${bulanInfo ? `<div style="font-size:.72rem;color:var(--t3)">Bulan: ${bulanInfo}</div>` : ""}
            <div style="font-size:.78rem;margin-top:.2rem">Transfer: <strong style="color:#3b82f6;font-size:.88rem">${fmtR(t.total_transfer)}</strong> <span style="font-size:.7rem;color:var(--t3)">(tagihan ${fmtR(t.nominal)}${t.fee > 0 ? " + fee " + fmtR(t.fee) : ""} + kode <span style="color:#3b82f6;font-weight:700">+${t.kode_unik}</span>)</span></div>
            <div style="font-size:.68rem;color:var(--t3);margin-top:.15rem">Dibuat: ${new Date(t.created_at).toLocaleString("id-ID")}</div>
            ${confirmedInfo}
            ${t.moota_mutation_id ? `<div style="font-size:.65rem;color:var(--accent);margin-top:.1rem">Moota ID: ${t.moota_mutation_id}</div>` : ""}
          </div>
          ${
            isPending
              ? `<div style="display:flex;flex-direction:column;gap:.3rem;flex-shrink:0;margin-left:.5rem">
            <button class="btn btn-sm" style="background:var(--green);color:#fff;font-size:.72rem" onclick="confirmBankTransfer(${t.id})"><i class="ri-checkbox-circle-line"></i> Konfirmasi</button>
            <button class="btn btn-danger btn-sm" style="font-size:.72rem" onclick="rejectBankTransfer(${t.id})"><i class="ri-close-circle-line"></i> Tolak</button>
          </div>`
              : ""
          }
        </div>
      </div>`;
      })
      .join("");
  }

  $("main").innerHTML = `<div class="page-header au">
    <h2><i class="ri-bank-line"></i> Konfirmasi Transfer ${pending > 0 ? '<span style="background:var(--red);color:#fff;padding:.1rem .45rem;border-radius:8px;font-size:.7rem;font-weight:700;margin-left:.3rem">' + pending + "</span>" : ""}</h2>
  </div>
  <div style="display:flex;gap:.3rem;overflow-x:auto;margin-bottom:.8rem;padding:.2rem 0">${tabs}</div>
  ${rows}`;
}

async function confirmBankTransfer(id) {
  if (
    !confirm("Yakin konfirmasi transfer ini? Pembayaran akan dicatat otomatis.")
  )
    return;
  try {
    const res = await api("/api/bank-transfers/" + id + "/confirm", {
      method: "POST",
    });
    toast("✅ " + (res.message || "Dikonfirmasi"));
    loadBankTransfers();
  } catch (e) {
    toast("Error: " + e.message);
  }
}

async function rejectBankTransfer(id) {
  if (!confirm("Yakin tolak transfer ini?")) return;
  try {
    const res = await api("/api/bank-transfers/" + id + "/reject", {
      method: "POST",
    });
    toast(res.message || "Ditolak");
    loadBankTransfers();
  } catch (e) {
    toast("Error: " + e.message);
  }
}

// ═══════════════════════════════════════════════════════════
// REKAP TUNGGAKAN — Admin/Bendahara grid (per-santri × per-month)
// ═══════════════════════════════════════════════════════════
async function loadRekapTunggakan() {
  const now = new Date(Date.now() + 7 * 3600000);
  const tahun = now.getFullYear();

  $("main").innerHTML =
    `<div class="page-header au"><h2><i class="ri-file-warning-line"></i> Rekap Tunggakan SPP</h2></div>
    <div class="card au" style="margin-bottom:.8rem">
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;align-items:end">
        <div class="fg" style="flex:1;min-width:100px"><label>Tahun</label>
          <select id="rekapTunggakanTahun" onchange="filterRekapTunggakan()">
            <option value="${tahun}">${tahun}</option>
            <option value="${tahun - 1}">${tahun - 1}</option>
            <option value="${tahun - 2}">${tahun - 2}</option>
          </select>
        </div>
        <button class="btn btn-primary btn-sm" onclick="filterRekapTunggakan()"><i class="ri-filter-3-line"></i> Filter</button>
        <button class="btn btn-outline btn-sm" id="btnExportTunggakan" onclick="exportTunggakan()"><i class="ri-file-excel-2-line"></i> Export Excel</button>
      </div>
    </div>
    <div id="rekapTunggakanResult"><p style="color:var(--t3)">Memuat data...</p></div>`;
  filterRekapTunggakan();
}

async function filterRekapTunggakan() {
  const tahun = $("rekapTunggakanTahun")?.value || new Date().getFullYear();
  const area = $("rekapTunggakanResult");
  area.innerHTML =
    '<div class="card"><p style="color:var(--t3)"><i class="ri-loader-4-line"></i> Memuat rekap tunggakan...</p></div>';

  try {
    const res = await api("/api/pembayaran/rekap-tunggakan?tahun=" + tahun);
    const data = res.data || [];
    const months = res.months || [];
    const monthLabels = months.map((m) => {
      const d = new Date(m + "-01");
      return d.toLocaleDateString("id-ID", { month: "short" });
    });

    window._rekapTunggakanData = { data, months, monthLabels, tahun };

    if (!data.length) {
      area.innerHTML =
        '<div class="card"><p style="color:var(--t3);text-align:center">Tidak ada data santri aktif.</p></div>';
      return;
    }

    // Summary
    let totalTunggakan = 0,
      totalLunas = 0;
    data.forEach((s) =>
      months.forEach((m) => {
        if (s.status[m] === "lunas") totalLunas++;
        else totalTunggakan++;
      }),
    );

    area.innerHTML = `
      <div style="display:flex;gap:.8rem;flex-wrap:wrap;margin-bottom:.8rem">
        <div class="card au" style="flex:1;min-width:140px;text-align:center;background:linear-gradient(135deg,rgba(22,163,74,.08),rgba(16,185,129,.05));border:1px solid rgba(22,163,74,.15)">
          <div style="font-size:1.6rem;font-weight:800;color:var(--green)">${totalLunas}</div>
          <div style="font-size:.78rem;color:var(--t2)">✅ Lunas</div>
        </div>
        <div class="card au" style="flex:1;min-width:140px;text-align:center;background:linear-gradient(135deg,rgba(239,68,68,.08),rgba(239,68,68,.05));border:1px solid rgba(239,68,68,.15)">
          <div style="font-size:1.6rem;font-weight:800;color:var(--red)">${totalTunggakan}</div>
          <div style="font-size:.78rem;color:var(--t2)">❌ Belum Bayar</div>
        </div>
        <div class="card au" style="flex:1;min-width:140px;text-align:center">
          <div style="font-size:1.6rem;font-weight:800;color:var(--accent)">${data.length}</div>
          <div style="font-size:.78rem;color:var(--t2)">👤 Santri</div>
        </div>
      </div>
      <div class="card"><div class="table-wrap"><table style="font-size:.78rem">
        <tr><th style="width:35px;text-align:center">No</th><th>Nama Santri</th>
          ${monthLabels.map((l) => '<th style="text-align:center;width:50px">' + l + "</th>").join("")}
        </tr>
        ${data
          .map(
            (
              s,
              i,
            ) => `<tr><td style="text-align:center">${i + 1}</td><td style="white-space:nowrap">${s.nama}</td>
          ${months
            .map((m) => {
              const st = s.status[m];
              if (st === "lunas")
                return '<td style="text-align:center"><span style="color:var(--green);font-weight:700">✅</span></td>';
              if (st === "kurang")
                return '<td style="text-align:center"><span style="color:#f59e0b;font-weight:700">⚠️</span></td>';
              return '<td style="text-align:center"><span style="color:var(--red);font-weight:700">❌</span></td>';
            })
            .join("")}
        </tr>`,
          )
          .join("")}
      </table></div></div>`;
  } catch (e) {
    area.innerHTML =
      '<div class="card"><p style="color:var(--red)">Error: ' +
      e.message +
      "</p></div>";
  }
}

function exportTunggakan() {
  const d = window._rekapTunggakanData;
  if (!d) return toast("Tidak ada data");
  const tahun = d.tahun;
  window.open(
    "/api/pembayaran/rekap-tunggakan/export-excel?tahun=" +
      tahun +
      "&token=" +
      encodeURIComponent(localStorage.getItem("token")),
    "_blank",
  );
}
// --- Tagihan Insidental ---
async function loadInsidental() {
  $('main').innerHTML = `
  <div class="page-header fade-up">
    <h2><i class="ri-receipt-line"></i> Tagihan Insidental</h2>
  </div>
  <div class="card fade-up" style="margin-bottom:1rem">
    <div style="display:flex;gap:0.5rem;flex-wrap:wrap">
      <div class="fg" style="flex:1;min-width:200px">
        <label>Cari Santri</label>
        <div style="position:relative">
          <input type="text" id="insSantriSearch" placeholder="Ketik nama santri..." autocomplete="off" oninput="searchInsSantri(this.value)">
          <div id="insSearchResults" class="search-results" style="max-height:200px;overflow-y:auto;position:absolute;top:100%;left:0;right:0;background:#fff;z-index:10;border:1px solid var(--border);border-radius:0 0 8px 8px;display:none;"></div>
        </div>
        <input type="hidden" id="insSantriId">
      </div>
      <button class="btn btn-primary" style="margin-top:23px" onclick="loadInsidentalSantri()"><i class="ri-search-line"></i> Tampilkan</button>
      <button class="btn btn-success" style="margin-top:23px" onclick="showBulkAssignInsidental()"><i class="ri-group-line"></i> Penetapan Massal</button>
      <button class="btn btn-outline" style="margin-top:23px" onclick="manageJenisInsidental()"><i class="ri-settings-3-line"></i> Master Jenis</button>
      <button class="btn btn-outline" style="margin-top:23px" onclick="showRekapInsidental()"><i class="ri-list-check-2"></i> Daftar/Rekap</button>
    </div>
  </div>
  <div id="insResult"></div>
  `;
  if (!window._santriFullList || window._santriFullList.length === 0) {
    try { window._santriFullList = await api('/api/santri'); } catch (e) {}
  }
}

window.searchInsSantri = function(q) {
  const res = $('insSearchResults');
  if (!q) { res.style.display='none'; return; }
  const match = (window._santriFullList||[]).filter(s => s.nama.toLowerCase().includes(q.toLowerCase())).slice(0, 10);
  if (!match.length) { res.style.display='none'; return; }
  res.innerHTML = match.map(s => `<div style="padding:0.5rem 0.8rem;border-bottom:1px solid var(--border);cursor:pointer" onclick="pickInsSantri(${s.id}, '${s.nama.replace(/'/g,"\\'").replace(/"/g,"&quot;")}')">${s.nama} - ${s.kamar_nama||'-'}</div>`).join('');
  res.style.display = 'block';
};

window.pickInsSantri = function(id, nama) {
  $('insSantriId').value = id;
  $('insSantriSearch').value = nama;
  $('insSearchResults').style.display = 'none';
};

window.loadInsidentalSantri = async function() {
  const sid = $('insSantriId').value;
  if (!sid) return toast('Pilih santri terlebih dahulu');
  $('insResult').innerHTML = '<div class="card"><p style="color:var(--t3)"><i class="ri-loader-4-line"></i> Memuat...</p></div>';
  try {
    const [tagihan, riwayat] = await Promise.all([
      api('/api/insidental/santri/' + sid),
      api('/api/insidental/santri/' + sid + '/riwayat')
    ]);
    
    let html = `<div class="card fade-up">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
        <h3>Daftar Tagihan</h3>
        <button class="btn btn-primary btn-sm" onclick="showAddTagihanInsidental(${sid})"><i class="ri-add-line"></i> Buat Tagihan</button>
      </div>
      <div class="table-wrap">
        <table>
          <tr><th>Jenis Tagihan</th><th>Nominal</th><th>Sisa Tagihan</th><th>Status</th><th>Aksi</th></tr>
          ${(tagihan||[]).map(t => {
            const isLunas = t.status === 'LUNAS';
            return `<tr>
              <td><strong>${t.nama}</strong></td>
              <td>${fmtRp(t.nominal)}</td>
              <td style="color:${isLunas?'var(--green)':'var(--red)'}">${fmtRp(t.sisa_tagihan)}</td>
              <td><span class="badge-${isLunas?'h':'a'}">${t.status}</span></td>
              <td>
                ${!isLunas ? `<button class="btn btn-primary btn-sm" onclick="showBayarInsidental(${t.id}, ${t.sisa_tagihan}, '${t.nama.replace(/'/g,"\\'")}')">Bayar</button>` : '-'}
              </td>
            </tr>`;
          }).join('')}
          ${(!tagihan || tagihan.length===0) ? '<tr><td colspan="5" style="text-align:center;color:var(--t3)">Belum ada tagihan insidental</td></tr>' : ''}
        </table>
      </div>
    </div>`;
    
    html += `<div class="card fade-up" style="margin-top:1rem">
      <h3>Riwayat Pembayaran</h3>
      <div class="table-wrap" style="margin-top:1rem">
        <table>
          <tr><th>Tanggal</th><th>Tagihan</th><th>Nominal Dibayar</th><th>Metode</th></tr>
          ${(riwayat||[]).map(r => `<tr>
            <td>${(r.tanggal||'').replace('T',' ').slice(0,16)}</td>
            <td>${r.nama}</td>
            <td style="color:var(--green)">+${fmtRp(r.nominal_dibayar)}</td>
            <td>${r.metode||'-'}</td>
          </tr>`).join('')}
          ${(!riwayat || riwayat.length===0) ? '<tr><td colspan="4" style="text-align:center;color:var(--t3)">Belum ada riwayat</td></tr>' : ''}
        </table>
      </div>
    </div>`;
    $('insResult').innerHTML = html;
  } catch (e) {
    $('insResult').innerHTML = '<div class="card"><p style="color:var(--red)">Error: ' + e.message + '</p></div>';
  }
};

window.manageJenisInsidental = async function() {
  $('modal').innerHTML = `
    <h3 style="margin-bottom:1rem">Master Jenis Tagihan Insidental</h3>
    <div id="jenisInsList"><p style="color:var(--t3)">Memuat...</p></div>
    <div style="margin-top:1rem;display:flex;gap:0.5rem">
      <button class="btn btn-outline btn-sm" onclick="hideModal()">Tutup</button>
    </div>
  `;
  showModal();
  try {
    const list = await api('/api/insidental/jenis');
    let html = `<div style="display:flex;gap:0.5rem;margin-bottom:1rem">
      <input type="text" id="newJenisNama" placeholder="Nama Jenis Baru" style="flex:1;padding:0.4rem;border:1px solid var(--border);border-radius:4px">
      <input type="number" id="newJenisNominal" placeholder="Nominal Default (Rp)" style="flex:1;padding:0.4rem;border:1px solid var(--border);border-radius:4px">
      <button class="btn btn-primary btn-sm" onclick="addJenisInsidental()">Tambah</button>
    </div>`;
    html += `<div class="table-wrap"><table>
      <tr><th>Nama Jenis</th><th>Nominal Default</th><th>Status</th><th>Aksi</th></tr>
      ${(list||[]).map(j => `<tr>
        <td>${j.nama}</td>
        <td>Rp ${parseInt(j.nominal).toLocaleString('id-ID')}</td>
        <td>${j.aktif?'Aktif':'Nonaktif'}</td>
        <td>
           <button class="btn btn-outline btn-sm" onclick="toggleJenisInsidental(${j.id}, ${!j.aktif})">${j.aktif?'Nonaktifkan':'Aktifkan'}</button>
        </td>
      </tr>`).join('')}
    </table></div>`;
    $('jenisInsList').innerHTML = html;
  } catch (e) {
    $('jenisInsList').innerHTML = `<p style="color:var(--red)">${e.message}</p>`;
  }
};

window.addJenisInsidental = async function() {
  const nama = $('newJenisNama').value;
  const nom = parseInt($('newJenisNominal').value) || 0;
  if (!nama || nom <= 0) return toast('Nama dan nominal wajib diisi!');
  try {
    await api('/api/insidental/jenis', { method:'POST', body: JSON.stringify({ nama: nama, nominal: nom }) });
    toast('Berhasil ditambahkan');
    manageJenisInsidental();
  } catch (e) { toast('Error: '+e.message); }
};

window.toggleJenisInsidental = async function(id, toAktif) {
  try {
    await api('/api/insidental/jenis/'+id, { method:'PUT', body: JSON.stringify({ aktif: toAktif }) });
    manageJenisInsidental();
  } catch (e) { toast('Error: '+e.message); }
};

window.showRekapInsidental = async function() {
  $('insResult').innerHTML = '<div class="card fade-up"><h3><i class="ri-loader-4-line"></i> Memuat jenis tagihan...</h3></div>';
  try {
    const list = await api('/api/insidental/jenis');
    let html = `<div class="card fade-up">
      <h3 style="margin-bottom:1rem"><i class="ri-list-check-2"></i> Rekap Tagihan Insidental</h3>
      <div style="display:flex;gap:0.5rem;margin-bottom:1rem">
        <select id="rekapJenisId" style="flex:1;padding:0.5rem;border:1px solid var(--border);border-radius:6px">
          <option value="">-- Pilih Jenis Tagihan --</option>
          ${(list||[]).map(j => `<option value="${j.id}">${j.nama}</option>`).join('')}
        </select>
        <button class="btn btn-primary" onclick="loadRekapInsidental()"><i class="ri-search-line"></i> Lihat Rekap</button>
      </div>
      <div id="rekapTableBox"></div>
    </div>`;
    $('insResult').innerHTML = html;
  } catch(e) {
    $('insResult').innerHTML = `<div class="card"><p style="color:var(--red)">${e.message}</p></div>`;
  }
};

window.loadRekapInsidental = async function() {
  const jid = $('rekapJenisId').value;
  if (!jid) return toast("Silakan pilih jenis tagihan");
  $('rekapTableBox').innerHTML = '<p style="color:var(--t3)">Memuat data...</p>';
  try {
    const data = await api('/api/insidental/rekap?tagihan_id=' + jid);
    const arr = Array.isArray(data) ? data : [];
    let html = `<div class="table-wrap"><table>
      <tr><th>Nama Santri</th><th>Kamar</th><th style="text-align:right">Total Tagihan</th><th style="text-align:right">Dibayar</th><th style="text-align:right">Sisa</th><th style="text-align:center">Status</th></tr>
      ${arr.map(s => {
        const isLunas = s.status === 'LUNAS';
        const badge = isLunas ? 'badge-h' : (s.status === 'KURANG' ? 'badge-i' : 'badge-a');
        return `<tr>
          <td>${s.nama}</td>
          <td>${s.kamar_nama||'-'}</td>
          <td style="text-align:right">${fmtRp(s.nominal)}</td>
          <td style="text-align:right">${fmtRp(s.nominal_dibayar)}</td>
          <td style="text-align:right;color:${isLunas?'inherit':'var(--red)'};font-weight:${isLunas?'normal':'700'}">${fmtRp(s.sisa)}</td>
          <td style="text-align:center"><span class="${badge}">${s.status}</span></td>
        </tr>`;
      }).join('')}
    </table></div>`;
    if(arr.length === 0) html = '<p style="color:var(--t3)">Tidak ada data untuk tagihan ini.</p>';
    $('rekapTableBox').innerHTML = html;
  } catch(e) {
    $('rekapTableBox').innerHTML = `<p style="color:var(--red)">${e.message}</p>`;
  }
};

window.showAddTagihanInsidental = async function(santriId) {
  $('modal').innerHTML = `<h3><i class="ri-loader-4-line"></i> Memuat jenis tagihan...</h3>`;
  showModal();
  try {
    const list = await api('/api/insidental/jenis');
    const aktifList = (list||[]).filter(x=>x.aktif);
    if (!aktifList.length) {
      $('modal').innerHTML = `<div style="padding:1rem">Tidak ada jenis tagihan aktif. Buat di Master Jenis. <br><br><button class="btn btn-outline" onclick="hideModal()">Tutup</button></div>`;
      return;
    }
    
    $('modal').innerHTML = `
      <h3 style="margin-bottom:1rem">Buat Tagihan Insidental</h3>
      <div class="fg">
        <label>Jenis Tagihan</label>
        <select id="atiJenis">
          ${aktifList.map(x=>`<option value="${x.id}">${x.nama}</option>`).join('')}
        </select>
      </div>
      <div class="fg">
        <label>Nominal Tagihan (Rp)</label>
        <input type="number" id="atiNominal" placeholder="Cth: 500000">
      </div>
      <div style="display:flex;gap:0.5rem;margin-top:1rem">
        <button class="btn btn-primary" onclick="doAddTagihanInsidental(${santriId})">Simpan Tagihan</button>
        <button class="btn btn-outline" onclick="hideModal()">Batal</button>
      </div>
    `;
  } catch(e) { toast(e.message); hideModal(); }
};

window.doAddTagihanInsidental = async function(santriId) {
  const jid = parseInt($('atiJenis').value);
  const nom = parseInt($('atiNominal').value);
  if (!jid || !nom) return toast('Mohon isi lengkap');
  
  try {
    await api('/api/insidental/santri', {
      method: 'POST',
      body: JSON.stringify({
        santri_id: santriId,
        tagihan_id: jid,
        nominal: nom
      })
    });
    toast('Tagihan berhasil dibuat');
    hideModal();
    loadInsidentalSantri();
  } catch (e) { toast('Error: ' + e.message); }
};

window.showBulkAssignInsidental = async function() {
  $('main').innerHTML = `
    <div class="page-header fade-up" style="display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem;justify-content:space-between">
      <div style="display:flex;align-items:center;gap:1rem;">
        <button class="btn btn-outline btn-sm" onclick="loadInsidental()"><i class="ri-arrow-left-line"></i> Kembali</button>
        <h2 style="margin:0"><i class="ri-group-line"></i> Penetapan Tagihan Massal</h2>
      </div>
      <button class="btn btn-primary" onclick="processBulkAssign()"><i class="ri-save-line"></i> Simpan Tagihan</button>
    </div>
    
    <div class="card fade-up">
      <div style="display:flex;gap:1rem;margin:1rem 0;flex-wrap:wrap">
        <div class="fg" style="flex:1;min-width:200px">
          <label>Pilih Jenis Tagihan</label>
          <select id="bulkJenisTagihan">
            <option value="">-- Pilih Jenis --</option>
          </select>
        </div>
        <div class="fg" style="flex:1;min-width:200px">
          <label>Set Nominal Massal</label>
          <div style="display:flex;gap:0.5rem">
            <input type="number" id="bulkNominalDefault" placeholder="Misal: 1000000">
            <button class="btn btn-outline" onclick="applyBulkNominal()">Terapkan</button>
          </div>
        </div>
      </div>
      <div style="display:flex;gap:1rem;margin:1rem 0;flex-wrap:wrap">
        <div class="fg" style="flex:1;min-width:200px">
          <label>Filter Kamar</label>
          <select id="bulkKamarFilter" onchange="renderBulkList()">
            <option value="">Semua Kamar</option>
          </select>
        </div>
        <div class="fg" style="flex:1;min-width:200px">
          <label>Cari Nama Santri</label>
          <input type="text" id="bulkSearchSantri" placeholder="Ketik nama santri..." onkeyup="renderBulkList()">
        </div>
      </div>
      
      <div style="display:flex;justify-content:space-between;align-items:center;margin:1rem 0 0.5rem">
        <label style="font-weight:bold;cursor:pointer"><input type="checkbox" id="bulkCheckAll" onchange="toggleBulkCheckAll(this.checked)" checked> Pilih Semua yang Tampil</label>
        <span id="bulkCountInfo"></span>
      </div>
      
      <div id="bulkSantriList" style="border:1px solid var(--border);border-radius:8px;background:#f8fafc;max-height: 60vh;overflow-y:auto;">
        <!-- Rendered by JS -->
      </div>
    </div>
  `;

  // Load Jenis
  try {
    const jenis = await api('/api/insidental/jenis');
    $('bulkJenisTagihan').innerHTML += jenis.filter(j=>j.aktif).map(j => `<option value="${j.id}" data-nom="${j.nominal}">${j.nama} - Rp ${parseInt(j.nominal).toLocaleString('id-ID')}</option>`).join('');
  } catch(e) {}
  
  // Load Santri
  if (!window._santriFullList) {
    try { window._santriFullList = await api('/api/santri'); } catch(e){}
  }
  
  const kamars = [...new Set((window._santriFullList||[]).map(s=>s.kamar_nama).filter(Boolean))].sort();
  $('bulkKamarFilter').innerHTML += kamars.map(k=>`<option value="${k}">${k}</option>`).join('');
  
  window.bulkSantriChecked = {};
  window.bulkSantriNominal = {};
  (window._santriFullList||[]).forEach(s => {
    window.bulkSantriChecked[s.id] = true;
    window.bulkSantriNominal[s.id] = 0;
  });
  
  $('bulkJenisTagihan').addEventListener('change', function(e) {
    const opt = this.options[this.selectedIndex];
    if(opt && opt.dataset.nom) {
       $('bulkNominalDefault').value = opt.dataset.nom;
       applyBulkNominal();
    }
  });

  renderBulkList();
};

window.renderBulkList = function() {
  const kFilter = $('bulkKamarFilter').value;
  const qFilter = ($('bulkSearchSantri').value || '').toLowerCase();
  
  let santris = window._santriFullList || [];
  if (kFilter) santris = santris.filter(s => s.kamar_nama === kFilter);
  if (qFilter) santris = santris.filter(s => s.nama.toLowerCase().includes(qFilter));
  
  $('bulkCountInfo').innerText = `Total Tampil: ${santris.length} Santri`;
  let html = '<table class="table" style="margin:0;font-size:0.85rem"><thead><tr><th style="width:50px"></th><th>Nama Santri</th><th>Kamar</th><th style="width:200px">Nominal Tagihan</th></tr></thead><tbody>';
  santris.forEach(s => {
    const isChecked = window.bulkSantriChecked[s.id] ? 'checked' : '';
    const nom = window.bulkSantriNominal[s.id] || 0;
    html += `
      <tr>
        <td style="text-align:center"><input type="checkbox" class="cb-santri" value="${s.id}" ${isChecked} onchange="window.bulkSantriChecked[${s.id}]=this.checked"></td>
        <td>${s.nama}</td>
        <td>${s.kamar_nama||'-'}</td>
        <td>
          <input type="number" class="nom-santri" style="width:100%;padding:4px" value="${nom}" onchange="window.bulkSantriNominal[${s.id}]=this.value">
        </td>
      </tr>
    `;
  });
  html += '</tbody></table>';
  $('bulkSantriList').innerHTML = html;
};

window.toggleBulkCheckAll = function(isChecked) {
  const kFilter = $('bulkKamarFilter').value;
  const qFilter = ($('bulkSearchSantri').value || '').toLowerCase();
  
  let santris = window._santriFullList || [];
  if (kFilter) santris = santris.filter(s => s.kamar_nama === kFilter);
  if (qFilter) santris = santris.filter(s => s.nama.toLowerCase().includes(qFilter));
  
  santris.forEach(s => { window.bulkSantriChecked[s.id] = isChecked; });
  renderBulkList();
};

window.applyBulkNominal = function() {
  const nom = $('bulkNominalDefault').value || 0;
  const kFilter = $('bulkKamarFilter').value;
  const qFilter = ($('bulkSearchSantri').value || '').toLowerCase();
  
  let santris = window._santriFullList || [];
  if (kFilter) santris = santris.filter(s => s.kamar_nama === kFilter);
  if (qFilter) santris = santris.filter(s => s.nama.toLowerCase().includes(qFilter));
  
  santris.forEach(s => { window.bulkSantriNominal[s.id] = nom; });
  renderBulkList();
  toast('Nominal diterapkan ke tabel di bawah. Silakan klik "Simpan Tagihan" di bagian atas untuk memproses.');
};

window.processBulkAssign = async function() {
  const tid = $('bulkJenisTagihan').value;
  if(!tid) return toast("Pilih jenis tagihan terlebih dahulu!");
  
  let list = [];
  (window._santriFullList||[]).forEach(s => {
     if(window.bulkSantriChecked[s.id] && window.bulkSantriNominal[s.id] > 0) {
       list.push({ santri_id: s.id, nominal: parseInt(window.bulkSantriNominal[s.id]) });
     }
  });
  
  if(list.length === 0) return toast("Tidak ada santri yang ditagihkan!");
  if(!confirm(`Anda akan menetapkan tagihan massal untuk ${list.length} santri. Lanjutkan?`)) return;
  
  try {
    await api('/api/insidental/bulk', {
      method: 'POST',
      body: JSON.stringify({ tagihan_id: parseInt(tid), santri_list: list })
    });
    toast('Berhasil menetapkan tagihan massal!');
    loadInsidental();
  } catch(e) { toast('Error: ' + e.message); }
};

window.showBayarInsidental = function(idTagihanSantri, sisa, nama) {
  $('modal').innerHTML = `
    <h3 style="margin-bottom:1rem">Bayar: ${nama}</h3>
    <div class="fg">
      <label>Sisa Tagihan</label>
      <input type="text" value="${fmtRp(sisa)}" disabled style="background:#f1f5f9;color:var(--t2)">
    </div>
    <div class="fg">
      <label>Nominal Bayar (Rp)</label>
      <input type="number" id="pbiNominal" value="${sisa}">
    </div>
    <div class="fg">
      <label>Keterangan (Opsional)</label>
      <input type="text" id="pbiKet" placeholder="Keterangan pembayaran...">
    </div>
    <div style="display:flex;gap:0.5rem;margin-top:1rem">
      <button class="btn btn-primary" onclick="doBayarInsidental(${idTagihanSantri}, ${sisa})">Proses Bayar</button>
      <button class="btn btn-outline" onclick="hideModal()">Batal</button>
    </div>
  `;
  showModal();
};

window.doBayarInsidental = async function(id, maks) {
  const n = parseInt($('pbiNominal').value);
  const k = $('pbiKet').value;
  if (!n || n <= 0) return toast('Nominal tidak valid');
  if (n > maks) return toast('Nominal melebihi sisa tagihan');
  
  try {
    await api('/api/insidental/bayar', {
      method: 'POST',
      body: JSON.stringify({
        tagihan_santri_id: id,
        nominal_dibayar: n,
        keterangan: k
      })
    });
    toast('Pembayaran berhasil');
    hideModal();
    loadInsidentalSantri();
  } catch(e) { toast('Error: ' + e.message); }
};

function selectPayMethodInsidental(method, el) {
  window._pgInsSelectedMethod = method;
  document.querySelectorAll("#pgInsMethodBtns label").forEach((l) => {
    l.style.borderColor = "var(--border)";
    l.style.background = "#f8fafc";
  });
  el.style.borderColor = method === "transfer_bank" ? "var(--green)" : "var(--blue)";
  el.style.background = method === "transfer_bank" ? "rgba(22,163,74,.06)" : "rgba(59,130,246,.06)";

  const ts = window._tenantSettings || {};
  const subtotal = window._pgInsSubtotal || 0;
  let fee = 0;
  if (method === "transfer_bank") {
    fee = ts.transfer_bank_fee_flat || 0;
  } else {
    fee = Math.round(subtotal * ((ts.pg_fee_percent || 0) / 100)) + (ts.pg_fee_flat || 0);
  }
  
  $("pgInsFee").textContent = fee > 0 ? "Rp " + fee.toLocaleString("id-ID") : "Gratis";
  $("pgInsTotal").textContent = "Rp " + (subtotal + fee).toLocaleString("id-ID");

  const btn = $("pgInsPayBtn");
  btn.disabled = false;
  btn.style.opacity = "1";
  if (method === "transfer_bank") {
    btn.innerHTML = '<i class="ri-bank-line"></i> Transfer Bank';
    btn.onclick = () => processPaymentInsidentalBankTransfer();
  } else {
    btn.innerHTML = '<i class="ri-secure-payment-line"></i> Bayar via Gateway';
    btn.onclick = () => processPaymentInsidental();
  }
}

async function processPaymentInsidentalBankTransfer() {
  const santriId = window._pgSantriId;
  const tagihanId = window._pgInsSelectedId;
  const nominal = window._pgInsSubtotal;
  const santriNama = window._pgSantriNama;

  if (!tagihanId || nominal <= 0) return toast("Pilih tagihan dan masukkan nominal");

  const btn = $("pgInsPayBtn");
  btn.disabled = true;
  btn.innerHTML = '<i class="ri-loader-4-line" style="animation:spin .8s linear infinite"></i> Memproses...';

  try {
    const res = await api("/api/wali/bank-transfer", {
      method: "POST",
      body: JSON.stringify({
        santri_id: santriId,
        tipe: "insidental",
        tagihan_santri_id: parseInt(tagihanId),
        nominal: nominal,
      }),
    });

    const expiryDate = new Date(res.expiry).toLocaleString("id-ID", {
      day: "numeric", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit",
    });

    $("modal").innerHTML = `<div style="max-width:420px;margin:0 auto">
      <h3 style="margin-bottom:.5rem;color:var(--green);display:flex;align-items:center;gap:.4rem"><i class="ri-checkbox-circle-line"></i> Instruksi Transfer</h3>
      <p style="font-size:.78rem;color:var(--t3);margin-bottom:.8rem">${santriNama}  Transfer ke rekening berikut:</p>

      <div style="background:linear-gradient(135deg,#f0fdf4,#ecfdf5);border:1.5px solid rgba(22,163,74,.2);border-radius:14px;padding:1rem;margin-bottom:.8rem">
        <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.6rem">
          <i class="ri-bank-line" style="font-size:1.2rem;color:var(--green)"></i>
          <strong style="font-size:.9rem">${res.rekening.bank}</strong>
        </div>
        <div style="display:flex;align-items:center;gap:.4rem;margin-bottom:.3rem">
          <span style="font-size:1.1rem;font-weight:800;letter-spacing:1px">${res.rekening.nomor}</span>
          <button onclick="navigator.clipboard.writeText('${res.rekening.nomor}');toast('?? Nomor rekening disalin!')" style="background:var(--green);color:#fff;border:none;border-radius:6px;padding:.2rem .5rem;font-size:.68rem;cursor:pointer"><i class="ri-file-copy-line"></i> Salin</button>
        </div>
        <div style="font-size:.78rem;color:var(--t3)">a.n. ${res.rekening.atas_nama}</div>
      </div>

      <div style="display:flex;justify-content:space-between;border-bottom:1px dashed var(--border);padding-bottom:.5rem;margin-bottom:.5rem;font-size:.85rem">
        <span style="color:var(--t3)">Nominal Tagihan:</span>
        <span style="font-weight:600">Rp ${res.nominal.toLocaleString("id-ID")}</span>
      </div>
      <div style="display:flex;justify-content:space-between;border-bottom:1px dashed var(--border);padding-bottom:.5rem;margin-bottom:.5rem;font-size:.85rem">
        <span style="color:var(--t3)">Biaya Admin:</span>
        <span style="font-weight:600">Rp ${res.fee.toLocaleString("id-ID")}</span>
      </div>
      <div style="display:flex;justify-content:space-between;border-bottom:1px dashed var(--border);padding-bottom:.5rem;margin-bottom:.5rem;font-size:.85rem">
        <span style="color:var(--t3)">Kode Unik:</span>
        <span style="font-weight:600;color:var(--orange)">Rp ${res.kode_unik}</span>
      </div>
      <div style="display:flex;justify-content:space-between;background:var(--blue-light);padding:.8rem;border-radius:10px;margin-bottom:1rem">
        <span style="font-weight:600;color:var(--blue)">Total Transfer:</span>
        <span style="font-weight:800;font-size:1.1rem;color:var(--blue)">Rp ${res.total_transfer.toLocaleString("id-ID")}</span>
      </div>

      <div style="background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.2);border-radius:8px;padding:.8rem;margin-bottom:1rem;font-size:.78rem;color:var(--red)">
        <i class="ri-error-warning-line"></i> Transfer <strong>TEPAT SESUAI NOMINAL</strong> di atas (termasuk 3 digit terakhir). Batas waktu pembayaran: <strong>${expiryDate}</strong>
      </div>

      <button class="btn btn-primary" style="width:100%" onclick="hideModal();window.location.hash='#riwayat';window.location.reload()">Selesai & Tutup</button>
    </div>`;
  } catch (err) {
    btn.disabled = false;
    btn.innerHTML = '<i class="ri-bank-line"></i> Transfer Bank';
    toast(err.message, "error");
  }
}

