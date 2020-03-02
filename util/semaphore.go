package util

import "sync"

// Pool struct
// https://www.golangtc.com/t/559e97d6b09ecc22f6000053
// control the number of Goroutine concurrency by waitGroup + channel
type Pool struct {
	queue chan int
	Wg    *sync.WaitGroup
	Size  int // pool size
}

// NewPool function,create a shared lock with parallelism shared credentials
func NewPool(cap, total int) *Pool {
	if cap < 1 {
		cap = 1
	}
	p := &Pool{
		queue: make(chan int, cap),
		Wg:    new(sync.WaitGroup),
	}
	p.Wg.Add(total)
	p.Size = 0
	return p
}

// Acquire function,get a voucher
func (p *Pool) Acquire() {
	p.queue <- 1
	p.Size++
}

// Release function,release a credential
func (p *Pool) Release() {
	<-p.queue
	p.Wg.Done()
	p.Size--
}

// TryAcquire function,try acquire credential
func (p *Pool) TryAcquire() bool {
	select {
	case p.queue <- 1:
		p.Size++
		return true
	default:
		return false

	}
}

// AvailableCredential function
func (p *Pool) AvailableCredential() int {
	return cap(p.queue) - p.Size
}
