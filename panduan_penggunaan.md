# Buku Panduan Penggunaan Aplikasi Manajemen Pesantren

Buku panduan ini merupakan dokumentasi lengkap mengenai fungsionalitas, kegunaan menu, dan tata cara penggunaan Aplikasi Manajemen Pesantren. Panduan ini ditujukan bagi seluruh pengelola dan pengguna di lingkungan pesantren (Admin, Ustadz, Bendahara, Pengelola Kantin, Kasir, dan Wali Santri).

---

## 1. Pendahuluan
Aplikasi ini dirancang untuk mendigitalisasi dan menyederhanakan manajemen administrasi, akademik, keuangan, dan logistik di lingkungan pesantren. Sistem ini berbasis *multi-role*, artinya setiap pengguna akan mendapatkan tampilan dan menu yang disesuaikan dengan tugas pokok dan fungsinya masing-masing.

---

## 2. Daftar Hak Akses (Role)
Sistem memiliki 6 tingkatan akses utama bagi pengelola dan pengguna pesantren, yaitu:
1. **Admin**: Administrator utama pesantren yang memiliki hak penuh untuk mengatur data master, santri, jadwal pelajaran, hingga pengaturan sistem.
2. **Ustadz / Guru**: Bertugas menangani kegiatan belajar mengajar, melakukan presensi, menginput nilai, serta mencatat pelanggaran dan perkembangan santri.
3. **Bendahara**: Bertanggung jawab penuh terhadap arus kas pesantren (uang masuk/keluar) serta administrasi pembayaran tagihan/SPP santri.
4. **Merchant Admin (Pengelola Kantin)**: Mengelola inventaris toko/kantin pesantren, menambah produk dagangan, dan melakukan penarikan saldo hasil penjualan.
5. **Kasir**: Petugas di lapangan yang melayani transaksi belanja santri di kantin menggunakan sistem *Tap* Kartu RFID atau pemotongan saldo uang saku (sangu).
6. **Wali Santri**: Orang tua/wali yang dapat masuk ke sistem untuk memantau kehadiran anak, nilai akademik, tagihan, dan melakukan *top-up* (isi ulang) uang saku anak dari rumah.

---

## 3. Penjelasan Fitur dan Tata Cara Penggunaan per Role

### A. ROLE: ADMIN (Administrator Utama)
Admin merupakan pusat kendali operasional pesantren. Menu yang tersedia bagi Admin mencakup seluruh aspek administrasi lembaga.

**1. Menu Beranda (Dashboard)**
- **Fungsi**: Pusat informasi yang menampilkan ringkasan data kegiatan dan situasi pesantren secara *real-time*.
- **Detail**: Menampilkan persentase kehadiran santri, jumlah kasus pelanggaran yang belum ditindaklanjuti, persentase SPP yang terkumpul, dan statistik jumlah santri aktif.
- **Tata Cara Penggunaan**: Admin dapat menggunakan **Tombol Aksi Cepat** di halaman ini untuk langsung melompat ke halaman *Tambah Santri*, *Input Absensi*, *Pembayaran*, *Catatan Guru*, atau *Pelanggaran* tanpa perlu mencari di menu samping.

**2. Menu Data Santri**
- **Fungsi**: Mengelola basis data seluruh santri yang terdaftar di pesantren.
- **Tata Cara Penggunaan**: 
  - **Mencari Santri**: Gunakan kolom pencarian di bagian atas untuk mencari nama santri, kamar, atau nama wali.
  - **Menambah Santri Baru**: Klik tombol **"+ Tambah"** di sudut kanan atas. Isi formulir pendaftaran seperti Nama, Kamar, Alamat, Nama Wali, Nomor HP Wali (untuk akses login wali), dan ID Kartu RFID (jika menggunakan kartu pintar). Lalu klik "Simpan".
  - **Import Data Massal**: Jika Anda memiliki data ratusan santri di Excel, klik tombol **"Import"** dan unggah file sesuai format template yang disediakan untuk memasukkan data secara instan.
  - **Edit/Hapus Data**: Pada baris nama santri di tabel, klik ikon pensil (Edit) untuk mengubah data, atau ikon tempat sampah (Hapus) untuk menghapus data santri yang sudah lulus/keluar.

**3. Menu PSB Online (Penerimaan Santri Baru)**
- **Fungsi**: Menyelenggarakan dan memantau pendaftaran santri baru secara *online*.
- **Tata Cara Penggunaan**: 
  - Di halaman PSB, klik **"Salin Link PSB"**.
  - Bagikan *link* (tautan) tersebut melalui WhatsApp atau brosur kepada calon pendaftar.
  - Setiap kali ada calon santri/wali yang mengisi formulir lewat tautan tersebut, datanya akan otomatis masuk ke tabel di halaman ini untuk direview oleh Admin.

