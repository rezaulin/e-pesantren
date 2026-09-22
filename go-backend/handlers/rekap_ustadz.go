package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
)

func RekapUstadzSummary(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	bulan := c.Query("bulan", helpers.MonthWIB())
	dari := c.Query("dari", "")
	sampai := c.Query("sampai", "")

	// Build date filter: if dari+sampai provided, use BETWEEN; else use LIKE month%
	var dateFilter string
	var dateArgs []interface{}
	if dari != "" && sampai != "" {
		dateFilter = "BETWEEN ? AND ?"
		dateArgs = []interface{}{dari, sampai}
	} else {
		dateFilter = "LIKE ?"
		dateArgs = []interface{}{bulan + "%"}
	}

	type SesiDetail struct {
		Tanggal    string `json:"tanggal"`
		RecordedAt string `json:"recorded_at"`
		Kelompok   string `json:"kelompok"`
		Kegiatan   string `json:"kegiatan"`
		Jumlah     int    `json:"jumlah"`
		Tipe       string `json:"tipe"`
	}
	type UstadzRekap struct {
		Total int          `json:"total"`
		Sesi  []SesiDetail `json:"sesi"`
	}
	grouped := map[string]*UstadzRekap{}

	// 1. Absensi Kegiatan (from absensi_sesi)
	rows1, err1 := config.DB.Query(
		`SELECT s.ustadz_username, COALESCE(k.nama,''), COALESCE(k.kegiatan_nama,''), s.tanggal, COALESCE(s.recorded_at,''), COUNT(*) as jumlah
		FROM absensi_sesi s LEFT JOIN kelompok k ON s.kelompok_id = k.id
		WHERE s.tenant_id = ? AND s.tanggal `+dateFilter+`
		GROUP BY s.ustadz_username, s.kelompok_id, s.tanggal, s.recorded_at ORDER BY s.tanggal DESC`,
		append([]interface{}{tid}, dateArgs...)...)
	if err1 != nil {
		fmt.Println("  [RekapUstadz] kegiatan query error:", err1)
	}
	if rows1 != nil {
		defer rows1.Close()
		for rows1.Next() {
			var jml int
			var u, kn, kgn, tgl, ra string
			rows1.Scan(&u, &kn, &kgn, &tgl, &ra, &jml)
			if u == "" {
				continue
			}
			if grouped[u] == nil {
				grouped[u] = &UstadzRekap{}
			}
			grouped[u].Total++
			key := kgn
			if key == "" {
				key = kn
			}
			if key == "" {
				key = "Lainnya"
			}
			grouped[u].Sesi = append(grouped[u].Sesi, SesiDetail{
				Tanggal: tgl, RecordedAt: ra, Kelompok: kn, Kegiatan: key, Jumlah: jml, Tipe: "kegiatan",
			})
		}
	}

	// 2. Absensi Sekolah (from absen_sekolah_sesi — simple query, no subquery)
	rows2, err2 := config.DB.Query(
		`SELECT ass.ustadz_username, COALESCE(ks.nama,''), ass.tanggal, COALESCE(ass.recorded_at,''),
			COALESCE(NULLIF(ass.mata_pelajaran,''), COALESCE(js.mata_pelajaran,''))
		FROM absen_sekolah_sesi ass
		LEFT JOIN kelas_sekolah ks ON ass.kelas_id = ks.id
		LEFT JOIN jadwal_sekolah js ON ass.jadwal_sekolah_id = js.id
		WHERE ass.tenant_id = ? AND ass.tanggal `+dateFilter+` AND ass.ustadz_username != ''
		ORDER BY ass.tanggal DESC`,
		append([]interface{}{tid}, dateArgs...)...)
	if err2 != nil {
		fmt.Println("  [RekapUstadz] sekolah query error:", err2)
	}
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var u, kn, tgl, ra, mp string
			rows2.Scan(&u, &kn, &tgl, &ra, &mp)
			if u == "" {
				continue
			}
			if grouped[u] == nil {
				grouped[u] = &UstadzRekap{}
			}
			grouped[u].Total++
			label := "Sekolah"
			if mp != "" {
				label = "Sekolah — " + mp
			} else if kn != "" {
				label = "Sekolah — " + kn
			}
			grouped[u].Sesi = append(grouped[u].Sesi, SesiDetail{
				Tanggal: tgl, RecordedAt: ra, Kelompok: kn, Kegiatan: label, Jumlah: 0, Tipe: "sekolah",
			})
		}
	}

	// 3. Absensi Diniyyah (from absen_diniyyah — grouped by ustadz, kelas, mapel, tanggal)
	rows3, err3 := config.DB.Query(
		`SELECT ad.ustadz_username, COALESCE(kd.nama,''), COALESCE(ad.mata_pelajaran,''), ad.tanggal, COALESCE(MAX(ad.created_at),''), COUNT(*) as jumlah
		FROM absen_diniyyah ad
		LEFT JOIN kelas_diniyyah kd ON ad.kelas_diniyyah_id = kd.id
		WHERE ad.tenant_id = ? AND ad.tanggal `+dateFilter+` AND ad.ustadz_username != ''
		GROUP BY ad.ustadz_username, ad.kelas_diniyyah_id, ad.mata_pelajaran, ad.tanggal
		ORDER BY ad.tanggal DESC`,
		append([]interface{}{tid}, dateArgs...)...)
	if err3 != nil {
		fmt.Println("  [RekapUstadz] diniyyah query error:", err3)
	}
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var jml int
			var u, kn, mp, tgl, ra string
			rows3.Scan(&u, &kn, &mp, &tgl, &ra, &jml)
			if u == "" {
				continue
			}
			if grouped[u] == nil {
				grouped[u] = &UstadzRekap{}
			}
			grouped[u].Total++
			label := "Diniyyah"
			if mp != "" {
				label = "Diniyyah — " + mp
			} else if kn != "" {
				label = "Diniyyah — " + kn
			}
			grouped[u].Sesi = append(grouped[u].Sesi, SesiDetail{
				Tanggal: tgl, RecordedAt: ra, Kelompok: kn, Kegiatan: label, Jumlah: jml, Tipe: "diniyyah",
			})
		}
	}

	// 4. Tahfidz (from tahfidz_sesi)
	rows4, err4 := config.DB.Query(
		`SELECT ts.ustadz_username, COALESCE(h.nama,''), ts.tanggal, COALESCE(ts.recorded_at,'')
		FROM tahfidz_sesi ts LEFT JOIN halaqoh h ON ts.halaqoh_id = h.id
		WHERE ts.tenant_id = ? AND ts.tanggal `+dateFilter+` AND ts.ustadz_username != ''
		ORDER BY ts.tanggal DESC`,
		append([]interface{}{tid}, dateArgs...)...)
	if err4 != nil {
		fmt.Println("  [RekapUstadz] tahfidz query error:", err4)
	}
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var u, hn, tgl, ra string
			rows4.Scan(&u, &hn, &tgl, &ra)
			if u == "" {
				continue
			}
			if grouped[u] == nil {
				grouped[u] = &UstadzRekap{}
			}
			grouped[u].Total++
			label := "Tahfidz"
			if hn != "" {
				label = "Tahfidz — " + hn
			}
			grouped[u].Sesi = append(grouped[u].Sesi, SesiDetail{
				Tanggal: tgl, RecordedAt: ra, Kelompok: hn, Kegiatan: label, Jumlah: 0, Tipe: "tahfidz",
			})
		}
	}

	return c.JSON(grouped)
}

