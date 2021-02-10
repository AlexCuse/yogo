package scanner

import (
	"testing"

	iex "github.com/goinvest/iexcloud/v2"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	scan := Scan{
		Name:   "Test",
		Source: "Quote.Close > .2",
	}

	require.NoError(t, scan.Compile())

	res, err := scan.Check(Target{
		Quote: iex.PreviousDay{
			Close: .6,
		},
		Stats: iex.KeyStats{},
	})

	require.True(t, res)
	require.Nil(t, err)

	res, err = scan.Check(Target{
		Quote: iex.PreviousDay{
			Close: .1,
		},
		Stats: iex.KeyStats{},
	})

	require.False(t, res)
	require.Nil(t, err)
}
