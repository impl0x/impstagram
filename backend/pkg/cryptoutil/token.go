package cryptoutil

import (
	"crypto/rand"
	"encoding/base64"
	mathrand "math/rand/v2"
)

// returns random bytes of [size] prefixed with [prefix]
func GenerateToken(prefix string, size uint16) string {
	b := make([]byte, size)
	rand.Read(b) // ignore error as it is never returned according to the documentation
	return prefix + base64.RawURLEncoding.EncodeToString(b)
}

const hex = "0123456789abcdef"

type BitSize uint8

const (
	Bit8   BitSize = 2 << 0
	Bit16  BitSize = 2 << 1
	Bit32  BitSize = 2 << 2
	Bit64  BitSize = 2 << 3
	Bit128 BitSize = 2 << 4
	Bit256 BitSize = 2 << 5
)

// Generates a pseudo random hex token of specified bit size
func GeneratePseudoHexToken(size BitSize) string {
	id := make([]byte, size)
	for i := range id {
		id[i] = hex[mathrand.IntN(int(size))]
	}
	return string(id)
}

// Generates a pseudo random hex token of 32 bits
func GeneratePseudoHexToken32Bit() string {
	return GeneratePseudoHexToken(Bit32)
}

// Generates a pseudo random hex token of 64 bits
func GeneratePseudoHexToken64Bit() string {
	return GeneratePseudoHexToken(Bit64)
}
