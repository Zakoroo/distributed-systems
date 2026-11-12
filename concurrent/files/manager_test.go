package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestManagerDoesNotValidateFilenames(t *testing.T) {
	manager := NewManager()
	filename := filepath.Join(t.TempDir(), "-caller-selected.txt")
	if err := manager.WithFile(filename, "w", 0o644, func(f *os.File) error {
		_, err := f.WriteString("allowed by filesystem")
		return err
	}); err != nil {
		t.Fatalf("manager imposed filename policy: %v", err)
	}
}

func TestManagerReadsCompleteVersionsAcrossPathAliases(t *testing.T) {
	manager := NewManager()
	root := t.TempDir()
	filename := filepath.Join(root, "file.txt")
	alias := root + "/./file.txt"
	if err := os.WriteFile(filename, []byte("leftright"), 0o644); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			for range 100 {
				err := manager.WithFile(alias, "w", 0o644, func(f *os.File) error {
					if _, err := f.WriteString("left"); err != nil {
						return err
					}
					runtime.Gosched()
					_, err := f.WriteString("right")
					return err
				})
				if err != nil {
					t.Error(err)
					return
				}
			}
		})
		wg.Go(func() {
			for range 100 {
				err := manager.WithFile(filename, "r", 0o644, func(f *os.File) error {
					data, err := io.ReadAll(f)
					if err != nil {
						return err
					}
					if string(data) != "leftright" {
						return fmt.Errorf("observed partial file version %q", data)
					}
					return nil
				})
				if err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
}
