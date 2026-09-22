package handlers

import (
	"bytes"
	"fmt"
	"strings"
	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

type AbsenItem struct {
	SantriID   int    `json:"santri_id"`
	Status     string `json:"status"`
	Keterangan string `json:"keterangan"`
}

// ── Absensi Kegiatan (bulk) ──────────────────────────────
func AbsensiBulk(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	uname := c.Locals("username").(string)
	role := c.Locals("role").(string)
	var body struct {
		Tanggal    string      `json:"tanggal"`
		KelompokID int         `json:"kelompok_id"`
		KegiatanID int         `json:"kegiatan_id"`
		Items      []AbsenItem `json:"items"`
	}
	c.BodyParser(&body)
	if len(body.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Data kosong"})
	}

	// Time-lock: check per-kelompok, not per-tenant
	if role == "ustadz" {
		now := helpers.JamSekarang()
		hari := helpers.HariIni()
		// Check if THIS kelompok has jadwal for today (or "Setiap Hari")
		var jadwalCount int
		config.DB.QueryRow(
			"SELECT COUNT(*) FROM jadwal_umum WHERE tenant_id = ? AND kelompok_id = ? AND (hari = ? OR hari = 'Setiap Hari')",
			tid, body.KelompokID, hari,
		).Scan(&jadwalCount)
		if jadwalCount > 0 {
			// This kelompok has schedule -> enforce ustadz + time window
			// 1. Check if this ustadz is assigned to this kelompok today
			var ustadzMatch int
			config.DB.QueryRow(
				"SELECT COUNT(*) FROM jadwal_umum WHERE tenant_id = ? AND kelompok_id = ? AND ustadz_username = ? AND (hari = ? OR hari = 'Setiap Hari')",
				tid, body.KelompokID, uname, hari,
			).Scan(&ustadzMatch)
			if ustadzMatch == 0 {
				return c.Status(403).JSON(fiber.Map{
					"message": "Anda tidak dijadwalkan untuk kelompok ini hari " + hari,
					"code":    "NOT_ASSIGNED",
				})
			}
			// 2. Check time window: jam_mulai <= now <= jam_selesai + 30 min
			var jamMulai, jamSelesai string
			config.DB.QueryRow(
				"SELECT jam_mulai, jam_selesai FROM jadwal_umum WHERE tenant_id = ? AND kelompok_id = ? AND ustadz_username = ? AND (hari = ? OR hari = 'Setiap Hari') ORDER BY jam_mulai LIMIT 1",
				tid, body.KelompokID, uname, hari,
			).Scan(&jamMulai, &jamSelesai)
			if jamMulai != "" && now < jamMulai {
				return c.Status(403).JSON(fiber.Map{
					"message": fmt.Sprintf("Belum waktunya absen. Mulai jam %s WIB (sekarang %s)", jamMulai, now),
					"code":    "NOT_IN_SCHEDULE",
				})
			}
			batas := helpers.AddMinutes(jamSelesai, 30)
			if jamSelesai != "" && now > batas {
				return c.Status(403).JSON(fiber.Map{
					"message": fmt.Sprintf("Waktu absen sudah lewat. Batas: %s WIB (sekarang %s)", batas, now),
					"code":    "SCHEDULE_EXPIRED",
				})
			}
		}
		// If jadwalCount == 0 -> this kelompok has no schedule, allow free attendance
	}

	recordedAt := helpers.ToDatetime()
	res, err := config.DB.Exec(
		"INSERT INTO absensi_sesi (tenant_id, ustadz_username, kelompok_id, tanggal, recorded_at) VALUES (?,?,?,?,?)",
		tid, uname, body.KelompokID, body.Tanggal, recordedAt,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan sesi: " + err.Error()})
	}
	sesiID, _ := res.LastInsertId()
	// Auto-izin: override status for santri with active perizinan
	izinMap := GetSantriIzinAktif(tid)
	for _, item := range body.Items {
		st := item.Status
		ket := item.Keterangan
		if izinKet, ok := izinMap[item.SantriID]; ok {
			st = "I"
			ket = "Izin resmi: " + izinKet
		}
		_, err := config.DB.Exec(
			"INSERT INTO absensi (tenant_id, santri_id, kelompok_id, sesi_id, kegiatan_id, tanggal, status, keterangan, recorded_by) VALUES (?,?,?,?,?,?,?,?,?)",
			tid, item.SantriID, body.KelompokID, sesiID, body.KegiatanID, body.Tanggal, st, ket, uid,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan absensi: " + err.Error()})
		}
	}
	return c.JSON(fiber.Map{"message": "Absensi tersimpan", "recorded_at": recordedAt})
}

// ── Rekap Absensi Kegiatan ───────────────────────────────
func GetRekap(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := "SELECT a.id, a.santri_id, COALESCE(s.nama,''), a.kelompok_id, COALESCE(k.nama,''), COALESCE(k.kegiatan_nama,''), a.tanggal, a.status, COALESCE(a.keterangan,'') FROM absensi a LEFT JOIN santri s ON a.santri_id = s.id LEFT JOIN kelompok k ON a.kelompok_id = k.id WHERE a.tenant_id = ?"
	args := []interface{}{tid}
	if v := c.Query("dari"); v != "" { q += " AND a.tanggal >= ?"; args = append(args, v) }
	if v := c.Query("sampai"); v != "" { q += " AND a.tanggal <= ?"; args = append(args, v) }
	if v := c.Query("santri_id"); v != "" { q += " AND a.santri_id = ?"; args = append(args, v) }
	if v := c.Query("kelompok_id"); v != "" { q += " AND a.kelompok_id = ?"; args = append(args, v) }
	q += " ORDER BY a.tanggal DESC"
	rows, _ := config.DB.Query(q, args...)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, sid, kid int
		var sn, kn, kgn, tgl, st, ket string
		rows.Scan(&id, &sid, &sn, &kid, &kn, &kgn, &tgl, &st, &ket)
		list = append(list, fiber.Map{"id": id, "santri_id": sid, "santri_nama": sn, "kelompok_id": kid, "kelompok_nama": kn, "kegiatan_nama": kgn, "tanggal": tgl, "status": st, "keterangan": ket})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

// ── Rekap Absensi Summary (grouped by santri) ────────────
func GetRekapSummary(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari")
	sampai := c.Query("sampai")
	kegNama := c.Query("kegiatan_nama")
	kelID := c.Query("kelompok_id")

	// 1. Count distinct sessions (pertemuan) in the period
	sesiQ := "SELECT COUNT(DISTINCT id) FROM absensi_sesi WHERE tenant_id = ?"
	sesiArgs := []interface{}{tid}
	if dari != "" { sesiQ += " AND tanggal >= ?"; sesiArgs = append(sesiArgs, dari) }
	if sampai != "" { sesiQ += " AND tanggal <= ?"; sesiArgs = append(sesiArgs, sampai) }
	if kelID != "" { sesiQ += " AND kelompok_id = ?"; sesiArgs = append(sesiArgs, kelID) }
	var jumlahPertemuan int
	config.DB.QueryRow(sesiQ, sesiArgs...).Scan(&jumlahPertemuan)

	// 2. Aggregate attendance per santri
	q := `SELECT a.santri_id, COALESCE(s.nama,''),
		SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END)
		FROM absensi a
		LEFT JOIN santri s ON a.santri_id = s.id
		LEFT JOIN kelompok k ON a.kelompok_id = k.id
		WHERE a.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND a.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND a.tanggal <= ?"; args = append(args, sampai) }
	if kegNama != "" { q += " AND k.kegiatan_nama = ?"; args = append(args, kegNama) }
	if kelID != "" { q += " AND a.kelompok_id = ?"; args = append(args, kelID) }
	q += " GROUP BY a.santri_id, s.nama ORDER BY s.nama"

	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	var totalAbsen int
	for rows.Next() {
		var sid, h, i, s, a int
		var nama string
		rows.Scan(&sid, &nama, &h, &i, &s, &a)
		total := h + i + s + a
		totalAbsen += total
		list = append(list, fiber.Map{"santri_id": sid, "santri_nama": nama, "H": h, "I": i, "S": s, "A": a, "total": total})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(fiber.Map{
		"jumlah_pertemuan": jumlahPertemuan,
		"jumlah_absensi":   totalAbsen,
		"data":             list,
	})
}

// ── Export Rekap Absensi ke Excel ─────────────────────────
func ExportRekapExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari")
	sampai := c.Query("sampai")
	kegNama := c.Query("kegiatan_nama")
	kelID := c.Query("kelompok_id")

	// Lembaga name
	var lembaga string
	config.DB.QueryRow("SELECT COALESCE(app_name,'Pesantren') FROM settings WHERE tenant_id = ?", tid).Scan(&lembaga)
	if lembaga == "" { lembaga = "Pesantren" }

	// Session count
	sesiQ := "SELECT COUNT(DISTINCT id) FROM absensi_sesi WHERE tenant_id = ?"
	sesiArgs := []interface{}{tid}
	if dari != "" { sesiQ += " AND tanggal >= ?"; sesiArgs = append(sesiArgs, dari) }
	if sampai != "" { sesiQ += " AND tanggal <= ?"; sesiArgs = append(sesiArgs, sampai) }
	if kelID != "" { sesiQ += " AND kelompok_id = ?"; sesiArgs = append(sesiArgs, kelID) }
	var jumlahPertemuan int
	config.DB.QueryRow(sesiQ, sesiArgs...).Scan(&jumlahPertemuan)

	// Aggregate
	q := `SELECT a.santri_id, COALESCE(s.nama,''),
		SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END)
		FROM absensi a
		LEFT JOIN santri s ON a.santri_id = s.id
		LEFT JOIN kelompok k ON a.kelompok_id = k.id
		WHERE a.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND a.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND a.tanggal <= ?"; args = append(args, sampai) }
	if kegNama != "" { q += " AND k.kegiatan_nama = ?"; args = append(args, kegNama) }
	if kelID != "" { q += " AND a.kelompok_id = ?"; args = append(args, kelID) }
	q += " GROUP BY a.santri_id, s.nama ORDER BY s.nama"

	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()

	type SantriRekap struct {
		Nama           string
		H, I, S, A     int
	}
	var data []SantriRekap
	for rows.Next() {
		var sid, h, i, s, a int
		var nama string
		rows.Scan(&sid, &nama, &h, &i, &s, &a)
		data = append(data, SantriRekap{Nama: nama, H: h, I: i, S: s, A: a})
	}

	periode := ""
	if dari != "" && sampai != "" { periode = dari + " s/d " + sampai }

	f := excelize.NewFile()
	sheet := "Rekap Absensi"
	f.SetSheetName("Sheet1", sheet)

	// Header styling
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1B4332"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	cellLeftStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D8F3DC"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})

	// Title
	f.MergeCell(sheet, "A1", "G1")
	f.SetCellValue(sheet, "A1", "REKAP ABSENSI - "+strings.ToUpper(lembaga))
	f.SetCellStyle(sheet, "A1", "G1", titleStyle)

	r := 3
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Periode")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), ": "+periode)
	r++
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Jumlah Pertemuan")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf(": %d sesi", jumlahPertemuan))
	r++
	kegLabel := kegNama
	if kegLabel == "" { kegLabel = "Semua" }
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Kegiatan")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), ": "+kegLabel)
	r += 2

	// Table headers
	hdrs := []string{"No", "Nama Santri", "Hadir", "Izin", "Sakit", "Alpha", "Total"}
	for i, h := range hdrs {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h)
		f.SetCellStyle(sheet, fmt.Sprintf("%s%d", col, r), fmt.Sprintf("%s%d", col, r), headerStyle)
	}
	r++

	// Set column widths
	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "G", 10)

	tH, tI, tS, tA := 0, 0, 0, 0
	for idx, d := range data {
		total := d.H + d.I + d.S + d.A
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), idx+1)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("A%d", r), cellStyle)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), d.Nama)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("B%d", r), cellLeftStyle)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), d.H)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), d.I)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), d.S)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), d.A)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), total)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", r), fmt.Sprintf("G%d", r), cellStyle)
		tH += d.H; tI += d.I; tS += d.S; tA += d.A
		r++
	}

	// Total row
	f.MergeCell(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("B%d", r))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "TOTAL")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", r), tH)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", r), tI)
	f.SetCellValue(sheet, fmt.Sprintf("E%d", r), tS)
	f.SetCellValue(sheet, fmt.Sprintf("F%d", r), tA)
	f.SetCellValue(sheet, fmt.Sprintf("G%d", r), tH+tI+tS+tA)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("G%d", r), totalStyle)

	var buf bytes.Buffer
	f.Write(&buf)
	filename := "Rekap_Absensi"
	if dari != "" { filename += "_" + dari }
	if sampai != "" { filename += "_" + sampai }
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", filename))
	return c.Send(buf.Bytes())
}

