package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ── List Perizinan ───────────────────────────────────────
func GetPerizinan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	status := c.Query("status")   // aktif, selesai, terlambat
	santriID := c.Query("santri_id")

	q := `SELECT p.id, p.santri_id, COALESCE(s.nama,''), COALESCE(p.keterangan,''),
		p.tipe_durasi, p.durasi, p.tanggal_mulai, p.tanggal_selesai, p.status,
		p.terlambat_durasi, COALESCE(p.terlambat_tipe,'hari'),
		COALESCE(p.tanggal_kembali,''), p.created_at
		FROM perizinan p
		LEFT JOIN santri s ON p.santri_id = s.id
		WHERE p.tenant_id = ?`
	args := []interface{}{tid}
	if status != "" {
		q += " AND p.status = ?"
		args = append(args, status)
	}
	if santriID != "" {
		q += " AND p.santri_id = ?"
		args = append(args, santriID)
	}
	q += " ORDER BY p.created_at DESC"

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, sid, durasi, terlambatDurasi int
		var nama, ket, tipeDurasi, mulai, selesai, st, terlambatTipe, kembali, createdAt string
		rows.Scan(&id, &sid, &nama, &ket, &tipeDurasi, &durasi, &mulai, &selesai, &st,
			&terlambatDurasi, &terlambatTipe, &kembali, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "santri_id": sid, "santri_nama": nama,
			"keterangan": ket, "tipe_durasi": tipeDurasi, "durasi": durasi,
			"tanggal_mulai": mulai, "tanggal_selesai": selesai, "status": st,
			"terlambat_durasi": terlambatDurasi, "terlambat_tipe": terlambatTipe,
			"tanggal_kembali": kembali, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ── Get Active Perizinan (santri yang sedang izin saat ini) ──
func GetPerizinanAktif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	now := helpers.ToDatetime()
	rows, err := config.DB.Query(
		`SELECT p.santri_id, COALESCE(s.nama,''), COALESCE(p.keterangan,''), p.tanggal_selesai
		FROM perizinan p LEFT JOIN santri s ON p.santri_id = s.id
		WHERE p.tenant_id = ? AND p.status = 'aktif' AND ? BETWEEN p.tanggal_mulai AND p.tanggal_selesai`,
		tid, now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var sid int
		var nama, ket, selesai string
		rows.Scan(&sid, &nama, &ket, &selesai)
		list = append(list, fiber.Map{"santri_id": sid, "santri_nama": nama, "keterangan": ket, "tanggal_selesai": selesai})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ── Create Perizinan ─────────────────────────────────────
func CreatePerizinan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		SantriID   int    `json:"santri_id"`
		Keterangan string `json:"keterangan"`
		TipeDurasi string `json:"tipe_durasi"` // "hari" or "jam"
		Durasi     int    `json:"durasi"`
		TglMulai   string `json:"tanggal_mulai"` // "2006-01-02 15:04" or "2006-01-02"
	}
	c.BodyParser(&body)
	if body.SantriID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Santri wajib dipilih"})
	}
	if body.Durasi <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Durasi harus lebih dari 0"})
	}
	if body.TipeDurasi == "" {
		body.TipeDurasi = "hari"
	}

	// Parse start time
	var mulai time.Time
	var err error
	if len(body.TglMulai) > 10 {
		mulai, err = time.ParseInLocation("2006-01-02 15:04", body.TglMulai, helpers.WIB)
	} else if body.TglMulai != "" {
		mulai, err = time.ParseInLocation("2006-01-02", body.TglMulai, helpers.WIB)
	} else {
		mulai = helpers.NowWIB()
	}
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format tanggal tidak valid"})
	}

	// Calculate end time
	var selesai time.Time
	if body.TipeDurasi == "jam" {
		selesai = mulai.Add(time.Duration(body.Durasi) * time.Hour)
	} else {
		selesai = mulai.AddDate(0, 0, body.Durasi)
	}

	mulaiStr := mulai.Format("2006-01-02 15:04:05")
	selesaiStr := selesai.Format("2006-01-02 15:04:05")

	res, err := config.DB.Exec(
		`INSERT INTO perizinan (tenant_id, santri_id, keterangan, tipe_durasi, durasi, tanggal_mulai, tanggal_selesai, status, created_by) 
		VALUES (?,?,?,?,?,?,?,'aktif',?)`,
		tid, body.SantriID, body.Keterangan, body.TipeDurasi, body.Durasi, mulaiStr, selesaiStr, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan: " + err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Perizinan berhasil dibuat", "tanggal_selesai": selesaiStr})
}

