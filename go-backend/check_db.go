// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "pesantren:jancok123@tcp(localhost:3306)/pesantren_multi?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("--- TENANTS ---")
	rows, _ := db.Query("SELECT id, subdomain, features FROM tenants")
	for rows.Next() {
		var id int
		var sub string
		var feat sql.NullString
		rows.Scan(&id, &sub, &feat)
		fmt.Printf("Tenant ID: %d | Subdomain: %s | Features: %s\n", id, sub, feat.String)
	}
	rows.Close()

	fmt.Println("\n--- USERS ---")
	rows2, _ := db.Query("SELECT id, tenant_id, username, role FROM users")
	for rows2.Next() {
		var id, tid int
		var user, role string
		rows2.Scan(&id, &tid, &user, &role)
		fmt.Printf("User ID: %d | Tenant: %d | Username: %s | Role: %s\n", id, tid, user, role)
	}
	rows2.Close()
}
