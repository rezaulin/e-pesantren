package handlers

import (
	"fmt"
	"sort"
	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
)

func Dashboard(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	role := c.Locals("role").(string)
	today := helpers.TodayWIB()

	if role == "wali" {
		uid := c.Locals("user_id").(int)
		rows, _ := config.DB.Query(`SELECT s.id, s.nama, COALESCE(s.kelas_diniyyah,''), COALESCE(k.nama,'-'), s.status, COALESCE(pk.nominal, 0) 
			FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id LEFT JOIN pembayaran_kategori pk ON s.kategori_spp_id = pk.id 
			WHERE s.tenant_id = ? AND s.wali_user_id = ? ORDER BY s.nama`, tid, uid)
		defer rows.Close()

		bulanNow := today[:7]
		bulanStart := bulanNow + "-01"
		// Next month start for range query (avoid LIKE on DATE)
		bulanEnd := helpers.NextMonthStart(bulanNow)

		// Collect all anak IDs first
		type anakInfo struct {
			ID, SID                int
			Nama, Kelas, Kamar, St string
			KatNominal             int64
		}
		var anakArr []anakInfo
		var sids []interface{}
		for rows.Next() {
			var a anakInfo
			rows.Scan(&a.SID, &a.Nama, &a.Kelas, &a.Kamar, &a.St, &a.KatNominal)
			anakArr = append(anakArr, a)
			sids = append(sids, a.SID)
		}
		if len(anakArr) == 0 {
			return c.JSON(fiber.Map{"role": "wali", "anak": []fiber.Map{}, "bulan": bulanNow})
		}

		// Build IN clause placeholder
		inPh := "?"
		for i := 1; i < len(sids); i++ {
			inPh += ",?"
		}

		// Base args: tenant_id, bulanStart, bulanEnd, ...sids
		baseArgs := func() []interface{} {
			a := []interface{}{tid, bulanStart, bulanEnd}
			return append(a, sids...)
		}

		// ── BATCH 1: Absensi kegiatan (per kegiatan_nama) ──
		absensiMap := map[int]map[string][4]int{} // sid → kegiatan_nama → [H,I,S,A]
		r1, _ := config.DB.Query(`SELECT a.santri_id, COALESCE(k.kegiatan_nama,'Lainnya'),
			COALESCE(SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END),0)
			FROM absensi a LEFT JOIN kelompok k ON a.kelompok_id = k.id
			WHERE a.tenant_id = ? AND a.tanggal >= ? AND a.tanggal < ? AND a.santri_id IN (`+inPh+`)
			GROUP BY a.santri_id, k.kegiatan_nama`, baseArgs()...)
		if r1 != nil {
			defer r1.Close()
			for r1.Next() {
				var sid, h, i, s, a int
				var keg string
				r1.Scan(&sid, &keg, &h, &i, &s, &a)
				if absensiMap[sid] == nil {
					absensiMap[sid] = map[string][4]int{}
				}
				absensiMap[sid][keg] = [4]int{h, i, s, a}
			}
		}

		// ── BATCH 2: Absen malam ──
		malamMap := map[int][4]int{}
		r2, _ := config.DB.Query(`SELECT santri_id,
			COALESCE(SUM(CASE WHEN status='H' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='I' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='S' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='A' THEN 1 ELSE 0 END),0)
			FROM absen_malam WHERE tenant_id = ? AND tanggal >= ? AND tanggal < ? AND santri_id IN (`+inPh+`)
			GROUP BY santri_id`, baseArgs()...)
		if r2 != nil {
			defer r2.Close()
			for r2.Next() {
				var sid, h, i, s, a int
				r2.Scan(&sid, &h, &i, &s, &a)
				malamMap[sid] = [4]int{h, i, s, a}
			}
		}

		// ── BATCH 3: Absen sekolah ──
		sekolahMap := map[int][4]int{}
		r3, _ := config.DB.Query(`SELECT santri_id,
			COALESCE(SUM(CASE WHEN status='H' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='I' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='S' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='A' THEN 1 ELSE 0 END),0)
			FROM absen_sekolah WHERE tenant_id = ? AND tanggal >= ? AND tanggal < ? AND santri_id IN (`+inPh+`)
			GROUP BY santri_id`, baseArgs()...)
		if r3 != nil {
			defer r3.Close()
			for r3.Next() {
				var sid, h, i, s, a int
				r3.Scan(&sid, &h, &i, &s, &a)
				sekolahMap[sid] = [4]int{h, i, s, a}
			}
		}

		// ── BATCH 4: Absen diniyyah ──
		diniyyahMap := map[int][4]int{}
		r4, _ := config.DB.Query(`SELECT santri_id,
			COALESCE(SUM(CASE WHEN status='H' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='I' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='S' THEN 1 ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN status='A' THEN 1 ELSE 0 END),0)
			FROM absen_diniyyah WHERE tenant_id = ? AND tanggal >= ? AND tanggal < ? AND santri_id IN (`+inPh+`)
			GROUP BY santri_id`, baseArgs()...)
		if r4 != nil {
			defer r4.Close()
			for r4.Next() {
				var sid, h, i, s, a int
				r4.Scan(&sid, &h, &i, &s, &a)
				diniyyahMap[sid] = [4]int{h, i, s, a}
			}
		}

		// ── BATCH 5: Catatan guru (latest 5 per santri) ──
		catatanMap := map[int][]fiber.Map{}
		r5, _ := config.DB.Query(`SELECT c.santri_id, COALESCE(c.catatan,''), COALESCE(c.tanggal,''), COALESCE(u.nama,'')
			FROM catatan_guru c LEFT JOIN users u ON c.created_by = u.id
			WHERE c.tenant_id = ? AND c.santri_id IN (`+inPh+`)
			ORDER BY c.santri_id, c.tanggal DESC`,
			append([]interface{}{tid}, sids...)...)
		if r5 != nil {
			defer r5.Close()
			for r5.Next() {
				var sid int
				var ct, tgl, guru string
				r5.Scan(&sid, &ct, &tgl, &guru)
				if len(catatanMap[sid]) < 5 {
					catatanMap[sid] = append(catatanMap[sid], fiber.Map{"catatan": ct, "tanggal": tgl, "guru": guru})
				}
			}
		}

		// ── BATCH 6: Pelanggaran (latest 5 per santri + total poin) ──
		pelanggaranMap := map[int][]fiber.Map{}
		poinMap := map[int]int{}
		r6, _ := config.DB.Query(`SELECT santri_id, COALESCE(jenis,''), COALESCE(deskripsi,''), COALESCE(poin,0),
			COALESCE(tanggal,''), COALESCE(status_takzir,'belum'), COALESCE(jenis_takzir,''), COALESCE(denda,0)
			FROM pelanggaran WHERE tenant_id = ? AND santri_id IN (`+inPh+`)
			ORDER BY santri_id, tanggal DESC`,
			append([]interface{}{tid}, sids...)...)
		if r6 != nil {
			defer r6.Close()
			for r6.Next() {
				var sid, poin int
				var jenis, desk, tgl, stTakzir, jTakzir string
				var denda int64
				r6.Scan(&sid, &jenis, &desk, &poin, &tgl, &stTakzir, &jTakzir, &denda)
				poinMap[sid] += poin
				if len(pelanggaranMap[sid]) < 5 {
					pelanggaranMap[sid] = append(pelanggaranMap[sid], fiber.Map{
						"jenis": jenis, "deskripsi": desk, "poin": poin, "tanggal": tgl,
						"status_takzir": stTakzir, "jenis_takzir": jTakzir, "denda": denda,
					})
				}
			}
		}

		// ── BATCH 7: Pembayaran potongan ──
		potonganMap := map[int]int64{}
		r7, _ := config.DB.Query(`SELECT santri_id, COALESCE(SUM(nominal),0)
			FROM pembayaran_potongan WHERE tenant_id = ? AND aktif = 1 AND santri_id IN (`+inPh+`)
			GROUP BY santri_id`,
			append([]interface{}{tid}, sids...)...)
		if r7 != nil {
			defer r7.Close()
			for r7.Next() {
				var sid int
				var nom int64
				r7.Scan(&sid, &nom)
				potonganMap[sid] = nom
			}
		}

		// ── BATCH 8: Pembayaran bulan ini ──
		bayarMap := map[int]int64{}
		r8, _ := config.DB.Query(`SELECT santri_id, COALESCE(SUM(nominal),0)
			FROM pembayaran WHERE tenant_id = ? AND bulan = ? AND santri_id IN (`+inPh+`)
			GROUP BY santri_id`,
			append([]interface{}{tid, bulanNow}, sids...)...)
		if r8 != nil {
			defer r8.Close()
			for r8.Next() {
				var sid int
				var nom int64
				r8.Scan(&sid, &nom)
				bayarMap[sid] = nom
			}
		}

		// Get total tarif for this month (shared, used for display)
		// Note: actual tagihan is computed per-santri via HitungTagihan

		// ── Assemble results ──
		var anakList []fiber.Map
		for _, a := range anakArr {
			sid := a.SID
			// Build kegiatan map for this child
			kegMap := fiber.Map{}
			for keg, v := range absensiMap[sid] {
				kegMap[keg] = fiber.Map{"H": v[0], "I": v[1], "S": v[2], "A": v[3]}
			}
			am := malamMap[sid]
			as := sekolahMap[sid]
			ad := diniyyahMap[sid]

			cg := catatanMap[sid]
			if cg == nil {
				cg = []fiber.Map{}
			}
			pl := pelanggaranMap[sid]
			if pl == nil {
				pl = []fiber.Map{}
			}

			tagihan := HitungTagihan(tid, sid, bulanNow)
			totalBayar := bayarMap[sid]
			pbStatus := "BELUM BAYAR"
			kekurangan := tagihan - totalBayar
			if tagihan == 0 {
				pbStatus = "GRATIS"
				kekurangan = 0
			} else if totalBayar >= tagihan {
				pbStatus = "LUNAS"
				kekurangan = 0
			} else if totalBayar > 0 {
				pbStatus = "KURANG"
			}

			anakList = append(anakList, fiber.Map{
				"id": sid, "nama": a.Nama, "kelas_diniyyah": a.Kelas, "kamar_nama": a.Kamar, "status": a.St,
				"absensi": fiber.Map{
					"kegiatan": kegMap,
					"malam":    fiber.Map{"H": am[0], "I": am[1], "S": am[2], "A": am[3]},
					"sekolah":  fiber.Map{"H": as[0], "I": as[1], "S": as[2], "A": as[3]},
					"diniyyah": fiber.Map{"H": ad[0], "I": ad[1], "S": ad[2], "A": ad[3]},
				},
				"catatan_guru": cg,
				"pelanggaran":  pl,
				"total_poin":   poinMap[sid],
				"pembayaran": fiber.Map{
					"bulan": bulanNow, "tagihan": tagihan, "total_bayar": totalBayar,
					"kekurangan": kekurangan, "status": pbStatus,
				},
			})
		}
		if anakList == nil {
			anakList = []fiber.Map{}
		}
		return c.JSON(fiber.Map{"role": "wali", "anak": anakList, "bulan": bulanNow})
	}

	bulanNow := today[:7]

	var totalSantri, totalKamar, kamarTerisi, hadir, izinSakit, alfa, santriSudahAbsen, pelanggaranPending int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE tenant_id = ? AND status = 'aktif'", tid).Scan(&totalSantri)
	config.DB.QueryRow("SELECT COUNT(*) FROM kamar WHERE tenant_id = ?", tid).Scan(&totalKamar)
	config.DB.QueryRow("SELECT COUNT(DISTINCT kamar_id) FROM santri WHERE tenant_id = ? AND status = 'aktif' AND kamar_id IS NOT NULL", tid).Scan(&kamarTerisi)

	config.DB.QueryRow("SELECT COUNT(*) FROM absensi WHERE tenant_id = ? AND tanggal = ? AND status = 'H'", tid, today).Scan(&hadir)
	config.DB.QueryRow("SELECT COUNT(*) FROM absensi WHERE tenant_id = ? AND tanggal = ? AND status IN ('I','S')", tid, today).Scan(&izinSakit)
	config.DB.QueryRow("SELECT COUNT(*) FROM absensi WHERE tenant_id = ? AND tanggal = ? AND status = 'A'", tid, today).Scan(&alfa)
	config.DB.QueryRow("SELECT COUNT(DISTINCT santri_id) FROM absensi WHERE tenant_id = ? AND tanggal = ?", tid, today).Scan(&santriSudahAbsen)

	config.DB.QueryRow("SELECT COUNT(*) FROM pelanggaran WHERE tenant_id = ? AND status_takzir = 'belum'", tid).Scan(&pelanggaranPending)

	// Calculate total tagihan per santri using helper
	// Get all active santri IDs
	santriIDs := []int{}
	siRows, _ := config.DB.Query("SELECT id FROM santri WHERE tenant_id = ? AND status = 'aktif'", tid)
	if siRows != nil {
		for siRows.Next() {
			var sid int
			siRows.Scan(&sid)
			santriIDs = append(santriIDs, sid)
		}
		siRows.Close()
	}

	tagihanMap := HitungTagihanBatch(tid, santriIDs, bulanNow)

	var totalTagihan int64
	for _, tag := range tagihanMap {
		totalTagihan += tag
	}
	if totalTagihan < 0 {
		totalTagihan = 0
	}

	var terkumpul int64
	config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND bulan = ?", tid, bulanNow).Scan(&terkumpul)

	var lunasCount int
	for _, sid := range santriIDs {
		tag := tagihanMap[sid]
		var bayar int64
		config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, sid, bulanNow).Scan(&bayar)
		if tag == 0 || bayar >= tag {
			lunasCount++
		}
	}

	tertunggakCount := totalSantri - lunasCount
	if tertunggakCount < 0 {
		tertunggakCount = 0
	}

	belumAbsenCount := totalSantri - santriSudahAbsen
	if belumAbsenCount < 0 {
		belumAbsenCount = 0
	}

	// Aktivitas Terbaru (UNION of Pembayaran, Pelanggaran, Catatan Guru)
	var aktivitas []fiber.Map

	// Catatan Guru
	cgRows, _ := config.DB.Query(`SELECT c.catatan, c.created_at, u.nama, s.nama 
		FROM catatan_guru c JOIN users u ON c.created_by = u.id JOIN santri s ON c.santri_id = s.id 
		WHERE c.tenant_id = ? ORDER BY c.created_at DESC LIMIT 3`, tid)
	if cgRows != nil {
		defer cgRows.Close()
		for cgRows.Next() {
			var txt, tgl, uName, sName string
			cgRows.Scan(&txt, &tgl, &uName, &sName)
			aktivitas = append(aktivitas, fiber.Map{"type": "catatan", "teks": "Catatan untuk " + sName, "waktu": tgl, "oleh": uName, "sub": txt})
		}
	}

	// Pelanggaran (Using tanggal as waktu for sorting)
	plRows, _ := config.DB.Query(`SELECT p.jenis, p.tanggal, s.nama 
		FROM pelanggaran p JOIN santri s ON p.santri_id = s.id 
		WHERE p.tenant_id = ? ORDER BY p.tanggal DESC, p.id DESC LIMIT 3`, tid)
	if plRows != nil {
		defer plRows.Close()
		for plRows.Next() {
			var jenis, tgl, sName string
			plRows.Scan(&jenis, &tgl, &sName)
			aktivitas = append(aktivitas, fiber.Map{"type": "pelanggaran", "teks": "Pelanggaran: " + jenis, "waktu": tgl + " 00:00:00", "oleh": sName, "sub": "Menunggu review"})
		}
	}

	// Pembayaran
	pbRows, _ := config.DB.Query(`SELECT p.nominal, p.created_at, u.nama, s.nama 
		FROM pembayaran p JOIN users u ON p.created_by = u.id JOIN santri s ON p.santri_id = s.id 
		WHERE p.tenant_id = ? ORDER BY p.created_at DESC LIMIT 3`, tid)
	if pbRows != nil {
		defer pbRows.Close()
		for pbRows.Next() {
			var nom int64
			var tgl, uName, sName string
			pbRows.Scan(&nom, &tgl, &uName, &sName)
			aktivitas = append(aktivitas, fiber.Map{"type": "pembayaran", "teks": "Pembayaran dari " + sName, "waktu": tgl, "oleh": uName, "sub": fmt.Sprintf("Rp %d", nom)})
		}
	}

	// Sort aktivitas by waktu DESC
	sort.Slice(aktivitas, func(i, j int) bool {
		return aktivitas[i]["waktu"].(string) > aktivitas[j]["waktu"].(string)
	})
	if len(aktivitas) > 5 {
		aktivitas = aktivitas[:5]
	}

	return c.JSON(fiber.Map{
		"total_santri":         totalSantri,
		"total_kamar":          totalKamar,
		"kamar_terisi":         kamarTerisi,
		"hadir_hari_ini":       hadir,
		"izin_sakit":           izinSakit,
		"alfa":                 alfa,
		"belum_absen_hari_ini": belumAbsenCount,
		"pelanggaran_pending":  pelanggaranPending,
		"tagihan_bulan_ini":    totalTagihan,
		"terkumpul_bulan_ini":  terkumpul,
		"lunas_count":          lunasCount,
		"tertunggak_count":     tertunggakCount,
		"aktivitas_terbaru":    aktivitas,
	})
}
