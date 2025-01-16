package util

import (
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func CreateULID() ulid.ULID {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
}

// Copy ULID bytes into UUID format
func ULIDToUUID(u ulid.ULID) uuid.UUID {
    var uuidBytes [16]byte
	
	// Convert ULID timestamp (uint64) to 6 bytes
	timestamp := u.Time()
	// First 6 bytes: ULID timestamp
	binary.BigEndian.PutUint64(uuidBytes[:8], timestamp)

    copy(uuidBytes[6:], u.Entropy()) // Remaining 10 bytes: ULID entropy

    return uuid.UUID(uuidBytes)
}

func UUIDToULID(u uuid.UUID) ulid.ULID {
	var ulidBytes [16]byte

	// Extract timestamp and entropy from the UUID
	copy(ulidBytes[:6], u[:6]) // First 6 bytes for timestamp
	copy(ulidBytes[6:], u[6:]) // Last 10 bytes for entropy

	// Convert the bytes back to a ULID
	return ulid.MustParse(string(ulidBytes[:]))
}