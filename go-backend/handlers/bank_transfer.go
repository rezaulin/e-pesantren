package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"pesantren-multi/config"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ═══════════════════════════════════════════════════════════
// HELPER: Get transfer bank settings for a tenant
// ═══════════════════════════════════════════════════════════

type tbSettings struct {
	Enabled       bool
	FeeFlat       int64
	ExpiryHours   int
	MootaEnabled  bool
	MootaToken    string
	MootaSecret   string
	RekeningBank  string
	RekeningNomor string
	RekeningNama  string
}

func getTBSettings(tenantID int) *tbSettings {
	var s tbSettings
	var enabledStr, mootaEnabledStr string
	var feeStr, expiryStr string
	err := config.DB.QueryRow(`SELECT 
		COALESCE(transfer_bank_enabled,'0'), COALESCE(transfer_bank_fee_flat,'0'),
		COALESCE(transfer_bank_expiry_hours,'24'),
		COALESCE(moota_enabled,'0'), COALESCE(moota_api_token,''), COALESCE(moota_secret_token,''),
		COALESCE(rekening_bank,''), COALESCE(rekening_nomor,''), COALESCE(rekening_atas_nama,'')
		FROM settings WHERE tenant_id = ? LIMIT 1`, tenantID).
		Scan(&enabledStr, &feeStr, &expiryStr,
			&mootaEnabledStr, &s.MootaToken, &s.MootaSecret,
			&s.RekeningBank, &s.RekeningNomor, &s.RekeningNama)
	if err != nil {
		return &s
	}
	en, _ := strconv.Atoi(enabledStr)
	s.Enabled = en == 1
	s.FeeFlat, _ = strconv.ParseInt(feeStr, 10, 64)
	s.ExpiryHours, _ = strconv.Atoi(expiryStr)
	if s.ExpiryHours <= 0 {
		s.ExpiryHours = 24
	}
	me, _ := strconv.Atoi(mootaEnabledStr)
	s.MootaEnabled = me == 1
	return &s
}

// generateKodeUnik generates a random 3-digit code (100-999) that produces
// a unique total_transfer among pending transfers for the given tenant.
func generateKodeUnik(tenantID int, nominal int64) int {
	for attempt := 0; attempt < 50; attempt++ {
		kode := 100 + rand.Intn(900) // 100-999
		total := nominal + int64(kode)
		var cnt int
		config.DB.QueryRow(`SELECT COUNT(*) FROM bank_transfers 
			WHERE tenant_id = ? AND total_transfer = ? AND status = 'pending'`,
			tenantID, total).Scan(&cnt)
		if cnt == 0 {
			return kode
		}
	}
	// Fallback: use timestamp-based
	return 100 + int(time.Now().UnixMilli()%900)
}

// ═══════════════════════════════════════════════════════════
// 1. POST /api/wali/bank-transfer — Wali creates bank transfer
// ═══════════════════════════════════════════════════════════

