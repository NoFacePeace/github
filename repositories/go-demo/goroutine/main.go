package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println("hello world")
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		time.Sleep(time.Second * 10)
		log.Fatal("hh")
	}()
	fmt.Println("starting")
	WaitOSInterrupt()
	fmt.Println("exit")
}

func WaitOSInterrupt() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
