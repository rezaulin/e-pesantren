package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

// importMu prevents duplicate import requests from concurrent clicks.
// Key = tenant_id, so different tenants can import in parallel.
var importMu sync.Map

// normalizePhone strips non-digit chars and normalizes Indonesian phone numbers.
// Handles Excel quirks: numbers stored as float (trailing .0), missing leading zero, etc.
func normalizePhone(hp string) string {
	hp = strings.TrimSpace(hp)
	// Remove trailing .0 or .00 (Excel stores numbers as float)
	if idx := strings.Index(hp, "."); idx > 0 {
		allZero := true
		for _, c := range hp[idx+1:] {
			if c != '0' {
				allZero = false
				break
			}
		}
		if allZero {
			hp = hp[:idx]
		}
	}
	// Strip all non-digit characters
	re := regexp.MustCompile(`[^0-9]`)
	hp = re.ReplaceAllString(hp, "")
	if hp == "" {
		return ""
	}
	// Normalize +62 / 62 prefix to 0
	if strings.HasPrefix(hp, "62") && len(hp) > 4 {
		hp = "0" + hp[2:]
	}
	// Add leading 0 if starts with 8 (common Excel issue: 0812... becomes 812...)
	if strings.HasPrefix(hp, "8") && len(hp) >= 9 && len(hp) <= 13 {
		hp = "0" + hp
	}
	return hp
}

