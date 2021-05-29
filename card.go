package mtgdb

type Card struct {
	ID              uint   `gorm:"primary_key"`
	EnName          string `gorm:"size:255;not null;index"`
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
	ScryfallId      string `gorm:"size:255;not null"`
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
