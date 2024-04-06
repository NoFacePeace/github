package main

import (
	"fmt"
	"sync"
)

func main() {
	for i := 0; i < 10; i++ {
		var once sync.Once
		once.Do(func() {
			fmt.Println("hello world")
		})
	}
}
