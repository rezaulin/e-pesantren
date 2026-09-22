package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var WIB *time.Location

func init() {
	WIB, _ = time.LoadLocation("Asia/Jakarta")
	if WIB == nil {
		WIB = time.FixedZone("WIB", 7*3600)
	}
}

var Hari = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

func NowWIB() time.Time      { return time.Now().In(WIB) }
func TodayWIB() string       { return NowWIB().Format("2006-01-02") }
func MonthWIB() string       { return NowWIB().Format("2006-01") }
func HariIni() string        { return Hari[NowWIB().Weekday()] }
func JamSekarang() string    { return NowWIB().Format("15:04") }
func ToDatetime() string     { return NowWIB().Format("2006-01-02 15:04:05") }

func AddMinutes(jam string, menit int) string {
	parts := strings.Split(jam, ":")
	if len(parts) != 2 { return jam }
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	total := h*60 + m + menit
	return fmt.Sprintf("%02d:%02d", (total/60)%24, total%60)
}

// NextMonthStart returns "YYYY-MM-01" for the month after the given "YYYY-MM" string.
// Used to convert LIKE 'YYYY-MM%' to efficient range queries: tanggal >= start AND tanggal < nextStart
func NextMonthStart(yearMonth string) string {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		// Fallback: just add -32 days worth
		return yearMonth + "-32"
	}
	next := t.AddDate(0, 1, 0)
	return next.Format("2006-01-02")
}
