package stocktwits

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

func NewSymbolApi(c *resty.Client) SymbolApi {
	return SymbolApi{
		c: c,
	}
}

type SymbolApi struct {
	c *resty.Client
}

func (s *SymbolApi) Get(ctx context.Context, symbol string) (Twits, error) {
	twits := Twits{}

	r, err := s.c.R().
		SetContext(ctx).
		SetResult(&twits).
		Get(fmt.Sprintf("/api/2/streams/symbol/%v.json", symbol))
	if err != nil {
		return twits, fmt.Errorf("fetching stocktwits for %v:%w", symbol, err)
	}
	if !r.IsSuccess() {
		return twits, fmt.Errorf("fetching stocktwits for %v, status code %v", symbol, r.Status())
	}

	return twits, nil
}

func (s *SymbolApi) Stream(ctx context.Context, symbols <-chan string, twits chan Twits) {
	for {
		select {
		case <-ctx.Done():
			return
		case sym, ok := <-symbols:
			if !ok {
				return
			}
			log := zerolog.Ctx(ctx).With().Str("symbol", sym).Logger()
			ctx := log.WithContext(ctx)

			r, err := s.Get(ctx, sym)
			if err != nil {
				log.Err(err).Msg("")
			}

			twits <- r

			log.Debug().Msg("new twits")
		}
	}
}

type Twits struct {
	Symbol struct {
		Symbol string `json:"symbol,omitempty"`
	} `json:"symbol,omitempty"`
	Cursor struct {
		Since int64 `json:"since,omitempty"`
	} `json:"cursor,omitempty"`
	Messages []struct {
		ID        int       `json:"id,omitempty"`
		Body      string    `json:"body,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		User      struct {
			ID                   int           `json:"id,omitempty"`
			Username             string        `json:"username,omitempty"`
			Name                 string        `json:"name,omitempty"`
			AvatarURL            string        `json:"avatar_url,omitempty"`
			AvatarURLSsl         string        `json:"avatar_url_ssl,omitempty"`
			JoinDate             string        `json:"join_date,omitempty"`
			Official             bool          `json:"official,omitempty"`
			Identity             string        `json:"identity,omitempty"`
			Classification       []interface{} `json:"classification,omitempty"`
			Followers            int           `json:"followers,omitempty"`
			Following            int           `json:"following,omitempty"`
			Ideas                int           `json:"ideas,omitempty"`
			WatchlistStocksCount int           `json:"watchlist_stocks_count,omitempty"`
			LikeCount            int           `json:"like_count,omitempty"`
			PlusTier             string        `json:"plus_tier,omitempty"`
			PremiumRoom          string        `json:"premium_room,omitempty"`
			TradeApp             bool          `json:"trade_app,omitempty"`
		} `json:"user,omitempty"`
		Source struct {
			ID    int    `json:"id,omitempty"`
			Title string `json:"title,omitempty"`
			URL   string `json:"url,omitempty"`
		} `json:"source,omitempty"`
		Symbols []struct {
			ID             int           `json:"id,omitempty"`
			Symbol         string        `json:"symbol,omitempty"`
			Title          string        `json:"title,omitempty"`
			Aliases        []interface{} `json:"aliases,omitempty"`
			IsFollowing    bool          `json:"is_following,omitempty"`
			WatchlistCount int           `json:"watchlist_count,omitempty"`
		} `json:"symbols,omitempty"`
		MentionedUsers []interface{} `json:"mentioned_users,omitempty"`
		Entities       struct {
			Giphy struct {
				ID    string  `json:"id,omitempty"`
				Ratio float64 `json:"ratio,omitempty"`
			} `json:"giphy,omitempty"`
			Sentiment struct {
				Basic string `json:"basic,omitempty"`
			} `json:"sentiment,omitempty"`
		} `json:"entities,omitempty"`
	} `json:"messages,omitempty"`
}
