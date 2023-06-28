package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/skip-mev/slinky/oracle/types"
	"github.com/skip-mev/slinky/providers"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"
)

// getPriceForPair returns the price of a currency pair. The price is fetched
// from the CoinGecko API in a single request for all pairs. Since the CoinGecko
// response will match some base denoms to quote denoms that should not be supported,
// we filter out pairs that are not supported by the provider.
//
// Response format:
//
//	{
//	  "cosmos": {
//	    "usd": 11.35
//	  },
//	  "bitcoin": {
//	    "usd": 10000
//	  }
//	}
func (p *Provider) getPrices(ctx context.Context) (map[oracletypes.CurrencyPair]types.QuotePrice, error) {
	url := getPriceEndpoint(p.bases, p.quotes)

	// make the request to url and unmarshal the response into respMap
	respMap := make(map[string]map[string]float64)

	if err := providers.GetWithContext(ctx, url, func(body []byte) error {
		return json.Unmarshal(body, &respMap)
	}); err != nil {
		return nil, err
	}

	prices := make(map[oracletypes.CurrencyPair]types.QuotePrice)

	for _, pair := range p.pairs {
		base := strings.ToLower(pair.Base)
		quote := strings.ToLower(pair.Quote)

		if _, ok := respMap[base]; !ok {
			continue
		}

		if _, ok := respMap[base][quote]; !ok {
			continue
		}

		quotePrice, err := types.NewQuotePrice(
			providers.Float64ToUint256(respMap[base][quote], pair.Decimals()),
			time.Now(),
		)
		if err != nil {
			continue
		}

		prices[pair] = quotePrice
	}

	return prices, nil
}

// getPriceEndpoint is the CoinGecko endpoint for getting the price of a
// currency pair.
func getPriceEndpoint(base, quote string) string {
	return fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s&precision=18", base, quote)
}
