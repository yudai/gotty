// Package umutex provides unblocking mutex
package umutex

// UnblockingMutex represents an unblocking mutex.
type UnblockingMutex struct {
	// Raw channel
	C chan bool
}

// New returnes a new unblocking mutex instance.
func New() *UnblockingMutex {
	return &UnblockingMutex{
		C: make(chan bool, 1),
	}
}

// TryLock tries to lock the mutex.
// When the mutex is free at the time, the function locks the mutex and return
// true. Otherwise false will be returned. In the both cases, this function
// doens't block and return the result immediately.
func (m UnblockingMutex) TryLock() (result bool) {
	select {
	case m.C <- true:
		return true
	default:
		return false
	}
}

// Unlock unclocks the mutex.
func (m UnblockingMutex) Unlock() {
	<-m.C
}

// ForceLock surely locks the mutex, however, this function blocks when the mutex is locked at the time.
func (m UnblockingMutex) ForceLock() {
	m.C <- false
}
