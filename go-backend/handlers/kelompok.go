package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"github.com/gofiber/fiber/v2"
)

func GetKelompok(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT id, nama, COALESCE(tipe,''), COALESCE(kegiatan_nama,''), COALESCE(kegiatan_id,0) FROM kelompok WHERE tenant_id = ?", tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() { var id, kid int; var n, t, kn string; rows.Scan(&id, &n, &t, &kn, &kid); list = append(list, fiber.Map{"id": id, "nama": n, "tipe": t, "kegiatan_nama": kn, "kegiatan_id": kid}) }
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateKelompok(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct{ Nama string `json:"nama"`; Tipe string `json:"tipe"`; KegiatanNama string `json:"kegiatan_nama"`; KegiatanID *int `json:"kegiatan_id"` }
	c.BodyParser(&body)
	var kid interface{} = nil
	if body.KegiatanID != nil { kid = *body.KegiatanID }
	res, _ := config.DB.Exec("INSERT INTO kelompok (tenant_id, nama, tipe, kegiatan_nama, kegiatan_id) VALUES (?,?,?,?,?)", tid, body.Nama, body.Tipe, body.KegiatanNama, kid)
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama, "tipe": body.Tipe})
}

func DeleteKelompok(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM kelompok WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}

func GetKelompokMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT sk.id, sk.santri_id, s.nama, COALESCE(s.kelas_diniyyah,'') FROM santri_kelompok sk JOIN santri s ON sk.santri_id = s.id WHERE sk.kelompok_id = ? AND sk.tenant_id = ? AND sk.status = 'active'", c.Params("id"), tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() { var skid, sid int; var n, k string; rows.Scan(&skid, &sid, &n, &k); list = append(list, fiber.Map{"sk_id": skid, "santri_id": sid, "santri_nama": n, "kelas_diniyyah": k}) }
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func AddKelompokMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct{ SantriID int `json:"santri_id"` }
	c.BodyParser(&body)
	if body.SantriID == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_id wajib"}) }
	config.DB.Exec("INSERT INTO santri_kelompok (tenant_id, santri_id, kelompok_id, status) VALUES (?,?,?,?)", tid, body.SantriID, c.Params("id"), "active")
	return c.JSON(fiber.Map{"message": "Anggota ditambahkan"})
}

func RemoveKelompokMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM santri_kelompok WHERE kelompok_id = ? AND santri_id = ? AND tenant_id = ?", c.Params("id"), c.Params("santri_id"), tid)
	return c.JSON(fiber.Map{"message": "Anggota dihapus"})
}

func BulkAddKelompokMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kelId := c.Params("id")
	var body struct{ SantriIDs []int `json:"santri_ids"` }
	c.BodyParser(&body)
	if len(body.SantriIDs) == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_ids wajib"}) }
	count := 0
	for _, sid := range body.SantriIDs {
		var ex int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelompok WHERE kelompok_id = ? AND santri_id = ? AND tenant_id = ? AND status = 'active'", kelId, sid, tid).Scan(&ex)
		if ex > 0 { continue }
		_, err := config.DB.Exec("INSERT INTO santri_kelompok (tenant_id, santri_id, kelompok_id, status) VALUES (?,?,?,?)", tid, sid, kelId, "active")
		if err == nil { count++ }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d anggota ditambahkan", count)})
}
