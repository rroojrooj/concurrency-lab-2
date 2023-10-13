package main

import (
	"fmt"
	"github.com/ChrisGora/semaphore"
	"math/rand"
	"sync"
	"time"
)

// buffer struct represents a circular buffer of integers.
type buffer struct {
	b                 []int // The actual buffer contents.
	size, read, write int   // Size of buffer, read and write indices.
}

// newBuffer initializes and returns a new buffer of the specified size.
func newBuffer(size int) buffer {
	return buffer{
		b:     make([]int, size),
		size:  size,
		read:  0,
		write: 0,
	}
}

// get method removes and returns the next integer from the buffer.
func (buffer *buffer) get() int {
	x := buffer.b[buffer.read]
	fmt.Println("Get\t", x, "\t", buffer)
	buffer.read = (buffer.read + 1) % len(buffer.b) // Move the read pointer in a circular manner.
	return x
}

// put method inserts an integer into the buffer.
func (buffer *buffer) put(x int) {
	buffer.b[buffer.write] = x
	fmt.Println("Put\t", x, "\t", buffer)
	buffer.write = (buffer.write + 1) % len(buffer.b) // Move the write pointer in a circular manner.
}

// producer function simulates the behavior of a producer.
func producer(buffer *buffer, spaceAvailable, workAvailable semaphore.Semaphore, mutex *sync.Mutex, start, delta int) {
	x := start
	for {
		// Wait until there's space available in the buffer.
		spaceAvailable.Wait()

		// Acquire the mutex to ensure exclusive access to the buffer.
		mutex.Lock()

		// Add a value to the buffer.
		buffer.put(x)
		fmt.Println("Put\t", x, "\t", buffer)
		x = x + delta // Adjust the value for the next iteration.

		// Release the mutex.
		mutex.Unlock()

		// Signal that work is available for the consumer.
		workAvailable.Post()

		// Sleep for a random duration between 0ms and 500ms.
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}
}

// consumer function simulates the behavior of a consumer.
func consumer(buffer *buffer, spaceAvailable, workAvailable semaphore.Semaphore, mutex *sync.Mutex) {
	for {
		// Wait until there's work available in the buffer.
		workAvailable.Wait()

		// Acquire the mutex to ensure exclusive access to the buffer.
		mutex.Lock()

		// Retrieve a value from the buffer.
		val := buffer.get()
		fmt.Println("Get\t", val, "\t", buffer)

		// Release the mutex.
		mutex.Unlock()

		// Signal that space is available in the buffer.
		spaceAvailable.Post()

		// Sleep for a random duration between 0ms and 5s.
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
	}
}

// The main function sets up the simulation.
func main() {
	// Create a new buffer of size 5.
	buffer := newBuffer(5)

	// Initialize a mutex for synchronization.
	mutex := &sync.Mutex{}

	// Initialize semaphores.
	// spaceAvailable tracks how many slots are free in the buffer.
	// workAvailable tracks how many slots are filled and ready to be consumed.
	spaceAvailable := semaphore.Init(5, 5)
	workAvailable := semaphore.Init(5, 0)

	// Start the producer and consumer goroutines.
	go producer(&buffer, spaceAvailable, workAvailable, mutex, 1, 1)
	go producer(&buffer, spaceAvailable, workAvailable, mutex, 1000, -1)

	// Start the consumer function (in the main goroutine).
	consumer(&buffer, spaceAvailable, workAvailable, mutex)
}
