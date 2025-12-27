package main

type RingBuffer struct {
	// buf is the actual buffer.
	buf []float64
	// index points to the value which will be overwritten next.
	index int
}

// NewRingBuffer creates a ring buffer with the given capacity.
// Panics if size <= 0.
func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		panic("ringbuffer size must be positive")
	}
	return &RingBuffer{
		buf: make([]float64, size),
	}
}

func (r *RingBuffer) Add(v float64) {
	r.buf[r.index] = v

	r.index++
	if r.index >= len(r.buf) {
		r.index = 0
	}
}

func (r *RingBuffer) Avg() float64 {
	var sum float64
	for _, v := range r.buf {
		sum += v
	}
	return sum / float64(len(r.buf))
}
