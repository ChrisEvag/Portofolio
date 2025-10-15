package utils

import (
	"fmt"
	"portofoliov1/types"
	"strings"
)

// FormatUSD - Μορφοποίηση USD
func FormatUSD(amount float64) string {
	if amount >= 1_000_000 {
		return fmt.Sprintf("$%.2fM", amount/1_000_000)
	} else if amount >= 1_000 {
		return fmt.Sprintf("$%.2fK", amount/1_000)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// ExtractDydxTokensFromMarkets - Ενημερωμένη με πραγματικές τιμές
func ExtractDydxTokensFromMarkets(markets []types.DydxMarket) map[string]types.TokenInfo {
	tokens := make(map[string]types.TokenInfo)

	for _, market := range markets {
		// Παράβλεψη μη ενεργών markets
		if market.Status != "" && market.Status != "ACTIVE" {
			continue
		}

		// Base asset token με πραγματική τιμή
		baseDenom := "dydx-" + market.BaseAsset
		if _, exists := tokens[baseDenom]; !exists && market.OraclePrice > 0 {
			tokens[baseDenom] = types.TokenInfo{
				Denom:     baseDenom,
				Symbol:    market.BaseAsset,
				Name:      market.BaseAsset + " Token",
				Price:     market.OraclePrice,
				Liquidity: market.Volume24H,
				PoolCount: 1,
				Chain:     "dydx",
			}
		}

		// Quote asset token (USDC) - συνήθως $1
		quoteDenom := "dydx-" + market.QuoteAsset
		if _, exists := tokens[quoteDenom]; !exists {
			price := 1.0 // USDC είναι συνήθως $1
			if market.QuoteAsset != "USD" {
				price = 1.0 // Προσωρινό, μπορούμε να βρούμε πραγματικές τιμές αργότερα
			}

			tokens[quoteDenom] = types.TokenInfo{
				Denom:     quoteDenom,
				Symbol:    market.QuoteAsset,
				Name:      market.QuoteAsset + " Token",
				Price:     price,
				Liquidity: 0,
				PoolCount: 1,
				Chain:     "dydx",
			}
		}
	}

	return tokens
}

// ExtractTokensFromPools - Ενημερωμένη για Osmosis (προσθήκη πραγματικών τιμών αργότερα)
func ExtractTokensFromPools(poolResponses []types.PoolResponse) map[string]types.TokenInfo {
	tokens := make(map[string]types.TokenInfo)

	for _, poolResponse := range poolResponses {
		for _, pool := range poolResponse.Pools {
			for _, asset := range pool.PoolAssets {
				denom := asset.Token.Denom

				if _, exists := tokens[denom]; !exists {
					tokens[denom] = types.TokenInfo{
						Denom:     denom,
						Symbol:    extractSymbolFromDenom(denom),
						Name:      extractNameFromDenom(denom),
						Price:     1.0, // Προσωρινό - θα αντικατασταθεί με Numia API
						Liquidity: 0,
						PoolCount: 0,
						Chain:     "osmosis",
					}
				}

				// Ενημέρωση pool count
				token := tokens[denom]
				token.PoolCount++
				tokens[denom] = token
			}
		}
	}

	return tokens
}

// GetTokenPrices - Τώρα επιστρέφει απλώς τα tokens χωρίς αλλαγή τιμών
func GetTokenPrices(tokens map[string]types.TokenInfo) []types.TokenInfo {
	var result []types.TokenInfo
	for _, token := range tokens {
		result = append(result, token)
	}
	return result
}

// Οι υπόλοιπες συναρτήσεις παραμένουν ίδιες...
func extractSymbolFromDenom(denom string) string {
	if strings.HasPrefix(denom, "ibc/") {
		return "IBC_" + denom[len(denom)-8:]
	}
	if strings.HasPrefix(denom, "dydx-") {
		return denom[5:] // Αφαίρεση "dydx-" prefix
	}
	if len(denom) > 8 {
		return strings.ToUpper(denom[:8])
	}
	return strings.ToUpper(denom)
}

func extractNameFromDenom(denom string) string {
	if strings.HasPrefix(denom, "dydx-") {
		return strings.Title(denom[5:]) + " Token"
	}
	if strings.HasPrefix(denom, "u") {
		return strings.Title(denom[1:]) + " Token"
	}
	return strings.Title(denom) + " Token"
}
