package config

import "fmt"

// AutoMigrate ensures all required tables and columns exist
func AutoMigrate() {
	// Create tables if not exist
	tables := []string{
		`CREATE TABLE IF NOT EXISTS kelas_sekolah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(100) NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_ks_tid (tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS santri_kelas (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			kelas_id INT NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_sk_tid (tenant_id),
			INDEX idx_sk_kid (kelas_id),
			INDEX idx_sk_sid (santri_id)
		)`,
		`CREATE TABLE IF NOT EXISTS jadwal_sekolah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			kelas_id INT NOT NULL,
			mata_pelajaran VARCHAR(200) NOT NULL,
			ustadz_username VARCHAR(100) NOT NULL,
			hari VARCHAR(20) NOT NULL,
			jam_mulai VARCHAR(10) NOT NULL,
			jam_selesai VARCHAR(10) NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_js_tid (tenant_id),
			INDEX idx_js_kid (kelas_id)
		)`,
		`CREATE TABLE IF NOT EXISTS absen_sekolah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			kelas_id INT DEFAULT NULL,
			jadwal_sekolah_id INT DEFAULT NULL,
			mata_pelajaran VARCHAR(200) DEFAULT NULL,
			tanggal VARCHAR(20) NOT NULL,
			status VARCHAR(10) NOT NULL,
			keterangan TEXT,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_as_tid (tenant_id),
			INDEX idx_as_kid (kelas_id)
		)`,
		`CREATE TABLE IF NOT EXISTS absen_malam (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			kamar_id INT DEFAULT NULL,
			tanggal VARCHAR(20) NOT NULL,
			status VARCHAR(10) NOT NULL,
			keterangan TEXT,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_am_tid (tenant_id),
			INDEX idx_am_kid (kamar_id)
		)`,
		`CREATE TABLE IF NOT EXISTS pembayaran_tarif (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			aktif TINYINT(1) DEFAULT 1,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pt_tid (tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS pembayaran_potongan (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			aktif TINYINT(1) DEFAULT 1,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pp_tid (tenant_id),
			INDEX idx_pp_sid (santri_id)
		)`,
		`CREATE TABLE IF NOT EXISTS pembayaran (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			bulan VARCHAR(7) NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			metode VARCHAR(50) DEFAULT 'tunai',
			keterangan TEXT,
			created_by INT NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pb_tid (tenant_id),
			INDEX idx_pb_sid (santri_id),
			INDEX idx_pb_bulan (bulan)
		)`,
		`CREATE TABLE IF NOT EXISTS catatan_keuangan (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			tipe ENUM('masuk','keluar') NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			keterangan TEXT,
			tanggal DATE NOT NULL,
			created_by INT NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_ck_tid (tenant_id),
			INDEX idx_ck_tgl (tanggal)
		)`,
		`CREATE TABLE IF NOT EXISTS absen_sekolah_sesi (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			kelas_id INT NOT NULL,
			jadwal_sekolah_id INT DEFAULT NULL,
			tanggal VARCHAR(20) NOT NULL,
			ustadz_username VARCHAR(100) DEFAULT '',
			recorded_at DATETIME DEFAULT NOW(),
			INDEX idx_ass_tid (tenant_id),
			INDEX idx_ass_tgl (tanggal)
		)`,
		// ── Madrasah Diniyyah Tables ──────────────────────────
		`CREATE TABLE IF NOT EXISTS kelas_diniyyah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(100) NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_kd_tid (tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS santri_kelas_diniyyah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			kelas_diniyyah_id INT NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_skd_tid (tenant_id),
			INDEX idx_skd_kid (kelas_diniyyah_id),
			INDEX idx_skd_sid (santri_id)
		)`,
		// ── Penilaian Tables ────────────────────────────────
		`CREATE TABLE IF NOT EXISTS mata_pelajaran (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			kelas_diniyyah_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_mp_tid (tenant_id),
			INDEX idx_mp_kdid (kelas_diniyyah_id)
		)`,
		`CREATE TABLE IF NOT EXISTS nilai_pelajaran (
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
			INDEX idx_np_tid (tenant_id),
			INDEX idx_np_sid (santri_id),
			INDEX idx_np_mpid (mata_pelajaran_id),
			INDEX idx_np_kdid (kelas_diniyyah_id),
			INDEX idx_np_smt (semester),
			UNIQUE KEY uq_np (tenant_id, santri_id, mata_pelajaran_id, semester)
		)`,
		`CREATE TABLE IF NOT EXISTS nilai_kegiatan (
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
			INDEX idx_nk_tid (tenant_id),
			INDEX idx_nk_sid (santri_id),
			INDEX idx_nk_kgid (kegiatan_id),
			INDEX idx_nk_klid (kelompok_id),
			INDEX idx_nk_bln (bulan),
			UNIQUE KEY uq_nk (tenant_id, santri_id, kegiatan_id, kelompok_id, bulan)
		)`,
		// ── Penilaian Sekolah Tables ───────────────────────────
		`CREATE TABLE IF NOT EXISTS mata_pelajaran_sekolah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			kelas_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_mps_tid (tenant_id),
			INDEX idx_mps_kid (kelas_id)
		)`,
		`CREATE TABLE IF NOT EXISTS nilai_sekolah (
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
			INDEX idx_ns_tid (tenant_id),
			INDEX idx_ns_sid (santri_id),
			INDEX idx_ns_mpsid (mata_pelajaran_sekolah_id),
			INDEX idx_ns_kid (kelas_id),
			INDEX idx_ns_smt (semester),
			UNIQUE KEY uq_ns (tenant_id, santri_id, mata_pelajaran_sekolah_id, semester)
		)`,
		// ── Absensi Diniyyah ──────────────────────────────────
		`CREATE TABLE IF NOT EXISTS absen_diniyyah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			kelas_diniyyah_id INT NOT NULL,
			tanggal DATE NOT NULL,
			status VARCHAR(1) DEFAULT 'H',
			recorded_by INT NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_ad_tid (tenant_id),
			INDEX idx_ad_sid (santri_id),
			INDEX idx_ad_kdid (kelas_diniyyah_id),
			INDEX idx_ad_tgl (tanggal),
			UNIQUE KEY uq_ad (tenant_id, santri_id, kelas_diniyyah_id, tanggal)
		)`,
		// ── Jadwal Diniyyah ──────────────────────────────────
		`CREATE TABLE IF NOT EXISTS jadwal_diniyyah (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			kelas_diniyyah_id INT NOT NULL,
			mata_pelajaran VARCHAR(100) NOT NULL,
			ustadz_username VARCHAR(100) DEFAULT '',
			hari VARCHAR(20) DEFAULT '',
			INDEX idx_jd_tid (tenant_id),
			INDEX idx_jd_kdid (kelas_diniyyah_id)
		)`,
		`CREATE TABLE IF NOT EXISTS pembayaran_kategori (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			aktif TINYINT(1) DEFAULT 1,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pk_tid (tenant_id)
		)`,
		// ── Tarif & Kategori Per Periode ─────────────────────
		`CREATE TABLE IF NOT EXISTS pembayaran_tarif_periode (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			tarif_id INT NOT NULL,
			bulan_mulai VARCHAR(7) NOT NULL,
			bulan_akhir VARCHAR(7) NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_ptp_tid (tenant_id),
			INDEX idx_ptp_tarifid (tarif_id),
			INDEX idx_ptp_bulan (bulan_mulai, bulan_akhir)
		)`,
		`CREATE TABLE IF NOT EXISTS pembayaran_kategori_periode (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			kategori_id INT NOT NULL,
			bulan_mulai VARCHAR(7) NOT NULL,
			bulan_akhir VARCHAR(7) NOT NULL,
			nominal BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pkp_tid (tenant_id),
			INDEX idx_pkp_katid (kategori_id),
			INDEX idx_pkp_bulan (bulan_mulai, bulan_akhir)
		)`,

		`CREATE TABLE IF NOT EXISTS perizinan (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			keterangan TEXT,
			tipe_durasi VARCHAR(10) DEFAULT 'hari',
			durasi INT DEFAULT 1,
			tanggal_mulai DATETIME NOT NULL,
			tanggal_selesai DATETIME NOT NULL,
			status VARCHAR(20) DEFAULT 'aktif',
			terlambat_durasi INT DEFAULT 0,
			terlambat_tipe VARCHAR(10) DEFAULT 'hari',
			tanggal_kembali DATETIME DEFAULT NULL,
			created_by INT NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pz_tid (tenant_id),
			INDEX idx_pz_sid (santri_id),
			INDEX idx_pz_status (status),
			INDEX idx_pz_mulai (tanggal_mulai),
			INDEX idx_pz_selesai (tanggal_selesai)
		)`,
		// ── E-Paket (Logistik Paket Santri) ──────────────────
		`CREATE TABLE IF NOT EXISTS paket (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			pengirim VARCHAR(300) DEFAULT '',
			keterangan TEXT,
			status ENUM('DI_GERBANG','DI_ASRAMA','SELESAI') DEFAULT 'DI_GERBANG',
			posisi_rak VARCHAR(100) DEFAULT '',
			waktu_gerbang DATETIME DEFAULT NOW(),
			waktu_asrama DATETIME DEFAULT NULL,
			waktu_diterima DATETIME DEFAULT NULL,
			created_by INT NOT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_pkt_tid (tenant_id),
			INDEX idx_pkt_sid (santri_id),
			INDEX idx_pkt_status (status)
		)`,
		// ── Tahfidz Qur'an ──────────────────────────────────
		`CREATE TABLE IF NOT EXISTS halaqoh (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			musyrif VARCHAR(200) DEFAULT '',
			ustadz_username VARCHAR(100) DEFAULT '',
			target_juz VARCHAR(100) DEFAULT '',
			target_surah VARCHAR(500) DEFAULT '',
			target_ayat INT DEFAULT 0,
			target_deadline DATE DEFAULT NULL,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_hlq_tid (tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS halaqoh_members (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			halaqoh_id INT NOT NULL,
			santri_id INT NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_hm_tid (tenant_id),
			INDEX idx_hm_hid (halaqoh_id),
			INDEX idx_hm_sid (santri_id),
			UNIQUE KEY uq_hm (tenant_id, halaqoh_id, santri_id)
		)`,
		`CREATE TABLE IF NOT EXISTS tahfidz_sesi (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			halaqoh_id INT NOT NULL,
			tanggal DATE NOT NULL,
			ustadz_username VARCHAR(100) NOT NULL,
			recorded_at DATETIME DEFAULT NOW(),
			INDEX idx_ts_tid (tenant_id),
			INDEX idx_ts_tgl (tanggal),
			UNIQUE KEY uq_ts (tenant_id, halaqoh_id, tanggal)
		)`,
		`CREATE TABLE IF NOT EXISTS tahfidz_nilai (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			sesi_id INT NOT NULL,
			santri_id INT NOT NULL,
			halaqoh_id INT NOT NULL,
			tanggal DATE NOT NULL,
			status CHAR(1) NOT NULL DEFAULT 'H',
			jenis_setoran VARCHAR(10) DEFAULT 'ziyadah',
			surah VARCHAR(100) DEFAULT '',
			ayat_dari INT DEFAULT 0,
			ayat_sampai INT DEFAULT 0,
			juz INT DEFAULT 0,
			nilai_tajwid TINYINT UNSIGNED DEFAULT 0,
			nilai_kelancaran TINYINT UNSIGNED DEFAULT 0,
			nilai_makhorijul TINYINT UNSIGNED DEFAULT 0,
			nilai_rata DECIMAL(5,2) DEFAULT 0,
			predikat VARCHAR(20) DEFAULT '',
			catatan TEXT,
			created_at DATETIME DEFAULT NOW(),
			updated_at DATETIME DEFAULT NOW(),
			INDEX idx_tn_tid (tenant_id),
			INDEX idx_tn_sesi (sesi_id),
			INDEX idx_tn_sid (santri_id),
			INDEX idx_tn_hid (halaqoh_id),
			INDEX idx_tn_tgl (tanggal)
		)`,
		`CREATE TABLE IF NOT EXISTS sangu_merchants (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			deskripsi TEXT,
			saldo BIGINT DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_sangu_m_tid (tenant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sangu_products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			merchant_id INT NOT NULL,
			nama VARCHAR(200) NOT NULL,
			harga BIGINT NOT NULL,
			stok INT DEFAULT 0,
			barcode_sku VARCHAR(100),
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_sangu_p_tid (tenant_id),
			INDEX idx_sangu_p_mid (merchant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sangu_transaksi (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			merchant_id INT NOT NULL,
			kasir_user_id INT,
			total BIGINT NOT NULL,
			status VARCHAR(20) DEFAULT 'sukses',
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_sangu_tr_tid (tenant_id),
			INDEX idx_sangu_tr_sid (santri_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sangu_transaksi_items (
			id INT AUTO_INCREMENT PRIMARY KEY,
			transaksi_id INT NOT NULL,
			product_id INT NOT NULL,
			qty INT NOT NULL,
			harga BIGINT NOT NULL,
			subtotal BIGINT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sangu_topup (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			nominal BIGINT NOT NULL,
			fee BIGINT DEFAULT 0,
			total BIGINT NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			payment_method VARCHAR(50),
			payment_ref VARCHAR(100),
			payment_url TEXT,
			created_at DATETIME DEFAULT NOW(),
			INDEX idx_sangu_tp_tid (tenant_id),
			INDEX idx_sangu_tp_sid (santri_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sangu_withdrawals (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			merchant_id INT NOT NULL,
			nominal BIGINT NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			keterangan TEXT,
			approved_by INT,
			created_at DATETIME DEFAULT NOW(),
			updated_at DATETIME,
			INDEX idx_sangu_w_tid (tenant_id),
			INDEX idx_sangu_w_mid (merchant_id)
		)`,
	}
	for _, q := range tables {
		if _, err := DB.Exec(q); err != nil {
			fmt.Println("  Table warning:", err.Error())
		}
	}

	// Add missing columns to existing tables (safe — ignores if already exist)
	alters := []struct{ table, col, def string }{
		{"jadwal_sekolah", "kelas_id", "INT NOT NULL DEFAULT 0"},
		{"jadwal_sekolah", "mata_pelajaran", "VARCHAR(200) NOT NULL DEFAULT ''"},
		{"jadwal_sekolah", "ustadz_username", "VARCHAR(100) NOT NULL DEFAULT ''"},
		{"jadwal_sekolah", "hari", "VARCHAR(20) NOT NULL DEFAULT ''"},
		{"jadwal_sekolah", "jam_mulai", "VARCHAR(10) NOT NULL DEFAULT ''"},
		{"jadwal_sekolah", "jam_selesai", "VARCHAR(10) NOT NULL DEFAULT ''"},
		{"absen_malam", "kamar_id", "INT DEFAULT NULL"},
		{"absen_sekolah", "kelas_id", "INT DEFAULT NULL"},
		{"absen_sekolah", "jadwal_sekolah_id", "INT DEFAULT NULL"},
		{"absen_sekolah", "mata_pelajaran", "VARCHAR(200) DEFAULT NULL"},
		{"santri", "nama_wali", "VARCHAR(200) DEFAULT ''"},
		{"santri", "no_hp", "VARCHAR(20) DEFAULT ''"},
		{"absen_diniyyah", "jadwal_diniyyah_id", "INT DEFAULT NULL"},
		{"absen_diniyyah", "mata_pelajaran", "VARCHAR(200) DEFAULT ''"},
		{"absen_diniyyah", "ustadz_username", "VARCHAR(100) DEFAULT ''"},
		{"absen_sekolah", "ustadz_username", "VARCHAR(100) DEFAULT ''"},
		{"absen_sekolah_sesi", "mata_pelajaran", "VARCHAR(200) DEFAULT ''"},
		{"santri", "tempat_lahir", "VARCHAR(100) DEFAULT ''"},
		{"santri", "tanggal_lahir", "DATE DEFAULT NULL"},
		{"santri", "jenis_kelamin", "VARCHAR(2) DEFAULT 'L'"},
		{"santri", "asal_sekolah", "VARCHAR(200) DEFAULT ''"},
		{"santri", "catatan_psb", "TEXT"},
		{"santri", "nama_ayah", "VARCHAR(200) DEFAULT ''"},
		{"santri", "nama_ibu", "VARCHAR(200) DEFAULT ''"},
		{"santri", "kategori_spp_id", "INT DEFAULT NULL"},
		{"settings", "rekening_bank", "VARCHAR(100) DEFAULT ''"},
		{"settings", "rekening_nomor", "VARCHAR(100) DEFAULT ''"},
		{"settings", "rekening_atas_nama", "VARCHAR(100) DEFAULT ''"},
		{"settings", "bendahara_telp", "VARCHAR(100) DEFAULT ''"},
		{"users", "merchant_id", "INT DEFAULT NULL"},
		{"settings", "psb_kuota", "INT DEFAULT 0"},
		// Takzir columns on pelanggaran
		{"pelanggaran", "status_takzir", "VARCHAR(10) DEFAULT 'belum'"},
		{"pelanggaran", "jenis_takzir", "VARCHAR(200) DEFAULT ''"},
		{"pelanggaran", "denda", "BIGINT DEFAULT 0"},
		{"pelanggaran", "takzir_at", "DATETIME DEFAULT NULL"},
		// RFID/NFC Card UID
		{"santri", "card_uid", "VARCHAR(50) DEFAULT NULL"},
		// Raport v2: Header and KKM
		{"settings", "nama_lembaga_sekolah", "VARCHAR(200) DEFAULT ''"},
		{"settings", "header_sekolah_line1", "VARCHAR(200) DEFAULT ''"},
		{"settings", "header_sekolah_line2", "VARCHAR(200) DEFAULT ''"},
		{"settings", "header_sekolah_alamat", "VARCHAR(300) DEFAULT ''"},
		{"settings", "logo_sekolah", "LONGTEXT"},
		{"settings", "nama_lembaga_diniyyah", "VARCHAR(200) DEFAULT ''"},
		{"settings", "header_diniyyah_alamat", "VARCHAR(300) DEFAULT ''"},
		{"settings", "logo_diniyyah", "LONGTEXT"},
		{"settings", "nama_lembaga_kegiatan", "VARCHAR(200) DEFAULT ''"},
		{"settings", "header_kegiatan_alamat", "VARCHAR(300) DEFAULT ''"},
		{"settings", "logo_kegiatan", "LONGTEXT"},
		{"nilai_pelajaran", "kkm", "DECIMAL(5,2) DEFAULT 75"},
		{"nilai_sekolah", "kkm", "DECIMAL(5,2) DEFAULT 75"},
	}
	// Update users role ENUM to include all roles
	DB.Exec("ALTER TABLE users MODIFY COLUMN role ENUM('superadmin','admin','ustadz','wali','bendahara','keamanan','kasir','merchant_admin') NOT NULL")

	// Add features column to tenants (JSON array of enabled feature keys)
	DB.Exec("ALTER TABLE tenants ADD COLUMN features TEXT")

	for _, a := range alters {
		q := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", a.table, a.col, a.def)
		_, err := DB.Exec(q)
		if err != nil {
			// "Duplicate column" is expected and fine
			if err.Error() != "" {
				// silently ignore — column already exists
			}
		}
	}

	// Add UNIQUE constraint on users(tenant_id, username) to prevent duplicate wali accounts
	DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uq_users_tenant_username ON users (tenant_id, username)")

	// Allow multiple absen_diniyyah per day per kelas (different jadwal = different record)
	DB.Exec("ALTER TABLE absen_diniyyah DROP INDEX uq_ad")
	DB.Exec("CREATE UNIQUE INDEX uq_ad2 ON absen_diniyyah (tenant_id, santri_id, kelas_diniyyah_id, tanggal, jadwal_diniyyah_id)")

	// ═══════════════════════════════════════════════════════════
	// TAHAP 1 OPTIMASI: Hemat Storage & RAM
	// ═══════════════════════════════════════════════════════════
	fmt.Println("  Running storage optimizations...")

	// ── 1A. Perkecil tipe data kolom ─────────────────────────
	// tanggal: VARCHAR(20) → DATE (21 bytes → 3 bytes per baris)
	// status: VARCHAR(10) → CHAR(1) (11 bytes → 1 byte per baris)
	// keterangan: TEXT → VARCHAR(500) (hemat overhead)
	// Catatan: MODIFY COLUMN aman untuk data existing selama format cocok
	dataTypeOptimizations := []string{
		// ── Tabel absensi (kegiatan) ──
		"ALTER TABLE absensi MODIFY COLUMN tanggal DATE NOT NULL",
		"ALTER TABLE absensi MODIFY COLUMN status CHAR(1) NOT NULL DEFAULT 'H'",
		"ALTER TABLE absensi MODIFY COLUMN keterangan VARCHAR(500) DEFAULT NULL",

		// ── Tabel absensi_sesi ──
		"ALTER TABLE absensi_sesi MODIFY COLUMN tanggal DATE NOT NULL",
		"ALTER TABLE absensi_sesi MODIFY COLUMN ustadz_username VARCHAR(50) NOT NULL",

		// ── Tabel absen_malam ──
		"ALTER TABLE absen_malam MODIFY COLUMN tanggal DATE NOT NULL",
		"ALTER TABLE absen_malam MODIFY COLUMN status CHAR(1) NOT NULL DEFAULT 'H'",
		"ALTER TABLE absen_malam MODIFY COLUMN keterangan VARCHAR(500) DEFAULT NULL",

		// ── Tabel absen_sekolah ──
		"ALTER TABLE absen_sekolah MODIFY COLUMN tanggal DATE NOT NULL",
		"ALTER TABLE absen_sekolah MODIFY COLUMN status CHAR(1) NOT NULL DEFAULT 'H'",
		"ALTER TABLE absen_sekolah MODIFY COLUMN keterangan VARCHAR(500) DEFAULT NULL",
		"ALTER TABLE absen_sekolah MODIFY COLUMN mata_pelajaran VARCHAR(100) DEFAULT NULL",
		"ALTER TABLE absen_sekolah MODIFY COLUMN ustadz_username VARCHAR(50) DEFAULT ''",

		// ── Tabel absen_sekolah_sesi ──
		"ALTER TABLE absen_sekolah_sesi MODIFY COLUMN tanggal DATE NOT NULL",
		"ALTER TABLE absen_sekolah_sesi MODIFY COLUMN ustadz_username VARCHAR(50) DEFAULT ''",
		"ALTER TABLE absen_sekolah_sesi MODIFY COLUMN mata_pelajaran VARCHAR(100) DEFAULT ''",

		// ── Tabel absen_diniyyah ──
		"ALTER TABLE absen_diniyyah MODIFY COLUMN status CHAR(1) NOT NULL DEFAULT 'H'",
		"ALTER TABLE absen_diniyyah MODIFY COLUMN mata_pelajaran VARCHAR(100) DEFAULT ''",
		"ALTER TABLE absen_diniyyah MODIFY COLUMN ustadz_username VARCHAR(50) DEFAULT ''",

		// ── Tabel pelanggaran ──
		"ALTER TABLE pelanggaran MODIFY COLUMN tanggal DATE NOT NULL",
		"ALTER TABLE pelanggaran MODIFY COLUMN jenis VARCHAR(50) NOT NULL",

		// ── Tabel catatan_guru ──
		"ALTER TABLE catatan_guru MODIFY COLUMN tanggal DATE NOT NULL",
	}

	for _, q := range dataTypeOptimizations {
		if _, err := DB.Exec(q); err != nil {
			// Silently continue — some may fail if data format needs manual fix
			fmt.Printf("  [optimize] warning: %s\n", err.Error())
		}
	}

	// ── 1B. Composite indexes (menggantikan single-column indexes) ──
	// Composite index lebih efisien karena 1 index melayani beberapa pola query
	compositeIndexes := []string{
		// absensi: query selalu filter tenant_id + tanggal, sering + kelompok_id
		"CREATE INDEX idx_absensi_lookup ON absensi (tenant_id, tanggal, kelompok_id)",
		// absen_malam: query selalu filter tenant_id + tanggal, sering + kamar_id
		"CREATE INDEX idx_am_lookup ON absen_malam (tenant_id, tanggal, kamar_id)",
		// absen_sekolah: query selalu filter tenant_id + tanggal, sering + kelas_id
		"CREATE INDEX idx_as_lookup ON absen_sekolah (tenant_id, tanggal, kelas_id)",
		// absen_diniyyah: query selalu filter tenant_id + tanggal, sering + kelas_diniyyah_id
		"CREATE INDEX idx_adin_lookup ON absen_diniyyah (tenant_id, tanggal, kelas_diniyyah_id)",
		// absensi_sesi: sering query by tenant + tanggal + kelompok
		"CREATE INDEX idx_asesi_lookup ON absensi_sesi (tenant_id, tanggal, kelompok_id)",
		// absen_sekolah_sesi: sering query by tenant + tanggal + kelas
		"CREATE INDEX idx_assesi_lookup ON absen_sekolah_sesi (tenant_id, tanggal, kelas_id)",
		// pelanggaran: sering filter by tenant + santri + tanggal
		"CREATE INDEX idx_plg_lookup ON pelanggaran (tenant_id, santri_id, tanggal)",
	}

	for _, q := range compositeIndexes {
		DB.Exec(q) // silently ignore if already exists (Duplicate key name)
	}

	// ── 1B2. Composite indexes for wali dashboard queries ──
	// Wali queries always filter by (tenant_id, santri_id, tanggal) — need dedicated indexes
	waliIndexes := []string{
		"CREATE INDEX idx_absensi_wali ON absensi (tenant_id, santri_id, tanggal)",
		"CREATE INDEX idx_am_wali ON absen_malam (tenant_id, santri_id, tanggal)",
		"CREATE INDEX idx_as_wali ON absen_sekolah (tenant_id, santri_id, tanggal)",
		"CREATE INDEX idx_ad_wali ON absen_diniyyah (tenant_id, santri_id, tanggal)",
		"CREATE INDEX idx_cg_wali ON catatan_guru (tenant_id, santri_id, tanggal)",
		"CREATE INDEX idx_pb_wali ON pembayaran (tenant_id, santri_id, bulan)",
		"CREATE INDEX idx_pp_wali ON pembayaran_potongan (tenant_id, santri_id, aktif)",
	}
	for _, q := range waliIndexes {
		DB.Exec(q)
	}

	// ── 1C. InnoDB Page Compression ──────────────────────────
	// Hemat 40-60% disk untuk tabel besar tanpa ubah kode apapun
	compressionTables := []string{
		"ALTER TABLE absensi ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE absen_malam ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE absen_sekolah ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE absen_diniyyah ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE absensi_sesi ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE absen_sekolah_sesi ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE pelanggaran ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
		"ALTER TABLE catatan_guru ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8",
	}

	for _, q := range compressionTables {
		if _, err := DB.Exec(q); err != nil {
			fmt.Printf("  [compress] warning: %s\n", err.Error())
		}
	}

	fmt.Println("  Storage optimizations complete")

	// ── Add is_active column to users table for enable/disable wali accounts ──
	DB.Exec("ALTER TABLE users ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1")

	// ── Payment Gateway: payment_transactions table ──
	DB.Exec(`CREATE TABLE IF NOT EXISTS payment_transactions (
		id INT AUTO_INCREMENT PRIMARY KEY,
		tenant_id INT NOT NULL,
		santri_id INT NOT NULL,
		bulan VARCHAR(7) NOT NULL,
		nominal BIGINT NOT NULL,
		fee BIGINT NOT NULL DEFAULT 0,
		total_bayar BIGINT NOT NULL,
		order_id VARCHAR(100) NOT NULL,
		gateway VARCHAR(20) DEFAULT 'midtrans',
		snap_token TEXT DEFAULT '',
		payment_type VARCHAR(50) DEFAULT '',
		transaction_status VARCHAR(30) DEFAULT 'pending',
		gateway_response JSON,
		created_at DATETIME DEFAULT NOW(),
		updated_at DATETIME DEFAULT NOW(),
		INDEX idx_ptx_tid (tenant_id),
		INDEX idx_ptx_oid (order_id),
		INDEX idx_ptx_status (transaction_status)
	)`)
	// Drop UNIQUE on order_id (multi-month payments share one order_id)
	DB.Exec("ALTER TABLE payment_transactions DROP INDEX order_id")

	// ── Payment Gateway config columns in settings ──
	pgAlters := []struct{ col, def string }{
		{"pg_provider", "VARCHAR(20) DEFAULT ''"},
		{"pg_server_key", "VARCHAR(255) DEFAULT ''"},
		{"pg_client_key", "VARCHAR(255) DEFAULT ''"},
		{"pg_is_production", "TINYINT(1) DEFAULT 0"},
		{"pg_fee_percent", "DECIMAL(5,2) DEFAULT 0"},
		{"pg_fee_flat", "BIGINT DEFAULT 0"},
	}
	for _, a := range pgAlters {
		DB.Exec(fmt.Sprintf("ALTER TABLE settings ADD COLUMN %s %s", a.col, a.def))
	}

	// ── RFID/NFC Card UID unique index ──
	DB.Exec("CREATE UNIQUE INDEX uq_card_uid ON santri(tenant_id, card_uid)")

	// ── Bank Transfer: transfer bank + kode unik ──
	DB.Exec(`CREATE TABLE IF NOT EXISTS bank_transfers (
		id INT AUTO_INCREMENT PRIMARY KEY,
		tenant_id INT NOT NULL,
		santri_id INT NOT NULL,
		tipe VARCHAR(20) NOT NULL DEFAULT 'spp',
		bulan_list TEXT DEFAULT '',
		nominal BIGINT NOT NULL,
		fee BIGINT NOT NULL DEFAULT 0,
		kode_unik INT NOT NULL,
		total_transfer BIGINT NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		confirmed_by INT DEFAULT NULL,
		confirmed_at DATETIME DEFAULT NULL,
		expired_at DATETIME DEFAULT NULL,
		moota_mutation_id VARCHAR(100) DEFAULT '',
		created_at DATETIME DEFAULT NOW(),
		INDEX idx_bt_tid (tenant_id),
		INDEX idx_bt_sid (santri_id),
		INDEX idx_bt_status (status),
		INDEX idx_bt_match (tenant_id, total_transfer, status)
	)`)

	// ── Transfer Bank config columns in settings ──
	tbAlters := []struct{ col, def string }{
		{"transfer_bank_enabled", "TINYINT(1) DEFAULT 0"},
		{"transfer_bank_fee_flat", "BIGINT DEFAULT 0"},
		{"transfer_bank_expiry_hours", "INT DEFAULT 24"},
		{"moota_enabled", "TINYINT(1) DEFAULT 0"},
		{"moota_api_token", "VARCHAR(255) DEFAULT ''"},
		{"moota_secret_token", "VARCHAR(255) DEFAULT ''"},
		{"active_semester", "VARCHAR(50) DEFAULT ''"},
		{"active_month", "VARCHAR(50) DEFAULT ''"},
		{"ttd_sekolah_nama", "VARCHAR(100) DEFAULT ''"},
		{"ttd_sekolah_img", "LONGTEXT"},
		{"ttd_diniyyah_nama", "VARCHAR(100) DEFAULT ''"},
		{"ttd_diniyyah_img", "LONGTEXT"},
		{"ttd_kegiatan_nama", "VARCHAR(100) DEFAULT ''"},
		{"ttd_kegiatan_img", "LONGTEXT"},
		{"logo_base64", "LONGTEXT"},
		{"komponen_nilai_sekolah", "TEXT"},
		{"komponen_nilai_diniyyah", "TEXT"},
	}
	for _, a := range tbAlters {
		DB.Exec(fmt.Sprintf("ALTER TABLE settings ADD COLUMN %s %s", a.col, a.def))
	}

	// ── Fix inconsistent tipe values in catatan_keuangan ──
	// Some payment gateways/bank transfers used 'Pemasukan' instead of 'masuk'
	DB.Exec("UPDATE catatan_keuangan SET tipe = 'masuk' WHERE tipe = 'Pemasukan'")
	DB.Exec("UPDATE catatan_keuangan SET tipe = 'keluar' WHERE tipe = 'Pengeluaran'")

	// ── Tagihan bulan mulai (batas bawah penagihan SPP) ──
	DB.Exec("ALTER TABLE settings ADD COLUMN tagihan_bulan_mulai VARCHAR(7) DEFAULT ''")

	fmt.Println("  Database migration complete")
}
