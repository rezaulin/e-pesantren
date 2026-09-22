//go:build ignore

package main

import (
	"fmt"
	"log"
	"pesantren-multi/config"
)

func main() {
	err := config.ConnectDB()
	if err != nil {
		log.Fatal("Gagal konek DB:", err)
	}
	defer config.DB.Close()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS tagihan_insidental (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			nama VARCHAR(255) NOT NULL,
			nominal BIGINT NOT NULL,
			aktif INT DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tagihan_insidental_santri (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			tagihan_id INT NOT NULL,
			nominal BIGINT NOT NULL,
			sisa_tagihan BIGINT NOT NULL,
			status VARCHAR(50) DEFAULT 'Belum Lunas',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pembayaran_insidental (
			id INT AUTO_INCREMENT PRIMARY KEY,
			tenant_id INT NOT NULL,
			santri_id INT NOT NULL,
			tagihan_santri_id INT NOT NULL,
			nominal_dibayar BIGINT NOT NULL,
			metode VARCHAR(50),
			keterangan TEXT,
			tanggal DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by INT NOT NULL
		)`,
	}

	for _, q := range queries {
		_, err := config.DB.Exec(q)
		if err != nil {
			fmt.Println("Error executing:", err)
		} else {
			fmt.Println("Success executing query")
		}
	}
	fmt.Println("Migrasi tabel insidental selesai.")
}
