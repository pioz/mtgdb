package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pioz/mtgdb"
)

func main() {
	log.Println("Importer initialization")
	importer := mtgdb.NewImporter("./data")
	// importer.OnlyTheseSetCodes = []string{"eld"}
	// importer.DownloadAssets = false
	log.Println("Downloading data")
	err := importer.DownloadData()
	if err != nil {
		panic(err)
	}

	log.Println("Open connection to database")
	db, err := gorm.Open("mysql", "root@tcp(127.0.0.1:3306)/mtgdb?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic("Failed to connect database")
	}
	log.Println("Database migration")
	mtgdb.AutoMigrate(db)

	log.Println("Filling database")
	start := time.Now()
	collection := importer.BuildCardsFromJson()
	err = mtgdb.BulkInsert(db, collection)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Imported %d cards in %s\n", len(collection), time.Since(start))
}
