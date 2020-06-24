package main

import "testing"

func TestExtractTitleAndYearMovies(t *testing.T) {
	var tests = []struct {
		input string
		title string
		year  int
	}{
		{"", "", 0},
		{"Joker (2019) [Bluray] [1080p] [YTS.LT]", "Joker", 2019},
		{"Joker.(2019).[Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker_(2019)_[Bluray]_[1080p]_[YTS.LT]", "Joker", 2019},
		{"Joker.2019.[Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker 2019 [Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker [ 2019 ] [Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker.[ 2019 ].[Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker [2019] [Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker.[2019].[Bluray].[1080p].[YTS.LT]", "Joker", 2019},
		{"Joker.[Bluray].[1080p].[YTS.LT]", "Joker", 0},
		{"2019.[Bluray].[1080p].[YTS.LT]", "2019", 0},
	}
	config := loadConfig()
	for _, test := range tests {
		title, year, _ := extractTitleAndYear(test.input, config)
		if title != test.title || year != test.year {
			t.Errorf("extractTitleAndYear(%q) = (%q, %d), want (%q, %d)", test.input, title, year, test.title, test.year)
		}
	}

}
