package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// ══════════════════════════════════════════════════════════
// LAYER 2: GO FIBER RATE LIMITER (berlapis)
// Layer 1 = Cloudflare, Layer 3 = Nginx, Layer 4 = Frontend
// ══════════════════════════════════════════════════════════

// getClientIP mendapatkan IP asli dari Cloudflare/Nginx header
func getClientIP(c *fiber.Ctx) string {
	// Prioritas: CF-Connecting-IP > X-Real-IP > X-Forwarded-For > RemoteIP
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return c.IP()
}

// ── Global Rate Limiter ──────────────────────────────────
// Membatasi semua request: 60 req/menit per IP
func GlobalRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"message": "Terlalu banyak request. Coba lagi nanti.",
				"code":    "RATE_LIMIT",
			})
		},
	})
}

// ── Landing Page Limiter ────────────────────────────────
// Lebih ketat untuk halaman landing: 20 req/menit per IP
func LandingRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "landing:" + getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).SendString("Too many requests")
		},
	})
}

// ── Login Brute-Force Limiter ───────────────────────────
// Login: 5 percobaan per menit per IP
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "login:" + getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"message": "Terlalu banyak percobaan login. Tunggu 1 menit.",
				"code":    "LOGIN_RATE_LIMIT",
			})
		},
	})
}

// ══════════════════════════════════════════════════════════
// CLICK BOMB DETECTOR — Deteksi pola klik abnormal
// ══════════════════════════════════════════════════════════

type clickRecord struct {
	count     int
	firstSeen time.Time
	blocked   bool
	blockedAt time.Time
}

var (
	clickTracker = make(map[string]*clickRecord)
	clickMu      sync.Mutex
)

func init() {
	// Bersihkan tracker setiap 10 menit
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			clickMu.Lock()
			now := time.Now()
			for ip, rec := range clickTracker {
				// Hapus record lama (>30 menit)
				if now.Sub(rec.firstSeen) > 30*time.Minute {
					delete(clickTracker, ip)
				}
				// Unblock setelah 15 menit
				if rec.blocked && now.Sub(rec.blockedAt) > 15*time.Minute {
					rec.blocked = false
					rec.count = 0
					rec.firstSeen = now
				}
			}
			clickMu.Unlock()
		}
	}()
}

// ClickBombDetector mendeteksi pola klik abnormal pada halaman landing
// Jika >30 request dalam 30 detik dari 1 IP = langsung block 15 menit
func ClickBombDetector() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := getClientIP(c)

		clickMu.Lock()
		rec, exists := clickTracker[ip]
		if !exists {
			clickTracker[ip] = &clickRecord{
				count:     1,
				firstSeen: time.Now(),
			}
			clickMu.Unlock()
			return c.Next()
		}

		// Sudah di-block?
		if rec.blocked {
			remaining := 15*time.Minute - time.Since(rec.blockedAt)
			clickMu.Unlock()
			fmt.Printf("🚫 [CLICK BOMB] IP %s BLOCKED — sisa %v\n", ip, remaining.Round(time.Second))
			return c.Status(403).JSON(fiber.Map{
				"message": "Akses diblokir sementara karena aktivitas mencurigakan.",
				"code":    "CLICK_BOMB_BLOCKED",
				"retry":   int(remaining.Seconds()),
			})
		}

		rec.count++
		elapsed := time.Since(rec.firstSeen)

		// Jika >30 request dalam 30 detik → BOMB DETECTED
		if rec.count > 30 && elapsed < 30*time.Second {
			rec.blocked = true
			rec.blockedAt = time.Now()
			clickMu.Unlock()
			fmt.Printf("🚨 [CLICK BOMB DETECTED] IP %s — %d req in %v — BLOCKED 15min\n", ip, rec.count, elapsed.Round(time.Millisecond))
			return c.Status(403).JSON(fiber.Map{
				"message": "Aktivitas mencurigakan terdeteksi. Diblokir 15 menit.",
				"code":    "CLICK_BOMB_DETECTED",
			})
		}

		// Reset window jika sudah lewat 30 detik
		if elapsed > 30*time.Second {
			rec.count = 1
			rec.firstSeen = time.Now()
		}

		clickMu.Unlock()
		return c.Next()
	}
}

// ══════════════════════════════════════════════════════════
// SECURITY HEADERS — Anti-clickjacking, XSS, dll
// ══════════════════════════════════════════════════════════

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// P1: Content Security Policy — cegah XSS inject script external
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://app.midtrans.com https://app.sandbox.midtrans.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; "+
				"font-src 'self' https://fonts.gstatic.com https://cdn.jsdelivr.net data:; "+
				"img-src 'self' data: blob: https:; "+
				"connect-src 'self' https://api.midtrans.com https://api.sandbox.midtrans.com; "+
				"frame-src https://app.midtrans.com https://app.sandbox.midtrans.com")
		return c.Next()
	}
}

// ══════════════════════════════════════════════════════════
// REGISTER RATE LIMITER — Anti-spam registrasi tenant
// ══════════════════════════════════════════════════════════

// RegisterRateLimiter membatasi registrasi tenant: 3 req/jam per IP
func RegisterRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "register:" + getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"message": "Terlalu banyak percobaan registrasi. Coba lagi dalam 1 jam.",
				"code":    "REGISTER_RATE_LIMIT",
			})
		},
	})
}
