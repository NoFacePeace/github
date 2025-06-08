package io

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	errTest  = errors.New("test")
	errClose = errors.New("close")
)

func TestSafeClose(t *testing.T) {
	err := SafeClose(nil, errTest)
	require.Equal(t, errTest, err)

	err = SafeClose(&testCloser{}, nil)
	require.Equal(t, errClose, err)

	err = SafeClose(&testCloser{}, errTest)
	require.Equal(t, errors.Join(errTest, errClose), err)
}

type testCloser struct{}

func (t *testCloser) Close() error {
	return errClose
}
