//go:build !linux

package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func GetSystemStats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"cpu_usage":    "N/A",
		"ram_total":    0,
		"ram_used":     0,
		"ram_percent":  "N/A",
		"disk_total":   0,
		"disk_used":    0,
		"disk_percent": "N/A",
	})
}
