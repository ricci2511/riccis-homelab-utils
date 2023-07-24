package dupescout

// Custom simple semaphore implementation.
type semaphore chan bool

func newSemaphore(size int) semaphore {
	return make(chan bool, size)
}

func (s *semaphore) acquire() {
	*s <- true
}

func (s *semaphore) release() {
	func() { <-*s }()
}
