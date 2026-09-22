#!/bin/bash
# ══════════════════════════════════════════════════════════
# E-PESANTREN ONE-CLICK INSTALLER v5.0 (Go/Fiber)
# Tested on: Ubuntu 22.04 / 24.04 LTS
# Stack: Go + Fiber + MariaDB + Nginx + Cloudflare
# Mode: FRESH INSTALL | UPGRADE | MIGRATE (pindah VPS)
# ══════════════════════════════════════════════════════════
set -e
export DEBIAN_FRONTEND=noninteractive

# ── Colors ──
RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[ OK ]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()  { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }
step()  { echo -e "\n${CYAN}══════════════════════════════════════════${NC}"; echo -e "${CYAN}  $1${NC}"; echo -e "${CYAN}══════════════════════════════════════════${NC}"; }

# ── Root check ──
if [ "$EUID" -ne 0 ]; then fail "Jalankan sebagai root: sudo bash install.sh"; fi

clear
echo -e "${GREEN}"
echo "╔══════════════════════════════════════════════════════╗"
echo "║       E-PESANTREN INSTALLER v5.0 (Go/Fiber)         ║"
echo "║     Sistem Manajemen Pesantren Multi-Tenant          ║"
echo "╠══════════════════════════════════════════════════════╣"
echo "║  1. FRESH INSTALL  — VPS baru, belum ada data       ║"
echo "║  2. UPGRADE        — Update kode di VPS yg sama     ║"
echo "║  3. MIGRATE        — Pindah dari VPS lama           ║"
echo "╚══════════════════════════════════════════════════════╝"
echo -e "${NC}"

read -p "Pilih mode [1/2/3]: " MODE
case "$MODE" in
  1) MODE_NAME="FRESH INSTALL" ;;
  2) MODE_NAME="UPGRADE" ;;
  3) MODE_NAME="MIGRATE" ;;
  *) fail "Pilihan tidak valid. Masukkan 1, 2, atau 3." ;;
esac

echo ""
echo -e "${CYAN}Mode: ${GREEN}${MODE_NAME}${NC}"
echo ""

# ══════════════════════════════════════════════════════════
# INPUT — Hanya 2 hal yang perlu diisi
# ══════════════════════════════════════════════════════════
APP_DIR="/var/www/pesantren-multi"
GO_VER="1.22.4"
PORT=3002

if [ "$MODE" == "2" ]; then
  # UPGRADE: baca dari .env yang sudah ada
  if [ -f "$APP_DIR/go-backend/.env" ]; then
    DOMAIN=$(grep DOMAIN "$APP_DIR/go-backend/.env" 2>/dev/null | cut -d= -f2 || echo "")
    DB_PASS=$(grep DB_PASS "$APP_DIR/go-backend/.env" | cut -d= -f2)
    JWT_SEC=$(grep JWT_SECRET "$APP_DIR/go-backend/.env" | cut -d= -f2)
    info "Konfigurasi lama ditemukan, tidak perlu input ulang."
  else
    warn ".env tidak ditemukan, perlu input manual."
    MODE="1"
  fi
fi

if [ "$MODE" != "2" ]; then
  read -p "🌐 Domain (contoh: e-pesantren.app): " DOMAIN
  [ -z "$DOMAIN" ] && fail "Domain wajib diisi!"

  read -sp "🔑 Password database: " DB_PASS; echo ""
  [ -z "$DB_PASS" ] && fail "Password database wajib diisi!"

  if [ "$MODE" == "3" ]; then
    echo ""
    echo -e "${YELLOW}══ MIGRATE: Backup dari VPS Lama ══${NC}"
    echo "  Kamu memerlukan file backup database dari VPS lama."
    echo "  Cara backup di VPS lama:"
    echo -e "  ${CYAN}mysqldump -u pesantren -p pesantren_multi > /tmp/backup.sql${NC}"
    echo "  Lalu transfer ke VPS baru:"
    echo -e "  ${CYAN}scp root@IP_VPS_LAMA:/tmp/backup.sql /tmp/backup.sql${NC}"
    echo ""
    read -p "📦 Path file backup SQL (kosongkan jika belum ada): " BACKUP_FILE
  fi

  JWT_SEC="pesantren-jwt-$(openssl rand -hex 16 2>/dev/null || date +%s%N)"
