package id

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var entropy = ulid.Monotonic(rand.Reader, 0)

func New() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func NewWithPrefix(prefix string) string {
	return prefix + ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func MustParse(s string) ulid.ULID {
	return ulid.MustParse(s)
}

func Parse(s string) (ulid.ULID, error) {
	return ulid.Parse(s)
}
