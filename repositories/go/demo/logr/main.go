package main

import (
	otellogr "go.opentelemetry.io/contrib/bridges/otellogr"
)

func main() {
	otellogr.NewLogSink("otel")
	// logr.New()
	// otel.SetLogger(logr.Logger{})
	// global.GetLoggerProvider().Logger()
}
