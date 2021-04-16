package mtgdb

import "time"

type Set struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"size:255;not null"`
	Code       string `gorm:"size:6;not null;uniqueIndex"`
	ParentCode string `gorm:"size:6;not null;index"`
	ReleasedAt *time.Time
	IconName   string `gorm:"size:255;not null"`
}

func (set *Set) ImagePath(dataImagesPath string) string {
	return SetImagePath(dataImagesPath, set.IconName)
}
