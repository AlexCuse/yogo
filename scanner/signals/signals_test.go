package signals_test

import (
	"github.com/alexcuse/yogo/scanner/signals"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheck(t *testing.T) {
	sig := signals.Signal{
		Name:   "Test",
		Source: "Quote.Close > .2",
	}

	scan, _ := signals.NewScan(sig)

	res, err := scan.Check(signals.Target{
		Quote: iex.PreviousDay{
			Close: .6,
		},
		Stats: iex.KeyStats{},
	})

	require.True(t, res)
	require.Nil(t, err)

	res, err = scan.Check(signals.Target{
		Quote: iex.PreviousDay{
			Close: .1,
		},
		Stats: iex.KeyStats{},
	})

	require.False(t, res)
	require.Nil(t, err)
}