// autoCreateWaliUser creates a user with role=wali for the given phone number.
// Returns the user ID (existing or new).
func autoCreateWaliUser(tid int, noHP, namaWali, namaSantri string) int {
	if noHP == "" {
		return 0
	}
	username := normalizePhone(noHP)
	if username == "" {
		return 0
	}

	// Check if user already exists
	var existID int
	err := config.DB.QueryRow("SELECT id FROM users WHERE username = ? AND tenant_id = ?", username, tid).Scan(&existID)
	if err == nil && existID > 0 {
		return existID
	}

	// Create new wali user
	displayName := namaWali
	if displayName == "" {
		displayName = "Wali " + namaSantri
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(username), 10)
	res, err := config.DB.Exec("INSERT INTO users (tenant_id, username, password_hash, role, nama) VALUES (?,?,?,?,?)",
		tid, username, string(hash), "wali", displayName)
	if err != nil {
		fmt.Println("⚠️ Auto-create wali error:", err)
		return 0
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func GetSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	// Single JOIN query — no N+1 for wali names
	q := `SELECT s.id, s.nama, COALESCE(s.alamat,''), 
		COALESCE((SELECT kd.nama FROM santri_kelas_diniyyah skd JOIN kelas_diniyyah kd ON skd.kelas_diniyyah_id = kd.id WHERE skd.santri_id = s.id AND skd.tenant_id = s.tenant_id AND skd.status = 'active' LIMIT 1), s.kelas_diniyyah, ''), 
		s.status,
		s.kamar_id, COALESCE(k.nama,'-'), COALESCE(s.wali_user_id,0),
		COALESCE(s.nama_wali,''), COALESCE(s.no_hp,''),
		COALESCE(u.nama,'-'), COALESCE(s.kategori_spp_id,0), COALESCE(pk.nama,'-'),
		COALESCE(s.card_uid,''), COALESCE(s.kelas_sekolah,'')
		FROM santri s
		LEFT JOIN kamar k ON s.kamar_id = k.id
		LEFT JOIN pembayaran_kategori pk ON s.kategori_spp_id = pk.id
		LEFT JOIN users u ON s.wali_user_id = u.id AND s.wali_user_id > 0
		WHERE s.tenant_id = ?`
	args := []interface{}{tid}
	if kid := c.Query("kamar_id"); kid != "" {
		q += " AND s.kamar_id = ?"
		args = append(args, kid)
	}
	if ksid := c.Query("kelas_sekolah_id"); ksid != "" {
		q += " AND s.id IN (SELECT santri_id FROM santri_kelas WHERE kelas_id = ? AND status='active')"
		args = append(args, ksid)
	}
	if kdid := c.Query("kelas_diniyyah_id"); kdid != "" {
		q += " AND s.id IN (SELECT santri_id FROM santri_kelas_diniyyah WHERE kelas_diniyyah_id = ? AND status='active')"
		args = append(args, kdid)
	}
	if st := c.Query("status"); st != "" {
		q += " AND s.status = ?"
		args = append(args, st)
	} else {
		q += " AND s.status NOT IN ('pendaftar','ditolak')"
	}
	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, kamarID, waliUID, katSppId int
		var nama, alamat, kelas, status, kamarNama, namaWaliText, noHP, waliUserNama, katSppNama, cardUID, kelasSekolah string
		rows.Scan(&id, &nama, &alamat, &kelas, &status, &kamarID, &kamarNama, &waliUID, &namaWaliText, &noHP, &waliUserNama, &katSppId, &katSppNama, &cardUID, &kelasSekolah)
		waliNama := "-"
		if waliUID > 0 && waliUserNama != "-" {
			waliNama = waliUserNama
		} else if namaWaliText != "" {
			waliNama = namaWaliText
		}
		list = append(list, fiber.Map{"id": id, "nama": nama, "alamat": alamat, "kelas_diniyyah": kelas, "kelas_sekolah": kelasSekolah, "status": status, "kamar_id": kamarID, "kamar_nama": kamarNama, "wali_nama": waliNama, "wali_user_id": waliUID, "no_hp": noHP, "kategori_spp_id": katSppId, "kategori_spp_nama": katSppNama, "card_uid": cardUID})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama        string `json:"nama"`
		KamarID     int    `json:"kamar_id"`
		Status      string `json:"status"`
		KelasDin    string `json:"kelas_diniyyah"`
		Alamat      string `json:"alamat"`
		NoHP        string `json:"no_hp"`
		NamaWali    string `json:"nama_wali"`
		KategoriSpp int    `json:"kategori_spp_id"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.KamarID == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nama & kamar wajib"})
	}
	if body.Status == "" {
		body.Status = "aktif"
	}

	// ── Cek Limit Kuota Santri (SaaS) ──
	var limit int
	config.DB.QueryRow("SELECT kuota_santri FROM tenants WHERE id = ?", tid).Scan(&limit)
	var currentCount int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE tenant_id = ?", tid).Scan(&currentCount)
	if currentCount >= limit {
		return c.Status(403).JSON(fiber.Map{"message": fmt.Sprintf("Kuota santri Anda penuh (%d/%d). Silakan upgrade paket untuk menambah santri tanpa batas.", currentCount, limit), "code": "LIMIT_EXCEEDED"})
	}

	// Auto-create wali user if no_hp provided
	var waliUserID int
	normalizedHP := normalizePhone(body.NoHP)
	if normalizedHP != "" {
		waliUserID = autoCreateWaliUser(tid, normalizedHP, body.NamaWali, body.Nama)
	}

	var id int64
	if body.KategoriSpp > 0 {
		res, errQuery := config.DB.Exec("INSERT INTO santri (tenant_id, nama, kamar_id, status, kelas_diniyyah, alamat, no_hp, nama_wali, wali_user_id, kategori_spp_id) VALUES (?,?,?,?,?,?,?,?,?,?)",
			tid, body.Nama, body.KamarID, body.Status, body.KelasDin, body.Alamat, normalizedHP, body.NamaWali, waliUserID, body.KategoriSpp)
		if errQuery != nil {
			return c.Status(500).JSON(fiber.Map{"message": errQuery.Error()})
		}
		id, _ = res.LastInsertId()
	} else {
		res, errQuery := config.DB.Exec("INSERT INTO santri (tenant_id, nama, kamar_id, status, kelas_diniyyah, alamat, no_hp, nama_wali, wali_user_id) VALUES (?,?,?,?,?,?,?,?,?)",
			tid, body.Nama, body.KamarID, body.Status, body.KelasDin, body.Alamat, normalizedHP, body.NamaWali, waliUserID)
		if errQuery != nil {
			return c.Status(500).JSON(fiber.Map{"message": errQuery.Error()})
		}
		id, _ = res.LastInsertId()
	}
	return c.JSON(fiber.Map{"id": id, "nama": body.Nama, "wali_user_id": waliUserID})
}

func UpdateSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	var body map[string]interface{}
	c.BodyParser(&body)
	allowed := []string{"nama", "status", "kelas_diniyyah", "kelompok_ngaji", "jenis_bakat", "kelas_sekolah", "alamat", "kamar_id", "nama_wali", "no_hp", "kategori_spp_id", "card_uid"}
	for _, f := range allowed {
		if v, ok := body[f]; ok {
			if f == "no_hp" {
				// Normalize and auto-create/update wali
				hp := normalizePhone(fmt.Sprintf("%v", v))
				config.DB.Exec("UPDATE santri SET no_hp = ? WHERE id = ? AND tenant_id = ?", hp, id, tid)
				if hp != "" {
					var sNama, namaWali string
					config.DB.QueryRow("SELECT nama, COALESCE(nama_wali,'') FROM santri WHERE id = ? AND tenant_id = ?", id, tid).Scan(&sNama, &namaWali)
					waliUID := autoCreateWaliUser(tid, hp, namaWali, sNama)
					if waliUID > 0 {
						config.DB.Exec("UPDATE santri SET wali_user_id = ? WHERE id = ? AND tenant_id = ?", waliUID, id, tid)
					}
				}
			} else {
				config.DB.Exec(fmt.Sprintf("UPDATE santri SET %s = ? WHERE id = ? AND tenant_id = ?", f), v, id, tid)
			}
		}
	}
	return c.JSON(fiber.Map{"message": "Santri diupdate"})
}

func DeleteSantri(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("id")
	// Cascade delete all related records
	tables := []string{
		"absen_harian", "absen_malam", "absen_sekolah", "absen_diniyyah",
		"catatan_guru", "pelanggaran",
		"nilai_kegiatan", "nilai_sekolah", "nilai_diniyyah",
		"santri_kelas", "santri_kelas_diniyyah",
		"kamar_members", "pembayaran", "pembayaran_potongan",
	}
	for _, t := range tables {
		config.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE santri_id = ? AND tenant_id = ?", t), sid, tid)
	}
	config.DB.Exec("DELETE FROM santri WHERE id = ? AND tenant_id = ?", sid, tid)
	return c.JSON(fiber.Map{"message": "Santri dan semua data terkait dihapus"})
}

func ImportSantriExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)

	// ── LOCK: prevent duplicate import from double-click ──
	lockKey := fmt.Sprintf("import_%d", tid)
	if _, loaded := importMu.LoadOrStore(lockKey, true); loaded {
		return c.Status(429).JSON(fiber.Map{
			"message": "⏳ Import sedang berjalan, mohon tunggu sampai selesai",
		})
	}
	defer importMu.Delete(lockKey)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "File Excel wajib diupload"})
	}
	f, _ := file.Open()
	defer f.Close()
	xlsx, err := excelize.OpenReader(f)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format Excel tidak valid"})
	}
	sheetName := xlsx.GetSheetName(0)
	excelRows, _ := xlsx.GetRows(sheetName)
	if len(excelRows) < 2 {
		return c.Status(400).JSON(fiber.Map{"message": "File Excel kosong atau hanya header"})
	}

	// Auto-detect column indices from header row
	colNama, colAlamat, colWali, colHP := -1, -1, -1, -1
	for ci, h := range excelRows[0] {
		hl := strings.ToLower(strings.TrimSpace(h))
		if colNama < 0 && (hl == "nama" || strings.Contains(hl, "nama santri") || strings.Contains(hl, "nama lengkap")) {
			colNama = ci
		} else if colAlamat < 0 && strings.Contains(hl, "alamat") {
			colAlamat = ci
		} else if colWali < 0 && strings.Contains(hl, "wali") {
			colWali = ci
		} else if colHP < 0 && (strings.Contains(hl, "telep") || strings.Contains(hl, "hp") || strings.Contains(hl, "phone") || strings.Contains(hl, "telepon") || strings.Contains(hl, "nomor")) {
			colHP = ci
		}
	}
	if colNama < 0 {
		colNama = 0
	}
	fmt.Printf("📋 Column mapping: nama=%d, alamat=%d, wali=%d, hp=%d\n", colNama, colAlamat, colWali, colHP)

	// ── PRE-LOAD existing santri names to prevent duplicates ──
	existingNames := map[string]bool{}
	existRows, _ := config.DB.Query("SELECT LOWER(TRIM(nama)) FROM santri WHERE tenant_id = ?", tid)
	var currentCount int
	if existRows != nil {
		defer existRows.Close()
		for existRows.Next() {
			var n string
			existRows.Scan(&n)
			existingNames[n] = true
			currentCount++
		}
	}

	// ── Cek Limit Kuota Santri (SaaS) ──
	var limit int
	config.DB.QueryRow("SELECT kuota_santri FROM tenants WHERE id = ?", tid).Scan(&limit)
	if currentCount >= limit {
		return c.Status(403).JSON(fiber.Map{"message": fmt.Sprintf("Kuota santri Anda sudah penuh (%d/%d). Silakan upgrade paket untuk menambah santri.", currentCount, limit), "code": "LIMIT_EXCEEDED"})
	}
	remainingQuota := limit - currentCount

	// ── PRE-LOAD existing wali users (username→id) to avoid repeated queries ──
	waliCache := map[string]int{}
	waliRows, _ := config.DB.Query("SELECT username, id FROM users WHERE tenant_id = ? AND role = 'wali'", tid)
	if waliRows != nil {
		defer waliRows.Close()
		for waliRows.Next() {
			var uname string
			var uid int
			waliRows.Scan(&uname, &uid)
			waliCache[uname] = uid
		}
	}

	// ── PARSE all rows first, skip duplicates ──
	type importRow struct {
		Nama     string
		Alamat   string
		NamaWali string
		NoHP     string
	}
	var pendingRows []importRow

	for i, row := range excelRows {
		if i == 0 {
			continue
		}
		if len(row) == 0 {
			continue
		}

		getCol := func(idx int) string {
			if idx < 0 || idx >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[idx])
		}

		nama := getCol(colNama)
		if nama == "" {
			continue
		}

		namaKey := strings.ToLower(strings.TrimSpace(nama))
		if existingNames[namaKey] {
			continue
		}
		existingNames[namaKey] = true

		alamat := getCol(colAlamat)
		namaWali := getCol(colWali)

		noHPRaw := getCol(colHP)
		if noHPRaw == "" && colHP >= 0 {
			colLetter := string(rune('A' + colHP))
			if colHP >= 26 {
				colLetter = "A" + string(rune('A'+colHP-26))
			}
			cellRef := fmt.Sprintf("%s%d", colLetter, i+1)
			val, _ := xlsx.GetCellValue(sheetName, cellRef)
			noHPRaw = strings.TrimSpace(val)
		}
		noHP := normalizePhone(noHPRaw)

		if len(pendingRows) >= remainingQuota {
			break // Stop adding if quota is reached
		}

		pendingRows = append(pendingRows, importRow{
			Nama: nama, Alamat: alamat, NamaWali: namaWali, NoHP: noHP,
		})
	}

	if len(pendingRows) == 0 {
		return c.JSON(fiber.Map{
			"message":  "Tidak ada data baru untuk diimport (semua sudah ada atau file kosong)",
			"imported": 0, "skipped": len(excelRows) - 1, "wali_created": 0,
		})
	}

	// ── Collect unique phone numbers that need wali accounts ──
	newWaliPhones := map[string]string{}
	for _, r := range pendingRows {
		if r.NoHP != "" {
			if _, cached := waliCache[r.NoHP]; !cached {
				if _, pending := newWaliPhones[r.NoHP]; !pending {
					dn := r.NamaWali
					if dn == "" {
						dn = "Wali " + r.Nama
					}
					newWaliPhones[r.NoHP] = dn
				}
			}
		}
	}

	// ── PRE-HASH all wali passwords in memory (bcrypt cost=4 for bulk import) ──
	// Cost 4 is sufficient for phone-number passwords; single-create uses cost 10
	type waliEntry struct {
		Phone       string
		DisplayName string
		Hash        string
	}
	var newWaliList []waliEntry
	for phone, displayName := range newWaliPhones {
		hash, _ := bcrypt.GenerateFromPassword([]byte(phone), 4)
		newWaliList = append(newWaliList, waliEntry{Phone: phone, DisplayName: displayName, Hash: string(hash)})
	}

	// ── BEGIN TRANSACTION: all-or-nothing insert ──
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memulai transaksi: " + err.Error()})
	}
	defer tx.Rollback() // no-op if already committed

	// ── INSERT wali users with INSERT IGNORE to prevent duplicates ──
	waliCount := 0
	for _, w := range newWaliList {
		res, err := tx.Exec(
			"INSERT IGNORE INTO users (tenant_id, username, password_hash, role, nama) VALUES (?,?,?,?,?)",
			tid, w.Phone, w.Hash, "wali", w.DisplayName)
		if err != nil {
			fmt.Println("⚠️ Auto-create wali error:", err)
			continue
		}
		uid, _ := res.LastInsertId()
		if uid > 0 {
			waliCache[w.Phone] = int(uid)
			waliCount++
		} else {
			// INSERT IGNORE hit a duplicate — fetch existing ID
			var existID int
			tx.QueryRow("SELECT id FROM users WHERE username = ? AND tenant_id = ?", w.Phone, tid).Scan(&existID)
			if existID > 0 {
				waliCache[w.Phone] = existID
			}
		}
	}

	// ── BATCH INSERT santri ──
	const batchSize = 100
	count := 0
	for start := 0; start < len(pendingRows); start += batchSize {
		end := start + batchSize
		if end > len(pendingRows) {
			end = len(pendingRows)
		}
		batch := pendingRows[start:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO santri (tenant_id, nama, alamat, nama_wali, no_hp, wali_user_id, kamar_id, status) VALUES ")
		args := make([]interface{}, 0, len(batch)*8)
		for j, r := range batch {
			if j > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("(?,?,?,?,?,?,1,'aktif')")
			waliUID := 0
			if r.NoHP != "" {
				if uid, ok := waliCache[r.NoHP]; ok {
					waliUID = uid
				}
			}
			args = append(args, tid, r.Nama, r.Alamat, r.NamaWali, r.NoHP, waliUID)
		}
		_, err := tx.Exec(sb.String(), args...)
		if err != nil {
			fmt.Println("⚠️ Batch insert error:", err)
			for _, r := range batch {
				waliUID := 0
				if r.NoHP != "" {
					if uid, ok := waliCache[r.NoHP]; ok {
						waliUID = uid
					}
				}
				_, e := tx.Exec("INSERT INTO santri (tenant_id, nama, alamat, nama_wali, no_hp, wali_user_id, kamar_id, status) VALUES (?,?,?,?,?,?,1,'aktif')",
					tid, r.Nama, r.Alamat, r.NamaWali, r.NoHP, waliUID)
				if e == nil {
					count++
				}
			}
		} else {
			count += len(batch)
		}
	}

	// ── COMMIT ──
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan data: " + err.Error()})
	}

	skipped := len(excelRows) - 1 - count
	return c.JSON(fiber.Map{
		"message":  fmt.Sprintf("%d santri diimport, %d dilewati (sudah ada), %d akun wali dibuat", count, skipped, waliCount),
		"imported": count, "skipped": skipped, "wali_created": waliCount,
		"columns": fiber.Map{"nama": colNama, "alamat": colAlamat, "wali": colWali, "hp": colHP},
	})
}

// BulkSetKategoriSpp — update kategori_spp_id untuk banyak santri sekaligus
func BulkSetKategoriSpp(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var payload struct {
		SantriIDs     []int `json:"santri_ids"`
		KategoriSppID *int  `json:"kategori_spp_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid JSON", "error": err.Error()})
	}

	if len(payload.SantriIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Tidak ada santri yang dipilih"})
	}

	if payload.KategoriSppID != nil && *payload.KategoriSppID == 0 {
		payload.KategoriSppID = nil
	}

	db := config.DB
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Database error", "error": err.Error()})
	}

	query := "UPDATE santri SET kategori_spp_id = ? WHERE id = ? AND tenant_id = ?"
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"message": "Database prepare error", "error": err.Error()})
	}
	defer stmt.Close()

	for _, id := range payload.SantriIDs {
		if payload.KategoriSppID == nil {
			_, err = stmt.Exec(nil, id, tid)
		} else {
			_, err = stmt.Exec(*payload.KategoriSppID, id, tid)
		}
		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"message": "Gagal mengupdate sebagian data", "error": err.Error()})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Database commit error", "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Kategori SPP berhasil diterapkan ke santri yang dipilih",
	})
}
