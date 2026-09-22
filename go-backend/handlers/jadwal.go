package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"github.com/gofiber/fiber/v2"
)

// ── Jadwal Kegiatan (Umum) ───────────────────────────────
func GetJadwalUmum(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT id, COALESCE(kelompok_id,0), COALESCE(ustadz_username,''), COALESCE(hari,''), COALESCE(jam_mulai,''), COALESCE(jam_selesai,'') FROM jadwal_umum WHERE tenant_id = ?", tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kid int; var u, h, jm, js string
		rows.Scan(&id, &kid, &u, &h, &jm, &js)
		list = append(list, fiber.Map{"id": id, "kelompok_id": kid, "ustadz_username": u, "hari": h, "jam_mulai": jm, "jam_selesai": js})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateJadwalUmum(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		KelompokID     int      `json:"kelompok_id"`
		UstadzUsername string   `json:"ustadz_username"`
		HariList       []string `json:"hari_list"`
		Hari           string   `json:"hari"`
		JamMulai       string   `json:"jam_mulai"`
		JamSelesai     string   `json:"jam_selesai"`
	}
	c.BodyParser(&body)
	// Build list of days: prefer hari_list, fallback to single hari
	days := body.HariList
	if len(days) == 0 && body.Hari != "" {
		days = []string{body.Hari}
	}
	if len(days) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Pilih minimal 1 hari"})
	}
	var lastID int64
	for _, hari := range days {
		res, _ := config.DB.Exec("INSERT INTO jadwal_umum (tenant_id, kelompok_id, ustadz_username, hari, jam_mulai, jam_selesai) VALUES (?,?,?,?,?,?)",
			tid, body.KelompokID, body.UstadzUsername, hari, body.JamMulai, body.JamSelesai)
		lastID, _ = res.LastInsertId()
	}
	return c.JSON(fiber.Map{"id": lastID, "message": fmt.Sprintf("Jadwal ditambahkan untuk %d hari", len(days))})
}

func DeleteJadwalUmum(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM jadwal_umum WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}

