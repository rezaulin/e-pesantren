package handlers

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func GetRaport(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("santri_id")
	bulan := c.Query("bulan")
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")

	// Build date filter clause helper
	buildDateFilter := func(col string) (string, []interface{}) {
		var clause string
		var extra []interface{}
		if tglMulai != "" && tglAkhir != "" {
			clause = " AND " + col + " BETWEEN ? AND ?"
			extra = append(extra, tglMulai, tglAkhir)
		} else if bulan != "" {
			clause = " AND " + col + " LIKE ?"
			extra = append(extra, bulan+"%")
		}
		return clause, extra
	}

	// Santri info
	var sID, kamarID, waliUID int
	var nama, alamat, kelasDiniyyah, kelasSekolah, status, kamarNama, namaWaliText string
	err := config.DB.QueryRow(`SELECT s.id, s.nama, COALESCE(s.alamat,''), COALESCE(s.kelas_diniyyah,''), 
		COALESCE((SELECT ks.nama FROM santri_kelas sk JOIN kelas_sekolah ks ON sk.kelas_id = ks.id WHERE sk.santri_id = s.id AND sk.tenant_id = s.tenant_id AND sk.status = 'active' LIMIT 1), '-'),
		s.status, s.kamar_id, COALESCE(k.nama,'-'), COALESCE(s.wali_user_id,0), COALESCE(s.nama_wali,'') 
		FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id WHERE s.id = ? AND s.tenant_id = ?`, sid, tid).
		Scan(&sID, &nama, &alamat, &kelasDiniyyah, &kelasSekolah, &status, &kamarID, &kamarNama, &waliUID, &namaWaliText)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}
	waliNama := "-"
	if waliUID > 0 {
		config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", waliUID).Scan(&waliNama)
	} else if namaWaliText != "" {
		waliNama = namaWaliText
	}

	// Absensi Kegiatan
	df, dfArgs := buildDateFilter("a.tanggal")
	q := "SELECT COALESCE(k.kegiatan_nama,'Lainnya'), a.status FROM absensi a LEFT JOIN kelompok k ON a.kelompok_id = k.id WHERE a.tenant_id = ? AND a.santri_id = ?" + df
	args := append([]interface{}{tid, sid}, dfArgs...)
	rows, _ := config.DB.Query(q, args...)
	defer rows.Close()
	type Rekap struct{ H, I, S, A int }
	rekap := map[string]*Rekap{}
	total := 0
	for rows.Next() {
		var keg, st string
		rows.Scan(&keg, &st)
		if rekap[keg] == nil {
			rekap[keg] = &Rekap{}
		}
		switch st {
		case "H":
			rekap[keg].H++
		case "I":
			rekap[keg].I++
		case "S":
			rekap[keg].S++
		case "A":
			rekap[keg].A++
		}
		total++
	}

	// Catatan Guru
	df2, df2Args := buildDateFilter("c.tanggal")
	cq := "SELECT c.id, COALESCE(c.catatan,''), COALESCE(c.tanggal,''), COALESCE(u.nama,'') FROM catatan_guru c LEFT JOIN users u ON c.created_by = u.id WHERE c.tenant_id = ? AND c.santri_id = ?" + df2 + " ORDER BY c.tanggal DESC"
	cargs := append([]interface{}{tid, sid}, df2Args...)
	crows, _ := config.DB.Query(cq, cargs...)
	defer crows.Close()
	var catatan []fiber.Map
	for crows.Next() {
		var id int
		var ct, tgl, gn string
		crows.Scan(&id, &ct, &tgl, &gn)
		catatan = append(catatan, fiber.Map{"id": id, "catatan": ct, "tanggal": tgl, "guru_nama": gn})
	}

	// Pelanggaran
	df3, df3Args := buildDateFilter("p.tanggal")
	pq := "SELECT p.id, COALESCE(p.jenis,''), COALESCE(p.deskripsi,''), COALESCE(p.poin,0), COALESCE(p.tanggal,''), COALESCE(p.status_takzir,'belum'), COALESCE(p.jenis_takzir,''), COALESCE(p.denda,0) FROM pelanggaran p WHERE p.tenant_id = ? AND p.santri_id = ?" + df3 + " ORDER BY p.tanggal DESC"
	pargs := append([]interface{}{tid, sid}, df3Args...)
	prows, _ := config.DB.Query(pq, pargs...)
	defer prows.Close()
	var pelanggaran []fiber.Map
	for prows.Next() {
		var id, poin int
		var denda int64
		var j, d, tgl, stTakzir, jTakzir string
		prows.Scan(&id, &j, &d, &poin, &tgl, &stTakzir, &jTakzir, &denda)
		pelanggaran = append(pelanggaran, fiber.Map{"id": id, "jenis": j, "deskripsi": d, "poin": poin, "tanggal": tgl, "status_takzir": stTakzir, "jenis_takzir": jTakzir, "denda": denda})
	}

	// Absen Malam
	df4, df4Args := buildDateFilter("tanggal")
	mqr := "SELECT status FROM absen_malam WHERE tenant_id = ? AND santri_id = ?" + df4
	margs := append([]interface{}{tid, sid}, df4Args...)
	mrows, _ := config.DB.Query(mqr, margs...)
	defer mrows.Close()
	rekapMalam := map[string]int{"H": 0, "I": 0, "S": 0, "A": 0}
	for mrows.Next() {
		var st string
		mrows.Scan(&st)
		rekapMalam[st]++
	}

	// Absen Sekolah -- per sesi (not per mata pelajaran)
	df5, df5Args := buildDateFilter("tanggal")
	sqr := "SELECT status FROM absen_sekolah WHERE tenant_id = ? AND santri_id = ?" + df5
	sargs := append([]interface{}{tid, sid}, df5Args...)
	srows, _ := config.DB.Query(sqr, sargs...)
	defer srows.Close()
	rekapSekolah := map[string]int{"H": 0, "I": 0, "S": 0, "A": 0}
	for srows.Next() {
		var st string
		srows.Scan(&st)
		rekapSekolah[st]++
	}

	// Count school sessions
	var jumlahSesiSekolah int
	sesiQ5 := "SELECT COUNT(*) FROM absen_sekolah_sesi WHERE tenant_id = ?"
	sesiArgs5 := []interface{}{tid}
	if tglMulai != "" && tglAkhir != "" {
		sesiQ5 += " AND tanggal BETWEEN ? AND ?"
		sesiArgs5 = append(sesiArgs5, tglMulai, tglAkhir)
	} else if bulan != "" {
		sesiQ5 += " AND tanggal LIKE ?"
		sesiArgs5 = append(sesiArgs5, bulan+"%")
	}
	config.DB.QueryRow(sesiQ5, sesiArgs5...).Scan(&jumlahSesiSekolah)

	// Absen Diniyyah
	df6, df6Args := buildDateFilter("ad.tanggal")
	dqr := `SELECT COALESCE(kd.nama,'Diniyyah'), ad.status FROM absen_diniyyah ad 
		LEFT JOIN kelas_diniyyah kd ON ad.kelas_diniyyah_id = kd.id 
		WHERE ad.tenant_id = ? AND ad.santri_id = ?` + df6
	dargs := append([]interface{}{tid, sid}, df6Args...)
	drows, _ := config.DB.Query(dqr, dargs...)
	defer drows.Close()
	rekapDiniyyah := map[string]*Rekap{}
	for drows.Next() {
		var kdNama, st string
		drows.Scan(&kdNama, &st)
		if rekapDiniyyah[kdNama] == nil {
			rekapDiniyyah[kdNama] = &Rekap{}
		}
		switch st {
		case "H":
			rekapDiniyyah[kdNama].H++
		case "I":
			rekapDiniyyah[kdNama].I++
		case "S":
			rekapDiniyyah[kdNama].S++
		case "A":
			rekapDiniyyah[kdNama].A++
		}
	}

	if catatan == nil {
		catatan = []fiber.Map{}
	}
	if pelanggaran == nil {
		pelanggaran = []fiber.Map{}
	}

	result := fiber.Map{
		"santri":         fiber.Map{"id": sID, "nama": nama, "alamat": alamat, "kelas_diniyyah": kelasDiniyyah, "kelas_sekolah": kelasSekolah, "status": status, "kamar_id": kamarID, "kamar_nama": kamarNama, "wali_nama": waliNama},
		"rekap":          rekap,
		"rekap_malam":    fiber.Map{"H": rekapMalam["H"], "I": rekapMalam["I"], "S": rekapMalam["S"], "A": rekapMalam["A"]},
		"rekap_sekolah":  fiber.Map{"H": rekapSekolah["H"], "I": rekapSekolah["I"], "S": rekapSekolah["S"], "A": rekapSekolah["A"], "jumlah_sesi": jumlahSesiSekolah},
		"rekap_diniyyah": rekapDiniyyah,
		"catatan_guru":   catatan,
		"pelanggaran":    pelanggaran,
		"total_records":  total,
	}

	// ── Nilai Pelajaran (if semester query param provided) ──
	semester := c.Query("semester")
	if semester != "" {
		npRows, _ := config.DB.Query(`SELECT mp.nama, np.nilai_harian, np.nilai_uts, np.nilai_uas, np.nilai_akhir, COALESCE(np.kkm, 75)
			FROM nilai_pelajaran np JOIN mata_pelajaran mp ON np.mata_pelajaran_id = mp.id
			WHERE np.tenant_id = ? AND np.santri_id = ? AND np.semester = ? ORDER BY mp.nama`, tid, sid, semester)
		var nilaiPelajaran []fiber.Map
		var totalNA float64
		var countNA int
		if npRows != nil {
			for npRows.Next() {
				var mpNama string
				var h, u, a, na, kkm float64
				npRows.Scan(&mpNama, &h, &u, &a, &na, &kkm)
				nilaiPelajaran = append(nilaiPelajaran, fiber.Map{"mata_pelajaran": mpNama, "nilai_harian": h, "nilai_uts": u, "nilai_uas": a, "nilai_akhir": na, "kkm": kkm})
				totalNA += na
				countNA++
			}
			npRows.Close()
		}
		if nilaiPelajaran == nil {
			nilaiPelajaran = []fiber.Map{}
		}
		rataRata := 0.0
		if countNA > 0 {
			rataRata = math.Round((totalNA/float64(countNA))*100) / 100
		}
		result["nilai_pelajaran"] = nilaiPelajaran
		result["rata_rata_pelajaran"] = rataRata

		// Peringkat kelas diniyyah
		var kdID int
		config.DB.QueryRow("SELECT kelas_diniyyah_id FROM santri_kelas_diniyyah WHERE santri_id = ? AND tenant_id = ? AND status = 'active' LIMIT 1", sid, tid).Scan(&kdID)
		if kdID > 0 {
			rkRows, _ := config.DB.Query(`SELECT np.santri_id, AVG(np.nilai_akhir) as avg_na
				FROM nilai_pelajaran np WHERE np.tenant_id = ? AND np.kelas_diniyyah_id = ? AND np.semester = ?
				GROUP BY np.santri_id ORDER BY avg_na DESC`, tid, kdID, semester)
			peringkat := 0
			totalSiswa := 0
			if rkRows != nil {
				for rkRows.Next() {
					var rsid int
					var avgNA float64
					rkRows.Scan(&rsid, &avgNA)
					totalSiswa++
					if fmt.Sprintf("%d", rsid) == fmt.Sprintf("%s", sid) {
						peringkat = totalSiswa
					}
				}
				rkRows.Close()
			}
			result["peringkat_kelas_diniyyah"] = fiber.Map{"peringkat": peringkat, "total_siswa": totalSiswa}
		}

		// ── Nilai Sekolah (also uses semester param) ──
		nsRows, _ := config.DB.Query(`SELECT mp.nama, ns.nilai_harian, ns.nilai_uts, ns.nilai_uas, ns.nilai_akhir, COALESCE(ns.kkm, 75)
			FROM nilai_sekolah ns JOIN mata_pelajaran_sekolah mp ON ns.mata_pelajaran_sekolah_id = mp.id
			WHERE ns.tenant_id = ? AND ns.santri_id = ? AND ns.semester = ? ORDER BY mp.nama`, tid, sid, semester)
		var nilaiSekolah []fiber.Map
		var totalNS float64
		var countNS int
		if nsRows != nil {
			for nsRows.Next() {
				var mpNama string
				var h, u, a, na, kkm float64
				nsRows.Scan(&mpNama, &h, &u, &a, &na, &kkm)
				nilaiSekolah = append(nilaiSekolah, fiber.Map{"mata_pelajaran": mpNama, "nilai_harian": h, "nilai_uts": u, "nilai_uas": a, "nilai_akhir": na, "kkm": kkm})
				totalNS += na
				countNS++
			}
			nsRows.Close()
		}
		if nilaiSekolah == nil {
			nilaiSekolah = []fiber.Map{}
		}
		rataRataSekolah := 0.0
		if countNS > 0 {
			rataRataSekolah = math.Round((totalNS/float64(countNS))*100) / 100
		}
		result["nilai_sekolah"] = nilaiSekolah
		result["rata_rata_sekolah"] = rataRataSekolah

		// Peringkat kelas sekolah
		var kelasSekolahID int
		config.DB.QueryRow("SELECT kelas_id FROM santri_kelas WHERE santri_id = ? AND tenant_id = ? AND status = 'active' LIMIT 1", sid, tid).Scan(&kelasSekolahID)
		if kelasSekolahID > 0 {
			rsRows, _ := config.DB.Query(`SELECT ns.santri_id, AVG(ns.nilai_akhir) as avg_na
				FROM nilai_sekolah ns WHERE ns.tenant_id = ? AND ns.kelas_id = ? AND ns.semester = ?
				GROUP BY ns.santri_id ORDER BY avg_na DESC`, tid, kelasSekolahID, semester)
			peringkatS := 0
			totalSiswaS := 0
			if rsRows != nil {
				for rsRows.Next() {
					var rsid int
					var avgNA float64
					rsRows.Scan(&rsid, &avgNA)
					totalSiswaS++
					if fmt.Sprintf("%d", rsid) == fmt.Sprintf("%s", sid) {
						peringkatS = totalSiswaS
					}
				}
				rsRows.Close()
			}
			result["peringkat_kelas_sekolah"] = fiber.Map{"peringkat": peringkatS, "total_siswa": totalSiswaS}
		}
	}

	// ── Nilai Kegiatan (if bulan_nilai query param provided) ──
	bulanNilai := c.Query("bulan_nilai")
	if bulanNilai != "" {
		nkRows, _ := config.DB.Query(`SELECT kg.nama, kl.nama, nk.nilai, COALESCE(nk.catatan,''), nk.kelompok_id,
			(SELECT COUNT(*) FROM absensi a WHERE a.tenant_id = nk.tenant_id AND a.santri_id = nk.santri_id AND a.kegiatan_id = nk.kegiatan_id AND a.tanggal LIKE CONCAT(nk.bulan, '-%') AND a.status = 'sakit') as sakit,
			(SELECT COUNT(*) FROM absensi a WHERE a.tenant_id = nk.tenant_id AND a.santri_id = nk.santri_id AND a.kegiatan_id = nk.kegiatan_id AND a.tanggal LIKE CONCAT(nk.bulan, '-%') AND a.status = 'izin') as izin,
			(SELECT COUNT(*) FROM absensi a WHERE a.tenant_id = nk.tenant_id AND a.santri_id = nk.santri_id AND a.kegiatan_id = nk.kegiatan_id AND a.tanggal LIKE CONCAT(nk.bulan, '-%') AND a.status = 'alpha') as alpha
			FROM nilai_kegiatan nk 
			JOIN kegiatan kg ON nk.kegiatan_id = kg.id 
			JOIN kelompok kl ON nk.kelompok_id = kl.id
			WHERE nk.tenant_id = ? AND nk.santri_id = ? AND nk.bulan = ? ORDER BY kg.nama`, tid, sid, bulanNilai)
		var nilaiKegiatan []fiber.Map
		var totalNK float64
		var countNK int
		var kelompokIDs []int
		if nkRows != nil {
			for nkRows.Next() {
				var kgNama, klNama, cat string
				var nilai float64
				var klID, sakit, izin, alpha int
				nkRows.Scan(&kgNama, &klNama, &nilai, &cat, &klID, &sakit, &izin, &alpha)
				nilaiKegiatan = append(nilaiKegiatan, fiber.Map{"kegiatan": kgNama, "kelompok": klNama, "nilai": nilai, "catatan": cat, "sakit": sakit, "izin": izin, "alpha": alpha})
				totalNK += nilai
				countNK++
				found := false
				for _, kid := range kelompokIDs {
					if kid == klID {
						found = true
						break
					}
				}
				if !found {
					kelompokIDs = append(kelompokIDs, klID)
				}
			}
			nkRows.Close()
		}
		if nilaiKegiatan == nil {
			nilaiKegiatan = []fiber.Map{}
		}
		rataRataKG := 0.0
		if countNK > 0 {
			rataRataKG = math.Round((totalNK/float64(countNK))*100) / 100
		}
		result["nilai_kegiatan"] = nilaiKegiatan
		result["rata_rata_kegiatan"] = rataRataKG

		// Peringkat per kelompok
		var peringkatList []fiber.Map
		for _, klID := range kelompokIDs {
			var klNama string
			config.DB.QueryRow("SELECT nama FROM kelompok WHERE id = ?", klID).Scan(&klNama)
			prRows, _ := config.DB.Query(`SELECT nk.santri_id, AVG(nk.nilai) as avg_n
				FROM nilai_kegiatan nk WHERE nk.tenant_id = ? AND nk.kelompok_id = ? AND nk.bulan = ?
				GROUP BY nk.santri_id ORDER BY avg_n DESC`, tid, klID, bulanNilai)
			peringkat := 0
			totalAnggota := 0
			if prRows != nil {
				for prRows.Next() {
					var rsid int
					var avgN float64
					prRows.Scan(&rsid, &avgN)
					totalAnggota++
					if fmt.Sprintf("%d", rsid) == fmt.Sprintf("%s", sid) {
						peringkat = totalAnggota
					}
				}
				prRows.Close()
			}
			peringkatList = append(peringkatList, fiber.Map{"kelompok": klNama, "peringkat": peringkat, "total_anggota": totalAnggota})
		}
		if peringkatList == nil {
			peringkatList = []fiber.Map{}
		}
		result["peringkat_kelompok"] = peringkatList
	}

	return c.JSON(result)
}

