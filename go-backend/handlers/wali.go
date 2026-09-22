package handlers

import (
	"fmt"
	"math"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// WaliGetAnak — returns all santri linked to the logged-in wali user
func WaliGetAnak(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	rows, err := config.DB.Query(`SELECT s.id, s.nama, COALESCE(s.alamat,''), COALESCE(s.kelas_diniyyah,''), 
		s.status, COALESCE(k.nama,'-'), COALESCE(s.no_hp,'')
		FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id 
		WHERE s.tenant_id = ? AND s.wali_user_id = ? ORDER BY s.nama`, tid, uid)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama, alamat, kelas, status, kamar, noHP string
		rows.Scan(&id, &nama, &alamat, &kelas, &status, &kamar, &noHP)
		list = append(list, fiber.Map{
			"id": id, "nama": nama, "alamat": alamat, "kelas_diniyyah": kelas,
			"status": status, "kamar_nama": kamar, "no_hp": noHP,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

// WaliGetRaport — get raport for a specific child (validates wali ownership)
func WaliGetRaport(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	sid := c.Params("santri_id")

	// Validate: santri belongs to this wali (skip for admin/superadmin)
	if role == "wali" {
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}

	// Delegate to existing GetRaport — set params and call
	return GetRaport(c)
}

// WaliGetPembayaran — get payment data for a specific child
func WaliGetPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	sid := c.Params("santri_id")
	bulan := c.Query("bulan")

	// Validate ownership
	if role == "wali" {
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}

	// Get santri info
	var sNama, kamarNama string
	config.DB.QueryRow("SELECT s.nama, COALESCE(k.nama,'-') FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id WHERE s.id = ? AND s.tenant_id = ?", sid, tid).Scan(&sNama, &kamarNama)
	if sNama == "" { return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"}) }

	// Get tarif using helper (period-aware)
	bulanQ := bulan
	if bulanQ == "" {
		// If no bulan specified, use current month for tagihan calculation
		bulanQ = helpers.TodayWIB()[:7]
	}
	sidInt, _ := strconv.Atoi(sid)
	tagihan := HitungTagihan(tid, sidInt, bulanQ)

	// Get riwayat pembayaran
	q := `SELECT id, bulan, nominal, metode, keterangan, created_at FROM (
            SELECT p.id, CAST(p.bulan AS CHAR) as bulan, p.nominal, CAST(p.metode AS CHAR) as metode, CAST(COALESCE(p.keterangan,'') AS CHAR) as keterangan, CAST(p.created_at AS CHAR) as created_at 
            FROM pembayaran p WHERE p.tenant_id = ? AND p.santri_id = ?
            UNION ALL
            SELECT pi.id, CAST(CONCAT('Insidental: ', COALESCE(ti.nama, '')) AS CHAR) as bulan, pi.nominal_dibayar as nominal, CAST(pi.metode AS CHAR) as metode, CAST('Tagihan Insidental' AS CHAR) as keterangan, CAST(pi.tanggal AS CHAR) as created_at 
            FROM pembayaran_insidental pi
            LEFT JOIN tagihan_insidental_santri tis ON pi.tagihan_santri_id = tis.id
            LEFT JOIN tagihan_insidental ti ON tis.tagihan_id = ti.id
            WHERE pi.tenant_id = ? AND pi.santri_id = ?
        ) combined
        WHERE 1=1`
	args := []interface{}{tid, sid, tid, sid}
	if bulan != "" {
		q += " AND bulan = ?"
		args = append(args, bulan)
	}
	q += " ORDER BY created_at DESC LIMIT 100"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()

	var riwayat []fiber.Map
	var totalBayar int64
	for rows.Next() {
		var id int; var bln, metode, ket, ca string; var nominal int64
		rows.Scan(&id, &bln, &nominal, &metode, &ket, &ca)
		riwayat = append(riwayat, fiber.Map{"id": id, "bulan": bln, "nominal": nominal, "metode": metode, "keterangan": ket, "created_at": ca})
		if bulan != "" { totalBayar += nominal }
	}
	if riwayat == nil { riwayat = []fiber.Map{} }

	// Calculate current month status if bulan specified
	if bulan != "" {
		config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, sid, bulan).Scan(&totalBayar)
	}

	status := "BELUM BAYAR"
	kekurangan := tagihan - totalBayar
	if tagihan == 0 { status = "GRATIS"; kekurangan = 0 } else if totalBayar >= tagihan { status = "LUNAS"; kekurangan = 0 } else if totalBayar > 0 { status = "KURANG" }

	return c.JSON(fiber.Map{
		"santri":     fiber.Map{"nama": sNama, "kamar": kamarNama},
		"tarif":      tagihan + hitungTotalPotongan(tid, sidInt),
		"potongan":   hitungTotalPotongan(tid, sidInt),
		"tagihan":    tagihan,
		"total_bayar": totalBayar,
		"kekurangan": kekurangan,
		"status":     status,
		"riwayat":    riwayat,
	})
}

// WaliGetRaportLengkap — extended raport with nilai + peringkat for wali
func WaliGetRaportLengkap(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	sid := c.Params("santri_id")
	semester := c.Query("semester")
	bulan := c.Query("bulan")

	// Validate ownership
	if role == "wali" {
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}

	// Santri info
	var sNama, sKelas, sKamar string
	config.DB.QueryRow(`SELECT s.nama, COALESCE(s.kelas_diniyyah,''), COALESCE(k.nama,'-') 
		FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id 
		WHERE s.id = ? AND s.tenant_id = ?`, sid, tid).Scan(&sNama, &sKelas, &sKamar)
	if sNama == "" { return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"}) }

	result := fiber.Map{
		"santri": fiber.Map{"nama": sNama, "kelas_diniyyah": sKelas, "kamar": sKamar},
	}

	// Nilai Pelajaran (if semester provided)
	if semester != "" {
		npRows, _ := config.DB.Query(`SELECT mp.nama, np.nilai_harian, np.nilai_uts, np.nilai_uas, np.nilai_akhir
			FROM nilai_pelajaran np JOIN mata_pelajaran mp ON np.mata_pelajaran_id = mp.id
			WHERE np.tenant_id = ? AND np.santri_id = ? AND np.semester = ? ORDER BY mp.nama`, tid, sid, semester)
		var nilaiPelajaran []fiber.Map
		var totalNA float64; var countNA int
		for npRows.Next() {
			var mpNama string; var h, u, a, na float64
			npRows.Scan(&mpNama, &h, &u, &a, &na)
			nilaiPelajaran = append(nilaiPelajaran, fiber.Map{"mata_pelajaran": mpNama, "nilai_harian": h, "nilai_uts": u, "nilai_uas": a, "nilai_akhir": na})
			totalNA += na; countNA++
		}
		npRows.Close()
		if nilaiPelajaran == nil { nilaiPelajaran = []fiber.Map{} }
		rataRata := 0.0
		if countNA > 0 { rataRata = math.Round((totalNA/float64(countNA))*100) / 100 }
		result["nilai_pelajaran"] = nilaiPelajaran
		result["rata_rata_pelajaran"] = rataRata

		// Peringkat kelas diniyyah — find which kelas diniyyah this santri belongs to
		var kdID int
		config.DB.QueryRow("SELECT kelas_diniyyah_id FROM santri_kelas_diniyyah WHERE santri_id = ? AND tenant_id = ? AND status = 'active' LIMIT 1", sid, tid).Scan(&kdID)
		if kdID > 0 {
			rankRows, _ := config.DB.Query(`SELECT np.santri_id, AVG(np.nilai_akhir) as avg_na
				FROM nilai_pelajaran np WHERE np.tenant_id = ? AND np.kelas_diniyyah_id = ? AND np.semester = ?
				GROUP BY np.santri_id ORDER BY avg_na DESC`, tid, kdID, semester)
			peringkat := 0; totalSiswa := 0
			for rankRows.Next() {
				var rsid int; var avgNA float64
				rankRows.Scan(&rsid, &avgNA)
				totalSiswa++
				if fmt.Sprintf("%d", rsid) == sid { peringkat = totalSiswa }
			}
			rankRows.Close()
			result["peringkat_kelas_diniyyah"] = fiber.Map{"peringkat": peringkat, "total_siswa": totalSiswa}
		}

		// ── Nilai Sekolah (also uses semester) ──
		nsRows, _ := config.DB.Query(`SELECT mp.nama, ns.nilai_harian, ns.nilai_uts, ns.nilai_uas, ns.nilai_akhir
			FROM nilai_sekolah ns JOIN mata_pelajaran_sekolah mp ON ns.mata_pelajaran_sekolah_id = mp.id
			WHERE ns.tenant_id = ? AND ns.santri_id = ? AND ns.semester = ? ORDER BY mp.nama`, tid, sid, semester)
		var nilaiSekolah []fiber.Map
		var totalNS float64; var countNS int
		if nsRows != nil {
			for nsRows.Next() {
				var mpNama string; var h, u, a, na float64
				nsRows.Scan(&mpNama, &h, &u, &a, &na)
				nilaiSekolah = append(nilaiSekolah, fiber.Map{"mata_pelajaran": mpNama, "nilai_harian": h, "nilai_uts": u, "nilai_uas": a, "nilai_akhir": na})
				totalNS += na; countNS++
			}
			nsRows.Close()
		}
		if nilaiSekolah == nil { nilaiSekolah = []fiber.Map{} }
		rataRataSekolah := 0.0
		if countNS > 0 { rataRataSekolah = math.Round((totalNS/float64(countNS))*100) / 100 }
		result["nilai_sekolah"] = nilaiSekolah
		result["rata_rata_sekolah"] = rataRataSekolah

		// Peringkat kelas sekolah
		var kelasSekolahID int
		config.DB.QueryRow("SELECT kelas_id FROM santri_kelas WHERE santri_id = ? AND tenant_id = ? AND status = 'active' LIMIT 1", sid, tid).Scan(&kelasSekolahID)
		if kelasSekolahID > 0 {
			rsRows, _ := config.DB.Query(`SELECT ns.santri_id, AVG(ns.nilai_akhir) as avg_na
				FROM nilai_sekolah ns WHERE ns.tenant_id = ? AND ns.kelas_id = ? AND ns.semester = ?
				GROUP BY ns.santri_id ORDER BY avg_na DESC`, tid, kelasSekolahID, semester)
			peringkatS := 0; totalSiswaS := 0
			if rsRows != nil {
				for rsRows.Next() {
					var rsid int; var avgNA float64
					rsRows.Scan(&rsid, &avgNA)
					totalSiswaS++
					if fmt.Sprintf("%d", rsid) == sid { peringkatS = totalSiswaS }
				}
				rsRows.Close()
			}
			result["peringkat_kelas_sekolah"] = fiber.Map{"peringkat": peringkatS, "total_siswa": totalSiswaS}
		}
	}

	// Nilai Kegiatan (if bulan provided)
	if bulan != "" {
		nkRows, _ := config.DB.Query(`SELECT kg.nama, kl.nama, nk.nilai, COALESCE(nk.catatan,''), nk.kelompok_id
			FROM nilai_kegiatan nk 
			JOIN kegiatan kg ON nk.kegiatan_id = kg.id 
			JOIN kelompok kl ON nk.kelompok_id = kl.id
			WHERE nk.tenant_id = ? AND nk.santri_id = ? AND nk.bulan = ? ORDER BY kg.nama`, tid, sid, bulan)
		var nilaiKegiatan []fiber.Map
		var totalNK float64; var countNK int
		var kelompokIDs []int
		for nkRows.Next() {
			var kgNama, klNama, catatan string; var nilai float64; var klID int
			nkRows.Scan(&kgNama, &klNama, &nilai, &catatan, &klID)
			nilaiKegiatan = append(nilaiKegiatan, fiber.Map{"kegiatan": kgNama, "kelompok": klNama, "nilai": nilai, "catatan": catatan})
			totalNK += nilai; countNK++
			// Track unique kelompok IDs
			found := false
			for _, kid := range kelompokIDs { if kid == klID { found = true; break } }
			if !found { kelompokIDs = append(kelompokIDs, klID) }
		}
		nkRows.Close()
		if nilaiKegiatan == nil { nilaiKegiatan = []fiber.Map{} }
		rataRataKG := 0.0
		if countNK > 0 { rataRataKG = math.Round((totalNK/float64(countNK))*100) / 100 }
		result["nilai_kegiatan"] = nilaiKegiatan
		result["rata_rata_kegiatan"] = rataRataKG

		// Peringkat per kelompok
		var peringkatList []fiber.Map
		for _, klID := range kelompokIDs {
			var klNama string
			config.DB.QueryRow("SELECT nama FROM kelompok WHERE id = ?", klID).Scan(&klNama)
			rkRows, _ := config.DB.Query(`SELECT nk.santri_id, AVG(nk.nilai) as avg_n
				FROM nilai_kegiatan nk WHERE nk.tenant_id = ? AND nk.kelompok_id = ? AND nk.bulan = ?
				GROUP BY nk.santri_id ORDER BY avg_n DESC`, tid, klID, bulan)
			peringkat := 0; totalAnggota := 0
			for rkRows.Next() {
				var rsid int; var avgN float64
				rkRows.Scan(&rsid, &avgN)
				totalAnggota++
				if fmt.Sprintf("%d", rsid) == sid { peringkat = totalAnggota }
			}
			rkRows.Close()
			peringkatList = append(peringkatList, fiber.Map{"kelompok": klNama, "peringkat": peringkat, "total_anggota": totalAnggota})
		}
		if peringkatList == nil { peringkatList = []fiber.Map{} }
		result["peringkat_kelompok"] = peringkatList
	}

	return c.JSON(result)
}

// ═══════════════════════════════════════════════════════════
// WALI: LAPORAN TAHFIDZ QUR'AN
// ═══════════════════════════════════════════════════════════

// WaliGetTahfidz — laporan tahfidz lengkap untuk wali santri
func WaliGetTahfidz(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	sid := c.Params("santri_id")

	// Validate ownership
	if role == "wali" {
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}

	// Santri info
	var sNama, sKamar string
	config.DB.QueryRow(`SELECT s.nama, COALESCE(k.nama,'-') FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id 
		WHERE s.id = ? AND s.tenant_id = ?`, sid, tid).Scan(&sNama, &sKamar)
	if sNama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}

	// Halaqoh info — cari halaqoh santri ini
	var halaqohID int
	var halaqohNama, musyrif, targetJuz, targetSurah, targetDeadline string
	var targetAyat int
	config.DB.QueryRow(`SELECT h.id, h.nama, COALESCE(h.musyrif,''), COALESCE(h.target_juz,''), 
		COALESCE(h.target_surah,''), h.target_ayat, COALESCE(h.target_deadline,'')
		FROM halaqoh_members hm JOIN halaqoh h ON hm.halaqoh_id = h.id AND hm.tenant_id = h.tenant_id
		WHERE hm.santri_id = ? AND hm.tenant_id = ? AND hm.status = 'active' LIMIT 1`, sid, tid).Scan(
		&halaqohID, &halaqohNama, &musyrif, &targetJuz, &targetSurah, &targetAyat, &targetDeadline)

	// Summary: absensi agregat
	var totalH, totalI, totalS, totalA, totalSesi, capaianAyat int
	var avgNilai float64
	config.DB.QueryRow(`SELECT 
		COALESCE(SUM(CASE WHEN status='H' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status='I' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status='S' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status='A' THEN 1 ELSE 0 END), 0),
		COUNT(*),
		COALESCE(AVG(CASE WHEN status='H' AND nilai_rata > 0 THEN nilai_rata END), 0),
		COALESCE(SUM(CASE WHEN jenis_setoran='ziyadah' AND status='H' THEN GREATEST(ayat_sampai - ayat_dari + 1, 0) ELSE 0 END), 0)
		FROM tahfidz_nilai WHERE tenant_id = ? AND santri_id = ?`, tid, sid).Scan(
		&totalH, &totalI, &totalS, &totalA, &totalSesi, &avgNilai, &capaianAyat)

	avgNilai = math.Round(avgNilai*100) / 100
	predikat := ""
	if avgNilai > 0 {
		predikat = getPredikat(avgNilai)
	}

	progressPct := 0.0
	if targetAyat > 0 {
		progressPct = math.Round(float64(capaianAyat)/float64(targetAyat)*10000) / 100
	}

	// Riwayat setoran (50 terbaru)
	rows, err := config.DB.Query(`SELECT tn.tanggal, tn.status, tn.jenis_setoran, COALESCE(tn.surah,''), 
		tn.ayat_dari, tn.ayat_sampai, tn.juz, tn.nilai_tajwid, tn.nilai_kelancaran, tn.nilai_makhorijul,
		tn.nilai_rata, COALESCE(tn.predikat,''), COALESCE(tn.catatan,'')
		FROM tahfidz_nilai tn WHERE tn.tenant_id = ? AND tn.santri_id = ? 
		ORDER BY tn.tanggal DESC, tn.id DESC LIMIT 50`, tid, sid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var riwayat []fiber.Map
	for rows.Next() {
		var ayatDari, ayatSampai, juz, tajwid, kelancaran, makhorijul int
		var tanggal, status, jenis, surah, pred, catatan string
		var nilaiRata float64
		rows.Scan(&tanggal, &status, &jenis, &surah, &ayatDari, &ayatSampai, &juz,
			&tajwid, &kelancaran, &makhorijul, &nilaiRata, &pred, &catatan)
		riwayat = append(riwayat, fiber.Map{
			"tanggal": tanggal, "status": status, "jenis_setoran": jenis,
			"surah": surah, "ayat_dari": ayatDari, "ayat_sampai": ayatSampai, "juz": juz,
			"nilai_tajwid": tajwid, "nilai_kelancaran": kelancaran, "nilai_makhorijul": makhorijul,
			"nilai_rata": nilaiRata, "predikat": pred, "catatan": catatan,
		})
	}
	if riwayat == nil {
		riwayat = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"santri":   fiber.Map{"nama": sNama, "kamar": sKamar},
		"halaqoh":  fiber.Map{"id": halaqohID, "nama": halaqohNama, "musyrif": musyrif, "target_juz": targetJuz, "target_surah": targetSurah, "target_ayat": targetAyat, "target_deadline": targetDeadline},
		"summary":  fiber.Map{"H": totalH, "I": totalI, "S": totalS, "A": totalA, "total_sesi": totalSesi, "avg_nilai": avgNilai, "predikat": predikat, "capaian_ayat": capaianAyat, "progress_pct": progressPct},
		"riwayat":  riwayat,
	})
}
