package xsync

import "sync"

type Conditioner interface {
	WaitAndDo(condition func() bool, do func())
	Wait(condition func() bool)
	DoAndNotify(do func())
	DoAndNotifyAll(do func())
}

type conditioner struct {
	cond sync.Cond
}

func NewConditioner() Conditioner {
	return &conditioner{}
}

func (c *conditioner) WaitAndDo(condition func() bool, do func()) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for !condition() {
		c.cond.Wait()
	}

	do()
}

func (c *conditioner) Wait(condition func() bool) {
	c.WaitAndDo(condition, func() {})
}

func (c *conditioner) DoAndNotify(do func()) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	do()
	c.cond.Signal()
}

func (c *conditioner) DoAndNotifyAll(do func()) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	do()
	c.cond.Broadcast()
}

type RWConditioner interface {
	Conditioner

	RDo(do func())
}

type rwconditioner struct {
	conditioner

	rwlock sync.RWMutex
}

func NewConditionerRW() RWConditioner {
	c := &rwconditioner{}
	c.cond.L = &c.rwlock
	return c
}

func (c *rwconditioner) RDo(do func()) {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()

	do()
}
