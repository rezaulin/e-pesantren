package handlers

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"pesantren-multi/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"golang.org/x/crypto/bcrypt"
)

// generateRandomPassword membuat password acak sepanjang n karakter
func generateRandomPassword(n int) string {
	const charset = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

// ═══════════════════════════════════════════════════════════
// PUBLIC: POST /api/register-tenant (Landing Page Registration)
// ═══════════════════════════════════════════════════════════

func RegisterTenantSaaS(c *fiber.Ctx) error {
	var body struct {
		PesantrenNama string `json:"pesantren_nama"`
		PendaftarNama string `json:"pendaftar_nama"`
		Kontak        string `json:"kontak"`
		Subdomain     string `json:"subdomain"`
		Paket         string `json:"paket"` // "trial", "standart", "premium"
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}

	if body.Subdomain == "" || body.PesantrenNama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Nama pesantren dan subdomain wajib diisi"})
	}

	// Cek ketersediaan subdomain
	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM tenants WHERE subdomain = ?", body.Subdomain).Scan(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Subdomain sudah terpakai"})
	}

	// Validasi paket
	if body.Paket != "trial" && body.Paket != "standart" && body.Paket != "premium" {
		body.Paket = "trial"
	}

	// Setup values
	kuotaSantri := 100
	fiturSangu := false
	harga := 0
	masaAktif := time.Now().AddDate(1, 0, 0) // gratis 1 tahun
	statusAwal := "active"

	switch body.Paket {
	case "standart":
		harga = 500000 // contoh harga
		kuotaSantri = 500
		fiturSangu = false
		masaAktif = time.Now().AddDate(1, 0, 0) // 1 tahun
		statusAwal = "pending_payment"
	case "premium":
		harga = 1000000 // contoh harga
		kuotaSantri = 999999
		fiturSangu = true
		masaAktif = time.Now().AddDate(1, 0, 0) // 1 tahun
		statusAwal = "pending_payment"
	}

	// 1. Insert saas_registrations tracking (terutama untuk yang berbayar)
	resReg, err := config.DB.Exec(`INSERT INTO saas_registrations 
		(pesantren_nama, pendaftar_nama, kontak, subdomain, paket, status) 
		VALUES (?, ?, ?, ?, ?, 'pending')`,
		body.PesantrenNama, body.PendaftarNama, body.Kontak, body.Subdomain, body.Paket)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan registrasi"})
	}
	regID, _ := resReg.LastInsertId()

	var snapToken string
	orderID := fmt.Sprintf("SAAS-%d-%d", regID, time.Now().Unix())

	if harga > 0 {
		// Midtrans integration
		var serverKey string
		var isProd bool
		config.DB.QueryRow("SELECT midtrans_server_key, midtrans_is_production FROM saas_settings WHERE id = 1").Scan(&serverKey, &isProd)

		if serverKey == "" {
			return c.Status(500).JSON(fiber.Map{"message": "Gateway pembayaran belum dikonfigurasi oleh Superadmin"})
		}

		env := midtrans.Sandbox
		if isProd {
			env = midtrans.Production
		}
		var s snap.Client
		s.New(serverKey, env)

		req := &snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: int64(harga),
			},
			Items: &[]midtrans.ItemDetails{
				{
					ID:    fmt.Sprintf("SAAS-%s", body.Paket),
					Name:  fmt.Sprintf("Langganan %s", body.Paket),
					Price: int64(harga),
					Qty:   1,
				},
			},
			CustomerDetail: &midtrans.CustomerDetails{
				FName: body.PendaftarNama,
				Phone: body.Kontak,
			},
		}

		snapResp, errSnap := s.CreateTransaction(req)
		if errSnap != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat transaksi: " + errSnap.GetMessage()})
		}
		snapToken = snapResp.Token

		// Update reg tracking
		config.DB.Exec("UPDATE saas_registrations SET midtrans_order_id = ?, snap_token = ? WHERE id = ?", orderID, snapToken, regID)
	}

	// 2. Buatkan Tenant (walaupun berbayar, kita buat tenantnya tapi dengan status suspended/pending_payment)
	// Jika trial, status langsung 'active'
	res, err := config.DB.Exec("INSERT INTO tenants (nama, subdomain, status, paket_saas, kuota_santri, fitur_sangu, expired_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		body.PesantrenNama, body.Subdomain, statusAwal, body.Paket, kuotaSantri, fiturSangu, masaAktif.Format("2006-01-02"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat tenant"})
	}
	tenantID, _ := res.LastInsertId()

	// 3. Buatkan user admin awal dengan password RANDOM
	defaultPassword := generateRandomPassword(10)
	hash, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	username := fmt.Sprintf("admin_%s", body.Subdomain)

	config.DB.Exec("INSERT INTO users (tenant_id, username, password_hash, role, nama, is_active) VALUES (?, ?, ?, 'admin', ?, 1)",
		tenantID, username, string(hash), body.PendaftarNama)

	// Log password untuk kirim via WhatsApp/email oleh admin
	log.Printf("[REGISTER] Tenant '%s' (subdomain: %s) terdaftar. Username: %s, Password: %s — KIRIM KE KONTAK: %s",
		body.PesantrenNama, body.Subdomain, username, defaultPassword, body.Kontak)

	// P1: JANGAN kirim password di response API — kirim via WA/email
	return c.JSON(fiber.Map{
		"message":     "Registrasi berhasil! Kredensial login akan dikirim ke WhatsApp Anda.",
		"snap_token":  snapToken,
		"username":    username,
		"status_awal": statusAwal,
		"paket":       body.Paket,
	})
}