fi

SA_USER="superadmin"
SA_PASS="superadmin123"

echo ""
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo -e "  Mode         : ${GREEN}${MODE_NAME}${NC}"
echo -e "  Domain       : ${GREEN}${DOMAIN:-tidak diset}${NC}"
echo -e "  Wildcard     : ${GREEN}*.${DOMAIN:-tidak diset}${NC}"
echo -e "  App Dir      : ${GREEN}${APP_DIR}${NC}"
[ "$MODE" == "1" ] && echo -e "  SuperAdmin   : ${GREEN}${SA_USER} / ${SA_PASS}${NC}"
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo ""
read -p "Lanjutkan? (y/n): " CONFIRM
[ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ] && exit 0
echo ""

# ── Fungsi migration (dipakai mode 2 & 3) ──
_run_migration() {
  mysql -u pesantren -p"${DB_PASS}" pesantren_multi <<EOSQL
-- ══ Tabel baru jika belum ada ══

CREATE TABLE IF NOT EXISTS kelas_sekolah (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS santri_kelas (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  kelas_id INT NOT NULL,
  status VARCHAR(20) DEFAULT 'active',
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id)
);
CREATE TABLE IF NOT EXISTS jadwal_sekolah (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  kelas_id INT NOT NULL,
  mata_pelajaran VARCHAR(200) NOT NULL,
  ustadz_username VARCHAR(100) NOT NULL,
  hari VARCHAR(20) NOT NULL,
  jam_mulai VARCHAR(10) NOT NULL,
  jam_selesai VARCHAR(10) NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(kelas_id)
);
CREATE TABLE IF NOT EXISTS absen_sekolah_sesi (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  kelas_id INT NOT NULL,
  jadwal_sekolah_id INT DEFAULT NULL,
  tanggal VARCHAR(20) NOT NULL,
  ustadz_username VARCHAR(100) DEFAULT '',
  recorded_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(tanggal)
);
CREATE TABLE IF NOT EXISTS pembayaran_tarif (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  nama VARCHAR(200) NOT NULL,
  nominal BIGINT NOT NULL DEFAULT 0,
  aktif TINYINT(1) DEFAULT 1,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id)
);
CREATE TABLE IF NOT EXISTS pembayaran_potongan (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  nama VARCHAR(200) NOT NULL,
  nominal BIGINT NOT NULL DEFAULT 0,
  aktif TINYINT(1) DEFAULT 1,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(santri_id)
);
CREATE TABLE IF NOT EXISTS pembayaran (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  bulan VARCHAR(7) NOT NULL,
  nominal BIGINT NOT NULL DEFAULT 0,
  metode VARCHAR(50) DEFAULT 'tunai',
  keterangan TEXT,
  created_by INT NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(santri_id),
  INDEX(bulan)
);
CREATE TABLE IF NOT EXISTS catatan_keuangan (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  tipe ENUM('masuk','keluar') NOT NULL,
  nominal BIGINT NOT NULL DEFAULT 0,
  keterangan TEXT,
  tanggal DATE NOT NULL,
  created_by INT NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(tanggal)
);

