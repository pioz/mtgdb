package mtgdb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Card struct {
	ID              uint   `gorm:"primary_key"`
	EnName          string `gorm:"size:255;not null;index;index:idxft_cards_en_name,class:FULLTEXT"`
	EsName          string `gorm:"size:255;not null"`
	FrName          string `gorm:"size:255;not null"`
	DeName          string `gorm:"size:255;not null"`
	ItName          string `gorm:"size:255;not null"`
	PtName          string `gorm:"size:255;not null"`
	JaName          string `gorm:"size:255;not null"`
	KoName          string `gorm:"size:255;not null"`
	RuName          string `gorm:"size:255;not null"`
	ZhsName         string `gorm:"size:255;not null"`
	ZhtName         string `gorm:"size:255;not null"`
	SetCode         string `gorm:"size:6;not null;uniqueIndex:idx_cards_set_code_collector_number"`
	Set             *Set   `gorm:"foreignkey:SetCode;references:Code;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	CollectorNumber string `gorm:"size:255;not null;uniqueIndex:idx_cards_set_code_collector_number"`
	Foil            bool   `gorm:"not null"`
	NonFoil         bool   `gorm:"not null"`
	HasBackSide     bool   `gorm:"not null"`
	ReleasedAt      *time.Time
	ScryfallId      string `gorm:"size:255;not null"`
	MtgoID          uint64
	ArenaID         uint64
	TcgplayerID     uint64
	CardmarketID    uint64
	Layout          string
	ManaCost        string
	CMC             float32
	TypeLine        string
	OracleText      string
	Power           string
	Toughness       string
	Colors          []string  `gorm:"type:json"`
	ColorIdentity   []string  `gorm:"type:json"`
	Keywords        []string  `gorm:"type:json"`
	ProducedMana    []string  `gorm:"type:json"`
	Legalities      MapString `gorm:"type:json"`
	Games           []string  `gorm:"type:json"`
	Oversized       bool
	Promo           bool
	Reprint         bool
	Variation       bool
	Digital         bool
	Rarity          string
	Watermark       string
	Artist          string
	BorderColor     string
	Frame           string
	FrameEffects    []string `gorm:"type:json"`
	SecurityStamp   string
	FullArt         bool
	Textless        bool
	Booster         bool
	StorySpotlight  bool
}

func (card *Card) IsValid() bool {
	return card.EnName != "" && card.SetCode != "" && card.CollectorNumber != ""
}

func (card *Card) ImagePath(dataImagesPath, locale string, backImage bool) string {
	return CardImagePath(dataImagesPath, card.SetCode, card.CollectorNumber, locale, backImage)
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

type MapString map[string]interface{}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *MapString) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal json value:", value))
	}

	result := MapString{}
	err := json.Unmarshal(bytes, &result)
	*j = MapString(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j MapString) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}
