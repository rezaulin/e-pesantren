package handlers

import (
	"fmt"
	"pesantren-multi/config"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetSettings(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)

	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE tenant_id = ?", tid).Scan(&count)
	if count == 0 {
		config.DB.Exec("INSERT INTO settings (tenant_id, app_name) VALUES (?, 'Pesantren')", tid)
	}

	// ── Base settings (always exist) ──
	var id, psbKuota int
	var appName, alamatLembaga, kepalaNama, namaKota string
	var rekeningBank, rekeningNomor, rekeningAtasNama, bendaharaTelp string
	var logoBase64, kompSekolah, kompDiniyyah, activeSemester, activeMonth string
	var ttdSekolahNama, ttdSekolahImg, ttdDiniyyahNama, ttdDiniyyahImg, ttdKegiatanNama, ttdKegiatanImg string
	var namaLembagaSekolah, headerSekolahLine1, headerSekolahLine2, headerSekolahAlamat, logoSekolah string
	var namaLembagaDiniyyah, headerDiniyyahAlamat, logoDiniyyah string
	var namaLembagaKegiatan, headerKegiatanAlamat, logoKegiatan string
	var tahunAjaranAktif, semesterAktif, tagihanBulanMulai string
	err := config.DB.QueryRow(`SELECT COALESCE(id,0), COALESCE(app_name,'Pesantren'), COALESCE(alamat_lembaga,''), 
		COALESCE(kepala_nama,''), COALESCE(nama_kota,''), COALESCE(rekening_bank,''), COALESCE(rekening_nomor,''), 
		COALESCE(rekening_atas_nama,''), COALESCE(bendahara_telp,''), COALESCE(psb_kuota,0),
		COALESCE(logo_base64,''), COALESCE(komponen_nilai_sekolah,'[{\"nama\":\"Harian\",\"bobot\":30},{\"nama\":\"UTS\",\"bobot\":30},{\"nama\":\"UAS\",\"bobot\":40}]'), COALESCE(komponen_nilai_diniyyah,'[{\"nama\":\"Harian\",\"bobot\":30},{\"nama\":\"UTS\",\"bobot\":30},{\"nama\":\"UAS\",\"bobot\":40}]'),
		COALESCE(active_semester,''), COALESCE(active_month,''),
		COALESCE(ttd_sekolah_nama,''), COALESCE(ttd_sekolah_img,''),
		COALESCE(ttd_diniyyah_nama,''), COALESCE(ttd_diniyyah_img,''),
		COALESCE(ttd_kegiatan_nama,''), COALESCE(ttd_kegiatan_img,''),
		COALESCE(nama_lembaga_sekolah,''), COALESCE(header_sekolah_line1,'PEMERINTAH DAERAH'), COALESCE(header_sekolah_line2,'DINAS PENDIDIKAN'), COALESCE(header_sekolah_alamat,''), COALESCE(logo_sekolah,''),
		COALESCE(nama_lembaga_diniyyah,''), COALESCE(header_diniyyah_alamat,''), COALESCE(logo_diniyyah,''),
		COALESCE(nama_lembaga_kegiatan,''), COALESCE(header_kegiatan_alamat,''), COALESCE(logo_kegiatan,''),
		COALESCE(tahun_ajaran_aktif,''), COALESCE(semester_aktif,''), COALESCE(tagihan_bulan_mulai,'')
		FROM settings WHERE tenant_id = ? LIMIT 1`, tid).
		Scan(&id, &appName, &alamatLembaga, &kepalaNama, &namaKota, &rekeningBank, &rekeningNomor,
			&rekeningAtasNama, &bendaharaTelp, &psbKuota, &logoBase64, &kompSekolah, &kompDiniyyah, &activeSemester, &activeMonth,
			&ttdSekolahNama, &ttdSekolahImg, &ttdDiniyyahNama, &ttdDiniyyahImg, &ttdKegiatanNama, &ttdKegiatanImg,
			&namaLembagaSekolah, &headerSekolahLine1, &headerSekolahLine2, &headerSekolahAlamat, &logoSekolah,
			&namaLembagaDiniyyah, &headerDiniyyahAlamat, &logoDiniyyah,
			&namaLembagaKegiatan, &headerKegiatanAlamat, &logoKegiatan,
			&tahunAjaranAktif, &semesterAktif, &tagihanBulanMulai)
	if err != nil {
		return c.JSON(fiber.Map{"app_name": "Pesantren"})
	}

	result := fiber.Map{
		"id": id, "app_name": appName, "alamat_lembaga": alamatLembaga,
		"kepala_nama": kepalaNama, "nama_kota": namaKota,
		"rekening_bank": rekeningBank, "rekening_nomor": rekeningNomor,
		"rekening_atas_nama": rekeningAtasNama, "bendahara_telp": bendaharaTelp,
		"psb_kuota": psbKuota,
		"logo_base64": logoBase64,
		"komponen_nilai_sekolah": kompSekolah,
		"komponen_nilai_diniyyah": kompDiniyyah,
		"active_semester": activeSemester,
		"tahun_ajaran_aktif": tahunAjaranAktif,
		"semester_aktif": semesterAktif,
		"active_month": activeMonth,
		"tagihan_bulan_mulai": tagihanBulanMulai,
		"ttd_sekolah_nama": ttdSekolahNama, "ttd_sekolah_img": ttdSekolahImg,
		"ttd_diniyyah_nama": ttdDiniyyahNama, "ttd_diniyyah_img": ttdDiniyyahImg,
		"ttd_kegiatan_nama": ttdKegiatanNama, "ttd_kegiatan_img": ttdKegiatanImg,
		"nama_lembaga_sekolah": namaLembagaSekolah, "header_sekolah_line1": headerSekolahLine1, "header_sekolah_line2": headerSekolahLine2, "header_sekolah_alamat": headerSekolahAlamat, "logo_sekolah": logoSekolah,
		"nama_lembaga_diniyyah": namaLembagaDiniyyah, "header_diniyyah_alamat": headerDiniyyahAlamat, "logo_diniyyah": logoDiniyyah,
		"nama_lembaga_kegiatan": namaLembagaKegiatan, "header_kegiatan_alamat": headerKegiatanAlamat, "logo_kegiatan": logoKegiatan,
	}

	// ── Payment Gateway fields (may not exist on older DB) ──
	var pgProvider, pgServerKey, pgClientKey, pgFeePercentStr, pgFeeFlatStr, pgIsProdStr string
	errPG := config.DB.QueryRow(`SELECT COALESCE(pg_provider,''), COALESCE(pg_server_key,''), COALESCE(pg_client_key,''), 
		COALESCE(pg_is_production,'0'), COALESCE(pg_fee_percent,'0'), COALESCE(pg_fee_flat,'0')
		FROM settings WHERE tenant_id = ? LIMIT 1`, tid).
		Scan(&pgProvider, &pgServerKey, &pgClientKey, &pgIsProdStr, &pgFeePercentStr, &pgFeeFlatStr)
	if errPG == nil {
		pgIsProduction, _ := strconv.Atoi(pgIsProdStr)
		pgFeePercent, _ := strconv.ParseFloat(pgFeePercentStr, 64)
		pgFeeFlat, _ := strconv.ParseInt(pgFeeFlatStr, 10, 64)

		// Mask server key for ALL roles (security)
		maskedServerKey := pgServerKey
		if len(pgServerKey) > 8 {
			maskedServerKey = pgServerKey[:4] + "****" + pgServerKey[len(pgServerKey)-4:]
		} else if pgServerKey != "" {
			maskedServerKey = "****"
		}

		result["pg_provider"] = pgProvider
		result["pg_server_key"] = maskedServerKey
		result["pg_client_key"] = pgClientKey
		result["pg_is_production"] = pgIsProduction
		result["pg_fee_percent"] = pgFeePercent
		result["pg_fee_flat"] = pgFeeFlat
	}

	// ── Transfer Bank fields (may not exist on older DB) ──
	var tbEnabledStr, tbFeeFlatStr, tbExpiryStr, mootaEnabledStr, mootaToken, mootaSecret string
	errTB := config.DB.QueryRow(`SELECT COALESCE(transfer_bank_enabled,'0'), COALESCE(transfer_bank_fee_flat,'0'),
		COALESCE(transfer_bank_expiry_hours,'24'),
		COALESCE(moota_enabled,'0'), COALESCE(moota_api_token,''), COALESCE(moota_secret_token,'')
		FROM settings WHERE tenant_id = ? LIMIT 1`, tid).
		Scan(&tbEnabledStr, &tbFeeFlatStr, &tbExpiryStr, &mootaEnabledStr, &mootaToken, &mootaSecret)
	if errTB == nil {
		tbEnabled, _ := strconv.Atoi(tbEnabledStr)
		tbFeeFlat, _ := strconv.ParseInt(tbFeeFlatStr, 10, 64)
		tbExpiry, _ := strconv.Atoi(tbExpiryStr)
		if tbExpiry <= 0 { tbExpiry = 24 }
		mootaEn, _ := strconv.Atoi(mootaEnabledStr)

		// Mask tokens for ALL roles (security)
		maskedMootaToken := mootaToken
		maskedMootaSecret := mootaSecret
		if len(mootaToken) > 8 {
			maskedMootaToken = mootaToken[:4] + "****" + mootaToken[len(mootaToken)-4:]
		} else if mootaToken != "" {
			maskedMootaToken = "****"
		}
		if len(mootaSecret) > 4 {
			maskedMootaSecret = "****"
		} else if mootaSecret != "" {
			maskedMootaSecret = "****"
		}

		result["transfer_bank_enabled"] = tbEnabled
		result["transfer_bank_fee_flat"] = tbFeeFlat
		result["transfer_bank_expiry_hours"] = tbExpiry
		result["moota_enabled"] = mootaEn
		result["moota_api_token"] = maskedMootaToken
		result["moota_secret_token"] = maskedMootaSecret
	}

	return c.JSON(result)
}

