package main

import (
	"context"
	"fmt"
	"time"
)

func main()  {
	ctx, cancel := context.WithCancel(context.Background())
	go watch(ctx, "监控1")
	go watch(ctx, "监控2")
	go watch(ctx, "监控3")

	time.Sleep(10 * time.Second)
	fmt.Println("可以了，通知监控停止")
	cancel()
	time.Sleep(5 * time.Second)
}


func watch(ctx context.Context, str string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println(str, "监控退出，停止了...")
			return
		default:
			fmt.Println(str, "goroutine监控中...")
			time.Sleep(2 * time.Second)
		}
	}
}