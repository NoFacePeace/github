package io

import (
	"errors"
	"io"
)

func SafeClose(c io.Closer, err error) error {
	if c == nil {
		return err
	}
	cerr := c.Close()
	if cerr == nil {
		return err
	}
	if err == nil {
		return cerr
	}
	return errors.Join(err, cerr)
}
