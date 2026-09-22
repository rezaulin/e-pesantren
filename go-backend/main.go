package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pesantren-multi/config"
	"pesantren-multi/handlers"
	"pesantren-multi/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	if err := config.ConnectDB(); err != nil {
		log.Fatal("Database connection failed:", err)
	}
	config.AutoMigrate()
	config.StartArchiveScheduler() // arsip otomatis + bersihkan log
	fmt.Println("✅ Database connected")

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
		// P2: Custom error handler — sembunyikan stack trace dari user
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("[ERROR] %s %s → %d: %v", c.Method(), c.Path(), code, err)
			return c.Status(code).JSON(fiber.Map{
				"message": "Terjadi kesalahan internal. Silakan coba lagi.",
			})
		},
	})
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// P1: CORS — hanya izinkan domain sendiri, bukan wildcard (*)
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "https://e-pesantren.app, https://*.e-pesantren.app"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Authorization",
		AllowCredentials: true,
	}))

	// ── Keamanan berlapis: anti click-bomb ──────────────
	app.Use(middleware.SecurityHeaders())   // Layer: security headers
	app.Use(middleware.GlobalRateLimiter()) // Layer: 60 req/menit per IP

	// ── Landing page ────────────────────────────────────
	publicDir := "../public"
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		publicDir = "public"
	}
	absPublic, _ := filepath.Abs(publicDir)
	fmt.Println("📁 Public dir:", absPublic)

	app.Get("/", middleware.ClickBombDetector(), middleware.LandingRateLimiter(), func(c *fiber.Ctx) error {
		// If Authorization header or token cookie present, redirect to /app
		if c.Get("Authorization") != "" {
			return c.Redirect("/app")
		}
		return c.SendFile(filepath.Join(absPublic, "landing.html"))
	})
	app.Get("/app", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		idx, err := os.ReadFile(filepath.Join(absPublic, "index.html"))
		if err != nil {
			return err
		}
		html := strings.Replace(string(idx), "?v=v21.0", "?v=v21.35", -1)
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})
	app.Get("/app/*", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		idx, err := os.ReadFile(filepath.Join(absPublic, "index.html"))
		if err != nil {
			return err
		}
		html := strings.Replace(string(idx), "?v=v21.0", "?v=v21.35", -1)
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})
	app.Static("/", absPublic, fiber.Static{
		Index:         "",
		CacheDuration: 7 * 24 * time.Hour,
	})

	// ── Auth (public) ───────────────────────────────────
	app.Post("/api/login", middleware.LoginRateLimiter(), handlers.Login)

	// ── SaaS Registration (public) ──────────────────────
	// P2: Rate limit register-tenant — 3 req/jam per IP untuk cegah spam
	app.Post("/api/register-tenant", middleware.RegisterRateLimiter(), handlers.RegisterTenantSaaS)
	app.Post("/api/saas/midtrans-callback", handlers.SaaSWebhook)

	// ── PSB (public, no auth) ───────────────────────────
	app.Get("/psb/:subdomain", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(absPublic, "psb.html"))
	})
	app.Get("/api/psb/:subdomain/info", handlers.GetPSBInfo)
	app.Post("/api/psb/:subdomain", handlers.SubmitPSB)

	// ── Payment Gateway Webhook (public, no auth) ──────
	app.Post("/api/payment/notification", handlers.PaymentNotification)
	app.Post("/api/moota/webhook", handlers.MootaWebhook)

	// ── Auth required ───────────────────────────────────
	auth := app.Group("/api", middleware.Authenticate)
	auth.Get("/me", handlers.Me)
	auth.Post("/change-password", handlers.UserChangePassword)
	auth.Get("/dashboard", handlers.Dashboard)

	// ── Settings ────────────────────────────────────────
	auth.Get("/settings", handlers.GetSettings)
	auth.Put("/settings", middleware.RequireAdmin, handlers.UpdateSettings)

	// ── SaaS Admin (Superadmin Only) ────────────────────
	auth.Get("/saas/settings", handlers.GetSaaSSettings)
	auth.Put("/saas/settings", handlers.UpdateSaaSSettings)
	auth.Put("/saas/tenants/:id/fitur", handlers.UpdateTenantSaaS)

	// ── Santri ──────────────────────────────────────────
	auth.Get("/santri", handlers.GetSantri)
	auth.Post("/santri", middleware.RequireAdmin, handlers.CreateSantri)
	auth.Put("/santri/bulk-kategori", middleware.RequireAdminOrBendahara, handlers.BulkSetKategoriSpp)
	auth.Put("/santri/:id", middleware.RequireAdmin, handlers.UpdateSantri)
	auth.Delete("/santri/:id", middleware.RequireAdmin, handlers.DeleteSantri)

	auth.Post("/santri/import-excel", middleware.RequireAdmin, handlers.ImportSantriExcel)
	auth.Get("/santri/:id/profile", handlers.GetSantriProfile)

	// ── PSB Admin ────────────────────────────────────────
	auth.Get("/pendaftar", middleware.RequireAdmin, handlers.GetPendaftar)
	auth.Put("/pendaftar/:id/terima", middleware.RequireAdmin, handlers.TerimaPendaftar)
	auth.Put("/pendaftar/:id/tolak", middleware.RequireAdmin, handlers.TolakPendaftar)

	// ── Kamar ───────────────────────────────────────────
	auth.Get("/kamar", handlers.GetKamar)
	auth.Post("/kamar", middleware.RequireAdmin, handlers.CreateKamar)
	auth.Put("/kamar/:id", middleware.RequireAdmin, handlers.UpdateKamar)
	auth.Delete("/kamar/:id", middleware.RequireAdmin, handlers.DeleteKamar)
	auth.Get("/kamar/:id/members", handlers.GetKamarMembers)
	auth.Post("/kamar/:id/members/bulk", middleware.RequireAdmin, handlers.BulkSetKamarMembers)

	// ── Kegiatan ────────────────────────────────────────
	auth.Get("/kegiatan", handlers.GetKegiatan)
	auth.Post("/kegiatan", middleware.RequireAdmin, handlers.CreateKegiatan)
	auth.Put("/kegiatan/:id", middleware.RequireAdmin, handlers.UpdateKegiatan)
	auth.Delete("/kegiatan/:id", middleware.RequireAdmin, handlers.DeleteKegiatan)

	// ── Kelompok ────────────────────────────────────────
	auth.Get("/kelompok", handlers.GetKelompok)
	auth.Post("/kelompok", middleware.RequireAdmin, handlers.CreateKelompok)
	auth.Delete("/kelompok/:id", middleware.RequireAdmin, handlers.DeleteKelompok)
	auth.Get("/kelompok/:id/members", handlers.GetKelompokMembers)
	auth.Post("/kelompok/:id/members", middleware.RequireAdmin, handlers.AddKelompokMember)
	auth.Post("/kelompok/:id/members/bulk", middleware.RequireAdmin, handlers.BulkAddKelompokMember)
	auth.Delete("/kelompok/:id/members/:santri_id", middleware.RequireAdmin, handlers.RemoveKelompokMember)

	// ── Absensi ─────────────────────────────────────────
	auth.Post("/absensi/bulk", handlers.AbsensiBulk)
	auth.Get("/rekap", handlers.GetRekap)
	auth.Get("/rekap/summary", handlers.GetRekapSummary)
	auth.Get("/rekap/export-excel", handlers.ExportRekapExcel)
	auth.Post("/absen-malam/bulk", handlers.AbsenMalamBulk)
	auth.Get("/absen-malam", handlers.GetAbsenMalam)
	auth.Get("/rekap-absen-malam", handlers.GetRekapAbsenMalam)
	auth.Get("/rekap-absen-malam/export-excel", handlers.ExportRekapMalamExcel)
	auth.Post("/absen-sekolah/bulk", handlers.AbsenSekolahBulk)
	auth.Get("/absen-sekolah", handlers.GetAbsenSekolah)
	auth.Get("/rekap-absen-sekolah", handlers.GetRekapAbsenSekolah)
	auth.Get("/rekap-absen-diniyyah", handlers.GetRekapAbsenDiniyyah)
	auth.Get("/rekap-absen-sekolah/export-excel", handlers.ExportRekapSekolahExcel)
	auth.Get("/rekap-absen-diniyyah/export-excel", handlers.ExportRekapDiniyyahExcel)
	auth.Get("/server-time", handlers.GetServerTime)

	// ── Perizinan ────────────────────────────────────────
	auth.Get("/perizinan", handlers.GetPerizinan)
	auth.Get("/perizinan/aktif", handlers.GetPerizinanAktif)
	auth.Post("/perizinan", middleware.RequireAdminOrKeamanan, handlers.CreatePerizinan)
	auth.Put("/perizinan/:id/kembali", middleware.RequireAdminOrKeamanan, handlers.PerizinanKembali)
	auth.Put("/perizinan/:id/terlambat", middleware.RequireAdminOrKeamanan, handlers.PerizinanTerlambat)
	auth.Delete("/perizinan/:id", middleware.RequireAdminOrKeamanan, handlers.DeletePerizinan)

	// ── Pelanggaran ─────────────────────────────────────
	auth.Get("/pelanggaran/rekap", handlers.GetRekapTakzir)
	auth.Get("/pelanggaran", handlers.GetPelanggaran)
	auth.Post("/pelanggaran", handlers.CreatePelanggaran)
	auth.Put("/pelanggaran/:id/takzir", handlers.UpdateTakzir)
	auth.Delete("/pelanggaran/:id", handlers.DeletePelanggaran)

	// ── Catatan Guru ────────────────────────────────────
	auth.Get("/catatan-guru", handlers.GetCatatanGuru)
	auth.Post("/catatan-guru", handlers.CreateCatatanGuru)
	auth.Delete("/catatan-guru/:id", handlers.DeleteCatatanGuru)

	// ── Jadwal ──────────────────────────────────────────
	auth.Get("/jadwal-umum", handlers.GetJadwalUmum)
	auth.Post("/jadwal-umum", middleware.RequireAdmin, handlers.CreateJadwalUmum)
	auth.Delete("/jadwal-umum/:id", middleware.RequireAdmin, handlers.DeleteJadwalUmum)
	auth.Get("/jadwal-sekolah", handlers.GetJadwalSekolah)
	auth.Post("/jadwal-sekolah", middleware.RequireAdmin, handlers.CreateJadwalSekolah)
	auth.Delete("/jadwal-sekolah/:id", middleware.RequireAdmin, handlers.DeleteJadwalSekolah)
	auth.Get("/jadwal-aktif", handlers.GetJadwalAktif)
	auth.Get("/jadwal-sekolah-aktif", handlers.GetJadwalSekolahAktif)

	// ── Pengumuman ──────────────────────────────────────
	auth.Get("/pengumuman", handlers.GetPengumuman)
	auth.Post("/pengumuman", middleware.RequireAdmin, handlers.CreatePengumuman)
	auth.Delete("/pengumuman/:id", middleware.RequireAdmin, handlers.DeletePengumuman)

	// ── Kelas Sekolah ───────────────────────────────────
	auth.Get("/kelas-sekolah", handlers.GetKelasSekolah)
	auth.Post("/kelas-sekolah", middleware.RequireAdmin, handlers.CreateKelasSekolah)
	auth.Put("/kelas-sekolah/:id", middleware.RequireAdmin, handlers.UpdateKelasSekolah)
	auth.Delete("/kelas-sekolah/:id", middleware.RequireAdmin, handlers.DeleteKelasSekolah)
	auth.Get("/kelas-sekolah/:id/members", handlers.GetKelasMembers)
	auth.Post("/kelas-sekolah/:id/members", middleware.RequireAdmin, handlers.AddKelasMember)
	auth.Post("/kelas-sekolah/:id/members/bulk", middleware.RequireAdmin, handlers.BulkAddKelasMember)
	auth.Delete("/kelas-sekolah/:id/members/:santri_id", middleware.RequireAdmin, handlers.RemoveKelasMember)

	// ── Users ───────────────────────────────────────────
	auth.Get("/users", middleware.RequireAdmin, handlers.GetUsers)
	auth.Post("/users", middleware.RequireAdmin, handlers.CreateUser)
	auth.Put("/users/bulk-toggle", middleware.RequireAdmin, handlers.BulkToggleUsers)
	auth.Put("/users/:id", middleware.RequireAdmin, handlers.UpdateUser)
	auth.Delete("/users/:id", middleware.RequireAdmin, handlers.DeleteUser)

	// ── Raport ──────────────────────────────────────────
	auth.Get("/raport-all/zip", handlers.RaportAllZip)
	auth.Get("/raport/:santri_id", handlers.GetRaport)
	auth.Get("/raport/:santri_id/pdf", handlers.RaportPDF)
	auth.Get("/raport/:santri_id/excel", handlers.RaportExcel)
	auth.Get("/raport-terpadu/data", handlers.GetRaportTerpaduData)
	auth.Get("/raport-terpadu/options", handlers.GetRaportTerpaduOptions)

	// ── Rekap Ustadz ────────────────────────────────────
	auth.Get("/rekap-ustadz/summary", middleware.RequireAdmin, handlers.RekapUstadzSummary)
	auth.Get("/rekap-ustadz", handlers.RekapUstadz)
	auth.Get("/rekap-ustadz/all", middleware.RequireAdmin, handlers.RekapUstadzAll)

	// ── Pembayaran ──────────────────────────────────────
	auth.Get("/pembayaran/tarif", handlers.GetTarif)
	auth.Post("/pembayaran/tarif", middleware.RequireAdmin, handlers.CreateTarif)
	auth.Put("/pembayaran/tarif/:id", middleware.RequireAdmin, handlers.UpdateTarif)
	auth.Delete("/pembayaran/tarif/:id", middleware.RequireAdmin, handlers.DeleteTarif)
	auth.Get("/pembayaran/tarif/:id/periode", handlers.GetTarifPeriode)
	auth.Post("/pembayaran/tarif/:id/periode", middleware.RequireAdmin, handlers.CreateTarifPeriode)
	auth.Delete("/pembayaran/tarif/periode/:id", middleware.RequireAdmin, handlers.DeleteTarifPeriode)
	auth.Get("/pembayaran/kategori", handlers.GetKategoriPembayaran)
	auth.Post("/pembayaran/kategori", middleware.RequireAdminOrBendahara, handlers.CreateKategoriPembayaran)
	auth.Put("/pembayaran/kategori/:id", middleware.RequireAdminOrBendahara, handlers.UpdateKategoriPembayaran)
	auth.Delete("/pembayaran/kategori/:id", middleware.RequireAdminOrBendahara, handlers.DeleteKategoriPembayaran)
	auth.Get("/pembayaran/kategori/:id/periode", handlers.GetKategoriPeriode)
	auth.Post("/pembayaran/kategori/:id/periode", middleware.RequireAdminOrBendahara, handlers.CreateKategoriPeriode)
	auth.Delete("/pembayaran/kategori/periode/:id", middleware.RequireAdminOrBendahara, handlers.DeleteKategoriPeriode)
	auth.Get("/pembayaran/potongan", handlers.GetPotongan)
	auth.Post("/pembayaran/potongan", middleware.RequireAdminOrBendahara, handlers.CreatePotongan)
	auth.Delete("/pembayaran/potongan/:id", middleware.RequireAdminOrBendahara, handlers.DeletePotongan)
	auth.Get("/pembayaran/rekap", handlers.GetRekapPembayaran)
	auth.Get("/pembayaran/riwayat", handlers.GetRiwayatPembayaran)
	auth.Get("/pembayaran/export-excel", handlers.ExportPembayaranExcel)
	auth.Get("/pembayaran", handlers.GetPembayaran)
	auth.Post("/pembayaran", middleware.RequireAdminOrBendahara, handlers.CreatePembayaran)
	auth.Post("/pembayaran/bulk", middleware.RequireAdminOrBendahara, handlers.BulkBayarSPP)
	auth.Delete("/pembayaran/:id", middleware.RequireAdminOrBendahara, handlers.DeletePembayaran)

	// -- Pembayaran Insidental --
	auth.Get("/insidental/jenis", handlers.GetJenisInsidental)
	auth.Post("/insidental/jenis", middleware.RequireAdminOrBendahara, handlers.CreateJenisInsidental)
	auth.Put("/insidental/jenis/:id", middleware.RequireAdminOrBendahara, handlers.UpdateJenisInsidental)
	auth.Get("/insidental/rekap", middleware.RequireAdminOrBendahara, handlers.GetRekapInsidental)
	auth.Get("/insidental/santri/:santri_id", handlers.GetTagihanInsidentalSantri)
	auth.Get("/insidental/santri/:santri_id/riwayat", handlers.GetRiwayatInsidentalSantri)
	auth.Post("/insidental/santri", middleware.RequireAdminOrBendahara, handlers.SetTagihanInsidentalSantri)
	auth.Post("/insidental/bulk", middleware.RequireAdminOrBendahara, handlers.SetTagihanInsidentalBulk)
	auth.Post("/insidental/bayar", middleware.RequireAdminOrBendahara, handlers.BayarInsidental)

	// -- Catatan Keuangan (Bendahara) --
	auth.Get("/keuangan/saldo", handlers.GetSaldoKeuangan)
	auth.Get("/keuangan/export-excel", handlers.ExportKeuanganExcel)
	auth.Get("/keuangan", handlers.GetCatatanKeuangan)
	auth.Post("/keuangan", middleware.RequireAdminOrBendahara, handlers.CreateCatatanKeuangan)
	auth.Delete("/keuangan/:id", middleware.RequireAdminOrBendahara, handlers.DeleteCatatanKeuangan)

	// ── Kelas Diniyyah (Madrasah Diniyyah) ───────────────
	auth.Get("/kelas-diniyyah", handlers.GetKelasDiniyyah)
	auth.Post("/kelas-diniyyah", middleware.RequireAdmin, handlers.CreateKelasDiniyyah)
	auth.Put("/kelas-diniyyah/:id", middleware.RequireAdmin, handlers.UpdateKelasDiniyyah)
	auth.Delete("/kelas-diniyyah/:id", middleware.RequireAdmin, handlers.DeleteKelasDiniyyah)
	auth.Get("/kelas-diniyyah/:id/members", handlers.GetKelasDiniyyahMembers)
	auth.Post("/kelas-diniyyah/:id/members", middleware.RequireAdmin, handlers.AddKelasDiniyyahMember)
	auth.Post("/kelas-diniyyah/:id/members/bulk", middleware.RequireAdmin, handlers.BulkAddKelasDiniyyahMembers)
	auth.Delete("/kelas-diniyyah/:id/members/:santri_id", middleware.RequireAdmin, handlers.RemoveKelasDiniyyahMember)

	// ── Mata Pelajaran Diniyyah ─────────────────────────
	auth.Get("/mata-pelajaran", handlers.GetMataPelajaran)
	auth.Post("/mata-pelajaran", middleware.RequireAdmin, handlers.CreateMataPelajaran)
	auth.Put("/mata-pelajaran/:id", middleware.RequireAdmin, handlers.UpdateMataPelajaran)
	auth.Delete("/mata-pelajaran/:id", middleware.RequireAdmin, handlers.DeleteMataPelajaran)

	// ── Mata Pelajaran Sekolah ───────────────────────────
	auth.Get("/mata-pelajaran-sekolah", handlers.GetMataPelajaranSekolah)
	auth.Post("/mata-pelajaran-sekolah", middleware.RequireAdmin, handlers.CreateMataPelajaranSekolah)
	auth.Put("/mata-pelajaran-sekolah/:id", middleware.RequireAdmin, handlers.UpdateMataPelajaranSekolah)
	auth.Delete("/mata-pelajaran-sekolah/:id", middleware.RequireAdmin, handlers.DeleteMataPelajaranSekolah)

	// ── Nilai Pelajaran Diniyyah & Kegiatan ──────────────
	auth.Get("/nilai-pelajaran", handlers.GetNilaiPelajaran)
	auth.Post("/nilai-pelajaran", middleware.RequireAdminOrUstadz, handlers.UpsertNilaiPelajaran)
	auth.Post("/nilai-pelajaran/bulk", middleware.RequireAdminOrUstadz, handlers.BulkUpsertNilaiPelajaran)
	auth.Get("/nilai-kegiatan", handlers.GetNilaiKegiatan)
	auth.Post("/nilai-kegiatan", middleware.RequireAdminOrUstadz, handlers.UpsertNilaiKegiatan)
	auth.Post("/nilai-kegiatan/bulk", middleware.RequireAdminOrUstadz, handlers.BulkUpsertNilaiKegiatan)

	// ── Nilai Sekolah ───────────────────────────────────
	auth.Get("/nilai-sekolah", handlers.GetNilaiSekolah)
	auth.Post("/nilai-sekolah", middleware.RequireAdminOrUstadz, handlers.UpsertNilaiSekolah)
	auth.Post("/nilai-sekolah/bulk", middleware.RequireAdminOrUstadz, handlers.BulkUpsertNilaiSekolah)

	// ── Peringkat ───────────────────────────────────────
	auth.Get("/peringkat/diniyyah/:kelas_diniyyah_id", handlers.GetPeringkatDiniyyah)
	auth.Get("/peringkat/sekolah/:kelas_id", handlers.GetPeringkatSekolah)
	auth.Get("/peringkat/kelompok/:kelompok_id", handlers.GetPeringkatKelompok)

	// ── Absensi Diniyyah ────────────────────────────────
	auth.Get("/absen-diniyyah", handlers.GetAbsenDiniyyah)
	auth.Post("/absen-diniyyah/bulk", middleware.RequireAdminOrUstadz, handlers.BulkAbsenDiniyyah)

	// ── Jadwal Diniyyah ─────────────────────────────────
	auth.Get("/jadwal-diniyyah", handlers.GetJadwalDiniyyah)
	auth.Post("/jadwal-diniyyah", middleware.RequireAdmin, handlers.CreateJadwalDiniyyah)
	auth.Delete("/jadwal-diniyyah/:id", middleware.RequireAdmin, handlers.DeleteJadwalDiniyyah)

	// ── Raport Penilaian Export ──────────────────────────
	auth.Get("/raport-penilaian/:santri_id/pdf", handlers.RaportPenilaianPDF)
	auth.Get("/raport-penilaian/:santri_id/excel", handlers.RaportPenilaianExcel)
	auth.Get("/raport-penilaian-all/zip", handlers.RaportPenilaianAllZip)

	// ── Wali Santri ─────────────────────────────────────
	auth.Get("/wali/anak", handlers.WaliGetAnak)
	auth.Get("/wali/raport/:santri_id", handlers.WaliGetRaport)
	auth.Get("/wali/raport-lengkap/:santri_id", handlers.WaliGetRaportLengkap)
	auth.Get("/wali/pembayaran/:santri_id", handlers.WaliGetPembayaran)
	auth.Get("/wali/tunggakan/:santri_id", handlers.WaliGetTunggakan)
	auth.Post("/wali/pay", middleware.RequireWali, handlers.WaliCreatePayment)
	auth.Post("/wali/pay-insidental", middleware.RequireWali, handlers.WaliCreatePaymentInsidental)
	auth.Post("/wali/verify-payment", middleware.RequireWali, handlers.WaliVerifyPayment)
	auth.Get("/wali/payment-status/:order_id", middleware.RequireWali, handlers.WaliPaymentStatus)
	auth.Get("/wali/tahfidz/:santri_id", handlers.WaliGetTahfidz)
	auth.Post("/wali/bank-transfer", middleware.RequireWali, handlers.WaliCreateBankTransfer)
	auth.Get("/wali/bank-transfers", middleware.RequireWali, handlers.WaliGetBankTransfers)

	// ── Payment History (admin/bendahara) ───────────────
	auth.Get("/payment/history", middleware.RequireAdminOrBendahara, handlers.GetPaymentHistory)
	auth.Get("/pembayaran/rekap-tunggakan", middleware.RequireAdminOrBendahara, handlers.GetRekapTunggakan)

	// ── Bank Transfer Admin ────────────────────────────
	auth.Get("/bank-transfers", middleware.RequireAdminOrBendahara, handlers.GetBankTransfers)
	auth.Get("/bank-transfers/count-pending", middleware.RequireAdminOrBendahara, handlers.CountPendingBankTransfers)
	auth.Post("/bank-transfers/:id/confirm", middleware.RequireAdminOrBendahara, handlers.ConfirmBankTransferHTTP)
	auth.Post("/bank-transfers/:id/reject", middleware.RequireAdminOrBendahara, handlers.RejectBankTransfer)

	// ── Tahfidz Qur'an ──────────────────────────────────
	auth.Get("/halaqoh", handlers.GetHalaqoh)
	auth.Post("/halaqoh", middleware.RequireAdmin, handlers.CreateHalaqoh)
	auth.Put("/halaqoh/:id", middleware.RequireAdmin, handlers.UpdateHalaqoh)
	auth.Delete("/halaqoh/:id", middleware.RequireAdmin, handlers.DeleteHalaqoh)
	auth.Get("/halaqoh/:id/members", handlers.GetHalaqohMembers)
	auth.Post("/halaqoh/:id/members/bulk", middleware.RequireAdmin, handlers.BulkAddHalaqohMembers)
	auth.Delete("/halaqoh/:id/members/:santri_id", middleware.RequireAdmin, handlers.RemoveHalaqohMember)
	auth.Post("/tahfidz/bulk", middleware.RequireAdminOrUstadz, handlers.TahfidzBulk)
	auth.Get("/tahfidz", handlers.GetTahfidz)
	auth.Get("/tahfidz/rekap", handlers.GetRekapTahfidz)
	auth.Get("/tahfidz/progress/:santri_id", handlers.GetProgressTahfidz)
	auth.Get("/tahfidz/export-excel", handlers.ExportTahfidzExcel)

	// ── E-Paket (Logistik Paket Santri) ─────────────────
	auth.Get("/paket/stats", middleware.RequireAdminUstadzKeamanan, handlers.GetPaketStats)
	auth.Get("/paket", middleware.RequireAdminUstadzKeamanan, handlers.GetPaket)
	auth.Post("/paket", middleware.RequireAdminUstadzKeamanan, handlers.CreatePaket)
	auth.Put("/paket/bulk-transfer", middleware.RequireAdminUstadzKeamanan, handlers.BulkTransferPaket)
	auth.Put("/paket/:id/serahkan", middleware.RequireAdminUstadzKeamanan, handlers.SerahkanPaket)
	auth.Put("/paket/:id/rak", middleware.RequireAdminUstadzKeamanan, handlers.UpdatePaketRak)
	auth.Delete("/paket/:id", middleware.RequireAdminUstadzKeamanan, handlers.DeletePaket)

	// ── Sangu Santri (Uang Saku & Kantin) ───────────────
	// Admin
	auth.Get("/sangu/merchants", middleware.RequireAdmin, handlers.GetSanguMerchants)
	auth.Post("/sangu/merchants", middleware.RequireAdmin, handlers.CreateSanguMerchant)
	auth.Put("/sangu/merchants/:id", middleware.RequireAdmin, handlers.UpdateSanguMerchant)
	auth.Get("/sangu/withdrawals", middleware.RequireAdmin, handlers.GetSanguWithdrawals)
	auth.Put("/sangu/withdrawals/:id/process", middleware.RequireAdmin, handlers.ProcessSanguWithdrawal)

	// Merchant
	auth.Get("/sangu/merchant/dashboard", handlers.GetMerchantDashboard)
	auth.Get("/sangu/merchant/products", handlers.GetMerchantProducts)
	auth.Post("/sangu/merchant/products", handlers.CreateMerchantProduct)
	auth.Put("/sangu/merchant/products/:id", handlers.UpdateMerchantProduct)
	auth.Delete("/sangu/merchant/products/:id", handlers.DeleteMerchantProduct)
	auth.Post("/sangu/merchant/withdrawals", handlers.RequestWithdrawal)
	auth.Get("/sangu/merchant/withdrawals", handlers.GetMerchantWithdrawals)

	// Kasir
	auth.Get("/sangu/kasir/santri", handlers.KasirGetSantriByCard)
	auth.Post("/sangu/kasir/checkout", handlers.KasirCheckout)

	// RFID Card Registration
	auth.Put("/sangu/register-card", middleware.RequireAdmin, handlers.RegisterCard)
	auth.Delete("/sangu/register-card/:santri_id", middleware.RequireAdmin, handlers.UnregisterCard)

	// Wali
	auth.Get("/sangu/wali/:santri_id", middleware.RequireWali, handlers.WaliGetSanguInfo)
	auth.Put("/sangu/wali/:santri_id/limit", middleware.RequireWali, handlers.WaliSetLimitHarian)
	auth.Get("/sangu/wali/:santri_id/riwayat", middleware.RequireWali, handlers.WaliGetRiwayatSangu)
	auth.Post("/sangu/wali/topup", middleware.RequireWali, handlers.WaliCreateTopupSangu)

	// ── Super Admin ─────────────────────────────────────
	sa := app.Group("/api/super", middleware.Authenticate, middleware.RequireSuperAdmin)
	sa.Get("/tenants", handlers.GetTenants)
	sa.Post("/tenants", handlers.CreateTenant)
	sa.Put("/tenants/:id", handlers.UpdateTenant)
	sa.Delete("/tenants/:id", handlers.DeleteTenant)
	sa.Get("/stats", handlers.SuperStats)
	sa.Put("/change-password", handlers.ChangePassword)
	sa.Put("/tenants/:id/extend", handlers.ExtendTenant)
	sa.Put("/tenants/:id/toggle", handlers.ToggleTenant)
	sa.Post("/broadcast", handlers.Broadcast)
	sa.Get("/server-stats", handlers.GetSystemStats)

	// ── Background: Auto-expire pending bank transfers ──
	handlers.StartBankTransferExpiry()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	fmt.Printf("🚀 Server running on port %s (WIB timezone)\n", port)
	log.Fatal(app.Listen(":" + port))
}
