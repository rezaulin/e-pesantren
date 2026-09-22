package handlers

import (
	"fmt"
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// -- CATATAN KEUANGAN (Uang Masuk / Keluar) --

func GetCatatanKeuangan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")
	tipe := c.Query("tipe") // masuk / keluar / empty=all

	q := `SELECT id, tipe, nominal, COALESCE(keterangan,''), tanggal, created_at 
		  FROM catatan_keuangan WHERE tenant_id = ?`
	args := []interface{}{tid}
	if tglMulai != "" {
		q += " AND tanggal >= ?"
		args = append(args, tglMulai)
	}
	if tglAkhir != "" {
		q += " AND tanggal <= ?"
		args = append(args, tglAkhir)
	}
	if tipe == "masuk" || tipe == "keluar" {
		q += " AND tipe = ?"
		args = append(args, tipe)
	}
	q += " ORDER BY tanggal DESC, id DESC LIMIT 500"

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var tipe, ket, tanggal, createdAt string
		var nominal int64
		rows.Scan(&id, &tipe, &nominal, &ket, &tanggal, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "tipe": tipe, "nominal": nominal,
			"keterangan": ket, "tanggal": tanggal, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func GetSaldoKeuangan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")

	// 1. Total uang masuk (semua: SPP + Insidental + Manual)
	// Semua pemasukan sudah tercatat di catatan_keuangan dengan tipe = 'masuk'
	qMasuk := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk'"
	argsMasuk := []interface{}{tid}
	if tglMulai != "" {
		qMasuk += " AND tanggal >= ?"
		argsMasuk = append(argsMasuk, tglMulai)
	}
	if tglAkhir != "" {
		qMasuk += " AND tanggal <= ?"
		argsMasuk = append(argsMasuk, tglAkhir)
	}
	var totalMasuk int64
	config.DB.QueryRow(qMasuk, argsMasuk...).Scan(&totalMasuk)

	// 2. Total SPP saja (untuk ditampilkan sebagai breakdown)
	qSpp := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk' AND keterangan LIKE 'SPP %'"
	argsSpp := []interface{}{tid}
	if tglMulai != "" {
		qSpp += " AND tanggal >= ?"
		argsSpp = append(argsSpp, tglMulai)
	}
	if tglAkhir != "" {
		qSpp += " AND tanggal <= ?"
		argsSpp = append(argsSpp, tglAkhir)
	}
	var totalSPP int64
	config.DB.QueryRow(qSpp, argsSpp...).Scan(&totalSPP)

	// 3. Total insidental (untuk breakdown)
	qIns := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk' AND keterangan LIKE 'Tagihan %'"
	argsIns := []interface{}{tid}
	if tglMulai != "" {
		qIns += " AND tanggal >= ?"
		argsIns = append(argsIns, tglMulai)
	}
	if tglAkhir != "" {
		qIns += " AND tanggal <= ?"
		argsIns = append(argsIns, tglAkhir)
	}
	var totalInsidental int64
	config.DB.QueryRow(qIns, argsIns...).Scan(&totalInsidental)

	// 4. Total uang keluar
	qKeluar := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'keluar'"
	argsKeluar := []interface{}{tid}
	if tglMulai != "" {
		qKeluar += " AND tanggal >= ?"
		argsKeluar = append(argsKeluar, tglMulai)
	}
	if tglAkhir != "" {
		qKeluar += " AND tanggal <= ?"
		argsKeluar = append(argsKeluar, tglAkhir)
	}
	var totalKeluar int64
	config.DB.QueryRow(qKeluar, argsKeluar...).Scan(&totalKeluar)

	// 5. Total masuk lain-lain (non-SPP, non-insidental)
	totalLain := totalMasuk - totalSPP - totalInsidental

	saldo := totalMasuk - totalKeluar

	return c.JSON(fiber.Map{
		"total_spp":        totalSPP,
		"total_insidental": totalInsidental,
		"total_masuk_lain": totalLain,
		"total_masuk":      totalMasuk,
		"total_keluar":     totalKeluar,
		"saldo":            saldo,
	})
}

func CreateCatatanKeuangan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		Tipe       string `json:"tipe"`
		Nominal    int64  `json:"nominal"`
		Keterangan string `json:"keterangan"`
		Tanggal    string `json:"tanggal"`
	}
	c.BodyParser(&body)
	if (body.Tipe != "masuk" && body.Tipe != "keluar") || body.Nominal <= 0 || body.Tanggal == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Tipe (masuk/keluar), nominal, dan tanggal wajib diisi"})
	}
	res, err := config.DB.Exec(
		"INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?,?,?,?,?,?)",
		tid, body.Tipe, body.Nominal, body.Keterangan, body.Tanggal, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Catatan keuangan ditambahkan"})
}

