package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	InitDB()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestIsWordInDB(t *testing.T) {
	testCases := []struct {
		name     string
		word     string
		expected bool
	}{
		{
			name:     "DB에 존재하는 단어",
			word:     "하늘",
			expected: true,
		},
		{
			name:     "DB에 존재하지 않는 단어",
			word:     "쀍쀍쀍",
			expected: false,
		},
		{
			name:     "빈 문자열",
			word:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsWordInDB(tc.word)
			assert.Equal(t, tc.expected, got, "IsWordInDB(%q) 결과가 예상과 다릅니다.", tc.word)
		})
	}
}

func TestGetRandomWordByLength(t *testing.T) {
	testCases := []struct {
		name        string
		length      int
		expectError bool
	}{
		{
			name:        "길이 2의 단어 가져오기",
			length:      2,
			expectError: false,
		},
		{
			name:        "길이 5의 단어 가져오기",
			length:      5,
			expectError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			word, err := GetRandomWordByLength(tc.length)
			if tc.expectError {
				assert.Error(t, err, "오류가 예상되었으나 발생하지 않았습니다.")
			} else {
				assert.NoError(t, err, "오류가 발생했습니다: %v", err)
				assert.Equal(t, tc.length, len([]rune(word)), "가져온 단어의 길이가 예상과 다릅니다.")
			}
		})
	}
}
