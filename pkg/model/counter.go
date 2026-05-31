package model

import "sync"

type Counter struct {
	sync.RWMutex
	cur, max int
}

func NewCounter(max int) *Counter {
	return &Counter{max: max}
}

func (c *Counter) Inc() (ok bool) {
	c.Lock()
	defer c.Unlock()
	if ok = c.cur < c.max; ok {
		c.cur += 1
	}
	return
}

func (c *Counter) Dec() {
	c.Lock()
	defer c.Unlock()
	c.cur -= 1
}