func DeleteCatatanKeuangan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM catatan_keuangan WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Catatan dihapus"})
}

func ExportKeuanganExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")

	// Helper for date filter
	addDateFilter := func(q string, args []interface{}, col string) (string, []interface{}) {
		if tglMulai != "" {
			q += fmt.Sprintf(" AND %s >= ?", col)
			args = append(args, tglMulai)
		}
		if tglAkhir != "" {
			q += fmt.Sprintf(" AND %s <= ?", col)
			args = append(args, tglAkhir)
		}
		return q, args
	}

	// Totals - semua dari catatan_keuangan (single source of truth)
	var totalSPP, totalMasuk, totalInsidental, totalKeluar int64

	// Total semua pemasukan
	qM := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk'"
	argsM := []interface{}{tid}
	qM, argsM = addDateFilter(qM, argsM, "tanggal")
	config.DB.QueryRow(qM, argsM...).Scan(&totalMasuk)

	// Total SPP saja
	qSpp := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk' AND keterangan LIKE 'SPP %'"
	argsSpp := []interface{}{tid}
	qSpp, argsSpp = addDateFilter(qSpp, argsSpp, "tanggal")
	config.DB.QueryRow(qSpp, argsSpp...).Scan(&totalSPP)

	// Total insidental
	qIns := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk' AND keterangan LIKE 'Tagihan %'"
	argsIns := []interface{}{tid}
	qIns, argsIns = addDateFilter(qIns, argsIns, "tanggal")
	config.DB.QueryRow(qIns, argsIns...).Scan(&totalInsidental)

	qK := "SELECT COALESCE(SUM(nominal),0) FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'keluar'"
	argsK := []interface{}{tid}
	qK, argsK = addDateFilter(qK, argsK, "tanggal")
	config.DB.QueryRow(qK, argsK...).Scan(&totalKeluar)

	f := excelize.NewFile()

	periode := "Semua Periode"
	if tglMulai != "" && tglAkhir != "" {
		periode = tglMulai + " s/d " + tglAkhir
	} else if tglMulai != "" {
		periode = "Sejak " + tglMulai
	} else if tglAkhir != "" {
		periode = "Sampai " + tglAkhir
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14, Color: "#16a34a"}})
	headerGreen, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#16a34a"}, Pattern: 1},
	})
	headerRed, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#ef4444"}, Pattern: 1},
	})
	greenFont, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "#16a34a"}})
	boldBig, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 13}})
	totalRowStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#f0fdf4"}, Pattern: 1},
	})
	totalRowRedStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#ef4444"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#fef2f2"}, Pattern: 1},
	})

	// ═══════════════ Sheet 1: RINGKASAN ═══════════════
	ring := "Ringkasan"
	f.SetSheetName("Sheet1", ring)
	f.SetCellValue(ring, "A1", "LAPORAN KEUANGAN")
	f.SetCellStyle(ring, "A1", "A1", titleStyle)
	f.SetCellValue(ring, "A2", "Periode: "+periode)

	f.SetCellValue(ring, "A4", "Uraian")
	f.SetCellValue(ring, "B4", "Jumlah (Rp)")
	f.SetCellStyle(ring, "A4", "B4", headerGreen)

	f.SetCellValue(ring, "A5", "Pemasukan SPP")
	f.SetCellValue(ring, "B5", totalSPP)
	f.SetCellStyle(ring, "B5", "B5", greenFont)
	f.SetCellValue(ring, "A6", "Pemasukan Insidental")
	f.SetCellValue(ring, "B6", totalInsidental)
	f.SetCellStyle(ring, "B6", "B6", greenFont)
	f.SetCellValue(ring, "A7", "Pemasukan Lain")
	f.SetCellValue(ring, "B7", totalMasuk-totalSPP-totalInsidental)
	f.SetCellStyle(ring, "B7", "B7", greenFont)
	f.SetCellValue(ring, "A8", "TOTAL PEMASUKAN")
	f.SetCellValue(ring, "B8", totalMasuk)
	f.SetCellStyle(ring, "A8", "B8", totalRowStyle)

	f.SetCellValue(ring, "A10", "Total Pengeluaran")
	f.SetCellValue(ring, "B10", totalKeluar)
	f.SetCellStyle(ring, "A10", "B10", totalRowRedStyle)

	f.SetCellValue(ring, "A12", "SALDO AKHIR")
	f.SetCellValue(ring, "B12", totalMasuk-totalKeluar)
	f.SetCellStyle(ring, "A12", "B12", boldBig)

	f.SetColWidth(ring, "A", "A", 30)
	f.SetColWidth(ring, "B", "B", 20)

	// ═══════════════ Sheet 2: UANG MASUK (Rinci) ═══════════════
	masuk := "Uang Masuk"
	f.NewSheet(masuk)
	f.SetCellValue(masuk, "A1", "RINCIAN UANG MASUK")
	f.SetCellStyle(masuk, "A1", "A1", titleStyle)
	f.SetCellValue(masuk, "A2", "Periode: "+periode)

	// Section A: SPP & Insidental (summary)
	f.SetCellValue(masuk, "A4", "A. PEMASUKAN SPP & INSIDENTAL")
	f.SetCellStyle(masuk, "A4", "A4", greenFont)
	mHeaders := []string{"No", "Jenis Pemasukan", "Nominal"}
	for i, h := range mHeaders {
		cell := fmt.Sprintf("%c5", 'A'+i)
		f.SetCellValue(masuk, cell, h)
	}
	f.SetCellStyle(masuk, "A5", "C5", headerGreen)

	f.SetCellValue(masuk, "A6", 1)
	f.SetCellValue(masuk, "B6", "Pembayaran SPP Santri")
	f.SetCellValue(masuk, "C6", totalSPP)
	f.SetCellStyle(masuk, "C6", "C6", greenFont)
	f.SetCellValue(masuk, "A7", 2)
	f.SetCellValue(masuk, "B7", "Pembayaran Insidental")
	f.SetCellValue(masuk, "C7", totalInsidental)
	f.SetCellStyle(masuk, "C7", "C7", greenFont)

	row := 9
	no := 1

	// Section B: Semua detail pemasukan
	f.SetCellValue(masuk, fmt.Sprintf("A%d", row), "B. RINCIAN SEMUA PEMASUKAN")
	f.SetCellStyle(masuk, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), greenFont)
	row++
	bHeaders := []string{"No", "Tanggal", "Nominal", "Keterangan"}
	for i, h := range bHeaders {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(masuk, cell, h)
	}
	f.SetCellStyle(masuk, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerGreen)
	row++

	qMasukD := `SELECT tanggal, nominal, COALESCE(keterangan,'') FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'masuk'`
	argsMD := []interface{}{tid}
	qMasukD, argsMD = addDateFilter(qMasukD, argsMD, "tanggal")
	qMasukD += " ORDER BY tanggal ASC"
	masukRows, _ := config.DB.Query(qMasukD, argsMD...)
	no = 1
	var sumMasuk int64
	if masukRows != nil {
		defer masukRows.Close()
		for masukRows.Next() {
			var tgl, ket string
			var nominal int64
			masukRows.Scan(&tgl, &nominal, &ket)
			f.SetCellValue(masuk, fmt.Sprintf("A%d", row), no)
			f.SetCellValue(masuk, fmt.Sprintf("B%d", row), tgl)
			f.SetCellValue(masuk, fmt.Sprintf("C%d", row), nominal)
			f.SetCellValue(masuk, fmt.Sprintf("D%d", row), ket)
			sumMasuk += nominal
			row++
			no++
		}
	}
	f.SetCellValue(masuk, fmt.Sprintf("B%d", row), "TOTAL")
	f.SetCellValue(masuk, fmt.Sprintf("C%d", row), sumMasuk)
	f.SetCellStyle(masuk, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), totalRowStyle)
	row += 2
	f.SetCellValue(masuk, fmt.Sprintf("B%d", row), "TOTAL UANG MASUK")
	f.SetCellValue(masuk, fmt.Sprintf("C%d", row), totalMasuk)
	f.SetCellStyle(masuk, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), boldBig)

	f.SetColWidth(masuk, "A", "A", 5)
	f.SetColWidth(masuk, "B", "B", 15)
	f.SetColWidth(masuk, "C", "C", 22)
	f.SetColWidth(masuk, "D", "D", 12)
	f.SetColWidth(masuk, "E", "E", 12)
	f.SetColWidth(masuk, "F", "F", 18)
	f.SetColWidth(masuk, "G", "G", 30)

	// ═══════════════ Sheet 3: UANG KELUAR (Rinci) ═══════════════
	keluar := "Uang Keluar"
	f.NewSheet(keluar)
	f.SetCellValue(keluar, "A1", "RINCIAN UANG KELUAR")
	f.SetCellStyle(keluar, "A1", "A1", titleStyle)
	f.SetCellValue(keluar, "A2", "Periode: "+periode)

	kHeaders := []string{"No", "Tanggal", "Nominal", "Keterangan"}
	for i, h := range kHeaders {
		cell := fmt.Sprintf("%c4", 'A'+i)
		f.SetCellValue(keluar, cell, h)
	}
	f.SetCellStyle(keluar, "A4", "D4", headerRed)

	qKeluarD := `SELECT tanggal, nominal, COALESCE(keterangan,'') FROM catatan_keuangan WHERE tenant_id = ? AND tipe = 'keluar'`
	argsKD := []interface{}{tid}
	qKeluarD, argsKD = addDateFilter(qKeluarD, argsKD, "tanggal")
	qKeluarD += " ORDER BY tanggal ASC"
	keluarRows, _ := config.DB.Query(qKeluarD, argsKD...)
	row = 5
	no = 1
	var sumKeluar int64
	if keluarRows != nil {
		defer keluarRows.Close()
		for keluarRows.Next() {
			var tgl, ket string
			var nominal int64
			keluarRows.Scan(&tgl, &nominal, &ket)
			f.SetCellValue(keluar, fmt.Sprintf("A%d", row), no)
			f.SetCellValue(keluar, fmt.Sprintf("B%d", row), tgl)
			f.SetCellValue(keluar, fmt.Sprintf("C%d", row), nominal)
			f.SetCellValue(keluar, fmt.Sprintf("D%d", row), ket)
			sumKeluar += nominal
			row++
			no++
		}
	}
	if no == 1 {
		f.SetCellValue(keluar, fmt.Sprintf("A%d", row), "-")
		f.SetCellValue(keluar, fmt.Sprintf("D%d", row), "Belum ada pengeluaran")
		row++
	}
	f.SetCellValue(keluar, fmt.Sprintf("B%d", row), "TOTAL PENGELUARAN")
	f.SetCellValue(keluar, fmt.Sprintf("C%d", row), sumKeluar)
	f.SetCellStyle(keluar, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), totalRowRedStyle)

	row += 2
	f.SetCellValue(keluar, fmt.Sprintf("B%d", row), "SALDO AKHIR")
	f.SetCellValue(keluar, fmt.Sprintf("C%d", row), totalMasuk-sumKeluar)
	f.SetCellStyle(keluar, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), boldBig)

	f.SetColWidth(keluar, "A", "A", 5)
	f.SetColWidth(keluar, "B", "B", 18)
	f.SetColWidth(keluar, "C", "C", 18)
	f.SetColWidth(keluar, "D", "D", 40)

	filename := "Laporan_Keuangan"
	if tglMulai != "" {
		filename += "_" + tglMulai
	}
	if tglAkhir != "" {
		filename += "_" + tglAkhir
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", filename))
	return f.Write(c.Response().BodyWriter())
}