// ── Absen Malam (bulk per kamar) ─────────────────────────
func AbsenMalamBulk(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Tanggal string      `json:"tanggal"`
		KamarID int         `json:"kamar_id"`
		Items   []AbsenItem `json:"items"`
	}
	c.BodyParser(&body)
	if len(body.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Data kosong"})
	}
	// Auto-izin: override status for santri with active perizinan
	izinMap := GetSantriIzinAktif(tid)
	for _, item := range body.Items {
		st := item.Status
		ket := item.Keterangan
		if izinKet, ok := izinMap[item.SantriID]; ok {
			st = "I"
			ket = "Izin resmi: " + izinKet
		}
		_, err := config.DB.Exec(
			"INSERT INTO absen_malam (tenant_id, santri_id, kamar_id, tanggal, status, keterangan) VALUES (?,?,?,?,?,?)",
			tid, item.SantriID, body.KamarID, body.Tanggal, st, ket)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan: " + err.Error()})
		}
	}
	return c.JSON(fiber.Map{"message": "Absen malam tersimpan"})
}

// ── Get Absen Malam (per tanggal+kamar) ──────────────────
func GetAbsenMalam(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tanggal := c.Query("tanggal")
	if tanggal == "" {
		tanggal = helpers.TodayWIB()
	}
	q := "SELECT am.id, am.santri_id, COALESCE(s.nama,''), am.status, COALESCE(am.keterangan,''), COALESCE(am.kamar_id,0) FROM absen_malam am LEFT JOIN santri s ON am.santri_id = s.id WHERE am.tenant_id = ? AND am.tanggal = ?"
	args := []interface{}{tid, tanggal}
	if kamarID := c.Query("kamar_id"); kamarID != "" {
		q += " AND am.kamar_id = ?"
		args = append(args, kamarID)
	}
	q += " ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, sid, kid int
		var nama, status, ket string
		rows.Scan(&id, &sid, &nama, &status, &ket, &kid)
		list = append(list, fiber.Map{"id": id, "santri_id": sid, "santri_nama": nama, "status": status, "keterangan": ket, "kamar_id": kid})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

// ── Rekap Absen Malam (per periode/kamar) ────────────────
func GetRekapAbsenMalam(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari"); sampai := c.Query("sampai"); kamarID := c.Query("kamar_id")

	// Count distinct sessions
	sesiQ := `SELECT COUNT(DISTINCT am.tanggal) FROM absen_malam am WHERE am.tenant_id = ?`
	sesiArgs := []interface{}{tid}
	if dari != "" { sesiQ += " AND am.tanggal >= ?"; sesiArgs = append(sesiArgs, dari) }
	if sampai != "" { sesiQ += " AND am.tanggal <= ?"; sesiArgs = append(sesiArgs, sampai) }
	if kamarID != "" { sesiQ += " AND am.kamar_id = ?"; sesiArgs = append(sesiArgs, kamarID) }
	var jumlahSesi int
	config.DB.QueryRow(sesiQ, sesiArgs...).Scan(&jumlahSesi)

	q := `SELECT am.santri_id, COALESCE(s.nama,''), COALESCE(k.nama,'-'),
		SUM(CASE WHEN am.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN am.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN am.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN am.status='A' THEN 1 ELSE 0 END)
		FROM absen_malam am
		LEFT JOIN santri s ON am.santri_id = s.id
		LEFT JOIN kamar k ON am.kamar_id = k.id
		WHERE am.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND am.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND am.tanggal <= ?"; args = append(args, sampai) }
	if kamarID != "" { q += " AND am.kamar_id = ?"; args = append(args, kamarID) }
	q += " GROUP BY am.santri_id, s.nama, k.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var sid, h, i, s, a int
		var nama, kamar string
		rows.Scan(&sid, &nama, &kamar, &h, &i, &s, &a)
		list = append(list, fiber.Map{"santri_id": sid, "santri_nama": nama, "kamar": kamar, "H": h, "I": i, "S": s, "A": a, "total": h + i + s + a})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(fiber.Map{"jumlah_sesi": jumlahSesi, "data": list})
}

// ── Export Rekap Absen Kamar → Excel ─────────────────────
func ExportRekapMalamExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari"); sampai := c.Query("sampai"); kamarID := c.Query("kamar_id")
	q := `SELECT am.santri_id, COALESCE(s.nama,''), COALESCE(k.nama,'-'),
		SUM(CASE WHEN am.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN am.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN am.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN am.status='A' THEN 1 ELSE 0 END)
		FROM absen_malam am
		LEFT JOIN santri s ON am.santri_id = s.id
		LEFT JOIN kamar k ON am.kamar_id = k.id
		WHERE am.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND am.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND am.tanggal <= ?"; args = append(args, sampai) }
	if kamarID != "" { q += " AND am.kamar_id = ?"; args = append(args, kamarID) }
	q += " GROUP BY am.santri_id, s.nama, k.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Rekap Kamar"
	f.SetSheetName("Sheet1", sheet)
	periode := dari + " s/d " + sampai
	f.SetCellValue(sheet, "A1", "REKAP ABSENSI KAMAR (ASRAMA)")
	f.SetCellValue(sheet, "A2", "Periode: "+periode)
	r := 4
	hdrs := []string{"No", "Nama Santri", "Kamar", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
	for i, h := range hdrs { col := string(rune('A'+i)); f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h) }
	r++
	no := 1
	for rows.Next() {
		var sid, h, i, s, a int; var nama, kamar string
		rows.Scan(&sid, &nama, &kamar, &h, &i, &s, &a)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), no)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), kamar)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), h)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), i)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), s)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), a)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), h+i+s+a)
		no++; r++
	}
	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=Rekap_Absen_Kamar.xlsx")
	return c.Send(buf.Bytes())
}

