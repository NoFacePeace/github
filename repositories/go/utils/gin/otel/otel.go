package otel

import (
	"errors"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var (
	logger = otelslog.NewLogger("gin", otelslog.WithSource(true))
)

func InjectPropagatorToResponseHeader(c *gin.Context) {
	propagators := otel.GetTextMapPropagator()
	propagators.Inject(c.Request.Context(), propagation.HeaderCarrier(c.Writer.Header()))
	c.Next()
}

func DebugPrintRouteFunc(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	logger.Debug("gin debug", "method", httpMethod, "path", absolutePath, "handler name", handlerName, "num handlers", nuHandlers)
}

// https://github.com/gin-gonic/gin/blob/master/logger.go#L162
func Logger(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	c.Next()
	param := gin.LogFormatterParams{}
	param.TimeStamp = time.Now()
	param.Latency = param.TimeStamp.Sub(start)
	param.ClientIP = c.ClientIP()
	param.Method = c.Request.Method
	param.StatusCode = c.Writer.Status()
	param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
	param.BodySize = c.Writer.Size()
	if raw != "" {
		path = path + "?" + raw
	}
	param.Path = path
	logger.InfoContext(c.Request.Context(), "gin", "timestamp", param.TimeStamp, "status code", param.StatusCode, "latency", param.Latency, "client ip", param.ClientIP, "method", param.Method, "path", param.Path, "error message", param.ErrorMessage)
}

// https://github.com/gin-gonic/gin/blob/master/recovery.go#L53
func Recovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			var brokenPipe bool
			if ne, ok := err.(*net.OpError); ok {
				var se *os.SyscallError
				if errors.As(ne, &se) {
					seStr := strings.ToLower(se.Error())
					if strings.Contains(seStr, "broken pipe") || strings.Contains(seStr, "connection reset by peer") {
						brokenPipe = true
					}
				}
			}
			if e, ok := err.(error); ok && errors.Is(e, http.ErrAbortHandler) {
				brokenPipe = true
			}
			if brokenPipe {
				logger.ErrorContext(c.Request.Context(), "recovery", "error", err, "request", c.Request)
			} else {
				logger.ErrorContext(c.Request.Context(), "recovery", "error", err, "request", c.Request, "stack", string(debug.Stack()))
			}
			if brokenPipe {
				c.Error(err.(error))
				c.Abort()
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}()
	c.Next()
}
