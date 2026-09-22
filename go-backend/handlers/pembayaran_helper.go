package handlers

import (
	"pesantren-multi/config"
)

// HitungTagihan calculates the tagihan for a specific santri in a specific bulan.
// Formula: SUM(tarif_per_bulan) + kategori_nominal_per_bulan - SUM(potongan)
// Tarif & kategori nominal use periode override if available, otherwise fallback to default.
func HitungTagihan(tenantID int, santriID int, bulan string) int64 {
	totalTarif := hitungTotalTarif(tenantID, bulan)
	katNominal := hitungKategoriNominal(tenantID, santriID, bulan)
	totalPotongan := hitungTotalPotongan(tenantID, santriID)

	tagihan := totalTarif + katNominal - totalPotongan
	if tagihan < 0 {
		tagihan = 0
	}
	return tagihan
}

// hitungTotalTarif calculates the sum of all active tarif for a given bulan.
// For each tarif, checks if there's a periode override for the bulan,
// otherwise uses the default nominal.
func hitungTotalTarif(tenantID int, bulan string) int64 {
	rows, err := config.DB.Query(
		"SELECT id, nominal FROM pembayaran_tarif WHERE tenant_id = ? AND aktif = 1", tenantID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var total int64
	for rows.Next() {
		var tarifID int
		var defaultNominal int64
		rows.Scan(&tarifID, &defaultNominal)

		// Check for periode override
		var periodeNominal int64
		var found bool
		err := config.DB.QueryRow(
			"SELECT nominal FROM pembayaran_tarif_periode WHERE tenant_id = ? AND tarif_id = ? AND bulan_mulai <= ? AND bulan_akhir >= ?",
			tenantID, tarifID, bulan, bulan).Scan(&periodeNominal)
		if err == nil {
			found = true
		}

		if found {
			total += periodeNominal
		} else {
			total += defaultNominal
		}
	}
	return total
}

// hitungKategoriNominal calculates the kategori nominal for a santri in a given bulan.
// Checks if there's a periode override, otherwise uses the default nominal.
func hitungKategoriNominal(tenantID int, santriID int, bulan string) int64 {
	var kategoriID int
	var defaultNominal int64
	err := config.DB.QueryRow(
		`SELECT pk.id, pk.nominal FROM santri s 
		 JOIN pembayaran_kategori pk ON s.kategori_spp_id = pk.id AND pk.aktif = 1
		 WHERE s.id = ? AND s.tenant_id = ?`, santriID, tenantID).Scan(&kategoriID, &defaultNominal)
	if err != nil {
		return 0 // No kategori assigned or not found
	}

	// Check for periode override
	var periodeNominal int64
	err = config.DB.QueryRow(
		"SELECT nominal FROM pembayaran_kategori_periode WHERE tenant_id = ? AND kategori_id = ? AND bulan_mulai <= ? AND bulan_akhir >= ?",
		tenantID, kategoriID, bulan, bulan).Scan(&periodeNominal)
	if err == nil {
		return periodeNominal
	}

	return defaultNominal
}

// hitungTotalPotongan calculates total active potongan for a santri.
func hitungTotalPotongan(tenantID int, santriID int) int64 {
	var total int64
	config.DB.QueryRow(
		"SELECT COALESCE(SUM(nominal),0) FROM pembayaran_potongan WHERE tenant_id = ? AND santri_id = ? AND aktif = 1",
		tenantID, santriID).Scan(&total)
	return total
}

// HitungTagihanBatch calculates tagihan for multiple santri in a specific bulan.
// More efficient than calling HitungTagihan in a loop because it fetches tarif once.
// Returns map[santriID]tagihan
func HitungTagihanBatch(tenantID int, santriIDs []int, bulan string) map[int]int64 {
	result := make(map[int]int64)
	if len(santriIDs) == 0 {
		return result
	}

	// 1. Calculate total tarif for this bulan (same for all santri)
	totalTarif := hitungTotalTarif(tenantID, bulan)

	// 2. For each santri, get kategori nominal + potongan
	for _, sid := range santriIDs {
		katNominal := hitungKategoriNominal(tenantID, sid, bulan)
		potongan := hitungTotalPotongan(tenantID, sid)

		tagihan := totalTarif + katNominal - potongan
		if tagihan < 0 {
			tagihan = 0
		}
		result[sid] = tagihan
	}

	return result
}
