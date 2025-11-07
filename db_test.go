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
