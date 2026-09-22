package handlers

import (
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetCatatanGuru(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT c.id, c.santri_id, COALESCE(s.nama,''), COALESCE(c.catatan,''), COALESCE(c.tanggal,'') FROM catatan_guru c LEFT JOIN santri s ON c.santri_id = s.id WHERE c.tenant_id = ? ORDER BY c.tanggal DESC", tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() { var id, sid int; var sn, ct, tgl string; rows.Scan(&id, &sid, &sn, &ct, &tgl); list = append(list, fiber.Map{"id": id, "santri_id": sid, "santri_nama": sn, "catatan": ct, "tanggal": tgl}) }
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateCatatanGuru(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int); uid := c.Locals("user_id").(int)
	var body struct { SantriID int `json:"santri_id"`; Catatan string `json:"catatan"`; Tanggal string `json:"tanggal"` }
	c.BodyParser(&body)
	if body.SantriID == 0 { return c.Status(400).JSON(fiber.Map{"message": "Santri wajib dipilih"}) }
	res, _ := config.DB.Exec("INSERT INTO catatan_guru (tenant_id, santri_id, catatan, tanggal, created_by) VALUES (?,?,?,?,?)", tid, body.SantriID, body.Catatan, body.Tanggal, uid)
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Catatan disimpan"})
}

func DeleteCatatanGuru(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM catatan_guru WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Catatan dihapus"})
}
