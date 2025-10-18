package types

// Generic Pool Types - these are generic interfaces for all DEXs
type IPool interface {
	GetId() string
	GetAssets() []IPoolAsset
}

type IPoolAsset interface {
	GetToken() ICoin
	GetWeight() string
}

type ICoin interface {
	GetDenom() string
	GetAmount() string
}

// BasicCoin implements ICoin interface
type BasicCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func (c BasicCoin) GetDenom() string  { return c.Denom }
func (c BasicCoin) GetAmount() string { return c.Amount }

// BasicPoolAsset implements IPoolAsset interface
type BasicPoolAsset struct {
	Token  BasicCoin `json:"token"`
	Weight string    `json:"weight"`
}

func (p BasicPoolAsset) GetToken() ICoin   { return p.Token }
func (p BasicPoolAsset) GetWeight() string { return p.Weight }

// BasicPool implements IPool interface
type BasicPool struct {
	Id          string           `json:"id"`
	Type        string           `json:"type"`
	PoolAssets  []BasicPoolAsset `json:"pool_assets"`
	TotalWeight string           `json:"total_weight"`
	SwapFee     string           `json:"swap_fee"`
	ExitFee     string           `json:"exit_fee"`
}

func (p BasicPool) GetId() string { return p.Id }
func (p BasicPool) GetAssets() []IPoolAsset {
	assets := make([]IPoolAsset, len(p.PoolAssets))
	for i, asset := range p.PoolAssets {
		assets[i] = asset
	}
	return assets
}

// Για τα ticks (spot price updates)
type SpotPriceTick struct {
	PoolId    uint64  `json:"pool_id"`
	Token0    string  `json:"token0"`
	Token1    string  `json:"token1"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// Για το concentrated liquidity
type ConcentratedPool struct {
	Id               string `json:"id"`
	CurrentSqrtPrice string `json:"current_sqrt_price"`
	CurrentTick      int64  `json:"current_tick"`
	Token0           string `json:"token0"`
	Token1           string `json:"token1"`
	TickSpacing      uint64 `json:"tick_spacing"`
	ExponentAtPrice  int32  `json:"exponent_at_price"`
}

// Για τα swaps
type SwapAmount struct {
	TokenIn  BasicCoin `json:"token_in"`
	TokenOut BasicCoin `json:"token_out"`
}

type SimulateSwapResponse struct {
	TokenIn  BasicCoin `json:"token_in"`
	TokenOut BasicCoin `json:"token_out"`
	Fee      BasicCoin `json:"fee"`
}

// Για τα pool statistics
type PoolStats struct {
	PoolId    string             `json:"pool_id"`
	Volume24h float64            `json:"volume_24h"`
	Volume7d  float64            `json:"volume_7d"`
	Fees24h   float64            `json:"fees_24h"`
	TVL       float64            `json:"tvl"`
	TokenAPRs map[string]float64 `json:"token_aprs"`
}
