package handlers

import (
	"database/sql"
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetPelanggaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query(`SELECT p.id, p.santri_id, COALESCE(s.nama,''), COALESCE(p.jenis,''),
		COALESCE(p.deskripsi,''), COALESCE(p.poin,0), COALESCE(p.tanggal,''),
		COALESCE(p.status_takzir,'belum'), COALESCE(p.jenis_takzir,''), COALESCE(p.denda,0),
		p.takzir_at
		FROM pelanggaran p LEFT JOIN santri s ON p.santri_id = s.id
		WHERE p.tenant_id = ? ORDER BY p.tanggal DESC`, tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, sid, poin int
		var denda int64
		var sn, j, d, t, statusTakzir, jenisTakzir string
		var takzirAt sql.NullTime
		rows.Scan(&id, &sid, &sn, &j, &d, &poin, &t, &statusTakzir, &jenisTakzir, &denda, &takzirAt)
		ta := ""
		if takzirAt.Valid {
			ta = takzirAt.Time.Format("2006-01-02 15:04")
		}
		list = append(list, fiber.Map{
			"id": id, "santri_id": sid, "santri_nama": sn,
			"jenis": j, "deskripsi": d, "poin": poin, "tanggal": t,
			"status_takzir": statusTakzir, "jenis_takzir": jenisTakzir,
			"denda": denda, "takzir_at": ta,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreatePelanggaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int); uid := c.Locals("user_id").(int)
	var body struct {
		SantriID int    `json:"santri_id"`
		Jenis    string `json:"jenis"`
		Deskripsi string `json:"deskripsi"`
		Poin     int    `json:"poin"`
		Tanggal  string `json:"tanggal"`
	}
	c.BodyParser(&body)
	if body.SantriID == 0 { return c.Status(400).JSON(fiber.Map{"message": "Santri wajib dipilih"}) }
	res, _ := config.DB.Exec(`INSERT INTO pelanggaran (tenant_id, santri_id, jenis, deskripsi, poin, tanggal, status_takzir, created_by)
		VALUES (?,?,?,?,?,?,'belum',?)`,
		tid, body.SantriID, body.Jenis, body.Deskripsi, body.Poin, body.Tanggal, uid)
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Pelanggaran dicatat"})
}

func UpdateTakzir(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	pelID := c.Params("id")
	var body struct {
		StatusTakzir string `json:"status_takzir"`
		JenisTakzir  string `json:"jenis_takzir"`
		Denda        int64  `json:"denda"`
	}
	c.BodyParser(&body)
	if body.StatusTakzir == "" { body.StatusTakzir = "sudah" }

	if body.StatusTakzir == "sudah" {
		config.DB.Exec(`UPDATE pelanggaran SET status_takzir=?, jenis_takzir=?, denda=?, takzir_at=NOW()
			WHERE id=? AND tenant_id=?`,
			body.StatusTakzir, body.JenisTakzir, body.Denda, pelID, tid)
	} else {
		// Reset takzir
		config.DB.Exec(`UPDATE pelanggaran SET status_takzir='belum', jenis_takzir='', denda=0, takzir_at=NULL
			WHERE id=? AND tenant_id=?`, pelID, tid)
	}
	return c.JSON(fiber.Map{"message": "Takzir diperbarui"})
}

func GetRekapTakzir(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	santriID := c.QueryInt("santri_id", 0)

	// If santri_id provided, return detail for that santri
	if santriID > 0 {
		rows, _ := config.DB.Query(`SELECT p.id, COALESCE(p.jenis,''), COALESCE(p.deskripsi,''),
			COALESCE(p.poin,0), COALESCE(p.tanggal,''),
			COALESCE(p.status_takzir,'belum'), COALESCE(p.jenis_takzir,''),
			COALESCE(p.denda,0), p.takzir_at
			FROM pelanggaran p WHERE p.tenant_id=? AND p.santri_id=?
			ORDER BY p.tanggal DESC`, tid, santriID)
		defer rows.Close()
		var detail []fiber.Map
		for rows.Next() {
			var id, poin int
			var denda int64
			var jenis, desk, tgl, st, jt string
			var takzirAt sql.NullTime
			rows.Scan(&id, &jenis, &desk, &poin, &tgl, &st, &jt, &denda, &takzirAt)
			ta := ""
			if takzirAt.Valid { ta = takzirAt.Time.Format("2006-01-02 15:04") }
			detail = append(detail, fiber.Map{
				"id": id, "jenis": jenis, "deskripsi": desk, "poin": poin,
				"tanggal": tgl, "status_takzir": st, "jenis_takzir": jt,
				"denda": denda, "takzir_at": ta,
			})
		}
		if detail == nil { detail = []fiber.Map{} }

		// Get santri name
		var nama string
		config.DB.QueryRow("SELECT COALESCE(nama,'') FROM santri WHERE id=?", santriID).Scan(&nama)

		// Total denda
		var totalDenda int64
		config.DB.QueryRow(`SELECT COALESCE(SUM(denda),0) FROM pelanggaran
			WHERE tenant_id=? AND santri_id=? AND status_takzir='sudah'`, tid, santriID).Scan(&totalDenda)

		return c.JSON(fiber.Map{
			"santri_id": santriID, "santri_nama": nama,
			"total_denda": totalDenda, "detail": detail,
		})
	}

	// Summary per santri
	rows, _ := config.DB.Query(`SELECT p.santri_id, COALESCE(s.nama,''),
		COUNT(*) as total,
		SUM(CASE WHEN p.status_takzir='sudah' THEN 1 ELSE 0 END) as sudah,
		SUM(CASE WHEN p.status_takzir='belum' OR p.status_takzir IS NULL THEN 1 ELSE 0 END) as belum,
		COALESCE(SUM(p.poin),0) as total_poin,
		COALESCE(SUM(CASE WHEN p.status_takzir='sudah' THEN p.denda ELSE 0 END),0) as total_denda
		FROM pelanggaran p LEFT JOIN santri s ON p.santri_id = s.id
		WHERE p.tenant_id=?
		GROUP BY p.santri_id, s.nama
		ORDER BY total DESC`, tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var sid, total, sudah, belum, totalPoin int
		var totalDenda int64
		var nama string
		rows.Scan(&sid, &nama, &total, &sudah, &belum, &totalPoin, &totalDenda)
		list = append(list, fiber.Map{
			"santri_id": sid, "santri_nama": nama,
			"total": total, "sudah_takzir": sudah, "belum_takzir": belum,
			"total_poin": totalPoin, "total_denda": totalDenda,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func DeletePelanggaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pelanggaran WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}
