package handlers

import (
	"pesantren-multi/config"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// normalizeCardUID cleans up RFID/NFC card UID input:
// strips spaces, colons, dashes and uppercases the hex string
func normalizeCardUID(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, ":", "")
	raw = strings.ReplaceAll(raw, "-", "")
	return strings.ToUpper(raw)
}

// KasirGetSantriByCard — lookup santri by RFID/NFC card UID
// Accepts ?barcode=xxx (backward compatible query param name)
// 1) Try card_uid match first
// 2) Fallback: try parsing as SANTRI-123 or plain ID (transition period)
func KasirGetSantriByCard(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	raw := c.Query("barcode") // reader sends card UID here
	if raw == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Input kartu tidak boleh kosong"})
	}

	cardUID := normalizeCardUID(raw)

	var santriID int
	var nama, kamar, kelas string
	var saldo, limitHarian int64
	var foundByCard bool

	// 1) Try lookup by card_uid
	err := config.DB.QueryRow(`
		SELECT s.id, s.nama, COALESCE(k.nama,'-'), COALESCE(s.kelas_diniyyah,'-'), 
		       COALESCE(s.sangu_saldo,0), COALESCE(s.sangu_limit_harian,0) 
		FROM santri s 
		LEFT JOIN kamar k ON s.kamar_id = k.id 
		WHERE s.card_uid=? AND s.tenant_id=? AND s.status='aktif'`, cardUID, tid).
		Scan(&santriID, &nama, &kamar, &kelas, &saldo, &limitHarian)

	if err == nil {
		foundByCard = true
	}

	// 2) Fallback: try SANTRI-123 or plain numeric ID (for transition)
	if !foundByCard {
		idStr := strings.ReplaceAll(raw, "SANTRI-", "")
		idStr = strings.TrimSpace(idStr)
		fallbackID, parseErr := strconv.Atoi(idStr)
		if parseErr != nil {
			return c.Status(404).JSON(fiber.Map{
				"message": "Kartu tidak terdaftar. Silakan registrasi kartu terlebih dahulu.",
			})
		}

		err = config.DB.QueryRow(`
			SELECT s.id, s.nama, COALESCE(k.nama,'-'), COALESCE(s.kelas_diniyyah,'-'), 
			       COALESCE(s.sangu_saldo,0), COALESCE(s.sangu_limit_harian,0) 
			FROM santri s 
			LEFT JOIN kamar k ON s.kamar_id = k.id 
			WHERE s.id=? AND s.tenant_id=? AND s.status='aktif'`, fallbackID, tid).
			Scan(&santriID, &nama, &kamar, &kelas, &saldo, &limitHarian)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
		}
	}

	// Calculate daily spending limit
	var terpakaiHariIni int64
	config.DB.QueryRow("SELECT COALESCE(SUM(total),0) FROM sangu_transaksi WHERE santri_id=? AND tenant_id=? AND DATE(created_at) = CURDATE()", santriID, tid).Scan(&terpakaiHariIni)

	sisaLimit := limitHarian - terpakaiHariIni
	if sisaLimit < 0 {
		sisaLimit = 0
	}

	return c.JSON(fiber.Map{
		"id":                santriID,
		"nama":              nama,
		"kamar":             kamar,
		"kelas":             kelas,
		"saldo":             saldo,
		"limit_harian":      limitHarian,
		"terpakai_hari_ini": terpakaiHariIni,
		"sisa_limit":        sisaLimit,
	})
}

