package threadpool

import (
	"sync"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
)

var logger = lorg.NewDiscarder()

// Processable interface must be implemented for processing task by thread
// pool.
type Processable interface {
	Process()
}

type dummyProc struct {
	fn func()
}

func (proc dummyProc) Process() {
	proc.fn()
}

func ProcFunc(fn func()) Processable {
	return dummyProc{fn: fn}
}

// ThreadPool have internal queue that actually is a channel.
type ThreadPool struct {
	queue    chan Processable
	threads  *sync.WaitGroup
	capacity int64
}

// SetLogger that will be used for logging messages.
func SetLogger(log lorg.Logger) {
	logger = log
}

// New creates completely ready thread pool that can spawn threads.
func New(queueSize int) *ThreadPool {
	pool := &ThreadPool{
		queue:   make(chan Processable, queueSize),
		threads: &sync.WaitGroup{},
	}

	return pool
}

// Spawn thread pool for processing tasks.
func (pool *ThreadPool) Spawn(count int) {
	capacity := pool.capacity

	for thread := 1; thread <= count; thread++ {
		atomic.AddInt64(&pool.capacity, 1)
		pool.threads.Add(1)

		go func(id int) {
			defer func() {
				pool.threads.Done()
				atomic.AddInt64(&pool.capacity, -1)

				logger.Debugf(
					"[thread:%d] thread has been released",
					pool.capacity,
				)
			}()

			for task := range pool.queue {
				logger.Debugf(
					"[thread:%d] processing task %q",
					id, task,
				)

				task.Process()
			}
		}(int(capacity) + thread)
	}

	atomic.AddInt64(&pool.capacity, int64(count))
}

// Push task into thread pool queue.
func (pool *ThreadPool) Push(task Processable) {
	pool.queue <- task
}

// Release thread pool queue and wait all threads.
func (pool *ThreadPool) Release() {
	close(pool.queue)
	pool.threads.Wait()
}

// GetCapacity returns current thread pool capacity.
func (pool *ThreadPool) GetCapacity() int {
	return int(pool.capacity)
}
