package auth

import (
	ulid "github.com/oklog/ulid/v2"
	"math/rand"
	"time"
)

func CreateULID() ulid.ULID {
	utime := time.Now()
	source := rand.NewSource(utime.UnixNano())
	entropy := ulid.Monotonic(rand.New(source), 0)
	id := ulid.MustNew(ulid.Timestamp(utime), entropy)
	return id
}
