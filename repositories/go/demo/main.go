package main

import (
	"fmt"
)

func main() {
	arr := []int{1}
	fmt.Println(arr)
	addInt(1, arr)
	fmt.Println(arr)
}

func addInt(num int, arr []int) {
	arr = append(arr, num)
	fmt.Println(arr)
}
