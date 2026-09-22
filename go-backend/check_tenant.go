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

	rows, err := db.Query("SELECT id, status, expired_at FROM tenants")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var status string
		var expired_at sql.NullString
		rows.Scan(&id, &status, &expired_at)
		fmt.Printf("id: %d, status: %s, expired_at: %v\n", id, status, expired_at.String)
	}
}
