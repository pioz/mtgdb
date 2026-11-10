package mtgdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var DownloadFile = downloadFile
var DownloadFileWhenChanged = downloadFileWhenChanged

// Reversible cards like the REX basic lands are single-faced in non-English
// prints and double-faced only in the English one. When a non-English print is
// imported first, the English print must still set HasBackSide.
func TestBuildCardEnPrintAfterNonEnSetsBackSide(t *testing.T) {
	importer := &Importer{
		ImageType:         "normal",
		cardCollection:    make(map[string]*Card),
		setCollection:     make(map[string]*Set),
		rulingsCollection: make(map[string]Rulings),
	}

	jaPrint := &cardJsonStruct{
		Name: "Forest", Lang: "ja", SetCode: "rex", CollectorNumber: "25", Layout: "normal",
		ImageUris: imagesCardJsonStruct{Normal: "https://example.com/ja.jpg"},
	}
	enPrint := &cardJsonStruct{
		Name: "Forest // Forest", Lang: "en", SetCode: "rex", CollectorNumber: "25", Layout: "reversible_card",
		CardFaces: []cardFaceStruct{
			{PrintedName: "Forest", Layout: "normal", ImageUris: imagesCardJsonStruct{Normal: "https://example.com/front.jpg"}},
			{PrintedName: "Forest", Layout: "normal", ImageUris: imagesCardJsonStruct{Normal: "https://example.com/back.jpg"}},
		},
	}

	importer.buildCard(jaPrint)
	importer.buildCard(enPrint)

	card := importer.cardCollection["rex-25"]
	assert.Equal(t, "Forest", card.EnName)
	assert.True(t, card.HasBackSide)
	assert.Equal(t, "https://example.com/front.jpg", card.FrontImageUrl)
	assert.Equal(t, "https://example.com/back.jpg", card.BackImageUrl)
}
