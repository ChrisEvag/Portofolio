package utils

import (
	"fmt"
	"portofoliov1/types"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// FormatOSMO - Μορφοποίηση OSMO
func FormatOSMO(amount float64) string {
	if amount >= 1_000_000 {
		return fmt.Sprintf("%.2fM OSMO", amount/1_000_000)
	} else if amount >= 1_000 {
		return fmt.Sprintf("%.2fK OSMO", amount/1_000)
	}
	return fmt.Sprintf("%.6f OSMO", amount)
}

// ExtractTokensFromPools - Εξαγωγή tokens από Osmosis pools
func ExtractTokensFromPools(pools []types.OsmosisPool) map[string]types.TokenInfo {
	tokens := make(map[string]types.TokenInfo)

	for _, pool := range pools {
		for _, asset := range pool.PoolAssets {
			denom := asset.Token.Denom

			if _, exists := tokens[denom]; !exists {
				tokens[denom] = types.TokenInfo{
					Denom:     denom,
					Symbol:    extractSymbolFromDenom(denom),
					Name:      extractNameFromDenom(denom),
					Price:     0.0, // Default to 0, will be set later from pools
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

// ConvertToOsmosisPools - Μετατροπή PoolResponse σε OsmosisPool
func ConvertToOsmosisPools(poolResponses []types.PoolResponse) []types.OsmosisPool {
	var result []types.OsmosisPool
	for _, pr := range poolResponses {
		for _, p := range pr.Pools {
			// Μετατροπή των pool assets
			var poolAssets []types.BasicPoolAsset
			for _, asset := range p.PoolAssets {
				poolAsset := types.BasicPoolAsset{
					Token: types.BasicCoin{
						Denom:  asset.Token.Denom,
						Amount: asset.Token.Amount,
					},
				}
				poolAssets = append(poolAssets, poolAsset)
			}

			pool := types.OsmosisPool{
				Id:         p.ID.String(),
				Type:       p.Type,
				PoolAssets: poolAssets,
				PoolParams: struct {
					SwapFee                  string      `json:"swap_fee"`
					ExitFee                  string      `json:"exit_fee"`
					SmoothWeightChangeParams interface{} `json:"smooth_weight_change_params"`
				}{
					SwapFee: p.SwapFee,
					ExitFee: p.ExitFee,
				},
				TotalShares: struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				}{
					Denom:  "gamm/pool/" + p.ID.String(),
					Amount: "0", // Θα χρειαστεί να το βρούμε από κάπου
				},
			}
			result = append(result, pool)
		}
	}
	return result
}

func extractSymbolFromDenom(denom string) string {
	if strings.HasPrefix(denom, "ibc/") {
		return "IBC_" + denom[len(denom)-8:]
	}
	if len(denom) > 8 {
		return strings.ToUpper(denom[:8])
	}
	return strings.ToUpper(denom)
}

func extractNameFromDenom(denom string) string {
	titleCaser := cases.Title(language.English)
	if strings.HasPrefix(denom, "u") {
		return titleCaser.String(denom[1:]) + " Token"
	}
	return titleCaser.String(denom) + " Token"
}
