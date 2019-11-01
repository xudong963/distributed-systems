package main

import (
	"context"
	"fmt"
	"time"
)

// 控制并发的两种经典的方式, context 和 waitGroup
// 一个网络请求 Request
// 每个 Request 都需要开启一个 goroutine 做一些事情
// 这些 goroutine 又可能会开启其它的 goroutine
// 所以我们需要一种可以跟踪 goroutine 的方案，才可以达到控制它们的目的
// 这就是Go语言为我们提供的 Context,称之为上下文非常贴切,它就是 goroutine 的上下文

func main() {
	// context.Background() 返回一个空的Context, 这个空的Context一般用于整个树的根节点
	// 然后用context.WithCancel(parent),创建一个可取消的子Context,作为参数传给goroutine
	// 这个子Context就开始跟踪这个goroutine了
	// 调用cancel函数就可以让ctx.Done()收到值, 然后结束
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("监控退出，停止了...")
				return
			default:
				fmt.Println("goroutine监控中...")
				time.Sleep(2 * time.Second)
			}
		}
	}(ctx)

	time.Sleep(10 * time.Second)
	fmt.Println("可以了，通知监控停止")

	cancel()

	//为了检测监控过是否停止，如果没有监控输出，就表示停止了
	time.Sleep(5 * time.Second)
}
