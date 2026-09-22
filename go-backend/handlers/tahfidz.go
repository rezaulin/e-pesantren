package handlers

import (
	"bytes"
	"fmt"
	"math"
	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

func getPredikat(rata float64) string {
	switch {
	case rata >= 90:
		return "Mumtaz"
	case rata >= 80:
		return "Jayyid Jiddan"
	case rata >= 70:
		return "Jayyid"
	case rata >= 60:
		return "Maqbul"
	default:
		return "Dho'if"
	}
}

// ═══════════════════════════════════════════════════════════
// HALAQOH CRUD
// ═══════════════════════════════════════════════════════════

func GetHalaqoh(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query(`SELECT h.id, h.nama, COALESCE(h.musyrif,''), COALESCE(h.ustadz_username,''),
		COALESCE(h.target_juz,''), COALESCE(h.target_surah,''), h.target_ayat, COALESCE(h.target_deadline,''),
		(SELECT COUNT(*) FROM halaqoh_members hm WHERE hm.halaqoh_id = h.id AND hm.tenant_id = h.tenant_id AND hm.status='active')
		FROM halaqoh h WHERE h.tenant_id = ? ORDER BY h.nama`, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, targetAyat, memberCount int
		var nama, musyrif, uname, tJuz, tSurah, tDeadline string
		rows.Scan(&id, &nama, &musyrif, &uname, &tJuz, &tSurah, &targetAyat, &tDeadline, &memberCount)
		list = append(list, fiber.Map{"id": id, "nama": nama, "musyrif": musyrif, "ustadz_username": uname,
			"target_juz": tJuz, "target_surah": tSurah, "target_ayat": targetAyat, "target_deadline": tDeadline, "member_count": memberCount})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateHalaqoh(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama           string `json:"nama"`
		Musyrif        string `json:"musyrif"`
		UstadzUsername  string `json:"ustadz_username"`
		TargetJuz      string `json:"target_juz"`
		TargetSurah    string `json:"target_surah"`
		TargetAyat     int    `json:"target_ayat"`
		TargetDeadline string `json:"target_deadline"`
	}
	c.BodyParser(&body)
	if body.Nama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Nama halaqoh wajib"})
	}
	var deadline interface{} = nil
	if body.TargetDeadline != "" {
		deadline = body.TargetDeadline
	}
	res, err := config.DB.Exec(`INSERT INTO halaqoh (tenant_id, nama, musyrif, ustadz_username, target_juz, target_surah, target_ayat, target_deadline) VALUES (?,?,?,?,?,?,?,?)`,
		tid, body.Nama, body.Musyrif, body.UstadzUsername, body.TargetJuz, body.TargetSurah, body.TargetAyat, deadline)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama})
}

func UpdateHalaqoh(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	var body map[string]interface{}
	c.BodyParser(&body)
	allowed := map[string]string{"nama": "VARCHAR", "musyrif": "VARCHAR", "ustadz_username": "VARCHAR", "target_juz": "VARCHAR", "target_surah": "VARCHAR", "target_ayat": "INT", "target_deadline": "DATE"}
	for k, v := range body {
		if _, ok := allowed[k]; ok {
			config.DB.Exec(fmt.Sprintf("UPDATE halaqoh SET %s = ? WHERE id = ? AND tenant_id = ?", k), v, id, tid)
		}
	}
	return c.JSON(fiber.Map{"message": "Halaqoh diperbarui"})
}

func DeleteHalaqoh(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	config.DB.Exec("DELETE FROM tahfidz_nilai WHERE halaqoh_id = ? AND tenant_id = ?", id, tid)
	config.DB.Exec("DELETE FROM tahfidz_sesi WHERE halaqoh_id = ? AND tenant_id = ?", id, tid)
	config.DB.Exec("DELETE FROM halaqoh_members WHERE halaqoh_id = ? AND tenant_id = ?", id, tid)
	config.DB.Exec("DELETE FROM halaqoh WHERE id = ? AND tenant_id = ?", id, tid)
	return c.JSON(fiber.Map{"message": "Halaqoh dihapus"})
}

