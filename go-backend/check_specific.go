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

	rows, err := db.Query("SELECT id, tenant_id, nama, no_hp FROM santri WHERE no_hp LIKE '%99910%' OR no_hp LIKE '%165999%'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, tenant_id int
		var nama, no_hp string
		rows.Scan(&id, &tenant_id, &nama, &no_hp)
		fmt.Printf("id: %d, tenant_id: %d, nama: %s, no_hp: %s\n", id, tenant_id, nama, no_hp)
		count++
	}
	fmt.Printf("Total found: %d\n", count)
}
