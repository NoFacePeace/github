package gin

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func Shutdown(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown error: [%w]", err)
	}
	return nil
}
