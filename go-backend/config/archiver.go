package config

import (
	"fmt"
	"time"
)

// ── Tabel arsip: data lama dipindah ke sini ──────────────

func createArchiveTables() {
	archiveTables := []string{
		// Arsip absensi kegiatan
		`CREATE TABLE IF NOT EXISTS absensi_arsip LIKE absensi`,
		// Arsip absensi sesi
		`CREATE TABLE IF NOT EXISTS absensi_sesi_arsip LIKE absensi_sesi`,
		// Arsip absen malam
		`CREATE TABLE IF NOT EXISTS absen_malam_arsip LIKE absen_malam`,
		// Arsip absen sekolah
		`CREATE TABLE IF NOT EXISTS absen_sekolah_arsip LIKE absen_sekolah`,
		// Arsip absen sekolah sesi
		`CREATE TABLE IF NOT EXISTS absen_sekolah_sesi_arsip LIKE absen_sekolah_sesi`,
		// Arsip absen diniyyah
		`CREATE TABLE IF NOT EXISTS absen_diniyyah_arsip LIKE absen_diniyyah`,
	}
	for _, q := range archiveTables {
		DB.Exec(q)
	}
}

// ── Proses arsip: pindahkan data > 1 tahun ──────────────

func runArchive() {
	cutoff := time.Now().AddDate(-1, 0, 0).Format("2006-01-02") // 1 tahun lalu
	fmt.Printf("  [archiver] Memindahkan data sebelum %s ke tabel arsip...\n", cutoff)

	// Daftar tabel yang diarsipkan (sumber → arsip)
	tables := []struct {
		src, dst, dateCol string
	}{
		{"absensi", "absensi_arsip", "tanggal"},
		{"absensi_sesi", "absensi_sesi_arsip", "tanggal"},
		{"absen_malam", "absen_malam_arsip", "tanggal"},
		{"absen_sekolah", "absen_sekolah_arsip", "tanggal"},
		{"absen_sekolah_sesi", "absen_sekolah_sesi_arsip", "tanggal"},
		{"absen_diniyyah", "absen_diniyyah_arsip", "tanggal"},
	}

	totalMoved := int64(0)

	for _, t := range tables {
		// 1. Copy ke tabel arsip (INSERT IGNORE agar tidak duplikat jika dijalankan ulang)
		insertQ := fmt.Sprintf(
			"INSERT IGNORE INTO %s SELECT * FROM %s WHERE %s < ?",
			t.dst, t.src, t.dateCol,
		)
		res, err := DB.Exec(insertQ, cutoff)
		if err != nil {
			fmt.Printf("  [archiver] ⚠ %s→%s insert error: %s\n", t.src, t.dst, err.Error())
			continue
		}
		moved, _ := res.RowsAffected()

		// 2. Hapus dari tabel utama (hanya yang sudah berhasil dicopy)
		if moved > 0 {
			deleteQ := fmt.Sprintf("DELETE FROM %s WHERE %s < ?", t.src, t.dateCol)
			DB.Exec(deleteQ, cutoff)
		}

		totalMoved += moved
		if moved > 0 {
			fmt.Printf("  [archiver] ✓ %s: %d baris diarsipkan\n", t.src, moved)
		}
	}

	if totalMoved == 0 {
		fmt.Println("  [archiver] Tidak ada data yang perlu diarsipkan")
	} else {
		fmt.Printf("  [archiver] Total %d baris dipindahkan ke arsip\n", totalMoved)
	}

	// 3. Bersihkan space yang sudah dibebaskan
	optimizeTables()

	// 4. Bersihkan binary log MariaDB
	purgeOldLogs()
}

// ── Optimize tables: kembalikan disk space ───────────────

func optimizeTables() {
	tables := []string{
		"absensi", "absensi_sesi",
		"absen_malam", "absen_sekolah", "absen_sekolah_sesi",
		"absen_diniyyah",
	}
	for _, t := range tables {
		_, err := DB.Exec("OPTIMIZE TABLE " + t)
		if err != nil {
			fmt.Printf("  [archiver] ⚠ optimize %s: %s\n", t, err.Error())
		}
	}
	fmt.Println("  [archiver] ✓ OPTIMIZE TABLE selesai — disk space dibebaskan")
}

// ── Bersihkan binary log MariaDB ────────────────────────

func purgeOldLogs() {
	// Hapus binary log lebih dari 3 hari (hemat disk, data tetap aman di tabel)
	_, err := DB.Exec("PURGE BINARY LOGS BEFORE DATE_SUB(NOW(), INTERVAL 3 DAY)")
	if err != nil {
		// Bisa gagal jika binary log tidak aktif — aman diabaikan
		fmt.Printf("  [archiver] binary log purge: %s (aman diabaikan)\n", err.Error())
	} else {
		fmt.Println("  [archiver] ✓ Binary log lama dibersihkan")
	}

	// Set agar binary log otomatis dihapus setelah 3 hari ke depan
	DB.Exec("SET GLOBAL expire_logs_days = 3")
	// Untuk MariaDB 10.6+ gunakan juga:
	DB.Exec("SET GLOBAL binlog_expire_logs_seconds = 259200") // 3 hari
}

// ── Scheduler: jalankan arsip otomatis tiap bulan ────────

func StartArchiveScheduler() {
	// Buat tabel arsip jika belum ada
	createArchiveTables()

	// Jalankan sekali saat startup
	go func() {
		// Tunggu 30 detik setelah server start agar tidak ganggu startup
		time.Sleep(30 * time.Second)
		fmt.Println("  [archiver] Menjalankan arsip awal...")
		runArchive()

		// Lalu jalankan tiap 30 hari
		ticker := time.NewTicker(30 * 24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Println("  [archiver] Menjalankan arsip bulanan...")
			runArchive()
		}
	}()
}
