package files

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Locker struct {
	mu        *sync.RWMutex
	lockMap   map[string]*sync.RWMutex
	lockCount *atomic.Int64
}

func NewLocker() *Locker {
	return &Locker{
		mu:        &sync.RWMutex{},
		lockMap:   make(map[string]*sync.RWMutex),
		lockCount: &atomic.Int64{},
	}
}

func (fl *Locker) WithReadLock(filename string, handler func() error) error {
	if err := fl.RLock(filename); err != nil {
		return err
	}
	defer fl.RUnlock(filename)
	return handler()
}

func (fl *Locker) WithWriteLock(filename string, handler func() error) error {
	if err := fl.WLock(filename); err != nil {
		return err
	}
	defer fl.WUnlock(filename)
	return handler()
}

func (fl *Locker) RLock(filename string) error {
	return fl.lock(filename, "r")
}

func (fl *Locker) RUnlock(filename string) error {
	return fl.unlock(filename, "r")
}

func (fl *Locker) WLock(filename string) error {
	return fl.lock(filename, "w")
}

func (fl *Locker) WUnlock(filename string) error {
	return fl.unlock(filename, "w")
}

func (fl *Locker) lock(filename string, mode string) error {
	if !IsValidFileName(filename) {
		return fmt.Errorf("Invalid filename")
	}
	if mode != "r" && mode != "w" {
		return fmt.Errorf("Invalid file access mode")
	}

	lockFile := func(targetLock *sync.RWMutex) {
		switch mode {
		case "r":
			targetLock.RLock() // Lock file in read-mode
		case "w":
			targetLock.Lock() // Lock file in write-mode
		}
	}

	fl.mu.RLock() // Aquire master-lock
	fl.lockCount.Add(1)

	if fileLock, ok := fl.lockMap[filename]; ok {
		// The file already has an entry in the lockMap
		fl.mu.RUnlock() // Realse master-lock
		lockFile(fileLock)

	} else {
		// Create new entry for the file
		fl.mu.RUnlock() // Release master-lock
		fl.mu.Lock()    // Aquire master-lock in write-mode
		if fileLock, ok := fl.lockMap[filename]; !ok {
			fl.lockMap[filename] = &sync.RWMutex{} // Create new entry for the file
			lockFile(fl.lockMap[filename])
			fl.mu.Unlock() // Release master-lock
		} else {
			fl.mu.Unlock() // Unlock master lock early to avoid dead-lock
			lockFile(fileLock)
		}

	}

	return nil
}

func (fl *Locker) unlock(filename string, mode string) error {
	if !IsValidFileName(filename) {
		return fmt.Errorf("Invalid filename")
	}
	if mode != "r" && mode != "w" {
		return fmt.Errorf("Invalid file access mode")
	}

	unlockFile := func(targetLock *sync.RWMutex) {
		switch mode {
		case "r":
			targetLock.RUnlock() // Lock file in read-mode
		case "w":
			targetLock.Unlock() // Lock file in write-mode
		}
	}

	fl.mu.RLock() // Aquire master-lock

	if fileLock, ok := fl.lockMap[filename]; ok {
		// The file already has an entry in the lockMap
		fl.mu.RUnlock() // Realse master-lock
		unlockFile(fileLock)
	} else {
		// Create new entry for the file
		fl.mu.RUnlock() // Release master-lock
		return fmt.Errorf("The file `%v` does not exist", filename)
	}

	fl.lockCount.Add(-1)
	if x := fl.lockCount.Load(); x == 0 {
		fl.mu.Lock()
		if x := fl.lockCount.Load(); x == 0 {
			clear(fl.lockMap)
		}
		fl.mu.Unlock()
	}
	return nil
}