-- ══ Madrasah Diniyyah ══
CREATE TABLE IF NOT EXISTS kelas_diniyyah (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS santri_kelas_diniyyah (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  kelas_diniyyah_id INT NOT NULL,
  status VARCHAR(20) DEFAULT 'active',
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(kelas_diniyyah_id),
  INDEX(santri_id)
);

-- ══ Penilaian ══
CREATE TABLE IF NOT EXISTS mata_pelajaran (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  kelas_diniyyah_id INT NOT NULL,
  nama VARCHAR(200) NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(kelas_diniyyah_id)
);
CREATE TABLE IF NOT EXISTS nilai_pelajaran (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  mata_pelajaran_id INT NOT NULL,
  kelas_diniyyah_id INT NOT NULL,
  semester VARCHAR(10) NOT NULL,
  nilai_harian DECIMAL(5,2) DEFAULT 0,
  nilai_uts DECIMAL(5,2) DEFAULT 0,
  nilai_uas DECIMAL(5,2) DEFAULT 0,
  nilai_akhir DECIMAL(5,2) DEFAULT 0,
  created_by INT NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  updated_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(santri_id),
  INDEX(mata_pelajaran_id),
  INDEX(kelas_diniyyah_id),
  INDEX(semester),
  UNIQUE KEY uq_np (tenant_id, santri_id, mata_pelajaran_id, semester)
);
CREATE TABLE IF NOT EXISTS nilai_kegiatan (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  kegiatan_id INT NOT NULL,
  kelompok_id INT NOT NULL,
  bulan VARCHAR(7) NOT NULL,
  nilai DECIMAL(5,2) DEFAULT 0,
  catatan TEXT,
  created_by INT NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  updated_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(santri_id),
  INDEX(kegiatan_id),
  INDEX(kelompok_id),
  INDEX(bulan),
  UNIQUE KEY uq_nk (tenant_id, santri_id, kegiatan_id, kelompok_id, bulan)
);
CREATE TABLE IF NOT EXISTS mata_pelajaran_sekolah (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  kelas_id INT NOT NULL,
  nama VARCHAR(200) NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(kelas_id)
);
CREATE TABLE IF NOT EXISTS nilai_sekolah (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tenant_id INT NOT NULL,
  santri_id INT NOT NULL,
  mata_pelajaran_sekolah_id INT NOT NULL,
  kelas_id INT NOT NULL,
  semester VARCHAR(10) NOT NULL,
  nilai_harian DECIMAL(5,2) DEFAULT 0,
  nilai_uts DECIMAL(5,2) DEFAULT 0,
  nilai_uas DECIMAL(5,2) DEFAULT 0,
  nilai_akhir DECIMAL(5,2) DEFAULT 0,
  created_by INT NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  updated_at DATETIME DEFAULT NOW(),
  INDEX(tenant_id),
  INDEX(santri_id),
  INDEX(mata_pelajaran_sekolah_id),
  INDEX(kelas_id),
  INDEX(semester),
  UNIQUE KEY uq_ns (tenant_id, santri_id, mata_pelajaran_sekolah_id, semester)
);

-- ══ Kolom baru (aman dijalankan berulang) ══
ALTER TABLE absen_malam ADD COLUMN IF NOT EXISTS kamar_id INT DEFAULT NULL;
ALTER TABLE absen_sekolah ADD COLUMN IF NOT EXISTS kelas_id INT DEFAULT NULL;
ALTER TABLE absen_sekolah ADD COLUMN IF NOT EXISTS jadwal_sekolah_id INT DEFAULT NULL;
ALTER TABLE absen_sekolah ADD COLUMN IF NOT EXISTS mata_pelajaran VARCHAR(200) DEFAULT NULL;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS features TEXT;
ALTER TABLE santri ADD COLUMN IF NOT EXISTS nama_wali VARCHAR(200) DEFAULT '';
ALTER TABLE santri ADD COLUMN IF NOT EXISTS no_hp VARCHAR(20) DEFAULT '';
ALTER TABLE settings ADD COLUMN IF NOT EXISTS dashboard_bg TEXT;
ALTER TABLE users MODIFY COLUMN role ENUM('superadmin','admin','ustadz','wali','bendahara') NOT NULL;
EOSQL
}

# ══════════════════════════════════════════════════════════
# STEP 1: UPDATE SISTEM & DEPENDENCIES
# ══════════════════════════════════════════════════════════
step "1/8 Update sistem & install dependencies"
apt update -y && apt upgrade -y
apt install -y curl wget git unzip ufw nginx mariadb-server mariadb-client openssl python3-pip
pip3 install bcrypt -q 2>/dev/null || pip3 install bcrypt --break-system-packages -q 2>/dev/null || true
ok "Sistem & dependencies siap"

