package mtgdb

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&Set{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Card{})
	if err != nil {
		panic(err)
	}
}