**4. Menu Data Master (Kamar, Kelas, & Jadwal)**
- **Fungsi**: Menyusun fondasi data operasional sebelum sistem digunakan secara aktif.
- **Tata Cara Penggunaan**: 
  - **Kamar**: Masuk ke menu Kamar, klik "Tambah", lalu isi nama kamar dan kapasitasnya.
  - **Kelas Sekolah / Diniyyah**: Tambahkan nama-nama kelas yang ada di lembaga Anda, lalu masukkan santri ke dalam kelas masing-masing.
  - **Jadwal Pelajaran**: Pilih kelas, tentukan mata pelajaran, pilih Ustadz yang mengajar, hari, dan rentang jam (Mulai - Selesai).

**5. Menu E-Paket (Logistik Barang Santri)**
- **Fungsi**: Mengelola pencatatan kiriman paket barang untuk santri dari luar pesantren agar tidak hilang/terselip.
- **Tata Cara Penggunaan**:
  - Saat kurir datang membawa paket, Admin/Satpam membuka menu ini dan klik **Tambah Paket**.
  - Pilih nama santri penerima, ketik nama pengirim (misal: "Ibu Maryam via JNE"), dan set status ke **"Di Gerbang"**.
  - Saat paket sudah dibawa ke asrama, ubah status menjadi **"Di Asrama"**.
  - Saat paket sudah diambil oleh santri, ubah status menjadi **"Selesai"**.

**6. Menu Pengaturan (Settings)**
- **Fungsi**: Mengatur identitas pesantren dan aturan jam operasional.
- **Tata Cara Penggunaan**: Masuk ke menu Pengaturan. Di sini Anda bisa mengunggah Logo Pesantren, mengubah Nama Aplikasi, Nama Kepala Pesantren, dan mengatur batasan jam mulai/akhir untuk absensi otomatis.

---

### B. ROLE: USTADZ / GURU
Ustadz bertanggung jawab penuh pada kedisiplinan dan capaian akademik santri. Sebagian besar aktivitas Ustadz akan berkutat pada Menu Kedisiplinan dan Sekolah Formal/Diniyyah.

**1. Menu Absensi & Rekap Absensi**
- **Fungsi**: Melakukan presensi kehadiran harian (Sekolah, Diniyyah, atau Apel Malam/Asrama).
- **Tata Cara Penggunaan**: 
  - Buka menu Absensi.
  - Pilih kelas atau kelompok ngaji yang akan diajar.
  - Daftar nama santri akan muncul. Tandai status masing-masing santri: **Hadir**, **Izin**, **Sakit**, atau **Alpa**.
  - Klik **Simpan**. Data ini akan otomatis masuk ke rekap akademik dan langsung bisa dilihat oleh Wali Santri.

**2. Menu Catatan Guru**
- **Fungsi**: Memberikan laporan perkembangan (perilaku, prestasi, atau nasihat khusus) yang bersifat personal kepada wali santri.
- **Tata Cara Penggunaan**: 
  - Buka menu **Catatan Guru**.
  - Pada bagian form, ketik nama santri, pilih tanggal, dan tuliskan isi catatan (Contoh: "Ananda hari ini sudah hafal Juz 30 dengan sangat lancar").
  - Klik **Simpan**. Catatan tersebut akan permanen tersimpan di rekam jejak santri.

**3. Menu Pelanggaran**
- **Fungsi**: Mencatat indisipliner santri, memberikan poin pelanggaran, dan memantau status penyelesaian hukuman (*Takzir*).
- **Tata Cara Penggunaan**:
  - Di halaman Pelanggaran, isi form **Catat Pelanggaran**: Ketik nama santri, pilih jenis pelanggaran (misal: "Terlambat Shalat Subuh"), masukkan poin, dan tulis deskripsi detailnya. Klik **Simpan**.
  - **Penyelesaian Takzir**: Pada tabel *Daftar Pelanggaran*, pelanggaran baru akan berstatus "Belum" ditakzir. Jika santri sudah menjalankan hukumannya, Ustadz wajib mengklik tombol **Aksi (Centang)** di baris tersebut untuk mengubah statusnya menjadi selesai ("Sudah").

**4. Menu Penilaian**
- **Fungsi**: Menginput nilai Ulangan, UTS, UAS untuk rapor santri.
- **Tata Cara Penggunaan**: Pilih Kelas, pilih Mata Pelajaran, pilih Semester, lalu masukkan angka nilai (0-100) pada masing-masing santri, kemudian Simpan.

---

### C. ROLE: BENDAHARA
Fokus utama Bendahara adalah pencatatan finansial dan pemantauan pembayaran tagihan santri.

**1. Menu Pembayaran SPP / Tagihan**
- **Fungsi**: Mengelola pengaturan tarif bulanan dan mencatat pemasukan dari wali santri.
- **Tata Cara Penggunaan**: 
  - **Set Tarif**: Bendahara dapat mengatur kategori SPP (misalnya: Kategori Reguler Rp 500.000, Yatim Rp 0).
  - **Verifikasi Bayar**: Jika wali santri membayar tunai, Bendahara mencari nama santri, memilih bulan tagihan yang akan dibayar, dan mengklik konfirmasi "Bayar Tunai". Sistem akan mencatat lunas dan membuat tanda terima digital.

