package db

import (
	"go.uber.org/dig"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Start(di *dig.Container) {
	db, err := gorm.Open(sqlite.Open("acc.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	di.Provide(func() *gorm.DB {
		return db
	})
}
