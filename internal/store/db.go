package store

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)


type Word struct {
	ID   int    `gorm:"column:id"`
	Word string `gorm:"column:word"`
	Part string `gorm:"column:part"`
}

type DBManager struct {
	DB *gorm.DB
}

func NewDBManager() (*DBManager, error) {
	db, err := gorm.Open(sqlite.Open("data/kr_korean.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	log.Println("Database connection successfully established.")
	return &DBManager{DB: db}, nil
}

func (Word) TableName() string {
	return "kr"
}
func (db *DBManager) IsWordInDB(word string) bool {
	if db.DB == nil {
		log.Println("Database is not initialized.")
		return false
	}

	w := strings.TrimSpace(word)
	normalized := normalizeWord(w)

	if normalized == "" {
		return false
	}

	var result Word
	res := db.DB.Raw(
		"SELECT * FROM kr WHERE word = ? OR REPLACE(REPLACE(REPLACE(word, '-', ''), '^', ''), ' ', '') = ? LIMIT 1",
		w, normalized,
	).Scan(&result)

	if res.Error != nil {
		log.Println("Error querying database:", res.Error)
		return false
	}

	return res.RowsAffected > 0
}

func (db *DBManager) GetRandomWordByLength(length int) (string, error) {
	if db.DB == nil {
		log.Println("Database is not initialized.")
		return "", fmt.Errorf("database is not initialized")
	}

	var result Word
	err := db.DB.Raw("SELECT * FROM kr WHERE LENGTH(word) = ? ORDER BY RANDOM() LIMIT 1", length).Scan(&result).Error

	if err != nil {
		return "", err
	}

	randomWord := strings.TrimSpace(result.Word)
	normalized := strings.ReplaceAll(randomWord, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "^", "")

	return normalized, nil
}

func normalizeWord(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(" ", "", "-", "", "^", "")
	return replacer.Replace(s)
}
