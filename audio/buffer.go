package audio

import (
	"sync"
)

type RingBuffer struct {
	buffer []byte
	size   int
	read   int
	write  int
	mutex  sync.Mutex
	cond   *sync.Cond
}

func NewRingBuffer(size int) *RingBuffer {
	rb := &RingBuffer{
		buffer: make([]byte, size),
		size:   size,
		read:   0,
		write:  0,
	}
	rb.cond = sync.NewCond(&rb.mutex)
	return rb
}

func (rb *RingBuffer) Write(data []byte) int {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	n := len(data)
	if n > rb.Available() {
		n = rb.Available()
	}

	for i := 0; i < n; i++ {
		rb.buffer[rb.write] = data[i]
		rb.write = (rb.write + 1) % rb.size
	}

	rb.cond.Signal()
	return n
}

func (rb *RingBuffer) Read(data []byte) int {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	n := len(data)
	available := rb.Used()
	
	if n > available {
		n = available
	}

	for i := 0; i < n; i++ {
		data[i] = rb.buffer[rb.read]
		rb.read = (rb.read + 1) % rb.size
	}

	return n
}

func (rb *RingBuffer) Used() int {
	if rb.write >= rb.read {
		return rb.write - rb.read
	}
	return rb.size - rb.read + rb.write
}

func (rb *RingBuffer) Available() int {
	return rb.size - rb.Used() - 1
}

func (rb *RingBuffer) Reset() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	rb.read = 0
	rb.write = 0
}
