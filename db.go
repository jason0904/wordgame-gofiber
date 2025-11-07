package main

import (
	"context"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Word struct {
	Id   int    `gorm:"column:id"`
	Word string `gorm:"column:word"`
	Part string `gorm:"column:part"`
}

func IsWordInDB(word string) bool {
	db, err := initDB()
	if err != nil {
		return false
	}

	var result Word
	err = db.WithContext(context.Background()).Where("word = ?", word).First(&result).Error

	return err == nil
}

// 비공개 메서드

func initDB() (*gorm.DB, error) {
	//db 출처 https://github.com/korean-word-game/db?tab=readme-ov-file
	db, err := gorm.Open(sqlite.Open("kr_korean.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db, nil
}
