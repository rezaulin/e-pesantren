package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

type nilaiRow struct {
	MapelNama string
	Harian    float64
	UTS       float64
	UAS       float64
	Akhir     float64
	Detail    string
}

type kegiatanRow struct {
	Kegiatan string
	Kelompok string
	Nilai    float64
	Catatan  string
}

func getNilaiPenilaian(tid int, sid string, semester string) ([]nilaiRow, float64, []nilaiRow, float64) {
	// Nilai Diniyyah
	var dinList []nilaiRow
	var dinTotal float64
	var dinCount int
	npRows, _ := config.DB.Query(`SELECT mp.nama, np.nilai_harian, np.nilai_uts, np.nilai_uas, np.nilai_akhir, COALESCE(np.nilai_detail,'{}')
		FROM nilai_pelajaran np JOIN mata_pelajaran mp ON np.mata_pelajaran_id = mp.id
		WHERE np.tenant_id = ? AND np.santri_id = ? AND np.semester = ? ORDER BY mp.nama`, tid, sid, semester)
	if npRows != nil {
		for npRows.Next() {
			var r nilaiRow
			npRows.Scan(&r.MapelNama, &r.Harian, &r.UTS, &r.UAS, &r.Akhir, &r.Detail)
			dinList = append(dinList, r)
			dinTotal += r.Akhir
			dinCount++
		}
		npRows.Close()
	}
	dinAvg := 0.0
	if dinCount > 0 {
		dinAvg = math.Round((dinTotal/float64(dinCount))*100) / 100
	}

	// Nilai Sekolah
	var sekList []nilaiRow
	var sekTotal float64
	var sekCount int
	nsRows, _ := config.DB.Query(`SELECT mp.nama, ns.nilai_harian, ns.nilai_uts, ns.nilai_uas, ns.nilai_akhir, COALESCE(ns.nilai_detail,'{}')
		FROM nilai_sekolah ns JOIN mata_pelajaran_sekolah mp ON ns.mata_pelajaran_sekolah_id = mp.id
		WHERE ns.tenant_id = ? AND ns.santri_id = ? AND ns.semester = ? ORDER BY mp.nama`, tid, sid, semester)
	if nsRows != nil {
		for nsRows.Next() {
			var r nilaiRow
			nsRows.Scan(&r.MapelNama, &r.Harian, &r.UTS, &r.UAS, &r.Akhir, &r.Detail)
			sekList = append(sekList, r)
			sekTotal += r.Akhir
			sekCount++
		}
		nsRows.Close()
	}
	sekAvg := 0.0
	if sekCount > 0 {
		sekAvg = math.Round((sekTotal/float64(sekCount))*100) / 100
	}

	return dinList, dinAvg, sekList, sekAvg
}

func getNilaiKegiatan(tid int, sid string, bulan string) ([]kegiatanRow, float64) {
	var list []kegiatanRow
	var total float64
	var count int
	rows, _ := config.DB.Query(`SELECT kg.nama, kl.nama, nk.nilai, COALESCE(nk.catatan,'')
		FROM nilai_kegiatan nk
		JOIN kegiatan kg ON nk.kegiatan_id = kg.id
		JOIN kelompok kl ON nk.kelompok_id = kl.id
		WHERE nk.tenant_id = ? AND nk.santri_id = ? AND nk.bulan = ? ORDER BY kg.nama`, tid, sid, bulan)
	if rows != nil {
		for rows.Next() {
			var r kegiatanRow
			rows.Scan(&r.Kegiatan, &r.Kelompok, &r.Nilai, &r.Catatan)
			list = append(list, r)
			total += r.Nilai
			count++
		}
		rows.Close()
	}
	avg := 0.0
	if count > 0 {
		avg = math.Round((total/float64(count))*100) / 100
	}
	return list, avg
}

