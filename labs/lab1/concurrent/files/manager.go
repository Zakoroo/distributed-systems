package files

import (
	"fmt"
	"os"
)

type Manager struct {
	locker *Locker
}

func NewFileManager() *Manager {
	return &Manager{
		locker: NewLocker(),
	}
}

func (fm *Manager) WithFile(filename, mode string, perm os.FileMode, handler func(*os.File) error) error {
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
		return fm.locker.WithReadLock(filename, func() error {
			return genericHandler(os.O_RDONLY)
		})
	case "w":
		return fm.locker.WithWriteLock(filename, func() error {
			return genericHandler(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)
		})
	case "a":
		return fm.locker.WithWriteLock(filename, func() error {
			return genericHandler(os.O_CREATE | os.O_WRONLY | os.O_APPEND)
		})
	default:
		return fmt.Errorf("Access mode `%v` is unsupported", mode)
	}
}
