package main

import (
	"fmt"

	"github.com/apache/pulsar-client-go/pulsaradmin"
)

func main() {

}

func pulsaradminDemo() {
	cfg := &pulsaradmin.Config{}
	admin, err := pulsaradmin.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	tenants, _ := admin.Tenants().List()
	fmt.Println(tenants)
}