// buildPenilaianExcel builds Excel with kategori filter: semua, sekolah, diniyyah, kegiatan
func buildPenilaianExcel(tid int, sid string, semester string, santriNama string, settings map[string]string, kategori string, bulan string) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	sheet := "Raport Penilaian"
	f.SetSheetName("Sheet1", sheet)

	lembaga := settings["app_name"]
	if lembaga == "" {
		lembaga = "Pesantren"
	}

	// Styles
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleBorder, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	styleBorderCenter, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Header
	f.MergeCell(sheet, "A1", "F1")
	f.SetCellValue(sheet, "A1", lembaga)
	f.SetCellStyle(sheet, "A1", "F1", styleTitle)

	titleSuffix := "Semester " + semester
	if kategori == "sekolah" {
		titleSuffix = "SEKOLAH — Semester " + semester
	} else if kategori == "diniyyah" {
		titleSuffix = "MADRASAH DINIYYAH — Semester " + semester
	} else if kategori == "kegiatan" {
		titleSuffix = "KEGIATAN — " + bulan
	}
	f.MergeCell(sheet, "A2", "F2")
	f.SetCellValue(sheet, "A2", "RAPORT PENILAIAN "+titleSuffix)
	f.SetCellStyle(sheet, "A2", "F2", styleTitle)

	f.MergeCell(sheet, "A3", "F3")
	f.SetCellValue(sheet, "A3", "Nama: "+santriNama)

	row := 5

	type Komp struct {
		Nama  string `json:"nama"`
		Bobot int    `json:"bobot"`
	}

	writeNilaiSection := func(title string, list []nilaiRow, avg float64, compsStr string) {
		var comps []Komp
		if err := json.Unmarshal([]byte(compsStr), &comps); err != nil || len(comps) == 0 {
			comps = []Komp{{Nama: "Harian", Bobot: 30}, {Nama: "UTS", Bobot: 30}, {Nama: "UAS", Bobot: 40}}
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), title)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleTitle)
		row++

		// Header Row
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Mata Pelajaran")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleHeader)
		
		colIdx := 2
		for _, c := range comps {
			colName, _ := excelize.ColumnNumberToName(colIdx)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, row), c.Nama)
			f.SetCellStyle(sheet, fmt.Sprintf("%s%d", colName, row), fmt.Sprintf("%s%d", colName, row), styleHeader)
			colIdx++
		}
		colNameAkhir, _ := excelize.ColumnNumberToName(colIdx)
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", colNameAkhir, row), "Nilai Akhir")
		f.SetCellStyle(sheet, fmt.Sprintf("%s%d", colNameAkhir, row), fmt.Sprintf("%s%d", colNameAkhir, row), styleHeader)
		row++

		// Data Rows
		for _, n := range list {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), n.MapelNama)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleBorder)

			var detail map[string]interface{}
			json.Unmarshal([]byte(n.Detail), &detail)

			cIdx := 2
			for _, c := range comps {
				colName, _ := excelize.ColumnNumberToName(cIdx)
				val := detail[c.Nama]
				if val == nil && c.Nama == "Harian" { val = n.Harian }
				if val == nil && c.Nama == "UTS" { val = n.UTS }
				if val == nil && c.Nama == "UAS" { val = n.UAS }
				if val == nil { val = "-" }
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, row), val)
				f.SetCellStyle(sheet, fmt.Sprintf("%s%d", colName, row), fmt.Sprintf("%s%d", colName, row), styleBorderCenter)
				cIdx++
			}
			colNameAkhir, _ := excelize.ColumnNumberToName(cIdx)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", colNameAkhir, row), n.Akhir)
			f.SetCellStyle(sheet, fmt.Sprintf("%s%d", colNameAkhir, row), fmt.Sprintf("%s%d", colNameAkhir, row), styleBorderCenter)
			row++
		}

		// Rata-rata
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Rata-rata")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleHeader)
		cIdx := 2
		for range comps {
			colName, _ := excelize.ColumnNumberToName(cIdx)
			f.SetCellStyle(sheet, fmt.Sprintf("%s%d", colName, row), fmt.Sprintf("%s%d", colName, row), styleHeader)
			cIdx++
		}
		colNameAkhir, _ = excelize.ColumnNumberToName(cIdx)
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", colNameAkhir, row), avg)
		f.SetCellStyle(sheet, fmt.Sprintf("%s%d", colNameAkhir, row), fmt.Sprintf("%s%d", colNameAkhir, row), styleHeader)
		row += 2
	}

	// Diniyyah
	if kategori == "semua" || kategori == "diniyyah" || kategori == "" {
		dinList, dinAvg, _, _ := getNilaiPenilaian(tid, sid, semester)
		writeNilaiSection("NILAI MADRASAH DINIYYAH", dinList, dinAvg, settings["komponen_nilai_diniyyah"])
	}

	// Sekolah
	if kategori == "semua" || kategori == "sekolah" || kategori == "" {
		_, _, sekList, sekAvg := getNilaiPenilaian(tid, sid, semester)
		writeNilaiSection("NILAI SEKOLAH", sekList, sekAvg, settings["komponen_nilai_sekolah"])
	}

	// Kegiatan
	if (kategori == "semua" || kategori == "kegiatan") && bulan != "" {
		kegList, kegAvg := getNilaiKegiatan(tid, sid, bulan)
		if len(kegList) > 0 {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "NILAI KEGIATAN ("+bulan+")")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleTitle)
			row++
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Kegiatan")
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Kelompok")
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Nilai")
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "Catatan")
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), styleHeader)
			row++
			for _, n := range kegList {
				f.SetCellValue(sheet, fmt.Sprintf("A%d", row), n.Kegiatan)
				f.SetCellValue(sheet, fmt.Sprintf("B%d", row), n.Kelompok)
				f.SetCellValue(sheet, fmt.Sprintf("C%d", row), n.Nilai)
				f.SetCellValue(sheet, fmt.Sprintf("D%d", row), n.Catatan)
				f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleBorder)
				f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), styleBorderCenter)
				f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleBorder)
				row++
			}
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Rata-rata")
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), kegAvg)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), styleHeader)
			f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleHeader)
			f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleHeader)
			row += 2
		}
	}

	// Column widths
	f.SetColWidth(sheet, "A", "A", 25)
	f.SetColWidth(sheet, "B", "Z", 15)

	buf, err := f.WriteToBuffer()
	return buf, err
}

func RaportPenilaianExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("santri_id")
	semester := c.Query("semester")
	if semester == "" {
		return c.Status(400).JSON(fiber.Map{"message": "semester wajib"})
	}
	var nama string
	config.DB.QueryRow("SELECT nama FROM santri WHERE id = ? AND tenant_id = ?", sid, tid).Scan(&nama)
	if nama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}
	settings := getSettings(tid)
	kategori := c.Query("kategori", "semua")
	bulan := c.Query("bulan")
	buf, err := buildPenilaianExcel(tid, sid, semester, nama, settings, kategori, bulan)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=RaportPenilaian_%s.xlsx", nama))
	return c.Send(buf.Bytes())
}

func RaportPenilaianPDF(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("santri_id")
	semester := c.Query("semester")
	if semester == "" {
		return c.Status(400).JSON(fiber.Map{"message": "semester wajib"})
	}
	var nama string
	config.DB.QueryRow("SELECT nama FROM santri WHERE id = ? AND tenant_id = ?", sid, tid).Scan(&nama)
	if nama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}
	settings := getSettings(tid)
	buf, err := buildPenilaianExcel(tid, sid, semester, nama, settings, "semua", "")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=RaportPenilaian_%s.xlsx", nama))
	return c.Send(buf.Bytes())
}