func RekapUstadz(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uname := c.Locals("username").(string)
	today := helpers.TodayWIB()
	var list []fiber.Map

	// 1. Kegiatan
	rows1, _ := config.DB.Query(
		`SELECT s.id, s.ustadz_username, COALESCE(k.nama,''), s.tanggal, COALESCE(s.recorded_at,'')
		FROM absensi_sesi s LEFT JOIN kelompok k ON s.kelompok_id = k.id
		WHERE s.tenant_id = ? AND s.ustadz_username = ? AND s.tanggal = ?`,
		tid, uname, today)
	if rows1 != nil {
		defer rows1.Close()
		for rows1.Next() {
			var id int
			var u, kn, tgl, ra string
			rows1.Scan(&id, &u, &kn, &tgl, &ra)
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": kn, "tanggal": tgl, "recorded_at": ra, "tipe": "kegiatan"})
		}
	}

	// 2. Sekolah
	rows2, _ := config.DB.Query(
		`SELECT s.id, s.ustadz_username, COALESCE(ks.nama,''), COALESCE(NULLIF(s.mata_pelajaran,''), COALESCE(js.mata_pelajaran,'')), s.tanggal, COALESCE(s.recorded_at,'')
		FROM absen_sekolah_sesi s
		LEFT JOIN kelas_sekolah ks ON s.kelas_id = ks.id
		LEFT JOIN jadwal_sekolah js ON s.jadwal_sekolah_id = js.id
		WHERE s.tenant_id = ? AND s.ustadz_username = ? AND s.tanggal = ?`,
		tid, uname, today)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var id int
			var u, kn, mp, tgl, ra string
			rows2.Scan(&id, &u, &kn, &mp, &tgl, &ra)
			label := "Sekolah — " + kn
			if mp != "" {
				label = "Sekolah — " + mp + " (" + kn + ")"
			}
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": label, "tanggal": tgl, "recorded_at": ra, "tipe": "sekolah"})
		}
	}

	// 3. Diniyyah
	rows3, _ := config.DB.Query(
		`SELECT 0, ad.ustadz_username, COALESCE(kd.nama,''), COALESCE(ad.mata_pelajaran,''), ad.tanggal, COALESCE(MAX(ad.created_at),'')
		FROM absen_diniyyah ad LEFT JOIN kelas_diniyyah kd ON ad.kelas_diniyyah_id = kd.id
		WHERE ad.tenant_id = ? AND ad.ustadz_username = ? AND ad.tanggal = ?
		GROUP BY ad.kelas_diniyyah_id, ad.mata_pelajaran, ad.tanggal`,
		tid, uname, today)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var id int
			var u, kn, mp, tgl, ra string
			rows3.Scan(&id, &u, &kn, &mp, &tgl, &ra)
			label := "Diniyyah — " + kn
			if mp != "" {
				label = "Diniyyah — " + mp + " (" + kn + ")"
			}
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": label, "tanggal": tgl, "recorded_at": ra, "tipe": "diniyyah"})
		}
	}

	// 4. Tahfidz
	rows4, _ := config.DB.Query(
		`SELECT ts.id, ts.ustadz_username, COALESCE(h.nama,''), ts.tanggal, COALESCE(ts.recorded_at,'')
		FROM tahfidz_sesi ts LEFT JOIN halaqoh h ON ts.halaqoh_id = h.id
		WHERE ts.tenant_id = ? AND ts.ustadz_username = ? AND ts.tanggal = ?`,
		tid, uname, today)
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var id int
			var u, hn, tgl, ra string
			rows4.Scan(&id, &u, &hn, &tgl, &ra)
			label := "Tahfidz"
			if hn != "" {
				label = "Tahfidz — " + hn
			}
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": label, "tanggal": tgl, "recorded_at": ra, "tipe": "tahfidz"})
		}
	}

	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

