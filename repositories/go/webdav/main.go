package main

import (
	"log"
	"log/slog"
	"net/http"
	"os/exec"

	"github.com/NoFacePeace/github/repositories/go/util/config"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/webdav"
)

type Config struct {
	Port     string
	User     string
	Password string
	Backup   []string
	Path     string
}

func main() {
	initLog()

	// init config
	cfg := &Config{}
	if err := config.ReadYamlFile("config.yaml", cfg); err != nil {
		slog.Error(err.Error())
		return
	}

	// 同步
	c := cron.New()
	c.AddFunc("@every 60s", func() {
		slog.Info("Starting rsync")
		for _, v := range cfg.Backup {
			cmd := exec.Command("rsync", "-av", cfg.Path+"/", "root@"+v+":"+cfg.Path)
			slog.Info("Starting rsync " + v)
			if err := cmd.Run(); err != nil {
				slog.Error(err.Error())
			}
			slog.Info("Finish rsync " + v)
		}
		slog.Info("Finish rsync")
	})
	c.Start()

	// webdav
	fs := &webdav.Handler{
		FileSystem: webdav.Dir(cfg.Path),
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
		if user != cfg.User && password != cfg.Password {
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