func RaportPenilaianAllZip(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	semester := c.Query("semester")
	kategori := c.Query("kategori", "semua")
	bulan := c.Query("bulan")
	kelasID := c.Query("kelas_id")

	if semester == "" && kategori != "kegiatan" {
		return c.Status(400).JSON(fiber.Map{"message": "semester wajib"})
	}
	if kategori == "kegiatan" && bulan == "" {
		return c.Status(400).JSON(fiber.Map{"message": "bulan wajib untuk kategori kegiatan"})
	}

	settings := getSettings(tid)

	// Build santri query based on kategori + kelas_id filter
	var santriQuery string
	var santriArgs []interface{}

	if kelasID != "" && kategori == "sekolah" {
		// Filter santri yang ada di kelas_sekolah tertentu
		santriQuery = `SELECT s.id, s.nama FROM santri s
			INNER JOIN santri_kelas sk ON s.id = sk.santri_id AND sk.tenant_id = s.tenant_id AND sk.status = 'active'
			WHERE s.tenant_id = ? AND s.status = 'aktif' AND sk.kelas_id = ? ORDER BY s.nama`
		santriArgs = []interface{}{tid, kelasID}
	} else if kelasID != "" && kategori == "diniyyah" {
		// Filter santri yang ada di kelas_diniyyah tertentu
		santriQuery = `SELECT s.id, s.nama FROM santri s
			INNER JOIN santri_kelas_diniyyah skd ON s.id = skd.santri_id AND skd.tenant_id = s.tenant_id AND skd.status = 'active'
			WHERE s.tenant_id = ? AND s.status = 'aktif' AND skd.kelas_diniyyah_id = ? ORDER BY s.nama`
		santriArgs = []interface{}{tid, kelasID}
	} else {
		// Semua santri aktif
		santriQuery = "SELECT id, nama FROM santri WHERE tenant_id = ? AND status = 'aktif' ORDER BY nama"
		santriArgs = []interface{}{tid}
	}

	rows, err := config.DB.Query(santriQuery, santriArgs...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	for rows.Next() {
		var id int
		var nama string
		rows.Scan(&id, &nama)
		sid := fmt.Sprintf("%d", id)
		buf, err := buildPenilaianExcel(tid, sid, semester, nama, settings, kategori, bulan)
		if err != nil {
			continue
		}
		// Use category prefix in filename
		prefix := "Raport"
		switch kategori {
		case "sekolah":
			prefix = "RaportSekolah"
		case "diniyyah":
			prefix = "RaportDiniyyah"
		case "kegiatan":
			prefix = "RaportKegiatan"
		default:
			prefix = "RaportPenilaian"
		}
		fw, _ := zw.Create(fmt.Sprintf("%s_%s.xlsx", prefix, nama))
		fw.Write(buf.Bytes())
	}
	zw.Close()

	// Get kelas name for ZIP filename
	kelasLabel := ""
	if kelasID != "" {
		if kategori == "sekolah" {
			config.DB.QueryRow("SELECT nama FROM kelas_sekolah WHERE id = ? AND tenant_id = ?", kelasID, tid).Scan(&kelasLabel)
		} else if kategori == "diniyyah" {
			config.DB.QueryRow("SELECT nama FROM kelas_diniyyah WHERE id = ? AND tenant_id = ?", kelasID, tid).Scan(&kelasLabel)
		}
	}

	// Filename for the ZIP
	zipName := fmt.Sprintf("Raport_%s_%s.zip", kategori, semester)
	if kelasLabel != "" {
		zipName = fmt.Sprintf("Raport_%s_%s_%s.zip", kategori, kelasLabel, semester)
	}
	if kategori == "kegiatan" {
		zipName = fmt.Sprintf("RaportKegiatan_%s.zip", bulan)
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipName))
	return c.Send(zipBuf.Bytes())
}

