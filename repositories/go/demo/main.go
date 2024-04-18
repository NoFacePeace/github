package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func main() {
	_, err := os.ReadFile("hhh")
	if err != nil {
		fmt.Printf("%+v\n", err)
		err = errors.Wrap(err, "hhhh")
		fmt.Printf("%+v\n", err)
	}
}
