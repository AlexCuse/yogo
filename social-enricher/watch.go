package social

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func NewWatchList(c *resty.Client) WatchList {
	return WatchList{
		c: c,
	}
}

type WatchList struct {
	c *resty.Client
}

func (w *WatchList) Symbols(ctx context.Context) ([]Watch, error) {
	wl := make([]Watch, 0)

	r, err := w.c.R().
		SetContext(ctx).
		SetResult(&wl).
		Get("/api/watch")
	if err != nil {
		return wl, fmt.Errorf("fetching watchlist:%w", err)
	}
	if !r.IsSuccess() {
		return wl, fmt.Errorf("fetching watchlist, status code %v", r.Status())
	}

	return wl, nil

}

type Watch struct {
	Symbol string `json:"symbol,omitempty"`
}
