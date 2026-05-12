package model

import "sync"

type Capacitor struct {
	sync.Mutex
	cur, cap int
}

func NewCapacitor(cap int) *Capacitor {
	return &Capacitor{cap: cap}
}

func (c *Capacitor) Inc() (ok bool) {
	c.Lock()
	defer c.Unlock()
	if ok = c.cur <= c.cap; ok {
		c.cur += 1
	}
	return
}

func (c *Capacitor) Dec() {
	c.Lock()
	defer c.Unlock()
	c.cur -= 1
}
