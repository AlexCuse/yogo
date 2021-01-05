package signals_test

import (
	"github.com/alexcuse/yogo/scanner/signals"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheck(t *testing.T){
	rule := "ChangePercent > .2"

	signal, _ := signals.NewSignal("Test", rule)

	res := signal.Check(iex.Quote{
		ChangePercent: 0.6,
	})

	require.True(t, res)

	res = signal.Check(iex.Quote{
		ChangePercent: 0.1,
	})

	require.False(t, res)
}
