package server_test

import (
	"github.com/alexcuse/yogo/scanner/server"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheck(t *testing.T) {
	scan := server.Scan{
		Name:   "Test",
		Source: "Quote.Close > .2",
	}

	require.NoError(t, scan.Compile())

	res, err := scan.Check(server.Target{
		Quote: iex.PreviousDay{
			Close: .6,
		},
		Stats: iex.KeyStats{},
	})

	require.True(t, res)
	require.Nil(t, err)

	res, err = scan.Check(server.Target{
		Quote: iex.PreviousDay{
			Close: .1,
		},
		Stats: iex.KeyStats{},
	})

	require.False(t, res)
	require.Nil(t, err)
}
