package importer_test

import (
	"testing"

	"github.com/pioz/mtgdb/importer"
	"github.com/stretchr/testify/assert"
)

func TestCardIsValid(t *testing.T) {
	card := importer.Card{}
	assert.False(t, card.IsValid())

	card = importer.Card{
		EnName:          "Gilded Goose",
		SetCode:         "eld",
		CollectorNumber: "160",
		IconName:        "eld",
	}
	assert.True(t, card.IsValid())
}

func TestCardSetName(t *testing.T) {
	card := importer.Card{}
	card.SetName("Goose", "it")
	assert.Empty(t, card.EnName)
	assert.Equal(t, "Goose", card.ItName)
}
