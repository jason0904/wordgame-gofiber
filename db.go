package main

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Word struct {
	Id   int    `gorm:"column:id"`
	Word string `gorm:"column:word"`
	Part string `gorm:"column:part"`
}

func (Word) TableName() string {
	return "kr"
}

func IsWordInDB(word string) bool {
	if DB == nil {
		log.Println("Database is not initialized.")
		return false
	}

	w := strings.TrimSpace(word)
	normalized := strings.ReplaceAll(w, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "^", "")

	var result Word
	// 원문 일치 또는 구분자 제거 후 일치 검사
	err := DB.Raw(
		"SELECT * FROM kr WHERE word = ? OR REPLACE(REPLACE(REPLACE(word, '-', ''), '^', ''), ' ', '') = ? LIMIT 1",
		word, normalized,
	).Scan(&result).Error

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

	randomWord := strings.TrimSpace(result.Word)
	normalized := strings.ReplaceAll(randomWord, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "^", "")

	return normalized, nil
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