# ══════════════════════════════════════════════════════════
# STEP 2: INSTALL GO
# ══════════════════════════════════════════════════════════
step "2/8 Install Go ${GO_VER}"
if ! /usr/local/go/bin/go version &>/dev/null; then
  info "Download Go ${GO_VER}..."
  wget -q "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  grep -q "/usr/local/go/bin" /etc/profile || echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
  grep -q "/usr/local/go/bin" ~/.bashrc || echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  ok "Go ${GO_VER} terinstall"
else
  ok "Go sudah terinstall: $(/usr/local/go/bin/go version | awk '{print $3}')"
fi
export PATH=$PATH:/usr/local/go/bin

# ══════════════════════════════════════════════════════════
# STEP 3: SETUP MARIADB
# ══════════════════════════════════════════════════════════
step "3/8 Setup MariaDB"
systemctl start mariadb && systemctl enable mariadb

if [ "$MODE" == "1" ] || [ "$MODE" == "3" ]; then
  mysql -u root <<EOSQL
CREATE DATABASE IF NOT EXISTS pesantren_multi CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'pesantren'@'localhost' IDENTIFIED BY '${DB_PASS}';
GRANT ALL PRIVILEGES ON pesantren_multi.* TO 'pesantren'@'localhost';
FLUSH PRIVILEGES;
EOSQL
  ok "Database & user dibuat"
fi

# ══════════════════════════════════════════════════════════
# STEP 4: DEPLOY FILE APLIKASI
# ══════════════════════════════════════════════════════════
step "4/8 Deploy file aplikasi"
mkdir -p $APP_DIR

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ "$MODE" == "2" ]; then
  # Upgrade: hanya update kode, jangan hapus .env
  info "Backup .env lama..."
  cp "$APP_DIR/go-backend/.env" /tmp/.env.backup 2>/dev/null || true
fi

info "Copy file aplikasi..."
cp -r "$SCRIPT_DIR/go-backend/." "$APP_DIR/go-backend/"
cp -r "$SCRIPT_DIR/public/." "$APP_DIR/public/"
cp "$SCRIPT_DIR/schema.sql" "$APP_DIR/" 2>/dev/null || true
chmod -R 755 "$APP_DIR"

if [ "$MODE" == "2" ]; then
  # Restore .env setelah upgrade
  cp /tmp/.env.backup "$APP_DIR/go-backend/.env" 2>/dev/null || true
  ok "Kode diupdate, .env tetap"
else
  ok "File aplikasi di-deploy ke $APP_DIR"
fi

# ══════════════════════════════════════════════════════════
# STEP 5: IMPORT / MIGRATE DATABASE
# ══════════════════════════════════════════════════════════
step "5/8 Setup database"

if [ "$MODE" == "1" ]; then
  # Fresh install: import schema baru
  info "Import schema database..."
  mysql -u pesantren -p"${DB_PASS}" pesantren_multi < "$APP_DIR/schema.sql"
  ok "Schema diimport"

  # Buat superadmin
  info "Membuat SuperAdmin..."
  SA_HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'${SA_PASS}', bcrypt.gensalt()).decode())")
  mysql -u pesantren -p"${DB_PASS}" pesantren_multi <<EOSQL
INSERT INTO tenants (nama, subdomain, status) VALUES ('System', 'system', 'active')
ON DUPLICATE KEY UPDATE nama=nama;
SET @tid = (SELECT id FROM tenants WHERE subdomain = 'system' LIMIT 1);
DELETE FROM users WHERE username = '${SA_USER}';
INSERT INTO users (tenant_id, username, password_hash, role, nama)
  VALUES (@tid, '${SA_USER}', '${SA_HASH}', 'superadmin', 'Super Admin');
EOSQL
  ok "SuperAdmin dibuat: ${SA_USER} / ${SA_PASS}"

elif [ "$MODE" == "3" ]; then
  # Migrate: restore backup dulu, lalu jalankan migration
  if [ -n "$BACKUP_FILE" ] && [ -f "$BACKUP_FILE" ]; then
    info "Restore backup database dari $BACKUP_FILE..."
    mysql -u pesantren -p"${DB_PASS}" pesantren_multi < "$BACKUP_FILE"
    ok "Backup ter-restore"
  else
    warn "File backup tidak ditemukan. Import schema kosong saja..."
    mysql -u pesantren -p"${DB_PASS}" pesantren_multi < "$APP_DIR/schema.sql"
    ok "Schema kosong diimport"
  fi
  # Jalankan migration (kolom baru) — aman dijalankan berkali-kali
  info "Menjalankan migration kolom baru..."
  _run_migration

