package mtgdb_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
)

const FIXTURES_PATH = "./fixtures"

func TestImport(t *testing.T) {
	defer os.RemoveAll("/tmp/mtgscan_test")

	importer := mtgdb.NewImporter(filepath.Join(FIXTURES_PATH, "data"))
	importer.DownloadAssets = true
	importer.ImagesDir = "/tmp/mtgscan_test/images"

	collection := importer.BuildCardsFromJson()
	sort.Slice(collection, func(i, j int) bool {
		return collection[i].ScryfallId > collection[j].ScryfallId
	})

	_, err := os.Stat("/tmp/mtgscan_test/images/sets/eld.png")
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat("/tmp/mtgscan_test/images/sets/isd.png")
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat("/tmp/mtgscan_test/images/sets/ust.png")
	assert.False(t, os.IsNotExist(err))

	assert.Equal(t, 8, len(collection))

	card := collection[0]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "1", card.CollectorNumber)
	assert.Equal(t, "fb6b12e7-bb93-4eb6-bad1-b256a6ccff4e", card.ScryfallId)
	assert.Equal(t, "eld", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/eld/eld_1.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[1]
	assert.True(t, card.IsToken)
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
	assert.Equal(t, "19", card.CollectorNumber)
	assert.Equal(t, "d6c65749-1774-4b36-891e-abf762c95cec", card.ScryfallId)
	assert.Equal(t, "eld", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/teld/teld_19.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[2]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "65", card.CollectorNumber)
	assert.Equal(t, "cb3587b9-e727-4f37-b4d6-1baa7316262f", card.ScryfallId)
	assert.Equal(t, "ust", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/ust/ust_65.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[3]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "191", card.CollectorNumber)
	assert.Equal(t, "abef512f-8f1d-4257-b16f-c0eed58670ec", card.ScryfallId)
	assert.Equal(t, "eld", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/eld/eld_191.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[4]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "1s", card.CollectorNumber)
	assert.Equal(t, "9a675b33-ab47-4a34-ab10-384e0de2f71f", card.ScryfallId)
	assert.Equal(t, "eld", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/peld/peld_1s.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[5]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "1p", card.CollectorNumber)
	assert.Equal(t, "77ba25cb-a8a6-46b6-82be-5c70e663dfdf", card.ScryfallId)
	assert.Equal(t, "eld", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/peld/peld_1p.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[6]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "176", card.CollectorNumber)
	assert.Equal(t, "25b54a1d-e201-453b-9173-b04e06ee6fb7", card.ScryfallId)
	assert.Equal(t, "isd", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/isd/isd_176.png")
	assert.False(t, os.IsNotExist(err))

	card = collection[7]
	assert.False(t, card.IsToken)
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
	assert.Equal(t, "334", card.CollectorNumber)
	assert.Equal(t, "0dbf3260-b956-40da-abc7-764781c9f26f", card.ScryfallId)
	assert.Equal(t, "eld", card.IconName)
	_, err = os.Stat("/tmp/mtgscan_test/images/cards/eld/eld_334.png")
	assert.False(t, os.IsNotExist(err))
}
