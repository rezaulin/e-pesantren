package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"pesantren-multi/config"
	"pesantren-multi/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
)

// RaportAllZip generates a ZIP file containing individual PDF raports for every santri
func RaportAllZip(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id").(int)
	tglMulai := c.Query("tgl_mulai")
	tglAkhir := c.Query("tgl_akhir")
	bulan := c.Query("bulan", helpers.MonthWIB())
	periodeLabel := bulan
	if tglMulai != "" && tglAkhir != "" {
		periodeLabel = tglMulai + " s/d " + tglAkhir
	}

	// Build date filter helper
	buildDF := func(col string) (string, []interface{}) {
		if tglMulai != "" && tglAkhir != "" {
			return " AND " + col + " BETWEEN ? AND ?", []interface{}{tglMulai, tglAkhir}
		}
		return " AND " + col + " LIKE ?", []interface{}{bulan + "%"}
	}

	// Get lembaga name
	var lembaga string
	config.DB.QueryRow("SELECT COALESCE(app_name,'Pesantren') FROM settings WHERE tenant_id = ?", tid).Scan(&lembaga)
	if lembaga == "" {
		lembaga = "Pesantren"
	}

	// Fetch all santri
	rows, err := config.DB.Query("SELECT s.id, s.nama, COALESCE(s.alamat,''), COALESCE(s.kelas_diniyyah,''), COALESCE(k.nama,'-'), COALESCE(s.wali_user_id,0), COALESCE(s.nama_wali,'') FROM santri s LEFT JOIN kamar k ON s.kamar_id = k.id WHERE s.tenant_id = ? ORDER BY s.nama", tid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data santri"})
	}
	defer rows.Close()

	type Santri struct {
		ID       int
		Nama     string
		Alamat   string
		Kelas    string
		Kamar    string
		WaliUID  int
		NamaWali string
	}
	var santriList []Santri
	for rows.Next() {
		var s Santri
		rows.Scan(&s.ID, &s.Nama, &s.Alamat, &s.Kelas, &s.Kamar, &s.WaliUID, &s.NamaWali)
		santriList = append(santriList, s)
	}

	if len(santriList) == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Tidak ada santri"})
	}

	// Create ZIP buffer
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	for _, santri := range santriList {
		sid := santri.ID
		sNama := santri.Nama
		sKelas := santri.Kelas
		kamarNama := santri.Kamar
		waliNama := "-"
		if santri.WaliUID > 0 {
			config.DB.QueryRow("SELECT nama FROM users WHERE id = ?", santri.WaliUID).Scan(&waliNama)
		} else if santri.NamaWali != "" {
			waliNama = santri.NamaWali
		}

		// ── Rekap Kegiatan ──
		type R struct{ H, I, S, A int }
		dfK, dfKArgs := buildDF("a.tanggal")
		kRows, _ := config.DB.Query("SELECT COALESCE(k.kegiatan_nama,'Lainnya'), a.status FROM absensi a LEFT JOIN kelompok k ON a.kelompok_id = k.id WHERE a.tenant_id = ? AND a.santri_id = ?"+dfK, append([]interface{}{tid, sid}, dfKArgs...)...)
		rekap := map[string]*R{}
		if kRows != nil {
			for kRows.Next() {
				var keg, st string
				kRows.Scan(&keg, &st)
				if rekap[keg] == nil { rekap[keg] = &R{} }
				switch st {
				case "H": rekap[keg].H++
				case "I": rekap[keg].I++
				case "S": rekap[keg].S++
				case "A": rekap[keg].A++
				}
			}
			kRows.Close()
		}

		// ── Rekap Sekolah ──
		dfS, dfSArgs := buildDF("tanggal")
		sRows, _ := config.DB.Query("SELECT COALESCE(mata_pelajaran,'Lainnya'), status FROM absen_sekolah WHERE tenant_id = ? AND santri_id = ?"+dfS, append([]interface{}{tid, sid}, dfSArgs...)...)
		rekapSek := map[string]*R{}
		if sRows != nil {
			for sRows.Next() {
				var mp, st string
				sRows.Scan(&mp, &st)
				if rekapSek[mp] == nil { rekapSek[mp] = &R{} }
				switch st {
				case "H": rekapSek[mp].H++
				case "I": rekapSek[mp].I++
				case "S": rekapSek[mp].S++
				case "A": rekapSek[mp].A++
				}
			}
			sRows.Close()
		}

		// ── Rekap Malam ──
		dfM, dfMArgs := buildDF("tanggal")
		mRows, _ := config.DB.Query("SELECT status FROM absen_malam WHERE tenant_id = ? AND santri_id = ?"+dfM, append([]interface{}{tid, sid}, dfMArgs...)...)
		rmH, rmI, rmS, rmA := 0, 0, 0, 0
		if mRows != nil {
			for mRows.Next() {
				var st string
				mRows.Scan(&st)
				switch st {
				case "H": rmH++
				case "I": rmI++
				case "S": rmS++
				case "A": rmA++
				}
			}
			mRows.Close()
		}

		// ── Catatan Guru ──
		dfC, dfCArgs := buildDF("tanggal")
		cRows, _ := config.DB.Query("SELECT COALESCE(catatan,''), COALESCE(tanggal,'') FROM catatan_guru WHERE tenant_id = ? AND santri_id = ?"+dfC, append([]interface{}{tid, sid}, dfCArgs...)...)
		var catatan []string
		if cRows != nil {
			for cRows.Next() {
				var ct, tgl string
				cRows.Scan(&ct, &tgl)
				catatan = append(catatan, fmt.Sprintf("[%s] %s", tgl, ct))
			}
			cRows.Close()
		}

		// ── Pelanggaran ──
		dfP, dfPArgs := buildDF("tanggal")
		pRows, _ := config.DB.Query("SELECT COALESCE(jenis,''), COALESCE(deskripsi,''), COALESCE(poin,0), COALESCE(tanggal,'') FROM pelanggaran WHERE tenant_id = ? AND santri_id = ?"+dfP, append([]interface{}{tid, sid}, dfPArgs...)...)
		var pelanggaran []string
		if pRows != nil {
			for pRows.Next() {
				var j, d, tgl string
				var p int
				pRows.Scan(&j, &d, &p, &tgl)
				pelanggaran = append(pelanggaran, fmt.Sprintf("[%s] %s - %s (%d poin)", tgl, j, d, p))
			}
			pRows.Close()
		}

		// ── Build PDF ──
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Helvetica", "B", 16)
		pdf.CellFormat(0, 8, strings.ToUpper(lembaga), "", 1, "C", false, 0, "")
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(0, 5, "LAPORAN BULANAN PERKEMBANGAN SANTRI", "", 1, "C", false, 0, "")
		pdf.Ln(3)
		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(5)

		// Identity
		pdf.SetFont("Helvetica", "", 10)
		y := pdf.GetY()
		pdf.Text(10, y, "Nama Santri"); pdf.Text(45, y, ": "+sNama)
		pdf.Text(120, y, "Kelas"); pdf.Text(150, y, ": "+sKelas)
		pdf.Text(10, y+5, "Asrama"); pdf.Text(45, y+5, ": "+kamarNama)
		pdf.Text(120, y+5, "Periode"); pdf.Text(150, y+5, ": "+periodeLabel)
		pdf.Text(10, y+10, "Wali"); pdf.Text(45, y+10, ": "+waliNama)
		pdf.SetY(y + 18)

		// A. Rekap Kegiatan
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, "A. Rekap Absensi", "", 1, "", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		w := []float64{60, 22, 22, 22, 22, 22}
		hdr := []string{"Kegiatan", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
		for i, h := range hdr { pdf.CellFormat(w[i], 6, h, "1", 0, "C", false, 0, ""); _ = h }
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 9)
		tH, tI, tS, tA := 0, 0, 0, 0
		for keg, r := range rekap {
			t := r.H + r.I + r.S + r.A
			pdf.CellFormat(w[0], 6, keg, "1", 0, "", false, 0, "")
			pdf.CellFormat(w[1], 6, fmt.Sprintf("%d", r.H), "1", 0, "C", false, 0, "")
			pdf.CellFormat(w[2], 6, fmt.Sprintf("%d", r.I), "1", 0, "C", false, 0, "")
			pdf.CellFormat(w[3], 6, fmt.Sprintf("%d", r.S), "1", 0, "C", false, 0, "")
			pdf.CellFormat(w[4], 6, fmt.Sprintf("%d", r.A), "1", 0, "C", false, 0, "")
			pdf.CellFormat(w[5], 6, fmt.Sprintf("%d", t), "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
			tH += r.H; tI += r.I; tS += r.S; tA += r.A
		}
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(w[0], 6, "TOTAL", "1", 0, "", false, 0, "")
		pdf.CellFormat(w[1], 6, fmt.Sprintf("%d", tH), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[2], 6, fmt.Sprintf("%d", tI), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[3], 6, fmt.Sprintf("%d", tS), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[4], 6, fmt.Sprintf("%d", tA), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[5], 6, fmt.Sprintf("%d", tH+tI+tS+tA), "1", 0, "C", false, 0, "")
		pdf.Ln(5)

		// B. Rekap Sekolah
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, "B. Rekap Absen Sekolah", "", 1, "", false, 0, "")
		if len(rekapSek) > 0 {
			pdf.SetFont("Helvetica", "B", 9)
			for i, h := range hdr {
				label := h
				if i == 0 { label = "Mata Pelajaran" }
				pdf.CellFormat(w[i], 6, label, "1", 0, "C", false, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Helvetica", "", 9)
			for mp, r := range rekapSek {
				t := r.H + r.I + r.S + r.A
				pdf.CellFormat(w[0], 6, mp, "1", 0, "", false, 0, "")
				pdf.CellFormat(w[1], 6, fmt.Sprintf("%d", r.H), "1", 0, "C", false, 0, "")
				pdf.CellFormat(w[2], 6, fmt.Sprintf("%d", r.I), "1", 0, "C", false, 0, "")
				pdf.CellFormat(w[3], 6, fmt.Sprintf("%d", r.S), "1", 0, "C", false, 0, "")
				pdf.CellFormat(w[4], 6, fmt.Sprintf("%d", r.A), "1", 0, "C", false, 0, "")
				pdf.CellFormat(w[5], 6, fmt.Sprintf("%d", t), "1", 0, "C", false, 0, "")
				pdf.Ln(-1)
			}
		} else {
			pdf.SetFont("Helvetica", "", 9)
			pdf.CellFormat(0, 5, "Belum ada data absen sekolah.", "", 1, "", false, 0, "")
		}
		pdf.Ln(3)

		// C. Rekap Malam
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, "C. Rekap Absen Malam", "", 1, "", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		wM := []float64{40, 22, 22, 22, 22, 22}
		hdrM := []string{"Keterangan", "Hadir", "Izin", "Sakit", "Alpa", "Total"}
		for i, h := range hdrM { pdf.CellFormat(wM[i], 6, h, "1", 0, "C", false, 0, ""); _ = h }
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(wM[0], 6, "Malam", "1", 0, "", false, 0, "")
		pdf.CellFormat(wM[1], 6, fmt.Sprintf("%d", rmH), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wM[2], 6, fmt.Sprintf("%d", rmI), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wM[3], 6, fmt.Sprintf("%d", rmS), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wM[4], 6, fmt.Sprintf("%d", rmA), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wM[5], 6, fmt.Sprintf("%d", rmH+rmI+rmS+rmA), "1", 0, "C", false, 0, "")
		pdf.Ln(8)

		// D. Pelanggaran
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, "D. Catatan Kedisiplinan", "", 1, "", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		if len(pelanggaran) > 0 {
			for i, p := range pelanggaran {
				pdf.CellFormat(0, 5, fmt.Sprintf("%d. %s", i+1, p), "", 1, "", false, 0, "")
			}
		} else {
			pdf.CellFormat(0, 5, "Tidak ada catatan pelanggaran.", "", 1, "", false, 0, "")
		}
		pdf.Ln(5)

		// E. Catatan
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, "E. Laporan Perkembangan", "", 1, "", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		if len(catatan) > 0 {
			for i, ct := range catatan {
				pdf.CellFormat(0, 5, fmt.Sprintf("%d. %s", i+1, ct), "", 1, "", false, 0, "")
			}
		} else {
			pdf.CellFormat(0, 5, "Belum ada catatan perkembangan.", "", 1, "", false, 0, "")
		}

		// Write PDF to buffer
		var pdfBuf bytes.Buffer
		pdf.Output(&pdfBuf)

		// Add to ZIP with santri name as filename
		safeName := strings.ReplaceAll(sNama, "/", "_")
		safeName = strings.ReplaceAll(safeName, "\\", "_")
		safeName = strings.ReplaceAll(safeName, ":", "_")
		filename := fmt.Sprintf("Raport_%s.pdf", safeName)

		fw, err := zipWriter.Create(filename)
		if err != nil {
			continue
		}
		fw.Write(pdfBuf.Bytes())
	}

	zipWriter.Close()

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=Raport_Semua_%s.zip", strings.ReplaceAll(periodeLabel, " ", "_")))
	return c.Send(zipBuf.Bytes())
}
