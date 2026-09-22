package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
)

// GetAbsenDiniyyah — list absensi for a kelas diniyyah on a date
func GetAbsenDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	q := `SELECT ad.id, ad.santri_id, s.nama, ad.kelas_diniyyah_id, ad.tanggal, ad.status,
		COALESCE(ad.jadwal_diniyyah_id,0), COALESCE(ad.mata_pelajaran,''), COALESCE(ad.ustadz_username,'')
		FROM absen_diniyyah ad JOIN santri s ON ad.santri_id = s.id WHERE ad.tenant_id = ?`
	args := []interface{}{tid}
	if kid := c.Query("kelas_diniyyah_id"); kid != "" {
		q += " AND ad.kelas_diniyyah_id = ?"
		args = append(args, kid)
	}
	if tgl := c.Query("tanggal"); tgl != "" {
		q += " AND ad.tanggal = ?"
		args = append(args, tgl)
	}
	if jid := c.Query("jadwal_diniyyah_id"); jid != "" {
		q += " AND ad.jadwal_diniyyah_id = ?"
		args = append(args, jid)
	}
	q += " ORDER BY s.nama"
	rows, err := config.DB.Query(q, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	defer rows.Close()
	var list []fiber.Map
	for rows.Next() {
		var id, sid, kdid, jdid int
		var nama, tgl, status, mapel, ustadz string
		rows.Scan(&id, &sid, &nama, &kdid, &tgl, &status, &jdid, &mapel, &ustadz)
		list = append(list, fiber.Map{"id": id, "santri_id": sid, "santri_nama": nama, "kelas_diniyyah_id": kdid, "tanggal": tgl, "status": status, "jadwal_diniyyah_id": jdid, "mata_pelajaran": mapel, "ustadz_username": ustadz})
	}
	if list == nil {
		list = []fiber.Map{}
	}
	return c.JSON(list)
}

// BulkAbsenDiniyyah — bulk insert absensi diniyyah with schedule enforcement
func BulkAbsenDiniyyah(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	uid := c.Locals("user_id").(int)
	uname := c.Locals("username").(string)
	role := c.Locals("role").(string)
	var body struct {
		Tanggal         string `json:"tanggal"`
		KelasDiniyyahID int    `json:"kelas_diniyyah_id"`
		JadwalDiniyyahID int   `json:"jadwal_diniyyah_id"`
		MataPelajaran   string `json:"mata_pelajaran"`
		Items           []struct {
			SantriID int    `json:"santri_id"`
			Status   string `json:"status"`
		} `json:"items"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request"})
	}
	if body.Tanggal == "" || body.KelasDiniyyahID == 0 || len(body.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "tanggal, kelas_diniyyah_id, dan items wajib"})
	}

	// Schedule enforcement: if jadwal_diniyyah exists for this kelas, only scheduled ustadz + admin can attend
	if role == "ustadz" && body.KelasDiniyyahID > 0 {
		var kelasJadwalCount int
		config.DB.QueryRow("SELECT COUNT(*) FROM jadwal_diniyyah WHERE tenant_id = ? AND kelas_diniyyah_id = ?", tid, body.KelasDiniyyahID).Scan(&kelasJadwalCount)
		if kelasJadwalCount > 0 {
			now := helpers.JamSekarang()
			hari := helpers.HariIni()
			var count int
			config.DB.QueryRow(
				"SELECT COUNT(*) FROM jadwal_diniyyah WHERE tenant_id = ? AND kelas_diniyyah_id = ? AND ustadz_username = ? AND hari = ? AND ? >= jam_mulai",
				tid, body.KelasDiniyyahID, uname, hari, now,
			).Scan(&count)
			if count == 0 {
				return c.Status(403).JSON(fiber.Map{
					"message": fmt.Sprintf("Tidak ada jadwal diniyyah untuk kelas ini hari %s atau belum waktunya (sekarang %s WIB)", hari, now),
					"code":    "NOT_IN_SCHEDULE",
				})
			}
		}
	}

	// Resolve mata_pelajaran from jadwal if provided
	if body.JadwalDiniyyahID > 0 && body.MataPelajaran == "" {
		config.DB.QueryRow("SELECT mata_pelajaran FROM jadwal_diniyyah WHERE id = ? AND tenant_id = ?",
			body.JadwalDiniyyahID, tid).Scan(&body.MataPelajaran)
	}

	count := 0
	// Auto-izin: override status for santri with active perizinan
	izinMap := GetSantriIzinAktif(tid)
	for _, item := range body.Items {
		if item.SantriID == 0 {
			continue
		}
		st := item.Status
		if st == "" {
			st = "H"
		}
		// Override if santri has active perizinan
		if _, ok := izinMap[item.SantriID]; ok {
			st = "I"
		}
		_, err := config.DB.Exec(`INSERT INTO absen_diniyyah (tenant_id, santri_id, kelas_diniyyah_id, jadwal_diniyyah_id, mata_pelajaran, tanggal, status, recorded_by, ustadz_username) 
			VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE status = VALUES(status), ustadz_username = VALUES(ustadz_username)`,
			tid, item.SantriID, body.KelasDiniyyahID, body.JadwalDiniyyahID, body.MataPelajaran, body.Tanggal, st, uid, uname)
		if err == nil {
			count++
		}
	}
	return c.JSON(fiber.Map{"message": "Absen diniyyah disimpan", "count": count})
}
