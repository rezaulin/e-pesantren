package handlers

import (
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
)

// GetSantriProfile — comprehensive santri lookup: kamar, kelas, pembayaran, catatan, pelanggaran
func GetSantriProfile(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	sid := c.Params("id")

	// Basic info
	var id, kamarID int
	var nama, alamat, kelasD, status, kamarNama, noHP, namaWali string
	err := config.DB.QueryRow(`SELECT s.id, s.nama, COALESCE(s.alamat,''), COALESCE(s.kelas_diniyyah,''), 
		s.status, s.kamar_id, COALESCE(k.nama,'-'), COALESCE(s.no_hp,''), COALESCE(s.nama_wali,'')
		FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id 
		WHERE s.id = ? AND s.tenant_id = ?`, sid, tid).Scan(&id, &nama, &alamat, &kelasD, &status, &kamarID, &kamarNama, &noHP, &namaWali)
	if err != nil || nama == "" {
		return c.Status(404).JSON(fiber.Map{"message": "Santri tidak ditemukan"})
	}

	// Kelas Sekolah
	var kelasSekolahNama string
	config.DB.QueryRow(`SELECT ks.nama FROM santri_kelas sk JOIN kelas_sekolah ks ON sk.kelas_id = ks.id 
		WHERE sk.santri_id = ? AND sk.tenant_id = ? AND sk.status = 'active' LIMIT 1`, sid, tid).Scan(&kelasSekolahNama)

	// Kelas Diniyyah (from kelas_diniyyah table)
	var kelasDiniyyahNama string
	config.DB.QueryRow(`SELECT kd.nama FROM santri_kelas_diniyyah skd JOIN kelas_diniyyah kd ON skd.kelas_diniyyah_id = kd.id 
		WHERE skd.santri_id = ? AND skd.tenant_id = ? AND skd.status = 'active' LIMIT 1`, sid, tid).Scan(&kelasDiniyyahNama)

	// Pembayaran status (current month tagihan vs bayar)
	var totalTarif, totalPotongan, totalBayar, katNominal int64
	config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran_tarif WHERE tenant_id = ? AND aktif = 1", tid).Scan(&totalTarif)
	config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran_potongan WHERE tenant_id = ? AND santri_id = ? AND aktif = 1", tid, sid).Scan(&totalPotongan)
	config.DB.QueryRow("SELECT COALESCE(pk.nominal,0) FROM santri s LEFT JOIN pembayaran_kategori pk ON s.kategori_spp_id = pk.id WHERE s.id = ? AND s.tenant_id = ?", sid, tid).Scan(&katNominal)
	tagihan := totalTarif + katNominal - totalPotongan
	if tagihan < 0 {
		tagihan = 0
	}

	// Get all months with payments
	var riwayatPembayaran []fiber.Map
	pRows, _ := config.DB.Query(`SELECT bulan, SUM(nominal) as total FROM pembayaran 
		WHERE tenant_id = ? AND santri_id = ? GROUP BY bulan ORDER BY bulan DESC LIMIT 12`, tid, sid)
	if pRows != nil {
		for pRows.Next() {
			var bln string
			var total int64
			pRows.Scan(&bln, &total)
			st := "LUNAS"
			kk := tagihan - total
			if kk < 0 {
				kk = 0
			}
			if tagihan == 0 {
				st = "GRATIS"
			} else if total >= tagihan {
				st = "LUNAS"
				kk = 0
			} else if total > 0 {
				st = "KURANG"
			} else {
				st = "BELUM BAYAR"
			}
			riwayatPembayaran = append(riwayatPembayaran, fiber.Map{"bulan": bln, "total_bayar": total, "tagihan": tagihan, "kekurangan": kk, "status": st})
		}
		pRows.Close()
	}
	if riwayatPembayaran == nil {
		riwayatPembayaran = []fiber.Map{}
	}

	// Total tunggakan (months with no/partial payment)
	config.DB.QueryRow("SELECT COALESCE(SUM(nominal),0) FROM pembayaran WHERE tenant_id = ? AND santri_id = ?", tid, sid).Scan(&totalBayar)

	// Catatan Guru (all time)
	var catatanGuru []fiber.Map
	cgRows, _ := config.DB.Query(`SELECT cg.catatan, cg.created_at, COALESCE(u.nama,'') 
		FROM catatan_guru cg LEFT JOIN users u ON cg.created_by = u.id 
		WHERE cg.santri_id = ? AND cg.tenant_id = ? ORDER BY cg.created_at DESC LIMIT 20`, sid, tid)
	if cgRows != nil {
		for cgRows.Next() {
			var catatan, ca, uNama string
			cgRows.Scan(&catatan, &ca, &uNama)
			catatanGuru = append(catatanGuru, fiber.Map{"catatan": catatan, "created_at": ca, "guru": uNama})
		}
		cgRows.Close()
	}
	if catatanGuru == nil {
		catatanGuru = []fiber.Map{}
	}

	// Pelanggaran (all time)
	var pelanggaran []fiber.Map
	plRows, _ := config.DB.Query(`SELECT p.jenis, COALESCE(p.deskripsi,''), COALESCE(p.poin,0), COALESCE(p.tanggal,''), p.created_at, COALESCE(u.nama,''), COALESCE(p.status_takzir,'belum'), COALESCE(p.jenis_takzir,''), COALESCE(p.denda,0) 
		FROM pelanggaran p LEFT JOIN users u ON p.created_by = u.id 
		WHERE p.santri_id = ? AND p.tenant_id = ? ORDER BY p.created_at DESC LIMIT 20`, sid, tid)
	if plRows != nil {
		for plRows.Next() {
			var poin int
			var denda int64
			var jenis, desk, tgl, ca, uNama, stTakzir, jTakzir string
			plRows.Scan(&jenis, &desk, &poin, &tgl, &ca, &uNama, &stTakzir, &jTakzir, &denda)
			pelanggaran = append(pelanggaran, fiber.Map{"jenis": jenis, "keterangan": desk, "poin": poin, "tanggal": tgl, "created_at": ca, "pencatat": uNama, "status_takzir": stTakzir, "jenis_takzir": jTakzir, "denda": denda})
		}
		plRows.Close()
	}
	if pelanggaran == nil {
		pelanggaran = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"santri": fiber.Map{
			"id": id, "nama": nama, "alamat": alamat, "kelas_diniyyah": kelasD,
			"status": status, "kamar_id": kamarID, "kamar_nama": kamarNama,
			"no_hp": noHP, "nama_wali": namaWali,
			"kelas_sekolah":      kelasSekolahNama,
			"kelas_diniyyah_nama": kelasDiniyyahNama,
		},
		"pembayaran": fiber.Map{
			"tarif_bulanan": tagihan,
			"total_bayar":   totalBayar,
			"riwayat":       riwayatPembayaran,
		},
		"catatan_guru": catatanGuru,
		"pelanggaran":  pelanggaran,
	})
}