func getSettings(tid int) map[string]string {
	s := map[string]string{}
	var appName, alamatLembaga, kepalaNama, namaKota, logoBase64, kompSekolah, kompDiniyyah string
	var activeSemester, activeMonth string
	var ttdSekolahNama, ttdSekolahImg, ttdDiniyyahNama, ttdDiniyyahImg, ttdKegiatanNama, ttdKegiatanImg string
	var tahunAjaranAktif, semesterAktif string
	var namaLembagaSekolah, headerSekolahLine1, headerSekolahLine2, headerSekolahAlamat, logoSekolah string
	var namaLembagaDiniyyah, headerDiniyyahAlamat, logoDiniyyah string
	var namaLembagaKegiatan, headerKegiatanAlamat, logoKegiatan string
	var semesterSekolah, semesterDiniyyah string
	err := config.DB.QueryRow(`SELECT COALESCE(app_name,'Pesantren'), COALESCE(alamat_lembaga,''), 
		COALESCE(kepala_nama,''), COALESCE(nama_kota,''), 
		COALESCE(logo_base64,''), COALESCE(komponen_nilai_sekolah,''), COALESCE(komponen_nilai_diniyyah,''),
		COALESCE(active_semester,''), COALESCE(active_month,''),
		COALESCE(ttd_sekolah_nama,''), COALESCE(ttd_sekolah_img,''),
		COALESCE(ttd_diniyyah_nama,''), COALESCE(ttd_diniyyah_img,''),
		COALESCE(ttd_kegiatan_nama,''), COALESCE(ttd_kegiatan_img,''),
		COALESCE(nama_lembaga_sekolah,''), COALESCE(header_sekolah_line1,'PEMERINTAH DAERAH'), COALESCE(header_sekolah_line2,'DINAS PENDIDIKAN'), COALESCE(header_sekolah_alamat,''), COALESCE(logo_sekolah,''),
		COALESCE(nama_lembaga_diniyyah,''), COALESCE(header_diniyyah_alamat,''), COALESCE(logo_diniyyah,''),
		COALESCE(nama_lembaga_kegiatan,''), COALESCE(header_kegiatan_alamat,''), COALESCE(logo_kegiatan,''),
		COALESCE(tahun_ajaran_aktif,''), COALESCE(semester_aktif,''),
		COALESCE(semester_sekolah,''), COALESCE(semester_diniyyah,'')
		FROM settings WHERE tenant_id = ? LIMIT 1`, tid).
		Scan(&appName, &alamatLembaga, &kepalaNama, &namaKota, &logoBase64, &kompSekolah, &kompDiniyyah, &activeSemester, &activeMonth,
			&ttdSekolahNama, &ttdSekolahImg, &ttdDiniyyahNama, &ttdDiniyyahImg, &ttdKegiatanNama, &ttdKegiatanImg,
			&namaLembagaSekolah, &headerSekolahLine1, &headerSekolahLine2, &headerSekolahAlamat, &logoSekolah,
			&namaLembagaDiniyyah, &headerDiniyyahAlamat, &logoDiniyyah,
			&namaLembagaKegiatan, &headerKegiatanAlamat, &logoKegiatan,
			&tahunAjaranAktif, &semesterAktif,
			&semesterSekolah, &semesterDiniyyah)
	if err == nil {
		s["app_name"] = appName
		s["alamat_lembaga"] = alamatLembaga
		s["kepala_nama"] = kepalaNama
		s["nama_kota"] = namaKota
		s["logo_base64"] = logoBase64
		s["komponen_nilai_sekolah"] = kompSekolah
		s["komponen_nilai_diniyyah"] = kompDiniyyah
		s["active_semester"] = activeSemester
		s["tahun_ajaran_aktif"] = tahunAjaranAktif
		s["semester_aktif"] = semesterAktif
		s["active_month"] = activeMonth
		s["ttd_sekolah_nama"] = ttdSekolahNama
		s["ttd_sekolah_img"] = ttdSekolahImg
		s["ttd_diniyyah_nama"] = ttdDiniyyahNama
		s["ttd_diniyyah_img"] = ttdDiniyyahImg
		s["ttd_kegiatan_nama"] = ttdKegiatanNama
		s["ttd_kegiatan_img"] = ttdKegiatanImg
		s["nama_lembaga_sekolah"] = namaLembagaSekolah
		s["header_sekolah_line1"] = headerSekolahLine1
		s["header_sekolah_line2"] = headerSekolahLine2
		s["header_sekolah_alamat"] = headerSekolahAlamat
		s["logo_sekolah"] = logoSekolah
		s["nama_lembaga_diniyyah"] = namaLembagaDiniyyah
		s["header_diniyyah_alamat"] = headerDiniyyahAlamat
		s["logo_diniyyah"] = logoDiniyyah
		s["nama_lembaga_kegiatan"] = namaLembagaKegiatan
		s["header_kegiatan_alamat"] = headerKegiatanAlamat
		s["logo_kegiatan"] = logoKegiatan
	}
	return s
}