elif [ "$MODE" == "2" ]; then
  # Upgrade: jalankan migration saja (schema & kolom baru)
  info "Menjalankan migration database..."
  _run_migration
fi

ok "Database siap"

# ══════════════════════════════════════════════════════════
# STEP 6: BUAT .env & BUILD BINARY
# ══════════════════════════════════════════════════════════
step "6/8 Build aplikasi Go"
cd "$APP_DIR/go-backend"

# Buat .env (skip jika UPGRADE dan .env sudah ada)
if [ "$MODE" != "2" ] || [ ! -f ".env" ]; then
  cat > .env <<EOF
DB_HOST=localhost
DB_USER=pesantren
DB_PASS=${DB_PASS}
DB_NAME=pesantren_multi
PORT=${PORT}
JWT_SECRET=${JWT_SEC}
DOMAIN=${DOMAIN}
EOF
  ok ".env dibuat"
else
  ok ".env tetap (mode UPGRADE)"
fi

info "Download dependencies & build..."
go mod tidy
go build -o pesantren-server .
chmod +x pesantren-server
ok "Binary berhasil dibuat: $(ls -lh pesantren-server | awk '{print $5}')"

# ══════════════════════════════════════════════════════════
# STEP 7: SYSTEMD SERVICE
# ══════════════════════════════════════════════════════════
step "7/8 Setup systemd service"
cat > /etc/systemd/system/pesantren-multi.service <<EOF
[Unit]
Description=E-Pesantren Multi-Tenant (Go/Fiber) v5
After=network.target mariadb.service
Wants=mariadb.service

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR/go-backend
ExecStart=$APP_DIR/go-backend/pesantren-server
Restart=always
RestartSec=5
Environment=TZ=Asia/Jakarta
EnvironmentFile=$APP_DIR/go-backend/.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable pesantren-multi

if [ "$MODE" == "2" ] || systemctl is-active --quiet pesantren-multi 2>/dev/null; then
  systemctl restart pesantren-multi
else
  systemctl start pesantren-multi
fi

sleep 3
if systemctl is-active --quiet pesantren-multi; then
  ok "Service berjalan ✓"
else
  warn "Service gagal start. Cek: journalctl -u pesantren-multi -n 30"
fi

# ══════════════════════════════════════════════════════════
# STEP 8: NGINX + SSL (skip jika mode UPGRADE & sudah ada)
# ══════════════════════════════════════════════════════════
step "8/8 Setup Nginx & SSL"
SSL_DIR="/etc/nginx/ssl"
mkdir -p $SSL_DIR

if [ "$MODE" == "2" ] && [ -f "$SSL_DIR/pesantren.crt" ]; then
  ok "SSL & Nginx sudah ada, skip (mode UPGRADE)"
  nginx -t && systemctl reload nginx
else
  info "Membuat SSL certificate wildcard (10 tahun)..."
  openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout "$SSL_DIR/pesantren.key" \
    -out "$SSL_DIR/pesantren.crt" \
    -subj "/CN=*.${DOMAIN}" \
    -addext "subjectAltName=DNS:${DOMAIN},DNS:*.${DOMAIN}" 2>/dev/null
  ok "SSL certificate dibuat"

  cat > /etc/nginx/sites-available/pesantren-multi <<NGINX
# ── Cloudflare Real IP ──
set_real_ip_from 173.245.48.0/20;
set_real_ip_from 103.21.244.0/22;
set_real_ip_from 103.22.200.0/22;
set_real_ip_from 103.31.4.0/22;
set_real_ip_from 141.101.64.0/18;
set_real_ip_from 108.162.192.0/18;
set_real_ip_from 190.93.240.0/20;
set_real_ip_from 188.114.96.0/20;
set_real_ip_from 197.234.240.0/22;
set_real_ip_from 198.41.128.0/17;
set_real_ip_from 162.158.0.0/15;
set_real_ip_from 104.16.0.0/13;
set_real_ip_from 104.24.0.0/14;
set_real_ip_from 172.64.0.0/13;
set_real_ip_from 131.0.72.0/22;
real_ip_header CF-Connecting-IP;

