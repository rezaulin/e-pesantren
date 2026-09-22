package handlers

import (
	"pesantren-multi/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GetPSBInfo — public endpoint, returns tenant name & address for PSB page
func GetPSBInfo(c *fiber.Ctx) error {
	subdomain := c.Params("subdomain")
	if subdomain == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Subdomain wajib"})
	}
	if idx := strings.Index(subdomain, "."); idx > 0 {
		subdomain = subdomain[:idx]
	}
	var nama, alamat string
	var kuota int
	err := config.DB.QueryRow("SELECT t.nama, COALESCE(s.alamat_lembaga,''), COALESCE(s.psb_kuota,0) FROM tenants t LEFT JOIN settings s ON s.tenant_id = t.id WHERE t.subdomain = ? AND t.status = 'active'", subdomain).Scan(&nama, &alamat, &kuota)
	if err != nil || nama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Pesantren tidak ditemukan"})
	}
	var tid int
	config.DB.QueryRow("SELECT id FROM tenants WHERE subdomain = ?", subdomain).Scan(&tid)
	var jumlah int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE tenant_id = ? AND status = 'pendaftar'", tid).Scan(&jumlah)
	return c.JSON(fiber.Map{"nama": nama, "alamat": alamat, "kuota": kuota, "jumlah_pendaftar": jumlah})
}

// SubmitPSB — public endpoint, no auth required
func SubmitPSB(c *fiber.Ctx) error {
	subdomain := c.Params("subdomain")
	if subdomain == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Subdomain wajib"})
	}
	// Strip domain suffix: "al-fatimiyyah.e-pesantren.app" → "al-fatimiyyah"
	if idx := strings.Index(subdomain, "."); idx > 0 {
		subdomain = subdomain[:idx]
	}
	// Resolve tenant
	var tid int
	var tStatus string
	err := config.DB.QueryRow("SELECT id, status FROM tenants WHERE subdomain = ?", subdomain).Scan(&tid, &tStatus)
	if err != nil || tid == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Pesantren tidak ditemukan"})
	}
	if tStatus != "active" {
		return c.Status(403).JSON(fiber.Map{"message": "Pendaftaran tidak tersedia"})
	}
	// Check PSB feature enabled
	var featStr *string
	config.DB.QueryRow("SELECT features FROM tenants WHERE id = ?", tid).Scan(&featStr)
	// If features is set but doesn't include 'psb', reject
	// (If features is NULL/empty = all features enabled)
	if featStr != nil && *featStr != "" && !containsFeature(*featStr, "psb") {
		return c.Status(403).JSON(fiber.Map{"message": "Pendaftaran online tidak aktif untuk pesantren ini"})
	}

	// Check PSB quota
	var kuota int
	config.DB.QueryRow("SELECT COALESCE(psb_kuota,0) FROM settings WHERE tenant_id = ?", tid).Scan(&kuota)
	if kuota > 0 {
		var jumlah int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE tenant_id = ? AND status = 'pendaftar'", tid).Scan(&jumlah)
		if jumlah >= kuota {
			return c.Status(403).JSON(fiber.Map{"message": "Maaf, kuota pendaftaran sudah penuh"})
		}
	}

	var body struct {
		Nama         string `json:"nama"`
		TempatLahir  string `json:"tempat_lahir"`
		TanggalLahir string `json:"tanggal_lahir"`
		JenisKelamin string `json:"jenis_kelamin"`
		Alamat       string `json:"alamat"`
		NamaAyah     string `json:"nama_ayah"`
		NamaIbu      string `json:"nama_ibu"`
		NoHP         string `json:"no_hp"`
		AsalSekolah  string `json:"asal_sekolah"`
		Catatan      string `json:"catatan"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if body.Nama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Nama wajib diisi"})
	}
	if body.JenisKelamin == "" {
		body.JenisKelamin = "L"
	}

	hp := normalizePhone(body.NoHP)

	_, err = config.DB.Exec(`INSERT INTO santri 
		(tenant_id, nama, tempat_lahir, tanggal_lahir, jenis_kelamin, alamat, 
		 nama_ayah, nama_ibu, no_hp, asal_sekolah, catatan_psb, 
		 nama_wali, kamar_id, status) 
		VALUES (?,?,?,NULLIF(?,''),?,?,?,?,?,?,?,?,0,'pendaftar')`,
		tid, body.Nama, body.TempatLahir, body.TanggalLahir, body.JenisKelamin,
		body.Alamat, body.NamaAyah, body.NamaIbu, hp, body.AsalSekolah,
		body.Catatan, body.NamaAyah)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan: " + err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Pendaftaran berhasil! Silakan tunggu informasi selanjutnya dari pesantren."})
}

func containsFeature(s, sub string) bool {
	return strings.Contains(s, `"`+sub+`"`) || strings.Contains(s, `'`+sub+`'`)
}

// GetPendaftar — admin: list calon santri
func GetPendaftar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query(`SELECT id, nama, COALESCE(tempat_lahir,''), COALESCE(tanggal_lahir,''), 
		COALESCE(jenis_kelamin,'L'), COALESCE(alamat,''), COALESCE(nama_ayah,''), COALESCE(nama_ibu,''),
		COALESCE(no_hp,''), COALESCE(asal_sekolah,''), COALESCE(catatan_psb,''), created_at
		FROM santri WHERE tenant_id = ? AND status = 'pendaftar' ORDER BY created_at DESC`, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama, tempatLahir, tglLahir, jk, alamat, ayah, ibu, hp, asal, catatan, createdAt string
		rows.Scan(&id, &nama, &tempatLahir, &tglLahir, &jk, &alamat, &ayah, &ibu, &hp, &asal, &catatan, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "nama": nama, "tempat_lahir": tempatLahir, "tanggal_lahir": tglLahir,
			"jenis_kelamin": jk, "alamat": alamat, "nama_ayah": ayah, "nama_ibu": ibu,
			"no_hp": hp, "asal_sekolah": asal, "catatan": catatan, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// TerimaPendaftar — admin: accept applicant → status = aktif, assign kamar, auto-create wali
func TerimaPendaftar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("id")
	var body struct {
		KamarID int `json:"kamar_id"`
	}
	c.BodyParser(&body)
	if body.KamarID == 0 {
		body.KamarID = 1
	}

	// Get santri info for wali creation
	var nama, hp, namaWali string
	config.DB.QueryRow("SELECT nama, COALESCE(no_hp,''), COALESCE(nama_ayah,'') FROM santri WHERE id = ? AND tenant_id = ? AND status = 'pendaftar'",
		sid, tid).Scan(&nama, &hp, &namaWali)
	if nama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Pendaftar tidak ditemukan"})
	}

	// Auto-create wali if has phone number
	waliUID := 0
	normalizedHP := normalizePhone(hp)
	if normalizedHP != "" {
		waliUID = autoCreateWaliUser(tid, normalizedHP, namaWali, nama)
	}

	// Update status to aktif
	config.DB.Exec("UPDATE santri SET status = 'aktif', kamar_id = ?, wali_user_id = ?, nama_wali = ? WHERE id = ? AND tenant_id = ?",
		body.KamarID, waliUID, namaWali, sid, tid)

	return c.JSON(fiber.Map{"message": "Santri diterima dan masuk ke sistem", "wali_created": waliUID > 0})
}

// TolakPendaftar — admin: reject applicant
func TolakPendaftar(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("id")
	config.DB.Exec("UPDATE santri SET status = 'ditolak' WHERE id = ? AND tenant_id = ? AND status = 'pendaftar'", sid, tid)
	return c.JSON(fiber.Map{"message": "Pendaftar ditolak"})
}
