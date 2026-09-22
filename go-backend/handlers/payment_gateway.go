package handlers

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"pesantren-multi/config"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

// ═══════════════════════════════════════════════════════════
// HELPER: Read payment gateway settings for a tenant
// ═══════════════════════════════════════════════════════════

type pgSettings struct {
	Provider     string
	ServerKey    string
	ClientKey    string
	IsProduction bool
	FeePercent   float64
	FeeFlat      int64
}

func getPGSettings(tenantID int) (*pgSettings, error) {
	var s pgSettings
	var isProdStr, feePercentStr, feeFlatStr string
	err := config.DB.QueryRow(`SELECT COALESCE(pg_provider,''), COALESCE(pg_server_key,''), 
		COALESCE(pg_client_key,''), COALESCE(pg_is_production,'0'), 
		COALESCE(pg_fee_percent,'0'), COALESCE(pg_fee_flat,'0') 
		FROM settings WHERE tenant_id = ? LIMIT 1`, tenantID).
		Scan(&s.Provider, &s.ServerKey, &s.ClientKey, &isProdStr, &feePercentStr, &feeFlatStr)
	if err != nil {
		return nil, err
	}
	isProd, _ := strconv.Atoi(isProdStr)
	s.IsProduction = isProd == 1
	s.FeePercent, _ = strconv.ParseFloat(feePercentStr, 64)
	s.FeeFlat, _ = strconv.ParseInt(feeFlatStr, 10, 64)
	return &s, nil
}

// reconcileItem is used by reconcileBatch to process multi-month settlements
type reconcileItem struct {
	ID, SantriID   int
	Bulan, Gateway string
	Nominal        int64
}

// ═══════════════════════════════════════════════════════════
// POST /api/wali/pay — Wali creates online payment
// ═══════════════════════════════════════════════════════════

