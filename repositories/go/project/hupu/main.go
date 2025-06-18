package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			slog.Info("program stop")
			stop()
			return
		case <-ticker.C:
			slog.Info("crawl start")
			crawlPost()
			slog.Info("crawl stop")
		}
	}
}

func crawlPost() {
	url := "https://bbs.hupu.com/love-postdate"
	resp, err := http.Get(url)
	if err != nil {
		slog.Info("http get error", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("status code %d", resp.StatusCode))
		return
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.Error("html parse error", "error", err)
		return
	}
	cnt := 0
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			href := ""
			match := false
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = a.Val
				}
				if a.Key == "class" && a.Val == "p-title" {
					match = true
				}
			}
			if match {
				cnt++
				arr := strings.Split(href, "/")
				arr = strings.Split(arr[1], ".")
				dir := "image/" + arr[0]
				if err := os.MkdirAll(dir, 0755); err != nil {
					slog.Error("os mk dir error", "error", err)
					return
				}
				fmt.Println("https://bbs.hupu.com" + href)
				crawlIamge(dir, "https://bbs.hupu.com"+href)
				if cnt == 3 {
					break
				}
			}
		}
	}
}

func crawlIamge(dir, url string) {
	resp, err := http.Get(url)
	if err != nil {
		slog.Info("http get error", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("status code %d", resp.StatusCode))
		return
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.Error("html parse error", "error", err)
		return
	}
	cnt := 0
	for n := range doc.Descendants() {
		if n.DataAtom == atom.Img {
			for _, a := range n.Attr {
				if a.Key == "data-origin" {
					fmt.Println(a.Val)
					downloadImage(a.Val, dir+"/"+strconv.Itoa(cnt)+".jpg")
					cnt++
				}
			}
		}
	}
}

func downloadImage(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create file
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy image data to file
	_, err = out.ReadFrom(resp.Body)
	return err
}
