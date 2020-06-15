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
		CollectorNumber: "160",
		SetCode:         "eld",
	}
	assert.True(t, card.IsValid())
}

func TestCardImagePath(t *testing.T) {
	card := mtgdb.Card{
		EnName:          "Gilded Goose",
		CollectorNumber: "160",
		SetCode:         "peld",
	}
	assert.Equal(t, "images/cards/peld/peld_160.jpg", card.ImagePath("./images", false))
	assert.Equal(t, "images/cards/peld/peld_160_back.jpg", card.ImagePath("./images", true))
}

func TestCardSetName(t *testing.T) {
	card := mtgdb.Card{}
	card.SetName("Goose", "it")
	assert.Empty(t, card.EnName)
	assert.Equal(t, "Goose", card.ItName)
}
