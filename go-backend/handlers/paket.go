package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ── List Paket ───────────────────────────────────────────
func GetPaket(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	status := c.Query("status") // DI_GERBANG, DI_ASRAMA, SELESAI
	search := c.Query("search")
	kamarID := c.Query("kamar_id")

	q := `SELECT p.id, p.santri_id, COALESCE(s.nama,''), COALESCE(k.nama,''),
		COALESCE(p.pengirim,''), COALESCE(p.keterangan,''), p.status,
		COALESCE(p.posisi_rak,''),
		COALESCE(DATE_FORMAT(p.waktu_gerbang,'%d-%m-%Y %H:%i'),''),
		COALESCE(DATE_FORMAT(p.waktu_asrama,'%d-%m-%Y %H:%i'),''),
		COALESCE(DATE_FORMAT(p.waktu_diterima,'%d-%m-%Y %H:%i'),''),
		COALESCE(DATE_FORMAT(p.created_at,'%d-%m-%Y %H:%i'),'')
		FROM paket p
		LEFT JOIN santri s ON p.santri_id = s.id
		LEFT JOIN kamar k ON s.kamar_id = k.id
		WHERE p.tenant_id = ?`
	args := []interface{}{tid}

	if status != "" {
		q += " AND p.status = ?"
		args = append(args, status)
	}
	if search != "" {
		q += " AND (s.nama LIKE ? OR p.pengirim LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	if kamarID != "" {
		q += " AND s.kamar_id = ?"
		args = append(args, kamarID)
	}

	q += " ORDER BY p.created_at DESC"

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, sid int
		var nama, kamarNama, pengirim, ket, st, rak, wGerbang, wAsrama, wDiterima, createdAt string
		rows.Scan(&id, &sid, &nama, &kamarNama, &pengirim, &ket, &st, &rak,
			&wGerbang, &wAsrama, &wDiterima, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "santri_id": sid, "santri_nama": nama,
			"kamar_nama": kamarNama, "pengirim": pengirim,
			"keterangan": ket, "status": st, "posisi_rak": rak,
			"waktu_gerbang": wGerbang, "waktu_asrama": wAsrama,
			"waktu_diterima": wDiterima, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ── Create Paket (Input di Gerbang) ─────────────────────
func CreatePaket(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		SantriID   int    `json:"santri_id"`
		Pengirim   string `json:"pengirim"`
		Keterangan string `json:"keterangan"`
	}
	c.BodyParser(&body)
	if body.SantriID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Santri wajib dipilih"})
	}

	now := helpers.ToDatetime()
	res, err := config.DB.Exec(
		`INSERT INTO paket (tenant_id, santri_id, pengirim, keterangan, status, waktu_gerbang, created_by, created_at)
		VALUES (?,?,?,?,'DI_GERBANG',?,?,NOW())`,
		tid, body.SantriID, body.Pengirim, body.Keterangan, now, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan: " + err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "📦 Paket berhasil dicatat"})
}

// ── Bulk Transfer ke Asrama (Checklist Massal) ──────────
func BulkTransferPaket(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		IDs      []int  `json:"ids"`
		PosisiRak string `json:"posisi_rak"`
	}
	c.BodyParser(&body)
	if len(body.IDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Pilih minimal 1 paket"})
	}

	now := helpers.ToDatetime()
	placeholders := make([]string, len(body.IDs))
	args := []interface{}{now, body.PosisiRak}
	for i, id := range body.IDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, tid)

	q := fmt.Sprintf(
		`UPDATE paket SET status = 'DI_ASRAMA', waktu_asrama = ?, posisi_rak = ?
		WHERE id IN (%s) AND tenant_id = ? AND status = 'DI_GERBANG'`,
		strings.Join(placeholders, ","))

	res, err := config.DB.Exec(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	affected, _ := res.RowsAffected()
	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("✅ %d paket dipindahkan ke asrama", affected),
		"count":   affected,
	})
}

// ── Serahkan Paket ke Santri ────────────────────────────
func SerahkanPaket(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	pid, _ := strconv.Atoi(c.Params("id"))
	now := helpers.ToDatetime()

	_, err := config.DB.Exec(
		`UPDATE paket SET status = 'SELESAI', waktu_diterima = ?
		WHERE id = ? AND tenant_id = ? AND status = 'DI_ASRAMA'`,
		now, pid, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "✅ Paket diserahkan ke santri"})
}

// ── Update Posisi Rak ───────────────────────────────────
func UpdatePaketRak(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	pid, _ := strconv.Atoi(c.Params("id"))
	var body struct {
		PosisiRak string `json:"posisi_rak"`
	}
	c.BodyParser(&body)

	_, err := config.DB.Exec(
		"UPDATE paket SET posisi_rak = ? WHERE id = ? AND tenant_id = ?",
		body.PosisiRak, pid, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Posisi rak diperbarui"})
}

// ── Delete Paket ────────────────────────────────────────
func DeletePaket(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM paket WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Paket dihapus"})
}

// ── Paket Stats ─────────────────────────────────────────
func GetPaketStats(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	today := helpers.TodayWIB()

	var diGerbang, diAsrama, selesai, totalHariIni int
	config.DB.QueryRow("SELECT COUNT(*) FROM paket WHERE tenant_id = ? AND status = 'DI_GERBANG'", tid).Scan(&diGerbang)
	config.DB.QueryRow("SELECT COUNT(*) FROM paket WHERE tenant_id = ? AND status = 'DI_ASRAMA'", tid).Scan(&diAsrama)
	config.DB.QueryRow("SELECT COUNT(*) FROM paket WHERE tenant_id = ? AND status = 'SELESAI' AND DATE(waktu_diterima) = ?", tid, today).Scan(&selesai)
	config.DB.QueryRow("SELECT COUNT(*) FROM paket WHERE tenant_id = ? AND DATE(created_at) = ?", tid, today).Scan(&totalHariIni)

	return c.JSON(fiber.Map{
		"di_gerbang":    diGerbang,
		"di_asrama":     diAsrama,
		"selesai_today": selesai,
		"total_today":   totalHariIni,
	})
}
