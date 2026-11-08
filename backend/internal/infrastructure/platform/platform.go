package platform

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type UUIDGenerator struct{}

func (UUIDGenerator) NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }
