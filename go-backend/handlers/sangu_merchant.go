package handlers

import (
	"database/sql"
	"pesantren-multi/config"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Helper: Ensure user belongs to a merchant
func getMerchantID(c *fiber.Ctx) (int, error) {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var mid int
	err := config.DB.QueryRow("SELECT COALESCE(merchant_id, 0) FROM users WHERE id=? AND tenant_id=?", uid, tid).Scan(&mid)
	if err != nil || mid == 0 {
		return 0, fiber.ErrForbidden
	}
	return mid, nil
}

// Merchant: Get Dashboard Info
func GetMerchantDashboard(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Anda bukan pengelola merchant"})
	}

	var saldo int64
	var nama sql.NullString
	config.DB.QueryRow("SELECT nama, saldo FROM sangu_merchants WHERE id=? AND tenant_id=?", mid, tid).Scan(&nama, &saldo)

	var produkCount int
	config.DB.QueryRow("SELECT COUNT(id) FROM sangu_products WHERE merchant_id=? AND tenant_id=?", mid, tid).Scan(&produkCount)

	rows, _ := config.DB.Query("SELECT id, total, created_at, santri_id FROM sangu_transaksi WHERE merchant_id=? AND tenant_id=? ORDER BY id DESC LIMIT 5", mid, tid)
	var transaksi []fiber.Map
	for rows != nil && rows.Next() {
		var trxId, santriId int
		var trxTotal int64
		var trxDate string
		rows.Scan(&trxId, &trxTotal, &trxDate, &santriId)
		transaksi = append(transaksi, fiber.Map{"id": trxId, "total": trxTotal, "created_at": trxDate, "santri_id": santriId})
	}
	if rows != nil {
		rows.Close()
	}
	if transaksi == nil {
		transaksi = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"merchant": fiber.Map{
			"nama":  nama.String,
			"saldo": saldo,
		},
		"produk_count":       produkCount,
		"transaksi_terakhir": transaksi,
	})
}

// Merchant: Get Products
func GetMerchantProducts(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Anda bukan pengelola merchant"})
	}

	rows, _ := config.DB.Query("SELECT id, nama, harga, stok, COALESCE(barcode_sku,'') FROM sangu_products WHERE merchant_id=? AND tenant_id=?", mid, tid)
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, stok int
		var harga int64
		var nama, barcode string
		rows.Scan(&id, &nama, &harga, &stok, &barcode)
		list = append(list, fiber.Map{
			"id":      id,
			"nama":    nama,
			"harga":   harga,
			"stok":    stok,
			"barcode": barcode,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateMerchantProduct(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var body struct {
		Nama    string `json:"nama"`
		Harga   int64  `json:"harga"`
		Stok    int    `json:"stok"`
		Barcode string `json:"barcode"`
	}
	c.BodyParser(&body)

	res, err := config.DB.Exec("INSERT INTO sangu_products (tenant_id, merchant_id, nama, harga, stok, barcode_sku) VALUES (?, ?, ?, ?, ?, ?)", tid, mid, body.Nama, body.Harga, body.Stok, body.Barcode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menambah produk"})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"message": "Produk ditambahkan", "id": id})
}

func UpdateMerchantProduct(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	id, _ := strconv.Atoi(c.Params("id"))

	var body struct {
		Nama    string `json:"nama"`
		Harga   int64  `json:"harga"`
		Stok    int    `json:"stok"`
		Barcode string `json:"barcode"`
	}
	c.BodyParser(&body)

	_, err = config.DB.Exec("UPDATE sangu_products SET nama=?, harga=?, stok=?, barcode_sku=? WHERE id=? AND merchant_id=? AND tenant_id=?", body.Nama, body.Harga, body.Stok, body.Barcode, id, mid, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update produk"})
	}
	return c.JSON(fiber.Map{"message": "Produk diupdate"})
}

func DeleteMerchantProduct(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	id, _ := strconv.Atoi(c.Params("id"))
	config.DB.Exec("DELETE FROM sangu_products WHERE id=? AND merchant_id=? AND tenant_id=?", id, mid, tid)
	return c.JSON(fiber.Map{"message": "Produk dihapus"})
}

func RequestWithdrawal(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var body struct {
		Nominal    int64  `json:"nominal"`
		Keterangan string `json:"keterangan"`
	}
	c.BodyParser(&body)

	var saldo int64
	config.DB.QueryRow("SELECT saldo FROM sangu_merchants WHERE id=? AND tenant_id=?", mid, tid).Scan(&saldo)
	if saldo < body.Nominal {
		return c.Status(400).JSON(fiber.Map{"message": "Saldo merchant tidak mencukupi"})
	}

	_, err = config.DB.Exec("INSERT INTO sangu_withdrawals (tenant_id, merchant_id, nominal, status, keterangan) VALUES (?, ?, ?, 'pending', ?)", tid, mid, body.Nominal, body.Keterangan)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengajukan penarikan"})
	}

	return c.JSON(fiber.Map{"message": "Pengajuan penarikan berhasil"})
}

func GetMerchantWithdrawals(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	mid, err := getMerchantID(c)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	rows, _ := config.DB.Query("SELECT id, nominal, status, keterangan, created_at FROM sangu_withdrawals WHERE merchant_id=? AND tenant_id=? ORDER BY id DESC", mid, tid)
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id int
		var nominal int64
		var status, ket, createdAt string
		rows.Scan(&id, &nominal, &status, &ket, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "nominal": nominal, "status": status, "keterangan": ket, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}
