package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/NoFacePeace/github/repositories/go-tpl/config"
	"golang.org/x/net/webdav"
)

type Config struct {
	Port int
}

func main() {
	cfg := &Config{}
	err := config.ReadYamlFile("config.yaml", cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("start")
	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), &webdav.Handler{
		FileSystem: webdav.Dir("."),
		LockSystem: webdav.NewMemLS(),
	})
}