// ── Jadwal Sekolah ───────────────────────────────────────
func GetJadwalSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT j.id, COALESCE(j.kelas_id,0), COALESCE(ks.nama,''), COALESCE(j.mata_pelajaran,''),
		COALESCE(j.ustadz_username,''), COALESCE(j.hari,''), COALESCE(j.jam_mulai,''), COALESCE(j.jam_selesai,'')
		FROM jadwal_sekolah j
		LEFT JOIN kelas_sekolah ks ON j.kelas_id = ks.id
		WHERE j.tenant_id = ?`
	args := []interface{}{tid}
	if v := c.Query("kelas_id"); v != "" {
		q += " AND j.kelas_id = ?"
		args = append(args, v)
	}
	q += " ORDER BY j.hari, j.jam_mulai"
	rows, _ := config.DB.Query(q, args...)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kid int; var kn, mp, u, h, jm, js string
		rows.Scan(&id, &kid, &kn, &mp, &u, &h, &jm, &js)
		list = append(list, fiber.Map{
			"id": id, "kelas_id": kid, "kelas": kn, "mata_pelajaran": mp,
			"ustadz_username": u, "hari": h, "jam_mulai": jm, "jam_selesai": js,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateJadwalSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		KelasID        int      `json:"kelas_id"`
		MataPelajaran  string   `json:"mata_pelajaran"`
		UstadzUsername string   `json:"ustadz_username"`
		HariList       []string `json:"hari_list"`
		Hari           string   `json:"hari"`
		JamMulai       string   `json:"jam_mulai"`
		JamSelesai     string   `json:"jam_selesai"`
	}
	c.BodyParser(&body)
	if body.MataPelajaran == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Mata pelajaran wajib"})
	}
	// Build list of days: prefer hari_list, fallback to single hari
	days := body.HariList
	if len(days) == 0 && body.Hari != "" {
		days = []string{body.Hari}
	}
	if len(days) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Pilih minimal 1 hari"})
	}
	var lastID int64
	for _, hari := range days {
		res, err := config.DB.Exec(
			"INSERT INTO jadwal_sekolah (tenant_id, kelas_id, mata_pelajaran, ustadz_username, hari, jam_mulai, jam_selesai) VALUES (?,?,?,?,?,?,?)",
			tid, body.KelasID, body.MataPelajaran, body.UstadzUsername, hari, body.JamMulai, body.JamSelesai)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal: " + err.Error()})
		}
		lastID, _ = res.LastInsertId()
	}
	// Auto-create mata pelajaran nilai if not exists
	var existCount int
	config.DB.QueryRow("SELECT COUNT(*) FROM mata_pelajaran_sekolah WHERE tenant_id = ? AND kelas_id = ? AND nama = ?",
		tid, body.KelasID, body.MataPelajaran).Scan(&existCount)
	if existCount == 0 {
		config.DB.Exec("INSERT INTO mata_pelajaran_sekolah (tenant_id, kelas_id, nama) VALUES (?,?,?)",
			tid, body.KelasID, body.MataPelajaran)
	}
	return c.JSON(fiber.Map{"id": lastID, "message": fmt.Sprintf("Jadwal ditambahkan untuk %d hari", len(days))})
}

func DeleteJadwalSekolah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM jadwal_sekolah WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}

// ── Jadwal Diniyyah ─────────────────────────────────────
func GetJadwalDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT j.id, COALESCE(j.kelas_diniyyah_id,0), COALESCE(kd.nama,''), COALESCE(j.mata_pelajaran,''),
		COALESCE(j.ustadz_username,''), COALESCE(j.hari,''), COALESCE(j.jam_mulai,''), COALESCE(j.jam_selesai,'')
		FROM jadwal_diniyyah j
		LEFT JOIN kelas_diniyyah kd ON j.kelas_diniyyah_id = kd.id
		WHERE j.tenant_id = ?`
	args := []interface{}{tid}
	if v := c.Query("kelas_diniyyah_id"); v != "" {
		q += " AND j.kelas_diniyyah_id = ?"
		args = append(args, v)
	}
	q += " ORDER BY j.hari, j.jam_mulai"
	rows, _ := config.DB.Query(q, args...)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kid int
		var kn, mp, u, h, jm, js string
		rows.Scan(&id, &kid, &kn, &mp, &u, &h, &jm, &js)
		list = append(list, fiber.Map{
			"id": id, "kelas_diniyyah_id": kid, "kelas": kn, "mata_pelajaran": mp,
			"ustadz_username": u, "hari": h, "jam_mulai": jm, "jam_selesai": js,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateJadwalDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		KelasDiniyyahID int      `json:"kelas_diniyyah_id"`
		MataPelajaran   string   `json:"mata_pelajaran"`
		UstadzUsername   string   `json:"ustadz_username"`
		HariList        []string `json:"hari_list"`
		Hari            string   `json:"hari"`
		JamMulai        string   `json:"jam_mulai"`
		JamSelesai      string   `json:"jam_selesai"`
	}
	c.BodyParser(&body)
	if body.MataPelajaran == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Mata pelajaran wajib"})
	}
	// Build list of days: prefer hari_list, fallback to single hari
	days := body.HariList
	if len(days) == 0 && body.Hari != "" {
		days = []string{body.Hari}
	}
	if len(days) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Pilih minimal 1 hari"})
	}
	var lastID int64
	for _, hari := range days {
		res, err := config.DB.Exec(
			"INSERT INTO jadwal_diniyyah (tenant_id, kelas_diniyyah_id, mata_pelajaran, ustadz_username, hari, jam_mulai, jam_selesai) VALUES (?,?,?,?,?,?,?)",
			tid, body.KelasDiniyyahID, body.MataPelajaran, body.UstadzUsername, hari, body.JamMulai, body.JamSelesai)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal: " + err.Error()})
		}
		lastID, _ = res.LastInsertId()
	}
	// Auto-create mata pelajaran nilai if not exists
	var existCount int
	config.DB.QueryRow("SELECT COUNT(*) FROM mata_pelajaran WHERE tenant_id = ? AND kelas_diniyyah_id = ? AND nama = ?",
		tid, body.KelasDiniyyahID, body.MataPelajaran).Scan(&existCount)
	if existCount == 0 {
		config.DB.Exec("INSERT INTO mata_pelajaran (tenant_id, kelas_diniyyah_id, nama) VALUES (?,?,?)",
			tid, body.KelasDiniyyahID, body.MataPelajaran)
	}
	return c.JSON(fiber.Map{"id": lastID, "message": fmt.Sprintf("Jadwal ditambahkan untuk %d hari", len(days))})
}

func DeleteJadwalDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM jadwal_diniyyah WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Dihapus"})
}

// ── Jadwal Kegiatan Aktif (hari ini, untuk ustadz) ───────
func GetJadwalAktif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	role := c.Locals("role").(string)
	uname := c.Locals("username").(string)
	hari := helpers.HariIni()
	now := helpers.JamSekarang()

	q := "SELECT j.id, j.kelompok_id, COALESCE(k.nama,''), j.ustadz_username, j.jam_mulai, j.jam_selesai FROM jadwal_umum j LEFT JOIN kelompok k ON j.kelompok_id = k.id WHERE j.tenant_id = ? AND (j.hari = ? OR j.hari = 'Setiap Hari')"
	args := []interface{}{tid, hari}
	if role != "admin" { q += " AND j.ustadz_username = ?"; args = append(args, uname) }
	rows, _ := config.DB.Query(q, args...)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kid int; var kn, u, jm, js string
		rows.Scan(&id, &kid, &kn, &u, &jm, &js)
		status := "belum_waktunya"
		if now >= jm && now <= helpers.AddMinutes(js, 60) { status = "siap_absen" } else if now > helpers.AddMinutes(js, 60) { status = "sudah_lewat" }
		list = append(list, fiber.Map{"id": id, "kelompok_id": kid, "kelompok_nama": kn, "ustadz_username": u, "jam_mulai": jm, "jam_selesai": js, "jenis": "umum", "status": status})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(fiber.Map{"hari": hari, "jam_sekarang": now, "jadwal": list})
}

// ── Jadwal Sekolah Aktif (hari ini, untuk ustadz/guru) ───
func GetJadwalSekolahAktif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	role := c.Locals("role").(string)
	uname := c.Locals("username").(string)
	hari := helpers.HariIni()
	now := helpers.JamSekarang()

	q := `SELECT j.id, j.kelas_id, COALESCE(ks.nama,''), j.mata_pelajaran,
		j.ustadz_username, j.jam_mulai, j.jam_selesai
		FROM jadwal_sekolah j
		LEFT JOIN kelas_sekolah ks ON j.kelas_id = ks.id
		WHERE j.tenant_id = ? AND j.hari = ?`
	args := []interface{}{tid, hari}
	if role != "admin" { q += " AND j.ustadz_username = ?"; args = append(args, uname) }
	q += " ORDER BY j.jam_mulai"
	rows, err := config.DB.Query(q, args...)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kid int; var kn, mp, u, jm, js string
		rows.Scan(&id, &kid, &kn, &mp, &u, &jm, &js)
		status := "belum_waktunya"
		if now >= jm && now <= helpers.AddMinutes(js, 60) { status = "siap_absen" } else if now > helpers.AddMinutes(js, 60) { status = "sudah_lewat" }
		list = append(list, fiber.Map{
			"id": id, "kelas_id": kid, "kelas": kn, "mata_pelajaran": mp,
			"ustadz_username": u, "jam_mulai": jm, "jam_selesai": js,
			"jenis": "sekolah", "status": status,
		})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(fiber.Map{"hari": hari, "jam_sekarang": now, "jadwal": list})
}
