package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	symbolChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

type GeneratorOptions struct {
	Length     int  `json:"length"`
	UseLower   bool `json:"use_lower"`
	UseUpper   bool `json:"use_upper"`
	UseDigits  bool `json:"use_digits"`
	UseSymbols bool `json:"use_symbols"`
}

func DefaultGeneratorOptions() GeneratorOptions {
	return GeneratorOptions{
		Length:     24,
		UseLower:   true,
		UseUpper:   true,
		UseDigits:  true,
		UseSymbols: true,
	}
}

func GeneratePasswordWithOptions(opts GeneratorOptions) (string, error) {
	if opts.Length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	var charset string
	if opts.UseLower {
		charset += lowerChars
	}
	if opts.UseUpper {
		charset += upperChars
	}
	if opts.UseDigits {
		charset += digitChars
	}
	if opts.UseSymbols {
		charset += symbolChars
	}

	if charset == "" {
		return "", errors.New("at least one character set must be selected")
	}

	charsetLen := big.NewInt(int64(len(charset)))
	result := make([]byte, opts.Length)

	for i := 0; i < opts.Length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate char: %w", err)
		}
		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}
