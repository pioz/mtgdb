package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/pioz/mtgdb"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func main() {
	var forceDownloadData, skipDownloadAssets, forceDownloadOlderAssets, downloadOnlyEnAssets, help bool
	var downloadConcurrency int
	var setsString string
	flag.BoolVar(&forceDownloadData, "f", false, "Update Scryfall database")
	flag.BoolVar(&skipDownloadAssets, "skip-assets", false, "Skip download of set and card images")
	flag.BoolVar(&forceDownloadOlderAssets, "ff", false, "Force re-download of card images, but only if the modified date is older")
	flag.BoolVar(&downloadOnlyEnAssets, "en", true, "Download card images only in EN language")
	flag.IntVar(&downloadConcurrency, "download-concurrency", 0, "Set max download concurrency")
	flag.StringVar(&setsString, "only", "", "Import some sets (es: -only eld,war)")
	flag.BoolVar(&help, "h", false, "Print this help")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
	sets := strings.Split(setsString, ",")
	// Append Token and Overside sets
	for _, set := range sets {
		sets = append(sets, fmt.Sprintf("t%s", set))
		sets = append(sets, fmt.Sprintf("o%s", set))
	}

	log.Println("Importer initialization")
	importer := mtgdb.NewImporter(os.Getenv("DATA_PATH"))
	importer.ForceDownloadData = forceDownloadData
	importer.DownloadAssets = !skipDownloadAssets
	importer.ForceDownloadOlderAssets = forceDownloadOlderAssets
	importer.DownloadOnlyEnAssets = downloadOnlyEnAssets
	if setsString != "" && len(sets) > 0 {
		importer.OnlyTheseSetCodes = sets
	}
	if downloadConcurrency > 0 {
		importer.SetDownloadConcurrency(downloadConcurrency)
	}

	log.Println("Downloading data")
	err := importer.DownloadData()
	if err != nil {
		panic(err)
	}

	log.Println("Open connection to database")
	db, err := gorm.Open("mysql", os.Getenv("DB_CONNECTION"))
	if err != nil {
		panic("Failed to connect database")
	}
	// db.LogMode(true)
	log.Println("Database migration")
	mtgdb.AutoMigrate(db)

	log.Println("Filling database")
	var beforeSetsCount, beforeCardsCount, afterSetsCount, afterCardsCount int
	db.Model(&mtgdb.Set{}).Count(&beforeSetsCount)
	db.Model(&mtgdb.Card{}).Count(&beforeCardsCount)
	start := time.Now()
	collection := importer.BuildCardsFromJson()
	err = mtgdb.BulkInsert(db, collection)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Processed %d cards in %s\n", len(collection), time.Since(start))
	db.Model(&mtgdb.Set{}).Count(&afterSetsCount)
	db.Model(&mtgdb.Card{}).Count(&afterCardsCount)
	log.Printf("Imported %d new sets and %d new cards\n", afterSetsCount-beforeSetsCount, afterCardsCount-beforeCardsCount)
}
