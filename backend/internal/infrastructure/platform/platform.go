// Пакет platform содержит реализации вспомогательных портов:
// генератор идентификаторов и системные часы.
package platform

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// UUIDGenerator формирует случайные идентификаторы (128 бит, hex).
type UUIDGenerator struct{}

// NewID возвращает новый случайный идентификатор.
func (UUIDGenerator) NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// SystemClock возвращает текущее время ОС в UTC.
type SystemClock struct{}

// Now реализует usecase.Clock.
func (SystemClock) Now() time.Time { return time.Now().UTC() }
