package mtgdb

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Card struct {
	EnName          string `gorm:"not null"`
	EsName          string
	FrName          string
	DeName          string
	ItName          string
	PtName          string
	JaName          string
	KoName          string
	RuName          string
	ZhsName         string
	ZhtName         string
	SetCode         string `gorm:"not null"`
	CollectorNumber string `gorm:"not null"`
	IsToken         bool   `gorm:"not null"`
	IconName        string `gorm:"not null"`
	ScryfallId      string
	ExternalId      string
}

func (card *Card) IsValid() bool {
	return card.EnName != "" && card.SetCode != "" && card.CollectorNumber != "" && card.IconName != ""
}

func (card *Card) ImagePath(dataImagesPath string) string {
	return filepath.Join(dataImagesPath, "cards", strings.ToLower(card.SetCode), fmt.Sprintf("%s_%s.jpg", strings.ToLower(card.SetCode), strings.ToLower(card.CollectorNumber)))
}

func (card *Card) SetIconPath(dataImagesPath string) string {
	return filepath.Join(dataImagesPath, "sets", fmt.Sprintf("%s.jpg", strings.ToLower(card.IconName)))
}

func (card *Card) SetName(name, language string) {
	switch language {
	case "es", "Spanish":
		card.EsName = name
	case "fr", "French":
		card.FrName = name
	case "de", "German":
		card.DeName = name
	case "it", "Italian":
		card.ItName = name
	case "pt", "Portuguese", "Portuguese (Brazil)":
		card.PtName = name
	case "ja", "jp", "Japanese":
		card.JaName = name
	case "ko", "Korean":
		card.KoName = name
	case "ru", "Russian":
		card.RuName = name
	case "zhs", "Chinese Simplified":
		card.ZhsName = name
	case "zht", "Chinese Traditional":
		card.ZhtName = name
	}
}
