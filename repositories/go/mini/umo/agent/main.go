package main

import "log/slog"

func main() {

	_, clear, err := setupApp()
	if err != nil {
		slog.Error("setupApp failed", "err", err)
		return
	}
	defer clear()
}

type App struct {
}

func setupApp() (*App, func(), error) {

	return &App{}, func() {}, nil
}
