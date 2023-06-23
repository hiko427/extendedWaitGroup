package extendedwaitgroup

import (
	"context"
	"errors"
	"testing"
	"time"
)

// 正しく終了や並行処理なのかを判定するテスト
func TestExtendedWaitGroupWithSuccessfulTasks(t *testing.T) {
	ewg := NewExtendedWaitGroup()
	for i := 0; i < 10; i++ {
		ewg.Go(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}
	if err := ewg.Wait(1 * time.Second); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// エラーテスト
func TestExtendedWaitGroupWithErrorTask(t *testing.T) {
	ewg := NewExtendedWaitGroup()
	ewg.Go(func(ctx context.Context) error {
		return errors.New("error")
	})
	if err := ewg.Wait(1 * time.Second); err == nil || err.Error() != "task 1: error" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// タイムアウトテスト
func TestExtendedWaitGroupWithTimeout(t *testing.T) {
	ewg := NewExtendedWaitGroup()
	for i := 0; i < 10; i++ {
		ewg.Go(func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		})
	}
	err := ewg.Wait(1 * time.Second)
	if err == nil {
		t.Error("Unexpected error: nil")
	} else if err.Error()[:7] != "timeout" {
		t.Errorf("Unexpected error: %v", err)
	}
	if runningTasks := ewg.RunningTasks(); runningTasks != 10 {
		t.Errorf("Wrong task number. Expected: 10, but: %v", runningTasks)
	}
}

// 実行中のタスク数
func (ewg *ExtendedWaitGroup) RunningTasks() int {
	var count int
	ewg.running.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// キャンセルテスト
func TestExtendedWaitGroupWithCancel(t *testing.T) {
	ewg := NewExtendedWaitGroup()
	ewg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-time.After(10 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	time.AfterFunc(1*time.Second, func() {
		ewg.Cancel()
	})
	err := ewg.Wait(2 * time.Second)
	if err == nil {
		t.Error("Expected error, but got none")
	}
	if runningTasks := ewg.RunningTasks(); runningTasks != 1 {
		t.Errorf("Wrong task number. Expected: 1, but got: %v", runningTasks)
	}
}

// Panic テスト
func TestExtendedWaitGroupWithPanicTask(t *testing.T) {
	ewg := NewExtendedWaitGroup()
	ewg.Go(func(ctx context.Context) error {
		panic("This is a test panic")
	})
	err := ewg.Wait(1 * time.Second)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else if err.Error() != "task 1 panicked: This is a test panic" {
		t.Errorf("Unexpected error: %v", err)
	}
	if runningTasks := ewg.RunningTasks(); runningTasks != 0 {
		t.Errorf("Wrong task number. Expected: 0, but got: %v", runningTasks)
	}
}
