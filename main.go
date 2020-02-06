package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pioz/mtgdb/importer"
)

func main() {
	log.Println("Importer initialization")
	imp := importer.NewImporter("./data")
	// imp.OnlyTheseSetCodes = []string{"eld"}
	// imp.DownloadAssets = false
	log.Println("Downloading data")
	err := imp.DownloadData()
	if err != nil {
		panic(err)
	}

	log.Println("Open connection to database")
	db, err := gorm.Open("mysql", "root@tcp(127.0.0.1:3306)/mtgscan?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic("Failed to connect database")
	}
	log.Println("Database migration")
	db.AutoMigrate(&importer.Card{})
	db.Model(&importer.Card{}).AddUniqueIndex("idx_cards_set_code_collector_number", "set_code", "collector_number", "is_token")

	log.Println("Filling database")
	start := time.Now()
	collection := imp.BuildCardsFromJson()
	err = importer.BulkInsert(db, collection, 1000)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Imported %d cards in %s\n", len(collection), time.Since(start))
}
