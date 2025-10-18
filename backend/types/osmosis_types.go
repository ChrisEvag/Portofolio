package types

// OsmosisPool represents a liquidity pool in the Osmosis blockchain
type OsmosisPool struct {
	Type       string `json:"@type"`
	Address    string `json:"address"`
	Id         string `json:"id"`
	PoolParams struct {
		SwapFee                  string      `json:"swap_fee"`
		ExitFee                  string      `json:"exit_fee"`
		SmoothWeightChangeParams interface{} `json:"smooth_weight_change_params"`
	} `json:"pool_params"`
	FuturePoolGovernor string `json:"future_pool_governor"`
	TotalShares        struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"total_shares"`
	PoolAssets []BasicPoolAsset `json:"pool_assets"`
}

func (p OsmosisPool) GetId() string {
	return p.Id
}

func (p OsmosisPool) GetAssets() []IPoolAsset {
	assets := make([]IPoolAsset, len(p.PoolAssets))
	for i, asset := range p.PoolAssets {
		assets[i] = asset
	}
	return assets
}

// OsmosisPoolAsset represents a token in an Osmosis pool
type OsmosisPoolAsset struct {
	BasicPoolAsset
}

// OsmosisSwapRoute represents a route for token swaps
type OsmosisSwapRoute struct {
	PoolId        string `json:"pool_id"`
	TokenInDenom  string `json:"token_in_denom"`
	TokenOutDenom string `json:"token_out_denom"`
}

// OsmosisRouteResponse represents the best route for a swap
type OsmosisRouteResponse struct {
	Routes []OsmosisSwapRoute `json:"routes"`
}

// OsmosisSpotPriceRequest represents a request for spot price
type OsmosisSpotPriceRequest struct {
	PoolId   string `json:"pool_id"`
	TokenIn  string `json:"token_in"`
	TokenOut string `json:"token_out"`
}

// OsmosisSpotPriceResponse represents the response from a spot price request
type OsmosisSpotPriceResponse struct {
	SpotPrice string `json:"spot_price"`
}

// OsmosisSwapRequest represents a request to simulate/execute a swap
type OsmosisSwapRequest struct {
	TokenIn BasicCoin          `json:"token_in"`
	Routes  []OsmosisSwapRoute `json:"routes"`
}

// OsmosisSwapResponse represents the response from a swap simulation/execution
type OsmosisSwapResponse struct {
	TokenOut BasicCoin `json:"token_out"`
}
