package main

import (
	"fmt"
	"log"
	"pesantren-multi/config"
)

func main() {
	config.ConnectDB()
	q := `SELECT p.id, CAST(p.bulan AS CHAR) as bulan, p.nominal, CAST(p.metode AS CHAR) as metode, CAST(COALESCE(p.keterangan,'') AS CHAR) as keterangan, CAST(p.created_at AS CHAR) as created_at 
          FROM pembayaran p LIMIT 1
          UNION ALL
          SELECT pi.id, CAST(CONCAT('Insidental: ', 'test') AS CHAR) as bulan, pi.nominal_dibayar as nominal, CAST(pi.metode AS CHAR) as metode, CAST('Tagihan Insidental' AS CHAR) as keterangan, CAST(pi.tanggal AS CHAR) as created_at 
          FROM pembayaran_insidental pi LIMIT 1`
	rows, err := config.DB.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("Success")
}
