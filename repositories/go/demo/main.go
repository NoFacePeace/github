package main

import "fmt"

type Test struct {
	m map[string]string
}

func main() {
	t := Test{}
	t.m["test"] = "test"
	fmt.Println(t)
}
