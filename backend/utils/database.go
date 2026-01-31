package utils

import (
	"codeagent-backend/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	err = DB.AutoMigrate(&models.LLMConfig{}, &models.Project{}, &models.Prompt{}, &models.TestCase{}, &models.LLMTestCase{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
