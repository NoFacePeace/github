package main

import (
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	m := &autocert.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		Prompt:     autocert.AcceptTOS,
		Email:      "example@example.org",
		HostPolicy: autocert.HostWhitelist(),
	}
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
	}
	s.ListenAndServeTLS("", "")
}
