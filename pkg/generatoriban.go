package pkg

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	cardNumberLength = 16
	cvvCodeLength    = 3
)

type Generator struct {
	countryCode string
	bankCode    string

	digitRunes []rune
}

func NewGenerator(countryCode string, bankCode string) *Generator {
	return &Generator{
		countryCode: countryCode,
		bankCode:    bankCode,
		digitRunes:  []rune("1234567890"),
	}
}

func (g Generator) GenerateRandomIban() string {
	return fmt.Sprintf("%s%s%s00000%s", g.countryCode, g.randStringRunes(2), g.bankCode, g.randStringRunes(14))
}

func (g Generator) GenerateRandomCardNumber() string {
	return g.randStringRunes(cardNumberLength)
}

func (g Generator) GenerateRandomCvv() string {
	return g.randStringRunes(cvvCodeLength)
}

func (g Generator) randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = g.digitRunes[rand.Intn(len(g.digitRunes))]
	}
	return string(b)
}
