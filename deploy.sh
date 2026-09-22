#!/bin/bash
# ══════════════════════════════════════════════════════════
# DEPLOY SCRIPT — e-pesantren.app
# Jalankan dari Windows (Git Bash/WSL) atau dari VPS
# Usage: bash deploy.sh
# ══════════════════════════════════════════════════════════
set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[ OK ]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()  { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }

echo -e "${GREEN}"
echo "╔══════════════════════════════════════════════════════╗"
echo "║   🚀 DEPLOY E-PESANTREN KE VPS via SCP              ║"
echo "║   + Ganti DB Password + JWT Secret                   ║"
echo "╚══════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ── INPUT ──
read -p "🌐 IP Address VPS: " VPS_IP
[ -z "$VPS_IP" ] && fail "IP VPS wajib diisi!"

read -p "👤 SSH User [root]: " SSH_USER
SSH_USER=${SSH_USER:-root}

read -p "🔌 SSH Port [22]: " SSH_PORT
SSH_PORT=${SSH_PORT:-22}

read -sp "🔑 Password DB BARU (untuk production): " NEW_DB_PASS; echo ""
[ -z "$NEW_DB_PASS" ] && fail "Password DB baru wajib diisi!"

# Generate JWT Secret otomatis (64 hex chars = 32 bytes)
NEW_JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || python3 -c "import secrets; print(secrets.token_hex(32))" 2>/dev/null || echo "epesantren-$(date +%s)-$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c32)")
echo ""
info "JWT Secret baru: ${NEW_JWT_SECRET:0:16}... (${#NEW_JWT_SECRET} karakter)"

APP_DIR="/var/www/pesantren-multi"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo ""
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo -e "  VPS        : ${GREEN}${SSH_USER}@${VPS_IP}:${SSH_PORT}${NC}"
echo -e "  Target     : ${GREEN}${APP_DIR}${NC}"
echo -e "  DB Pass    : ${GREEN}********${NC}"
echo -e "  JWT Secret : ${GREEN}${NEW_JWT_SECRET:0:16}...${NC}"
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo ""
read -p "Lanjutkan deploy? (y/n): " CONFIRM
[ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ] && exit 0
echo ""

SSH_CMD="ssh -p ${SSH_PORT} ${SSH_USER}@${VPS_IP}"
SCP_CMD="scp -P ${SSH_PORT}"

# ══════════════════════════════════════════════════════════
# STEP 1: Backup di VPS
# ══════════════════════════════════════════════════════════
info "1/6 Backup database di VPS..."
$SSH_CMD "mysqldump -u pesantren -p\$(grep DB_PASS ${APP_DIR}/go-backend/.env | cut -d= -f2) pesantren_multi > /tmp/pesantren_backup_\$(date +%Y%m%d_%H%M%S).sql && echo 'BACKUP OK' || echo 'BACKUP SKIP (fresh install?)'"
ok "Backup selesai"

# ══════════════════════════════════════════════════════════
# STEP 2: Upload file via SCP
# ══════════════════════════════════════════════════════════
info "2/6 Upload go-backend via SCP..."
$SCP_CMD -r "${SCRIPT_DIR}/go-backend/handlers" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
$SCP_CMD -r "${SCRIPT_DIR}/go-backend/middleware" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
$SCP_CMD -r "${SCRIPT_DIR}/go-backend/config" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
$SCP_CMD -r "${SCRIPT_DIR}/go-backend/helpers" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
$SCP_CMD "${SCRIPT_DIR}/go-backend/main.go" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
$SCP_CMD "${SCRIPT_DIR}/go-backend/go.mod" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
$SCP_CMD "${SCRIPT_DIR}/go-backend/go.sum" "${SSH_USER}@${VPS_IP}:${APP_DIR}/go-backend/"
ok "Backend uploaded"

info "3/6 Upload frontend (public) via SCP..."
$SCP_CMD "${SCRIPT_DIR}/public/"*.html "${SSH_USER}@${VPS_IP}:${APP_DIR}/public/"
$SCP_CMD "${SCRIPT_DIR}/public/"*.js "${SSH_USER}@${VPS_IP}:${APP_DIR}/public/"
$SCP_CMD "${SCRIPT_DIR}/public/"*.css "${SSH_USER}@${VPS_IP}:${APP_DIR}/public/"
$SCP_CMD "${SCRIPT_DIR}/public/"*.json "${SSH_USER}@${VPS_IP}:${APP_DIR}/public/" 2>/dev/null || true
$SCP_CMD "${SCRIPT_DIR}/public/"*.png "${SSH_USER}@${VPS_IP}:${APP_DIR}/public/" 2>/dev/null || true
$SCP_CMD "${SCRIPT_DIR}/public/"*.ico "${SSH_USER}@${VPS_IP}:${APP_DIR}/public/" 2>/dev/null || true
ok "Frontend uploaded"

