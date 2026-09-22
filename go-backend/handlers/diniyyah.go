package handlers

import (
	"fmt"
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
)

// ═══════════════════════════════════════════════════════════
// KELAS DINIYYAH CRUD
// ═══════════════════════════════════════════════════════════

func GetKelasDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama FROM kelas_diniyyah WHERE tenant_id = ? ORDER BY nama", tid)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int; var nama string
		rows.Scan(&id, &nama)
		var jml int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelas_diniyyah WHERE kelas_diniyyah_id = ? AND tenant_id = ? AND status = 'active'", id, tid).Scan(&jml)
		var jmlMapel int
		config.DB.QueryRow("SELECT COUNT(*) FROM mata_pelajaran WHERE kelas_diniyyah_id = ? AND tenant_id = ?", id, tid).Scan(&jmlMapel)
		list = append(list, fiber.Map{"id": id, "nama": nama, "jumlah_santri": jml, "jumlah_mapel": jmlMapel})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateKelasDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct{ Nama string `json:"nama"` }
	c.BodyParser(&body)
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	res, err := config.DB.Exec("INSERT INTO kelas_diniyyah (tenant_id, nama) VALUES (?,?)", tid, body.Nama)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama})
}

func UpdateKelasDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct{ Nama string `json:"nama"` }
	c.BodyParser(&body)
	if body.Nama == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama wajib"}) }
	config.DB.Exec("UPDATE kelas_diniyyah SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Kelas diniyyah diperbarui"})
}

func DeleteKelasDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kid := c.Params("id")
	config.DB.Exec("DELETE FROM nilai_pelajaran WHERE kelas_diniyyah_id = ? AND tenant_id = ?", kid, tid)
	config.DB.Exec("DELETE FROM mata_pelajaran WHERE kelas_diniyyah_id = ? AND tenant_id = ?", kid, tid)
	config.DB.Exec("DELETE FROM santri_kelas_diniyyah WHERE kelas_diniyyah_id = ? AND tenant_id = ?", kid, tid)
	config.DB.Exec("DELETE FROM kelas_diniyyah WHERE id = ? AND tenant_id = ?", kid, tid)
	return c.JSON(fiber.Map{"message": "Kelas diniyyah dihapus"})
}

// ═══════════════════════════════════════════════════════════
// ANGGOTA KELAS DINIYYAH
// ═══════════════════════════════════════════════════════════

func GetKelasDiniyyahMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kdID := c.Params("id")
	rows, err := config.DB.Query(`SELECT s.id, s.nama, COALESCE(s.alamat,''), s.status, skd.id as skd_id
		FROM santri_kelas_diniyyah skd
		JOIN santri s ON skd.santri_id = s.id
		WHERE skd.kelas_diniyyah_id = ? AND skd.tenant_id = ? AND skd.status = 'active'
		ORDER BY s.nama`, kdID, tid)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, skdID int; var nama, alamat, status string
		rows.Scan(&id, &nama, &alamat, &status, &skdID)
		list = append(list, fiber.Map{"id": id, "nama": nama, "alamat": alamat, "status": status, "santri_kelas_diniyyah_id": skdID})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func AddKelasDiniyyahMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kdID := c.Params("id")
	var body struct{ SantriID int `json:"santri_id"` }
	c.BodyParser(&body)
	if body.SantriID == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_id wajib"}) }
	// Check duplicate
	var ex int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelas_diniyyah WHERE kelas_diniyyah_id = ? AND santri_id = ? AND tenant_id = ? AND status = 'active'", kdID, body.SantriID, tid).Scan(&ex)
	if ex > 0 { return c.Status(400).JSON(fiber.Map{"message": "Santri sudah ada di kelas ini"}) }
	config.DB.Exec("INSERT INTO santri_kelas_diniyyah (tenant_id, santri_id, kelas_diniyyah_id, status) VALUES (?,?,?,?)", tid, body.SantriID, kdID, "active")
	return c.JSON(fiber.Map{"message": "Santri ditambahkan ke kelas diniyyah"})
}

func RemoveKelasDiniyyahMember(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM santri_kelas_diniyyah WHERE kelas_diniyyah_id = ? AND santri_id = ? AND tenant_id = ?", c.Params("id"), c.Params("santri_id"), tid)
	return c.JSON(fiber.Map{"message": "Santri dihapus dari kelas diniyyah"})
}

// BulkAddKelasDiniyyahMembers — add multiple santri at once
func BulkAddKelasDiniyyahMembers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kdID := c.Params("id")
	var body struct{ SantriIDs []int `json:"santri_ids"` }
	c.BodyParser(&body)
	if len(body.SantriIDs) == 0 { return c.Status(400).JSON(fiber.Map{"message": "santri_ids wajib"}) }
	count := 0
	for _, sid := range body.SantriIDs {
		var ex int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri_kelas_diniyyah WHERE kelas_diniyyah_id = ? AND santri_id = ? AND tenant_id = ? AND status = 'active'", kdID, sid, tid).Scan(&ex)
		if ex > 0 { continue }
		_, err := config.DB.Exec("INSERT INTO santri_kelas_diniyyah (tenant_id, santri_id, kelas_diniyyah_id, status) VALUES (?,?,?,?)", tid, sid, kdID, "active")
		if err == nil { count++ }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d santri ditambahkan", count)})
}
