package mtgdb

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&Set{})
	db.AutoMigrate(&Card{})
}
