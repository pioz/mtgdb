package mtgdb

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
	Set             Set    `gorm:"foreignkey:Code;association_foreignkey:SetCode"`
	CollectorNumber string `gorm:"not null"`
	IsToken         bool   `gorm:"not null"`
	ScryfallId      string
}

func (self *Card) IsValid() bool {
	return self.EnName != "" && self.SetCode != "" && self.CollectorNumber != ""
}

func (self *Card) ImagePath(dataImagesPath string) string {
	return CardImagePath(dataImagesPath, self.SetCode, self.CollectorNumber)
}

func (self *Card) SetName(name, language string) {
	switch language {
	case "es", "Spanish":
		self.EsName = name
	case "fr", "French":
		self.FrName = name
	case "de", "German":
		self.DeName = name
	case "it", "Italian":
		self.ItName = name
	case "pt", "Portuguese", "Portuguese (Brazil)":
		self.PtName = name
	case "ja", "jp", "Japanese":
		self.JaName = name
	case "ko", "Korean":
		self.KoName = name
	case "ru", "Russian":
		self.RuName = name
	case "zhs", "Chinese Simplified":
		self.ZhsName = name
	case "zht", "Chinese Traditional":
		self.ZhtName = name
	}
}
