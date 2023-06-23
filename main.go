package main

import (
	"context"
	"fmt"
	"time"

	ewglib "github.com/knowledge-work-internship-2023-6-teamA/extendedwaitgroup/waitgroup"
)

func main() {
	ewg := ewglib.NewExtendedWaitGroup()
	for i := 0; i < 10; i++ {
		i := i
		ewg.Go(func(ctx context.Context) error {
			fmt.Println(i)
			time.Sleep(6 * time.Second)
			return nil
		})
	}
	// time.AfterFunc(1*time.Second, func() {
	// 	ewg.Cancel()
	// })
	err := ewg.Wait(4 * time.Second)
	if err != nil {
		fmt.Println("エラーが発生しました:", err)
	} else {
		fmt.Println("全てのタスクが正常に終了しました。")
	}
}
