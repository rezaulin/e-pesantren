package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"pesantren-multi/config"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func GetTenants(c *fiber.Ctx) error {
	rows, err := config.DB.Query("SELECT id, nama, subdomain, COALESCE(alamat,''), status, COALESCE(expired_at,''), COALESCE(created_at,''), COALESCE(features,''), COALESCE(paket_saas,'trial'), COALESCE(kuota_santri, 100), COALESCE(fitur_sangu, 0), COALESCE(masa_aktif, '') FROM tenants ORDER BY created_at DESC")
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kuotaSantri int; var nama, sub, alamat, status, exp, ca, features, paketSaas, masaAktif string
		var fiturSangu bool
		rows.Scan(&id, &nama, &sub, &alamat, &status, &exp, &ca, &features, &paketSaas, &kuotaSantri, &fiturSangu, &masaAktif)
		list = append(list, fiber.Map{"id": id, "nama": nama, "subdomain": sub, "alamat": alamat, "status": status, "expired_at": nilStr(exp), "created_at": ca, "features": nilStr(features), "paket_saas": paketSaas, "kuota_santri": kuotaSantri, "fitur_sangu": fiturSangu, "masa_aktif": masaAktif})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateTenant(c *fiber.Ctx) error {
	var body struct {
		Nama      string   `json:"nama"`
		Subdomain string   `json:"subdomain"`
		Alamat    string   `json:"alamat"`
		Features  []string `json:"features"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.Subdomain == "" { return c.Status(400).JSON(fiber.Map{"message": "Nama & subdomain wajib"}) }
	expired := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	// Encode features as JSON string
	var featuresStr *string
	if len(body.Features) > 0 {
		b, _ := json.Marshal(body.Features)
		s := string(b)
		featuresStr = &s
	}
	res, err := config.DB.Exec("INSERT INTO tenants (nama, subdomain, alamat, expired_at, features) VALUES (?, ?, ?, ?, ?)", body.Nama, body.Subdomain, body.Alamat, expired, featuresStr)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	tid, _ := res.LastInsertId()
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	uname := body.Subdomain + "_admin"
	config.DB.Exec("INSERT INTO users (tenant_id, username, password_hash, role, nama) VALUES (?, ?, ?, 'admin', ?)", tid, uname, string(hash), "Admin "+body.Nama)
	config.DB.Exec("INSERT INTO settings (tenant_id, app_name) VALUES (?, ?)", tid, body.Nama)
	return c.JSON(fiber.Map{"id": tid, "nama": body.Nama, "subdomain": body.Subdomain, "message": "Tenant dibuat. Login: " + uname + " / admin123"})
}

func UpdateTenant(c *fiber.Ctx) error {
	id := c.Params("id")
	var body map[string]interface{}
	c.BodyParser(&body)
	if nama, ok := body["nama"].(string); ok && nama != "" { config.DB.Exec("UPDATE tenants SET nama = ? WHERE id = ?", nama, id) }
	if status, ok := body["status"].(string); ok && status != "" { config.DB.Exec("UPDATE tenants SET status = ? WHERE id = ?", status, id) }
	if alamat, ok := body["alamat"].(string); ok { config.DB.Exec("UPDATE tenants SET alamat = ? WHERE id = ?", alamat, id) }
	// Handle features update
	if featuresRaw, ok := body["features"]; ok {
		if featuresArr, ok := featuresRaw.([]interface{}); ok {
			var features []string
			for _, f := range featuresArr { if s, ok := f.(string); ok { features = append(features, s) } }
			b, _ := json.Marshal(features)
			config.DB.Exec("UPDATE tenants SET features = ?, updated_at = NOW() WHERE id = ?", string(b), id)
		} else if featuresRaw == nil {
			config.DB.Exec("UPDATE tenants SET features = NULL, updated_at = NOW() WHERE id = ?", id)
		}
	}
	return c.JSON(fiber.Map{"message": "Tenant diupdate"})
}

func DeleteTenant(c *fiber.Ctx) error {
	config.DB.Exec("DELETE FROM tenants WHERE id = ?", c.Params("id"))
	return c.JSON(fiber.Map{"message": "Tenant dihapus (CASCADE hapus semua data)"})
}

func SuperStats(c *fiber.Ctx) error {
	rows, _ := config.DB.Query("SELECT id, nama, subdomain, status, COALESCE(expired_at,'') FROM tenants")
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int; var nama, sub, status, exp string
		rows.Scan(&id, &nama, &sub, &status, &exp)
		var sc, uc int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE tenant_id = ?", id).Scan(&sc)
		config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE tenant_id = ?", id).Scan(&uc)
		list = append(list, fiber.Map{"id": id, "nama": nama, "subdomain": sub, "status": status, "expired_at": nilStr(exp), "total_santri": sc, "total_users": uc})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func ChangePassword(c *fiber.Ctx) error {
	var body struct{ OldPassword string `json:"old_password"`; NewPassword string `json:"new_password"` }
	c.BodyParser(&body)
	if body.OldPassword == "" || body.NewPassword == "" { return c.Status(400).JSON(fiber.Map{"message": "Password lama dan baru wajib diisi"}) }
	if len(body.NewPassword) < 6 { return c.Status(400).JSON(fiber.Map{"message": "Password baru minimal 6 karakter"}) }
	uid := c.Locals("user_id").(int)
	var hash string
	config.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", uid).Scan(&hash)
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.OldPassword)) != nil { return c.Status(400).JSON(fiber.Map{"message": "Password lama salah"}) }
	newHash, _ := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	config.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(newHash), uid)
	return c.JSON(fiber.Map{"message": "Password berhasil diubah"})
}

func ExtendTenant(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct{ Months int `json:"months"` }
	c.BodyParser(&body)
	if body.Months < 1 { return c.Status(400).JSON(fiber.Map{"message": "Jumlah bulan wajib minimal 1"}) }
	var expAt sql.NullString; var status string
	config.DB.QueryRow("SELECT COALESCE(expired_at,''), status FROM tenants WHERE id = ?", id).Scan(&expAt, &status)
	now := time.Now()
	base := now
	if expAt.Valid && expAt.String != "" {
		if t, err := time.Parse("2006-01-02", expAt.String); err == nil && t.After(now) { base = t }
	}
	newExp := base.AddDate(0, body.Months, 0).Format("2006-01-02")
	config.DB.Exec("UPDATE tenants SET expired_at = ?, status = 'active', updated_at = NOW() WHERE id = ?", newExp, id)
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Tenant diperpanjang %d bulan sampai %s", body.Months, newExp), "expired_at": newExp})
}

func ToggleTenant(c *fiber.Ctx) error {
	id := c.Params("id")
	var status string
	config.DB.QueryRow("SELECT status FROM tenants WHERE id = ?", id).Scan(&status)
	newStatus := "active"
	if status == "active" { newStatus = "suspended" }
	config.DB.Exec("UPDATE tenants SET status = ?, updated_at = NOW() WHERE id = ?", newStatus, id)
	msg := "diaktifkan"
	if newStatus == "suspended" { msg = "disuspend" }
	return c.JSON(fiber.Map{"message": "Tenant " + msg, "status": newStatus})
}

func Broadcast(c *fiber.Ctx) error {
	var body struct{ Judul string `json:"judul"`; Isi string `json:"isi"` }
	c.BodyParser(&body)
	if body.Judul == "" || body.Isi == "" { return c.Status(400).JSON(fiber.Map{"message": "Judul dan isi wajib diisi"}) }
	uid := c.Locals("user_id").(int)
	tanggal := time.Now().Format("2006-01-02")
	rows, _ := config.DB.Query("SELECT id FROM tenants WHERE subdomain != 'system'")
	defer rows.Close()
	count := 0
	for rows.Next() {
		var tid int; rows.Scan(&tid)
		config.DB.Exec("INSERT INTO pengumuman (tenant_id, judul, isi, tanggal, created_by) VALUES (?, ?, ?, ?, ?)", tid, body.Judul, body.Isi, tanggal, uid)
		count++
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Pengumuman dikirim ke %d tenant", count)})
}

func nilStr(s string) interface{} {
	if s == "" { return nil }
	return s
}

func atoi(s string) int { v, _ := strconv.Atoi(s); return v }
