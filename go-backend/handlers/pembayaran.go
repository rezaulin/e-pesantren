package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// ── TARIF ──────────────────────────────────────────────

func GetTarif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama, nominal, aktif FROM pembayaran_tarif WHERE tenant_id = ? ORDER BY id", tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama string
		var nominal int64
		var aktif int
		rows.Scan(&id, &nama, &nominal, &aktif)
		list = append(list, fiber.Map{"id": id, "nama": nama, "nominal": nominal, "aktif": aktif})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateTarif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama    string `json:"nama"`
		Nominal int64  `json:"nominal"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.Nominal <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nama dan nominal wajib diisi"})
	}
	res, err := config.DB.Exec("INSERT INTO pembayaran_tarif (tenant_id, nama, nominal) VALUES (?,?,?)", tid, body.Nama, body.Nominal)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Tarif ditambahkan"})
}

func UpdateTarif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	var body struct {
		Nama    string `json:"nama"`
		Nominal int64  `json:"nominal"`
		Aktif   *int   `json:"aktif"`
	}
	c.BodyParser(&body)
	if body.Nama != "" {
		config.DB.Exec("UPDATE pembayaran_tarif SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, id, tid)
	}
	if body.Nominal > 0 {
		config.DB.Exec("UPDATE pembayaran_tarif SET nominal = ? WHERE id = ? AND tenant_id = ?", body.Nominal, id, tid)
	}
	if body.Aktif != nil {
		config.DB.Exec("UPDATE pembayaran_tarif SET aktif = ? WHERE id = ? AND tenant_id = ?", *body.Aktif, id, tid)
	}
	return c.JSON(fiber.Map{"message": "Tarif diupdate"})
}

func DeleteTarif(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pembayaran_tarif WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	// Also delete related periodes
	config.DB.Exec("DELETE FROM pembayaran_tarif_periode WHERE tarif_id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Tarif dihapus"})
}

// ── TARIF PERIODE ──────────────────────────────────────

func GetTarifPeriode(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tarifID := c.Params("id")
	rows, err := config.DB.Query("SELECT id, bulan_mulai, bulan_akhir, nominal FROM pembayaran_tarif_periode WHERE tenant_id = ? AND tarif_id = ? ORDER BY bulan_mulai", tid, tarifID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var mulai, akhir string
		var nominal int64
		rows.Scan(&id, &mulai, &akhir, &nominal)
		list = append(list, fiber.Map{"id": id, "bulan_mulai": mulai, "bulan_akhir": akhir, "nominal": nominal})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateTarifPeriode(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tarifID := c.Params("id")
	var body struct {
		BulanMulai string `json:"bulan_mulai"`
		BulanAkhir string `json:"bulan_akhir"`
		Nominal    int64  `json:"nominal"`
	}
	c.BodyParser(&body)
	if body.BulanMulai == "" || body.BulanAkhir == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Bulan mulai dan akhir wajib diisi"})
	}
	if body.BulanMulai > body.BulanAkhir {
		return c.Status(400).JSON(fiber.Map{"message": "Bulan mulai harus sebelum bulan akhir"})
	}
	if body.Nominal < 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nominal tidak boleh negatif"})
	}
	// Check overlap
	var overlap int
	config.DB.QueryRow("SELECT COUNT(*) FROM pembayaran_tarif_periode WHERE tenant_id = ? AND tarif_id = ? AND bulan_mulai <= ? AND bulan_akhir >= ?",
		tid, tarifID, body.BulanAkhir, body.BulanMulai).Scan(&overlap)
	if overlap > 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Periode tumpang tindih dengan periode yang sudah ada"})
	}
	res, err := config.DB.Exec("INSERT INTO pembayaran_tarif_periode (tenant_id, tarif_id, bulan_mulai, bulan_akhir, nominal) VALUES (?,?,?,?,?)",
		tid, tarifID, body.BulanMulai, body.BulanAkhir, body.Nominal)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Periode ditambahkan"})
}

func DeleteTarifPeriode(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pembayaran_tarif_periode WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Periode dihapus"})
}

// ── KATEGORI PEMBAYARAN SPP (PER SANTRI) ───────────────

func GetKategoriPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	rows, err := config.DB.Query("SELECT id, nama, nominal, aktif FROM pembayaran_kategori WHERE tenant_id = ? ORDER BY id", tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var nama string
		var nominal int64
		var aktif int
		rows.Scan(&id, &nama, &nominal, &aktif)
		list = append(list, fiber.Map{"id": id, "nama": nama, "nominal": nominal, "aktif": aktif})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateKategoriPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		Nama    string `json:"nama"`
		Nominal int64  `json:"nominal"`
	}
	c.BodyParser(&body)
	if body.Nama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Nama wajib diisi"})
	}
	res, err := config.DB.Exec("INSERT INTO pembayaran_kategori (tenant_id, nama, nominal) VALUES (?,?,?)", tid, body.Nama, body.Nominal)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Kategori ditambahkan"})
}

func UpdateKategoriPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	id := c.Params("id")
	var body struct {
		Nama    string `json:"nama"`
		Nominal int64  `json:"nominal"`
		Aktif   *int   `json:"aktif"`
	}
	c.BodyParser(&body)
	if body.Nama != "" {
		config.DB.Exec("UPDATE pembayaran_kategori SET nama = ? WHERE id = ? AND tenant_id = ?", body.Nama, id, tid)
	}
	if body.Nominal > 0 || (body.Nominal == 0 && body.Nama == "") { // If 0 provided explicitly via JSON it will still parse, but checking if nominal is present is trickier with int64. Let's assume updating to 0 is done via explicit call.
		config.DB.Exec("UPDATE pembayaran_kategori SET nominal = ? WHERE id = ? AND tenant_id = ?", body.Nominal, id, tid)
	}
	if body.Aktif != nil {
		config.DB.Exec("UPDATE pembayaran_kategori SET aktif = ? WHERE id = ? AND tenant_id = ?", *body.Aktif, id, tid)
	}
	return c.JSON(fiber.Map{"message": "Kategori diupdate"})
}

func DeleteKategoriPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pembayaran_kategori WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	
	// Set santri yg pakai kategori ini ke NULL
	config.DB.Exec("UPDATE santri SET kategori_spp_id = NULL WHERE kategori_spp_id = ? AND tenant_id = ?", c.Params("id"), tid)
	// Delete related periodes
	config.DB.Exec("DELETE FROM pembayaran_kategori_periode WHERE kategori_id = ? AND tenant_id = ?", c.Params("id"), tid)
	
	return c.JSON(fiber.Map{"message": "Kategori dihapus"})
}

// ── KATEGORI PERIODE ──────────────────────────────────

func GetKategoriPeriode(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kategoriID := c.Params("id")
	rows, err := config.DB.Query("SELECT id, bulan_mulai, bulan_akhir, nominal FROM pembayaran_kategori_periode WHERE tenant_id = ? AND kategori_id = ? ORDER BY bulan_mulai", tid, kategoriID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id int
		var mulai, akhir string
		var nominal int64
		rows.Scan(&id, &mulai, &akhir, &nominal)
		list = append(list, fiber.Map{"id": id, "bulan_mulai": mulai, "bulan_akhir": akhir, "nominal": nominal})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreateKategoriPeriode(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kategoriID := c.Params("id")
	var body struct {
		BulanMulai string `json:"bulan_mulai"`
		BulanAkhir string `json:"bulan_akhir"`
		Nominal    int64  `json:"nominal"`
	}
	c.BodyParser(&body)
	if body.BulanMulai == "" || body.BulanAkhir == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Bulan mulai dan akhir wajib diisi"})
	}
	if body.BulanMulai > body.BulanAkhir {
		return c.Status(400).JSON(fiber.Map{"message": "Bulan mulai harus sebelum bulan akhir"})
	}
	if body.Nominal < 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nominal tidak boleh negatif"})
	}
	// Check overlap
	var overlap int
	config.DB.QueryRow("SELECT COUNT(*) FROM pembayaran_kategori_periode WHERE tenant_id = ? AND kategori_id = ? AND bulan_mulai <= ? AND bulan_akhir >= ?",
		tid, kategoriID, body.BulanAkhir, body.BulanMulai).Scan(&overlap)
	if overlap > 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Periode tumpang tindih dengan periode yang sudah ada"})
	}
	res, err := config.DB.Exec("INSERT INTO pembayaran_kategori_periode (tenant_id, kategori_id, bulan_mulai, bulan_akhir, nominal) VALUES (?,?,?,?,?)",
		tid, kategoriID, body.BulanMulai, body.BulanAkhir, body.Nominal)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()
	return c.JSON(fiber.Map{"id": id, "message": "Periode ditambahkan"})
}

func DeleteKategoriPeriode(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pembayaran_kategori_periode WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Periode dihapus"})
}

// ── POTONGAN ───────────────────────────────────────────

func GetPotongan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT pp.id, pp.santri_id, s.nama, pp.nama, pp.nominal, pp.aktif 
		  FROM pembayaran_potongan pp 
		  LEFT JOIN santri s ON pp.santri_id = s.id 
		  WHERE pp.tenant_id = ? AND pp.aktif = 1 ORDER BY pp.id DESC`
	rows, err := config.DB.Query(q, tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, santriID int
		var santriNama, nama string
		var nominal int64
		var aktif int
		rows.Scan(&id, &santriID, &santriNama, &nama, &nominal, &aktif)
		list = append(list, fiber.Map{"id": id, "santri_id": santriID, "santri_nama": santriNama, "nama": nama, "nominal": nominal, "aktif": aktif})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func CreatePotongan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var body struct {
		SantriIDs []int  `json:"santri_ids"`
		Nama      string `json:"nama"`
		Nominal   int64  `json:"nominal"`
	}
	c.BodyParser(&body)
	if body.Nama == "" || body.Nominal <= 0 || len(body.SantriIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Nama, nominal, dan santri wajib diisi"})
	}
	count := 0
	for _, sid := range body.SantriIDs {
		// Check if already has same potongan
		var existing int
		config.DB.QueryRow("SELECT COUNT(*) FROM pembayaran_potongan WHERE tenant_id = ? AND santri_id = ? AND nama = ? AND aktif = 1", tid, sid, body.Nama).Scan(&existing)
		if existing > 0 {
			continue
		}
		_, err := config.DB.Exec("INSERT INTO pembayaran_potongan (tenant_id, santri_id, nama, nominal) VALUES (?,?,?,?)", tid, sid, body.Nama, body.Nominal)
		if err == nil {
			count++
		}
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Potongan ditambahkan untuk %d santri", count)})
}

func DeletePotongan(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("UPDATE pembayaran_potongan SET aktif = 0 WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Potongan dihapus"})
}

