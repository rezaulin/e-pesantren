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

	rows, err := db.Query("SELECT id, username, role FROM users WHERE tenant_id = 2")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, role string
		rows.Scan(&id, &username, &role)
		fmt.Printf("id: %d, username: %s, role: %s\n", id, username, role)
	}
}
