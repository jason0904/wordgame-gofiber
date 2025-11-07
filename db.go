package main

import (
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

// InitDB 함수는 한번만 호출되게 해야함. game에서 해야하나..
func InitDB() {
	//db 출처 https://github.com/korean-word-game/db?tab=readme-ov-file
	db, err := gorm.Open(sqlite.Open("kr_korean.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	log.Println("Database connection successfully established.")
}
