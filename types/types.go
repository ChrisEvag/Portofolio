package types

import "encoding/json"

type PoolResponse struct {
	Pools []struct {
		ID         json.Number `json:"id"`
		PoolAssets []struct {
			Token struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"token"`
		} `json:"pool_assets"`
	} `json:"pools"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// Pool - Απλοποιημένη δομή pool
type Pool struct {
	ID         string
	PoolAssets []PoolAsset
}

type PoolAsset struct {
	Token Token
}

type Token struct {
	Denom  string
	Amount string
}

// Μετατροπή από PoolResponse σε Pool
func ConvertPoolResponseToPool(poolResponse PoolResponse) []Pool {
	var pools []Pool

	for _, respPool := range poolResponse.Pools {
		pool := Pool{
			ID: respPool.ID.String(),
		}

		for _, asset := range respPool.PoolAssets {
			poolAsset := PoolAsset{
				Token: Token{
					Denom:  asset.Token.Denom,
					Amount: asset.Token.Amount,
				},
			}
			pool.PoolAssets = append(pool.PoolAssets, poolAsset)
		}

		pools = append(pools, pool)
	}

	return pools
}

type NumiaToken struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Denom     string  `json:"denom"`
	Liquidity float64 `json:"liquidity"`
}

type TokenInfo struct {
	Denom       string  `json:"denom"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Liquidity   float64 `json:"liquidity"`
	PoolCount   int     `json:"pool_count"`
	Source      string  `json:"source"`
	BlockHeight int64   `json:"block_height"`
	Chain       string  `json:"chain"` // "osmosis" ή "dydx"
}

// DYDX SPECIFIC STRUCTS
type DydxMarketResponse struct {
	Markets map[string]DydxMarket `json:"markets"`
}

// DydxMarket - Ενημερωμένη δομή με τιμές
type DydxMarket struct {
	MarketID     string  `json:"marketId"`
	Ticker       string  `json:"ticker"`
	BaseAsset    string  `json:"baseAsset"`
	QuoteAsset   string  `json:"quoteAsset"`
	OraclePrice  float64 `json:"oraclePrice,omitempty"`
	Volume24H    float64 `json:"volume24H,omitempty"`
	MinExchanges int     `json:"minExchanges"`
	Status       string  `json:"status,omitempty"`
}
