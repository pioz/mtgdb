package mtgdb_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const FIXTURES_PATH = "./testdata"
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
	importer.DownloadOnlyEnAssets = false
	importer.ImagesDir = filepath.Join(TEMP_DIR, "images")

	collection, downloadedImagesCount := importer.BuildCardsFromJson()
	sort.Slice(collection, func(i, j int) bool {
		return collection[i].ScryfallId > collection[j].ScryfallId
	})

	_, err := os.Stat(filepath.Join(importer.ImagesDir, "/sets/eld.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/sets/isd.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/sets/ust.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/sets/war.jpg"))
	assert.False(t, os.IsNotExist(err))

	assert.Equal(t, 9, len(collection))
	assert.Equal(t, uint32(42), downloadedImagesCount)

	card := collection[0]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "eld", card.Set.ParentCode)
	assert.Equal(t, "Throne of Eldraine", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "expansion", card.Set.Typology)
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "1", card.CollectorNumber)
	assert.Equal(t, "fb6b12e7-bb93-4eb6-bad1-b256a6ccff4e", card.ScryfallId)
	for _, lang := range []string{"en", "de", "es", "fr", "it", "ja", "ko", "pt", "ru", "zhs", "zht"} {
		_, err = os.Stat(filepath.Join(importer.ImagesDir, fmt.Sprintf("/cards/eld/eld_1_%s.jpg", lang)))
		assert.False(t, os.IsNotExist(err))
		_, err = os.Stat(filepath.Join(importer.ImagesDir, fmt.Sprintf("/cards/eld/eld_1_%s_back.jpg", lang)))
		assert.True(t, os.IsNotExist(err))
	}
	assert.Equal(t, "35df179a-c0e6-4ac1-a861-e6e9b4d1614d", card.OracleId)
	assert.Equal(t, uint64(0), card.MtgoID)
	assert.Equal(t, uint64(0), card.ArenaID)
	assert.Equal(t, uint64(0), card.TcgplayerID)
	assert.Equal(t, uint64(0), card.CardmarketID)
	assert.Equal(t, "normal", card.Layout)
	assert.Equal(t, "{2}{W}", card.ManaCost)
	assert.Equal(t, float32(3), card.CMC)
	assert.Equal(t, "Creature — Human Knight", card.TypeLine)
	assert.Equal(t, "When Acclaimed Contender enters the battlefield, if you control another Knight, look at the top five cards of your library. You may reveal a Knight, Aura, Equipment, or legendary artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.", card.OracleText)
	assert.Equal(t, "3", card.Power)
	assert.Equal(t, "3", card.Toughness)
	assert.Equal(t, mtgdb.SliceString{"W"}, card.Colors)
	assert.Equal(t, mtgdb.SliceString{"W"}, card.ColorIdentity)
	assert.Equal(t, mtgdb.SliceString(nil), card.Keywords)
	assert.Equal(t, mtgdb.SliceString(nil), card.ProducedMana)
	assert.Equal(t, mtgdb.MapString{"brawl": "legal", "commander": "legal", "duel": "legal", "future": "legal", "historic": "legal", "legacy": "legal", "modern": "legal", "oldschool": "not_legal", "pauper": "not_legal", "penny": "legal", "pioneer": "legal", "standard": "legal", "vintage": "legal"}, card.Legalities)
	assert.Equal(t, mtgdb.SliceString{"arena", "mtgo", "paper"}, card.Games)
	assert.Equal(t, false, card.Oversized)
	assert.Equal(t, false, card.Promo)
	assert.Equal(t, false, card.Reprint)
	assert.Equal(t, false, card.Variation)
	assert.Equal(t, false, card.Digital)
	assert.Equal(t, "rare", card.Rarity)
	assert.Equal(t, "", card.Watermark)
	assert.Equal(t, "David Gaillet", card.Artist)
	assert.Equal(t, "black", card.BorderColor)
	assert.Equal(t, "2015", card.Frame)
	assert.Equal(t, mtgdb.SliceString(nil), card.FrameEffects)
	assert.Equal(t, "", card.SecurityStamp)
	assert.Equal(t, false, card.FullArt)
	assert.Equal(t, false, card.Textless)
	assert.Equal(t, true, card.Booster)
	assert.Equal(t, false, card.StorySpotlight)
	assert.Equal(t, 2, len(card.Rulings))
	assert.Equal(t, "Acclaimed Contender’s ability won’t trigger if you don’t control another Knight immediately after it enters the battlefield. If you don’t control another Knight as that ability resolves, the ability has no effect. This doesn’t have the be the same Knight at both times, however.", card.Rulings[0].Comment)
	assert.Equal(t, "2019-10-04 00:00:00", card.Rulings[0].PublishedAt.Format("2006-01-02 15:04:05"))
	assert.Equal(t, "Acclaimed Contender’s ability can get you at most one card from the top five cards, no matter how many other Knights you control.", card.Rulings[1].Comment)
	assert.Equal(t, "2019-10-04 00:00:00", card.Rulings[1].PublishedAt.Format("2006-01-02 15:04:05"))

	card = collection[1]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "eld", card.Set.ParentCode)
	assert.Equal(t, "Throne of Eldraine Tokens", card.Set.Name)
	assert.Equal(t, "2019-09-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "token", card.Set.Typology)
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "19", card.CollectorNumber)
	assert.Equal(t, "d6c65749-1774-4b36-891e-abf762c95cec", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/teld/teld_19_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	assert.Equal(t, "6a5090b1-5eb6-4709-8208-ff3678be5756", card.OracleId)
	assert.Equal(t, uint64(0), card.MtgoID)
	assert.Equal(t, uint64(0), card.ArenaID)
	assert.Equal(t, uint64(0), card.TcgplayerID)
	assert.Equal(t, uint64(0), card.CardmarketID)
	assert.Equal(t, "emblem", card.Layout)
	assert.Equal(t, "", card.ManaCost)
	assert.Equal(t, float32(0), card.CMC)
	assert.Equal(t, "Emblem", card.TypeLine)
	assert.Equal(t, "Creatures you control get +3/+3 and have trample.", card.OracleText)
	assert.Equal(t, "", card.Power)
	assert.Equal(t, "", card.Toughness)
	assert.Equal(t, mtgdb.SliceString{}, card.Colors)
	assert.Equal(t, mtgdb.SliceString{}, card.ColorIdentity)
	assert.Equal(t, mtgdb.SliceString(nil), card.Keywords)
	assert.Equal(t, mtgdb.SliceString(nil), card.ProducedMana)
	assert.Equal(t, mtgdb.MapString{"brawl": "not_legal", "commander": "not_legal", "duel": "not_legal", "future": "not_legal", "historic": "not_legal", "legacy": "not_legal", "modern": "not_legal", "oldschool": "not_legal", "pauper": "not_legal", "penny": "not_legal", "pioneer": "not_legal", "standard": "not_legal", "vintage": "not_legal"}, card.Legalities)
	assert.Equal(t, mtgdb.SliceString{"paper"}, card.Games)
	assert.Equal(t, false, card.Oversized)
	assert.Equal(t, false, card.Promo)
	assert.Equal(t, false, card.Reprint)
	assert.Equal(t, false, card.Variation)
	assert.Equal(t, false, card.Digital)
	assert.Equal(t, "common", card.Rarity)
	assert.Equal(t, "", card.Watermark)
	assert.Equal(t, "Eric Deschamps", card.Artist)
	assert.Equal(t, "black", card.BorderColor)
	assert.Equal(t, "2015", card.Frame)
	assert.Equal(t, mtgdb.SliceString(nil), card.FrameEffects)
	assert.Equal(t, "", card.SecurityStamp)
	assert.Equal(t, false, card.FullArt)
	assert.Equal(t, false, card.Textless)
	assert.Equal(t, false, card.Booster)
	assert.Equal(t, false, card.StorySpotlight)
	assert.Equal(t, 0, len(card.Rulings))

	card = collection[2]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "ust", card.Set.ParentCode)
	assert.Equal(t, "Unstable", card.Set.Name)
	assert.Equal(t, "2017-12-08 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "funny", card.Set.Typology)
	assert.Equal(t, "ust", card.Set.IconName)
	assert.Equal(t, "65", card.CollectorNumber)
	assert.Equal(t, "cb3587b9-e727-4f37-b4d6-1baa7316262f", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/ust/ust_65_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	assert.Equal(t, "38bcba8b-2838-4ac8-9976-f9ccaa94fdba", card.OracleId)
	assert.Equal(t, uint64(0), card.MtgoID)
	assert.Equal(t, uint64(0), card.ArenaID)
	assert.Equal(t, uint64(153145), card.TcgplayerID)
	assert.Equal(t, uint64(0), card.CardmarketID)
	assert.Equal(t, "normal", card.Layout)
	assert.Equal(t, "{2}{B}", card.ManaCost)
	assert.Equal(t, float32(3), card.CMC)
	assert.Equal(t, "Enchantment", card.TypeLine)
	assert.Equal(t, "{3}{B}, Exile a permanent you control with a League of Dastardly Doom watermark: Return a permanent card with a League of Dastardly Doom watermark from your graveyard to the battlefield.", card.OracleText)
	assert.Equal(t, "", card.Power)
	assert.Equal(t, "", card.Toughness)
	assert.Equal(t, mtgdb.SliceString{"B"}, card.Colors)
	assert.Equal(t, mtgdb.SliceString{"B"}, card.ColorIdentity)
	assert.Equal(t, mtgdb.SliceString(nil), card.Keywords)
	assert.Equal(t, mtgdb.SliceString(nil), card.ProducedMana)
	assert.Equal(t, mtgdb.MapString{"brawl": "not_legal", "commander": "not_legal", "duel": "not_legal", "future": "not_legal", "historic": "not_legal", "legacy": "not_legal", "modern": "not_legal", "oldschool": "not_legal", "pauper": "not_legal", "penny": "not_legal", "pioneer": "not_legal", "standard": "not_legal", "vintage": "not_legal"}, card.Legalities)
	assert.Equal(t, mtgdb.SliceString{"paper"}, card.Games)
	assert.Equal(t, false, card.Oversized)
	assert.Equal(t, false, card.Promo)
	assert.Equal(t, false, card.Reprint)
	assert.Equal(t, false, card.Variation)
	assert.Equal(t, false, card.Digital)
	assert.Equal(t, "uncommon", card.Rarity)
	assert.Equal(t, "leagueofdastardlydoom", card.Watermark)
	assert.Equal(t, "Alex Konstad", card.Artist)
	assert.Equal(t, "silver", card.BorderColor)
	assert.Equal(t, "2015", card.Frame)
	assert.Equal(t, mtgdb.SliceString(nil), card.FrameEffects)
	assert.Equal(t, "", card.SecurityStamp)
	assert.Equal(t, false, card.FullArt)
	assert.Equal(t, false, card.Textless)
	assert.Equal(t, true, card.Booster)
	assert.Equal(t, false, card.StorySpotlight)
	assert.Equal(t, 0, len(card.Rulings))

	card = collection[3]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "eld", card.Set.ParentCode)
	assert.Equal(t, "Throne of Eldraine", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "expansion", card.Set.Typology)
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "191", card.CollectorNumber)
	assert.Equal(t, "abef512f-8f1d-4257-b16f-c0eed58670ec", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/eld/eld_191_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	// TODO other fields

	card = collection[4]
	assert.True(t, card.Foil)
	assert.False(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "eld", card.Set.ParentCode)
	assert.Equal(t, "Throne of Eldraine Promos", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "promo", card.Set.Typology)
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "1s", card.CollectorNumber)
	assert.Equal(t, "9a675b33-ab47-4a34-ab10-384e0de2f71f", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/peld/peld_1s_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	// TODO other fields

	card = collection[5]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "eld", card.Set.ParentCode)
	assert.Equal(t, "Throne of Eldraine Promos", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "promo", card.Set.Typology)
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "1p", card.CollectorNumber)
	assert.Equal(t, "77ba25cb-a8a6-46b6-82be-5c70e663dfdf", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/peld/peld_1p_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	// TODO other fields

	card = collection[6]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
	assert.Equal(t, "Nissa, Who Shakes the World", card.EnName)
	assert.Equal(t, "", card.EsName)
	assert.Equal(t, "", card.FrName)
	assert.Equal(t, "", card.DeName)
	assert.Equal(t, "", card.ItName)
	assert.Equal(t, "", card.PtName)
	assert.Equal(t, "世界を揺るがす者、ニッサ", card.JaName)
	assert.Equal(t, "", card.KoName)
	assert.Equal(t, "", card.RuName)
	assert.Equal(t, "", card.ZhsName)
	assert.Equal(t, "", card.ZhtName)
	assert.Equal(t, "war", card.SetCode)
	assert.Equal(t, "war", card.Set.Code)
	assert.Equal(t, "war", card.Set.ParentCode)
	assert.Equal(t, "War of the Spark", card.Set.Name)
	assert.Equal(t, "2019-05-03 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "expansion", card.Set.Typology)
	assert.Equal(t, "war", card.Set.IconName)
	assert.Equal(t, "169★", card.CollectorNumber)
	assert.Equal(t, "25d63632-c019-4f34-926a-42f829a4665c", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/war/war_169★_ja.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/war/war_169★_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	// TODO other fields

	card = collection[7]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.True(t, card.HasBackSide)
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
	assert.Equal(t, "isd", card.Set.ParentCode)
	assert.Equal(t, "Innistrad", card.Set.Name)
	assert.Equal(t, "2011-09-30 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "expansion", card.Set.Typology)
	assert.Equal(t, "isd", card.Set.IconName)
	assert.Equal(t, "176", card.CollectorNumber)
	assert.Equal(t, "25b54a1d-e201-453b-9173-b04e06ee6fb7", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/isd/isd_176_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/isd/isd_176_en_back.jpg"))
	assert.False(t, os.IsNotExist(err))
	// TODO other fields

	card = collection[8]
	assert.True(t, card.Foil)
	assert.True(t, card.NonFoil)
	assert.False(t, card.HasBackSide)
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
	assert.Equal(t, "eld", card.Set.ParentCode)
	assert.Equal(t, "Throne of Eldraine", card.Set.Name)
	assert.Equal(t, "2019-10-04 00:00:00 +0000 UTC", card.Set.ReleasedAt.String())
	assert.Equal(t, "expansion", card.Set.Typology)
	assert.Equal(t, "eld", card.Set.IconName)
	assert.Equal(t, "334", card.CollectorNumber)
	assert.Equal(t, "0dbf3260-b956-40da-abc7-764781c9f26f", card.ScryfallId)
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/eld/eld_334_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	// TODO other fields
}

func TestImporterBuildCardsFromJsonDownloadOnlyEnAssets(t *testing.T) {
	defer os.RemoveAll(TEMP_DIR)

	importer := mtgdb.NewImporter(filepath.Join(FIXTURES_PATH, "data"))
	importer.DownloadAssets = true
	importer.ImagesDir = filepath.Join(TEMP_DIR, "images")

	collection, downloadedImagesCount := importer.BuildCardsFromJson()
	sort.Slice(collection, func(i, j int) bool {
		return collection[i].ScryfallId > collection[j].ScryfallId
	})

	assert.Equal(t, uint32(11), downloadedImagesCount)
	// Index 6 is Nissa Japan
	card := collection[6]
	assert.Equal(t, "Nissa, Who Shakes the World", card.EnName)
	assert.Equal(t, "世界を揺るがす者、ニッサ", card.JaName)
	_, err := os.Stat(filepath.Join(importer.ImagesDir, "/cards/war/war_169★_en.jpg"))
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(importer.ImagesDir, "/cards/war/war_169★_ja.jpg"))
	assert.True(t, os.IsNotExist(err))
}

func TestBulkInsert(t *testing.T) {
	dbConnection := os.Getenv("DB_CONNECTION")
	if dbConnection == "" {
		dbConnection = "root@tcp(127.0.0.1:3306)/mtgdb_test?charset=utf8mb4&parseTime=True"
	}
	db, err := gorm.Open(mysql.Open(dbConnection), nil)
	if err != nil {
		panic(err)
	}
	db.Config.Logger = db.Config.Logger.LogMode(logger.Error)
	if os.Getenv("DB_LOG") == "1" {
		db.Config.Logger = db.Config.Logger.LogMode(logger.Info)
	}
	mtgdb.AutoMigrate(db)

	cards := []mtgdb.Card{
		{
			EnName:          "Gilded Goose",
			CollectorNumber: "160",
			SetCode:         "eld",
			Set: &mtgdb.Set{
				Name:     "Throne of Eldraine",
				Code:     "eld",
				IconName: "eld",
			},
			Foil:    true,
			NonFoil: true,
		}, {
			EnName:          "Acclaimed Contender",
			CollectorNumber: "1",
			SetCode:         "eld",
			Set: &mtgdb.Set{
				Name:     "Throne of Eldraine",
				Code:     "eld",
				IconName: "eld",
			},
			Foil:    true,
			NonFoil: true,
		}, {
			EnName:          "Daybreak Ranger // Nightfall Predator",
			CollectorNumber: "176",
			SetCode:         "isd",
			Set: &mtgdb.Set{
				Name:     "Innistrad",
				Code:     "isd",
				IconName: "isd",
			},
			Foil:    true,
			NonFoil: true,
		},
	}

	err = mtgdb.BulkInsert(db, cards)
	if err != nil {
		t.Fatal(err)
	}

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
	err := os.MkdirAll(TEMP_DIR, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(TEMP_DIR, "teferi.png")
	defer os.RemoveAll(TEMP_DIR)
	url := "https://c1.scryfall.com/file/scryfall-cards/normal/front/5/d/5d10b752-d9cb-419d-a5c4-d4ee1acb655e.jpg?1562736365"

	// Test download file
	err = mtgdb.DownloadFile(file, url)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, os.IsNotExist(err))

	// Test download file with different SHA1
	downloaded, err := mtgdb.DownloadFileWhenChanged(file, url, nil, "differentSHA1")
	if err != nil {
		t.Fatal(err)
	}
	currentFileTime := stat.ModTime()
	stat, err = os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, currentFileTime.Before(stat.ModTime()))
	assert.True(t, downloaded)

	// Test download file with same SHA1
	downloaded, err = mtgdb.DownloadFileWhenChanged(file, url, nil, "8b2ee43e87867e87a8fca7bfff0c7498f1d1fea8")
	if err != nil {
		t.Fatal(err)
	}
	currentFileTime = stat.ModTime()
	stat, err = os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, currentFileTime.Equal(stat.ModTime()))
	assert.False(t, downloaded)

	// Test download file with older time
	olderTime, _ := time.Parse(time.RFC3339, "1990-01-01T00:00:00.00Z")
	err = os.Chtimes(file, olderTime, olderTime)
	if err != nil {
		t.Fatal(err)
	}
	stat, err = os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	downloaded, err = mtgdb.DownloadFileWhenChanged(file, url, stat, "")
	if err != nil {
		t.Fatal(err)
	}
	stat, err = os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, olderTime.Equal(stat.ModTime()))
	assert.True(t, downloaded)

	// Test download file with newer time
	newerTime, _ := time.Parse(time.RFC3339, "2120-06-01T00:00:00.00Z")
	err = os.Chtimes(file, newerTime, newerTime)
	if err != nil {
		t.Fatal(err)
	}
	stat, err = os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	downloaded, err = mtgdb.DownloadFileWhenChanged(file, url, stat, "")
	if err != nil {
		t.Fatal(err)
	}
	stat, err = os.Stat(file)
	if err != nil {
		panic(err)
	}
	assert.True(t, newerTime.Equal(stat.ModTime()))
	assert.False(t, downloaded)
}
