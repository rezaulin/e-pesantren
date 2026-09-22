# 🕌 e-Pesantren Multi-Tenant

Sistem manajemen pesantren berbasis web dengan arsitektur multi-tenant. Satu instalasi untuk banyak pesantren.

![Status](https://img.shields.io/badge/Status-Production_Ready-brightgreen) ![Go](https://img.shields.io/badge/Go-1.22+-00ADD8) ![Fiber](https://img.shields.io/badge/Fiber-v2-blue) ![MariaDB](https://img.shields.io/badge/MariaDB-10.6+-blue)

## ✨ Fitur

| Modul | Deskripsi |
|-------|-----------|
| **Multi-Tenant** | Satu server untuk banyak pesantren, data terisolasi |
| **Kontrol Fitur** | Aktifkan/nonaktifkan modul per tenant |
| **Dashboard** | Statistik realtime per role (Admin/Ustadz/Wali/Bendahara) + Jam WIB |
| **Manajemen Santri** | CRUD santri + import Excel massal + profil lengkap + cascade delete |
| **PSB Online** | Pendaftaran Santri Baru publik + dashboard review admin + kuota |
| **Kamar & Kelompok** | Kelola asrama, kelompok belajar + anggota |
| **Kelas Sekolah** | Kelas sekolah dengan manajemen anggota + edit nama |
| **Kelas Diniyyah** | Madrasah Diniyyah: kelas, mata pelajaran, penilaian |
| **Kegiatan** | Kelola kegiatan (pokok/tambahan) + kelompok per kegiatan |
| **Absensi Kegiatan** | Absensi per kegiatan/kelompok dengan status H/I/S/A |
| **Absen Sekolah** | Absensi per kelas → pilih jadwal pelajaran → cegah duplikasi |
| **Absen Diniyyah** | Absensi per kelas diniyyah |
| **Absen Malam** | Absensi malam per kamar, terintegrasi ke raport |
| **Jadwal Kegiatan** | Jadwal per hari (multi-hari + "Setiap Hari") + time-lock ustadz |
| **Jadwal Pelajaran** | Jadwal sekolah & diniyyah per kelas |
| **Penilaian** | Nilai harian/UTS/UAS per mapel (Sekolah + Diniyyah + Kegiatan) |
| **Peringkat** | Ranking otomatis per kelas/kelompok |
| **Pembayaran SPP** | Tarif rincian, potongan per santri, pembayaran bulanan, export Excel |
| **Catatan Bendahara** | Catatan keuangan masuk/keluar + saldo + export Excel |
| **Raport** | Rekap kegiatan + sekolah + malam + pelanggaran + catatan guru, PDF & Excel |
| **Raport Penilaian** | Raport nilai akademik (Diniyyah + Sekolah + Kegiatan), PDF & Excel |
| **Raport Bulk** | Download raport semua santri sekaligus dalam ZIP |
| **Catatan Guru** | Catatan perkembangan santri |
| **Pelanggaran** | Pencatatan pelanggaran dengan sistem poin |
| **Perizinan** | Izin keluar santri + tracking kembali/terlambat |
| **Rekap Ustadz** | Monitoring keaktifan ustadz mengajar per bulan |
| **Wali Dashboard** | Dashboard khusus wali: lihat raport, pembayaran anak |
| **Pengaturan** | Nama lembaga, alamat, kepala, kota, logo, background |
| **Multi-Role** | SuperAdmin, Admin, Ustadz, Bendahara, Wali Santri |
| **PWA** | Progressive Web App, bisa di-install di HP |

## 🛠️ Tech Stack

- **Backend:** Go (Golang) + Fiber v2
- **Database:** MariaDB / MySQL
- **Frontend:** Vanilla JS SPA (Clean Light Blue Theme) + PWA
- **Auth:** JWT (golang-jwt) + bcrypt
- **Export:** gofpdf (PDF) + excelize (Excel)
- **Server:** Nginx reverse proxy + systemd
- **DNS/SSL:** Cloudflare (Proxied + Full)

---

## 🚀 Instalasi (One-Click)

### Persyaratan
- VPS **Ubuntu 22.04 / 24.04 LTS** (minimal RAM 1GB)
- Domain aktif di **Cloudflare**

### Langkah Instalasi

**1. Upload project ke VPS:**
```bash
scp -r . root@IP_SERVER:/root/pesantren-multi
```

**2. SSH ke server dan jalankan installer:**
```bash
ssh root@IP_SERVER
cd /root/pesantren-multi
chmod +x install.sh
sudo bash install.sh
```

**3. Pilih mode & masukkan informasi:**
```
╔══════════════════════════════════════════╗
║  1. FRESH INSTALL  — VPS baru           ║
║  2. UPGRADE        — Update kode        ║
║  3. MIGRATE        — Pindah VPS         ║
╚══════════════════════════════════════════╝

🌐 Domain      : pesantren.example.com
🔑 Password DB : ********
```

**Installer otomatis menangani:**
- ✅ Install Go 1.22, MariaDB, Nginx
- ✅ Build binary Go dari source
- ✅ Buat database + user + import schema
- ✅ Auto-migrate tabel & kolom baru (aman dijalankan berulang)
- ✅ Generate JWT secret
- ✅ SSL certificate wildcard (self-signed, 10 tahun)
- ✅ Nginx wildcard `*.domain.com` + Cloudflare headers
- ✅ Systemd service (auto-start & auto-restart)
- ✅ Firewall (SSH + HTTP/HTTPS)
- ✅ SuperAdmin: `superadmin / superadmin123`

### Setting Cloudflare

Setelah install, setting **3 hal** di Cloudflare:

| No | Setting | Nilai |
|----|---------|-------|
| 1 | DNS → A Record | `pesantren.com` → `IP_SERVER` (☁️ Proxied) |
| 2 | DNS → CNAME | `*` → `pesantren.com` (☁️ Proxied) |
| 3 | SSL/TLS → Mode | **Full** (bukan Full Strict) |

---

## 🖥️ Deploy 2 Web di 1 VPS (Multi-Deploy)

> Panduan jika kamu ingin menjalankan **2 instance e-Pesantren** (atau web lain) di satu VPS yang sama dengan domain berbeda.

### Konsep

```
VPS (1 server)
├── App 1: pesantren-a.com  → port 3002 → DB: pesantren_multi_a
└── App 2: pesantren-b.com  → port 3003 → DB: pesantren_multi_b
```

Setiap instance punya:
- **Port** berbeda (3002, 3003, dst)
- **Database** berbeda
- **Nginx config** berbeda
- **Systemd service** berbeda
- **Direktori** berbeda

### Langkah 1 — Siapkan Direktori

```bash
# Instance 1 (sudah ada dari install pertama)
/var/www/pesantren-multi/

# Instance 2 (copy project baru)
cp -r /root/pesantren-multi /root/pesantren-multi-2
```

### Langkah 2 — Buat Database Kedua

```bash
mysql -u root <<SQL
CREATE DATABASE IF NOT EXISTS pesantren_multi_b
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'pesantren_b'@'localhost'
  IDENTIFIED BY 'PASSWORD_DB_2_KAMU';

GRANT ALL PRIVILEGES ON pesantren_multi_b.*
  TO 'pesantren_b'@'localhost';

FLUSH PRIVILEGES;
SQL

# Import schema
mysql -u pesantren_b -p'PASSWORD_DB_2_KAMU' pesantren_multi_b < /root/pesantren-multi-2/schema.sql
```

### Langkah 3 — Konfigurasi .env Instance 2

```bash
cat > /var/www/pesantren-multi-2/go-backend/.env <<EOF
DB_HOST=localhost
DB_USER=pesantren_b
DB_PASS=PASSWORD_DB_2_KAMU
DB_NAME=pesantren_multi_b
PORT=3003
JWT_SECRET=$(openssl rand -hex 32)
DOMAIN=pesantren-b.com
EOF
```

> ⚠️ **Penting:** `PORT` harus berbeda dari instance pertama (3002). Gunakan 3003, 3004, dst.

### Langkah 4 — Build & Deploy Instance 2

```bash
# Copy file ke direktori deploy
APP_DIR_2="/var/www/pesantren-multi-2"
mkdir -p $APP_DIR_2
cp -r /root/pesantren-multi-2/go-backend/. $APP_DIR_2/go-backend/
cp -r /root/pesantren-multi-2/public/. $APP_DIR_2/public/
cp /root/pesantren-multi-2/schema.sql $APP_DIR_2/

# Build binary
cd $APP_DIR_2/go-backend
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go build -o pesantren-server .
chmod +x pesantren-server
```

### Langkah 5 — Buat Systemd Service Kedua

```bash
cat > /etc/systemd/system/pesantren-multi-2.service <<EOF
[Unit]
Description=E-Pesantren Instance 2 (Go/Fiber)
After=network.target mariadb.service
Wants=mariadb.service

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/pesantren-multi-2/go-backend
ExecStart=/var/www/pesantren-multi-2/go-backend/pesantren-server
Restart=always
RestartSec=5
Environment=TZ=Asia/Jakarta
EnvironmentFile=/var/www/pesantren-multi-2/go-backend/.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable pesantren-multi-2
systemctl start pesantren-multi-2
```

### Langkah 6 — Konfigurasi Nginx Instance 2

```bash
# Buat SSL untuk domain kedua
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
  -keyout /etc/nginx/ssl/pesantren-b.key \
  -out /etc/nginx/ssl/pesantren-b.crt \
  -subj "/CN=*.pesantren-b.com" \
  -addext "subjectAltName=DNS:pesantren-b.com,DNS:*.pesantren-b.com"

# Buat config nginx
cat > /etc/nginx/sites-available/pesantren-multi-2 <<'NGINX'
server {
    listen 80;
    server_name pesantren-b.com *.pesantren-b.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name pesantren-b.com *.pesantren-b.com;

    ssl_certificate     /etc/nginx/ssl/pesantren-b.crt;
    ssl_certificate_key /etc/nginx/ssl/pesantren-b.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript;
    gzip_min_length 256;

    location / {
        proxy_pass http://127.0.0.1:3003;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 300s;
        client_max_body_size 10M;
    }
}
NGINX

# Aktifkan & reload
ln -sf /etc/nginx/sites-available/pesantren-multi-2 /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
```

### Langkah 7 — Setting Cloudflare Domain Kedua

| No | Setting | Nilai |
|----|---------|-------|
| 1 | DNS → A Record | `pesantren-b.com` → `IP_SERVER` (☁️ Proxied) |
| 2 | DNS → CNAME | `*` → `pesantren-b.com` (☁️ Proxied) |
| 3 | SSL/TLS → Mode | **Full** |

### Langkah 8 — Buat SuperAdmin Instance 2

```bash
# Generate hash password
SA_HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'superadmin123', bcrypt.gensalt()).decode())")

mysql -u pesantren_b -p'PASSWORD_DB_2_KAMU' pesantren_multi_b <<SQL
INSERT INTO tenants (nama, subdomain, status) VALUES ('System', 'system', 'active')
ON DUPLICATE KEY UPDATE nama=nama;
SET @tid = (SELECT id FROM tenants WHERE subdomain = 'system' LIMIT 1);
DELETE FROM users WHERE username = 'superadmin';
INSERT INTO users (tenant_id, username, password_hash, role, nama)
  VALUES (@tid, 'superadmin', '${SA_HASH}', 'superadmin', 'Super Admin');
SQL
```

### Ringkasan Multi-Deploy

| Setting | Instance 1 | Instance 2 |
|---------|-----------|-----------|
| Direktori | `/var/www/pesantren-multi/` | `/var/www/pesantren-multi-2/` |
| Port | 3002 | 3003 |
| Database | `pesantren_multi` | `pesantren_multi_b` |
| DB User | `pesantren` | `pesantren_b` |
| Domain | `pesantren-a.com` | `pesantren-b.com` |
| Service | `pesantren-multi` | `pesantren-multi-2` |
| Nginx Config | `pesantren-multi` | `pesantren-multi-2` |
| SSL Cert | `pesantren.crt` | `pesantren-b.crt` |

### Perintah Berguna Multi-Deploy

```bash
# Cek status kedua instance
systemctl status pesantren-multi
systemctl status pesantren-multi-2

# Restart masing-masing
systemctl restart pesantren-multi
systemctl restart pesantren-multi-2

# Log masing-masing
journalctl -u pesantren-multi -f
journalctl -u pesantren-multi-2 -f

# Backup database masing-masing
mysqldump -u pesantren -p pesantren_multi > backup-instance1.sql
mysqldump -u pesantren_b -p pesantren_multi_b > backup-instance2.sql
```

---

## 🔄 Pindah VPS (Migrate)

> Cukup 3 langkah. Data tidak hilang.

### Langkah 1 — Backup di VPS Lama

```bash
ssh root@IP_VPS_LAMA
mysqldump -u pesantren -p pesantren_multi > /tmp/backup-pesantren.sql
ls -lh /tmp/backup-pesantren.sql
```

### Langkah 2 — Transfer ke VPS Baru

```bash
# Dari laptop/PC kamu
scp root@IP_VPS_LAMA:/tmp/backup-pesantren.sql /tmp/backup-pesantren.sql
scp /tmp/backup-pesantren.sql root@IP_VPS_BARU:/tmp/backup-pesantren.sql

scp -r root@IP_VPS_LAMA:/root/pesantren-multi /tmp/pesantren-multi
scp -r /tmp/pesantren-multi root@IP_VPS_BARU:/root/pesantren-multi
```

> **Alternatif:** Jika kodenya ada di GitHub/Git, cukup `git clone` di VPS baru.

### Langkah 3 — Install di VPS Baru

```bash
ssh root@IP_VPS_BARU
cd /root/pesantren-multi
chmod +x install.sh
sudo bash install.sh
```

```
Pilih mode: 3 (MIGRATE)
🌐 Domain      : pesantren.example.com     ← domain yang sama / baru
🔑 Password DB : ********                  ← password baru bebas
📦 Path backup : /tmp/backup-pesantren.sql ← file backup tadi
```

### Langkah 4 — Update DNS Cloudflare

| Setting | Nilai Lama | Nilai Baru |
|---------|-----------|------------|
| DNS A Record | IP VPS Lama | **IP VPS Baru** |

### Checklist Pindah VPS

- [ ] Backup database dari VPS lama
- [ ] Transfer backup + kode ke VPS baru
- [ ] Jalankan `install.sh` mode 3 (MIGRATE)
- [ ] Update A Record Cloudflare ke IP VPS baru
- [ ] Test login di `https://domain-anda/app`
- [ ] Verifikasi data santri & absensi masih ada
- [ ] Matikan VPS lama (opsional, setelah yakin)

---

## 📖 Panduan Penggunaan

### Login Pertama

Buka `https://domain-anda.com/app` dan login:

| Role | Username | Password |
|------|----------|----------|
| Super Admin | `superadmin` | `superadmin123` |

### Alur Kerja

```
SuperAdmin → Buat Tenant (+ pilih fitur aktif) → Auto-generate Admin
    ↓
Admin Tenant → Setup Kamar, Kegiatan, Kelompok, Kelas, Diniyyah
    ↓
Admin → Tambah Santri + Buat User Ustadz/Wali/Bendahara
    ↓
Ustadz → Input Absensi, Nilai, Catatan Guru, Pelanggaran
    ↓
Bendahara → Kelola Pembayaran SPP, Catatan Keuangan
    ↓
Wali → Lihat Dashboard Anak, Raport, Pembayaran
```

### Membuat Tenant Baru

1. Login sebagai **superadmin**
2. Klik **Kelola Tenant** → **+ Tambah**
3. Isi nama & subdomain (contoh: `al-falah`)
4. Pilih fitur yang diaktifkan
5. Sistem otomatis membuat admin tenant
6. Login dari `al-falah.domain.com/app` sebagai `al-falah_admin / admin123`

### Import Santri dari Excel

Format file Excel (.xlsx):

| Kolom A | Kolom B | Kolom C |
|---------|---------|---------|
| Nama Santri | Alamat | Nama Wali |

---

## 🗂️ Struktur Project

```
pesantren-multi/
├── README.md                  ← Dokumentasi
├── schema.sql                 ← Database DDL + migration
├── install.sh                 ← Installer script
├── go-backend/                ← Backend (Go/Fiber)
│   ├── .env                   ← Konfigurasi database & JWT
│   ├── main.go                ← Entry point + routing
│   ├── go.mod                 ← Go module dependencies
│   ├── config/
│   │   ├── database.go        ← MySQL connection pool
│   │   └── migrate.go         ← Auto-migration
│   ├── middleware/
│   │   └── auth.go            ← JWT authentication + role guard
│   ├── helpers/
│   │   └── waktu.go           ← Timezone WIB helpers
│   └── handlers/
│       ├── auth.go            ← Login & /api/me
│       ├── super_admin.go     ← Tenant management + broadcast
│       ├── dashboard.go       ← Dashboard data per role
│       ├── users.go           ← User CRUD
│       ├── santri.go          ← Santri CRUD + Excel import
│       ├── santri_profile.go  ← Profil lengkap santri
│       ├── psb.go             ← Pendaftaran Santri Baru (publik)
│       ├── kamar.go           ← Kamar + members
│       ├── kegiatan.go        ← Kegiatan CRUD
│       ├── kelompok.go        ← Kelompok + members
│       ├── kelas_sekolah.go   ← Kelas sekolah + anggota
│       ├── diniyyah.go        ← Kelas Diniyyah + anggota
│       ├── absensi.go         ← Absensi kegiatan/malam/sekolah
│       ├── absen_diniyyah.go  ← Absensi diniyyah
│       ├── jadwal.go          ← Jadwal kegiatan + sekolah
│       ├── nilai.go           ← Nilai Diniyyah + Kegiatan
│       ├── nilai_sekolah.go   ← Nilai Sekolah
│       ├── pelanggaran.go     ← Pelanggaran CRUD
│       ├── perizinan.go       ← Perizinan santri
│       ├── catatan_guru.go    ← Catatan guru
│       ├── pengumuman.go      ← Pengumuman
│       ├── raport.go          ← Raport santri + PDF/Excel
│       ├── raport_all.go      ← Raport bulk → ZIP
│       ├── raport_penilaian.go← Raport penilaian akademik
│       ├── rekap_ustadz.go    ← Rekap aktivitas ustadz
│       ├── pembayaran.go      ← SPP: tarif, potongan, bayar
│       ├── keuangan.go        ← Catatan keuangan + saldo
│       ├── wali.go            ← Dashboard wali santri
│       └── settings.go        ← Pengaturan lembaga
└── public/                    ← Frontend SPA
    ├── index.html             ← HTML shell (app)
    ├── landing.html           ← Landing page
    ├── psb.html               ← Form PSB publik
    ├── style.css              ← Design system
    ├── app-core.js            ← Core: auth, dashboard, sidebar
    ├── app-absensi.js         ← Modul absensi
    ├── app-views.js           ← Santri, raport, rekap
    ├── app-extra.js           ← Admin: settings, jadwal, users, dll
    ├── app-penilaian.js       ← Modul penilaian akademik
    ├── app-nilai-kegiatan.js  ← Nilai kegiatan
    ├── manifest.json          ← PWA manifest
    ├── sw.js                  ← Service worker
    ├── icon-192.png           ← App icon
    └── icon-512.png           ← App icon large
```

---

## ⚙️ Perintah Operasional

```bash
# Status aplikasi
systemctl status pesantren-multi

# Restart aplikasi
systemctl restart pesantren-multi

# Lihat log realtime
journalctl -u pesantren-multi -f

# Lihat 50 baris log terakhir
journalctl -u pesantren-multi -n 50

# Stop aplikasi
systemctl stop pesantren-multi

# Restart Nginx
systemctl restart nginx
```

## ♻️ Update Aplikasi (di VPS yang sama)

Cukup jalankan installer lagi dengan mode **UPGRADE** — data aman, .env tidak berubah:

```bash
cd /root/pesantren-multi
git pull  # jika pakai git, atau upload kode baru dulu
sudo bash install.sh
# Pilih mode: 2 (UPGRADE)
```

Installer akan:
- ✅ Update kode & build ulang binary
- ✅ Jalankan migration kolom & tabel baru
- ✅ Restart service
- ✅ Tidak mengubah data, .env, atau SSL

## 🔒 Keamanan

- Ganti password superadmin segera setelah login pertama
- SSL via Cloudflare Full mode
- JWT secret di-konfigurasi via `.env`
- Setiap tenant data terisolasi via `tenant_id`
- Password di-hash dengan bcrypt
- Login dibatasi per subdomain tenant (slug validation)

---

**Made with ❤️ for Pesantren Indonesia**
