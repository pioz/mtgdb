package mtgdb

import "time"

type Set struct {
	Name       string `gorm:"not null"`
	Code       string `gorm:"not null"`
	ReleasedAt *time.Time
	IconName   string `gorm:"not null"`
}

func (self *Set) ImagePath(dataImagesPath string) string {
	return SetImagePath(dataImagesPath, self.IconName)
}
