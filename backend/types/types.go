package types

import "encoding/json"

type PoolResponse struct {
	Pools []struct {
		ID         json.Number `json:"id"`
		Type       string      `json:"type"`
		PoolAssets []struct {
			Token struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"token"`
		} `json:"pool_assets"`
		TotalWeight string `json:"total_weight"`
		SwapFee     string `json:"swap_fee"`
		ExitFee     string `json:"exit_fee"`
	} `json:"pools"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// Μετατροπή από PoolResponse σε BasicPool
func ConvertPoolResponseToPool(poolResponse PoolResponse) []BasicPool {
	var pools []BasicPool

	for _, respPool := range poolResponse.Pools {
		pool := BasicPool{
			Id: respPool.ID.String(),
		}

		for _, asset := range respPool.PoolAssets {
			poolAsset := BasicPoolAsset{
				Token: BasicCoin{
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

type TokenInfo struct {
	Denom       string  `json:"denom"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Liquidity   float64 `json:"liquidity"`
	PoolCount   int     `json:"pool_count"`
	Source      string  `json:"source"`
	BlockHeight int64   `json:"block_height"`
	Chain       string  `json:"chain"` // "osmosis"
	LogoURI     string  `json:"logo_uri,omitempty"`
}

// BlockHeightResponse - Απάντηση από το API για το τρέχον block height
type BlockHeightResponse struct {
	Height int64 `json:"height"`
}
