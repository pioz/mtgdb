package mtgdb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Card struct {
	ID uint `gorm:"primary_key"`

	EnName  string `gorm:"size:255;not null;index;index:idxft_cards_en_name,class:FULLTEXT"`
	EsName  string `gorm:"size:255;not null"`
	FrName  string `gorm:"size:255;not null"`
	DeName  string `gorm:"size:255;not null"`
	ItName  string `gorm:"size:255;not null"`
	PtName  string `gorm:"size:255;not null"`
	JaName  string `gorm:"size:255;not null"`
	KoName  string `gorm:"size:255;not null"`
	RuName  string `gorm:"size:255;not null"`
	ZhsName string `gorm:"size:255;not null"`
	ZhtName string `gorm:"size:255;not null"`

	SetCode string `gorm:"size:6;not null;uniqueIndex:idx_cards_set_code_collector_number"`
	Set     *Set   `gorm:"foreignkey:SetCode;references:Code;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`

	CollectorNumber string `gorm:"size:255;not null;uniqueIndex:idx_cards_set_code_collector_number"`
	Foil            bool   `gorm:"not null"`
	NonFoil         bool   `gorm:"not null"`
	HasBackSide     bool   `gorm:"not null"`
	ReleasedAt      *time.Time
	FrontImageUrl   string `gorm:"size:255;not null"`
	BackImageUrl    string `gorm:"size:255"`

	Artist             string `gorm:"size:255"`
	ArtistBack         string `gorm:"size:255"`
	Booster            bool
	BorderColor        string `gorm:"size:255"`
	CMC                float32
	CMCBack            float32
	ColorIdentity      SliceString `gorm:"type:json"`
	ColorIndicator     SliceString `gorm:"type:json"`
	ColorIndicatorBack SliceString `gorm:"type:json"`
	Colors             SliceString `gorm:"type:json"`
	ColorsBack         SliceString `gorm:"type:json"`
	ContentWarning     bool
	Digital            bool
	Finishes           SliceString `gorm:"type:json"`
	FlavorName         string      `gorm:"size:255"`
	FlavorText         string
	FlavorTextBack     string
	Frame              string      `gorm:"size:255"`
	FrameEffects       SliceString `gorm:"type:json"`
	FullArt            bool
	Games              SliceString `gorm:"type:json"`
	HandModifier       string      `gorm:"size:255"`
	Keywords           SliceString `gorm:"type:json"`
	Layout             string      `gorm:"size:255"`
	LayoutBack         string      `gorm:"size:255"`
	Legalities         MapString   `gorm:"type:json"`
	LifeModifier       string      `gorm:"size:255"`
	Loyalty            string      `gorm:"size:255"`
	LoyaltyBack        string      `gorm:"size:255"`
	ManaCost           string      `gorm:"size:255"`
	ManaCostBack       string      `gorm:"size:255"`
	OracleText         string
	OracleTextBack     string
	Oversized          bool
	Power              string      `gorm:"size:255"`
	PowerBack          string      `gorm:"size:255"`
	ProducedMana       SliceString `gorm:"type:json"`
	Promo              bool
	Rarity             string `gorm:"size:255"`
	Reprint            bool
	Reserved           bool
	SecurityStamp      string `gorm:"size:255"`
	StorySpotlight     bool
	Textless           bool
	Toughness          string `gorm:"size:255"`
	ToughnessBack      string `gorm:"size:255"`
	TypeLine           string `gorm:"size:255"`
	TypeLineBack       string `gorm:"size:255"`
	Variation          bool
	Watermark          string `gorm:"size:255"`
	WatermarkBack      string `gorm:"size:255"`

	ScryfallID   string `gorm:"size:255;not null"`
	OracleID     string `gorm:"size:255"`
	MtgoID       uint64
	ArenaID      uint64
	TcgplayerID  uint64
	CardmarketID uint64

	Rulings Rulings `gorm:"type:json"`
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

// Custom sql datatypes

type MapString map[string]interface{}

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

func (j MapString) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

type SliceString []string

func (j *SliceString) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal json value:", value))
	}

	result := SliceString{}
	err := json.Unmarshal(bytes, &result)
	*j = SliceString(result)
	return err
}

func (j SliceString) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

type Ruling struct {
	PublishedAt *time.Time `json:"published_at"`
	Comment     string     `json:"comment"`
}

func (j *Ruling) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal json value:", value))
	}

	result := Ruling{}
	err := json.Unmarshal(bytes, &result)
	*j = Ruling(result)
	return err
}

func (j Ruling) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Rulings []Ruling

func (j *Rulings) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal json value:", value))
	}

	result := Rulings{}
	err := json.Unmarshal(bytes, &result)
	*j = Rulings(result)
	return err
}

func (j Rulings) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}
