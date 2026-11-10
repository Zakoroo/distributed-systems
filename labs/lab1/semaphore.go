package main

import (
	"fmt"
	"sync"
)

// The current implementation of Semaphore uses a spin-lock mechanism, i.e., keep waiting in loop while there are no available resouce. This could be improved further using condition variable (performance optimization).
type Semaphore struct {
	mu    sync.Mutex
	count uint
}

// Initializer for semaphore. Notice that semaphores only accept positive integers as initial value, i.e., if initialized with 0 then an error will be returned.
func NewSemaphore(maxCount uint) (*Semaphore, error) {
	if maxCount == 0 {
		return nil, fmt.Errorf("Semaphore cannot be initialized with count == 0")
	}
	return &Semaphore{count: maxCount}, nil
}

// Increments the semaphore's count by 1.
func (c *Semaphore) Up() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

// Decrements the semaphore's count by 1 iff the count is greater than 0, else it will block the current go routine until another one calls `Up()`.
func (c *Semaphore) Down() {
	for {
		c.mu.Lock()
		if c.count == 0 {
			c.mu.Unlock()
			continue
		} else {
			c.count--
			c.mu.Unlock()
			break
		}
	}
}