func RekapUstadzAll(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	today := helpers.TodayWIB()
	var list []fiber.Map

	// 1. Kegiatan
	rows1, _ := config.DB.Query(
		`SELECT s.id, s.ustadz_username, COALESCE(k.nama,''), s.tanggal, COALESCE(s.recorded_at,'')
		FROM absensi_sesi s LEFT JOIN kelompok k ON s.kelompok_id = k.id
		WHERE s.tenant_id = ? AND s.tanggal = ? ORDER BY s.ustadz_username`,
		tid, today)
	if rows1 != nil {
		defer rows1.Close()
		for rows1.Next() {
			var id int
			var u, kn, tgl, ra string
			rows1.Scan(&id, &u, &kn, &tgl, &ra)
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": kn, "tanggal": tgl, "recorded_at": ra, "tipe": "kegiatan"})
		}
	}

	// 2. Sekolah
	rows2, _ := config.DB.Query(
		`SELECT s.id, s.ustadz_username, COALESCE(ks.nama,''), COALESCE(NULLIF(s.mata_pelajaran,''), COALESCE(js.mata_pelajaran,'')), s.tanggal, COALESCE(s.recorded_at,'')
		FROM absen_sekolah_sesi s
		LEFT JOIN kelas_sekolah ks ON s.kelas_id = ks.id
		LEFT JOIN jadwal_sekolah js ON s.jadwal_sekolah_id = js.id
		WHERE s.tenant_id = ? AND s.tanggal = ? ORDER BY s.ustadz_username`,
		tid, today)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var id int
			var u, kn, mp, tgl, ra string
			rows2.Scan(&id, &u, &kn, &mp, &tgl, &ra)
			label := "Sekolah — " + kn
			if mp != "" {
				label = "Sekolah — " + mp + " (" + kn + ")"
			}
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": label, "tanggal": tgl, "recorded_at": ra, "tipe": "sekolah"})
		}
	}

	// 3. Diniyyah
	rows3, _ := config.DB.Query(
		`SELECT 0, ad.ustadz_username, COALESCE(kd.nama,''), COALESCE(ad.mata_pelajaran,''), ad.tanggal, COALESCE(MAX(ad.created_at),'')
		FROM absen_diniyyah ad LEFT JOIN kelas_diniyyah kd ON ad.kelas_diniyyah_id = kd.id
		WHERE ad.tenant_id = ? AND ad.tanggal = ? AND ad.ustadz_username != ''
		GROUP BY ad.ustadz_username, ad.kelas_diniyyah_id, ad.mata_pelajaran, ad.tanggal
		ORDER BY ad.ustadz_username`,
		tid, today)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var id int
			var u, kn, mp, tgl, ra string
			rows3.Scan(&id, &u, &kn, &mp, &tgl, &ra)
			label := "Diniyyah — " + kn
			if mp != "" {
				label = "Diniyyah — " + mp + " (" + kn + ")"
			}
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": label, "tanggal": tgl, "recorded_at": ra, "tipe": "diniyyah"})
		}
	}

	// 4. Tahfidz
	rows4, _ := config.DB.Query(
		`SELECT ts.id, ts.ustadz_username, COALESCE(h.nama,''), ts.tanggal, COALESCE(ts.recorded_at,'')
		FROM tahfidz_sesi ts LEFT JOIN halaqoh h ON ts.halaqoh_id = h.id
		WHERE ts.tenant_id = ? AND ts.tanggal = ? AND ts.ustadz_username != ''
		ORDER BY ts.ustadz_username`,
		tid, today)
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var id int
			var u, hn, tgl, ra string
			rows4.Scan(&id, &u, &hn, &tgl, &ra)
			label := "Tahfidz"
			if hn != "" {
				label = "Tahfidz — " + hn
			}
			list = append(list, fiber.Map{"id": id, "ustadz_username": u, "kelompok_nama": label, "tanggal": tgl, "recorded_at": ra, "tipe": "tahfidz"})
		}
	}

	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}