// RegisterCard — assign RFID/NFC card UID to a santri
// PUT /api/sangu/register-card
// Body: { "santri_id": 123, "card_uid": "04A3B2C1" }
func RegisterCard(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)

	var body struct {
		SantriID int    `json:"santri_id"`
		CardUID  string `json:"card_uid"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format data tidak valid"})
	}

	cardUID := normalizeCardUID(body.CardUID)
	if cardUID == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Card UID tidak boleh kosong"})
	}
	if body.SantriID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Santri ID wajib diisi"})
	}

	// Verify santri exists
	var existName string
	err := config.DB.QueryRow("SELECT nama FROM santri WHERE id=? AND tenant_id=?", body.SantriID, tid).Scan(&existName)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}

	// Check if card is already assigned to another santri
	var otherID int
	var otherName string
	err = config.DB.QueryRow("SELECT id, nama FROM santri WHERE card_uid=? AND tenant_id=? AND id!=?", cardUID, tid, body.SantriID).Scan(&otherID, &otherName)
	if err == nil {
		return c.Status(409).JSON(fiber.Map{
			"message": "Kartu ini sudah terdaftar untuk santri lain: " + otherName,
		})
	}

	// Assign card
	_, err = config.DB.Exec("UPDATE santri SET card_uid=? WHERE id=? AND tenant_id=?", cardUID, body.SantriID, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan kartu: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Kartu RFID berhasil didaftarkan untuk " + existName,
		"card_uid": cardUID,
	})
}

// UnregisterCard — remove RFID/NFC card from a santri
// DELETE /api/sangu/register-card/:santri_id
func UnregisterCard(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	santriID := c.Params("santri_id")

	_, err := config.DB.Exec("UPDATE santri SET card_uid=NULL WHERE id=? AND tenant_id=?", santriID, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus kartu"})
	}

	return c.JSON(fiber.Map{"message": "Kartu RFID berhasil dicopot"})
}

// Process POS Transaction
func KasirCheckout(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	// Kasir must belong to a merchant
	var mid int
	err := config.DB.QueryRow("SELECT COALESCE(merchant_id, 0) FROM users WHERE id=? AND tenant_id=?", uid, tid).Scan(&mid)
	if err != nil || mid == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak. Anda bukan kasir."})
	}

	var body struct {
		SantriID int   `json:"santri_id"`
		Total    int64 `json:"total"`
		Items    []struct {
			ProductID int   `json:"product_id"`
			Qty       int   `json:"qty"`
			Harga     int64 `json:"harga"`
		} `json:"items"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format data tidak valid"})
	}

	if len(body.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Keranjang kosong"})
	}

	// Fetch actual prices and calculate total on backend
	var calculatedTotal int64
	for i, item := range body.Items {
		var actualPrice int64
		var currentStok int
		err := config.DB.QueryRow("SELECT harga, stok FROM sangu_products WHERE id=? AND merchant_id=? AND tenant_id=?", item.ProductID, mid, tid).Scan(&actualPrice, &currentStok)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Produk tidak valid"})
		}
		if currentStok < item.Qty {
			return c.Status(400).JSON(fiber.Map{"message": "Stok tidak mencukupi untuk salah satu produk"})
		}
		body.Items[i].Harga = actualPrice
		calculatedTotal += actualPrice * int64(item.Qty)
	}

	body.Total = calculatedTotal

	// Validation
	var saldo, limitHarian int64
	err = config.DB.QueryRow("SELECT COALESCE(sangu_saldo,0), COALESCE(sangu_limit_harian,0) FROM santri WHERE id=? AND tenant_id=?", body.SantriID, tid).Scan(&saldo, &limitHarian)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak valid"})
	}

	if saldo < body.Total {
		return c.Status(400).JSON(fiber.Map{"message": "Saldo santri tidak mencukupi"})
	}

	var terpakaiHariIni int64
	config.DB.QueryRow("SELECT COALESCE(SUM(total),0) FROM sangu_transaksi WHERE santri_id=? AND tenant_id=? AND DATE(created_at) = CURDATE() AND status='sukses'", body.SantriID, tid).Scan(&terpakaiHariIni)

	if limitHarian > 0 && (terpakaiHariIni+body.Total > limitHarian) {
		return c.Status(400).JSON(fiber.Map{"message": "Transaksi melebihi limit harian santri"})
	}

	// Execute transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Database error"})
	}

	res, err := tx.Exec("INSERT INTO sangu_transaksi (tenant_id, santri_id, merchant_id, kasir_user_id, total, status) VALUES (?, ?, ?, ?, ?, 'sukses')", tid, body.SantriID, mid, uid, body.Total)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan transaksi"})
	}
	trxID, _ := res.LastInsertId()

	for _, item := range body.Items {
		subtotal := int64(item.Qty) * item.Harga
		tx.Exec("INSERT INTO sangu_transaksi_items (transaksi_id, product_id, qty, harga, subtotal) VALUES (?, ?, ?, ?, ?)", trxID, item.ProductID, item.Qty, item.Harga, subtotal)
		// Reduce stock
		tx.Exec("UPDATE sangu_products SET stok = stok - ? WHERE id=? AND stok >= ?", item.Qty, item.ProductID, item.Qty)
	}

	// Deduct Santri Balance
	_, err = tx.Exec("UPDATE santri SET sangu_saldo = sangu_saldo - ? WHERE id=? AND sangu_saldo >= ?", body.Total, body.SantriID, body.Total)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memotong saldo"})
	}

	// Add Merchant Balance
	_, err = tx.Exec("UPDATE sangu_merchants SET saldo = saldo + ? WHERE id=?", body.Total, mid)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menambah saldo merchant"})
	}

	tx.Commit()
	return c.JSON(fiber.Map{"message": "Transaksi berhasil"})
}
