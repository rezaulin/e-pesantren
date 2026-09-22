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

	rows, err := db.Query("SELECT id, tenant_id, nama, no_hp, wali_user_id FROM santri ORDER BY id DESC LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, tenant_id int
		var nama, no_hp string
		var wali_user_id sql.NullInt64
		rows.Scan(&id, &tenant_id, &nama, &no_hp, &wali_user_id)
		fmt.Printf("Santri id: %d, tenant_id: %d, nama: %s, no_hp: %s, wali_user_id: %v\n", id, tenant_id, nama, no_hp, wali_user_id.Int64)
	}
}
