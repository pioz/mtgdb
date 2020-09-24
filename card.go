package mtgdb

type Card struct {
	ID              uint   `gorm:"primary_key"`
	EnName          string `gorm:"not null"`
	EsName          string `gorm:"not null"`
	FrName          string `gorm:"not null"`
	DeName          string `gorm:"not null"`
	ItName          string `gorm:"not null"`
	PtName          string `gorm:"not null"`
	JaName          string `gorm:"not null"`
	KoName          string `gorm:"not null"`
	RuName          string `gorm:"not null"`
	ZhsName         string `gorm:"not null"`
	ZhtName         string `gorm:"not null"`
	SetCode         string `gorm:"not null"`
	Set             *Set   `gorm:"foreignkey:Code;association_foreignkey:SetCode"`
	CollectorNumber string `gorm:"not null"`
	Foil            bool   `gorm:"not null"`
	NonFoil         bool   `gorm:"not null"`
	HasBackSide     bool   `gorm:"not null"`
	IsToken         bool   `gorm:"not null"`
	ScryfallId      string
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
