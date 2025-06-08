package main

import "fmt"

func main() {
	num := ret()
	fmt.Println(num)
}

func ret() (num int) {
	num = 1
	return 2
}
