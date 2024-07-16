package db

import (
	"acc-server-manager/local/model"

	"go.uber.org/dig"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Start(di *dig.Container) {
	db, err := gorm.Open(sqlite.Open("acc.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = di.Provide(func() *gorm.DB {
		return db
	})
	if err != nil {
		panic("failed to bind database")
	}
	Migrate(db)
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&model.ApiModel{})
	if err != nil {
		panic("failed to migrate model.ApiModel")
	}
	db.FirstOrCreate(&model.ApiModel{Api: "Works"})
}