// ═══════════════════════════════════════════════════════════
// HALAQOH MEMBERS
// ═══════════════════════════════════════════════════════════

func GetHalaqohMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query(`SELECT hm.id, hm.santri_id, s.nama, COALESCE(s.kamar_id,0), COALESCE(k.nama,'')
		FROM halaqoh_members hm JOIN santri s ON hm.santri_id = s.id LEFT JOIN kamar k ON s.kamar_id = k.id
		WHERE hm.halaqoh_id = ? AND hm.tenant_id = ? AND hm.status = 'active' ORDER BY s.nama`, c.Params("id"), tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var hmid, sid, kid int
		var nama, kamar string
		rows.Scan(&hmid, &sid, &nama, &kid, &kamar)
		list = append(list, fiber.Map{"hm_id": hmid, "santri_id": sid, "santri_nama": nama, "kamar_nama": kamar})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func BulkAddHalaqohMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	hid := c.Params("id")
	var body struct {
		SantriIDs []int `json:"santri_ids"`
	}
	c.BodyParser(&body)
	if len(body.SantriIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "santri_ids wajib"})
	}
	count := 0
	for _, sid := range body.SantriIDs {
		var ex int
		config.DB.QueryRow("SELECT COUNT(*) FROM halaqoh_members WHERE halaqoh_id = ? AND santri_id = ? AND tenant_id = ? AND status = 'active'", hid, sid, tid).Scan(&ex)
		if ex > 0 {
			continue
		}
		_, err := config.DB.Exec("INSERT INTO halaqoh_members (tenant_id, santri_id, halaqoh_id, status) VALUES (?,?,?,?)", tid, sid, hid, "active")
		if err == nil {
			count++
		}
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d anggota ditambahkan", count)})
}

func RemoveHalaqohMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM halaqoh_members WHERE halaqoh_id = ? AND santri_id = ? AND tenant_id = ?", c.Params("id"), c.Params("santri_id"), tid)
	return c.JSON(fiber.Map{"message": "Anggota dihapus"})
}

// ═══════════════════════════════════════════════════════════
// TAHFIDZ BULK (Absensi + Penilaian sekaligus)
// ═══════════════════════════════════════════════════════════

