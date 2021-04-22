package main

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:pumpkin123@/ubikom")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	//ctx := context.Background()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS pkeys (Public BINARY(33) NOT NULL PRIMARY KEY, Timestamp BIGINT);")
	if err != nil {
		panic(err)
	}
}
