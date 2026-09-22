package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetKamar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama, COALESCE(kapasitas,10), COALESCE(pengurus,'') FROM kamar WHERE tenant_id = ?", tid)
	if err != nil {
		fmt.Println("Error GetKamar:", err)
		return c.Status(500).JSON(fiber.Map{"message": "Database error: " + err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kap int; var nama, pengurus string
		rows.Scan(&id, &nama, &kap, &pengurus)
		var jml int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE kamar_id = ? AND status = 'aktif'", id).Scan(&jml)
		list = append(list, fiber.Map{"id": id, "nama": nama, "kapasitas": kap, "pengurus": pengurus, "jumlah_santri": jml})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateKamar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct{ Nama string `json:"nama"`; Kapasitas int `json:"kapasitas"`; Pengurus string `json:"pengurus"` }
	c.BodyParser(&body)
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	if body.Kapasitas == 0 { body.Kapasitas = 10 }
	res, _ := config.DB.Exec("INSERT INTO kamar (tenant_id, nama, kapasitas, pengurus) VALUES (?,?,?,?)", tid, body.Nama, body.Kapasitas, body.Pengurus)
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama, "kapasitas": body.Kapasitas, "pengurus": body.Pengurus})
}

func UpdateKamar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	var body struct{ Nama string `json:"nama"`; Kapasitas int `json:"kapasitas"`; Pengurus string `json:"pengurus"` }
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid input"})
	}
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	if body.Kapasitas == 0 { body.Kapasitas = 10 }
	_, err := config.DB.Exec("UPDATE kamar SET nama=?, kapasitas=?, pengurus=? WHERE id=? AND tenant_id=?", body.Nama, body.Kapasitas, body.Pengurus, id, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update kamar"})
	}
	return c.JSON(fiber.Map{"message": "Kamar diupdate"})
}

func DeleteKamar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM kamar WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Kamar dihapus"})
}

func GetKamarMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT id, nama, COALESCE(kelas_diniyyah,''), status FROM santri WHERE kamar_id = ? AND tenant_id = ? AND status = 'aktif'", c.Params("id"), tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() { var id int; var n, k, s string; rows.Scan(&id, &n, &k, &s); list = append(list, fiber.Map{"id": id, "nama": n, "kelas_diniyyah": k, "status": s}) }
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func BulkSetKamarMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kamarID := c.Params("id")
	var body struct{ SantriIDs []int `json:"santri_ids"` }
	c.BodyParser(&body)
	if len(body.SantriIDs) == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_ids wajib"}) }
	count := 0
	for _, sid := range body.SantriIDs {
		_, err := config.DB.Exec("UPDATE santri SET kamar_id = ? WHERE id = ? AND tenant_id = ?", kamarID, sid, tid)
		if err == nil { count++ }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d santri dipindahkan ke kamar", count)})
}
