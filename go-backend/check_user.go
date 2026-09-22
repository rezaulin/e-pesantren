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

	rows, err := db.Query("SELECT id, tenant_id, username, role, is_active FROM users WHERE username = '081216599910'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, tenant_id int
		var username, role string
		var is_active sql.NullInt64
		rows.Scan(&id, &tenant_id, &username, &role, &is_active)
		fmt.Printf("id: %d, tenant_id: %d, username: %s, role: %s, is_active: %v\n", id, tenant_id, username, role, is_active.Int64)
	}
}
