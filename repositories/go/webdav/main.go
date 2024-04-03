package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/NoFacePeace/github/repositories/go/util/config"
	"golang.org/x/net/webdav"
)

type Config struct {
	Port string
}

var (
	// flags
	u string
	p string
	d string
)

func main() {
	initFlags()
	initLog()

	// init config
	cfg := &Config{}
	if err := config.ReadYamlFile("config.yaml", cfg); err != nil {
		slog.Error(err.Error())
	}
	fs := &webdav.Handler{
		FileSystem: webdav.Dir("."),
		LockSystem: webdav.NewMemLS(),
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok {
			slog.Warn("Basic Auth Required")
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if user != u && password != p {
			slog.Warn("Basic Auth Failed")
			w.WriteHeader(http.StatusUnauthorized)
		}
		fs.ServeHTTP(w, r)
	})
	slog.Info("starting server")
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		slog.Error(err.Error())
	}
}

func initLog() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func initFlags() {
	flag.StringVar(&u, "u", "user", "user name")
	flag.StringVar(&p, "p", "password", "password")
	flag.StringVar(&d, "d", ".", "data")
	flag.Parse()
}
