package signals_test

import (
	"github.com/alexcuse/yogo/scanner/signals"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheck(t *testing.T) {
	rule := "Quote.Close > .2"

	signal, _ := signals.NewSignal("Test", rule)

	res, err := signal.Check(signals.Target{
		Quote: iex.PreviousDay{
			Close: .6,
		},
		Stats: iex.KeyStats{},
	})

	require.True(t, res)
	require.Nil(t, err)

	res, err = signal.Check(signals.Target{
		Quote: iex.PreviousDay{
			Close: .1,
		},
		Stats: iex.KeyStats{},
	})

	require.False(t, res)
	require.Nil(t, err)
}
