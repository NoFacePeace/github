package main

import (
	"fmt"
	"sync"
)

func main() {
	in := make(chan int)
	out := make(chan int)
	go func() {
		for {
			arr := []int{}
			for i := 0; i < 10; i++ {
				arr = append(arr, <-in)
			}
			for i := 0; i < 10; i++ {
				out <- arr[i]
			}
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(num int) {
			in <- num
			fmt.Println(num, <-out)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
