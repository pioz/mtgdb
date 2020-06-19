package mtgdb_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
)

const FIXTURES_PATH = "./fixtures"
const TEMP_DIR = "/tmp/mtgdb_test"

// func TestImporterDownloadData(t *testing.T) {
// 	defer os.RemoveAll(TEMP_DIR)

// 	importer := mtgdb.NewImporter(TEMP_DIR)
// 	importer.DownloadData()
// 	_, err := os.Stat(filepath.Join(TEMP_DIR, "all_sets.json"))
// 	assert.False(t, os.IsNotExist(err))
// 	_, err = os.Stat(filepath.Join(TEMP_DIR, "all_cards.json"))
// 	assert.False(t, os.IsNotExist(err))
// }

func TestImporterBuildCardsFromJson(t *testing.T) {
	defer os.RemoveAll(TEMP_DIR)

	importer := mtgdb.NewImporter(filepath.Join(FIXTURES_PATH, "data"))
	importer.DownloadAssets = true
	importer.ImagesDir = filepath.Join(TEMP_DIR, "images")

	collection := importer.BuildCardsFromJson()
	sort.Slice(collection, func(i, j int) bool {
		return collection[i].ScryfallId > collection[j].ScryfallId
	})

	_, err := os.Stat(filepath.Join(importer.ImagesDir, "/sets/eld.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/sets/isd.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/sets/ust.jpg"))
	assert.False(t, os.IsNotExist(err))

	assert.Equal(t, 8, len(collection))

	card := collection[0]
	assert.False(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "Acclaimed Contender", card.EnName)
	assert.Equal(t, "Contendiente aclamada", card.EsName)
	assert.Equal(t, "Concurrente acclamée", card.FrName)
	assert.Equal(t, "Bejubelte Wettstreiterin", card.DeName)
	assert.Equal(t, "Contendente Acclamata", card.ItName)
	assert.Equal(t, "Competidora Aclamada", card.PtName)
	assert.Equal(t, "評判高い挑戦者", card.JaName)
	assert.Equal(t, "칭송받는 경쟁자", card.KoName)
	assert.Equal(t, "Превозносимая Претендентка", card.RuName)
	assert.Equal(t, "受誉竞争者", card.ZhsName)
	assert.Equal(t, "受譽競爭者", card.ZhtName)
	assert.Equal(t, "eld", card.SetCode)
	assert.Equal(t, "eld", card.Set.Code)
	assert.Equal(t, "Throne of Eldraine", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "1", card.CollectorNumber)
	assert.Equal(t, "fb6b12e7-bb93-4eb6-bad1-b256a6ccff4e", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/eld/eld_1.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/isd/eld_1_back.jpg"))
	assert.True(t, os.IsNotExist(err))

	card = collection[1]
	assert.True(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "Garruk, Cursed Huntsman Emblem", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "teld", card.SetCode)
	assert.Equal(t, "teld", card.Set.Code)
	assert.Equal(t, "Throne of Eldraine Tokens", card.Set.Name)
	assert.Equal(t, "2019-09-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "19", card.CollectorNumber)
	assert.Equal(t, "d6c65749-1774-4b36-891e-abf762c95cec", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/teld/teld_19.jpg"))
	assert.False(t, os.IsNotExist(err))

	card = collection[2]
	assert.False(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "\"Rumors of My Death . . .\"", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "ust", card.SetCode)
	assert.Equal(t, "ust", card.Set.Code)
	assert.Equal(t, "Unstable", card.Set.Name)
	assert.Equal(t, "2017-12-08 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "ust", card.Set.IconName)
	assert.Equal(t, "65", card.CollectorNumber)
	assert.Equal(t, "cb3587b9-e727-4f37-b4d6-1baa7316262f", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/ust/ust_65.jpg"))
	assert.False(t, os.IsNotExist(err))

	card = collection[3]
	assert.False(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "Garruk, Cursed Huntsman", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "eld", card.SetCode)
	assert.Equal(t, "eld", card.Set.Code)
	assert.Equal(t, "Throne of Eldraine", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "191", card.CollectorNumber)
	assert.Equal(t, "abef512f-8f1d-4257-b16f-c0eed58670ec", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/eld/eld_191.jpg"))
	assert.False(t, os.IsNotExist(err))

	card = collection[4]
	assert.False(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "Acclaimed Contender", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "peld", card.SetCode)
	assert.Equal(t, "peld", card.Set.Code)
	assert.Equal(t, "Throne of Eldraine Promos", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "1s", card.CollectorNumber)
	assert.Equal(t, "9a675b33-ab47-4a34-ab10-384e0de2f71f", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/peld/peld_1s.jpg"))
	assert.False(t, os.IsNotExist(err))

	card = collection[5]
	assert.False(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "Acclaimed Contender", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "peld", card.SetCode)
	assert.Equal(t, "peld", card.Set.Code)
	assert.Equal(t, "Throne of Eldraine Promos", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "1p", card.CollectorNumber)
	assert.Equal(t, "77ba25cb-a8a6-46b6-82be-5c70e663dfdf", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/peld/peld_1p.jpg"))
	assert.False(t, os.IsNotExist(err))

	card = collection[6]
	assert.False(t, card.IsToken)
	assert.True(t, card.IsDoubleFaced)
	assert.Equal(t, "Daybreak Ranger // Nightfall Predator", card.EnName)
	assert.Equal(t, "Guardabosque del amanecer // Depredadora del anochecer", card.EsName)
	assert.Equal(t, "Ranger de l'aube // Prédateur du crépuscule", card.FrName)
	assert.Equal(t, "Morgengrauen-Waldläufer // Nachtbeginn-Jäger", card.DeName)
	assert.Equal(t, "Ranger dell'Alba // Predatrice del Crepuscolo", card.ItName)
	assert.Equal(t, "Patrulheiro do Amanhecer // Predador do Anoitecer", card.PtName)
	assert.Equal(t, "夜明けのレインジャー // 黄昏の捕食者", card.JaName)
	assert.Equal(t, "여명의 레인저 // 해질녘의 포식자", card.KoName)
	assert.Equal(t, "Рассветная Обходчица // Сумеречная Хищница", card.RuName)
	assert.Equal(t, "破晓护林人 // 夜幕掠食者", card.ZhsName)
	assert.Equal(t, "破曉護林人 // 夜幕掠食者", card.ZhtName)
	assert.Equal(t, "isd", card.SetCode)
	assert.Equal(t, "isd", card.Set.Code)
	assert.Equal(t, "Innistrad", card.Set.Name)
	assert.Equal(t, "2011-09-30 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "isd", card.Set.IconName)
	assert.Equal(t, "176", card.CollectorNumber)
	assert.Equal(t, "25b54a1d-e201-453b-9173-b04e06ee6fb7", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/isd/isd_176.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/isd/isd_176_back.jpg"))
	assert.False(t, os.IsNotExist(err))

	card = collection[7]
	assert.False(t, card.IsToken)
	assert.False(t, card.IsDoubleFaced)
	assert.Equal(t, "Acclaimed Contender", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "eld", card.SetCode)
	assert.Equal(t, "eld", card.Set.Code)
	assert.Equal(t, "Throne of Eldraine", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "334", card.CollectorNumber)
	assert.Equal(t, "0dbf3260-b956-40da-abc7-764781c9f26f", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/eld/eld_334.jpg"))
	assert.False(t, os.IsNotExist(err))
}

func TestBulkInsert(t *testing.T) {
	dbConnection := os.Getenv("DB_CONNECTION")
	if dbConnection == "" {
		dbConnection = "root@tcp(127.0.0.1:3306)/mtgdb_test?charset=utf8mb4&parseTime=True"
	}
	db, err := gorm.Open("mysql", dbConnection)
	if err != nil {
		panic(err)
	}
	mtgdb.AutoMigrate(db)

	cards := []mtgdb.Card{
		mtgdb.Card{
			EnName:          "Gilded Goose",
			CollectorNumber: "160",
			SetCode:         "eld",
			Set: &mtgdb.Set{
				Name:     "Throne of Eldraine",
				Code:     "eld",
				IconName: "eld",
			},
		},
		mtgdb.Card{
			EnName:          "Acclaimed Contender",
			CollectorNumber: "1",
			SetCode:         "eld",
			Set: &mtgdb.Set{
				Name:     "Throne of Eldraine",
				Code:     "eld",
				IconName: "eld",
			},
		},
		mtgdb.Card{
			EnName:          "Daybreak Ranger // Nightfall Predator",
			CollectorNumber: "176",
			SetCode:         "isd",
			Set: &mtgdb.Set{
				Name:     "Innistrad",
				Code:     "isd",
				IconName: "isd",
			},
		},
	}

	mtgdb.BulkInsert(db, cards)

	db.Preload("Set").Order("en_name").Find(&cards)
	assert.Equal(t, 3, len(cards))

	assert.Equal(t, "Acclaimed Contender", cards[0].EnName)
	assert.Equal(t, "1", cards[0].CollectorNumber)
	assert.Equal(t, "eld", cards[0].SetCode)
	assert.Equal(t, "Throne of Eldraine", cards[0].Set.Name)
	assert.Equal(t, "eld", cards[0].Set.Code)

	assert.Equal(t, "Daybreak Ranger // Nightfall Predator", cards[1].EnName)
	assert.Equal(t, "176", cards[1].CollectorNumber)
	assert.Equal(t, "isd", cards[1].SetCode)
	assert.Equal(t, "Innistrad", cards[1].Set.Name)
	assert.Equal(t, "isd", cards[1].Set.Code)

	assert.Equal(t, "Gilded Goose", cards[2].EnName)
	assert.Equal(t, "160", cards[2].CollectorNumber)
	assert.Equal(t, "eld", cards[2].SetCode)
	assert.Equal(t, "Throne of Eldraine", cards[2].Set.Name)
	assert.Equal(t, "eld", cards[2].Set.Code)
}

func TestDownloadFile(t *testing.T) {
	os.MkdirAll(TEMP_DIR, os.ModePerm)
	file := filepath.Join(TEMP_DIR, "teferi.png")
	defer os.RemoveAll(TEMP_DIR)
	url := "https://img.scryfall.com/cards/normal/front/5/d/5d10b752-d9cb-419d-a5c4-d4ee1acb655e.jpg?1562736365"

	mtgdb.DownloadFile(file, url, nil)
	_, err := os.Stat(file)
	assert.False(t, os.IsNotExist(err))

	olderTime, _ := time.Parse(time.RFC3339, "1990-01-01T00:00:00.00Z")
	err = os.Chtimes(file, olderTime, olderTime)
	stat, _ := os.Stat(file)
	mtgdb.DownloadFile(file, url, stat)
	stat, _ = os.Stat(file)
	assert.False(t, olderTime.Equal(stat.ModTime()))

	newerTime, _ := time.Parse(time.RFC3339, "2020-06-01T00:00:00.00Z")
	err = os.Chtimes(file, newerTime, newerTime)
	stat, _ = os.Stat(file)
	mtgdb.DownloadFile(file, url, stat)
	stat, _ = os.Stat(file)
	assert.True(t, newerTime.Equal(stat.ModTime()))
}
