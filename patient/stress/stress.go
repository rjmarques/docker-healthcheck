package stress

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

type CPUBurner struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup

	numThreads int
}

func NewCPUBurner(numThreads int) *CPUBurner {
	return &CPUBurner{
		numThreads: numThreads,
	}
}

func (b *CPUBurner) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		fmt.Println("already burning")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.wg.Add(b.numThreads)
	for i := 0; i < b.numThreads; i++ {
		go cpuBurn(i, ctx, &b.wg)
	}
	b.cancel = cancel
}

func (b *CPUBurner) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel == nil {
		fmt.Println("already stopped")
		return
	}

	b.cancel()     // signal go routines to exit
	b.wg.Wait()    // and wait for them to do so
	b.cancel = nil // mark the process a stopped
}

func cpuBurn(id int, ctx context.Context, wg *sync.WaitGroup) {
	fmt.Printf("starting cpu burn, routine %d\n", id)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("stopping cpu burn, routine %d\n", id)
			wg.Done()
			return
		default:
			for i := 0; i < 134217728; i++ {
				// burn, baby, burn! disco inferno!
			}
			runtime.Gosched()
		}
	}
}
