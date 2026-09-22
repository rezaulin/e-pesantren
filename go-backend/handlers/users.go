package handlers

import (
	"pesantren-multi/config"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, _ := config.DB.Query("SELECT id, username, nama, role, COALESCE(is_active,1), COALESCE(created_at,''), COALESCE(merchant_id,0) FROM users WHERE tenant_id = ? ORDER BY FIELD(role,'admin','ustadz','bendahara','keamanan','wali','merchant_admin','kasir'), nama", tid)
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, isActive, merchantID int; var u, n, r, ca string
		rows.Scan(&id, &u, &n, &r, &isActive, &ca, &merchantID)
		list = append(list, fiber.Map{"id": id, "username": u, "nama": n, "role": r, "is_active": isActive, "created_at": ca, "merchant_id": merchantID})
	}
	if list == nil { list = []fiber.Map{} }
	return c.JSON(list)
}

func CreateUser(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Role       string `json:"role"`
		Nama       string `json:"nama"`
		MerchantID *int   `json:"merchant_id"`
	}
	c.BodyParser(&body)
	if body.Username == "" || body.Password == "" || body.Nama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Semua field wajib"})
	}
	if body.Role == "" { body.Role = "ustadz" }
	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	res, err := config.DB.Exec("INSERT INTO users (tenant_id, username, password_hash, role, nama, merchant_id) VALUES (?, ?, ?, ?, ?, ?)",
		tid, body.Username, string(hash), body.Role, body.Nama, body.MerchantID)
	if err != nil { return c.Status(500).JSON(fiber.Map{"message": err.Error()}) }
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"message": "User ditambahkan", "user": fiber.Map{"id": id, "username": body.Username, "nama": body.Nama, "role": body.Role, "merchant_id": body.MerchantID}})
}

func UpdateUser(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	idStr := c.Params("id")
	id, _ := strconv.Atoi(idStr)

	var body struct {
		Role       string `json:"role"`
		Password   string `json:"password"`
		Nama       string `json:"nama"`
		IsActive   *int   `json:"is_active"`
		MerchantID *int   `json:"merchant_id"`
	}
	c.BodyParser(&body)

	// Update role if provided
	if body.Role != "" {
		allowed := map[string]bool{"admin": true, "ustadz": true, "bendahara": true, "keamanan": true, "wali": true, "merchant_admin": true, "kasir": true}
		if !allowed[body.Role] {
			return c.Status(400).JSON(fiber.Map{"message": "Role tidak valid"})
		}
		config.DB.Exec("UPDATE users SET role = ? WHERE id = ? AND tenant_id = ?", body.Role, id, tid)
	}

	// Update password if provided
	if body.Password != "" {
		if len(body.Password) < 6 {
			return c.Status(400).JSON(fiber.Map{"message": "Password minimal 6 karakter"})
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
		config.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ? AND tenant_id = ?", string(hash), id, tid)
	}

	// Update nama if provided
	if body.Nama != "" {
		config.DB.Exec("UPDATE users SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, id, tid)
	}

	// Update is_active if provided
	if body.IsActive != nil {
		config.DB.Exec("UPDATE users SET is_active = ? WHERE id = ? AND tenant_id = ?", *body.IsActive, id, tid)
	}

	// Update merchant_id if provided
	if body.MerchantID != nil {
		config.DB.Exec("UPDATE users SET merchant_id = ? WHERE id = ? AND tenant_id = ?", *body.MerchantID, id, tid)
	}

	return c.JSON(fiber.Map{"message": "User diperbarui"})
}

// BulkToggleUsers — activate/deactivate multiple users at once
func BulkToggleUsers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		IDs      []int `json:"ids"`
		IsActive int   `json:"is_active"`
	}
	c.BodyParser(&body)
	if len(body.IDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Pilih minimal 1 pengguna"})
	}
	// Build placeholder string
	placeholders := make([]string, len(body.IDs))
	args := make([]interface{}, 0, len(body.IDs)+2)
	args = append(args, body.IsActive, tid)
	for i, id := range body.IDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := "UPDATE users SET is_active = ? WHERE tenant_id = ? AND id IN (" + strings.Join(placeholders, ",") + ")"
	config.DB.Exec(query, args...)

	status := "diaktifkan"
	if body.IsActive == 0 {
		status = "dinonaktifkan"
	}
	return c.JSON(fiber.Map{"message": strconv.Itoa(len(body.IDs)) + " pengguna " + status})
}

func DeleteUser(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	idStr := c.Params("id")
	id, _ := strconv.Atoi(idStr)
	if id == uid { return c.Status(400).JSON(fiber.Map{"message": "Tidak bisa hapus diri sendiri"}) }
	config.DB.Exec("DELETE FROM users WHERE id = ? AND tenant_id = ?", id, tid)
	return c.JSON(fiber.Map{"message": "User dihapus"})
}