func TahfidzBulk(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uname := c.Locals("username").(string)
	var body struct {
		Tanggal   string `json:"tanggal"`
		HalaqohID int    `json:"halaqoh_id"`
		Items     []struct {
			SantriID       int     `json:"santri_id"`
			Status         string  `json:"status"`
			JenisSetoran   string  `json:"jenis_setoran"`
			Surah          string  `json:"surah"`
			AyatDari       int     `json:"ayat_dari"`
			AyatSampai     int     `json:"ayat_sampai"`
			Juz            int     `json:"juz"`
			NilaiTajwid    int     `json:"nilai_tajwid"`
			NilaiKelancaran int    `json:"nilai_kelancaran"`
			NilaiMakhorijul int    `json:"nilai_makhorijul"`
			Catatan        string  `json:"catatan"`
		} `json:"items"`
	}
	c.BodyParser(&body)
	if body.Tanggal == "" || body.HalaqohID == 0 || len(body.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "tanggal, halaqoh_id, dan items wajib"})
	}

	// Upsert sesi
	recordedAt := helpers.ToDatetime()
	config.DB.Exec(`INSERT INTO tahfidz_sesi (tenant_id, halaqoh_id, tanggal, ustadz_username, recorded_at) VALUES (?,?,?,?,?)
		ON DUPLICATE KEY UPDATE ustadz_username = VALUES(ustadz_username), recorded_at = VALUES(recorded_at)`,
		tid, body.HalaqohID, body.Tanggal, uname, recordedAt)
	var sesiID int
	config.DB.QueryRow("SELECT id FROM tahfidz_sesi WHERE tenant_id = ? AND halaqoh_id = ? AND tanggal = ?", tid, body.HalaqohID, body.Tanggal).Scan(&sesiID)

	// Auto-izin
	izinMap := GetSantriIzinAktif(tid)
	count := 0
	for _, item := range body.Items {
		if item.SantriID == 0 {
			continue
		}
		st := item.Status
		if st == "" {
			st = "H"
		}
		if _, ok := izinMap[item.SantriID]; ok {
			st = "I"
		}

		var nilaiRata float64
		predikat := ""
		if st == "H" && (item.NilaiTajwid > 0 || item.NilaiKelancaran > 0 || item.NilaiMakhorijul > 0) {
			nilaiRata = float64(item.NilaiTajwid+item.NilaiKelancaran+item.NilaiMakhorijul) / 3.0
			nilaiRata = math.Round(nilaiRata*100) / 100
			predikat = getPredikat(nilaiRata)
		}

		jenisSetoran := item.JenisSetoran
		if jenisSetoran == "" {
			jenisSetoran = "ziyadah"
		}

		// Delete existing then insert (simpler than upsert for multiple columns)
		config.DB.Exec("DELETE FROM tahfidz_nilai WHERE tenant_id = ? AND sesi_id = ? AND santri_id = ?", tid, sesiID, item.SantriID)
		_, err := config.DB.Exec(`INSERT INTO tahfidz_nilai (tenant_id, sesi_id, santri_id, halaqoh_id, tanggal, status,
			jenis_setoran, surah, ayat_dari, ayat_sampai, juz, nilai_tajwid, nilai_kelancaran, nilai_makhorijul, nilai_rata, predikat, catatan)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			tid, sesiID, item.SantriID, body.HalaqohID, body.Tanggal, st,
			jenisSetoran, item.Surah, item.AyatDari, item.AyatSampai, item.Juz,
			item.NilaiTajwid, item.NilaiKelancaran, item.NilaiMakhorijul, nilaiRata, predikat, item.Catatan)
		if err == nil {
			count++
		}
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Tahfidz tersimpan (%d santri)", count), "recorded_at": recordedAt})
}

// ═══════════════════════════════════════════════════════════
// GET TAHFIDZ (per tanggal)
// ═══════════════════════════════════════════════════════════

func GetTahfidz(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tanggal := c.Query("tanggal")
	if tanggal == "" {
		tanggal = helpers.TodayWIB()
	}
	q := `SELECT tn.id, tn.santri_id, s.nama, tn.halaqoh_id, tn.status, tn.jenis_setoran, COALESCE(tn.surah,''),
		tn.ayat_dari, tn.ayat_sampai, tn.juz, tn.nilai_tajwid, tn.nilai_kelancaran, tn.nilai_makhorijul,
		tn.nilai_rata, COALESCE(tn.predikat,''), COALESCE(tn.catatan,'')
		FROM tahfidz_nilai tn JOIN santri s ON tn.santri_id = s.id WHERE tn.tenant_id = ? AND tn.tanggal = ?`
	args := []interface{}{tid, tanggal}
	if hid := c.Query("halaqoh_id"); hid != "" {
		q += " AND tn.halaqoh_id = ?"
		args = append(args, hid)
	}
	q += " ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, sid, hid, ayatDari, ayatSampai, juz, tajwid, kelancaran, makhorijul int
		var nama, status, jenis, surah, predikat, catatan string
		var nilaiRata float64
		rows.Scan(&id, &sid, &nama, &hid, &status, &jenis, &surah, &ayatDari, &ayatSampai, &juz,
			&tajwid, &kelancaran, &makhorijul, &nilaiRata, &predikat, &catatan)
		list = append(list, fiber.Map{"id": id, "santri_id": sid, "santri_nama": nama, "halaqoh_id": hid,
			"status": status, "jenis_setoran": jenis, "surah": surah, "ayat_dari": ayatDari, "ayat_sampai": ayatSampai,
			"juz": juz, "nilai_tajwid": tajwid, "nilai_kelancaran": kelancaran, "nilai_makhorijul": makhorijul,
			"nilai_rata": nilaiRata, "predikat": predikat, "catatan": catatan})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ═══════════════════════════════════════════════════════════
// REKAP TAHFIDZ
// ═══════════════════════════════════════════════════════════

func GetRekapTahfidz(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari")
	sampai := c.Query("sampai")
	halaqohID := c.Query("halaqoh_id")

	// Count sessions
	sesiQ := "SELECT COUNT(DISTINCT id) FROM tahfidz_sesi WHERE tenant_id = ?"
	sesiArgs := []interface{}{tid}
	if dari != "" {
		sesiQ += " AND tanggal >= ?"
		sesiArgs = append(sesiArgs, dari)
	}
	if sampai != "" {
		sesiQ += " AND tanggal <= ?"
		sesiArgs = append(sesiArgs, sampai)
	}
	if halaqohID != "" {
		sesiQ += " AND halaqoh_id = ?"
		sesiArgs = append(sesiArgs, halaqohID)
	}
	var jumlahSesi int
	config.DB.QueryRow(sesiQ, sesiArgs...).Scan(&jumlahSesi)

	// Get target info
	var targetAyat int
	var targetJuz, targetDeadline string
	if halaqohID != "" {
		config.DB.QueryRow("SELECT COALESCE(target_juz,''), target_ayat, COALESCE(target_deadline,'') FROM halaqoh WHERE id = ? AND tenant_id = ?", halaqohID, tid).Scan(&targetJuz, &targetAyat, &targetDeadline)
	}

	// Aggregate per santri
	q := `SELECT tn.santri_id, s.nama,
		SUM(CASE WHEN tn.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN tn.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN tn.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN tn.status='A' THEN 1 ELSE 0 END),
		COALESCE(AVG(CASE WHEN tn.status='H' AND tn.nilai_rata > 0 THEN tn.nilai_rata END), 0),
		COALESCE(SUM(CASE WHEN tn.status='H' AND tn.jenis_setoran='ziyadah' THEN GREATEST(tn.ayat_sampai - tn.ayat_dari + 1, 0) ELSE 0 END), 0)
		FROM tahfidz_nilai tn JOIN santri s ON tn.santri_id = s.id WHERE tn.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" {
		q += " AND tn.tanggal >= ?"
		args = append(args, dari)
	}
	if sampai != "" {
		q += " AND tn.tanggal <= ?"
		args = append(args, sampai)
	}
	if halaqohID != "" {
		q += " AND tn.halaqoh_id = ?"
		args = append(args, halaqohID)
	}
	q += " GROUP BY tn.santri_id, s.nama ORDER BY s.nama"

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var sid, h, i, s, a, capaianAyat int
		var nama string
		var avgNilai float64
		rows.Scan(&sid, &nama, &h, &i, &s, &a, &avgNilai, &capaianAyat)
		avgNilai = math.Round(avgNilai*100) / 100
		predikat := ""
		if avgNilai > 0 {
			predikat = getPredikat(avgNilai)
		}
		pct := 0.0
		if targetAyat > 0 {
			pct = math.Round(float64(capaianAyat) / float64(targetAyat) * 10000) / 100
		}
		list = append(list, fiber.Map{"santri_id": sid, "santri_nama": nama, "H": h, "I": i, "S": s, "A": a,
			"total": h + i + s + a, "avg_nilai": avgNilai, "predikat": predikat,
			"capaian_ayat": capaianAyat, "progress_pct": pct})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"jumlah_sesi": jumlahSesi, "target_ayat": targetAyat, "target_juz": targetJuz, "target_deadline": targetDeadline, "data": list})
}

