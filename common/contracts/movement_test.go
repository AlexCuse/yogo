package contracts

import (
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPreviousDayMovement(t *testing.T) {
	pd := iex.PreviousDay{
		Symbol: "ALX",
		Date:   iex.Date(time.Now()),
		Open:   123.4,
		Close:  567.8,
		Volume: 37,
	}

	mvmt := PreviousDayMovement(pd)

	require.Equal(t, pd.Symbol, mvmt.Symbol)
	require.Equal(t, pd.Date.String(), mvmt.Date.String())
	require.Equal(t, pd.Open, mvmt.Open)
	require.Equal(t, pd.Close, mvmt.Close)
	require.Equal(t, pd.Volume, mvmt.Volume)
}
