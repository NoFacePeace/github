package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	ch := make(chan http.ResponseWriter, 100)
	go func() {
		for {
			arr := []http.ResponseWriter{}
			for i := 0; i < 3; i++ {
				arr = append(arr, <-ch)
			}
			for _, w := range arr {
				fmt.Println("Hello, world!")
				fmt.Fprintf(w, "Hello, %q", "world!")
			}
		}
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ch <- w
	})
	fmt.Println("start")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
