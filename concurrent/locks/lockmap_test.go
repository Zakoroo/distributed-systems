package locks

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGenericKeysAndExclusiveUpdates(t *testing.T) {
	type resource struct{ ID int }
	var locks LockMap[resource] // Also exercise the zero value.
	var wg sync.WaitGroup
	value := 0
	for range 20 {
		wg.Go(func() {
			for range 100 {
				if err := locks.WithWriteLock(resource{1}, func() error {
					previous := value
					runtime.Gosched()
					value = previous + 1
					return nil
				}); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	if value != 2000 {
		t.Fatalf("exclusive updates = %d, want 2000", value)
	}
}

func TestReadersShareKeyAndOtherKeysProgress(t *testing.T) {
	locks := NewLockMap[string]()
	held := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- locks.WithReadLock("not/a/filename?", func() error {
			close(held)
			<-release
			return nil
		})
	}()
	<-held
	t.Cleanup(func() { close(release); <-finished })
	for _, operation := range []func() error{
		func() error { return locks.WithReadLock("not/a/filename?", func() error { return nil }) },
		func() error { return locks.WithWriteLock("other", func() error { return nil }) },
	} {
		done := make(chan error, 1)
		go func() { done <- operation() }()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("independent operation blocked on an unrelated/read-shared lock")
		}
	}
}

func TestCallbackFailureAndPanicReleaseLock(t *testing.T) {
	locks := NewLockMap[string]()
	failure := errors.New("callback failed")
	if err := locks.WithWriteLock("key", func() error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("callback error = %v, want %v", err, failure)
	}
	func() {
		defer func() {
			if recovered := recover(); recovered != failure {
				t.Errorf("panic = %v, want %v", recovered, failure)
			}
		}()
		locks.WithWriteLock("key", func() error { panic(failure) })
	}()
	done := make(chan error, 1)
	go func() { done <- locks.WithWriteLock("key", func() error { return nil }) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("failed callback left the lock held")
	}
}
