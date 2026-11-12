package files

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Zakoroo/distributed-systems/concurrent/locks"
)

type Manager struct {
	locks *locks.LockMap[string]
}

func NewManager() *Manager {
	return &Manager{
		locks: locks.NewLockMap[string](),
	}
}

// WithFile synchronizes file access. Callers are responsible for filename and
// path policy; this method does not validate names or restrict filesystem access.
func (fm *Manager) WithFile(filename, mode string, perm os.FileMode, handler func(*os.File) error) error {
	// Normalize lexical aliases so relative and absolute paths share a lock key.
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	genericHandler := func(flags int) error {
		f, err := os.OpenFile(filename, flags, perm)
		if err != nil {
			return err
		}
		defer f.Close()
		return handler(f)
	}

	switch mode {
	case "r":
		return fm.locks.WithReadLock(filename, func() error {
			return genericHandler(os.O_RDONLY)
		})
	case "w":
		return fm.locks.WithWriteLock(filename, func() error {
			return genericHandler(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)
		})
	case "a":
		return fm.locks.WithWriteLock(filename, func() error {
			return genericHandler(os.O_CREATE | os.O_WRONLY | os.O_APPEND)
		})
	default:
		return fmt.Errorf("Access mode `%v` is unsupported", mode)
	}
}
