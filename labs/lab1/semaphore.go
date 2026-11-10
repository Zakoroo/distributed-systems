package main

import "sync"

type Semaphore struct {
	mu    sync.Mutex
	count uint
}

func NewSemaphore(maxCount uint) Semaphore {
	return Semaphore{count: maxCount}
}

func (c *Semaphore) Up() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

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
