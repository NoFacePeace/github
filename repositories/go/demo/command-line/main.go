package main

import "flag"

var (
	name string
)

func init() {
	flag.StringVar(&name, "name", "name", "name")
	flag.Parse()
}

func main() {
}
