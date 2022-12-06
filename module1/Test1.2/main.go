/*
基于 Channel 编写一个简单的单线程生产者消费者模型：
队列：
队列长度 10，队列元素类型为 int
生产者：
每 1 秒往队列中放入一个类型为 int 的元素，队列满时生产者可以阻塞
消费者：
每一秒从队列中获取一个元素并打印，队列为空时消费者阻塞
*/
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	messages := make(chan int, 10)
	//设置一个10秒的context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go producer(ctx, messages)
	go producer(ctx, messages)
	go consumer(messages)
	consumer(messages)
	// time.Sleep(11 * time.Second)
}

//producer
func producer(ctx context.Context, ch chan<- int) {
	//定时器 每个1秒想ticker.C中放入数据
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		//select 来轮询
		select {
		case <-ctx.Done():
			fmt.Println("producer done")
			close(ch)
		default:
			fmt.Println("produce message :", 100)
			ch <- 100
		}
	}
}
func consumer(ch <-chan int) {
	for i := range ch {
		fmt.Println("get message:", i)
	}
}
