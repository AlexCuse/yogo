package social

import "context"

func NewCalculator(wl *WatchList, symbols chan string) Calculator {
	return Calculator{
		wl,
		symbols,
	}
}

type Calculator struct {
	wl      *WatchList
	symbols chan string
}

func (c *Calculator) Start(ctx context.Context) error {
	watchList, err := c.wl.Symbols(ctx)
	if err != nil {
		return err
	}

	for _, w := range watchList {
		c.symbols <- w.Symbol
	}

	return nil
}