// ── Tandai Sudah Kembali (tepat waktu) ───────────────────
func PerizinanKembali(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	pid, _ := strconv.Atoi(c.Params("id"))
	now := helpers.ToDatetime()

	_, err := config.DB.Exec(
		`UPDATE perizinan SET status = 'selesai', tanggal_kembali = ? WHERE id = ? AND tenant_id = ? AND status = 'aktif'`,
		now, pid, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Santri telah kembali"})
}

// ── Tandai Terlambat + Auto Pelanggaran ──────────────────
func PerizinanTerlambat(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	pid, _ := strconv.Atoi(c.Params("id"))
	var body struct {
		TerlambatDurasi int    `json:"terlambat_durasi"`
		TerlambatTipe   string `json:"terlambat_tipe"` // "hari" or "jam"
	}
	c.BodyParser(&body)
	if body.TerlambatDurasi <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Durasi terlambat harus lebih dari 0"})
	}
	if body.TerlambatTipe == "" {
		body.TerlambatTipe = "hari"
	}

	now := helpers.ToDatetime()
	tanggalHariIni := helpers.TodayWIB()

	// Get perizinan info for pelanggaran description
	var santriID int
	var keterangan string
	err := config.DB.QueryRow(
		"SELECT santri_id, COALESCE(keterangan,'') FROM perizinan WHERE id = ? AND tenant_id = ?",
		pid, tid).Scan(&santriID, &keterangan)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Perizinan tidak ditemukan"})
	}

	// Update perizinan status
	_, err = config.DB.Exec(
		`UPDATE perizinan SET status = 'terlambat', terlambat_durasi = ?, terlambat_tipe = ?, tanggal_kembali = ? 
		WHERE id = ? AND tenant_id = ?`,
		body.TerlambatDurasi, body.TerlambatTipe, now, pid, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	// Auto-insert pelanggaran (1 poin per insiden)
	deskripsi := fmt.Sprintf("Terlambat %d %s dari izin: %s", body.TerlambatDurasi, body.TerlambatTipe, keterangan)
	_, err = config.DB.Exec(
		"INSERT INTO pelanggaran (tenant_id, santri_id, jenis, deskripsi, poin, tanggal, created_by) VALUES (?,?,?,?,?,?,?)",
		tid, santriID, "Terlambat kembali dari izin", deskripsi, 1, tanggalHariIni, uid)
	if err != nil {
		fmt.Println("Auto-pelanggaran insert error:", err)
	}

	return c.JSON(fiber.Map{"message": "Santri ditandai terlambat & pelanggaran tercatat"})
}

// ── Delete Perizinan ─────────────────────────────────────
func DeletePerizinan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM perizinan WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Perizinan dihapus"})
}

// ── Helper: Get santri IDs yang sedang izin aktif ────────
// Used by absensi handlers to auto-mark as "I"
func GetSantriIzinAktif(tenantID int) map[int]string {
	result := make(map[int]string)
	now := helpers.ToDatetime()
	rows, err := config.DB.Query(
		`SELECT santri_id, COALESCE(keterangan,'Izin resmi') FROM perizinan 
		WHERE tenant_id = ? AND status = 'aktif' AND ? BETWEEN tanggal_mulai AND tanggal_selesai`,
		tenantID, now)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var sid int
		var ket string
		rows.Scan(&sid, &ket)
		result[sid] = ket
	}
	return result
}
