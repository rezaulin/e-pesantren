package handlers

import (
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT id, nama, COALESCE(kategori,'tambahan') FROM kegiatan WHERE tenant_id = ?", tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() { var id int; var n, k string; rows.Scan(&id, &n, &k); list = append(list, fiber.Map{"id": id, "nama": n, "kategori": k}) }
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct{ Nama string `json:"nama"`; Kategori string `json:"kategori"` }
	c.BodyParser(&body)
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	if body.Kategori == "" { body.Kategori = "tambahan" }
	res, _ := config.DB.Exec("INSERT INTO kegiatan (tenant_id, nama, kategori) VALUES (?,?,?)", tid, body.Nama, body.Kategori)
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama, "kategori": body.Kategori})
}

func DeleteKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM kegiatan WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Kegiatan dihapus"})
}

func UpdateKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama     string `json:"nama"`
		Kategori string `json:"kategori"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}
	if body.Nama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"})
	}
	if body.Kategori == "" {
		body.Kategori = "tambahan"
	}
	config.DB.Exec("UPDATE kegiatan SET nama = ?, kategori = ? WHERE id = ? AND tenant_id = ?",
		body.Nama, body.Kategori, c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Kegiatan diperbarui", "nama": body.Nama, "kategori": body.Kategori})
}