// ═══════════════════════════════════════════════════════════
// PUBLIC: POST /api/saas/midtrans-callback (Midtrans Webhook)
// ═══════════════════════════════════════════════════════════

func SaaSWebhook(c *fiber.Ctx) error {
	var notif map[string]interface{}
	if err := json.Unmarshal(c.Body(), &notif); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid JSON"})
	}

	orderID, _ := notif["order_id"].(string)
	if orderID == "" {
		return c.Status(400).JSON(fiber.Map{"message": "order_id missing"})
	}

	var regID int
	var status string
	var subdomain string
	err := config.DB.QueryRow("SELECT id, status, subdomain FROM saas_registrations WHERE midtrans_order_id = ?", orderID).Scan(&regID, &status, &subdomain)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Registration not found"})
	}

	// Verify signature
	var serverKey string
	config.DB.QueryRow("SELECT midtrans_server_key FROM saas_settings WHERE id = 1").Scan(&serverKey)

	// P1: Signature verification WAJIB — jangan skip
	sigKey, _ := notif["signature_key"].(string)
	if sigKey == "" {
		return c.Status(403).JSON(fiber.Map{"message": "Missing signature_key — webhook rejected"})
	}
	statusCode, _ := notif["status_code"].(string)
	grossAmount, _ := notif["gross_amount"].(string)
	raw := orderID + statusCode + grossAmount + serverKey
	h := sha512.Sum512([]byte(raw))
	expectedSig := fmt.Sprintf("%x", h)
	if sigKey != expectedSig {
		return c.Status(403).JSON(fiber.Map{"message": "Invalid signature"})
	}

	transactionStatus, _ := notif["transaction_status"].(string)

	if transactionStatus == "capture" || transactionStatus == "settlement" {
		config.DB.Exec("UPDATE saas_registrations SET status = 'paid' WHERE id = ?", regID)
		config.DB.Exec("UPDATE tenants SET status = 'active' WHERE subdomain = ?", subdomain)
	} else if transactionStatus == "expire" || transactionStatus == "cancel" || transactionStatus == "deny" {
		config.DB.Exec("UPDATE saas_registrations SET status = 'failed' WHERE id = ?", regID)
		config.DB.Exec("UPDATE tenants SET status = 'suspended' WHERE subdomain = ?", subdomain)
	}

	return c.JSON(fiber.Map{"message": "OK"})
}

// ═══════════════════════════════════════════════════════════
// SUPERADMIN: GET /api/saas/settings
// ═══════════════════════════════════════════════════════════

func GetSaaSSettings(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "superadmin" {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var s struct {
		ServerKey    string `json:"midtrans_server_key"`
		ClientKey    string `json:"midtrans_client_key"`
		IsProduction bool   `json:"midtrans_is_production"`
	}
	config.DB.QueryRow("SELECT midtrans_server_key, midtrans_client_key, midtrans_is_production FROM saas_settings WHERE id = 1").Scan(&s.ServerKey, &s.ClientKey, &s.IsProduction)
	return c.JSON(s)
}

// ═══════════════════════════════════════════════════════════
// SUPERADMIN: PUT /api/saas/settings
// ═══════════════════════════════════════════════════════════

func UpdateSaaSSettings(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "superadmin" {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var body struct {
		ServerKey    string `json:"midtrans_server_key"`
		ClientKey    string `json:"midtrans_client_key"`
		IsProduction bool   `json:"midtrans_is_production"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid body"})
	}

	_, err := config.DB.Exec("UPDATE saas_settings SET midtrans_server_key=?, midtrans_client_key=?, midtrans_is_production=? WHERE id=1",
		body.ServerKey, body.ClientKey, body.IsProduction)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update saas settings"})
	}

	return c.JSON(fiber.Map{"message": "Tersimpan"})
}

// ═══════════════════════════════════════════════════════════
// SUPERADMIN: PUT /api/saas/tenants/:id/fitur
// ═══════════════════════════════════════════════════════════

func UpdateTenantSaaS(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "superadmin" {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	tenantID := c.Params("id")
	var body struct {
		Paket       string `json:"paket_saas"`
		KuotaSantri int    `json:"kuota_santri"`
		FiturSangu  bool   `json:"fitur_sangu"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid body"})
	}

	_, err := config.DB.Exec("UPDATE tenants SET paket_saas=?, kuota_santri=?, fitur_sangu=? WHERE id=?",
		body.Paket, body.KuotaSantri, body.FiturSangu, tenantID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update tenant"})
	}

	return c.JSON(fiber.Map{"message": "Pengaturan fitur tersimpan"})
}
