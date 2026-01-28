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

	part1 := rb.size - rb.write
	if part1 > n {
		part1 = n
	}
	copy(rb.buffer[rb.write:], data[:part1])

	if n > part1 {
		copy(rb.buffer[0:], data[part1:n])
	}

	rb.write = (rb.write + n) % rb.size
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

	actualRead := n

	part1 := rb.size - rb.read
	if part1 > actualRead {
		part1 = actualRead
	}
	copy(data[:part1], rb.buffer[rb.read:])

	if actualRead > part1 {
		copy(data[part1:actualRead], rb.buffer[0:])
	}

	rb.read = (rb.read + actualRead) % rb.size

	if len(data) > actualRead {
		for i := actualRead; i < len(data); i++ {
			data[i] = 0
		}
	}

	return actualRead
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
