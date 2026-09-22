package middleware

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"pesantren-multi/config"
)

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		log.Fatal("❌ JWT_SECRET environment variable wajib diset! Jangan jalankan server tanpa JWT_SECRET.")
	}
	if len(s) < 32 {
		log.Println("⚠️  WARNING: JWT_SECRET terlalu pendek (minimal 32 karakter untuk keamanan)")
	}
	return []byte(s)
}

type JWTClaims struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Nama     string `json:"nama"`
	TenantID int    `json:"tenant_id"`
	jwt.RegisteredClaims
}

func GenerateToken(id int, username, role, nama string, tenantID int) (string, error) {
	claims := JWTClaims{
		ID: id, Username: username, Role: role, Nama: nama, TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour))},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func Authenticate(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if tokenStr == "" || tokenStr == auth { return c.Status(401).JSON(fiber.Map{"message": "Token tidak ada"}) }
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) { return jwtSecret(), nil })
	if err != nil || !token.Valid { return c.Status(401).JSON(fiber.Map{"message": "Token tidak valid"}) }
	claims := token.Claims.(*JWTClaims)
	c.Locals("user_id", claims.ID)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)
	c.Locals("nama", claims.Nama)
	c.Locals("tenant_id", claims.TenantID)
	return c.Next()
}

func RequireAdmin(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" && role != "superadmin" { return c.Status(403).JSON(fiber.Map{"message": "Hanya admin"}) }
	return c.Next()
}

func RequireAdminOrBendahara(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" && role != "superadmin" && role != "bendahara" {
		return c.Status(403).JSON(fiber.Map{"message": "Hanya admin atau bendahara"})
	}
	return c.Next()
}

func RequireSuperAdmin(c *fiber.Ctx) error {
	if c.Locals("role").(string) != "superadmin" { return c.Status(403).JSON(fiber.Map{"message": "Hanya super admin"}) }
	return c.Next()
}

func RequireAdminOrUstadz(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" && role != "superadmin" && role != "ustadz" {
		return c.Status(403).JSON(fiber.Map{"message": "Hanya admin atau ustadz"})
	}
	return c.Next()
}

func RequireAdminOrKeamanan(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" && role != "superadmin" && role != "keamanan" {
		return c.Status(403).JSON(fiber.Map{"message": "Hanya admin atau keamanan"})
	}
	return c.Next()
}

func RequireAdminUstadzKeamanan(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" && role != "superadmin" && role != "ustadz" && role != "keamanan" {
		return c.Status(403).JSON(fiber.Map{"message": "Hanya admin, ustadz, atau keamanan"})
	}
	return c.Next()
}

func RequireWali(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "wali" && role != "admin" && role != "superadmin" {
		return c.Status(403).JSON(fiber.Map{"message": "Akses tidak diizinkan"})
	}
	return c.Next()
}

func CheckTenantExpiry(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role == "superadmin" { return c.Next() }
	tid := c.Locals("tenant_id").(int)
	var status string
	var expiredAt *string
	err := config.DB.QueryRow("SELECT status, expired_at FROM tenants WHERE id = ?", tid).Scan(&status, &expiredAt)
	if err != nil { return c.Next() }
	if status == "suspended" { return c.Status(403).JSON(fiber.Map{"message": "Akun pesantren Anda telah disuspend", "code": "TENANT_SUSPENDED"}) }
	if expiredAt != nil && *expiredAt != "" {
		exp, _ := time.Parse("2006-01-02", *expiredAt)
		if time.Now().After(exp) { return c.Status(403).JSON(fiber.Map{"message": "Langganan pesantren Anda telah habis", "code": "TENANT_EXPIRED"}) }
	}
	return c.Next()
}
