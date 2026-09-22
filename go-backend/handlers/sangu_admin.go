package handlers

import (
	"database/sql"
	"pesantren-multi/config"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Admin Pesantren: Get all merchants
func GetSanguMerchants(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama, pemilik, kontak, saldo, status FROM sangu_merchants WHERE tenant_id = ?", tid)
	if err != nil {
		return c.JSON([]fiber.Map{})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id int
		var saldo int64
		var nama, pemilik, kontak, status sql.NullString
		rows.Scan(&id, &nama, &pemilik, &kontak, &saldo, &status)
		list = append(list, fiber.Map{
			"id":        id,
			"nama":      nama.String,
			"pemilik":   pemilik.String,
			"kontak":    kontak.String,
			"saldo":     saldo,
			"status":    status.String,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// Admin Pesantren: Create merchant
func CreateSanguMerchant(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama    string `json:"nama"`
		Pemilik string `json:"pemilik"`
		Kontak  string `json:"kontak"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}

	res, err := config.DB.Exec("INSERT INTO sangu_merchants (tenant_id, nama, pemilik, kontak) VALUES (?, ?, ?, ?)", tid, body.Nama, body.Pemilik, body.Kontak)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat merchant"})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"message": "Merchant berhasil dibuat", "id": id})
}

func UpdateSanguMerchant(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id, _ := strconv.Atoi(c.Params("id"))
	var body struct {
		Nama    string `json:"nama"`
		Pemilik string `json:"pemilik"`
		Kontak  string `json:"kontak"`
		Status  string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}

	_, err := config.DB.Exec("UPDATE sangu_merchants SET nama=?, pemilik=?, kontak=?, status=? WHERE id=? AND tenant_id=?", body.Nama, body.Pemilik, body.Kontak, body.Status, id, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update merchant"})
	}
	return c.JSON(fiber.Map{"message": "Merchant berhasil diupdate"})
}

// Admin Pesantren: Get Pending Withdrawals
func GetSanguWithdrawals(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query(`
		SELECT w.id, w.merchant_id, m.nama, w.nominal, w.status, w.keterangan, w.created_at 
		FROM sangu_withdrawals w
		JOIN sangu_merchants m ON w.merchant_id = m.id
		WHERE w.tenant_id = ? ORDER BY w.id DESC`, tid)
	if err != nil {
		return c.JSON([]fiber.Map{})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, merchant_id int
		var nominal int64
		var m_nama, status, keterangan, created_at string
		rows.Scan(&id, &merchant_id, &m_nama, &nominal, &status, &keterangan, &created_at)
		list = append(list, fiber.Map{
			"id":            id,
			"merchant_id":   merchant_id,
			"merchant_nama": m_nama,
			"nominal":       nominal,
			"status":        status,
			"keterangan":    keterangan,
			"created_at":    created_at,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// Admin Pesantren: Approve/Reject Withdrawal
func ProcessSanguWithdrawal(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	id, _ := strconv.Atoi(c.Params("id"))

	var body struct {
		Action string `json:"action"` // "approve" or "reject"
		Status string `json:"status"`
	}
	c.BodyParser(&body)
	if body.Action == "" { body.Action = body.Status }

	var merchant_id int
	var nominal int64
	var status string
	err := config.DB.QueryRow("SELECT merchant_id, nominal, status FROM sangu_withdrawals WHERE id=? AND tenant_id=?", id, tid).Scan(&merchant_id, &nominal, &status)
	if err != nil || status != "pending" {
		return c.Status(400).JSON(fiber.Map{"message": "Withdrawal tidak valid atau sudah diproses"})
	}

	if body.Action == "approve" || body.Action == "sukses" {
		// Deduk merchant saldo and update withdrawal
		tx, _ := config.DB.Begin()
		_, err1 := tx.Exec("UPDATE sangu_merchants SET saldo = saldo - ? WHERE id=? AND tenant_id=? AND saldo >= ?", nominal, merchant_id, tid, nominal)
		if err1 != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"message": "Gagal proses, mungkin saldo tidak cukup"})
		}
		_, err2 := tx.Exec("UPDATE sangu_withdrawals SET status='sukses', approved_by=?, updated_at=NOW() WHERE id=?", uid, id)
		if err2 != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"message": "Gagal proses withdrawal"})
		}
		tx.Commit()
		return c.JSON(fiber.Map{"message": "Penarikan berhasil disetujui"})
	} else if body.Action == "reject" || body.Action == "ditolak" {
		config.DB.Exec("UPDATE sangu_withdrawals SET status='ditolak', approved_by=?, updated_at=NOW() WHERE id=?", uid, id)
		return c.JSON(fiber.Map{"message": "Penarikan ditolak"})
	}

	return c.Status(400).JSON(fiber.Map{"message": "Aksi tidak valid"})
}
