package aslbus

import "sync"

// FIFO standard fifo buffer
type FIFO struct {
	lock   *sync.RWMutex
	items  []interface{}
	length int
	size   int
}

// NewFIFO creates a new fifo buffer
func NewFIFO(size int) *FIFO {
	return &FIFO{lock: new(sync.RWMutex), size: size}
}

// Items returns the current items in the buffer
func (f *FIFO) Items() []interface{} {
	f.lock.Lock()
	defer f.lock.Unlock()
	myItems := f.items
	return myItems
}

// Next returns the next element to be popped in the FIFO
func (f *FIFO) Next() interface{} {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.length > 0 {
		return f.items[0]
	}
	return nil
}

// Size returns the current fifo size
func (f *FIFO) Size() int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.size
}

// SetSize allow the size of the fifo to be modified
func (f *FIFO) SetSize(s int) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.size = s
}

// Full returns a true if the buffer has reached it maximum size
func (f *FIFO) Full() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(f.items) >= f.size {
		return true
	}
	return false
}

// Empty returns a true the buffer is empty
func (f *FIFO) Empty() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(f.items) == 0 {
		return true
	}
	return false
}

// Push add item to the fifo returns a false if this action was unsuccessful
func (f *FIFO) Push(item interface{}) bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(f.items) < f.size {
		f.items = append(f.items, item)
		f.length = len(f.items)
		return true
	}
	return false
}

// Pop the oldest item
func (f *FIFO) Pop() interface{} {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(f.items) == 0 {
		return nil
	}
	item := f.items[0]
	f.items = f.items[1:]
	f.length = len(f.items)
	return item
}

// Length return the current length of the fifo
func (f *FIFO) Length() int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.length
}

// MapFIFO implements a FIFO buffer but wrapped as if it's a map
type MapFIFO struct {
	lock *sync.RWMutex
	keys []string
	fifo *FIFO
}

// NewMapFIFO creates new map FIFO
func NewMapFIFO(size int) *MapFIFO {
	f := &MapFIFO{lock: new(sync.RWMutex)}
	f.fifo = NewFIFO(size)
	return f
}

// SetSize modifies the size of the fifo contained within
func (f *MapFIFO) SetSize(s int) {
	f.fifo.SetSize(s)
}

// Next returns the next element to be popped in the FIFO
func (f *MapFIFO) Next() interface{} {
	return f.fifo.Next()
}

// Size returns the current fifo size
func (f *MapFIFO) Size() int {
	return f.fifo.Size()
}

// Full returns a true if the buffer has reached it maximum size
func (f *MapFIFO) Full() bool {
	return f.fifo.Full()
}

// Empty returns a true the buffer is empty
func (f *MapFIFO) Empty() bool {
	return f.fifo.Empty()
}

// Push pushes the item into the buffer using the key as a references
// if the key exists in the fifo it will not be added
func (f *MapFIFO) Push(key string, item interface{}) bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	for _, myKey := range f.keys {
		if myKey == key {
			return false
		}
	}
	f.keys = append(f.keys, key)
	good := f.fifo.Push(item)
	if !good {
		return false
	}
	return true
}

// Pop returns the oldest element in the fifo removes its key from the buffer
func (f *MapFIFO) Pop() interface{} {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.fifo.Length() == 0 {
		return nil
	}
	f.keys = f.keys[1:]
	return f.fifo.Pop()
}
