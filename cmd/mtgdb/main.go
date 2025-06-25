package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pioz/mtgdb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func createCustomItems(db *gorm.DB) error {
	scope := db.Clauses(clause.OnConflict{UpdateAll: true}).Session(&gorm.Session{CreateBatchSize: 1000})

	releasedAt, err := time.Parse(time.RFC3339, "1993-08-05T00:00:00.000Z")
	if err != nil {
		return err
	}
	set := mtgdb.Set{Name: "MTG Print Extra", Code: "extra", ParentCode: "extra", ReleasedAt: &releasedAt, Typology: "extra", IconName: "default"}
	err = scope.Create(&set).Error
	if err != nil {
		return err
	}

	cards := []mtgdb.Card{
		{EnName: "Back", SetCode: "extra", CollectorNumber: "001", Foil: false, NonFoil: true, HasBackSide: false, ReleasedAt: &releasedAt},
		{EnName: "Back", SetCode: "extra", CollectorNumber: "002", Foil: false, NonFoil: true, HasBackSide: false, ReleasedAt: &releasedAt},
		{EnName: "Back", SetCode: "extra", CollectorNumber: "003", Foil: false, NonFoil: true, HasBackSide: false, ReleasedAt: &releasedAt},
		{EnName: "Back", SetCode: "extra", CollectorNumber: "004", Foil: false, NonFoil: true, HasBackSide: false, ReleasedAt: &releasedAt},
		{EnName: "Back", SetCode: "extra", CollectorNumber: "005", Foil: false, NonFoil: true, HasBackSide: false, ReleasedAt: &releasedAt},
	}
	return scope.Omit("Set").Create(&cards).Error
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func main() {
	var forceDownloadData, skipDownloadAssets, forceDownloadOlderAssets, forceDownloadDiffSha1, forceDownloadAssets, downloadOnlyEnAssets, displayProgressBar, help bool
	var downloadConcurrency int
	var setsString string
	flag.BoolVar(&forceDownloadData, "u", false, "Update Scryfall database")
	flag.BoolVar(&skipDownloadAssets, "skip-assets", false, "Skip download of set and card images")
	flag.BoolVar(&forceDownloadOlderAssets, "ftime", false, "Force re-download of card images, but only if the modified date is older")
	// flag.BoolVar(&forceDownloadDiffSha1, "fsha1", false, "Force re-download of card images, but only if the sha1sum is changed")
	flag.BoolVar(&forceDownloadAssets, "f", false, "Force re-download of card images")
	flag.BoolVar(&downloadOnlyEnAssets, "en", true, "Download card images only in EN language")
	flag.IntVar(&downloadConcurrency, "download-concurrency", 0, "Set max download concurrency")
	flag.StringVar(&setsString, "only", "", "Import some sets (es: -only eld,war)")
	flag.BoolVar(&displayProgressBar, "p", false, "Display progress bar")
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
	importer.ForceDownloadDiffSha1 = forceDownloadDiffSha1
	importer.ForceDownloadAssets = forceDownloadAssets
	importer.DownloadOnlyEnAssets = downloadOnlyEnAssets
	importer.DisplayProgressBar = displayProgressBar
	if setsString != "" && len(sets) > 0 {
		importer.OnlyTheseSetCodes = sets
	}
	if downloadConcurrency > 0 {
		importer.SetDownloadConcurrency(downloadConcurrency)
	}

	// Start

	log.Println("Downloading data")
	err := importer.DownloadData()
	if err != nil {
		panic(err)
	}

	log.Println("Open connection to database")
	db, err := gorm.Open(mysql.Open(os.Getenv("DB_CONNECTION")), nil)
	if err != nil {
		panic("Failed to connect database")
	}
	db.Config.Logger = db.Config.Logger.LogMode(logger.Error)
	if os.Getenv("DB_LOG") == "1" {
		db.Config.Logger = db.Config.Logger.LogMode(logger.Info)
	}
	log.Println("Database migration")
	mtgdb.AutoMigrate(db)

	err = createCustomItems(db)
	if err != nil {
		panic(err)
	}

	log.Println("Filling database")
	var beforeSetsCount, beforeCardsCount, afterSetsCount, afterCardsCount int64
	var scryfallIds []string
	db.Model(&mtgdb.Set{}).Count(&beforeSetsCount)
	db.Model(&mtgdb.Card{}).Pluck("scryfall_id", &scryfallIds)
	beforeCardsCount = int64(len(scryfallIds))
	start := time.Now()
	collection, downloadedImagesCount := importer.BuildCardsFromJson()
	err = mtgdb.BulkInsert(db, collection)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Processed %d cards in %s\n", len(collection), time.Since(start))
	db.Model(&mtgdb.Set{}).Count(&afterSetsCount)
	db.Model(&mtgdb.Card{}).Count(&afterCardsCount)
	log.Printf("Imported %d new sets and %d new cards (%d images updated)\n", afterSetsCount-beforeSetsCount, afterCardsCount-beforeCardsCount, downloadedImagesCount)

	// Remove deleted cards ONLY if no filter on sets
	if setsString == "" {
		collectionScryfallIds := make(map[string]struct{})
		for _, card := range collection {
			collectionScryfallIds[card.ScryfallID] = struct{}{}
		}
		scryfallIdsNotFound := make([]string, 0)
		for _, scryfallId := range scryfallIds {
			if _, found := collectionScryfallIds[scryfallId]; !found && scryfallId != "" {
				scryfallIdsNotFound = append(scryfallIdsNotFound, scryfallId)
			}
		}
		db.Where("scryfall_id IN (?)", scryfallIdsNotFound).Delete(mtgdb.Card{})
		log.Printf("Deleted %d cards\n", len(scryfallIdsNotFound))
	}
}
