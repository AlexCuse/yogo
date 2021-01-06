package signals_test

import (
	"github.com/alexcuse/yogo/common/contracts"
	"github.com/alexcuse/yogo/scanner/signals"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheck(t *testing.T) {
	rule := "Close > .2"

	signal, _ := signals.NewSignal("Test", rule)

	res, err := signal.Check(contracts.Movement{
		Close: 0.6,
	})

	require.True(t, res)
	require.Nil(t, err)

	res, err = signal.Check(contracts.Movement{
		Close: 0.1,
	})

	require.False(t, res)
	require.Nil(t, err)
}
