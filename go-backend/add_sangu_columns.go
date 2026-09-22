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

	_, err = db.Exec("ALTER TABLE santri ADD COLUMN sangu_saldo BIGINT DEFAULT 0")
	if err != nil {
		fmt.Println("sangu_saldo:", err)
	} else {
		fmt.Println("sangu_saldo added")
	}

	_, err = db.Exec("ALTER TABLE santri ADD COLUMN sangu_limit_harian BIGINT DEFAULT 0")
	if err != nil {
		fmt.Println("sangu_limit_harian:", err)
	} else {
		fmt.Println("sangu_limit_harian added")
	}
}
