// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "pesantren:jancok123@tcp(localhost:3306)/pesantren_multi")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Add pemilik and kontak to sangu_merchants if it doesn't exist
	_, err = db.Exec("ALTER TABLE sangu_merchants ADD COLUMN pemilik VARCHAR(100) DEFAULT NULL")
	if err != nil {
		fmt.Println("Error (might already exist):", err)
	} else {
		fmt.Println("Berhasil menambahkan kolom pemilik ke tabel sangu_merchants")
	}

	_, err = db.Exec("ALTER TABLE sangu_merchants ADD COLUMN kontak VARCHAR(50) DEFAULT NULL")
	if err != nil {
		fmt.Println("Error (might already exist):", err)
	} else {
		fmt.Println("Berhasil menambahkan kolom kontak ke tabel sangu_merchants")
	}
}