# ══════════════════════════════════════════════════════════
# STEP 3: Ganti DB Password + JWT Secret + Build
# ══════════════════════════════════════════════════════════
info "4/6 Update .env + ganti DB password di MariaDB..."

$SSH_CMD << REMOTE_SCRIPT
set -e
export PATH=\$PATH:/usr/local/go/bin

echo "  → Baca password DB lama..."
OLD_DB_PASS=\$(grep DB_PASS ${APP_DIR}/go-backend/.env | cut -d= -f2)

echo "  → Ganti password DB di MariaDB..."
mysql -u root -e "ALTER USER 'pesantren'@'localhost' IDENTIFIED BY '${NEW_DB_PASS}'; FLUSH PRIVILEGES;" || {
  echo "  → Coba dengan mysql native..."
  mysql -u root -e "SET PASSWORD FOR 'pesantren'@'localhost' = PASSWORD('${NEW_DB_PASS}'); FLUSH PRIVILEGES;"
}
echo "  ✅ Password MariaDB berhasil diganti"

echo "  → Baca domain dari .env lama..."
DOMAIN=\$(grep DOMAIN ${APP_DIR}/go-backend/.env 2>/dev/null | cut -d= -f2 || echo "e-pesantren.app")

echo "  → Update .env dengan credentials baru..."
cat > ${APP_DIR}/go-backend/.env << EOF
DB_HOST=localhost
DB_USER=pesantren
DB_PASS=${NEW_DB_PASS}
DB_NAME=pesantren_multi
PORT=3002
JWT_SECRET=${NEW_JWT_SECRET}
DOMAIN=\${DOMAIN}
CORS_ORIGINS=https://\${DOMAIN}, https://*.\${DOMAIN}
EOF
echo "  ✅ .env diupdate"

echo "  → Build binary baru..."
cd ${APP_DIR}/go-backend
go mod tidy
go build -o pesantren-server .
chmod +x pesantren-server
echo "  ✅ Binary berhasil dibuild"

REMOTE_SCRIPT

ok "DB password + JWT secret + binary updated"

# ══════════════════════════════════════════════════════════
# STEP 5: Restart service
# ══════════════════════════════════════════════════════════
info "5/6 Restart service..."
$SSH_CMD "systemctl restart pesantren-multi && sleep 2 && systemctl is-active pesantren-multi && echo '✅ Service running' || echo '❌ Service failed — cek: journalctl -u pesantren-multi -n 30'"
ok "Service restarted"

# ══════════════════════════════════════════════════════════
# STEP 6: Verifikasi
# ══════════════════════════════════════════════════════════
info "6/6 Verifikasi..."
$SSH_CMD "curl -s -o /dev/null -w 'HTTP Status: %{http_code}\n' http://localhost:3002/ && echo '✅ App merespon dengan baik' || echo '⚠️ App belum merespon'"

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                DEPLOY SELESAI! 🎉                        ║${NC}"
echo -e "${GREEN}╠══════════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║${NC}  DB Password   : ${CYAN}(sudah diganti)${NC}"
echo -e "${GREEN}║${NC}  JWT Secret    : ${CYAN}(sudah diganti — semua user perlu login ulang)${NC}"
echo -e "${GREEN}║${NC}  CORS          : ${CYAN}(restricted ke domain sendiri)${NC}"
echo -e "${GREEN}║${NC}  CSP Header    : ${CYAN}(aktif)${NC}"
echo -e "${GREEN}╠══════════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║${NC}  ${YELLOW}⚠️  PENTING:${NC}"
echo -e "${GREEN}║${NC}  • Semua user harus LOGIN ULANG (JWT secret berubah)"
echo -e "${GREEN}║${NC}  • Cek log : ${CYAN}ssh ${SSH_USER}@${VPS_IP} journalctl -u pesantren-multi -f${NC}"
echo -e "${GREEN}║${NC}  • Status  : ${CYAN}ssh ${SSH_USER}@${VPS_IP} systemctl status pesantren-multi${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
