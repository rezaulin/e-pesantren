package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetKelasSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama FROM kelas_sekolah WHERE tenant_id = ? ORDER BY nama", tid)
	if err != nil {
		fmt.Println("❌ GetKelasSekolah error:", err)
		return c.Status(500).JSON(fiber.Map{"message": "DB error: " + err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int; var n string
		rows.Scan(&id, &n)
		var jml int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelas WHERE kelas_id = ? AND tenant_id = ? AND status = 'active'", id, tid).Scan(&jml)
		list = append(list, fiber.Map{"id": id, "nama": n, "jumlah_santri": jml})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateKelasSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct { Nama string `json:"nama"` }
	c.BodyParser(&body)
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	res, err := config.DB.Exec("INSERT INTO kelas_sekolah (tenant_id, nama) VALUES (?,?)", tid, body.Nama)
	if err != nil {
		fmt.Println("❌ CreateKelasSekolah error:", err)
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat kelas: " + err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama})
}

func UpdateKelasSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct { Nama string `json:"nama"` }
	c.BodyParser(&body)
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	_, err := config.DB.Exec("UPDATE kelas_sekolah SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, c.Params("id"), tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update: " + err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Kelas diperbarui"})
}

func DeleteKelasSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM santri_kelas WHERE kelas_id = ? AND tenant_id = ?", c.Params("id"), tid)
	config.DB.Exec("DELETE FROM jadwal_sekolah WHERE kelas_id = ? AND tenant_id = ?", c.Params("id"), tid)
	config.DB.Exec("DELETE FROM kelas_sekolah WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}

func GetKelasMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kelasID := c.Params("id")
	rows, err := config.DB.Query(
		`SELECT s.id, s.nama, COALESCE(s.kelas_diniyyah,''), s.status, sk.id as sk_id 
		 FROM santri_kelas sk 
		 JOIN santri s ON sk.santri_id = s.id 
		 WHERE sk.kelas_id = ? AND sk.tenant_id = ? AND sk.status = 'active' 
		 ORDER BY s.nama`,
		kelasID, tid)
	if err != nil {
		fmt.Println("❌ GetKelasMembers error:", err)
		return c.Status(500).JSON(fiber.Map{"message": "DB error: " + err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, skID int; var nama, kelas, status string
		rows.Scan(&id, &nama, &kelas, &status, &skID)
		list = append(list, fiber.Map{"id": id, "nama": nama, "kelas_diniyyah": kelas, "status": status, "santri_kelas_id": skID})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func AddKelasMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kelasID := c.Params("id")
	var body struct { SantriID int `json:"santri_id"` }
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid body: " + err.Error()})
	}
	if body.SantriID == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_id wajib"}) }

	// Check duplicate
	var ex int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelas WHERE kelas_id = ? AND santri_id = ? AND tenant_id = ? AND status = 'active'", kelasID, body.SantriID, tid).Scan(&ex)
	if err != nil {
		fmt.Println("❌ AddKelasMember check error:", err)
		return c.Status(500).JSON(fiber.Map{"message": "DB error saat cek duplikat: " + err.Error()})
	}
	if ex > 0 { return c.Status(400).JSON(fiber.Map{"message": "Santri sudah ada di kelas ini"}) }

	// Insert
	_, err = config.DB.Exec("INSERT INTO santri_kelas (tenant_id, santri_id, kelas_id, status) VALUES (?,?,?,?)", tid, body.SantriID, kelasID, "active")
	if err != nil {
		fmt.Println("❌ AddKelasMember insert error:", err)
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menambah santri: " + err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Santri ditambahkan ke kelas"})
}

func RemoveKelasMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	_, err := config.DB.Exec("DELETE FROM santri_kelas WHERE kelas_id = ? AND santri_id = ? AND tenant_id = ?", c.Params("id"), c.Params("santri_id"), tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal hapus: " + err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Santri dihapus dari kelas"})
}

func BulkAddKelasMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kelasID := c.Params("id")
	var body struct{ SantriIDs []int `json:"santri_ids"` }
	c.BodyParser(&body)
	if len(body.SantriIDs) == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_ids wajib"}) }
	count := 0
	for _, sid := range body.SantriIDs {
		var ex int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelas WHERE kelas_id = ? AND santri_id = ? AND tenant_id = ? AND status = 'active'", kelasID, sid, tid).Scan(&ex)
		if ex > 0 { continue }
		_, err := config.DB.Exec("INSERT INTO santri_kelas (tenant_id, santri_id, kelas_id, status) VALUES (?,?,?,?)", tid, sid, kelasID, "active")
		if err == nil { count++ }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d santri ditambahkan", count)})
}