// ── Absen Sekolah (bulk per kelas+jadwal) ────────────────
func AbsenSekolahBulk(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uname := c.Locals("username").(string)
	role := c.Locals("role").(string)
	var body struct {
		Tanggal         string      `json:"tanggal"`
		KelasID         int         `json:"kelas_id"`
		JadwalSekolahID int         `json:"jadwal_sekolah_id"`
		MataPelajaran   string      `json:"mata_pelajaran"`
		Items           []AbsenItem `json:"items"`
	}
	c.BodyParser(&body)
	if len(body.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Data kosong"})
	}

	// Time-lock: only if jadwal_sekolah exists for THIS kelas
	if role == "ustadz" && body.KelasID > 0 {
		var kelasJadwalCount int
		config.DB.QueryRow("SELECT COUNT(*) FROM jadwal_sekolah WHERE tenant_id = ? AND kelas_id = ?", tid, body.KelasID).Scan(&kelasJadwalCount)
		if kelasJadwalCount > 0 {
			// This kelas has schedule -> enforce ustadz + time window
			now := helpers.JamSekarang()
			hari := helpers.HariIni()
			var count int
			config.DB.QueryRow(
				"SELECT COUNT(*) FROM jadwal_sekolah WHERE tenant_id = ? AND kelas_id = ? AND ustadz_username = ? AND hari = ? AND ? >= jam_mulai",
				tid, body.KelasID, uname, hari, now,
			).Scan(&count)
			if count == 0 {
				return c.Status(403).JSON(fiber.Map{
					"message": fmt.Sprintf("Tidak ada jadwal untuk kelas ini hari %s atau belum waktunya (sekarang %s WIB)", hari, now),
					"code":    "NOT_IN_SCHEDULE",
				})
			}
		}
		// If kelasJadwalCount == 0 -> this kelas has no schedule, allow free attendance
	}

	// Ambil info jadwal jika jadwal_sekolah_id diberikan
	if body.JadwalSekolahID > 0 && body.MataPelajaran == "" {
		config.DB.QueryRow("SELECT mata_pelajaran FROM jadwal_sekolah WHERE id = ? AND tenant_id = ?",
			body.JadwalSekolahID, tid).Scan(&body.MataPelajaran)
	}

	// Track session
	res, _ := config.DB.Exec(
		"INSERT INTO absen_sekolah_sesi (tenant_id, kelas_id, jadwal_sekolah_id, tanggal, ustadz_username, mata_pelajaran) VALUES (?,?,?,?,?,?)",
		tid, body.KelasID, body.JadwalSekolahID, body.Tanggal, uname, body.MataPelajaran)
	sesiID, _ := res.LastInsertId()
	_ = sesiID

	// Auto-izin: override status for santri with active perizinan
	izinMap := GetSantriIzinAktif(tid)
	for _, item := range body.Items {
		st := item.Status
		ket := item.Keterangan
		if izinKet, ok := izinMap[item.SantriID]; ok {
			st = "I"
			ket = "Izin resmi: " + izinKet
		}
		_, err := config.DB.Exec(
			"INSERT INTO absen_sekolah (tenant_id, santri_id, kelas_id, jadwal_sekolah_id, mata_pelajaran, tanggal, status, keterangan, ustadz_username) VALUES (?,?,?,?,?,?,?,?,?)",
			tid, item.SantriID, body.KelasID, body.JadwalSekolahID, body.MataPelajaran, body.Tanggal, st, ket, uname)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan: " + err.Error()})
		}
	}
	return c.JSON(fiber.Map{"message": "Absen sekolah tersimpan"})
}