func WaliCreatePayment(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	var body struct {
		SantriID     int      `json:"santri_id"`
		Bulan        string   `json:"bulan"`      // single month (backward compat)
		BulanList    []string `json:"bulan_list"` // multi-month
		NominalBebas int64    `json:"nominal_bebas"`
	}
	c.BodyParser(&body)

	// Support both single, multi-month, and nominal_bebas
	var bulanListToProcess []string
	if body.NominalBebas > 0 {
		// generate last 36 months just to be safe, oldest first
		now := time.Now()
		for i := 36; i >= 0; i-- {
			t := now.AddDate(0, -i, 0)
			bulan := t.Format("2006-01")
			tagihan := HitungTagihan(tid, body.SantriID, bulan)
			if tagihan <= 0 {
				continue
			}
			var totalBayar int64
			config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?",
				tid, body.SantriID, bulan).Scan(&totalBayar)
			if tagihan-totalBayar > 0 {
				bulanListToProcess = append(bulanListToProcess, bulan)
			}
		}
	} else {
		bulanListToProcess = body.BulanList
		if len(bulanListToProcess) == 0 && body.Bulan != "" {
			bulanListToProcess = []string{body.Bulan}
		}
	}

	if body.SantriID == 0 || len(bulanListToProcess) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "santri_id dan bulan/nominal wajib diisi"})
	}

	// Validate santri belongs to this wali
	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?",
		body.SantriID, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
	}

	// Get santri name
	var santriNama string
	config.DB.QueryRow("SELECT nama FROM santri WHERE id = ?", body.SantriID).Scan(&santriNama)

	// Calculate total sisa across all months (tagihan varies per bulan)
	type bulanSisa struct {
		Bulan string
		Sisa  int64
	}
	var bulanItems []bulanSisa
	var totalSisa int64
	var nominalBebasSisa = body.NominalBebas

	for _, bulan := range bulanListToProcess {
		tagihan := HitungTagihan(tid, body.SantriID, bulan)
		if tagihan <= 0 {
			continue
		}

		var totalBayar int64
		config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?",
			tid, body.SantriID, bulan).Scan(&totalBayar)
		sisa := tagihan - totalBayar
		if sisa <= 0 {
			continue // already paid
		}

		nominalUntukBulanIni := sisa
		if body.NominalBebas > 0 {
			if nominalBebasSisa <= 0 {
				break
			}
			if nominalBebasSisa < sisa {
				nominalUntukBulanIni = nominalBebasSisa
			}
			nominalBebasSisa -= nominalUntukBulanIni
		}

		bulanItems = append(bulanItems, bulanSisa{Bulan: bulan, Sisa: nominalUntukBulanIni})
		totalSisa += nominalUntukBulanIni

		// Expire old pending transactions
		config.DB.Exec(`UPDATE payment_transactions SET transaction_status = 'expire' 
			WHERE tenant_id = ? AND santri_id = ? AND bulan = ? AND transaction_status = 'pending'`,
			tid, body.SantriID, bulan)
	}

	if len(bulanItems) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Semua tagihan sudah lunas"})
	}

	// Get payment gateway settings
	pg, err := getPGSettings(tid)
	if err != nil || pg.Provider == "" || pg.ServerKey == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Payment gateway belum dikonfigurasi oleh admin pesantren"})
	}

	// Calculate fee on total
	fee := int64(math.Ceil(float64(totalSisa)*pg.FeePercent/100.0)) + pg.FeeFlat
	totalPay := totalSisa + fee

	// Generate unique order ID (use "MULTI" when >1 month)
	bulanLabel := bulanItems[0].Bulan
	if len(bulanItems) > 1 {
		bulanLabel = fmt.Sprintf("MULTI%d", len(bulanItems))
	}
	orderID := fmt.Sprintf("PES%d-%d-%s-%d", tid, body.SantriID, bulanLabel, time.Now().UnixMilli())

	// Get wali name
	var waliNama string
	config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", uid).Scan(&waliNama)

	// ── Create transaction based on gateway provider ──
	var snapToken, redirectURL string

	switch pg.Provider {
	case "midtrans":
		env := midtrans.Sandbox
		if pg.IsProduction {
			env = midtrans.Production
		}
		var s snap.Client
		s.New(pg.ServerKey, env)

		// Build item details — one per month
		items := []midtrans.ItemDetails{}
		for _, bi := range bulanItems {
			items = append(items, midtrans.ItemDetails{
				ID:    fmt.Sprintf("SPP-%s", bi.Bulan),
				Name:  fmt.Sprintf("SPP %s - %s", bi.Bulan, santriNama),
				Price: bi.Sisa,
				Qty:   1,
			})
		}

		// Add fee as separate item if > 0
		if fee > 0 {
			items = append(items, midtrans.ItemDetails{
				ID:    "FEE",
				Name:  "Biaya Admin",
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
				Finish: c.BaseURL() + "/app#pembayaran-wali",
			},
		}

		snapResp, errSnap := s.CreateTransaction(req)
		if errSnap != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat transaksi: " + errSnap.GetMessage()})
		}
		snapToken = snapResp.Token
		redirectURL = snapResp.RedirectURL

	default:
		return c.Status(400).JSON(fiber.Map{"message": "Provider gateway '" + pg.Provider + "' belum didukung"})
	}

	// Save to payment_transactions — one row per month, same order_id
	log.Printf("[PAY-CREATE] Inserting %d rows for order_id=%s", len(bulanItems), orderID)
	for _, bi := range bulanItems {
		feePerMonth := int64(0) // fee only on first row to avoid double-counting
		if bi.Bulan == bulanItems[0].Bulan {
			feePerMonth = fee
		}
		_, errIns := config.DB.Exec(`INSERT INTO payment_transactions 
			(tenant_id, santri_id, bulan, nominal, fee, total_bayar, order_id, gateway, snap_token, transaction_status) 
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			tid, body.SantriID, bi.Bulan, bi.Sisa, feePerMonth, bi.Sisa+feePerMonth, orderID, pg.Provider, snapToken, "pending")
		if errIns != nil {
			log.Printf("[PAY-CREATE] ERROR inserting bulan=%s: %v", bi.Bulan, errIns)
		} else {
			log.Printf("[PAY-CREATE] OK inserted bulan=%s nominal=%d", bi.Bulan, bi.Sisa)
		}
	}

	return c.JSON(fiber.Map{
		"snap_token":   snapToken,
		"redirect_url": redirectURL,
		"order_id":     orderID,
		"nominal":      totalSisa,
		"fee":          fee,
		"total_bayar":  totalPay,
	})
}

// ═══════════════════════════════════════════════════════════
// POST /api/payment/notification — Webhook from gateway
// ═══════════════════════════════════════════════════════════

func PaymentNotification(c *fiber.Ctx) error {
	var notif map[string]interface{}
	if err := json.Unmarshal(c.Body(), &notif); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid JSON"})
	}

	orderID, _ := notif["order_id"].(string)
	if orderID == "" {
		return c.Status(400).JSON(fiber.Map{"message": "order_id missing"})
	}

	// Find ALL transactions for this order_id (multi-month support)
	rows, err := config.DB.Query(`SELECT id, tenant_id, santri_id, bulan, nominal, gateway 
		FROM payment_transactions WHERE order_id = ?`, orderID)
	if err != nil {
		// Try sangu_topup
		return HandleSanguWebhook(c, notif)
	}
	defer rows.Close()

	type txRow struct {
		ID, TenantID, SantriID int
		Bulan, Gateway         string
		Nominal                int64
	}
	var txRows []txRow
	for rows.Next() {
		var t txRow
		rows.Scan(&t.ID, &t.TenantID, &t.SantriID, &t.Bulan, &t.Nominal, &t.Gateway)
		txRows = append(txRows, t)
	}
	rows.Close() // Close immediately before making more DB calls
	if len(txRows) == 0 {
		return HandleSanguWebhook(c, notif)
	}

	// Use first row for signature verification
	tenantID := txRows[0].TenantID
	gateway := txRows[0].Gateway

	// Get gateway server key for signature verification
	pg, err := getPGSettings(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to read settings"})
	}

	transactionStatus, _ := notif["transaction_status"].(string)
	fraudStatus, _ := notif["fraud_status"].(string)
	paymentType, _ := notif["payment_type"].(string)

	// Verify signature for Midtrans (only when signature_key is present)
	if gateway == "midtrans" {
		sigKey, _ := notif["signature_key"].(string)
		if sigKey != "" {
			statusCode, _ := notif["status_code"].(string)
			grossAmount, _ := notif["gross_amount"].(string)
			raw := orderID + statusCode + grossAmount + pg.ServerKey
			h := sha512.Sum512([]byte(raw))
			expectedSig := fmt.Sprintf("%x", h)
			if sigKey != expectedSig {
				return c.Status(403).JSON(fiber.Map{"message": "Invalid signature"})
			}
		}
	}

	// Store the raw response on all rows
	respJSON, _ := json.Marshal(notif)

	// Process each transaction row
	log.Printf("[PAYMENT] Processing %d rows for order_id=%s, status=%s", len(txRows), orderID, transactionStatus)

	// Update gateway_response on all rows first
	for _, tx := range txRows {
		config.DB.Exec("UPDATE payment_transactions SET gateway_response = ?, payment_type = ?, updated_at = NOW() WHERE id = ?",
			string(respJSON), paymentType, tx.ID)
	}

	if transactionStatus == "capture" || transactionStatus == "settlement" {
		if transactionStatus == "capture" && fraudStatus != "" && fraudStatus != "accept" {
			for _, tx := range txRows {
				config.DB.Exec("UPDATE payment_transactions SET transaction_status = 'fraud' WHERE id = ?", tx.ID)
			}
			return c.JSON(fiber.Map{"message": "fraud detected"})
		}
		batchRows := make([]reconcileItem, len(txRows))
		for i, tx := range txRows {
			batchRows[i] = reconcileItem{tx.ID, tx.SantriID, tx.Bulan, tx.Gateway, tx.Nominal}
		}
		reconcileBatch(tenantID, batchRows, paymentType)
	} else if transactionStatus == "expire" || transactionStatus == "cancel" || transactionStatus == "deny" {
		for _, tx := range txRows {
			config.DB.Exec("UPDATE payment_transactions SET transaction_status = ? WHERE id = ?", transactionStatus, tx.ID)
		}
	} else if transactionStatus == "pending" {
		for _, tx := range txRows {
			config.DB.Exec("UPDATE payment_transactions SET transaction_status = 'pending' WHERE id = ?", tx.ID)
		}
	}

	return c.JSON(fiber.Map{"message": "OK"})
}

// reconcileBatch processes all months of a multi-month order at once.
// This avoids connection-pool exhaustion from calling DB in a tight loop.
func reconcileBatch(tenantID int, txRows []reconcileItem, paymentType string) {

	if len(txRows) == 0 {
		return
	}

	santriID := txRows[0].SantriID
	gateway := txRows[0].Gateway

	for _, tx := range txRows {
		// Idempotency: skip if already settled
		var currentStatus string
		config.DB.QueryRow("SELECT transaction_status FROM payment_transactions WHERE id = ?", tx.ID).Scan(&currentStatus)
		if currentStatus == "settlement" {
			log.Printf("[RECONCILE] txID=%d already settled, skipping", tx.ID)
			continue
		}

		// Update transaction status
		config.DB.Exec("UPDATE payment_transactions SET transaction_status = 'settlement', payment_type = ? WHERE id = ?", paymentType, tx.ID)

		if strings.HasPrefix(tx.Bulan, "INS-") {
			// Pembayaran Insidental
			tagihanIDStr := strings.TrimPrefix(tx.Bulan, "INS-")
			tagihanID, _ := strconv.Atoi(tagihanIDStr)

			// 1. Ambil sisa tagihan
			var sisa int64
			config.DB.QueryRow("SELECT sisa_tagihan FROM tagihan_insidental_santri WHERE id = ?", tagihanID).Scan(&sisa)
			sisaBaru := sisa - tx.Nominal
			status := "Belum Lunas"
			if sisaBaru <= 0 {
				status = "Lunas"
				sisaBaru = 0
			}

			// 2. Update tagihan insidental
			config.DB.Exec("UPDATE tagihan_insidental_santri SET sisa_tagihan = ?, status = ? WHERE id = ?", sisaBaru, status, tagihanID)

			// 3. Insert ke pembayaran_insidental
			config.DB.Exec("INSERT INTO pembayaran_insidental (tenant_id, santri_id, tagihan_santri_id, nominal_dibayar, metode, created_by) VALUES (?,?,?,?,?,0)",
				tenantID, santriID, tagihanID, tx.Nominal, gateway)

			// 4. Insert ke catatan_keuangan
			keteranganKas := fmt.Sprintf("Pembayaran Insidental Online via %s (%s)", gateway, paymentType)
			config.DB.Exec("INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?, 'masuk', ?, ?, ?, 0)",
				tenantID, tx.Nominal, keteranganKas, time.Now().Format("2006-01-02"))
			
			log.Printf("[RECONCILE-INS] OK pembayaran insidental txID=%d nominal=%d", tx.ID, tx.Nominal)

		} else {
			// Pembayaran SPP
			keterangan := fmt.Sprintf("Pembayaran Online via %s (%s)", gateway, paymentType)
			_, errPb := config.DB.Exec(`INSERT INTO pembayaran (tenant_id, santri_id, bulan, nominal, metode, keterangan, created_by) 
				VALUES (?,?,?,?,?,?,0)`,
				tenantID, santriID, tx.Bulan, tx.Nominal, gateway, keterangan)

			// Insert ke catatan keuangan
			config.DB.Exec("INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?, 'masuk', ?, ?, ?, 0)",
				tenantID, tx.Nominal, keterangan, time.Now().Format("2006-01-02"))

			if errPb != nil {
				log.Printf("[RECONCILE] ERROR pembayaran txID=%d bulan=%s: %v", tx.ID, tx.Bulan, errPb)
			} else {
				log.Printf("[RECONCILE] OK pembayaran txID=%d bulan=%s nominal=%d", tx.ID, tx.Bulan, tx.Nominal)
			}
		}
	}
	log.Printf("[RECONCILE] Batch done: %d rows processed for santri=%d", len(txRows), santriID)
}

// reconcileSettlement — single-month wrapper (used by WaliVerifyPayment)
func reconcileSettlement(txID, tenantID, santriID int, bulan string, nominal int64, gateway, paymentType string) {
	reconcileBatch(tenantID, []reconcileItem{{txID, santriID, bulan, gateway, nominal}}, paymentType)
}

// ═══════════════════════════════════════════════════════════
// POST /api/wali/verify-payment — Actively check Midtrans & reconcile
// ═══════════════════════════════════════════════════════════

func WaliVerifyPayment(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)

	var body struct {
		OrderID string `json:"order_id"`
	}
	c.BodyParser(&body)
	if body.OrderID == "" {
		return c.Status(400).JSON(fiber.Map{"message": "order_id wajib"})
	}

	// Find ALL transactions for this order_id (multi-month support)
	rows, err := config.DB.Query(`SELECT id, santri_id, bulan, nominal, gateway, transaction_status 
		FROM payment_transactions WHERE order_id = ? AND tenant_id = ?`, body.OrderID, tid)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	defer rows.Close()

	type txRow struct {
		ID, SantriID       int
		Bulan, Gateway     string
		Nominal            int64
		CurrentStatus      string
	}
	var txRows []txRow
	for rows.Next() {
		var t txRow
		rows.Scan(&t.ID, &t.SantriID, &t.Bulan, &t.Nominal, &t.Gateway, &t.CurrentStatus)
		txRows = append(txRows, t)
	}
	rows.Close() // Close immediately before making more DB calls
	if len(txRows) == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}

	// Check if ALL already settled
	allSettled := true
	for _, t := range txRows {
		if t.CurrentStatus != "settlement" {
			allSettled = false
			break
		}
	}
	if allSettled {
		return c.JSON(fiber.Map{"transaction_status": "settlement", "message": "Sudah lunas"})
	}

	gateway := txRows[0].Gateway

	// Get PG settings to call Midtrans API
	pg, err := getPGSettings(tid)
	if err != nil || pg.ServerKey == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Payment gateway belum dikonfigurasi"})
	}

	if gateway == "midtrans" {
		env := midtrans.Sandbox
		if pg.IsProduction {
			env = midtrans.Production
		}
		var core coreapi.Client
		core.New(pg.ServerKey, env)

		resp, errMT := core.CheckTransaction(body.OrderID)
		if errMT != nil {
			return c.JSON(fiber.Map{"transaction_status": txRows[0].CurrentStatus, "message": "Menunggu konfirmasi"})
		}

		txStatus := resp.TransactionStatus
		fraudStatus := resp.FraudStatus
		paymentType := resp.PaymentType

		if txStatus == "capture" || txStatus == "settlement" {
			if txStatus == "capture" && fraudStatus != "" && fraudStatus != "accept" {
				for _, t := range txRows {
					config.DB.Exec("UPDATE payment_transactions SET transaction_status = 'fraud' WHERE id = ?", t.ID)
				}
				return c.JSON(fiber.Map{"transaction_status": "fraud"})
			}
			// Reconcile ALL months
			for _, t := range txRows {
				reconcileSettlement(t.ID, tid, t.SantriID, t.Bulan, t.Nominal, t.Gateway, paymentType)
			}
			return c.JSON(fiber.Map{"transaction_status": "settlement", "message": "Pembayaran berhasil!"})
		} else if txStatus == "expire" || txStatus == "cancel" || txStatus == "deny" {
			for _, t := range txRows {
				config.DB.Exec("UPDATE payment_transactions SET transaction_status = ? WHERE id = ?", txStatus, t.ID)
			}
			return c.JSON(fiber.Map{"transaction_status": txStatus})
		}

		return c.JSON(fiber.Map{"transaction_status": txStatus, "message": "Menunggu pembayaran"})
	}

	return c.JSON(fiber.Map{"transaction_status": txRows[0].CurrentStatus})
}

// ═══════════════════════════════════════════════════════════
// GET /api/wali/payment-status/:order_id — Check local status
// ═══════════════════════════════════════════════════════════

func WaliPaymentStatus(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	orderID := c.Params("order_id")

	var status, paymentType string
	var nominal, fee, totalBayar int64
	err := config.DB.QueryRow(`SELECT transaction_status, COALESCE(payment_type,''), nominal, fee, total_bayar 
		FROM payment_transactions WHERE order_id = ? AND tenant_id = ?`, orderID, tid).
		Scan(&status, &paymentType, &nominal, &fee, &totalBayar)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"order_id":           orderID,
		"transaction_status": status,
		"payment_type":       paymentType,
		"nominal":            nominal,
		"fee":                fee,
		"total_bayar":        totalBayar,
	})
}

// ═══════════════════════════════════════════════════════════
// GET /api/payment/history — Admin/Bendahara view all online payments
// ═══════════════════════════════════════════════════════════

func GetPaymentHistory(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	status := c.Query("status") // filter: pending/settlement/expire/all
	limit := c.Query("limit", "100")
	lim, _ := strconv.Atoi(limit)
	if lim <= 0 || lim > 500 {
		lim = 100
	}

	q := `SELECT pt.id, pt.santri_id, COALESCE(s.nama,''), pt.bulan, pt.nominal, pt.fee, pt.total_bayar, 
		pt.order_id, pt.gateway, COALESCE(pt.payment_type,''), pt.transaction_status, pt.created_at 
		FROM payment_transactions pt 
		LEFT JOIN santri s ON pt.santri_id = s.id 
		WHERE pt.tenant_id = ?`
	args := []interface{}{tid}
	if status != "" && status != "all" {
		q += " AND pt.transaction_status = ?"
		args = append(args, status)
	}
	q += " ORDER BY pt.id DESC LIMIT ?"
	args = append(args, lim)

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, santriID int
		var santriNama, bulan, orderID, gw, payType, txStatus, createdAt string
		var nominal, fee, totalBayar int64
		rows.Scan(&id, &santriID, &santriNama, &bulan, &nominal, &fee, &totalBayar,
			&orderID, &gw, &payType, &txStatus, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriID, "santri_nama": santriNama,
			"bulan": bulan, "nominal": nominal, "fee": fee, "total_bayar": totalBayar,
			"order_id": orderID, "gateway": gw, "payment_type": payType,
			"transaction_status": txStatus, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ═══════════════════════════════════════════════════════════
// GET /api/wali/tunggakan/:santri_id — Get all unpaid months
// ═══════════════════════════════════════════════════════════

func WaliGetTunggakan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	sid := c.Params("santri_id")

	// Validate ownership
	if role == "wali" {
		var cnt int
		config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?", sid, tid, uid).Scan(&cnt)
		if cnt == 0 {
			return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
		}
	}

	// Get tagihan_bulan_mulai from settings
	var tagihanBulanMulai string
	config.DB.QueryRow("SELECT COALESCE(tagihan_bulan_mulai,'') FROM settings WHERE tenant_id = ?", tid).Scan(&tagihanBulanMulai)

	// Get payment gateway fee info
	pg, _ := getPGSettings(tid)
	var feePercent float64
	var feeFlat int64
	if pg != nil {
		feePercent = pg.FeePercent
		feeFlat = pg.FeeFlat
	}

	// Generate months from tagihan_bulan_mulai up to current month
	now := time.Now()
	bulanSekarang := now.Format("2006-01")
	var tunggakan []fiber.Map

	// Determine start month
	startMonth := bulanSekarang // default: only current month if no setting
	if tagihanBulanMulai != "" {
		startMonth = tagihanBulanMulai
	} else {
		// Fallback: 12 months back if no setting defined
		startMonth = now.AddDate(-1, 0, 0).Format("2006-01")
	}

	// Parse start month
	startYear := 0
	startMon := 0
	fmt.Sscanf(startMonth, "%d-%d", &startYear, &startMon)

	endYear := 0
	endMon := 0
	fmt.Sscanf(bulanSekarang, "%d-%d", &endYear, &endMon)

	sidInt, _ := strconv.Atoi(sid)

	// Iterate from startMonth to current month
	y, m := startYear, startMon
	for {
		bulan := fmt.Sprintf("%04d-%02d", y, m)
		if bulan > bulanSekarang {
			break
		}

		tagihan := HitungTagihan(tid, sidInt, bulan)
		if tagihan <= 0 {
			m++
			if m > 12 { m = 1; y++ }
			continue
		}

		var totalBayar int64
		config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?",
			tid, sid, bulan).Scan(&totalBayar)

		sisa := tagihan - totalBayar
		if sisa <= 0 {
			m++
			if m > 12 { m = 1; y++ }
			continue
		}

		// Check if there's a pending transaction
		var pendingOrderID, pendingToken string
		config.DB.QueryRow(`SELECT COALESCE(order_id,''), COALESCE(snap_token,'') FROM payment_transactions 
			WHERE tenant_id = ? AND santri_id = ? AND bulan = ? AND transaction_status = 'pending' 
			ORDER BY id DESC LIMIT 1`, tid, sid, bulan).Scan(&pendingOrderID, &pendingToken)

		fee := int64(math.Ceil(float64(sisa)*feePercent/100.0)) + feeFlat
		total := sisa + fee

		tunggakan = append(tunggakan, fiber.Map{
			"bulan":            bulan,
			"tagihan":          tagihan,
			"dibayar":          totalBayar,
			"sisa":             sisa,
			"fee":              fee,
			"total":            total,
			"pending_order_id": pendingOrderID,
			"pending_token":    pendingToken,
		})

		m++
		if m > 12 { m = 1; y++ }
	}

	if tunggakan == nil {
		tunggakan = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"tunggakan":        tunggakan,
		"fee_percent":      feePercent,
		"fee_flat":         feeFlat,
	})
}

// ═══════════════════════════════════════════════════════════
// GET /api/pembayaran/rekap-tunggakan — Admin/Bendahara grid
// ═══════════════════════════════════════════════════════════

func GetRekapTunggakan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tahun := c.Query("tahun")
	if tahun == "" {
		tahun = time.Now().Format("2006")
	}

	// Generate months Jan-Dec for the year
	months := make([]string, 12)
	for i := 0; i < 12; i++ {
		months[i] = fmt.Sprintf("%s-%02d", tahun, i+1)
	}

	// Get all active santri
	rows, err := config.DB.Query("SELECT s.id, s.nama FROM santri s WHERE s.tenant_id = ? AND s.status = 'aktif' ORDER BY s.nama", tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	type santriInfo struct {
		ID   int
		Nama string
	}
	var santriList []santriInfo
	var sids []interface{}
	for rows.Next() {
		var s santriInfo
		rows.Scan(&s.ID, &s.Nama)
		santriList = append(santriList, s)
		sids = append(sids, s.ID)
	}
	if len(santriList) == 0 {
		return c.JSON(fiber.Map{"data": []fiber.Map{}, "months": months})
	}

	// Build IN clause
	inPh := "?"
	for i := 1; i < len(sids); i++ {
		inPh += ",?"
	}

	// Get all payments for the year in one query
	bayarMap := map[int]map[string]int64{} // sid -> bulan -> total
	rb, _ := config.DB.Query(`SELECT santri_id, bulan, COALESCE(SUM(nominal),0) 
		FROM pembayaran WHERE tenant_id = ? AND bulan >= ? AND bulan <= ? AND santri_id IN (`+inPh+`)
		GROUP BY santri_id, bulan`,
		append([]interface{}{tid, months[0], months[11]}, sids...)...)
	if rb != nil {
		defer rb.Close()
		for rb.Next() {
			var sid int
			var bulan string
			var nom int64
			rb.Scan(&sid, &bulan, &nom)
			if bayarMap[sid] == nil {
				bayarMap[sid] = map[string]int64{}
			}
			bayarMap[sid][bulan] = nom
		}
	}

	// Get tagihan_bulan_mulai from settings
	var tagihanBulanMulai string
	config.DB.QueryRow("SELECT COALESCE(tagihan_bulan_mulai,'') FROM settings WHERE tenant_id = ?", tid).Scan(&tagihanBulanMulai)

	// Assemble result
	var data []fiber.Map
	for _, s := range santriList {
		status := fiber.Map{}
		for _, m := range months {
			// Skip bulan sebelum tagihan_bulan_mulai
			if tagihanBulanMulai != "" && m < tagihanBulanMulai {
				status[m] = "-"
				continue
			}

			tagihan := HitungTagihan(tid, s.ID, m)

			bayar := int64(0)
			if bayarMap[s.ID] != nil {
				bayar = bayarMap[s.ID][m]
			}
			if tagihan == 0 {
				status[m] = "gratis"
			} else if bayar >= tagihan {
				status[m] = "lunas"
			} else if bayar > 0 {
				status[m] = "kurang"
			} else {
				status[m] = "belum"
			}
		}

		data = append(data, fiber.Map{
			"id":     s.ID,
			"nama":   s.Nama,
			"status": status,
		})
	}

	if data == nil {
		data = []fiber.Map{}
	}

	return c.JSON(fiber.Map{"data": data, "months": months})
}

// ═══════════════════════════════════════════════════════════
// SANGU TOPUP WEBHOOK HANDLER
// ═══════════════════════════════════════════════════════════

func HandleSanguWebhook(c *fiber.Ctx, notif map[string]interface{}) error {
	orderID, _ := notif["order_id"].(string)
	
	var id, tenantID, santriID int
	var nominal int64
	var currentStatus, gateway string
	
	err := config.DB.QueryRow("SELECT id, tenant_id, santri_id, nominal, status, payment_method FROM sangu_topup WHERE payment_ref = ?", orderID).Scan(&id, &tenantID, &santriID, &nominal, &currentStatus, &gateway)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Transaction not found"})
	}
	
	transactionStatus, _ := notif["transaction_status"].(string)
	fraudStatus, _ := notif["fraud_status"].(string)
	
	// Get gateway server key for signature verification
	pg, err := getPGSettings(tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to read settings"})
	}
	
	// Verify signature for Midtrans
	if gateway == "midtrans" {
		sigKey, _ := notif["signature_key"].(string)
		if sigKey != "" {
			statusCode, _ := notif["status_code"].(string)
			grossAmount, _ := notif["gross_amount"].(string)
			raw := orderID + statusCode + grossAmount + pg.ServerKey
			h := sha512.Sum512([]byte(raw))
			expectedSig := fmt.Sprintf("%x", h)
			if sigKey != expectedSig {
				return c.Status(403).JSON(fiber.Map{"message": "Invalid signature"})
			}
		}
	}
	
	if currentStatus == "sukses" {
		return c.JSON(fiber.Map{"message": "Already success"})
	}
	
	if transactionStatus == "capture" || transactionStatus == "settlement" {
		if transactionStatus == "capture" && fraudStatus != "" && fraudStatus != "accept" {
			config.DB.Exec("UPDATE sangu_topup SET status = 'gagal', updated_at = NOW() WHERE id = ?", id)
			return c.JSON(fiber.Map{"message": "fraud detected"})
		}
		
		// Update status and add balance
		tx, _ := config.DB.Begin()
		_, err1 := tx.Exec("UPDATE sangu_topup SET status = 'sukses', updated_at = NOW() WHERE id = ?", id)
		_, err2 := tx.Exec("UPDATE santri SET sangu_saldo = sangu_saldo + ? WHERE id = ?", nominal, santriID)
		
		if err1 == nil && err2 == nil {
			tx.Commit()
		} else {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"message": "Gagal update saldo"})
		}
	} else if transactionStatus == "expire" || transactionStatus == "cancel" || transactionStatus == "deny" {
		config.DB.Exec("UPDATE sangu_topup SET status = 'gagal', updated_at = NOW() WHERE id = ?", id)
	}
	
	return c.JSON(fiber.Map{"message": "OK"})
}

// ═══════════════════════════════════════════════════════════
// POST /api/wali/pay-insidental — Create PG transaction for Insidental
// ═══════════════════════════════════════════════════════════
func WaliCreatePaymentInsidental(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	var body struct {
		TagihanSantriID int   `json:"tagihan_santri_id"`
		SantriID        int   `json:"santri_id"`
		NominalBebas    int64 `json:"nominal_bebas"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	if body.SantriID == 0 || body.TagihanSantriID == 0 || body.NominalBebas <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak lengkap atau nominal tidak valid"})
	}

	// Validate santri belongs to this wali
	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?",
		body.SantriID, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
	}

	// Cek sisa tagihan
	var sisa int64
	var namaTagihan string
	err := config.DB.QueryRow(`
		SELECT t.sisa_tagihan, m.nama 
		FROM tagihan_insidental_santri t
		JOIN tagihan_insidental m ON t.tagihan_id = m.id
		WHERE t.id = ? AND t.tenant_id = ?`, body.TagihanSantriID, tid).Scan(&sisa, &namaTagihan)

	if err != nil || sisa <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Tagihan tidak ditemukan atau sudah lunas"})
	}

	if body.NominalBebas > sisa {
		return c.Status(400).JSON(fiber.Map{"message": "Nominal melebihi sisa tagihan"})
	}

	// Get payment gateway settings
	pg, errPG := getPGSettings(tid)
	if errPG != nil || pg.Provider == "" || pg.ServerKey == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Payment gateway belum dikonfigurasi"})
	}

	fee := int64(math.Ceil(float64(body.NominalBebas)*pg.FeePercent/100.0)) + pg.FeeFlat
	totalPay := body.NominalBebas + fee

	bulanLabel := fmt.Sprintf("INS-%d", body.TagihanSantriID)
	orderID := fmt.Sprintf("PES%d-%d-%s-%d", tid, body.SantriID, bulanLabel, time.Now().UnixMilli())

	var waliNama, santriNama string
	config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", uid).Scan(&waliNama)
	config.DB.QueryRow("SELECT nama FROM santri WHERE id = ?", body.SantriID).Scan(&santriNama)

	// Create transaction based on gateway provider
	var snapToken, redirectURL string

	switch pg.Provider {
	case "midtrans":
		env := midtrans.Sandbox
		if pg.IsProduction {
			env = midtrans.Production
		}
		var s snap.Client
		s.New(pg.ServerKey, env)

		req := &snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: totalPay,
			},
			CustomerDetail: &midtrans.CustomerDetails{
				FName: waliNama,
			},
			Items: &[]midtrans.ItemDetails{
				{
					ID:    bulanLabel,
					Price: body.NominalBebas,
					Qty:   1,
					Name:  fmt.Sprintf("%s - %s", santriNama, namaTagihan),
				},
				{
					ID:    "FEE",
					Price: fee,
					Qty:   1,
					Name:  "Biaya Layanan Payment Gateway",
				},
			},
		}
		snapResp, errSnap := s.CreateTransaction(req)
		if errSnap != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal create Midtrans: " + errSnap.Error()})
		}
		snapToken = snapResp.Token
		redirectURL = snapResp.RedirectURL

	default:
		return c.Status(400).JSON(fiber.Map{"message": "Provider belum didukung"})
	}

	// Insert into payment_transactions
	_, errIns := config.DB.Exec(`INSERT INTO payment_transactions 
		(tenant_id, santri_id, bulan, nominal, fee, total_bayar, order_id, gateway, snap_token, transaction_status) 
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		tid, body.SantriID, bulanLabel, body.NominalBebas, fee, totalPay, orderID, pg.Provider, snapToken, "pending")
	if errIns != nil {
		log.Printf("[PAY-CREATE-INSIDENTAL] ERROR inserting: %v", errIns)
	}

	return c.JSON(fiber.Map{
		"snap_token":   snapToken,
		"redirect_url": redirectURL,
		"order_id":     orderID,
		"nominal":      body.NominalBebas,
		"fee":          fee,
		"total_bayar":  totalPay,
	})
}
