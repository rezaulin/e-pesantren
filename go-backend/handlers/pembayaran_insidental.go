package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"github.com/gofiber/fiber/v2"
)

// ── MASTER JENIS TAGIHAN INSIDENTAL ──────────────────────────────────────────────

func GetJenisInsidental(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama, nominal, aktif FROM tagihan_insidental WHERE tenant_id = ? ORDER BY id", tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama string
		var nominal int64
		var aktif int
		rows.Scan(&id, &nama, &nominal, &aktif)
		list = append(list, fiber.Map{"id": id, "nama": nama, "nominal": nominal, "aktif": aktif})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateJenisInsidental(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama    string `json:"nama"`
		Nominal int64  `json:"nominal"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.Nominal <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nama dan nominal wajib diisi"})
	}
	res, err := config.DB.Exec("INSERT INTO tagihan_insidental (tenant_id, nama, nominal) VALUES (?,?,?)", tid, body.Nama, body.Nominal)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Jenis tagihan insidental ditambahkan"})
}

func UpdateJenisInsidental(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	var body struct {
		Nama    string `json:"nama"`
		Nominal int64  `json:"nominal"`
		Aktif   *int   `json:"aktif"`
	}
	c.BodyParser(&body)
	if body.Nama != "" {
		config.DB.Exec("UPDATE tagihan_insidental SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, id, tid)
	}
	if body.Nominal > 0 {
		config.DB.Exec("UPDATE tagihan_insidental SET nominal = ? WHERE id = ? AND tenant_id = ?", body.Nominal, id, tid)
	}
	if body.Aktif != nil {
		config.DB.Exec("UPDATE tagihan_insidental SET aktif = ? WHERE id = ? AND tenant_id = ?", *body.Aktif, id, tid)
	}
	return c.JSON(fiber.Map{"message": "Diupdate"})
}

// ── TAGIHAN SANTRI ─────────────────────────────────────────────────────────────

func GetTagihanInsidentalSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	santriID := c.Params("santri_id")
	
	// Check wali ownership
	role := c.Locals("role").(string)
	if role == "wali" {
		uid := c.Locals("user_id").(int)
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", santriID, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}
	
	q := `SELECT t.id, m.nama, t.nominal, t.sisa_tagihan, t.status 
	      FROM tagihan_insidental_santri t
		  JOIN tagihan_insidental m ON t.tagihan_id = m.id
		  WHERE t.tenant_id = ? AND t.santri_id = ? ORDER BY t.id DESC`
		  
	rows, err := config.DB.Query(q, tid, santriID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama, status string
		var nom, sisa int64
		rows.Scan(&id, &nama, &nom, &sisa, &status)
		list = append(list, fiber.Map{"id": id, "nama": nama, "nominal": nom, "sisa_tagihan": sisa, "status": status})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func GetRekapInsidental(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tagihanID := c.Query("tagihan_id")
	if tagihanID == "" {
		return c.Status(400).JSON(fiber.Map{"message": "tagihan_id diperlukan"})
	}

	q := `SELECT s.nama, COALESCE(k.nama, '-') as kamar_nama, t.nominal, (t.nominal - t.sisa_tagihan) as nominal_dibayar, t.sisa_tagihan as sisa, t.status
	      FROM tagihan_insidental_santri t
		  JOIN santri s ON t.santri_id = s.id
		  LEFT JOIN kamar k ON s.kamar_id = k.id
		  WHERE t.tenant_id = ? AND t.tagihan_id = ? 
		  ORDER BY s.nama ASC`
		  
	rows, err := config.DB.Query(q, tid, tagihanID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	
	var list []fiber.Map
	for rows.Next() {
		var nama, kamar, status string
		var nom, dibayar, sisa int64
		rows.Scan(&nama, &kamar, &nom, &dibayar, &sisa, &status)
		list = append(list, fiber.Map{
			"nama": nama, 
			"kamar_nama": kamar, 
			"nominal": nom, 
			"nominal_dibayar": dibayar, 
			"sisa": sisa, 
			"status": status,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func SetTagihanInsidentalSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		SantriIDs []int `json:"santri_ids"`
		TagihanID int   `json:"tagihan_id"`
		Nominal   int64 `json:"nominal"`
	}
	c.BodyParser(&body)
	
	if len(body.SantriIDs) == 0 || body.TagihanID == 0 || body.Nominal <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak lengkap"})
	}

	for _, sid := range body.SantriIDs {
		// Cek apakah sudah ada tagihan ini untuk santri ini
		var count int
		config.DB.QueryRow("SELECT COUNT(*) FROM tagihan_insidental_santri WHERE tenant_id = ? AND santri_id = ? AND tagihan_id = ?", tid, sid, body.TagihanID).Scan(&count)
		if count == 0 {
			config.DB.Exec("INSERT INTO tagihan_insidental_santri (tenant_id, santri_id, tagihan_id, nominal, sisa_tagihan) VALUES (?,?,?,?,?)",
				tid, sid, body.TagihanID, body.Nominal, body.Nominal)
		}
	}

	return c.JSON(fiber.Map{"message": "Berhasil menetapkan tagihan ke santri"})
}

func SetTagihanInsidentalBulk(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	
	type SantriTagihan struct {
		SantriID int   `json:"santri_id"`
		Nominal  int64 `json:"nominal"`
	}
	
	var body struct {
		TagihanID  int             `json:"tagihan_id"`
		SantriList []SantriTagihan `json:"santri_list"`
	}
	
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	
	if body.TagihanID == 0 || len(body.SantriList) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak lengkap"})
	}
	
	tx, err := config.DB.Begin()
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer tx.Rollback()
	
	for _, item := range body.SantriList {
		if item.Nominal <= 0 { continue }
		var count int
		tx.QueryRow("SELECT COUNT(*) FROM tagihan_insidental_santri WHERE tenant_id = ? AND santri_id = ? AND tagihan_id = ?", tid, item.SantriID, body.TagihanID).Scan(&count)
		if count == 0 {
			_, err = tx.Exec("INSERT INTO tagihan_insidental_santri (tenant_id, santri_id, tagihan_id, nominal, sisa_tagihan) VALUES (?,?,?,?,?)",
				tid, item.SantriID, body.TagihanID, item.Nominal, item.Nominal)
			if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
		}
	}
	tx.Commit()
	
	return c.JSON(fiber.Map{"message": "Berhasil menetapkan tagihan massal"})
}

// ── PROSES PEMBAYARAN ──────────────────────────────────────────────────────────

func BayarInsidental(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	
	var body struct {
		TagihanSantriID int    `json:"tagihan_santri_id"`
		SantriID        int    `json:"santri_id"`
		NominalDibayar  int64  `json:"nominal_dibayar"`
		Metode          string `json:"metode"` // Tunai, Transfer, dll
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	
	if body.NominalDibayar <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nominal tidak valid"})
	}

	// 1. Cek Sisa Tagihan
	var sisa int64
	var namaTagihan string
	err := config.DB.QueryRow(`
		SELECT t.sisa_tagihan, m.nama 
		FROM tagihan_insidental_santri t
		JOIN tagihan_insidental m ON t.tagihan_id = m.id
		WHERE t.id = ? AND t.tenant_id = ?`, body.TagihanSantriID, tid).Scan(&sisa, &namaTagihan)
		
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Tagihan tidak ditemukan"})
	}
	
	if body.NominalDibayar > sisa {
		return c.Status(400).JSON(fiber.Map{"message": "Nominal dibayar melebihi sisa tagihan"})
	}

	// 2. Hitung Sisa Baru
	sisaBaru := sisa - body.NominalDibayar
	status := "Belum Lunas"
	if sisaBaru == 0 {
		status = "Lunas"
	}

	// Gunakan TX untuk konsistensi
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer tx.Rollback()

	// 3. Update Tagihan Santri
	_, err = tx.Exec("UPDATE tagihan_insidental_santri SET sisa_tagihan = ?, status = ? WHERE id = ?", sisaBaru, status, body.TagihanSantriID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	// 4. Insert Riwayat Pembayaran Insidental
	_, err = tx.Exec("INSERT INTO pembayaran_insidental (tenant_id, santri_id, tagihan_santri_id, nominal_dibayar, metode, created_by) VALUES (?,?,?,?,?,?)",
		tid, body.SantriID, body.TagihanSantriID, body.NominalDibayar, body.Metode, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	// 5. Insert ke Buku Kas / Catatan Keuangan
	var santriNama string
	tx.QueryRow("SELECT nama FROM santri WHERE id = ? AND tenant_id = ?", body.SantriID, tid).Scan(&santriNama)
	
	ket := fmt.Sprintf("Tagihan %s a.n. %s", namaTagihan, santriNama)
	
	_, err = tx.Exec("INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?,?,?,?,?,?)",
		tid, "masuk", body.NominalDibayar, ket, helpers.TodayWIB(), uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	tx.Commit()

	return c.JSON(fiber.Map{"message": "Pembayaran berhasil diproses"})
}

func GetRiwayatInsidentalSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	santriID := c.Params("santri_id")
	
	// Check wali ownership
	role := c.Locals("role").(string)
	if role == "wali" {
		uid := c.Locals("user_id").(int)
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", santriID, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}
	
	rows, err := config.DB.Query(`
		SELECT p.id, m.nama, p.nominal_dibayar, p.tanggal, p.metode
		FROM pembayaran_insidental p
		JOIN tagihan_insidental_santri t ON p.tagihan_santri_id = t.id
		JOIN tagihan_insidental m ON t.tagihan_id = m.id
		WHERE p.tenant_id = ? AND p.santri_id = ? ORDER BY p.id DESC`, tid, santriID)
		
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama, tgl, metode string
		var nom int64
		rows.Scan(&id, &nama, &nom, &tgl, &metode)
		list = append(list, fiber.Map{"id":id, "nama":nama, "nominal_dibayar":nom, "tanggal":tgl, "metode":metode})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}
