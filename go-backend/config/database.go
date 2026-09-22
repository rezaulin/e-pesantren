package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() error {
	host := envOr("DB_HOST", "localhost")
	user := envOr("DB_USER", "pesantren")
	pass := envOr("DB_PASS", "pesantren2026")
	name := envOr("DB_NAME", "pesantren_multi")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&loc=Local&charset=utf8mb4", user, pass, host, name)
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil { return err }
	DB.SetMaxOpenConns(50)                 // max 50 koneksi (cukup untuk banyak wali concurrent)
	DB.SetMaxIdleConns(10)                 // idle 10 koneksi
	DB.SetConnMaxLifetime(5 * time.Minute) // recycle koneksi tiap 5 menit
	return DB.Ping()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
