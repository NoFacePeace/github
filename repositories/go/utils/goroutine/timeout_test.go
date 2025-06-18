package goroutine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGoWithTimeout(t *testing.T) {
	// timeout
	err := GoWithTimeout(func() error {
		time.Sleep(time.Second)
		return nil
	}, time.Millisecond)
	require.ErrorIs(t, err, errTimeout)

	// success
	err = GoWithTimeout(func() error {
		return nil
	}, time.Millisecond)
	require.Nil(t, err)
}
