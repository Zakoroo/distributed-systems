package semaphores

import (
	"fmt"
	"sync"
)

// Semaphore implemented using a condition variable and a mutex to manage concurrency.
type Semaphore struct {
	mu    *sync.Mutex
	cond  *sync.Cond
	count uint
}

// Initializer for semaphore. Notice that semaphores only accept positive integers as initial value, i.e., if initialized with 0 then an error will be returned.
func NewSemaphore(maxCount uint) (*Semaphore, error) {
	if maxCount == 0 {
		return nil, fmt.Errorf("Semaphore cannot be initialized with count == 0")
	}

	mu := &sync.Mutex{}
	return &Semaphore{
		mu:    mu,
		cond:  sync.NewCond(mu),
		count: maxCount,
	}, nil
}

// Increments the semaphore's count by 1.
func (c *Semaphore) Up() {
	c.mu.Lock()
	c.count++
	c.cond.Signal()
	c.mu.Unlock()
}

// Decrements the semaphore's count by 1 iff the count is greater than 0, else it will block the current go routine until another one calls `Up()`.
func (c *Semaphore) Down() {
	c.mu.Lock()
	for c.count == 0 {
		c.cond.Wait()
	}
	c.count--
	c.mu.Unlock()
}