func UpdateSettings(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)

	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE tenant_id = ?", tid).Scan(&count)
	if count == 0 {
		config.DB.Exec("INSERT INTO settings (tenant_id, app_name) VALUES (?, 'Pesantren')", tid)
	}

	var body map[string]interface{}
	c.BodyParser(&body)
	delete(body, "id")
	delete(body, "tenant_id")

	allowed := map[string]bool{
		"app_name": true, "alamat_lembaga": true, "kepala_nama": true, "nama_kota": true,
		"rekening_bank": true, "rekening_nomor": true, "rekening_atas_nama": true, "bendahara_telp": true,
		"psb_kuota": true,
		"pg_provider": true, "pg_server_key": true, "pg_client_key": true,
		"pg_is_production": true, "pg_fee_percent": true, "pg_fee_flat": true,
		"transfer_bank_enabled": true, "transfer_bank_fee_flat": true, "transfer_bank_expiry_hours": true,
		"moota_enabled": true, "moota_api_token": true, "moota_secret_token": true,
		"logo_base64": true, "komponen_nilai_sekolah": true, "komponen_nilai_diniyyah": true,
		"active_semester": true, "active_month": true,
		"tahun_ajaran_aktif": true, "semester_aktif": true,
		"tagihan_bulan_mulai": true,
		"ttd_sekolah_nama": true, "ttd_sekolah_img": true,
		"ttd_diniyyah_nama": true, "ttd_diniyyah_img": true,
		"ttd_kegiatan_nama": true, "ttd_kegiatan_img": true,
		"nama_lembaga_sekolah": true, "header_sekolah_line1": true, "header_sekolah_line2": true, "header_sekolah_alamat": true, "logo_sekolah": true,
		"nama_lembaga_diniyyah": true, "header_diniyyah_alamat": true, "logo_diniyyah": true,
		"nama_lembaga_kegiatan": true, "header_kegiatan_alamat": true, "logo_kegiatan": true,
	}
	for k, v := range body {
		if allowed[k] {
			// Skip updating secrets if they are masked
			if (k == "pg_server_key" || k == "moota_api_token" || k == "moota_secret_token") && v != nil {
				if strVal, ok := v.(string); ok && strings.Contains(strVal, "****") {
					continue
				}
			}
			config.DB.Exec(fmt.Sprintf("UPDATE settings SET %s = ? WHERE tenant_id = ?", k), v, tid)
		}
	}

	return c.JSON(fiber.Map{"message": "Settings disimpan"})
}
