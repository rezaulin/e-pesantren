package main
import (
	"fmt"
	"log"
	"pesantren-multi/config"
)
func main() {
	config.ConnectDB()
	rows, err := config.DB.Query("SELECT tenant_features FROM users WHERE role='admin' LIMIT 1")
	if err != nil { log.Fatal(err) }
	defer rows.Close()
	for rows.Next() {
		var f string
		rows.Scan(&f)
		fmt.Printf("FEATURES: %s\n", f)
	}
}
