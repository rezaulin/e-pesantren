package handlers

import (
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetPengumuman(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT id, COALESCE(judul,''), COALESCE(isi,''), COALESCE(tanggal,''), COALESCE(created_at,'') FROM pengumuman WHERE tenant_id = ? ORDER BY created_at DESC", tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() { var id int; var j, i, t, ca string; rows.Scan(&id, &j, &i, &t, &ca); list = append(list, fiber.Map{"id": id, "judul": j, "isi": i, "tanggal": t, "created_at": ca}) }
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreatePengumuman(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int); uid := c.Locals("user_id").(int)
	var body struct { Judul string `json:"judul"`; Isi string `json:"isi"` }
	c.BodyParser(&body)
	res, _ := config.DB.Exec("INSERT INTO pengumuman (tenant_id, judul, isi, created_by) VALUES (?,?,?,?)", tid, body.Judul, body.Isi, uid)
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Pengumuman ditambahkan"})
}

func DeletePengumuman(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pengumuman WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}
