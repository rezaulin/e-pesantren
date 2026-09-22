package handlers

import (
	"fmt"
	"math"
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
)

// ═══════════════════════════════════════════════════════════
// MATA PELAJARAN CRUD (relasi ke kelas_diniyyah)
// ═══════════════════════════════════════════════════════════

func GetMataPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := "SELECT mp.id, mp.kelas_diniyyah_id, mp.nama, COALESCE(kd.nama,'') FROM mata_pelajaran mp LEFT JOIN kelas_diniyyah kd ON mp.kelas_diniyyah_id = kd.id WHERE mp.tenant_id = ?"
	args := []interface{}{tid}
	if kid := c.Query("kelas_diniyyah_id"); kid != "" {
		q += " AND mp.kelas_diniyyah_id = ?"
		args = append(args, kid)
	}
	q += " ORDER BY mp.kelas_diniyyah_id, mp.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kdID int; var nama, kdNama string
		rows.Scan(&id, &kdID, &nama, &kdNama)
		list = append(list, fiber.Map{"id": id, "kelas_diniyyah_id": kdID, "nama": nama, "kelas_diniyyah_nama": kdNama})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateMataPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		KelasDiniyyahID int    `json:"kelas_diniyyah_id"`
		Nama            string `json:"nama"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.KelasDiniyyahID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nama dan kelas_diniyyah_id wajib"})
	}
	res, err := config.DB.Exec("INSERT INTO mata_pelajaran (tenant_id, kelas_diniyyah_id, nama) VALUES (?,?,?)", tid, body.KelasDiniyyahID, body.Nama)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama, "kelas_diniyyah_id": body.KelasDiniyyahID})
}

func UpdateMataPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama            string `json:"nama"`
		KelasDiniyyahID *int   `json:"kelas_diniyyah_id"`
	}
	c.BodyParser(&body)
	if body.Nama != "" {
		config.DB.Exec("UPDATE mata_pelajaran SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, c.Params("id"), tid)
	}
	if body.KelasDiniyyahID != nil {
		config.DB.Exec("UPDATE mata_pelajaran SET kelas_diniyyah_id = ? WHERE id = ? AND tenant_id = ?", *body.KelasDiniyyahID, c.Params("id"), tid)
	}
	return c.JSON(fiber.Map{"message": "Mata pelajaran diperbarui"})
}

func DeleteMataPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM nilai_pelajaran WHERE mata_pelajaran_id = ? AND tenant_id = ?", c.Params("id"), tid)
	config.DB.Exec("DELETE FROM mata_pelajaran WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Mata pelajaran dihapus"})
}

// ═══════════════════════════════════════════════════════════
// NILAI PELAJARAN (per kelas diniyyah, per semester)
// ═══════════════════════════════════════════════════════════

func computeNilaiAkhir(harian, uts, uas float64) float64 {
	na := (harian * 0.30) + (uts * 0.30) + (uas * 0.40)
	return math.Round(na*100) / 100
}

func GetNilaiPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT np.id, np.santri_id, s.nama, np.mata_pelajaran_id, mp.nama, np.kelas_diniyyah_id,
	      np.semester, np.nilai_harian, np.nilai_uts, np.nilai_uas, np.nilai_akhir
	      FROM nilai_pelajaran np
	      JOIN santri s ON np.santri_id = s.id
	      JOIN mata_pelajaran mp ON np.mata_pelajaran_id = mp.id
	      WHERE np.tenant_id = ?`
	args := []interface{}{tid}
	if kid := c.Query("kelas_diniyyah_id"); kid != "" { q += " AND np.kelas_diniyyah_id = ?"; args = append(args, kid) }
	if smt := c.Query("semester"); smt != "" { q += " AND np.semester = ?"; args = append(args, smt) }
	if sid := c.Query("santri_id"); sid != "" { q += " AND np.santri_id = ?"; args = append(args, sid) }
	q += " ORDER BY s.nama, mp.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, santriID, mpID, kdID int
		var santriNama, mpNama, semester string
		var harian, uts, uas, akhir float64
		rows.Scan(&id, &santriID, &santriNama, &mpID, &mpNama, &kdID, &semester, &harian, &uts, &uas, &akhir)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriID, "santri_nama": santriNama,
			"mata_pelajaran_id": mpID, "mata_pelajaran": mpNama, "kelas_diniyyah_id": kdID,
			"semester": semester, "nilai_harian": harian, "nilai_uts": uts,
			"nilai_uas": uas, "nilai_akhir": akhir,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func UpsertNilaiPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		SantriID        int     `json:"santri_id"`
		MataPelajaranID int     `json:"mata_pelajaran_id"`
		KelasDiniyyahID int     `json:"kelas_diniyyah_id"`
		Semester        string  `json:"semester"`
		NilaiHarian     float64 `json:"nilai_harian"`
		NilaiUTS        float64 `json:"nilai_uts"`
		NilaiUAS        float64 `json:"nilai_uas"`
		KKM             float64 `json:"kkm"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}
	if body.SantriID == 0 || body.MataPelajaranID == 0 || body.KelasDiniyyahID == 0 || body.Semester == "" {
		return c.Status(400).JSON(fiber.Map{"message": "santri_id, mata_pelajaran_id, kelas_diniyyah_id, dan semester wajib"})
	}
	if body.KKM == 0 {
		body.KKM = 75
	}
	nilaiAkhir := computeNilaiAkhir(body.NilaiHarian, body.NilaiUTS, body.NilaiUAS)
	res, err := config.DB.Exec(`INSERT INTO nilai_pelajaran 
		(tenant_id, santri_id, mata_pelajaran_id, kelas_diniyyah_id, semester, nilai_harian, nilai_uts, nilai_uas, nilai_akhir, kkm, created_by) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE 
		nilai_harian = VALUES(nilai_harian), nilai_uts = VALUES(nilai_uts), 
		nilai_uas = VALUES(nilai_uas), nilai_akhir = VALUES(nilai_akhir), kkm = VALUES(kkm), updated_at = NOW()`,
		tid, body.SantriID, body.MataPelajaranID, body.KelasDiniyyahID, body.Semester,
		body.NilaiHarian, body.NilaiUTS, body.NilaiUAS, nilaiAkhir, body.KKM, uid)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "nilai_akhir": nilaiAkhir, "message": "Nilai disimpan"})
}

// ═══════════════════════════════════════════════════════════
// NILAI KEGIATAN (per kelompok, per bulan)
// ═══════════════════════════════════════════════════════════

func GetNilaiKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT nk.id, nk.santri_id, s.nama, nk.kegiatan_id, kg.nama, nk.kelompok_id, kl.nama,
	      nk.bulan, nk.nilai, COALESCE(nk.catatan,'')
	      FROM nilai_kegiatan nk
	      JOIN santri s ON nk.santri_id = s.id
	      JOIN kegiatan kg ON nk.kegiatan_id = kg.id
	      JOIN kelompok kl ON nk.kelompok_id = kl.id
	      WHERE nk.tenant_id = ?`
	args := []interface{}{tid}
	if kid := c.Query("kelompok_id"); kid != "" { q += " AND nk.kelompok_id = ?"; args = append(args, kid) }
	if bln := c.Query("bulan"); bln != "" { q += " AND nk.bulan = ?"; args = append(args, bln) }
	if sid := c.Query("santri_id"); sid != "" { q += " AND nk.santri_id = ?"; args = append(args, sid) }
	if kgid := c.Query("kegiatan_id"); kgid != "" { q += " AND nk.kegiatan_id = ?"; args = append(args, kgid) }
	q += " ORDER BY s.nama, kg.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, santriID, kegiatanID, kelompokID int
		var santriNama, kegiatanNama, kelompokNama, bulan, catatan string
		var nilai float64
		rows.Scan(&id, &santriID, &santriNama, &kegiatanID, &kegiatanNama, &kelompokID, &kelompokNama, &bulan, &nilai, &catatan)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriID, "santri_nama": santriNama,
			"kegiatan_id": kegiatanID, "kegiatan": kegiatanNama,
			"kelompok_id": kelompokID, "kelompok": kelompokNama,
			"bulan": bulan, "nilai": nilai, "catatan": catatan,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func UpsertNilaiKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		SantriID   int     `json:"santri_id"`
		KegiatanID int     `json:"kegiatan_id"`
		KelompokID int     `json:"kelompok_id"`
		Bulan      string  `json:"bulan"`
		Nilai      float64 `json:"nilai"`
		Catatan    string  `json:"catatan"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}
	if body.SantriID == 0 || body.KegiatanID == 0 || body.KelompokID == 0 || body.Bulan == "" {
		return c.Status(400).JSON(fiber.Map{"message": "santri_id, kegiatan_id, kelompok_id, dan bulan wajib"})
	}
	res, err := config.DB.Exec(`INSERT INTO nilai_kegiatan 
		(tenant_id, santri_id, kegiatan_id, kelompok_id, bulan, nilai, catatan, created_by) 
		VALUES (?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE 
		nilai = VALUES(nilai), catatan = VALUES(catatan), updated_at = NOW()`,
		tid, body.SantriID, body.KegiatanID, body.KelompokID, body.Bulan, body.Nilai, body.Catatan, uid)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Nilai kegiatan disimpan"})
}

// ═══════════════════════════════════════════════════════════
// PERINGKAT
// ═══════════════════════════════════════════════════════════

// GetPeringkatDiniyyah — ranking santri per kelas diniyyah berdasar rata-rata nilai_akhir per semester
func GetPeringkatDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kdID := c.Params("kelas_diniyyah_id")
	semester := c.Query("semester")
	if semester == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Parameter semester wajib (e.g. 2025-1)"})
	}
	rows, err := config.DB.Query(`
		SELECT np.santri_id, s.nama, AVG(np.nilai_akhir) as rata_rata
		FROM nilai_pelajaran np JOIN santri s ON np.santri_id = s.id
		WHERE np.tenant_id = ? AND np.kelas_diniyyah_id = ? AND np.semester = ?
		GROUP BY np.santri_id, s.nama ORDER BY rata_rata DESC`, tid, kdID, semester)
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

// GetPeringkatKelompok — ranking santri per kelompok berdasar rata-rata nilai kegiatan per bulan
func GetPeringkatKelompok(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kelompokID := c.Params("kelompok_id")
	bulan := c.Query("bulan")
	if bulan == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Parameter bulan wajib (e.g. 2025-06)"})
	}
	rows, err := config.DB.Query(`
		SELECT nk.santri_id, s.nama, AVG(nk.nilai) as rata_rata
		FROM nilai_kegiatan nk JOIN santri s ON nk.santri_id = s.id
		WHERE nk.tenant_id = ? AND nk.kelompok_id = ? AND nk.bulan = ?
		GROUP BY nk.santri_id, s.nama ORDER BY rata_rata DESC`, tid, kelompokID, bulan)
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

// ═══════════════════════════════════════════════════════════
// BULK INPUT
// ═══════════════════════════════════════════════════════════

func BulkUpsertNilaiPelajaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		KelasDiniyyahID int    `json:"kelas_diniyyah_id"`
		MataPelajaranID int    `json:"mata_pelajaran_id"`
		Semester        string `json:"semester"`
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
	if body.KelasDiniyyahID == 0 || body.MataPelajaranID == 0 || body.Semester == "" || len(body.Data) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "kelas_diniyyah_id, mata_pelajaran_id, semester, dan data wajib"})
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
		_, err := config.DB.Exec(`INSERT INTO nilai_pelajaran 
			(tenant_id, santri_id, mata_pelajaran_id, kelas_diniyyah_id, semester, nilai_harian, nilai_uts, nilai_uas, nilai_akhir, kkm, nilai_detail, created_by) 
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE 
			nilai_harian=VALUES(nilai_harian), nilai_uts=VALUES(nilai_uts), nilai_uas=VALUES(nilai_uas), nilai_akhir=VALUES(nilai_akhir), kkm=VALUES(kkm), nilai_detail=VALUES(nilai_detail), updated_at=NOW()`,
			tid, d.SantriID, body.MataPelajaranID, body.KelasDiniyyahID, body.Semester, d.NilaiHarian, d.NilaiUTS, d.NilaiUAS, na, d.KKM, d.NilaiDetail, uid)
		if err == nil { count++ }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d nilai disimpan", count)})
}

func BulkUpsertNilaiKegiatan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		KegiatanID int    `json:"kegiatan_id"`
		KelompokID int    `json:"kelompok_id"`
		Bulan      string `json:"bulan"`
		Data []struct {
			SantriID int     `json:"santri_id"`
			Nilai    float64 `json:"nilai"`
			Catatan  string  `json:"catatan"`
		} `json:"data"`
	}
	if err := c.BodyParser(&body); err != nil { return c.Status(400).JSON(fiber.Map{"message": "Invalid request"}) }
	if body.KegiatanID == 0 || body.KelompokID == 0 || body.Bulan == "" || len(body.Data) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "kegiatan_id, kelompok_id, bulan, dan data wajib"})
	}
	count := 0
	for _, d := range body.Data {
		if d.SantriID == 0 { continue }
		_, err := config.DB.Exec(`INSERT INTO nilai_kegiatan 
			(tenant_id, santri_id, kegiatan_id, kelompok_id, bulan, nilai, catatan, created_by) 
			VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE 
			nilai=VALUES(nilai), catatan=VALUES(catatan), updated_at=NOW()`,
			tid, d.SantriID, body.KegiatanID, body.KelompokID, body.Bulan, d.Nilai, d.Catatan, uid)
		if err == nil { count++ }
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("%d nilai kegiatan disimpan", count)})
}
