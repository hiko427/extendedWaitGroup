package main

import (
	"fmt"
	"time"

	ewglib "github.com/knowledge-work-internship-2023-6-teamA/extendedwaitgroup/waitgroup"
)

func main() {
	ewg := ewglib.NewExtendedWaitGroup()
	for i := 0; i < 10; i++ {
		ewg.Go(func() error {
			time.Sleep(2 * time.Second)
			return nil
		})
	}
	err := ewg.Wait(1 * time.Second)
	if err != nil {
		fmt.Println("エラーが発生しました:", err)
	} else {
		fmt.Println("全てのタスクが正常に終了しました。")
	}
}
