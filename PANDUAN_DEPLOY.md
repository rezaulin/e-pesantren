# Panduan Lengkap Build & Deploy E-Pesantren v20

Dokumen ini berisi langkah-langkah standar setiap kali Anda melakukan perubahan kode (baik Frontend maupun Backend) pada komputer lokal (Windows) dan ingin mengunggahnya ke server production (VPS Ubuntu Linux).

---

## BAGIAN 1: PROSES BUILD DI LOKAL (WINDOWS)

### A. Build Backend (Go)
Karena server VPS menggunakan Linux, Anda wajib mem-build aplikasi Go menggunakan format Linux (meskipun Anda nge-build dari Windows).
Gunakan **PowerShell** untuk proses ini (jangan gunakan CMD biasa).

1. Buka terminal PowerShell di dalam folder `go-backend`:
   ```powershell
   cd "d:\abi\backup pesantren\pesantren v20\go-backend"
   ```
2. Atur environment variable untuk Linux dan jalankan build:
   ```powershell
   $env:GOOS="linux"
   $env:GOARCH="amd64"
   go build -o pesantren-server
   ```
*(Akan menghasilkan file `pesantren-server` tanpa ekstensi .exe)*

### B. Build Frontend (Javascript & HTML)
Setiap kali Anda merubah file JS di folder `public` (misal: `app-core.js`), Anda harus mengecilkan ukurannya (minify) dan mengupdate cache buster di `index.html`.

1. Buka terminal di folder root project:
   ```powershell
   cd "d:\abi\backup pesantren\pesantren v20"
   ```
2. Jalankan script build (pastikan Node.js terinstall):
   ```powershell
   node build_frontend.js
   ```
3. **SANGAT PENTING**: Buka file `public/index.html`, cari bagian `<script defer src="dist/app-core-v2.min.js?v=..."></script>` dan **ubah angka versinya** (misalnya dari `?v=123` menjadi `?v=124`). Jika tidak diubah, browser wali/bendahara tidak akan bisa melihat update terbaru karena nyangkut di cache browser!

---

## BAGIAN 2: UPLOAD KE SERVER VPS (SCP)

Karena user `rezaulin` (atau `ubuntu`) tidak memiliki hak akses langsung untuk menimpa file yang sedang berjalan di `/var/www/`, Anda harus memindahkannya ke folder Home (`~/`) terlebih dahulu.

1. Buka terminal PowerShell. Pastikan kunci SSH (`jancok123.pem`) ada di folder Downloads.
2. Upload file **Backend**:
   ```powershell
   cd "d:\abi\backup pesantren\pesantren v20\go-backend"
   scp -i "C:\Users\DELL\Downloads\jancok123.pem" pesantren-server rezaulin@103.175.216.218:~/
   ```
3. Upload folder **Frontend (public)**:
   ```powershell
   cd "d:\abi\backup pesantren\pesantren v20"
   scp -r -i "C:\Users\DELL\Downloads\jancok123.pem" public rezaulin@103.175.216.218:~/
   ```

*(Catatan: Ganti `rezaulin` menjadi user SSH server Anda yang sebenarnya jika berbeda).*

---

## BAGIAN 3: TERAPKAN PEMBARUAN DI VPS

Setelah file berhasil terkirim 100%, sekarang kita masuk ke VPS untuk menerapkan perubahannya.

1. Masuk ke VPS menggunakan SSH:
   ```powershell
   ssh -i "C:\Users\DELL\Downloads\jancok123.pem" rezaulin@103.175.216.218
   ```
2. Setelah berhasil login, jadikan diri Anda sebagai Root (Administrator):
   ```bash
   sudo su
   ```
3. Matikan service aplikasi sementara agar file tidak terkunci:
   ```bash
   systemctl stop pesantren-multi
   ```
4. Pindahkan file Backend dan berikan izin eksekusi (`chmod +x`):
   ```bash
   mv /home/rezaulin/pesantren-server /var/www/pesantren-multi/go-backend/
   chmod +x /var/www/pesantren-multi/go-backend/pesantren-server
   ```
5. Pindahkan dan timpa folder Frontend:
   ```bash
   cp -r /home/rezaulin/public/* /var/www/pesantren-multi/public/
   ```
6. Nyalakan kembali service aplikasinya:
   ```bash
   systemctl start pesantren-multi
   systemctl status pesantren-multi
   ```
*(Jika indikatornya berwarna hijau `active (running)`, maka proses update berhasil! Silakan refresh browser Anda).*
