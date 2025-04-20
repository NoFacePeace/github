package main

import "fmt"

func main() {
	arr := [5]int{1, 2, 3, 4, 5}
	arr1 := arr[:3]
	arr1[0] = 0
	fmt.Println(arr)
	fmt.Println(arr1)
	arr1 = append(arr1, 5, 5)

	fmt.Println(arr)
	fmt.Println(arr1)
	arr1[1] = 0
	fmt.Println(arr)
	fmt.Println(arr1)
}
