package main

import (
	"errors"
	"fmt"
)

func main() {
	err := call()
	fmt.Println(err)
}

func call() (err error) {
	defer func() {
		fmt.Println(err)
	}()
	ok, geterr := get()
	fmt.Println(ok)
	return geterr
}

func get() (bool, error) {
	return true, errors.New("test error")
}