// ── PEMBAYARAN ─────────────────────────────────────────

func GetPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	bulan := c.Query("bulan")
	if bulan == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Parameter bulan wajib (format: 2026-04)"})
	}

	// Get all active santri
	santriRows, err := config.DB.Query("SELECT s.id, s.nama, COALESCE(k.nama,'-'), COALESCE(pk.nama, 'Default') FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id LEFT JOIN pembayaran_kategori pk ON s.kategori_spp_id = pk.id WHERE s.tenant_id = ? AND s.status = 'aktif' ORDER BY s.nama", tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer santriRows.Close()

	var result []fiber.Map
	for santriRows.Next() {
		var sid int
		var nama, kamar, katNama string
		santriRows.Scan(&sid, &nama, &kamar, &katNama)

		tagihan := HitungTagihan(tid, sid, bulan)

		// Get total bayar for this bulan
		var totalBayar int64
		config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, sid, bulan).Scan(&totalBayar)

		// Status
		status := "BELUM BAYAR"
		kekurangan := tagihan - totalBayar
		if tagihan == 0 {
			status = "GRATIS"
			kekurangan = 0
		} else if totalBayar >= tagihan {
			status = "LUNAS"
			kekurangan = 0
		} else if totalBayar > 0 {
			status = "KURANG"
		}

		potongan := hitungTotalPotongan(tid, sid)

		result = append(result, fiber.Map{
			"santri_id":      sid,
			"nama":           nama,
			"kamar":          kamar,
			"kategori_nama":  katNama,
			"tarif":          tagihan + potongan,
			"potongan":       potongan,
			"tagihan":        tagihan,
			"total_bayar":    totalBayar,
			"kekurangan":     kekurangan,
			"status":         status,
		})
	}
	if result == nil {
		result = []fiber.Map{}
	}
	return c.JSON(result)
}

