package mtgdb

type Card struct {
	ID              uint   `gorm:"primary_key"`
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
	Set             *Set   `gorm:"foreignkey:Code;association_foreignkey:SetCode"`
	CollectorNumber string `gorm:"not null"`
	IsToken         bool   `gorm:"not null"`
	IsDoubleFace    bool   `gorm:"not null"`
	ScryfallId      string
}

func (card *Card) IsValid() bool {
	return card.EnName != "" && card.SetCode != "" && card.CollectorNumber != ""
}

func (card *Card) ImagePath(dataImagesPath string, backImage bool) string {
	return CardImagePath(dataImagesPath, card.SetCode, card.CollectorNumber, backImage)
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