// ── Get Absen Sekolah (cek sudah diabsen?) ───────────────
func GetAbsenSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tanggal := c.Query("tanggal")
	if tanggal == "" { tanggal = helpers.TodayWIB() }
	q := `SELECT a.id, a.santri_id, COALESCE(s.nama,''), a.status, COALESCE(a.keterangan,''),
		COALESCE(a.mata_pelajaran,''), COALESCE(a.kelas_id,0), COALESCE(a.jadwal_sekolah_id,0)
		FROM absen_sekolah a LEFT JOIN santri s ON a.santri_id = s.id
		WHERE a.tenant_id = ? AND a.tanggal = ?`
	args := []interface{}{tid, tanggal}
	if v := c.Query("kelas_id"); v != "" { q += " AND a.kelas_id = ?"; args = append(args, v) }
	if v := c.Query("jadwal_sekolah_id"); v != "" { q += " AND a.jadwal_sekolah_id = ?"; args = append(args, v) }
	q += " ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, sid, kid, jid int
		var nama, status, ket, mapel string
		rows.Scan(&id, &sid, &nama, &status, &ket, &mapel, &kid, &jid)
		list = append(list, fiber.Map{"id": id, "santri_id": sid, "santri_nama": nama, "status": status, "keterangan": ket, "mata_pelajaran": mapel, "kelas_id": kid, "jadwal_sekolah_id": jid})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