server {
    listen 80;
    server_name ${DOMAIN} *.${DOMAIN};
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name ${DOMAIN} *.${DOMAIN};

    ssl_certificate     ${SSL_DIR}/pesantren.crt;
    ssl_certificate_key ${SSL_DIR}/pesantren.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 256;

    location / {
        proxy_pass http://127.0.0.1:${PORT};
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 300s;
        client_max_body_size 10M;
    }
}
NGINX

  rm -f /etc/nginx/sites-enabled/default
  ln -sf /etc/nginx/sites-available/pesantren-multi /etc/nginx/sites-enabled/
  nginx -t && systemctl reload nginx
  ok "Nginx + SSL wildcard *.${DOMAIN} aktif"
fi

# Firewall
ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable
ok "Firewall: SSH + HTTP/HTTPS diizinkan"

# ══════════════════════════════════════════════════════════
# SELESAI
# ══════════════════════════════════════════════════════════
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         ${MODE_NAME} BERHASIL! 🎉                   ${NC}${GREEN}║${NC}"
echo -e "${GREEN}╠══════════════════════════════════════════════════════════╣${NC}"
if [ -n "$DOMAIN" ]; then
echo -e "${GREEN}║${NC}  🌐 Landing    : ${CYAN}https://${DOMAIN}${NC}"
echo -e "${GREEN}║${NC}  📱 Dashboard  : ${CYAN}https://${DOMAIN}/app${NC}"
fi
echo -e "${GREEN}║${NC}  🖥️  Server IP  : ${CYAN}${SERVER_IP}${NC}"
[ "$MODE" == "1" ] && echo -e "${GREEN}║${NC}  👤 SuperAdmin : ${CYAN}${SA_USER} / ${SA_PASS}${NC}"
echo -e "${GREEN}╠══════════════════════════════════════════════════════════╣${NC}"
if [ -n "$DOMAIN" ]; then
echo -e "${GREEN}║${NC}  ${YELLOW}⚡ SETUP CLOUDFLARE (wajib jika domain baru):${NC}"
echo -e "${GREEN}║${NC}  1. DNS A     : ${CYAN}${DOMAIN}${NC} → ${CYAN}${SERVER_IP}${NC} (Proxied ☁️)"
echo -e "${GREEN}║${NC}  2. DNS CNAME : ${CYAN}*${NC} → ${CYAN}${DOMAIN}${NC} (Proxied ☁️)"
echo -e "${GREEN}║${NC}  3. SSL/TLS   : ${CYAN}Full${NC} (bukan Full Strict)"
echo -e "${GREEN}╠══════════════════════════════════════════════════════════╣${NC}"
fi
echo -e "${GREEN}║${NC}  ${YELLOW}📋 CARA PAKAI:${NC}"
echo -e "${GREEN}║${NC}  • Login superadmin → Buat Tenant"
echo -e "${GREEN}║${NC}  • Admin tenant login di: ${CYAN}nama-pesantren.${DOMAIN}/app${NC}"
echo -e "${GREEN}╠══════════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║${NC}  ${YELLOW}🔧 PERINTAH BERGUNA:${NC}"
echo -e "${GREEN}║${NC}  Status  : ${CYAN}systemctl status pesantren-multi${NC}"
echo -e "${GREEN}║${NC}  Log     : ${CYAN}journalctl -u pesantren-multi -f${NC}"
echo -e "${GREEN}║${NC}  Restart : ${CYAN}systemctl restart pesantren-multi${NC}"
echo -e "${GREEN}║${NC}  Update  : ${CYAN}bash install.sh${NC} (pilih mode 2)"
echo -e "${GREEN}║${NC}  Backup  : ${CYAN}mysqldump -u pesantren -p pesantren_multi > backup.sql${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
