package handlers

import (
	"database/sql"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"pesantren-multi/middleware"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *fiber.Ctx) error {
	var body struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Subdomain string `json:"subdomain"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}

	var id, tenantID int
	var username, hash, role, nama string
	var isActive int

	// ── Resolve tenant from subdomain first ──────────────
	// Try superadmin first (superadmin can login from any subdomain)
	errSA := config.DB.QueryRow("SELECT id, tenant_id, username, password_hash, role, nama, COALESCE(is_active,1) FROM users WHERE username = ? AND role = 'superadmin'", body.Username).
		Scan(&id, &tenantID, &username, &hash, &role, &nama, &isActive)

	if errSA != nil {
		// Not a superadmin — must resolve tenant from subdomain
		resolvedTenantID := 0
		var tenantStatus string
		var tenantExpAt sql.NullString

		if body.Subdomain != "" && body.Subdomain != "localhost" && body.Subdomain != "app" {
			// Production: resolve tenant by subdomain
			err := config.DB.QueryRow("SELECT id, status, expired_at FROM tenants WHERE subdomain = ?", body.Subdomain).
				Scan(&resolvedTenantID, &tenantStatus, &tenantExpAt)
			if err != nil || resolvedTenantID == 0 {
				return c.Status(401).JSON(fiber.Map{"message": "Lembaga tidak ditemukan"})
			}
		} else {
			// Development (localhost/app): find user and get their tenant
			// Use the first matching user — acceptable for dev only
			err := config.DB.QueryRow("SELECT id, tenant_id, username, password_hash, role, nama, COALESCE(is_active,1) FROM users WHERE username = ? AND role != 'superadmin' LIMIT 1", body.Username).
				Scan(&id, &tenantID, &username, &hash, &role, &nama, &isActive)
			if err != nil {
				return c.Status(401).JSON(fiber.Map{"message": "Username/password salah"})
			}
			resolvedTenantID = tenantID
			// Fetch tenant status for dev mode too
			config.DB.QueryRow("SELECT status, expired_at FROM tenants WHERE id = ?", resolvedTenantID).Scan(&tenantStatus, &tenantExpAt)
		}

		// Check tenant status
		if tenantStatus == "suspended" {
			return c.Status(403).JSON(fiber.Map{"message": "Akun pesantren Anda telah disuspend. Hubungi administrator.", "code": "TENANT_SUSPENDED"})
		}
		// Pengecekan expired dipindah ke response supaya bisa nampilin halaman Masa Aktif Habis di dalam app
		// if tenantExpAt.Valid && tenantExpAt.String != "" && tenantExpAt.String < helpers.TodayWIB() {
		// 	return c.Status(403).JSON(fiber.Map{"message": "Langganan pesantren Anda telah habis. Hubungi administrator.", "code": "TENANT_EXPIRED"})
		// }

		// Query user scoped to the resolved tenant
		if body.Subdomain != "" && body.Subdomain != "localhost" && body.Subdomain != "app" {
			err := config.DB.QueryRow("SELECT id, tenant_id, username, password_hash, role, nama, COALESCE(is_active,1) FROM users WHERE username = ? AND tenant_id = ?", body.Username, resolvedTenantID).
				Scan(&id, &tenantID, &username, &hash, &role, &nama, &isActive)
			if err != nil {
				return c.Status(401).JSON(fiber.Map{"message": "Username/password salah"})
			}
		}
	}

	// Verify password
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Username/password salah"})
	}

	// Check if user is active
	if isActive == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Akun Anda telah dinonaktifkan. Hubungi admin pesantren."})
	}

	// Get tenant features & subdomain & saas info & nama
	var featuresStr sql.NullString
	var tenantSubdomain, tenantNama string
	var paketSaas string
	var kuotaSantri int
	var fiturSangu bool
	var masaAktif sql.NullString

	config.DB.QueryRow("SELECT COALESCE(subdomain,''), COALESCE(nama,'E-Pesantren'), features, paket_saas, kuota_santri, fitur_sangu, masa_aktif FROM tenants WHERE id = ?", tenantID).
		Scan(&tenantSubdomain, &tenantNama, &featuresStr, &paketSaas, &kuotaSantri, &fiturSangu, &masaAktif)

	isExpired := false
	if masaAktif.Valid && masaAktif.String != "" && masaAktif.String < helpers.TodayWIB() {
		isExpired = true
	}

	token, err := middleware.GenerateToken(id, username, role, nama, tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Token error"})
	}
	userMap := fiber.Map{
		"id": id, "username": username, "role": role, "nama": nama, "tenant_id": tenantID, 
		"tenant_nama": tenantNama, "subdomain": tenantSubdomain, "paket_saas": paketSaas, "kuota_santri": kuotaSantri,
		"fitur_sangu": fiturSangu, "is_expired": isExpired,
	}
	if featuresStr.Valid && featuresStr.String != "" {
		userMap["features"] = featuresStr.String
	}
	return c.JSON(fiber.Map{"token": token, "user": userMap})
}

func Me(c *fiber.Ctx) error {
	uid := c.Locals("user_id").(int)
	var id, tenantID int
	var username, role, nama string
	err := config.DB.QueryRow("SELECT id, username, role, nama, tenant_id FROM users WHERE id = ?", uid).Scan(&id, &username, &role, &nama, &tenantID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "User tidak ditemukan"})
	}
	
	// Get tenant saas info
	var paketSaas string
	var kuotaSantri int
	var fiturSangu bool
	var masaAktif sql.NullString
	config.DB.QueryRow("SELECT paket_saas, kuota_santri, fitur_sangu, masa_aktif FROM tenants WHERE id = ?", tenantID).
		Scan(&paketSaas, &kuotaSantri, &fiturSangu, &masaAktif)

	isExpired := false
	if masaAktif.Valid && masaAktif.String != "" && masaAktif.String < helpers.TodayWIB() {
		isExpired = true
	}

	result := fiber.Map{
		"id": id, "username": username, "role": role, "nama": nama, "tenant_id": tenantID,
		"paket_saas": paketSaas, "kuota_santri": kuotaSantri, "fitur_sangu": fiturSangu, "is_expired": isExpired,
	}
	// Include tenant features & subdomain
	var featuresStr sql.NullString
	var tenantSubdomain string
	config.DB.QueryRow("SELECT COALESCE(subdomain,''), features FROM tenants WHERE id = ?", tenantID).Scan(&tenantSubdomain, &featuresStr)
	result["subdomain"] = tenantSubdomain
	if featuresStr.Valid && featuresStr.String != "" {
		result["features"] = featuresStr.String
	}
	return c.JSON(result)
}

func UserChangePassword(c *fiber.Ctx) error {
	uid := c.Locals("user_id").(int)
	tid := c.Locals("tenant_id").(int)

	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}

	if len(body.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"message": "Password minimal 6 karakter"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengenkripsi password"})
	}

	_, err = config.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ? AND tenant_id = ?", string(hash), uid, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan password"})
	}

	return c.JSON(fiber.Map{"message": "Password berhasil diubah"})
}
