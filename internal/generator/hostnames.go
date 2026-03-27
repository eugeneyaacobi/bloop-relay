package generator

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandomLabel(length int) (string, error) {
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b.WriteByte(letters[n.Int64()])
	}
	return b.String(), nil
}

func BuildHostname(label, domain string) string {
	return label + "." + domain
}
