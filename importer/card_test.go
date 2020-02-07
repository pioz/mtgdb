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

func TestCardImagePath(t *testing.T) {
	card := importer.Card{
		EnName:          "Gilded Goose",
		SetCode:         "peld",
		CollectorNumber: "160",
		IconName:        "eld",
	}
	assert.Equal(t, "images/cards/peld/peld_160.jpg", card.ImagePath("./images"))
}

func TestCardSetIconPath(t *testing.T) {
	card := importer.Card{
		EnName:          "Gilded Goose",
		SetCode:         "peld",
		CollectorNumber: "160",
		IconName:        "eld",
	}
	assert.Equal(t, "images/sets/eld.jpg", card.SetIconPath("./images"))
}

func TestCardSetName(t *testing.T) {
	card := importer.Card{}
	card.SetName("Goose", "it")
	assert.Empty(t, card.EnName)
	assert.Equal(t, "Goose", card.ItName)
}
