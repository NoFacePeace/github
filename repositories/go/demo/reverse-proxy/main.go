// reverse-proxy 演示如何使用 httputil.ReverseProxy 创建一个简单的反向代理。
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "反向代理的监听地址")
	targetAddr := flag.String("target", "http://localhost:8081", "上游服务地址")
	flag.Parse()

	target, err := url.Parse(*targetAddr)
	if err != nil {
		log.Fatalf("parse upstream target: %v", err)
	}
	if target.Scheme == "" || target.Host == "" {
		log.Fatal("upstream target must include a scheme and host, for example http://localhost:8081")
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.IdleConnTimeout = 90 * time.Second
	proxy.Transport = transport
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy %s %s to %s failed: %v", r.Method, r.URL, target, err)
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("proxying %s %s -> %s", r.Method, r.URL, target)
		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:              *listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("reverse proxy listening on %s; upstream: %s", *listenAddr, target)
	log.Fatal(server.ListenAndServe())
}