func RaportPDF(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("santri_id")
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")
	bulan := c.Query("bulan", helpers.MonthWIB())
	// Prefer range over bulan
	periodeLabel := bulan
	if tglMulai != "" && tglAkhir != "" {
		periodeLabel = tglMulai + " s/d " + tglAkhir
	}

	// Build date filter
	buildDF := func(col string) (string, []interface{}) {
		if tglMulai != "" && tglAkhir != "" {
			return " AND " + col + " BETWEEN ? AND ?", []interface{}{tglMulai, tglAkhir}
		}
		return " AND " + col + " LIKE ?", []interface{}{bulan + "%"}
	}

	var sNama, sAlamat, sKelas, kamarNama, namaWaliText2 string
	var waliUID int
	config.DB.QueryRow("SELECT s.nama, COALESCE(s.alamat,''), COALESCE(s.kelas_diniyyah,''), COALESCE(k.nama,'-'), COALESCE(s.wali_user_id,0), COALESCE(s.nama_wali,'') FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id WHERE s.id = ? AND s.tenant_id = ?", sid, tid).
		Scan(&sNama, &sAlamat, &sKelas, &kamarNama, &waliUID, &namaWaliText2)
	if sNama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}
	waliNama := "-"
	if waliUID > 0 {
		config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", waliUID).Scan(&waliNama)
	} else if namaWaliText2 != "" {
		waliNama = namaWaliText2
	}
	var lembaga string
	config.DB.QueryRow("SELECT COALESCE(app_name,'Pesantren') FROM settings WHERE tenant_id = ?", tid).Scan(&lembaga)
	if lembaga == "" {
		lembaga = "Pesantren"
	}

	// Rekap Kegiatan
	dfK, dfKArgs := buildDF("a.tanggal")
	rows, _ := config.DB.Query("SELECT COALESCE(k.kegiatan_nama,'Lainnya'), a.status FROM absensi a LEFT JOIN kelompok k ON a.kelompok_id = k.id WHERE a.tenant_id = ? AND a.santri_id = ?"+dfK, append([]interface{}{tid, sid}, dfKArgs...)...)
	type R struct{ H, I, S, A int }
	rekap := map[string]*R{}
	for rows.Next() {
		var keg, st string
		rows.Scan(&keg, &st)
		if rekap[keg] == nil {
			rekap[keg] = &R{}
		}
		switch st {
		case "H":
			rekap[keg].H++
		case "I":
			rekap[keg].I++
		case "S":
			rekap[keg].S++
		case "A":
			rekap[keg].A++
		}
	}
	rows.Close()

	// Rekap Sekolah per sesi (not per mata pelajaran)
	dfS, dfSArgs := buildDF("tanggal")
	sRows, _ := config.DB.Query("SELECT status FROM absen_sekolah WHERE tenant_id = ? AND santri_id = ?"+dfS, append([]interface{}{tid, sid}, dfSArgs...)...)
	sH, sI, sS, sA := 0, 0, 0, 0
	for sRows.Next() {
		var st string
		sRows.Scan(&st)
		switch st {
		case "H":
			sH++
		case "I":
			sI++
		case "S":
			sS++
		case "A":
			sA++
		}
	}
	sRows.Close()

	// C. Rekap Absen Malam
	dfM, dfMArgs := buildDF("tanggal")
	mRows, _ := config.DB.Query("SELECT status FROM absen_malam WHERE tenant_id = ? AND santri_id = ?"+dfM, append([]interface{}{tid, sid}, dfMArgs...)...)
	rmH, rmI, rmS, rmA := 0, 0, 0, 0
	for mRows.Next() {
		var st string
		mRows.Scan(&st)
		switch st {
		case "H":
			rmH++
		case "I":
			rmI++
		case "S":
			rmS++
		case "A":
			rmA++
		}
	}
	mRows.Close()

	// Rekap Diniyyah
	dfD, dfDArgs := buildDF("ad.tanggal")
	dRows, _ := config.DB.Query(`SELECT COALESCE(kd.nama,'Diniyyah'), ad.status FROM absen_diniyyah ad 
		LEFT JOIN kelas_diniyyah kd ON ad.kelas_diniyyah_id = kd.id 
		WHERE ad.tenant_id = ? AND ad.santri_id = ?`+dfD, append([]interface{}{tid, sid}, dfDArgs...)...)
	type RD struct{ H, I, S, A int }
	rekapDin := map[string]*RD{}
	for dRows.Next() {
		var kn, st string
		dRows.Scan(&kn, &st)
		if rekapDin[kn] == nil {
			rekapDin[kn] = &RD{}
		}
		switch st {
		case "H":
			rekapDin[kn].H++
		case "I":
			rekapDin[kn].I++
		case "S":
			rekapDin[kn].S++
		case "A":
			rekapDin[kn].A++
		}
	}
	dRows.Close()

	dfC, dfCArgs := buildDF("tanggal")
	catRows, _ := config.DB.Query("SELECT COALESCE(catatan,''), COALESCE(tanggal,'') FROM catatan_guru WHERE tenant_id = ? AND santri_id = ?"+dfC, append([]interface{}{tid, sid}, dfCArgs...)...)
	var catatan []string
	for catRows.Next() {
		var ct, tgl string
		catRows.Scan(&ct, &tgl)
		catatan = append(catatan, fmt.Sprintf("[%s] %s", tgl, ct))
	}
	catRows.Close()

	dfP, dfPArgs := buildDF("tanggal")
	pelRows, _ := config.DB.Query("SELECT COALESCE(jenis,''), COALESCE(deskripsi,''), COALESCE(poin,0), COALESCE(tanggal,'') FROM pelanggaran WHERE tenant_id = ? AND santri_id = ?"+dfP, append([]interface{}{tid, sid}, dfPArgs...)...)
	var pelanggaran []string
	for pelRows.Next() {
		var j, d, tgl string
		var p int
		pelRows.Scan(&j, &d, &p, &tgl)
		pelanggaran = append(pelanggaran, fmt.Sprintf("[%s] %s - %s (%d poin)", tgl, j, d, p))
	}
	pelRows.Close()

	// Build PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 8, strings.ToUpper(lembaga), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 5, "LAPORAN BULANAN PERKEMBANGAN SANTRI", "", 1, "C", false, 0, "")
	pdf.Ln(3)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	// Identity
	pdf.SetFont("Helvetica", "", 10)
	y := pdf.GetY()
	pdf.Text(10, y, "Nama Santri")
	pdf.Text(45, y, ": "+sNama)
	pdf.Text(120, y, "Kelas")
	pdf.Text(150, y, ": "+sKelas)
	pdf.Text(10, y+5, "Asrama")
	pdf.Text(45, y+5, ": "+kamarNama)
	pdf.Text(120, y+5, "Periode")
	pdf.Text(150, y+5, ": "+periodeLabel)
	pdf.Text(10, y+10, "Wali")
	pdf.Text(45, y+10, ": "+waliNama)
	pdf.SetY(y + 18)

	// A. Rekap
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "A. Rekap Absensi", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 9)
	w := []float64{60, 22, 22, 22, 22, 22}
	hdr := []string{"Kegiatan", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
	for i, h := range hdr {
		pdf.CellFormat(w[i], 6, h, "1", 0, "C", false, 0, "")
		_ = h
	}
	pdf.Ln(-1)
	pdf.SetFont("Helvetica", "", 9)
	tH, tI, tS, tA := 0, 0, 0, 0
	for keg, r := range rekap {
		t := r.H + r.I + r.S + r.A
		pdf.CellFormat(w[0], 6, keg, "1", 0, "", false, 0, "")
		pdf.CellFormat(w[1], 6, fmt.Sprintf("%d", r.H), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[2], 6, fmt.Sprintf("%d", r.I), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[3], 6, fmt.Sprintf("%d", r.S), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[4], 6, fmt.Sprintf("%d", r.A), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[5], 6, fmt.Sprintf("%d", t), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
		tH += r.H
		tI += r.I
		tS += r.S
		tA += r.A
	}
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(w[0], 6, "TOTAL", "1", 0, "", false, 0, "")
	pdf.CellFormat(w[1], 6, fmt.Sprintf("%d", tH), "1", 0, "C", false, 0, "")
	pdf.CellFormat(w[2], 6, fmt.Sprintf("%d", tI), "1", 0, "C", false, 0, "")
	pdf.CellFormat(w[3], 6, fmt.Sprintf("%d", tS), "1", 0, "C", false, 0, "")
	pdf.CellFormat(w[4], 6, fmt.Sprintf("%d", tA), "1", 0, "C", false, 0, "")
	pdf.CellFormat(w[5], 6, fmt.Sprintf("%d", tH+tI+tS+tA), "1", 0, "C", false, 0, "")
	pdf.Ln(5)

	// B. Rekap Absen Sekolah (per sesi)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "B. Rekap Absen Sekolah", "", 1, "", false, 0, "")
	sekolahTotal := sH + sI + sS + sA
	if sekolahTotal > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		wS := []float64{60, 22, 22, 22, 22, 22}
		hdrS := []string{"Keterangan", "Hadir", "Izin", "Sakit", "Alpa", "Total Sesi"}
		for i, h := range hdrS {
			pdf.CellFormat(wS[i], 6, h, "1", 0, "C", false, 0, "")
			_ = h
		}
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(wS[0], 6, "Sekolah", "1", 0, "", false, 0, "")
		pdf.CellFormat(wS[1], 6, fmt.Sprintf("%d", sH), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wS[2], 6, fmt.Sprintf("%d", sI), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wS[3], 6, fmt.Sprintf("%d", sS), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wS[4], 6, fmt.Sprintf("%d", sA), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wS[5], 6, fmt.Sprintf("%d", sekolahTotal), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	} else {
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 5, "Belum ada data absen sekolah.", "", 1, "", false, 0, "")
	}
	pdf.Ln(3)

	// C. Rekap Absen Malam
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "C. Rekap Absen Malam", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 9)
	wM := []float64{40, 22, 22, 22, 22, 22}
	hdrM := []string{"Keterangan", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
	for i, h := range hdrM {
		pdf.CellFormat(wM[i], 6, h, "1", 0, "C", false, 0, "")
		_ = h
	}
	pdf.Ln(-1)
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(wM[0], 6, "Malam", "1", 0, "", false, 0, "")
	pdf.CellFormat(wM[1], 6, fmt.Sprintf("%d", rmH), "1", 0, "C", false, 0, "")
	pdf.CellFormat(wM[2], 6, fmt.Sprintf("%d", rmI), "1", 0, "C", false, 0, "")
	pdf.CellFormat(wM[3], 6, fmt.Sprintf("%d", rmS), "1", 0, "C", false, 0, "")
	pdf.CellFormat(wM[4], 6, fmt.Sprintf("%d", rmA), "1", 0, "C", false, 0, "")
	pdf.CellFormat(wM[5], 6, fmt.Sprintf("%d", rmH+rmI+rmS+rmA), "1", 0, "C", false, 0, "")
	pdf.Ln(5)

	// C2. Rekap Absen Diniyyah
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "C2. Rekap Absen Madrasah Diniyyah", "", 1, "", false, 0, "")
	if len(rekapDin) > 0 {
		pdf.SetFont("Helvetica", "B", 9)
		wD := []float64{60, 22, 22, 22, 22, 22}
		hdrD := []string{"Kelas", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
		for i, h := range hdrD {
			pdf.CellFormat(wD[i], 6, h, "1", 0, "C", false, 0, "")
			_ = h
		}
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 9)
		for kelas, rd := range rekapDin {
			dt := rd.H + rd.I + rd.S + rd.A
			pdf.CellFormat(wD[0], 6, kelas, "1", 0, "", false, 0, "")
			pdf.CellFormat(wD[1], 6, fmt.Sprintf("%d", rd.H), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wD[2], 6, fmt.Sprintf("%d", rd.I), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wD[3], 6, fmt.Sprintf("%d", rd.S), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wD[4], 6, fmt.Sprintf("%d", rd.A), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wD[5], 6, fmt.Sprintf("%d", dt), "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
		}
	} else {
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 5, "Belum ada data absen diniyyah.", "", 1, "", false, 0, "")
	}
	pdf.Ln(3)

	// D. Pelanggaran
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "D. Catatan Kedisiplinan", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	if len(pelanggaran) > 0 {
		for i, p := range pelanggaran {
			pdf.CellFormat(0, 5, fmt.Sprintf("%d. %s", i+1, p), "", 1, "", false, 0, "")
		}
	} else {
		pdf.CellFormat(0, 5, "Tidak ada catatan pelanggaran.", "", 1, "", false, 0, "")
	}
	pdf.Ln(5)

	// E. Catatan
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "E. Laporan Perkembangan", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	if len(catatan) > 0 {
		for i, ct := range catatan {
			pdf.CellFormat(0, 5, fmt.Sprintf("%d. %s", i+1, ct), "", 1, "", false, 0, "")
		}
	} else {
		pdf.CellFormat(0, 5, "Belum ada catatan perkembangan.", "", 1, "", false, 0, "")
	}
	pdf.Ln(3)

	// F. Nilai Kegiatan
	bulanNilai := c.Query("bulan_nilai")
	if bulanNilai != "" {
		nkRows, _ := config.DB.Query(`SELECT kg.nama, kl.nama, nk.nilai, COALESCE(nk.catatan,'')
			FROM nilai_kegiatan nk
			JOIN kegiatan kg ON nk.kegiatan_id = kg.id
			JOIN kelompok kl ON nk.kelompok_id = kl.id
			WHERE nk.tenant_id = ? AND nk.santri_id = ? AND nk.bulan = ? ORDER BY kg.nama`, tid, sid, bulanNilai)
		type NK struct {
			kegiatan, kelompok, catatan string
			nilai                       float64
		}
		var nkList []NK
		var totalNK float64
		if nkRows != nil {
			for nkRows.Next() {
				var n NK
				nkRows.Scan(&n.kegiatan, &n.kelompok, &n.nilai, &n.catatan)
				nkList = append(nkList, n)
				totalNK += n.nilai
			}
			nkRows.Close()
		}
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, "F. Nilai Kegiatan ("+bulanNilai+")", "", 1, "", false, 0, "")
		if len(nkList) > 0 {
			pdf.SetFont("Helvetica", "B", 9)
			wN := []float64{45, 40, 25, 60}
			hdrN := []string{"Kegiatan", "Kelompok", "Nilai", "Catatan"}
			for i, h := range hdrN {
				pdf.CellFormat(wN[i], 6, h, "1", 0, "C", false, 0, "")
				_ = h
			}
			pdf.Ln(-1)
			pdf.SetFont("Helvetica", "", 9)
			for _, n := range nkList {
				pdf.CellFormat(wN[0], 6, n.kegiatan, "1", 0, "", false, 0, "")
				pdf.CellFormat(wN[1], 6, n.kelompok, "1", 0, "", false, 0, "")
				pdf.CellFormat(wN[2], 6, fmt.Sprintf("%.1f", n.nilai), "1", 0, "C", false, 0, "")
				pdf.CellFormat(wN[3], 6, n.catatan, "1", 0, "", false, 0, "")
				pdf.Ln(-1)
			}
			rata := 0.0
			if len(nkList) > 0 {
				rata = math.Round((totalNK/float64(len(nkList)))*100) / 100
			}
			pdf.SetFont("Helvetica", "B", 9)
			pdf.CellFormat(wN[0]+wN[1], 6, "Rata-rata", "1", 0, "", false, 0, "")
			pdf.CellFormat(wN[2], 6, fmt.Sprintf("%.1f", rata), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wN[3], 6, "", "1", 0, "", false, 0, "")
			pdf.Ln(-1)
		} else {
			pdf.SetFont("Helvetica", "", 9)
			pdf.CellFormat(0, 5, "Belum ada data nilai kegiatan.", "", 1, "", false, 0, "")
		}
	}

	var buf bytes.Buffer
	pdf.Output(&buf)
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=raport-%s.pdf", strings.ReplaceAll(sNama, " ", "_")))
	return c.Send(buf.Bytes())
}

func RaportExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("santri_id")
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")
	bulan := c.Query("bulan", helpers.MonthWIB())
	periode := bulan
	if tglMulai != "" && tglAkhir != "" {
		periode = tglMulai + " s/d " + tglAkhir
	}
	buildDF2 := func(col string) (string, []interface{}) {
		if tglMulai != "" && tglAkhir != "" {
			return " AND " + col + " BETWEEN ? AND ?", []interface{}{tglMulai, tglAkhir}
		}
		return " AND " + col + " LIKE ?", []interface{}{bulan + "%"}
	}

	var sNama, sKelas, kamarNama, namaWaliText3 string
	var waliUID int
	config.DB.QueryRow("SELECT s.nama, COALESCE(s.kelas_diniyyah,''), COALESCE(k.nama,'-'), COALESCE(s.wali_user_id,0), COALESCE(s.nama_wali,'') FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id WHERE s.id = ? AND s.tenant_id = ?", sid, tid).
		Scan(&sNama, &sKelas, &kamarNama, &waliUID, &namaWaliText3)
	if sNama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}
	waliNama := "-"
	if waliUID > 0 {
		config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", waliUID).Scan(&waliNama)
	} else if namaWaliText3 != "" {
		waliNama = namaWaliText3
	}
	var lembaga string
	config.DB.QueryRow("SELECT COALESCE(app_name,'Pesantren') FROM settings WHERE tenant_id = ?", tid).Scan(&lembaga)

	dfXK, dfXKArgs := buildDF2("a.tanggal")
	rows, _ := config.DB.Query("SELECT COALESCE(k.kegiatan_nama,'Lainnya'), a.status FROM absensi a LEFT JOIN kelompok k ON a.kelompok_id = k.id WHERE a.tenant_id = ? AND a.santri_id = ?"+dfXK, append([]interface{}{tid, sid}, dfXKArgs...)...)
	type R struct{ H, I, S, A int }
	rekap := map[string]*R{}
	for rows.Next() {
		var keg, st string
		rows.Scan(&keg, &st)
		if rekap[keg] == nil {
			rekap[keg] = &R{}
		}
		switch st {
		case "H":
			rekap[keg].H++
		case "I":
			rekap[keg].I++
		case "S":
			rekap[keg].S++
		case "A":
			rekap[keg].A++
		}
	}
	rows.Close()

	f := excelize.NewFile()
	sheet := "Raport"
	f.SetSheetName("Sheet1", sheet)
	f.MergeCell(sheet, "A1", "E1")
	f.SetCellValue(sheet, "A1", strings.ToUpper(lembaga))
	r := 3
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Nama")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), ": "+sNama)
	r++
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Kamar")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), ": "+kamarNama)
	r++
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Wali")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), ": "+waliNama)
	r++
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Periode")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), ": "+periode)
	r += 2

	hdrs := []string{"Kegiatan", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
	for i, h := range hdrs {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h)
	}
	r++
	tH, tI, tS, tA := 0, 0, 0, 0
	for keg, rv := range rekap {
		t := rv.H + rv.I + rv.S + rv.A
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), keg)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), rv.H)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), rv.I)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), rv.S)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), rv.A)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), t)
		tH += rv.H
		tI += rv.I
		tS += rv.S
		tA += rv.A
		r++
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "TOTAL")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), tH)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", r), tI)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", r), tS)
	f.SetCellValue(sheet, fmt.Sprintf("E%d", r), tA)
	f.SetCellValue(sheet, fmt.Sprintf("F%d", r), tH+tI+tS+tA)
	r += 2

	// Nilai Kegiatan
	bulanNilai := c.Query("bulan_nilai")
	if bulanNilai != "" {
		nkRows, _ := config.DB.Query(`SELECT kg.nama, kl.nama, nk.nilai, COALESCE(nk.catatan,'')
			FROM nilai_kegiatan nk
			JOIN kegiatan kg ON nk.kegiatan_id = kg.id
			JOIN kelompok kl ON nk.kelompok_id = kl.id
			WHERE nk.tenant_id = ? AND nk.santri_id = ? AND nk.bulan = ? ORDER BY kg.nama`, tid, sid, bulanNilai)
		if nkRows != nil {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "NILAI KEGIATAN ("+bulanNilai+")")
			r++
			nkHdrs := []string{"Kegiatan", "Kelompok", "Nilai", "Catatan"}
			for i, h := range nkHdrs {
				col := string(rune('A' + i))
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h)
			}
			r++
			var totalNK float64
			var countNK int
			for nkRows.Next() {
				var kgNama, klNama, cat string
				var nilai float64
				nkRows.Scan(&kgNama, &klNama, &nilai, &cat)
				f.SetCellValue(sheet, fmt.Sprintf("A%d", r), kgNama)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", r), klNama)
				f.SetCellValue(sheet, fmt.Sprintf("C%d", r), nilai)
				f.SetCellValue(sheet, fmt.Sprintf("D%d", r), cat)
				totalNK += nilai
				countNK++
				r++
			}
			nkRows.Close()
			if countNK > 0 {
				rata := math.Round((totalNK/float64(countNK))*100) / 100
				f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Rata-rata")
				f.SetCellValue(sheet, fmt.Sprintf("C%d", r), rata)
			}
		}
	}

	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", strings.ReplaceAll(sNama, " ", "_")))
	return c.Send(buf.Bytes())
}
