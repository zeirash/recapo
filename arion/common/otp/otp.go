package otp

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type entry struct {
	code      string
	expiresAt time.Time
	sentAt    time.Time
}

var store sync.Map

const TTL = 10 * time.Minute
const Cooldown = 60 * time.Second

// Generate creates a random 6-digit OTP for the given email, stores it, and returns the code.
func Generate(email string) string {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	store.Store(email, entry{code: code, expiresAt: time.Now().Add(TTL), sentAt: time.Now()})
	return code
}

// CanResend returns true if no OTP has been sent for the key within the cooldown period.
func CanResend(key string) bool {
	v, ok := store.Load(key)
	if !ok {
		return true
	}
	return time.Since(v.(entry).sentAt) >= Cooldown
}

// Verify returns true if the code matches and has not expired.
func Verify(email, code string) bool {
	v, ok := store.Load(email)
	if !ok {
		return false
	}
	e := v.(entry)
	if time.Now().After(e.expiresAt) {
		store.Delete(email)
		return false
	}
	return e.code == code
}

// Delete removes the OTP for the given email from the store.
func Delete(email string) {
	store.Delete(email)
}
