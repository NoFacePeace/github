package grafana

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/NoFacePeace/github/repositories/go/project/quant/indicator"
	"github.com/gin-gonic/gin"
)

type Grafana struct {
	server *http.Server
	cfg    *config
}

func New(opts ...Option) *Grafana {
	cfg := &config{
		port:    80,
		timeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &Grafana{
		cfg: cfg,
	}
}

func (g *Grafana) Start() {
	router := gin.Default()
	router.GET("/", g.Test)
	router.GET("/stock/price", g.GetPrice)
	router.GET("/stock/sma/cross", g.GetSMACross)
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(g.cfg.port),
		Handler: router,
	}
	g.server = srv
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
}

func (g *Grafana) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), g.cfg.timeout)
	defer cancel()
	if err := g.server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func (g *Grafana) Test(c *gin.Context) {
	c.String(http.StatusOK, "Welcome Gin Server")
}

func (g *Grafana) GetPrice(c *gin.Context) {
	code := c.Query("code")
	ps, err := indicator.AllPrice(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ps)
}

func (g *Grafana) GetSMACross(c *gin.Context) {
	code := c.Query("code")
	ps, err := indicator.AllPrice(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	s, _ := strconv.Atoi(c.Query("short"))
	l, _ := strconv.Atoi(c.Query("long"))
	short := indicator.SMA(ps, s)
	long := indicator.SMA(ps, l)
	cross := indicator.GoldenCross(ps, short, long)
	buy := []indicator.Point{}
	sell := []indicator.Point{}
	for i := 0; i < len(cross); i++ {
		if i%2 == 0 {
			buy = append(buy, cross[i])
		} else {
			sell = append(sell, cross[i])
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"price": ps,
		"short": short,
		"long":  long,
		"buy":   buy,
		"sell":  sell,
	})
}

type config struct {
	port    int
	timeout time.Duration
}
type Option interface {
	apply(*config)
}
type portOption int

func (p portOption) apply(opts *config) {
	opts.port = int(p)
}

func WithPort(port int) Option {
	return portOption(port)
}
