package mtgdb

import "time"

type Set struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"not null"`
	Code       string `gorm:"not null"`
	ParentCode string `gorm:"not null"`
	ReleasedAt *time.Time
	IconName   string `gorm:"not null"`
}

func (set *Set) ImagePath(dataImagesPath string) string {
	return SetImagePath(dataImagesPath, set.IconName)
}
