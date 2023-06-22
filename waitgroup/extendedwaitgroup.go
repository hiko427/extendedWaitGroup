package extendedwaitgroup

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type ExtendedWaitGroup struct {
	wg      sync.WaitGroup
	errChan chan error
	taskID  int32
	running sync.Map
}

func NewExtendedWaitGroup() *ExtendedWaitGroup {
	return &ExtendedWaitGroup{
		errChan: make(chan error, 1),
	}
}

func (ewg *ExtendedWaitGroup) Add(delta int) {
	ewg.wg.Add(delta)
}

func (ewg *ExtendedWaitGroup) Done(id int32) {
	ewg.running.Delete(id)
	ewg.wg.Done()
}

func (ewg *ExtendedWaitGroup) Go(f func() error) {
	id := atomic.AddInt32(&ewg.taskID, 1)
	ewg.Add(1)
	ewg.running.Store(id, true)
	go func() {
		defer ewg.Done(id)
		if err := f(); err != nil {
			select {
			case ewg.errChan <- fmt.Errorf("task %d: %w", id, err):
			default:
			}
		}
	}()
}

func (ewg *ExtendedWaitGroup) Wait(timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		defer close(c)
		ewg.wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case err := <-ewg.errChan:
		return err
	case <-time.After(timeout):
		var runningTasks []int32
		ewg.running.Range(func(key, value interface{}) bool {
			runningTasks = append(runningTasks, key.(int32))
			return true
		})
		return fmt.Errorf("timeout, still running tasks: %v", runningTasks)
	}
}
