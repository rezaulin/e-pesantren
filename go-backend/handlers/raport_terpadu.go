package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"pesantren-multi/config"

	"github.com/gofiber/fiber/v2"
)

// GetRaportTerpaduData returns JSON data containing merged Nilai + Absensi for rendering in the browser
func GetRaportTerpaduData(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	kategori := c.Query("kategori", "sekolah") // sekolah, diniyyah, kegiatan
	semester := c.Query("semester")
	bulan := c.Query("bulan")
	santriID := c.Query("santri_id")
	kelasID := c.Query("kelas_id") // also used for kamar_id if kategori is kegiatan

	if semester == "" && kategori != "kegiatan" {
		return c.Status(400).JSON(fiber.Map{"message": "semester wajib dipilih"})
	}
	if kategori == "kegiatan" && bulan == "" {
		return c.Status(400).JSON(fiber.Map{"message": "bulan wajib dipilih untuk kategori kegiatan"})
	}

	// 1. Get Settings
	settings := getSettings(tid)

	// 2. Build Santri Query
	var santriQuery string
	var santriArgs []interface{}

	if santriID != "" {
		santriQuery = "SELECT id, nama FROM santri WHERE tenant_id = ? AND id = ?"
		santriArgs = []interface{}{tid, santriID}
	} else if kelasID != "" {
		if kategori == "sekolah" {
			santriQuery = `SELECT s.id, s.nama FROM santri s
				INNER JOIN santri_kelas sk ON s.id = sk.santri_id AND sk.tenant_id = s.tenant_id AND sk.status = 'active'
				WHERE s.tenant_id = ? AND s.status = 'aktif' AND sk.kelas_id = ? ORDER BY s.nama`
			santriArgs = []interface{}{tid, kelasID}
		} else if kategori == "diniyyah" {
			santriQuery = `SELECT s.id, s.nama FROM santri s
				INNER JOIN santri_kelas_diniyyah skd ON s.id = skd.santri_id AND skd.tenant_id = s.tenant_id AND skd.status = 'active'
				WHERE s.tenant_id = ? AND s.status = 'aktif' AND skd.kelas_diniyyah_id = ? ORDER BY s.nama`
			santriArgs = []interface{}{tid, kelasID}
		} else if kategori == "kegiatan" {
			santriQuery = `SELECT id, nama FROM santri WHERE tenant_id = ? AND status = 'aktif' AND kamar_id = ? ORDER BY nama`
			santriArgs = []interface{}{tid, kelasID}
		}
	} else {
		// All active santri
		santriQuery = "SELECT id, nama FROM santri WHERE tenant_id = ? AND status = 'aktif' ORDER BY nama"
		santriArgs = []interface{}{tid}
	}

	rows, err := config.DB.Query(santriQuery, santriArgs...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()

	var result []fiber.Map

	for rows.Next() {
		var id int
		var nama string
		rows.Scan(&id, &nama)

		// 3. Get Identitas Tambahan
		var kamarNama, kelasSekolah, kelasDiniyyah string
		config.DB.QueryRow(`
			SELECT COALESCE(k.nama, ''), 
				(SELECT GROUP_CONCAT(ks.nama SEPARATOR ', ') FROM santri_kelas sk JOIN kelas_sekolah ks ON sk.kelas_id = ks.id WHERE sk.santri_id = s.id AND sk.status = 'active'),
				(SELECT GROUP_CONCAT(kd.nama SEPARATOR ', ') FROM santri_kelas_diniyyah skd JOIN kelas_diniyyah kd ON skd.kelas_diniyyah_id = kd.id WHERE skd.santri_id = s.id AND skd.status = 'active')
			FROM santri s
			LEFT JOIN kamar k ON s.kamar_id = k.id
			WHERE s.id = ? AND s.tenant_id = ?
		`, id, tid).Scan(&kamarNama, &kelasSekolah, &kelasDiniyyah)
		if kelasSekolah == "" { kelasSekolah = "-" }
		if kelasDiniyyah == "" { kelasDiniyyah = "-" }
		if kamarNama == "" { kamarNama = "-" }

		santriData := fiber.Map{
			"id":             id,
			"nama":           nama,
			"kamar":          kamarNama,
			"kelas_sekolah":  kelasSekolah,
			"kelas_diniyyah": kelasDiniyyah,
		}

		// 4. Get Nilai & Absensi based on kategori
		var nilaiList []fiber.Map
		var avg float64
		var absensi []fiber.Map

		if kategori == "sekolah" {
			// Nilai Sekolah
			nsRows, _ := config.DB.Query(`SELECT mp.nama, COALESCE(ns.nilai_harian, 0), COALESCE(ns.nilai_uts, 0), COALESCE(ns.nilai_uas, 0), COALESCE(ns.nilai_akhir, 0), COALESCE(ns.nilai_detail,'{}')
				FROM nilai_sekolah ns JOIN mata_pelajaran_sekolah mp ON ns.mata_pelajaran_sekolah_id = mp.id
				WHERE ns.tenant_id = ? AND ns.santri_id = ? AND ns.semester = ? ORDER BY mp.nama`, tid, id, semester)
			var total float64
			var count int
			if nsRows != nil {
				for nsRows.Next() {
					var mNama, det string
					var h, ut, ua, ak float64
					nsRows.Scan(&mNama, &h, &ut, &ua, &ak, &det)
					var detail map[string]interface{}
					json.Unmarshal([]byte(det), &detail)
					nilaiList = append(nilaiList, fiber.Map{
						"mapel":  mNama,
						"harian": h,
						"uts":    ut,
						"uas":    ua,
						"akhir":  ak,
						"detail": detail,
					})
					total += ak
					count++
				}
				nsRows.Close()
			}
			if count > 0 { avg = math.Round((total/float64(count))*100) / 100 }

			// Absensi Sekolah
			abRows, _ := config.DB.Query(`SELECT status, COUNT(*) FROM absen_sekolah WHERE tenant_id = ? AND santri_id = ? GROUP BY status`, tid, id)
			var hadir, izin, sakit, alpa int
			if abRows != nil {
				for abRows.Next() {
					var st string
					var c int
					abRows.Scan(&st, &c)
					switch st {
					case "H": hadir += c
					case "I": izin += c
					case "S": sakit += c
					case "A": alpa += c
					}
				}
				abRows.Close()
			}
			absensi = append(absensi, fiber.Map{"kegiatan": "Sekolah Formal", "H": hadir, "I": izin, "S": sakit, "A": alpa})

		} else if kategori == "diniyyah" {
			// Nilai Diniyyah
			ndRows, _ := config.DB.Query(`SELECT mp.nama, COALESCE(nd.nilai_harian, 0), COALESCE(nd.nilai_uts, 0), COALESCE(nd.nilai_uas, 0), COALESCE(nd.nilai_akhir, 0), COALESCE(nd.nilai_detail,'{}')
				FROM nilai_pelajaran nd JOIN mata_pelajaran mp ON nd.mata_pelajaran_id = mp.id
				WHERE nd.tenant_id = ? AND nd.santri_id = ? AND nd.semester = ? ORDER BY mp.nama`, tid, id, semester)
			var total float64
			var count int
			if ndRows != nil {
				for ndRows.Next() {
					var mNama, det string
					var h, ut, ua, ak float64
					ndRows.Scan(&mNama, &h, &ut, &ua, &ak, &det)
					var detail map[string]interface{}
					json.Unmarshal([]byte(det), &detail)
					nilaiList = append(nilaiList, fiber.Map{
						"mapel":  mNama,
						"harian": h,
						"uts":    ut,
						"uas":    ua,
						"akhir":  ak,
						"detail": detail,
					})
					total += ak
					count++
				}
				ndRows.Close()
			}
			if count > 0 { avg = math.Round((total/float64(count))*100) / 100 }

			// Absensi Diniyyah
			abRows, _ := config.DB.Query(`SELECT status, COUNT(*) FROM absen_diniyyah WHERE tenant_id = ? AND santri_id = ? GROUP BY status`, tid, id)
			var hadir, izin, sakit, alpa int
			if abRows != nil {
				for abRows.Next() {
					var st string
					var c int
					abRows.Scan(&st, &c)
					switch st {
					case "H": hadir += c
					case "I": izin += c
					case "S": sakit += c
					case "A": alpa += c
					}
				}
				abRows.Close()
			}
			absensi = append(absensi, fiber.Map{"kegiatan": "Madrasah Diniyyah", "H": hadir, "I": izin, "S": sakit, "A": alpa})

		} else if kategori == "kegiatan" {
			// Nilai Kegiatan
			nkRows, _ := config.DB.Query(`SELECT kg.nama, kl.nama, COALESCE(nk.nilai, 0), COALESCE(nk.catatan,'')
				FROM nilai_kegiatan nk
				JOIN kegiatan kg ON nk.kegiatan_id = kg.id
				JOIN kelompok kl ON nk.kelompok_id = kl.id
				WHERE nk.tenant_id = ? AND nk.santri_id = ? AND nk.bulan = ? ORDER BY kg.nama`, tid, id, bulan)
			var total float64
			var count int
			if nkRows != nil {
				for nkRows.Next() {
					var kNama, klNama, cat string
					var n float64
					nkRows.Scan(&kNama, &klNama, &n, &cat)
					nilaiList = append(nilaiList, fiber.Map{
						"kegiatan": kNama,
						"kelompok": klNama,
						"nilai":    n,
						"catatan":  cat,
					})
					total += n
					count++
				}
				nkRows.Close()
			}
			if count > 0 { avg = math.Round((total/float64(count))*100) / 100 }

			// Absensi Kegiatan
			abRows, _ := config.DB.Query(`SELECT COALESCE(k.kegiatan_nama,'Lainnya'), a.status, COUNT(*) 
				FROM absensi a LEFT JOIN kelompok k ON a.kelompok_id = k.id 
				WHERE a.tenant_id = ? AND a.santri_id = ? AND a.tanggal LIKE ?
				GROUP BY k.kegiatan_nama, a.status`, tid, id, bulan+"%")
			type Rekap struct{ H, I, S, A int }
			rekap := make(map[string]*Rekap)
			if abRows != nil {
				for abRows.Next() {
					var keg, st string
					var c int
					abRows.Scan(&keg, &st, &c)
					if rekap[keg] == nil { rekap[keg] = &Rekap{} }
					switch st {
					case "H": rekap[keg].H += c
					case "I": rekap[keg].I += c
					case "S": rekap[keg].S += c
					case "A": rekap[keg].A += c
					}
				}
				abRows.Close()
			}
			for k, r := range rekap {
				absensi = append(absensi, fiber.Map{"kegiatan": k, "H": r.H, "I": r.I, "S": r.S, "A": r.A})
			}
		}

		if nilaiList == nil { nilaiList = []fiber.Map{} }
		if absensi == nil { absensi = []fiber.Map{} }

		santriData["nilai"] = nilaiList
		santriData["rata_rata"] = avg
		santriData["absensi"] = absensi

		result = append(result, santriData)
	}

	if result == nil { result = []fiber.Map{} }

	return c.JSON(fiber.Map{
		"settings": settings,
		"data":     result,
	})
}

func GetRaportTerpaduOptions(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	var optSekolah, optDiniyyah, optKamar string

	rowsS, _ := config.DB.Query("SELECT id, nama FROM kelas_sekolah WHERE tenant_id = ? ORDER BY nama", tid)
	if rowsS != nil {
		for rowsS.Next() {
			var id int
			var nama string
			rowsS.Scan(&id, &nama)
			optSekolah += fmt.Sprintf(`<option value="%d">%s</option>`, id, nama)
		}
		rowsS.Close()
	}

	rowsD, _ := config.DB.Query("SELECT id, nama FROM kelas_diniyyah WHERE tenant_id = ? ORDER BY nama", tid)
	if rowsD != nil {
		for rowsD.Next() {
			var id int
			var nama string
			rowsD.Scan(&id, &nama)
			optDiniyyah += fmt.Sprintf(`<option value="%d">%s</option>`, id, nama)
		}
		rowsD.Close()
	}

	rowsK, _ := config.DB.Query("SELECT id, nama FROM kamar WHERE tenant_id = ? ORDER BY nama", tid)
	if rowsK != nil {
		for rowsK.Next() {
			var id int
			var nama string
			rowsK.Scan(&id, &nama)
			optKamar += fmt.Sprintf(`<option value="%d">%s</option>`, id, nama)
		}
		rowsK.Close()
	}

	type SantriMin struct {
		ID            int    `json:"id"`
		Nama          string `json:"nama"`
		KamarID       int    `json:"kamar_id"`
		KamarNama     string `json:"kamar_nama"`
		KelasSekolah  string `json:"kelas_sekolah"`
		KelasDiniyyah string `json:"kelas_diniyyah"`
		KelasSekolahID  int    `json:"kelas_sekolah_id"`
		KelasDiniyyahID int    `json:"kelas_diniyyah_id"`
	}

	var santriList []SantriMin
	rows, err := config.DB.Query(`
		SELECT s.id, s.nama, 
		       COALESCE(s.kamar_id, 0), COALESCE(k.nama, ''), 
		       COALESCE((SELECT kelas_id FROM santri_kelas sk WHERE sk.santri_id = s.id AND sk.tenant_id = s.tenant_id AND sk.status = 'active' LIMIT 1), 0),
		       COALESCE((SELECT ks.nama FROM santri_kelas sk JOIN kelas_sekolah ks ON sk.kelas_id = ks.id WHERE sk.santri_id = s.id AND sk.tenant_id = s.tenant_id AND sk.status = 'active' LIMIT 1), s.kelas_sekolah, ''),
		       COALESCE((SELECT kelas_diniyyah_id FROM santri_kelas_diniyyah skd WHERE skd.santri_id = s.id AND skd.tenant_id = s.tenant_id AND skd.status = 'active' LIMIT 1), 0),
		       COALESCE((SELECT kd.nama FROM santri_kelas_diniyyah skd JOIN kelas_diniyyah kd ON skd.kelas_diniyyah_id = kd.id WHERE skd.santri_id = s.id AND skd.tenant_id = s.tenant_id AND skd.status = 'active' LIMIT 1), s.kelas_diniyyah, '')
		FROM santri s
		LEFT JOIN kamar k ON s.kamar_id = k.id
		WHERE s.tenant_id = ? AND s.status = 'aktif'
		ORDER BY s.nama`, tid)
	if err == nil && rows != nil {
		for rows.Next() {
			var s SantriMin
			rows.Scan(&s.ID, &s.Nama, &s.KamarID, &s.KamarNama, &s.KelasSekolahID, &s.KelasSekolah, &s.KelasDiniyyahID, &s.KelasDiniyyah)
			santriList = append(santriList, s)
		}
		rows.Close()
	}

	if santriList == nil {
		santriList = []SantriMin{}
	}

	return c.JSON(fiber.Map{
		"optSekolah":  optSekolah,
		"optDiniyyah": optDiniyyah,
		"optKamar":    optKamar,
		"santri":      santriList,
	})
}
