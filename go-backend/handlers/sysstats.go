//go:build linux

package handlers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

// getCPUSample reads /proc/stat and returns idle and total CPU ticks
func getCPUSample() (idle, total uint64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return 0, 0, fmt.Errorf("invalid format")
			}
			for i, field := range fields[1:] {
				val, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					return 0, 0, err
				}
				total += val
				if i == 3 { // The 4th field is idle
					idle = val
				}
			}
			return
		}
	}
	return 0, 0, fmt.Errorf("cpu info not found")
}

func getCPUUsage() float64 {
	idle1, total1, err := getCPUSample()
	if err != nil {
		return 0.0
	}
	time.Sleep(200 * time.Millisecond) // 200ms delay to calculate diff
	idle2, total2, err := getCPUSample()
	if err != nil {
		return 0.0
	}

	idleTicks := float64(idle2 - idle1)
	totalTicks := float64(total2 - total1)

	if totalTicks > 0 {
		return 100.0 * (totalTicks - idleTicks) / totalTicks
	}
	return 0.0
}

func GetSystemStats(c *fiber.Ctx) error {
	// 1. RAM
	var sys syscall.Sysinfo_t
	err := syscall.Sysinfo(&sys)
	
	ramTotal := uint64(0)
	ramFree := uint64(0)
	if err == nil {
		ramTotal = sys.Totalram * uint64(sys.Unit)
		ramFree = sys.Freeram * uint64(sys.Unit)
	}
	
	// Read /proc/meminfo for more accurate MemAvailable if possible (optional, Sysinfo freeram + bufferram is okay, but /proc/meminfo MemAvailable is better).
	// We'll use a quick parse of MemAvailable if it exists.
	file, errMem := os.Open("/proc/meminfo")
	if errMem == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						ramTotal = val * 1024 // in bytes
					}
				}
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						ramFree = val * 1024 // in bytes
					}
				}
			}
		}
		file.Close()
	}

	ramUsed := ramTotal - ramFree
	ramPercent := float64(0)
	if ramTotal > 0 {
		ramPercent = float64(ramUsed) / float64(ramTotal) * 100.0
	}

	// 2. Disk Space on root "/"
	var stat syscall.Statfs_t
	err = syscall.Statfs("/", &stat)
	diskTotal := uint64(0)
	diskFree := uint64(0)
	if err == nil {
		diskTotal = stat.Blocks * uint64(stat.Bsize)
		diskFree = stat.Bavail * uint64(stat.Bsize)
	}
	diskUsed := diskTotal - diskFree
	diskPercent := float64(0)
	if diskTotal > 0 {
		diskPercent = float64(diskUsed) / float64(diskTotal) * 100.0
	}

	// 3. CPU Usage
	cpuPercent := getCPUUsage()

	return c.JSON(fiber.Map{
		"cpu_usage":    fmt.Sprintf("%.1f", cpuPercent),
		"ram_total":    ramTotal,
		"ram_used":     ramUsed,
		"ram_percent":  fmt.Sprintf("%.1f", ramPercent),
		"disk_total":   diskTotal,
		"disk_used":    diskUsed,
		"disk_percent": fmt.Sprintf("%.1f", diskPercent),
	})
}
