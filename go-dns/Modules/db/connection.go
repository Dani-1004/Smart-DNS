package db

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	var err error
	DB, err = gorm.Open(sqlite.Open("dns_records.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
	err = DB.AutoMigrate(&DNSRecord{})
	if err != nil {
		panic("failed to migrate database" + err.Error())
	}
	fmt.Println("Database connection established")

}
