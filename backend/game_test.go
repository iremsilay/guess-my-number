package main

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		secret, guess      string
		numbers, positions int
	}{
		{"9724", "9732", 3, 2},
		{"0123", "0213", 4, 2},
		{"0123", "1230", 4, 0},
		{"0123", "0123", 4, 4},
		{"0123", "4567", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.secret+"/"+tt.guess, func(t *testing.T) {
			numbers, positions := score(tt.secret, tt.guess)
			if numbers != tt.numbers || positions != tt.positions {
				t.Errorf("score = (%d, %d), want (%d, %d)", numbers, positions, tt.numbers, tt.positions)
			}
		})
	}
}