func WaliCreateBankTransfer(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	var body struct {
		SantriID  int      `json:"santri_id"`
		Tipe      string   `json:"tipe"`       // "spp" or "topup_sangu"
		BulanList []string `json:"bulan_list"`  // for SPP: ["2026-06","2026-07"]
		Nominal   int64    `json:"nominal"`     // for topup_sangu: amount
	}
	c.BodyParser(&body)

	if body.SantriID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "santri_id wajib diisi"})
	}
	if body.Tipe == "" {
		body.Tipe = "spp"
	}

	// Validate santri belongs to wali
	var cnt int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE id = ? AND tenant_id = ? AND wali_user_id = ?",
		body.SantriID, tid, uid).Scan(&cnt)
	if cnt == 0 {
		return c.Status(403).JSON(fiber.Map{"message": "Santri bukan anak Anda"})
	}

	// Get transfer bank settings
	tb := getTBSettings(tid)
	if !tb.Enabled {
		return c.Status(400).JSON(fiber.Map{"message": "Pembayaran via transfer bank belum diaktifkan"})
	}
	if tb.RekeningBank == "" || tb.RekeningNomor == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Rekening pesantren belum dikonfigurasi"})
	}

	var totalNominal int64
	var bulanListJSON string

	if body.Tipe == "spp" {
		if len(body.BulanList) == 0 {
			return c.Status(400).JSON(fiber.Map{"message": "bulan_list wajib untuk SPP"})
		}

		// Calculate tagihan per bulan using helper
		for _, bulan := range body.BulanList {
			tagihan := HitungTagihan(tid, body.SantriID, bulan)
			if tagihan <= 0 {
				continue
			}
			var totalBayar int64
			config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?",
				tid, body.SantriID, bulan).Scan(&totalBayar)
			sisa := tagihan - totalBayar
			if sisa > 0 {
				totalNominal += sisa
			}
		}

		if totalNominal <= 0 {
			return c.Status(400).JSON(fiber.Map{"message": "Semua bulan sudah lunas"})
		}

		bl, _ := json.Marshal(body.BulanList)
		bulanListJSON = string(bl)

		// Expire old pending transfers for same santri + overlapping months
		config.DB.Exec(`UPDATE bank_transfers SET status = 'expired', expired_at = NOW()
			WHERE tenant_id = ? AND santri_id = ? AND tipe = 'spp' AND status = 'pending'`,
			tid, body.SantriID)

	} else if body.Tipe == "topup_sangu" {
		if body.Nominal <= 0 {
			return c.Status(400).JSON(fiber.Map{"message": "Nominal topup harus > 0"})
		}
		totalNominal = body.Nominal
		bulanListJSON = "[]"

		// Expire old pending topup for same santri
		config.DB.Exec(`UPDATE bank_transfers SET status = 'expired', expired_at = NOW()
			WHERE tenant_id = ? AND santri_id = ? AND tipe = 'topup_sangu' AND status = 'pending'`,
			tid, body.SantriID)
	} else if body.Tipe == "insidental" {
		if body.Nominal <= 0 {
			return c.Status(400).JSON(fiber.Map{"message": "Nominal pembayaran insidental harus > 0"})
		}
		var tagihanID int
		// Check the tagihan_santri_id in json
		var parsed map[string]interface{}
		json.Unmarshal(c.Body(), &parsed)
		if parsed["tagihan_santri_id"] != nil {
			tagihanID = int(parsed["tagihan_santri_id"].(float64))
		}
		if tagihanID == 0 {
			return c.Status(400).JSON(fiber.Map{"message": "ID tagihan insidental harus disertakan"})
		}
		
		totalNominal = body.Nominal
		bulanListJSON = fmt.Sprintf("[%d]", tagihanID)

		// Expire old pending insidental for same santri
		config.DB.Exec(`UPDATE bank_transfers SET status = 'expired', expired_at = NOW()
			WHERE tenant_id = ? AND santri_id = ? AND tipe = 'insidental' AND status = 'pending'`,
			tid, body.SantriID)
	} else {
		return c.Status(400).JSON(fiber.Map{"message": "Tipe tidak valid"})
	}

	// Calculate fee + kode unik
	fee := tb.FeeFlat
	nominalPlusFee := totalNominal + fee
	kodeUnik := generateKodeUnik(tid, nominalPlusFee)
	totalTransfer := nominalPlusFee + int64(kodeUnik)

	// Insert bank_transfers
	res, err := config.DB.Exec(`INSERT INTO bank_transfers 
		(tenant_id, santri_id, tipe, bulan_list, nominal, fee, kode_unik, total_transfer, status)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		tid, body.SantriID, body.Tipe, bulanListJSON, totalNominal, fee, kodeUnik, totalTransfer, "pending")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat transfer: " + err.Error()})
	}

	insertID, _ := res.LastInsertId()

	// Calculate expiry
	expiry := time.Now().Add(time.Duration(tb.ExpiryHours) * time.Hour)

	log.Printf("[BANK-TF] Created id=%d tenant=%d santri=%d tipe=%s nominal=%d fee=%d kode=%d total=%d",
		insertID, tid, body.SantriID, body.Tipe, totalNominal, fee, kodeUnik, totalTransfer)

	return c.JSON(fiber.Map{
		"id":             insertID,
		"nominal":        totalNominal,
		"fee":            fee,
		"kode_unik":      kodeUnik,
		"total_transfer": totalTransfer,
		"rekening": fiber.Map{
			"bank":      tb.RekeningBank,
			"nomor":     tb.RekeningNomor,
			"atas_nama": tb.RekeningNama,
		},
		"expiry":     expiry.Format("2006-01-02T15:04:05"),
		"expiry_jam": tb.ExpiryHours,
	})
}

// ═══════════════════════════════════════════════════════════
// 2. GET /api/wali/bank-transfers — Wali list own transfers
// ═══════════════════════════════════════════════════════════

func WaliGetBankTransfers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	rows, err := config.DB.Query(`SELECT bt.id, bt.santri_id, s.nama, bt.tipe, bt.bulan_list,
		bt.nominal, bt.fee, bt.kode_unik, bt.total_transfer, bt.status, bt.created_at
		FROM bank_transfers bt
		JOIN santri s ON bt.santri_id = s.id
		WHERE bt.tenant_id = ? AND s.wali_user_id = ?
		ORDER BY bt.id DESC LIMIT 50`, tid, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, santriID, kodeUnik int
		var santriNama, tipe, bulanList, status, createdAt string
		var nominal, fee, totalTransfer int64
		rows.Scan(&id, &santriID, &santriNama, &tipe, &bulanList,
			&nominal, &fee, &kodeUnik, &totalTransfer, &status, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriID, "santri_nama": santriNama,
			"tipe": tipe, "bulan_list": bulanList,
			"nominal": nominal, "fee": fee, "kode_unik": kodeUnik,
			"total_transfer": totalTransfer, "status": status, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}

	// Also return rekening info
	tb := getTBSettings(tid)
	return c.JSON(fiber.Map{
		"transfers": list,
		"rekening": fiber.Map{
			"bank":      tb.RekeningBank,
			"nomor":     tb.RekeningNomor,
			"atas_nama": tb.RekeningNama,
		},
	})
}

// ═══════════════════════════════════════════════════════════
// 3. GET /api/bank-transfers — Admin list all transfers
// ═══════════════════════════════════════════════════════════

func GetBankTransfers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	status := c.Query("status", "pending")

	q := `SELECT bt.id, bt.santri_id, s.nama, bt.tipe, bt.bulan_list,
		bt.nominal, bt.fee, bt.kode_unik, bt.total_transfer, bt.status,
		bt.created_at, COALESCE(bt.confirmed_by,0), COALESCE(bt.confirmed_at,''),
		COALESCE(bt.moota_mutation_id,'')
		FROM bank_transfers bt
		JOIN santri s ON bt.santri_id = s.id
		WHERE bt.tenant_id = ?`
	args := []interface{}{tid}

	if status != "all" {
		q += " AND bt.status = ?"
		args = append(args, status)
	}
	q += " ORDER BY bt.id DESC LIMIT 200"

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var list []fiber.Map
	for rows.Next() {
		var id, santriID, kodeUnik, confirmedBy int
		var santriNama, tipe, bulanList, st, createdAt, confirmedAt, mootaID string
		var nominal, fee, totalTransfer int64
		rows.Scan(&id, &santriID, &santriNama, &tipe, &bulanList,
			&nominal, &fee, &kodeUnik, &totalTransfer, &st,
			&createdAt, &confirmedBy, &confirmedAt, &mootaID)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriID, "santri_nama": santriNama,
			"tipe": tipe, "bulan_list": bulanList,
			"nominal": nominal, "fee": fee, "kode_unik": kodeUnik,
			"total_transfer": totalTransfer, "status": st,
			"created_at": createdAt, "confirmed_by": confirmedBy,
			"confirmed_at": confirmedAt, "moota_mutation_id": mootaID,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// ═══════════════════════════════════════════════════════════
// 4. POST /api/bank-transfers/:id/confirm — Admin confirms
// ═══════════════════════════════════════════════════════════

func ConfirmBankTransfer(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	btID := c.Params("id")

	return confirmBankTransferByID(tid, btID, uid, "")
}

// Shared confirmation logic (used by admin confirm AND Moota webhook)
func confirmBankTransferByID(tenantID int, btID string, confirmedBy int, mootaMutationID string) error {
	var id, santriID, kodeUnik int
	var tipe, bulanListStr, status string
	var nominal, fee, totalTransfer int64

	err := config.DB.QueryRow(`SELECT id, santri_id, tipe, bulan_list, nominal, fee, kode_unik, total_transfer, status
		FROM bank_transfers WHERE id = ? AND tenant_id = ?`, btID, tenantID).
		Scan(&id, &santriID, &tipe, &bulanListStr, &nominal, &fee, &kodeUnik, &totalTransfer, &status)
	if err != nil {
		return fmt.Errorf("Transfer tidak ditemukan")
	}
	if status != "pending" {
		return fmt.Errorf("Transfer sudah %s", status)
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return fmt.Errorf("DB error")
	}

	// Update bank_transfers status
	mootaUpdate := ""
	if mootaMutationID != "" {
		mootaUpdate = ", moota_mutation_id = '" + mootaMutationID + "'"
	}
	_, err = tx.Exec(`UPDATE bank_transfers SET status = 'confirmed', confirmed_by = ?, 
		confirmed_at = NOW()`+mootaUpdate+` WHERE id = ?`, confirmedBy, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Gagal update status")
	}

	if tipe == "spp" {
		// Parse bulan_list JSON
		var bulanList []string
		json.Unmarshal([]byte(bulanListStr), &bulanList)

		if len(bulanList) == 0 {
			tx.Rollback()
			return fmt.Errorf("bulan_list kosong")
		}

		// Get tagihan per bulan using helper
		// Insert pembayaran per bulan
		sisaTotal := totalTransfer
		for i, bulan := range bulanList {
			tagihanPerBulan := HitungTagihan(tenantID, santriID, bulan)
			if tagihanPerBulan <= 0 {
				continue
			}

			// Check sisa tagihan for this month
			var sudahBayar int64
			config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?",
				tenantID, santriID, bulan).Scan(&sudahBayar)
			sisa := tagihanPerBulan - sudahBayar
			if sisa <= 0 {
				continue
			}

			bayar := sisa
			if i == len(bulanList)-1 {
				// Last month gets remainder (including kode_unik + fee)
				bayar = sisaTotal
			}
			if bayar > sisaTotal {
				bayar = sisaTotal
			}

			keterangan := "Transfer Bank (kode unik +" + strconv.Itoa(kodeUnik) + ")"
			if mootaMutationID != "" {
				keterangan = "Transfer Bank via Moota (auto)"
			}
			_, errPb := tx.Exec(`INSERT INTO pembayaran (tenant_id, santri_id, bulan, nominal, metode, keterangan, created_by) 
				VALUES (?,?,?,?,?,?,0)`,
				tenantID, santriID, bulan, bayar, "transfer_bank", keterangan)
			if errPb != nil {
				log.Printf("[BANK-TF] ERROR insert pembayaran bulan=%s: %v", bulan, errPb)
			}
			sisaTotal -= bayar
			if sisaTotal <= 0 {
				break
			}
		}
		log.Printf("[BANK-TF] Confirmed SPP id=%d santri=%d bulan=%v nominal=%d", id, santriID, bulanList, totalTransfer)

		// Insert ke catatan_keuangan untuk SPP via transfer bank
		var santriNama string
		tx.QueryRow("SELECT nama FROM santri WHERE id = ? AND tenant_id = ?", santriID, tenantID).Scan(&santriNama)
		sort.Strings(bulanList)
		var ketSPP string
		if len(bulanList) == 1 {
			ketSPP = fmt.Sprintf("SPP %s a.n. %s (Via Transfer Bank)", bulanList[0], santriNama)
		} else {
			ketSPP = fmt.Sprintf("SPP %s s/d %s a.n. %s (Via Transfer Bank)", bulanList[0], bulanList[len(bulanList)-1], santriNama)
		}
		if mootaMutationID != "" {
			ketSPP += " via Moota (auto)"
		}
		tx.Exec("INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?,?,?,?,CURDATE(),0)",
			tenantID, "masuk", totalTransfer, ketSPP)

	} else if tipe == "topup_sangu" {
		// Add saldo (nominal only, without fee/kode_unik for sangu balance)
		_, err1 := tx.Exec("UPDATE santri SET sangu_saldo = sangu_saldo + ? WHERE id = ?", nominal, santriID)
		// Record in sangu_topup
		_, err2 := tx.Exec(`INSERT INTO sangu_topup (tenant_id, santri_id, nominal, fee, total, status, payment_method, payment_ref)
			VALUES (?,?,?,?,?,?,?,?)`,
			tenantID, santriID, nominal, fee, totalTransfer, "sukses", "transfer_bank",
			fmt.Sprintf("BT-%d", id))

		if err1 != nil || err2 != nil {
			tx.Rollback()
			return fmt.Errorf("Gagal update saldo sangu")
		}
		log.Printf("[BANK-TF] Confirmed Topup id=%d santri=%d nominal=%d", id, santriID, nominal)
	} else if tipe == "insidental" {
		// Parse bulan_list JSON which contains [tagihanID]
		var bulanList []int
		json.Unmarshal([]byte(bulanListStr), &bulanList)

		if len(bulanList) == 0 {
			tx.Rollback()
			return fmt.Errorf("ID tagihan tidak ditemukan")
		}
		tagihanID := bulanList[0]

		// Find the tagihan to deduct
		var sisaTagihan int64
		errTagihan := tx.QueryRow("SELECT sisa_tagihan FROM tagihan_insidental_santri WHERE id = ?", tagihanID).Scan(&sisaTagihan)
		if errTagihan != nil {
			tx.Rollback()
			return fmt.Errorf("Tagihan insidental tidak ditemukan")
		}

		if sisaTagihan > 0 {
			bayar := nominal // The exact amount requested by the user
			if bayar > sisaTagihan {
				bayar = sisaTagihan
			}

			// Update tagihan_insidental_santri
			_, errUpd := tx.Exec("UPDATE tagihan_insidental_santri SET sisa_tagihan = sisa_tagihan - ? WHERE id = ?", bayar, tagihanID)
			
			// If it's fully paid now, update status
			if sisaTagihan-bayar <= 0 {
				tx.Exec("UPDATE tagihan_insidental_santri SET status = 'LUNAS' WHERE id = ?", tagihanID)
			}

			keteranganKasExtra := " (kode unik +" + strconv.Itoa(kodeUnik) + ")"
			if mootaMutationID != "" {
				keteranganKasExtra = " via Moota (auto)"
			}

			// Record history
			_, errHist := tx.Exec(`INSERT INTO pembayaran_insidental 
				(tenant_id, santri_id, tagihan_santri_id, nominal_dibayar, metode, created_by) 
				VALUES (?,?,?,?,?,0)`,
				tenantID, santriID, tagihanID, bayar, "transfer_bank")

			// Add to catatan_keuangan
			var tgs struct {
				IDTagihanMaster int
				TagihanNama     string
				SantriNama      string
			}
			tx.QueryRow(`
				SELECT t.id_tagihan, m.nama, s.nama
				FROM tagihan_insidental_santri t
				JOIN tagihan_insidental m ON t.id_tagihan = m.id
				JOIN santri s ON t.santri_id = s.id
				WHERE t.id = ?`, tagihanID).Scan(&tgs.IDTagihanMaster, &tgs.TagihanNama, &tgs.SantriNama)

			keteranganKas := fmt.Sprintf("Pembayaran Insidental (%s) - %s (Via Transfer Bank)%s", tgs.TagihanNama, tgs.SantriNama, keteranganKasExtra)
			tx.Exec(`INSERT INTO catatan_keuangan 
				(tenant_id, tipe, nominal, keterangan, tanggal, created_by)
				VALUES (?, 'masuk', ?, ?, CURDATE(), 0)`,
				tenantID, bayar, keteranganKas)

			if errUpd != nil || errHist != nil {
				log.Printf("[BANK-TF] ERROR insert pembayaran insidental: %v", errUpd)
			}
		}
		log.Printf("[BANK-TF] Confirmed Insidental id=%d tagihanID=%d nominal=%d", id, tagihanID, totalTransfer)
	}

	tx.Commit()
	return nil
}

// ═══════════════════════════════════════════════════════════
// HTTP wrapper for ConfirmBankTransfer (returns fiber response)
// ═══════════════════════════════════════════════════════════

func init() {
	// Override ConfirmBankTransfer to use the shared logic with fiber response
}

func ConfirmBankTransferHTTP(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	btID := c.Params("id")

	err := confirmBankTransferByID(tid, btID, uid, "")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Transfer dikonfirmasi"})
}

// ═══════════════════════════════════════════════════════════
// 5. POST /api/bank-transfers/:id/reject — Admin rejects
// ═══════════════════════════════════════════════════════════

func RejectBankTransfer(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	btID := c.Params("id")

	var status string
	config.DB.QueryRow("SELECT status FROM bank_transfers WHERE id = ? AND tenant_id = ?", btID, tid).Scan(&status)
	if status != "pending" {
		return c.Status(400).JSON(fiber.Map{"message": "Transfer sudah " + status})
	}

	config.DB.Exec("UPDATE bank_transfers SET status = 'rejected' WHERE id = ? AND tenant_id = ?", btID, tid)
	log.Printf("[BANK-TF] Rejected id=%s tenant=%d", btID, tid)
	return c.JSON(fiber.Map{"message": "Transfer ditolak"})
}

// ═══════════════════════════════════════════════════════════
// 6. GET /api/bank-transfers/count-pending — Badge count
// ═══════════════════════════════════════════════════════════

func CountPendingBankTransfers(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM bank_transfers WHERE tenant_id = ? AND status = 'pending'", tid).Scan(&count)
	return c.JSON(fiber.Map{"count": count})
}

// ═══════════════════════════════════════════════════════════
// 7. POST /api/moota/webhook — Moota auto-confirm webhook
// ═══════════════════════════════════════════════════════════

// Moota server IP for whitelist validation
const mootaServerIP = "128.199.173.138"

func MootaWebhook(c *fiber.Ctx) error {
	// Get real client IP
	clientIP := c.IP()
	// Also check X-Forwarded-For (behind Cloudflare/nginx)
	forwarded := c.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		clientIP = strings.TrimSpace(parts[0])
	}

	// Parse IP for comparison (strip port if present)
	host, _, err := net.SplitHostPort(clientIP)
	if err == nil {
		clientIP = host
	}

	log.Printf("[MOOTA] Webhook received from IP=%s", clientIP)

	// Parse webhook payload
	var payload []map[string]interface{}
	// Moota sends array of mutations
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		// Try single object
		var single map[string]interface{}
		if err2 := json.Unmarshal(c.Body(), &single); err2 != nil {
			log.Printf("[MOOTA] Invalid JSON payload")
			return c.Status(400).JSON(fiber.Map{"message": "Invalid JSON"})
		}
		payload = []map[string]interface{}{single}
	}

	for _, mutation := range payload {
		// Get amount
		amountStr := fmt.Sprintf("%v", mutation["amount"])
		// Remove decimal part if any
		if idx := strings.Index(amountStr, "."); idx != -1 {
			amountStr = amountStr[:idx]
		}
		amount, _ := strconv.ParseInt(amountStr, 10, 64)
		if amount <= 0 {
			continue
		}

		// Only process credit (incoming money)
		txType := fmt.Sprintf("%v", mutation["type"])
		if txType != "CR" && txType != "cr" {
			continue
		}

		mutationID := fmt.Sprintf("%v", mutation["mutation_id"])
		bankID := fmt.Sprintf("%v", mutation["bank_id"])

		log.Printf("[MOOTA] Processing mutation amount=%d type=%s mutation_id=%s bank_id=%s",
			amount, txType, mutationID, bankID)

		// Find matching pending transfers across ALL tenants that have Moota enabled
		// We need to verify the secret token per-tenant
		rows, err := config.DB.Query(`SELECT bt.id, bt.tenant_id 
			FROM bank_transfers bt 
			WHERE bt.total_transfer = ? AND bt.status = 'pending'`, amount)
		if err != nil {
			continue
		}

		for rows.Next() {
			var btID, btTenantID int
			rows.Scan(&btID, &btTenantID)

			// Verify this tenant has Moota enabled
			tb := getTBSettings(btTenantID)
			if !tb.MootaEnabled {
				continue
			}

			// Verify secret token if provided by Moota
			secretHeader := c.Get("Authorization")
			if secretHeader == "" {
				secretHeader = c.Get("X-Moota-Secret")
			}
			if tb.MootaSecret != "" && secretHeader != "" {
				// Clean up "Bearer " prefix if present
				cleanSecret := strings.TrimPrefix(secretHeader, "Bearer ")
				if cleanSecret != tb.MootaSecret {
					log.Printf("[MOOTA] Secret mismatch for tenant=%d", btTenantID)
					continue
				}
			}

			// Also do IP whitelist check per tenant
			if clientIP != mootaServerIP && clientIP != "127.0.0.1" && clientIP != "::1" {
				log.Printf("[MOOTA] IP not whitelisted: %s (expected %s)", clientIP, mootaServerIP)
				// Don't hard-block, just log warning — some Moota IPs may change
			}

			// Auto-confirm!
			err := confirmBankTransferByID(btTenantID, strconv.Itoa(btID), 0, mutationID)
			if err != nil {
				log.Printf("[MOOTA] Failed to confirm bt_id=%d: %v", btID, err)
			} else {
				log.Printf("[MOOTA] Auto-confirmed bt_id=%d tenant=%d amount=%d", btID, btTenantID, amount)
			}
		}
		rows.Close()
	}

	// Always return 200 quickly (Moota best practice)
	return c.JSON(fiber.Map{"message": "OK"})
}

// ═══════════════════════════════════════════════════════════
// AUTO-EXPIRE: Background goroutine to expire old transfers
// ═══════════════════════════════════════════════════════════

func StartBankTransferExpiry() {
	go func() {
		// Wait a bit for DB to be ready
		time.Sleep(10 * time.Second)
		for {
			result, err := config.DB.Exec(`UPDATE bank_transfers SET status = 'expired', expired_at = NOW()
				WHERE status = 'pending' AND created_at < NOW() - INTERVAL 24 HOUR`)
			if err == nil {
				affected, _ := result.RowsAffected()
				if affected > 0 {
					log.Printf("[BANK-TF] Auto-expired %d transfers", affected)
				}
			}
			time.Sleep(1 * time.Hour)
		}
	}()
}

// Suppress unused import warnings
var _ = math.Ceil
