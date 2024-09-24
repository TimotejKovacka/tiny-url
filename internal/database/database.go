package database

import (
	"github.com/TimotejKovacka/tiny-url/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	dsn := "host=localhost user=admin password=admin dbname=test_db port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	registerModels(db)

	return db
}

func registerModels(db *gorm.DB) {
	db.AutoMigrate(&models.UrlModel{})
}
