package mtgdb_test

import (
	"testing"

	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
)

func TestCardIsValid(t *testing.T) {
	card := mtgdb.Card{}
	assert.False(t, card.IsValid())

	card = mtgdb.Card{
		EnName:          "Gilded Goose",
		SetCode:         "eld",
		CollectorNumber: "160",
		IconName:        "eld",
	}
	assert.True(t, card.IsValid())
}

func TestCardImagePath(t *testing.T) {
	card := mtgdb.Card{
		EnName:          "Gilded Goose",
		SetCode:         "PELD",
		CollectorNumber: "160",
		IconName:        "eld",
	}
	assert.Equal(t, "images/cards/peld/peld_160.jpg", card.ImagePath("./images"))
}

func TestCardSetIconPath(t *testing.T) {
	card := mtgdb.Card{
		EnName:          "Gilded Goose",
		SetCode:         "peld",
		CollectorNumber: "160",
		IconName:        "eld",
	}
	assert.Equal(t, "images/sets/eld.jpg", card.SetIconPath("./images"))
}

func TestCardSetName(t *testing.T) {
	card := mtgdb.Card{}
	card.SetName("Goose", "it")
	assert.Empty(t, card.EnName)
	assert.Equal(t, "Goose", card.ItName)
}
