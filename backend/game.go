package main

import (
	"crypto/rand"
	"math/big"
)

const (
	codeLength  = 4
	maxAttempts = 10

	inProgress = "in_progress"
	won        = "won"
	lost       = "lost"
)

type game struct {
	secret   string
	attempts int
	status   string
}

type guessResult struct {
	CorrectNumbers    int    `json:"correctNumbers"`
	CorrectPositions  int    `json:"correctPositions"`
	RemainingAttempts int    `json:"remainingAttempts"`
	Status            string `json:"status"`
}

func generateCode() (string, error) {
	digits := []byte("0123456789")
	for i := 0; i < codeLength; i++ {
		offset, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits)-i)))
		if err != nil {
			return "", err
		}
		j := i + int(offset.Int64())
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits[:codeLength]), nil
}

func validCode(code string) bool {
	if len(code) != codeLength {
		return false
	}

	var seen [10]bool
	for i := 0; i < len(code); i++ {
		digit := code[i]
		if digit < '0' || digit > '9' || seen[digit-'0'] {
			return false
		}
		seen[digit-'0'] = true
	}
	return true
}

// score expects two valid codes. Digits in the correct position count in both totals.
func score(secret, guess string) (numbers, positions int) {
	var secretDigits [10]bool
	for i := 0; i < codeLength; i++ {
		secretDigits[secret[i]-'0'] = true
	}

	for i := 0; i < codeLength; i++ {
		if secretDigits[guess[i]-'0'] {
			numbers++
		}
		if secret[i] == guess[i] {
			positions++
		}
	}
	return numbers, positions
}

func (g *game) submit(guess string) guessResult {
	numbers, positions := score(g.secret, guess)
	g.attempts++
	remaining := maxAttempts - g.attempts

	switch {
	case positions == codeLength:
		g.status = won
	case remaining == 0:
		g.status = lost
	}

	return guessResult{
		CorrectNumbers:    numbers,
		CorrectPositions:  positions,
		RemainingAttempts: remaining,
		Status:            g.status,
	}
}
