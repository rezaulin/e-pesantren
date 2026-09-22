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

	// Add merchant_id to users if it doesn't exist
	_, err = db.Exec("ALTER TABLE users ADD COLUMN merchant_id INT DEFAULT NULL")
	if err != nil {
		fmt.Println("Error (might already exist):", err)
	} else {
		fmt.Println("Berhasil menambahkan kolom merchant_id ke tabel users")
	}
}
