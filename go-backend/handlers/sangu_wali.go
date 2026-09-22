package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

func WaliGetSanguInfo(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	sid := c.Params("santri_id")

	// Validate ownership
	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
	}

	var saldo, limitHarian int64
	err := config.DB.QueryRow("SELECT COALESCE(sangu_saldo,0), COALESCE(sangu_limit_harian,0) FROM santri WHERE id=? AND tenant_id=?", sid, tid).Scan(&saldo, &limitHarian)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"saldo":        saldo,
		"limit_harian": limitHarian,
	})
}

func WaliSetLimitHarian(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	sid := c.Params("santri_id")

	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
	}

	var body struct {
		Limit int64 `json:"limit_harian"`
	}
	c.BodyParser(&body)

	config.DB.Exec("UPDATE santri SET sangu_limit_harian=? WHERE id=? AND tenant_id=?", body.Limit, sid, tid)
	return c.JSON(fiber.Map{"message": "Limit harian berhasil diatur"})
}

func WaliGetRiwayatSangu(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	sid := c.Params("santri_id")

	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
	}

	var riwayat []fiber.Map

	// Transaksi belanja
	rows, _ := config.DB.Query("SELECT t.total, t.created_at, m.nama FROM sangu_transaksi t JOIN sangu_merchants m ON t.merchant_id=m.id WHERE t.santri_id=? AND t.tenant_id=? AND t.status='sukses' ORDER BY t.id DESC LIMIT 50", sid, tid)
	defer rows.Close()
	for rows.Next() {
		var total int64
		var createdAt, mNama string
		rows.Scan(&total, &createdAt, &mNama)
		riwayat = append(riwayat, fiber.Map{
			"tipe":       "belanja",
			"nominal":    total,
			"keterangan": "Belanja di " + mNama,
			"created_at": createdAt,
		})
	}

	// Topup
	rows2, _ := config.DB.Query("SELECT nominal, status, payment_method, created_at FROM sangu_topup WHERE santri_id=? AND tenant_id=? AND status='sukses' ORDER BY id DESC LIMIT 50", sid, tid)
	defer rows2.Close()
	for rows2.Next() {
		var nominal int64
		var status, method, createdAt string
		rows2.Scan(&nominal, &status, &method, &createdAt)
		riwayat = append(riwayat, fiber.Map{
			"tipe":       "topup",
			"nominal":    nominal,
			"keterangan": "Topup via " + method,
			"created_at": createdAt,
		})
	}

	if riwayat == nil {
		riwayat = []fiber.Map{}
	}
	return c.JSON(riwayat)
}

func WaliCreateTopupSangu(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	var body struct {
		SantriID int   `json:"santri_id"`
		Nominal  int64 `json:"nominal"`
	}
	c.BodyParser(&body)

	if body.Nominal < 10000 {
		return c.Status(400).JSON(fiber.Map{"message": "Minimal topup Rp 10.000"})
	}

	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", body.SantriID, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var santriNama string
	config.DB.QueryRow("SELECT nama FROM santri WHERE id = ?", body.SantriID).Scan(&santriNama)

	pg, err := getPGSettings(tid)
	if err != nil || pg.Provider == "" || pg.ServerKey == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Payment gateway belum dikonfigurasi"})
	}

	fee := pg.FeeFlat
	totalPay := body.Nominal + fee

	orderID := fmt.Sprintf("SANGU%d-%d-%d", tid, body.SantriID, time.Now().UnixMilli())
	
	var waliNama string
	config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", uid).Scan(&waliNama)

	var snapToken, redirectURL string

	switch pg.Provider {
	case "midtrans":
		env := midtrans.Sandbox
		if pg.IsProduction {
			env = midtrans.Production
		}
		var s snap.Client
		s.New(pg.ServerKey, env)

		items := []midtrans.ItemDetails{
			{
				ID:    "SANGU",
				Name:  "Topup Sangu - " + santriNama,
				Price: body.Nominal,
				Qty:   1,
			},
		}

		if fee > 0 {
			items = append(items, midtrans.ItemDetails{
				ID:    "FEE",
				Name:  "Biaya Layanan",
				Price: fee,
				Qty:   1,
			})
		}

		req := &snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: totalPay,
			},
			Items: &items,
			CustomerDetail: &midtrans.CustomerDetails{
				FName: waliNama,
			},
			Callbacks: &snap.Callbacks{
				Finish: c.BaseURL() + "/app#sangu-wali",
			},
		}

		snapResp, errSnap := s.CreateTransaction(req)
		if errSnap != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat transaksi"})
		}
		snapToken = snapResp.Token
		redirectURL = snapResp.RedirectURL

	default:
		return c.Status(400).JSON(fiber.Map{"message": "Provider gateway '" + pg.Provider + "' belum didukung"})
	}

	_, errIns := config.DB.Exec("INSERT INTO sangu_topup (tenant_id, santri_id, nominal, fee, total, status, payment_method, payment_ref, payment_url) VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?)",
		tid, body.SantriID, body.Nominal, fee, totalPay, pg.Provider, orderID, redirectURL)
	
	if errIns != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan data topup"})
	}

	return c.JSON(fiber.Map{
		"snap_token":   snapToken,
		"redirect_url": redirectURL,
		"order_id":     orderID,
		"nominal":      body.Nominal,
		"fee":          fee,
		"total_bayar":  totalPay,
	})
}
