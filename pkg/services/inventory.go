package services

import (
	rand2 "crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// SKUGenerator handles generation of Stock Keeping Unit codes
type SKUGenerator struct {
	Prefix     string
	Separator  string
	DigitCount int
}

// NewSKUGenerator creates a new SKU generator with default settings
func NewSKUGenerator() *SKUGenerator {
	return &SKUGenerator{
		Prefix:     "",
		Separator:  "-",
		DigitCount: 6,
	}
}

// GenerateSKU generates a new SKU with optional prefix
func (g *SKUGenerator) GenerateSKU() string {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(g.DigitCount)), nil)
	n, _ := rand2.Int(rand2.Reader, max)
	number := fmt.Sprintf("%0*d", g.DigitCount, n.Int64())

	if g.Prefix != "" {
		return fmt.Sprintf("%s%s%s", g.Prefix, g.Separator, number)
	}
	return number
}

// GenerateTimedSKU generates a SKU including timestamp components
func (g *SKUGenerator) GenerateTimedSKU() string {
	now := time.Now()
	timeComponent := fmt.Sprintf("%02d%02d", now.Month(), now.Day())

	// Generate random suffix
	suffix, _ := rand2.Int(rand2.Reader, big.NewInt(9999))

	if g.Prefix != "" {
		return fmt.Sprintf("%s%s%s%s%04d",
			g.Prefix, g.Separator,
			timeComponent, g.Separator,
			suffix.Int64())
	}
	return fmt.Sprintf("%s%s%04d", timeComponent, g.Separator, suffix.Int64())
}

// GeneratePatternSKU generates a SKU based on a pattern
func (g *SKUGenerator) GeneratePatternSKU(pattern string) string {
	result := strings.Builder{}

	for _, char := range pattern {
		switch char {
		case 'A': // Random uppercase letter
			result.WriteRune(rune('A' + randInt(26)))
		case 'N': // Random number
			result.WriteRune(rune('0' + randInt(10)))
		case '-': // Separator
			result.WriteRune('-')
		default: // Keep any other character as is
			result.WriteRune(char)
		}
	}

	return result.String()
}

// randInt generates random integer for internal use
func randInt(max int) int {
	n, _ := rand2.Int(rand2.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}