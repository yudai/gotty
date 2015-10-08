package umutex

import (
	"sync"
	"testing"
)

func TestTryLock(t *testing.T) {
	var result bool

	mutex := New()

	result = mutex.TryLock()
	if result != true {
		t.Error()
	}

	result = mutex.TryLock()
	if result != false {
		t.Error()
	}

	mutex.Unlock()

	result = mutex.TryLock()
	if result != true {
		t.Error()
	}
}

func TestForceLock(t *testing.T) {
	var result bool

	mutex := New()

	result = mutex.TryLock()
	if result != true {
		t.Error()
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mutex.ForceLock()
	}()

	mutex.Unlock()
	wg.Wait()
}
