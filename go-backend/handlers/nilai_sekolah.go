package handlers

import (
	"fmt"
	"math"
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
)

// ═══════════════════════════════════════════════════════════
// MATA PELAJARAN SEKOLAH CRUD (relasi ke kelas_sekolah)
// ═══════════════════════════════════════════════════════════

func GetMataPelajaranSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := "SELECT mp.id, mp.kelas_id, mp.nama, COALESCE(ks.nama,'') FROM mata_pelajaran_sekolah mp LEFT JOIN kelas_sekolah ks ON mp.kelas_id = ks.id WHERE mp.tenant_id = ?"
	args := []interface{}{tid}
	if kid := c.Query("kelas_id"); kid != "" {
		q += " AND mp.kelas_id = ?"
		args = append(args, kid)
	}
	q += " ORDER BY mp.kelas_id, mp.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kelasID int; var nama, kelasNama string
		rows.Scan(&id, &kelasID, &nama, &kelasNama)
		list = append(list, fiber.Map{"id": id, "kelas_id": kelasID, "nama": nama, "kelas_nama": kelasNama})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateMataPelajaranSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		KelasID int    `json:"kelas_id"`
		Nama    string `json:"nama"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.KelasID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nama dan kelas_id wajib"})
	}
	res, err := config.DB.Exec("INSERT INTO mata_pelajaran_sekolah (tenant_id, kelas_id, nama) VALUES (?,?,?)", tid, body.KelasID, body.Nama)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama, "kelas_id": body.KelasID})
}

func UpdateMataPelajaranSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama    string `json:"nama"`
		KelasID *int   `json:"kelas_id"`
	}
	c.BodyParser(&body)
	if body.Nama != "" {
		config.DB.Exec("UPDATE mata_pelajaran_sekolah SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, c.Params("id"), tid)
	}
	if body.KelasID != nil {
		config.DB.Exec("UPDATE mata_pelajaran_sekolah SET kelas_id = ? WHERE id = ? AND tenant_id = ?", *body.KelasID, c.Params("id"), tid)
	}
	return c.JSON(fiber.Map{"message": "Mata pelajaran sekolah diperbarui"})
}

func DeleteMataPelajaranSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM nilai_sekolah WHERE mata_pelajaran_sekolah_id = ? AND tenant_id = ?", c.Params("id"), tid)
	config.DB.Exec("DELETE FROM mata_pelajaran_sekolah WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Mata pelajaran sekolah dihapus"})
}

// ═══════════════════════════════════════════════════════════
// NILAI SEKOLAH (per kelas sekolah, per semester)
// ═══════════════════════════════════════════════════════════

func GetNilaiSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT ns.id, ns.santri_id, s.nama, ns.mata_pelajaran_sekolah_id, mp.nama, ns.kelas_id,
	      ns.semester, ns.nilai_harian, ns.nilai_uts, ns.nilai_uas, ns.nilai_akhir
	      FROM nilai_sekolah ns
	      JOIN santri s ON ns.santri_id = s.id
	      JOIN mata_pelajaran_sekolah mp ON ns.mata_pelajaran_sekolah_id = mp.id
	      WHERE ns.tenant_id = ?`
	args := []interface{}{tid}
	if kid := c.Query("kelas_id"); kid != "" { q += " AND ns.kelas_id = ?"; args = append(args, kid) }
	if smt := c.Query("semester"); smt != "" { q += " AND ns.semester = ?"; args = append(args, smt) }
	if sid := c.Query("santri_id"); sid != "" { q += " AND ns.santri_id = ?"; args = append(args, sid) }
	q += " ORDER BY s.nama, mp.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, santriID, mpID, kelasID int
		var santriNama, mpNama, semester string
		var harian, uts, uas, akhir float64
		rows.Scan(&id, &santriID, &santriNama, &mpID, &mpNama, &kelasID, &semester, &harian, &uts, &uas, &akhir)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriID, "santri_nama": santriNama,
			"mata_pelajaran_sekolah_id": mpID, "mata_pelajaran": mpNama, "kelas_id": kelasID,
			"semester": semester, "nilai_harian": harian, "nilai_uts": uts,
			"nilai_uas": uas, "nilai_akhir": akhir,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func UpsertNilaiSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		SantriID              int     `json:"santri_id"`
		MataPelajaranSekolahID int    `json:"mata_pelajaran_sekolah_id"`
		KelasID               int     `json:"kelas_id"`
		Semester              string  `json:"semester"`
		NilaiHarian           float64 `json:"nilai_harian"`
		NilaiUTS              float64 `json:"nilai_uts"`
		NilaiUAS              float64 `json:"nilai_uas"`
		KKM                   float64 `json:"kkm"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}
	if body.SantriID == 0 || body.MataPelajaranSekolahID == 0 || body.KelasID == 0 || body.Semester == "" {
		return c.Status(400).JSON(fiber.Map{"message": "santri_id, mata_pelajaran_sekolah_id, kelas_id, dan semester wajib"})
	}
	if body.KKM == 0 {
		body.KKM = 75
	}
	nilaiAkhir := computeNilaiAkhir(body.NilaiHarian, body.NilaiUTS, body.NilaiUAS)
	res, err := config.DB.Exec(`INSERT INTO nilai_sekolah 
		(tenant_id, santri_id, mata_pelajaran_sekolah_id, kelas_id, semester, nilai_harian, nilai_uts, nilai_uas, nilai_akhir, kkm, created_by) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE 
		nilai_harian = VALUES(nilai_harian), nilai_uts = VALUES(nilai_uts), 
		nilai_uas = VALUES(nilai_uas), nilai_akhir = VALUES(nilai_akhir), kkm = VALUES(kkm), updated_at = NOW()`,
		tid, body.SantriID, body.MataPelajaranSekolahID, body.KelasID, body.Semester,
		body.NilaiHarian, body.NilaiUTS, body.NilaiUAS, nilaiAkhir, body.KKM, uid)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nilai_akhir": nilaiAkhir, "message": "Nilai sekolah disimpan"})
}

func BulkUpsertNilaiSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		KelasID                int    `json:"kelas_id"`
		MataPelajaranSekolahID int    `json:"mata_pelajaran_sekolah_id"`
		Semester               string `json:"semester"`
		Data []struct {
			SantriID    int     `json:"santri_id"`
			NilaiHarian float64 `json:"nilai_harian"`
			NilaiUTS    float64 `json:"nilai_uts"`
			NilaiUAS    float64 `json:"nilai_uas"`
			NilaiAkhir  float64 `json:"nilai_akhir"`
			NilaiDetail string  `json:"nilai_detail"`
			KKM         float64 `json:"kkm"`
		} `json:"data"`
	}
	if err := c.BodyParser(&body); err != nil { return c.Status(400).JSON(fiber.Map{"message": "Invalid request"}) }
	if body.KelasID == 0 || body.MataPelajaranSekolahID == 0 || body.Semester == "" || len(body.Data) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "kelas_id, mata_pelajaran_sekolah_id, semester, dan data wajib"})
	}
	count := 0
	for _, d := range body.Data {
		if d.SantriID == 0 { continue }
		na := d.NilaiAkhir
		if na == 0 && (d.NilaiHarian > 0 || d.NilaiUTS > 0 || d.NilaiUAS > 0) {
			na = computeNilaiAkhir(d.NilaiHarian, d.NilaiUTS, d.NilaiUAS)
		}
		if d.KKM == 0 {
			d.KKM = 75
		}
		_, err := config.DB.Exec(`INSERT INTO nilai_sekolah 
			(tenant_id, santri_id, mata_pelajaran_sekolah_id, kelas_id, semester, nilai_harian, nilai_uts, nilai_uas, nilai_akhir, kkm, nilai_detail, created_by) 
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE 
			nilai_harian=VALUES(nilai_harian), nilai_uts=VALUES(nilai_uts), nilai_uas=VALUES(nilai_uas), nilai_akhir=VALUES(nilai_akhir), kkm=VALUES(kkm), nilai_detail=VALUES(nilai_detail), updated_at=NOW()`,
			tid, d.SantriID, body.MataPelajaranSekolahID, body.KelasID, body.Semester, d.NilaiHarian, d.NilaiUTS, d.NilaiUAS, na, d.KKM, d.NilaiDetail, uid)
		if err == nil { count++ } else { fmt.Println("Error insert:", err) }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d nilai sekolah disimpan", count)})
}

// ═══════════════════════════════════════════════════════════
// PERINGKAT SEKOLAH
// ═══════════════════════════════════════════════════════════

func GetPeringkatSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kelasID := c.Params("kelas_id")
	semester := c.Query("semester")
	if semester == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Parameter semester wajib (e.g. 2025-1)"})
	}
	rows, err := config.DB.Query(`
		SELECT ns.santri_id, s.nama, AVG(ns.nilai_akhir) as rata_rata
		FROM nilai_sekolah ns JOIN santri s ON ns.santri_id = s.id
		WHERE ns.tenant_id = ? AND ns.kelas_id = ? AND ns.semester = ?
		GROUP BY ns.santri_id, s.nama ORDER BY rata_rata DESC`, tid, kelasID, semester)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	rank := 1
	for rows.Next() {
		var santriID int; var nama string; var rataRata float64
		rows.Scan(&santriID, &nama, &rataRata)
		rataRata = math.Round(rataRata*100) / 100
		list = append(list, fiber.Map{"peringkat": rank, "santri_id": santriID, "santri_nama": nama, "rata_rata": rataRata})
		rank++
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}
