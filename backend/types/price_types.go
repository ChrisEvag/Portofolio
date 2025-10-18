package types

import "time"

// TokenPrice represents a price entry for a token
type TokenPrice struct {
	Symbol    string    `json:"symbol"`
	Denom     string    `json:"denom"`
	PriceUSD  float64   `json:"price_usd"`
	PriceOSMO float64   `json:"price_osmo"` // νέο πεδίο
	Timestamp time.Time `json:"timestamp"`
}

// PoolPrice represents price data for a liquidity pool pair
type PoolPrice struct {
	PoolID              string    `json:"pool_id"`
	Token0Symbol        string    `json:"token0_symbol"`
	Token0Denom         string    `json:"token0_denom"`
	Token0Amount        string    `json:"token0_amount"`
	Token1Symbol        string    `json:"token1_symbol"`
	Token1Denom         string    `json:"token1_denom"`
	Token1Amount        string    `json:"token1_amount"`
	PriceOSMO           float64   `json:"price_osmo"`             // Τιμή του Token0 σε Token1
	PriceToken0ToToken1 float64   `json:"price_token0_to_token1"` // Τιμή Token0 -> Token1
	PriceToken1ToToken0 float64   `json:"price_token1_to_token0"` // Τιμή Token1 -> Token0
	LiquidityUSD        float64   `json:"liquidity_usd"`          // Total liquidity in USD
	Timestamp           time.Time `json:"timestamp"`
}
