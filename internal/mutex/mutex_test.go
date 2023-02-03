package mutex

import (
	"testing"
	"time"
)

func TestKeyValueLock(t *testing.T) {
	keyValueMutex := New()
	keyValueMutex.Lock("foo")

	doneChannel := make(chan struct{})
	go func() {
		keyValueMutex.Lock("foo")
		close(doneChannel)
	}()

	select {
	case <-doneChannel:
		t.Fatal("Second lock was able to be taken. This shouldn't happen.")
	case <-time.After(50 * time.Millisecond):
		// Test passing.
	}
}

func TestKeyValueUnlock(t *testing.T) {
	keyValueMutex := New()
	keyValueMutex.Lock("foo")
	keyValueMutex.Unlock("foo")

	doneChannel := make(chan struct{})
	go func() {
		keyValueMutex.Lock("foo")
		close(doneChannel)
	}()

	select {
	case <-doneChannel:
		// Test passing.
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Second lock blocked after unlock. This shouldn't happen.")
	}
}

func TestKeyValueDifferentKeys(t *testing.T) {
	keyValueMutex := New()
	keyValueMutex.Lock("foo")

	doneChannel := make(chan struct{})
	go func() {
		keyValueMutex.Lock("bar")
		close(doneChannel)
	}()

	select {
	case <-doneChannel:
		// Test passing.
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Second lock on a different key blocked. This shouldn't happen.")
	}
}