func BulkBayarSPP(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)

	var payload struct {
		SantriIDs []int    `json:"santri_ids"`
		Bulans    []string `json:"bulans"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request payload"})
	}

	if len(payload.SantriIDs) == 0 || len(payload.Bulans) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Pilih minimal 1 santri dan 1 bulan"})
	}

	var grandTotal int64
	for _, sid := range payload.SantriIDs {
		var santriNama string
		config.DB.QueryRow("SELECT nama FROM santri WHERE id = ? AND tenant_id = ?", sid, tid).Scan(&santriNama)

		var paidBulans []string
		var santriTotal int64
		for _, bulan := range payload.Bulans {
			tagihan := HitungTagihan(tid, sid, bulan)
			if tagihan <= 0 {
				continue
			}

			var totalBayar int64
			config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, sid, bulan).Scan(&totalBayar)

			kekurangan := tagihan - totalBayar
			if kekurangan > 0 {
				_, err := config.DB.Exec("INSERT INTO pembayaran (tenant_id, santri_id, bulan, nominal, created_by) VALUES (?, ?, ?, ?, ?)", tid, sid, bulan, kekurangan, uid)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"message": err.Error()})
				}
				santriTotal += kekurangan
				paidBulans = append(paidBulans, bulan)
			}
		}

		// Auto-insert to catatan_keuangan
		if santriTotal > 0 && len(paidBulans) > 0 {
			sort.Strings(paidBulans)
			var ket string
			if len(paidBulans) == 1 {
				ket = fmt.Sprintf("SPP %s a.n. %s", paidBulans[0], santriNama)
			} else {
				ket = fmt.Sprintf("SPP %s s/d %s a.n. %s", paidBulans[0], paidBulans[len(paidBulans)-1], santriNama)
			}
			config.DB.Exec("INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?,?,?,?,?,?)",
				tid, "masuk", santriTotal, ket, helpers.TodayWIB(), uid)
			grandTotal += santriTotal
		}
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Pembayaran massal berhasil disimpan (Rp %d)", grandTotal)})
}

func CreatePembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	var body struct {
		SantriID   int    `json:"santri_id"`
		Bulan      string `json:"bulan"`
		Nominal    int64  `json:"nominal"`
		Metode     string `json:"metode"`
		Keterangan string `json:"keterangan"`
	}
	c.BodyParser(&body)
	if body.SantriID == 0 || body.Bulan == "" || body.Nominal <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Santri, bulan, dan nominal wajib diisi"})
	}
	if body.Metode == "" {
		body.Metode = "tunai"
	}
	res, err := config.DB.Exec("INSERT INTO pembayaran (tenant_id, santri_id, bulan, nominal, metode, keterangan, created_by) VALUES (?,?,?,?,?,?,?)",
		tid, body.SantriID, body.Bulan, body.Nominal, body.Metode, body.Keterangan, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	id, _ := res.LastInsertId()

	// Auto-insert to catatan_keuangan
	var santriNama string
	config.DB.QueryRow("SELECT nama FROM santri WHERE id = ? AND tenant_id = ?", body.SantriID, tid).Scan(&santriNama)
	ket := fmt.Sprintf("SPP %s a.n. %s", body.Bulan, santriNama)
	if body.Keterangan != "" {
		ket += " - " + body.Keterangan
	}
	config.DB.Exec("INSERT INTO catatan_keuangan (tenant_id, tipe, nominal, keterangan, tanggal, created_by) VALUES (?,?,?,?,?,?)",
		tid, "masuk", body.Nominal, ket, helpers.TodayWIB(), uid)

	// Check status after payment using helper
	tagihan := HitungTagihan(tid, body.SantriID, body.Bulan)
	var totalBayar int64
	config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, body.SantriID, body.Bulan).Scan(&totalBayar)

	status := "KURANG"
	if totalBayar >= tagihan {
		status = "LUNAS"
	}

	return c.JSON(fiber.Map{"id": id, "message": "Pembayaran dicatat", "status": status})
}

func DeletePembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	config.DB.Exec("DELETE FROM pembayaran WHERE id = ? AND tenant_id = ?", c.Params("id"), tid)
	return c.JSON(fiber.Map{"message": "Pembayaran dihapus"})
}

// ── REKAP ──────────────────────────────────────────────

func GetRekapPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	bulan := c.Query("bulan")
	if bulan == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Parameter bulan wajib"})
	}

	// Total santri aktif
	var totalSantri int
	config.DB.QueryRow("SELECT COUNT(*) FROM santri WHERE tenant_id = ? AND status = 'aktif'", tid).Scan(&totalSantri)

	// Total bayar bulan ini
	var totalBayar int64
	config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND bulan = ?", tid, bulan).Scan(&totalBayar)

	// Count lunas/kurang/belum — hitung per santri pakai helper
	santriRows, _ := config.DB.Query("SELECT id FROM santri WHERE tenant_id = ? AND status = 'aktif'", tid)
	defer santriRows.Close()
	lunas := 0
	belum := 0
	kurang := 0
	var totalTagihan int64
	var totalPotonganAll int64
	for santriRows.Next() {
		var sid int
		santriRows.Scan(&sid)
		tag := HitungTagihan(tid, sid, bulan)
		totalTagihan += tag
		totalPotonganAll += hitungTotalPotongan(tid, sid)
		var bayar int64
		config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, sid, bulan).Scan(&bayar)
		if tag == 0 || bayar >= tag {
			lunas++
		} else if bayar > 0 {
			kurang++
		} else {
			belum++
		}
	}

	persen := 0
	if totalTagihan > 0 {
		persen = int(totalBayar * 100 / totalTagihan)
		if persen > 100 {
			persen = 100
		}
	}

	return c.JSON(fiber.Map{
		"bulan":          bulan,
		"total_tarif":    hitungTotalTarif(tid, bulan),
		"total_santri":   totalSantri,
		"total_tagihan":  totalTagihan,
		"total_bayar":    totalBayar,
		"total_potongan": totalPotonganAll,
		"persen":         persen,
		"lunas":          lunas,
		"kurang":         kurang,
		"belum":          belum,
	})
}

// ── EXPORT EXCEL ───────────────────────────────────────

func ExportPembayaranExcel(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)

	// Support both single bulan (backward compat) and period (bulan_mulai + bulan_akhir)
	bulanMulai := c.Query("bulan_mulai")
	bulanAkhir := c.Query("bulan_akhir")
	singleBulan := c.Query("bulan")

	if bulanMulai == "" && singleBulan != "" {
		bulanMulai = singleBulan
		bulanAkhir = singleBulan
	}
	if bulanMulai == "" || bulanAkhir == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Parameter bulan_mulai dan bulan_akhir wajib"})
	}
	if bulanMulai > bulanAkhir {
		return c.Status(400).JSON(fiber.Map{"message": "bulan_mulai harus <= bulan_akhir"})
	}

	// Generate list of months
	months := []string{}
	cur := bulanMulai
	for cur <= bulanAkhir {
		months = append(months, cur)
		// Parse and increment month
		parts := []byte(cur)
		y := int(parts[0]-'0')*1000 + int(parts[1]-'0')*100 + int(parts[2]-'0')*10 + int(parts[3]-'0')
		m := int(parts[5]-'0')*10 + int(parts[6]-'0')
		m++
		if m > 12 {
			m = 1
			y++
		}
		cur = fmt.Sprintf("%04d-%02d", y, m)
	}

	// Filters
	kid := c.Query("kamar_id")
	kelasID := c.Query("kelas_id")

	// Build query to get santri with kelas diniyyah
	q := `SELECT s.id, s.nama, COALESCE(k.nama,'-'), COALESCE(kd.nama, '-')
		FROM santri s 
		LEFT JOIN kamar k ON s.kamar_id = k.id
		LEFT JOIN santri_kelas_diniyyah skd2 ON s.id = skd2.santri_id AND skd2.tenant_id = s.tenant_id AND skd2.status = 'active'
		LEFT JOIN kelas_diniyyah kd ON skd2.kelas_diniyyah_id = kd.id`

	if kelasID != "" {
		q += " INNER JOIN santri_kelas_diniyyah skd ON s.id = skd.santri_id AND skd.kelas_diniyyah_id = ? AND skd.status = 'active'"
	}

	q += " WHERE s.tenant_id = ? AND s.status = 'aktif'"
	args := []interface{}{}
	if kelasID != "" {
		args = append(args, kelasID)
	}
	args = append(args, tid)

	if kid != "" {
		q += " AND s.kamar_id = ?"
		args = append(args, kid)
	}
	q += " ORDER BY k.nama, s.nama"

	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	// Collect santri data
	type santriRow struct {
		ID    int
		Nama  string
		Kamar string
		Kelas string
	}
	var santriList []santriRow
	for rows.Next() {
		var s santriRow
		rows.Scan(&s.ID, &s.Nama, &s.Kamar, &s.Kelas)
		if s.Kamar == "" || s.Kamar == "-" {
			s.Kamar = "Belum Ada Kamar"
		}
		if s.Kelas == "" || s.Kelas == "-" {
			s.Kelas = "-"
		}
		santriList = append(santriList, s)
	}

	f := excelize.NewFile()
	sheet := "Rekap Pembayaran"
	f.SetSheetName("Sheet1", sheet)

	// ── Build headers ──
	// Fixed: No | Nama | Kelas | Kamar | Potongan
	// Per bulan: Tagihan | Bayar | Status
	// End: Total Tagihan | Total Bayar | Sisa
	headers := []string{"No", "Nama Santri", "Kelas", "Kamar", "Potongan"}
	for _, m := range months {
		// Format bulan label (e.g. "Jan 2026")
		label := m // fallback
		parts := []byte(m)
		y := string(parts[0:4])
		mi := int(parts[5]-'0')*10 + int(parts[6]-'0')
		monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Ags", "Sep", "Okt", "Nov", "Des"}
		if mi >= 1 && mi <= 12 {
			label = monthNames[mi] + " " + y
		}
		headers = append(headers, label+" Tagihan", label+" Bayar", label+" Status")
	}
	headers = append(headers, "Total Tagihan", "Total Bayar", "Sisa")

	colIdx := func(i int) string {
		if i < 26 {
			return string(rune('A' + i))
		}
		return string(rune('A'+i/26-1)) + string(rune('A'+i%26))
	}

	for i, h := range headers {
		f.SetCellValue(sheet, colIdx(i)+"1", h)
	}

	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#16a34a"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", WrapText: true},
	})
	f.SetCellStyle(sheet, "A1", colIdx(len(headers)-1)+"1", headerStyle)

	// Number format style for currency
	numStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 3}) // #,##0

	row := 2
	no := 1
	var grandTagihan, grandBayar int64
	var subTagihan, subBayar int64
	currentKamar := ""
	firstKamar := true

	for _, s := range santriList {
		if currentKamar != s.Kamar {
			if !firstKamar {
				// Subtotal row for previous kamar
				subStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#f3f4f6"}, Pattern: 1}})
				f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "TOTAL KAMAR "+currentKamar)
				totalColStart := 5 + len(months)*3
				f.SetCellValue(sheet, colIdx(totalColStart)+fmt.Sprintf("%d", row), subTagihan)
				f.SetCellValue(sheet, colIdx(totalColStart+1)+fmt.Sprintf("%d", row), subBayar)
				f.SetCellValue(sheet, colIdx(totalColStart+2)+fmt.Sprintf("%d", row), subTagihan-subBayar)
				f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), colIdx(len(headers)-1)+fmt.Sprintf("%d", row), subStyle)
				row++
			}
			// Kamar header
			kamarStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "#16a34a"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#dcfce7"}, Pattern: 1}})
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "KAMAR: "+s.Kamar)
			f.MergeCell(sheet, fmt.Sprintf("A%d", row), colIdx(len(headers)-1)+fmt.Sprintf("%d", row))
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), colIdx(len(headers)-1)+fmt.Sprintf("%d", row), kamarStyle)
			row++

			currentKamar = s.Kamar
			subTagihan = 0
			subBayar = 0
			firstKamar = false
		}

		pot := hitungTotalPotongan(tid, s.ID)

		// Fixed columns
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), no)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), s.Nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), s.Kelas)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), s.Kamar)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), pot)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), numStyle)

		var santriTotalTagihan, santriTotalBayar int64

		for mi, bulan := range months {
			tagihan := HitungTagihan(tid, s.ID, bulan)
			var bayar int64
			config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ? AND bulan = ?", tid, s.ID, bulan).Scan(&bayar)

			status := "BELUM"
			if tagihan == 0 {
				status = "GRATIS"
			} else if bayar >= tagihan {
				status = "LUNAS"
			} else if bayar > 0 {
				status = "KURANG"
			}

			colBase := 5 + mi*3
			f.SetCellValue(sheet, colIdx(colBase)+fmt.Sprintf("%d", row), tagihan)
			f.SetCellValue(sheet, colIdx(colBase+1)+fmt.Sprintf("%d", row), bayar)
			f.SetCellValue(sheet, colIdx(colBase+2)+fmt.Sprintf("%d", row), status)
			f.SetCellStyle(sheet, colIdx(colBase)+fmt.Sprintf("%d", row), colIdx(colBase+1)+fmt.Sprintf("%d", row), numStyle)

			// Color status cell
			var stColor string
			switch status {
			case "LUNAS":
				stColor = "#16a34a"
			case "KURANG":
				stColor = "#f59e0b"
			case "BELUM":
				stColor = "#ef4444"
			case "GRATIS":
				stColor = "#6b7280"
			}
			if stColor != "" {
				stStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: stColor, Size: 9}, Alignment: &excelize.Alignment{Horizontal: "center"}})
				f.SetCellStyle(sheet, colIdx(colBase+2)+fmt.Sprintf("%d", row), colIdx(colBase+2)+fmt.Sprintf("%d", row), stStyle)
			}

			santriTotalTagihan += tagihan
			santriTotalBayar += bayar
		}

		// Total columns
		totalColStart := 5 + len(months)*3
		sisa := santriTotalTagihan - santriTotalBayar
		if sisa < 0 {
			sisa = 0
		}
		f.SetCellValue(sheet, colIdx(totalColStart)+fmt.Sprintf("%d", row), santriTotalTagihan)
		f.SetCellValue(sheet, colIdx(totalColStart+1)+fmt.Sprintf("%d", row), santriTotalBayar)
		f.SetCellValue(sheet, colIdx(totalColStart+2)+fmt.Sprintf("%d", row), sisa)
		f.SetCellStyle(sheet, colIdx(totalColStart)+fmt.Sprintf("%d", row), colIdx(totalColStart+2)+fmt.Sprintf("%d", row), numStyle)

		grandTagihan += santriTotalTagihan
		grandBayar += santriTotalBayar
		subTagihan += santriTotalTagihan
		subBayar += santriTotalBayar

		row++
		no++
	}

	// Last kamar subtotal
	if !firstKamar {
		subStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#f3f4f6"}, Pattern: 1}})
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "TOTAL KAMAR "+currentKamar)
		totalColStart := 5 + len(months)*3
		f.SetCellValue(sheet, colIdx(totalColStart)+fmt.Sprintf("%d", row), subTagihan)
		f.SetCellValue(sheet, colIdx(totalColStart+1)+fmt.Sprintf("%d", row), subBayar)
		f.SetCellValue(sheet, colIdx(totalColStart+2)+fmt.Sprintf("%d", row), subTagihan-subBayar)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), colIdx(len(headers)-1)+fmt.Sprintf("%d", row), subStyle)
		row++
	}

	// Grand total
	totalStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11}})
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "TOTAL KESELURUHAN")
	totalColStart := 5 + len(months)*3
	f.SetCellValue(sheet, colIdx(totalColStart)+fmt.Sprintf("%d", row), grandTagihan)
	f.SetCellValue(sheet, colIdx(totalColStart+1)+fmt.Sprintf("%d", row), grandBayar)
	f.SetCellValue(sheet, colIdx(totalColStart+2)+fmt.Sprintf("%d", row), grandTagihan-grandBayar)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), colIdx(len(headers)-1)+fmt.Sprintf("%d", row), totalStyle)

	// Column widths
	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "B", "B", 22)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 14)
	f.SetColWidth(sheet, "E", "E", 12)
	// Per-bulan columns
	for i := 0; i < len(months)*3; i++ {
		col := colIdx(5 + i)
		f.SetColWidth(sheet, col, col, 13)
	}
	// Total columns
	for i := 0; i < 3; i++ {
		col := colIdx(5 + len(months)*3 + i)
		f.SetColWidth(sheet, col, col, 15)
	}

	filename := fmt.Sprintf("Rekap_Pembayaran_%s_sd_%s.xlsx", bulanMulai, bulanAkhir)
	if bulanMulai == bulanAkhir {
		filename = fmt.Sprintf("Rekap_Pembayaran_%s.xlsx", bulanMulai)
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return f.Write(c.Response().BodyWriter())
}

// ── RIWAYAT PEMBAYARAN PER SANTRI ──────────────────────

func GetRiwayatPembayaran(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	bulan := c.Query("bulan")
	santriID := c.Query("santri_id")
	q := `SELECT id, santri_id, nama, bulan, nominal, metode, keterangan, created_at FROM (
			SELECT p.id, p.santri_id, s.nama, CAST(p.bulan AS CHAR) as bulan, p.nominal, CAST(p.metode AS CHAR) as metode, CAST(COALESCE(p.keterangan,'') AS CHAR) as keterangan, CAST(p.created_at AS CHAR) as created_at 
			FROM pembayaran p LEFT JOIN santri s ON p.santri_id = s.id 
			WHERE p.tenant_id = ?
			UNION ALL
			SELECT pi.id, pi.santri_id, s.nama, CAST(CONCAT('Insidental: ', COALESCE(ti.nama, '')) AS CHAR) as bulan, pi.nominal_dibayar as nominal, CAST(pi.metode AS CHAR) as metode, CAST('Tagihan Insidental' AS CHAR) as keterangan, CAST(pi.tanggal AS CHAR) as created_at 
			FROM pembayaran_insidental pi
			LEFT JOIN santri s ON pi.santri_id = s.id
			LEFT JOIN tagihan_insidental_santri tis ON pi.tagihan_santri_id = tis.id
			LEFT JOIN tagihan_insidental ti ON tis.tagihan_id = ti.id
			WHERE pi.tenant_id = ?
		) combined
		WHERE 1=1`
	args := []interface{}{tid, tid}
	if bulan != "" {
		q += " AND bulan = ?"
		args = append(args, bulan)
	}
	if santriID != "" {
		q += " AND santri_id = ?"
		args = append(args, santriID)
	}
	q += " ORDER BY created_at DESC LIMIT 200"
	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, santriId int
		var nama, bln, metode, ket string
		var nominal int64
		var createdAt string
		rows.Scan(&id, &santriId, &nama, &bln, &nominal, &metode, &ket, &createdAt)
		list = append(list, fiber.Map{
			"id": id, "santri_id": santriId, "santri_nama": nama,
			"bulan": bln, "nominal": nominal, "metode": metode,
			"keterangan": ket, "created_at": createdAt,
		})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}