// ── Rekap Absen Sekolah (per sesi, grouped by santri) ────
func GetRekapAbsenSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari")
	sampai := c.Query("sampai")
	kelasID := c.Query("kelas_id")

	// Count total sessions
	sesiQ := "SELECT COUNT(*) FROM absen_sekolah_sesi WHERE tenant_id = ?"
	sesiArgs := []interface{}{tid}
	if dari != "" { sesiQ += " AND tanggal >= ?"; sesiArgs = append(sesiArgs, dari) }
	if sampai != "" { sesiQ += " AND tanggal <= ?"; sesiArgs = append(sesiArgs, sampai) }
	if kelasID != "" { sesiQ += " AND kelas_id = ?"; sesiArgs = append(sesiArgs, kelasID) }
	var jumlahSesi int
	config.DB.QueryRow(sesiQ, sesiArgs...).Scan(&jumlahSesi)

	q := `SELECT a.santri_id, COALESCE(s.nama,''), COALESCE(ks.nama,'-'),
		SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END)
		FROM absen_sekolah a
		LEFT JOIN santri s ON a.santri_id = s.id
		LEFT JOIN kelas_sekolah ks ON a.kelas_id = ks.id
		WHERE a.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND a.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND a.tanggal <= ?"; args = append(args, sampai) }
	if kelasID != "" { q += " AND a.kelas_id = ?"; args = append(args, kelasID) }
	q += " GROUP BY a.santri_id, s.nama, ks.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var sid, h, i, s, a int
		var nama, kelas string
		rows.Scan(&sid, &nama, &kelas, &h, &i, &s, &a)
		list = append(list, fiber.Map{"santri_id": sid, "santri_nama": nama, "kelas": kelas, "H": h, "I": i, "S": s, "A": a, "total": h + i + s + a})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(fiber.Map{"jumlah_sesi": jumlahSesi, "data": list})
}

// ── Rekap Absen Diniyyah ────────────────────────────────
func GetRekapAbsenDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari")
	sampai := c.Query("sampai")
	kelasID := c.Query("kelas_diniyyah_id")

	// Count distinct sessions (unique tanggal + jadwal_diniyyah_id per kelas)
	sesiQ := `SELECT COUNT(DISTINCT CONCAT(a.tanggal, '-', COALESCE(a.jadwal_diniyyah_id,0))) FROM absen_diniyyah a WHERE a.tenant_id = ?`
	sesiArgs := []interface{}{tid}
	if dari != "" { sesiQ += " AND a.tanggal >= ?"; sesiArgs = append(sesiArgs, dari) }
	if sampai != "" { sesiQ += " AND a.tanggal <= ?"; sesiArgs = append(sesiArgs, sampai) }
	if kelasID != "" { sesiQ += " AND a.kelas_diniyyah_id = ?"; sesiArgs = append(sesiArgs, kelasID) }
	var jumlahSesi int
	config.DB.QueryRow(sesiQ, sesiArgs...).Scan(&jumlahSesi)

	q := `SELECT a.santri_id, COALESCE(s.nama,''), COALESCE(kd.nama,'-'),
		SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END)
		FROM absen_diniyyah a
		LEFT JOIN santri s ON a.santri_id = s.id
		LEFT JOIN kelas_diniyyah kd ON a.kelas_diniyyah_id = kd.id
		WHERE a.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND a.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND a.tanggal <= ?"; args = append(args, sampai) }
	if kelasID != "" { q += " AND a.kelas_diniyyah_id = ?"; args = append(args, kelasID) }
	q += " GROUP BY a.santri_id, s.nama, kd.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var sid, h, i, s, a int
		var nama, kelas string
		rows.Scan(&sid, &nama, &kelas, &h, &i, &s, &a)
		list = append(list, fiber.Map{"santri_id": sid, "santri_nama": nama, "kelas": kelas, "H": h, "I": i, "S": s, "A": a, "total": h + i + s + a})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(fiber.Map{"jumlah_sesi": jumlahSesi, "data": list})
}

