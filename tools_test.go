package main

import (
	"testing"
	"time"
)

func TestGetDateDixy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "valid date in Russian - January",
			input:    "15 янв 2023",
			expected: getTimeMoscowLayout("15 01 2023", "02 01 2006"),
		},
		{
			name:     "valid date in Russian - February",
			input:    "28 фев 2023",
			expected: getTimeMoscowLayout("28 02 2023", "02 01 2006"),
		},
		{
			name:     "valid date in Russian - December",
			input:    "31 дек 2023",
			expected: getTimeMoscowLayout("31 12 2023", "02 01 2006"),
		},
		{
			name:     "empty input returns zero time",
			input:    "",
			expected: time.Time{},
		},
		{
			name:     "invalid month",
			input:    "15 xyz 2023",
			expected: time.Time{},
		},
		{
			name:     "partial date",
			input:    "15 янв",
			expected: time.Time{},
		},
		{
			name:     "valid numeric date",
			input:    "15 01 2023",
			expected: getTimeMoscowLayout("15 01 2023", "02 01 2006"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDateDixy(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("Expected: %v, got: %v", tt.expected, result)
			}
		})
	}
}
