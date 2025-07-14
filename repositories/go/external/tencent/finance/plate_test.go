package finance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListPlates(t *testing.T) {
	ps, err := ListPlates(PlateTypeHY2)
	require.Nil(t, err)
	require.NotEqual(t, 0, len(ps))
}
