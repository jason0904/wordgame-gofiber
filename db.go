package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Word struct {
	Id   int    `gorm:"column:id"`
	Word string `gorm:"column:word"`
	Part string `gorm:"column:part"`
}

// GORM에게 kr테이블을 사용함을 알려줌.
func (Word) TableName() string {
	return "kr"
}

func IsWordInDB(word string) bool {
	if DB == nil {
		log.Println("Database is not initialized.")
		return false
	}

	var result Word
	err := DB.Where("word = ?", word).First(&result).Error

	return err == nil
}

func GetRandomWordByLength(length int) (string, error) {
	if DB == nil {
		log.Println("Database is not initialized.")
		return "", fmt.Errorf("database is not initialized")
	}

	var result Word
	err := DB.Raw("SELECT * FROM kr WHERE LENGTH(word) = ? ORDER BY RANDOM() LIMIT 1", length).Scan(&result).Error

	if err != nil {
		return "", err
	}

	return result.Word, nil
}

func InitDB() error {
	db, err := gorm.Open(sqlite.Open("kr_korean.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	DB = db
	log.Println("Database connection successfully established.")
	return nil
}