// ═══════════════════════════════════════════════════════════
// PROGRESS INDIVIDUAL
// ═══════════════════════════════════════════════════════════

func GetProgressTahfidz(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("santri_id")

	rows, err := config.DB.Query(`SELECT tn.tanggal, tn.jenis_setoran, COALESCE(tn.surah,''), tn.ayat_dari, tn.ayat_sampai, tn.juz,
		tn.nilai_tajwid, tn.nilai_kelancaran, tn.nilai_makhorijul, tn.nilai_rata, COALESCE(tn.predikat,''), COALESCE(tn.catatan,''), tn.status
		FROM tahfidz_nilai tn WHERE tn.tenant_id = ? AND tn.santri_id = ? ORDER BY tn.tanggal DESC LIMIT 50`, tid, sid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var ayatDari, ayatSampai, juz, tajwid, kelancaran, makhorijul int
		var tanggal, jenis, surah, predikat, catatan, status string
		var nilaiRata float64
		rows.Scan(&tanggal, &jenis, &surah, &ayatDari, &ayatSampai, &juz,
			&tajwid, &kelancaran, &makhorijul, &nilaiRata, &predikat, &catatan, &status)
		list = append(list, fiber.Map{"tanggal": tanggal, "jenis_setoran": jenis, "surah": surah,
			"ayat_dari": ayatDari, "ayat_sampai": ayatSampai, "juz": juz,
			"nilai_tajwid": tajwid, "nilai_kelancaran": kelancaran, "nilai_makhorijul": makhorijul,
			"nilai_rata": nilaiRata, "predikat": predikat, "catatan": catatan, "status": status})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ═══════════════════════════════════════════════════════════
// EXPORT EXCEL
// ═══════════════════════════════════════════════════════════

func ExportTahfidzExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	dari := c.Query("dari")
	sampai := c.Query("sampai")
	halaqohID := c.Query("halaqoh_id")

	q := `SELECT tn.santri_id, s.nama,
		SUM(CASE WHEN tn.status='H' THEN 1 ELSE 0 END),
		SUM(CASE WHEN tn.status='I' THEN 1 ELSE 0 END),
		SUM(CASE WHEN tn.status='S' THEN 1 ELSE 0 END),
		SUM(CASE WHEN tn.status='A' THEN 1 ELSE 0 END),
		COALESCE(AVG(CASE WHEN tn.status='H' AND tn.nilai_rata > 0 THEN tn.nilai_rata END), 0),
		COALESCE(SUM(CASE WHEN tn.jenis_setoran='ziyadah' THEN GREATEST(tn.ayat_sampai - tn.ayat_dari + 1, 0) ELSE 0 END), 0)
		FROM tahfidz_nilai tn JOIN santri s ON tn.santri_id = s.id WHERE tn.tenant_id = ?`
	args := []interface{}{tid}
	if dari != "" {
		q += " AND tn.tanggal >= ?"
		args = append(args, dari)
	}
	if sampai != "" {
		q += " AND tn.tanggal <= ?"
		args = append(args, sampai)
	}
	if halaqohID != "" {
		q += " AND tn.halaqoh_id = ?"
		args = append(args, halaqohID)
	}
	q += " GROUP BY tn.santri_id, s.nama ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Rekap Tahfidz"
	f.SetSheetName("Sheet1", sheet)
	periode := dari + " s/d " + sampai
	f.SetCellValue(sheet, "A1", "REKAP TAHFIDZ QUR'AN")
	f.SetCellValue(sheet, "A2", "Periode: "+periode)
	r := 4
	hdrs := []string{"No", "Nama Santri", "Hadir", "Izin", "Sakit", "Alpa", "Capaian Ayat", "Rata-rata", "Predikat"}
	for i, h := range hdrs {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), h)
	}
	r++
	no := 1
	for rows.Next() {
		var sid, h, i, s, a, ayat int
		var nama string
		var avg float64
		rows.Scan(&sid, &nama, &h, &i, &s, &a, &avg, &ayat)
		avg = math.Round(avg*100) / 100
		pred := ""
		if avg > 0 {
			pred = getPredikat(avg)
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), no)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), h)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), i)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), s)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), a)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), ayat)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), avg)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", r), pred)
		no++
		r++
	}
	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=Rekap_Tahfidz.xlsx")
	return c.Send(buf.Bytes())
}