**2. Menu Catatan Keuangan**
- **Fungsi**: Mencatat seluruh arus kas di luar SPP, seperti uang donasi, atau pengeluaran belanja operasional pesantren.
- **Tata Cara Penggunaan**: Klik "Tambah Catatan", pilih tipe aliran dana (Uang Masuk / Uang Keluar), masukkan nominal (misal: Rp 1.500.000), isi keterangan (misal: "Beli alat kebersihan asrama"), lalu pilih tanggal dan klik Simpan.

---

### D. ROLE: WALI SANTRI
Akses khusus bagi orang tua untuk memantau anak secara transparan dari jauh.

**1. Dashboard Wali**
- **Fungsi**: Saat pertama *login*, wali santri akan melihat ringkasan tagihan bulan ini yang belum dibayar, sisa saldo uang saku (sangu) anak di kantin, dan grafik kehadiran anak selama sebulan terakhir.

**2. Menu Sangu (Uang Saku)**
- **Fungsi**: Untuk melakukan isi ulang (*top-up*) saldo belanja anak di kantin pesantren.
- **Tata Cara Penggunaan**: 
  - Buka menu Sangu, masukkan nominal (contoh: 100000).
  - Sistem akan mengarahkan wali ke instruksi transfer atau kode Virtual Account (VA).
  - Jika pembayaran berhasil, saldo akan otomatis masuk ke akun anak dan bisa langsung digunakan untuk jajan di kantin.

**3. Menu Laporan Akademik & Kedisiplinan**
- **Fungsi**: Memantau rapor nilai anak, melihat "Catatan Guru" yang diberikan oleh Ustadz, dan melihat apakah anak memiliki catatan "Pelanggaran" atau denda.

---

### E. ROLE: MERCHANT ADMIN (Pengelola Kantin / Koperasi)
Bertanggung jawab atas ketersediaan barang dan perputaran uang kantin.

**1. Menu Produk Kantin**
- **Fungsi**: Membuat katalog barang dagangan atau menu makanan.
- **Tata Cara Penggunaan**: Klik "Tambah Produk", ketik nama barang (misal: "Susu Beruang"), masukkan harga jual, masukkan stok awal, dan *scan* barcode barang menggunakan *scanner* (jika ada).

**2. Menu Penarikan Saldo (Withdrawal)**
- **Fungsi**: Memindahkan dana hasil penjualan dari sistem dompet digital pesantren ke kas fisik pengelola kantin.
- **Tata Cara Penggunaan**: Jika kantin sudah mendapatkan omzet (misal Rp 2.000.000), Merchant Admin mengajukan penarikan dana di menu ini. Setelah diverifikasi oleh Bendahara pusat, uang akan diserahkan secara fisik/transfer ke pengelola kantin.

---

### F. ROLE: KASIR KANTIN
Petugas kasir bertugas melayani transaksi santri di kantin menggunakan sistem Point of Sales (POS).

**1. Menu Point of Sales (POS)**
- **Fungsi**: Melakukan transaksi jual-beli tanpa uang tunai (*cashless*).
- **Tata Cara Penggunaan**: 
  1. Santri datang membawa barang belanjaan dan Kartu Santri (RFID).
  2. Kasir menempelkan (Tap) Kartu Santri ke alat pembaca (NFC/RFID Reader). Foto dan nama santri beserta sisa saldonya akan muncul di layar.
  3. Kasir men-scan *barcode* barang atau mengklik barang dari menu katalog di layar.
  4. Total belanja akan terkalkulasi otomatis. Klik **"Bayar"**.
  5. Saldo uang saku (sangu) santri akan langsung terpotong, dan saldo Merchant otomatis bertambah.

---

## 4. Alur Proses Operasional Harian (SOP Utama)

Untuk memahami bagaimana seluruh *Role* saling berkaitan, berikut adalah alur kerjanya:

- **Proses Administrasi & Setup**: 
  Admin menerima siswa baru via **PSB Online** -> Admin melengkapi **Data Santri** dan mendistribusikan santri ke Kamar & Kelas -> Admin mendaftarkan kartu RFID untuk masing-masing santri.
  
- **Proses KBM & Kedisiplinan**: 
  Setiap pagi, Ustadz membuka HP/Laptop untuk menginput **Absensi**. Jika ada santri yang berbuat ulah, Ustadz langsung membuka menu **Pelanggaran** dan memotong poin santri. Hasil input dari Ustadz ini secara *real-time* langsung masuk ke aplikasi **Wali Santri**.
  
- **Proses Jajan & Uang Saku**: 
  Wali Santri melakukan transfer **Top-up Sangu** dari rumah -> Saldo masuk ke akun anak -> Anak menggunakan kartu RFID untuk jajan ke **Kasir Kantin** -> Transaksi selesai tanpa uang tunai -> Malam harinya, **Merchant Admin** melihat total omzet dan mengajukan *Withdrawal* ke Bendahara.
