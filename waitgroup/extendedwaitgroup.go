package extendedwaitgroup

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// WaitGroupを拡張した構造体
type ExtendedWaitGroup struct {
	wg      sync.WaitGroup     // WaitGroup
	errChan chan error         // エラー収集用channel
	taskID  int32              // タスクIDを保持するための変数
	running sync.Map           // 実行中のタスクを保持するmap
	ctx     context.Context    // context
	cancel  context.CancelFunc // contextのcancel関数
}

// インスタンスの作成
func NewExtendedWaitGroup() *ExtendedWaitGroup {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExtendedWaitGroup{
		errChan: make(chan error, 1),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// WaitGroupのAddを呼び出す
func (ewg *ExtendedWaitGroup) Add(delta int) {
	ewg.wg.Add(delta)
}

// 指定したタスクIDを削除し、WaitGroupのDoneを呼び出す
func (ewg *ExtendedWaitGroup) Done(id int32) {
	ewg.running.Delete(id)
	ewg.wg.Done()
}

// Goはタスクを非同期に実行し、エラーが発生した場合はエラーchannelに送信する
func (ewg *ExtendedWaitGroup) Go(f func(context.Context) error) {
	id := atomic.AddInt32(&ewg.taskID, 1)
	ewg.Add(1)
	ewg.running.Store(id, true)
	go func() {
		// panicタスクをキャッチしてエラーとして収集
		defer func() {
			if r := recover(); r != nil {
				select {
				case ewg.errChan <- fmt.Errorf("task %d panicked: %v", id, r):
				default:
				}
			}
			ewg.Done(id)
		}()
		if err := f(ewg.ctx); err != nil {
			select {
			case ewg.errChan <- fmt.Errorf("task %d: %w", id, err):
			default:
			}
		}
	}()
}

// 全てのタスクをキャンセルする
func (ewg *ExtendedWaitGroup) Cancel() {
	ewg.cancel()
}

// 全てのタスクが完了するか、エラーが発生するか、タイムアウトするまで待機する
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
	case <-ewg.ctx.Done():
		var runningTasks []int32
		ewg.running.Range(func(key, value interface{}) bool {
			runningTasks = append(runningTasks, key.(int32))
			return true
		})
		return fmt.Errorf("cancelled, still running tasks: %v", runningTasks)
	}
}
