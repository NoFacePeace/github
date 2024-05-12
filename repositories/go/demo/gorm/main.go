package main

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Uid int `gorm:"uniqueIndex"`
}

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatal(err)
	}
	p := &Product{
		Uid: 1,
	}
	if tx := db.Create(p); tx.Error != nil {
		log.Fatal(tx.Error)
	}
	if tx := db.Delete(p); tx.Error != nil {
		log.Fatal(tx.Error)
	}
	p.ID = 0
	if tx := db.Create(p); tx.Error != nil {
		log.Fatal(tx.Error)
	}
}