// ── Export Rekap Sekolah → Excel ──────────────────────────
func ExportRekapSekolahExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari"); sampai := c.Query("sampai"); kelasID := c.Query("kelas_id")
	q := `SELECT a.santri_id, COALESCE(s.nama,''), COALESCE(ks.nama,'-'),
		SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END)
		FROM absen_sekolah a
		LEFT JOIN santri s ON a.santri_id = s.id
		LEFT JOIN kelas_sekolah ks ON a.kelas_id = ks.id
		WHERE a.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND a.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND a.tanggal <= ?"; args = append(args, sampai) }
	if kelasID != "" { q += " AND a.kelas_id = ?"; args = append(args, kelasID) }
	q += " GROUP BY a.santri_id, s.nama, ks.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Rekap Sekolah"
	f.SetSheetName("Sheet1", sheet)
	periode := dari + " s/d " + sampai
	f.SetCellValue(sheet, "A1", "REKAP ABSENSI SEKOLAH")
	f.SetCellValue(sheet, "A2", "Periode: "+periode)
	r := 4
	hdrs := []string{"No", "Nama Santri", "Kelas", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
	for i, h := range hdrs { col := string(rune('A'+i)); f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h) }
	r++
	no := 1
	for rows.Next() {
		var sid, h, i, s, a int; var nama, kelas string
		rows.Scan(&sid, &nama, &kelas, &h, &i, &s, &a)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), no)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), kelas)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), h)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), i)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), s)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), a)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), h+i+s+a)
		no++; r++
	}
	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=Rekap_Absen_Sekolah.xlsx")
	return c.Send(buf.Bytes())
}

// ── Export Rekap Diniyyah → Excel ────────────────────────
func ExportRekapDiniyyahExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari"); sampai := c.Query("sampai"); kelasID := c.Query("kelas_diniyyah_id")
	q := `SELECT a.santri_id, COALESCE(s.nama,''), COALESCE(kd.nama,'-'),
		SUM(CASE WHEN a.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN a.status='A' THEN 1 ELSE 0 END)
		FROM absen_diniyyah a
		LEFT JOIN santri s ON a.santri_id = s.id
		LEFT JOIN kelas_diniyyah kd ON a.kelas_diniyyah_id = kd.id
		WHERE a.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" { q += " AND a.tanggal >= ?"; args = append(args, dari) }
	if sampai != "" { q += " AND a.tanggal <= ?"; args = append(args, sampai) }
	if kelasID != "" { q += " AND a.kelas_diniyyah_id = ?"; args = append(args, kelasID) }
	q += " GROUP BY a.santri_id, s.nama, kd.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Rekap Diniyyah"
	f.SetSheetName("Sheet1", sheet)
	periode := dari + " s/d " + sampai
	f.SetCellValue(sheet, "A1", "REKAP ABSENSI MADRASAH DINIYYAH")
	f.SetCellValue(sheet, "A2", "Periode: "+periode)
	r := 4
	hdrs := []string{"No", "Nama Santri", "Kelas", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
	for i, h := range hdrs { col := string(rune('A'+i)); f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h) }
	r++
	no := 1
	for rows.Next() {
		var sid, h, i, s, a int; var nama, kelas string
		rows.Scan(&sid, &nama, &kelas, &h, &i, &s, &a)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), no)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), kelas)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), h)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), i)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), s)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), a)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), h+i+s+a)
		no++; r++
	}
	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=Rekap_Absen_Diniyyah.xlsx")
	return c.Send(buf.Bytes())
}

// ── Server Time ──────────────────────────────────────────
func GetServerTime(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"waktu":    helpers.ToDatetime(),
		"tanggal":  helpers.TodayWIB(),
		"jam":      helpers.JamSekarang(),
		"hari":     helpers.HariIni(),
		"timezone": "WIB (UTC+7)",
	})
}
