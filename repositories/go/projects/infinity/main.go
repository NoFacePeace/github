package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/external/tencent/finance"
	"github.com/NoFacePeace/github/repositories/go/external/tencent/gu"
	"github.com/NoFacePeace/github/repositories/go/utils/signal"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := signal.SetupSignalHandler()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/tencent/gu/hkminute", func(c *gin.Context) {
		code := c.Query("code")
		ps, err := gu.HKMinute(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ps)
	})
	r.GET("/tencent/finance/plates", func(c *gin.Context) {
		plates, err := finance.ListPlates(finance.PlateTypeHY2)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("finance list plates error: [%w]", err)})
			return
		}
		c.JSON(http.StatusOK, plates)
	})
	r.GET("/tencent/finance/kline/:name", func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": errors.New("bad request").Error()})
			return
		}
		plates, err := finance.ListPlates(finance.PlateTypeHY2)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": fmt.Errorf("finance list plates error: [%w]", err).Error(),
			})
			return
		}
		code := ""
		for _, plate := range plates {
			if plate.Name == name {
				code = plate.Code
				break
			}
		}
		ps, err := finance.GetKline(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Errorf("finance get all kline error: [%w]", err).Error()})
			return
		}
		from := c.Query("from")
		if from != "" {
			t, _ := strconv.Atoi(from)
			d := time.UnixMilli(int64(t))
			for i := 0; i < len(ps); i++ {
				if ps[i].Date.After(d) {
					ps = ps[i:]
					break
				}
			}
		}
		c.JSON(http.StatusOK, ps)
	})
	r.GET("/tencent/finance/change/:name", func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": errors.New("bad request").Error()})
			return
		}
		plates, err := finance.ListPlates(finance.PlateTypeHY2)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": fmt.Errorf("finance list plates error: [%w]", err).Error(),
			})
			return
		}
		code := ""
		for _, plate := range plates {
			if plate.Name == name {
				code = plate.Code
				break
			}
		}
		ps, err := finance.GetKline(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Errorf("finance get all kline error: [%w]", err).Error()})
			return
		}
		arr := []map[string]any{}
		last := ps[len(ps)-1]
		for _, p := range ps {
			m := map[string]any{}
			m["date"] = p.Date
			m["percent"] = (p.Last - last.Last) / p.Last * 100
			arr = append(arr, m)
		}
		c.JSON(http.StatusOK, arr)
	})
	r.GET("/tencent/finance/low", func(c *gin.Context) {
		plates, err := finance.ListPlates(finance.PlateTypeHY2, finance.DirectOptionUp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": fmt.Errorf("finance list plates error: [%w]", err).Error(),
			})
			return
		}
		ret := []map[string]any{}
		for _, plate := range plates {
			ps, err := finance.GetKline(plate.Code)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": fmt.Errorf("finance get kline error: [%w]", err).Error(),
				})
				return
			}
			if len(ps) == 0 {
				continue
			}
			n := len(ps)
			last := ps[n-1]
			mn := last.Last
			mx := last.Last
			date := last.Date
			for i := n - 1; i >= 0; i-- {
				if ps[i].Last < mn {
					break
				}
				if ps[i].Last >= mx {
					mx = ps[i].Last
					date = ps[i].Date
				}
			}
			percent := (mn - mx) / mx * 100
			if percent == 0 {
				break
			}
			tmp := map[string]any{
				"name":    plate.Name,
				"percent": (mn - mx) / mx * 100,
				"date":    date,
			}
			ret = append(ret, tmp)

		}
		c.JSON(http.StatusOK, ret)
	})
	r.GET("/tencent/finance/plates/volume/low", func(c *gin.Context) {
		plates, err := finance.ListPlates(finance.PlateTypeHY2, finance.DirectOptionUp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": fmt.Errorf("finance list plates error: [%w]", err).Error(),
			})
			return
		}
		ret := []map[string]any{}
		for _, plate := range plates {
			ps, err := finance.GetKline(plate.Code)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": fmt.Errorf("finance get kline error: [%w]", err).Error(),
				})
				return
			}
			if len(ps) == 0 {
				continue
			}
			from := c.Query("from")
			if from != "" {
				t, _ := strconv.Atoi(from)
				d := time.UnixMilli(int64(t))
				for i := 0; i < len(ps); i++ {
					if ps[i].Date.After(d) {
						ps = ps[i:]
						break
					}
				}
			}
			n := len(ps)
			if n == 0 {
				continue
			}
			mn := ps[0]
			for i := 1; i < n; i++ {
				if ps[i].Volume <= mn.Volume {
					mn = ps[i]
				}
			}
			tmp := map[string]any{
				"name":   plate.Name,
				"volume": mn.Volume,
				"date":   mn.Date,
			}
			ret = append(ret, tmp)

		}
		c.JSON(http.StatusOK, ret)
	})
	r.GET("/tencent/finance/plates/volume/high", func(c *gin.Context) {
		plates, err := finance.ListPlates(finance.PlateTypeHY2, finance.DirectOptionUp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": fmt.Errorf("finance list plates error: [%w]", err).Error(),
			})
			return
		}
		ret := []map[string]any{}
		for _, plate := range plates {
			ps, err := finance.GetKline(plate.Code)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": fmt.Errorf("finance get kline error: [%w]", err).Error(),
				})
				return
			}
			if len(ps) == 0 {
				continue
			}
			from := c.Query("from")
			if from != "" {
				t, _ := strconv.Atoi(from)
				d := time.UnixMilli(int64(t))
				for i := 0; i < len(ps); i++ {
					if ps[i].Date.After(d) {
						ps = ps[i:]
						break
					}
				}
			}
			n := len(ps)
			if n == 0 {
				continue
			}
			mx := ps[0]
			for i := 1; i < n; i++ {
				if ps[i].Volume >= mx.Volume {
					mx = ps[i]
				}
			}
			tmp := map[string]any{
				"name":   plate.Name,
				"volume": mx.Volume,
				"date":   mx.Date,
			}
			ret = append(ret, tmp)

		}
		c.JSON(http.StatusOK, ret)
	})
	srv := &http.Server{
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("http server listen and server error", "error", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown error", "error", err)
		os.Exit(1)
	}
}
