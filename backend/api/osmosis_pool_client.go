package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"portofoliov1/types"
)

type OsmosisPoolClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewOsmosisPoolClient() *OsmosisPoolClient {
	return &OsmosisPoolClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://lcd.osmosis.zone", // Χρησιμοποιούμε το επίσημο LCD API
	}
}

// GetPoolById επιστρέφει λεπτομέρειες για ένα συγκεκριμένο pool
func (c *OsmosisPoolClient) GetPoolById(poolId string) (*types.OsmosisPool, error) {
	url := fmt.Sprintf("%s/pools/%s", c.baseURL, poolId)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά την ανάκτηση pool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("μη αναμενόμενο status code: %d", resp.StatusCode)
	}

	var pool types.OsmosisPool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, fmt.Errorf("σφάλμα κατά το parsing του pool: %w", err)
	}

	return &pool, nil
}

// GetAllPools επιστρέφει όλα τα διαθέσιμα pools
func (c *OsmosisPoolClient) GetAllPools(limit int, offset int) ([]types.OsmosisPool, error) {
	url := fmt.Sprintf("%s/osmosis/gamm/v1beta1/pools?pagination.limit=%d&pagination.offset=%d", c.baseURL, limit, offset)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά την ανάκτηση pools: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("μη αναμενόμενο status code: %d", resp.StatusCode)
	}

	var poolsResponse struct {
		Pools []types.OsmosisPool `json:"pools"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&poolsResponse); err != nil {
		return nil, fmt.Errorf("σφάλμα κατά το parsing των pools: %w", err)
	}

	return poolsResponse.Pools, nil
}

// GetSpotPrice επιστρέφει την τρέχουσα τιμή μεταξύ δύο tokens σε ένα pool
func (c *OsmosisPoolClient) GetSpotPrice(poolId string, tokenIn string, tokenOut string) (float64, error) {
	url := fmt.Sprintf("%s/osmosis/gamm/v1beta1/pools/%s/prices?base_asset_denom=%s&quote_asset_denom=%s", c.baseURL, poolId, tokenIn, tokenOut)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("σφάλμα κατά την ανάκτηση spot price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("μη αναμενόμενο status code: %d", resp.StatusCode)
	}

	var response struct {
		SpotPrice string `json:"spot_price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("σφάλμα κατά το parsing του price: %w", err)
	}

	price, err := strconv.ParseFloat(response.SpotPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("σφάλμα κατά τη μετατροπή του price: %w", err)
	}

	return price, nil
}

// SimulateSwap προσομοιώνει ένα swap και επιστρέφει το αναμενόμενο αποτέλεσμα
func (c *OsmosisPoolClient) SimulateSwap(poolId string, tokenIn types.BasicCoin) (*types.SimulateSwapResponse, error) {
	url := fmt.Sprintf("%s/osmosis/gamm/v1beta1/pools/%s/estimate/swap", c.baseURL, poolId)

	body, err := json.Marshal(map[string]interface{}{
		"token_in": tokenIn,
	})
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά τη σειριοποίηση του request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά την προσομοίωση του swap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("μη αναμενόμενο status code: %d", resp.StatusCode)
	}

	var result types.SimulateSwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("σφάλμα κατά το parsing του simulation result: %w", err)
	}

	return &result, nil
}

// GetPoolStats επιστρέφει στατιστικά για ένα pool (volume, TVL, APR κλπ)
func (c *OsmosisPoolClient) GetPoolStats(poolId string) (*types.PoolStats, error) {
	url := fmt.Sprintf("%s/osmosis/gamm/v1beta1/pools/%s/stats", c.baseURL, poolId)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά την ανάκτηση pool stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("μη αναμενόμενο status code: %d", resp.StatusCode)
	}

	var stats types.PoolStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("σφάλμα κατά το parsing των stats: %w", err)
	}

	return &stats, nil
}

// GetBlockHeight επιστρέφει το τρέχον block height του Osmosis chain
func (c *OsmosisPoolClient) GetBlockHeight() (*types.BlockHeightResponse, error) {
	url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά την ανάκτηση block height: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("μη αναμενόμενο status code: %d", resp.StatusCode)
	}

	var response struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("σφάλμα κατά το parsing του block height: %w", err)
	}

	height, err := strconv.ParseInt(response.Block.Header.Height, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("σφάλμα κατά τη μετατροπή του height: %w", err)
	}

	return &types.BlockHeightResponse{Height: height}, nil
}

// CalculateSpotPrices υπολογίζει τις τιμές όλων των tokens σε USD
func (c *OsmosisPoolClient) CalculateSpotPrices(pools []types.OsmosisPool, assetService *types.AssetService) (map[string]float64, error) {
	// Βοηθητικοί χάρτες
	prices := make(map[string]float64)       // Τελικές τιμές σε USD
	poolPrices := make(map[string][]float64) // Τιμές από διάφορα pools

	// Γνωστά stablecoins και η τιμή τους σε USD
	stableCoins := map[string]float64{
		"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858": 1.0, // USDC
		"ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4": 1.0, // USDT
		"ibc/6329DD8CF31A334DD5BE3F68C846C9FE313281362B37686A62343BAC1EB1546D": 1.0, // BUSD
		"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7": 1.0, // DAI
	}

	// Σιωπηλός υπολογισμός - no logs
	var stablePools int
	for _, pool := range pools {

		if len(pool.PoolAssets) != 2 {
			continue
		}

		asset0 := pool.PoolAssets[0]
		asset1 := pool.PoolAssets[1]

		// Έλεγξε αν το pool περιέχει stablecoin
		hasStable := false
		var stableDenom, otherDenom string
		var stableAmount, otherAmount float64

		for denom := range stableCoins {
			if asset0.Token.Denom == denom {
				hasStable = true
				stableDenom = asset0.Token.Denom
				otherDenom = asset1.Token.Denom
				stableAmount, _ = strconv.ParseFloat(asset0.Token.Amount, 64)
				otherAmount, _ = strconv.ParseFloat(asset1.Token.Amount, 64)
				break
			} else if asset1.Token.Denom == denom {
				hasStable = true
				stableDenom = asset1.Token.Denom
				otherDenom = asset0.Token.Denom
				stableAmount, _ = strconv.ParseFloat(asset1.Token.Amount, 64)
				otherAmount, _ = strconv.ParseFloat(asset0.Token.Amount, 64)
				break
			}
		}

		if hasStable && stableAmount > 0 && otherAmount > 0 {
			stablePools++
			// Υπολόγισε τιμή σε USD
			stableExp := assetService.GetExponent(stableDenom)
			otherExp := assetService.GetExponent(otherDenom)

			stableValue := stableAmount / math.Pow10(stableExp) * stableCoins[stableDenom]
			otherValue := otherAmount / math.Pow10(otherExp)

			price := stableValue / otherValue
			if price > 0 && price < 1e12 { // Φιλτράρισμα εξωφρενικών τιμών
				symbol := assetService.GetSymbol(otherDenom)
				if symbol != "" {
					poolPrices[symbol] = append(poolPrices[symbol], price)
				}
			}
		}
	}

	// Σιωπηλή ολοκλήρωση - no logs

	// Υπολόγισε μέσες τιμές
	for symbol, priceList := range poolPrices {
		if len(priceList) > 0 {
			var sum float64
			for _, p := range priceList {
				sum += p
			}
			avgPrice := sum / float64(len(priceList))
			prices[symbol] = avgPrice

			// Αποθήκευση OSMO price χωρίς log
			if symbol == "OSMO" {
				assetService.SetOsmoUsdPrice(avgPrice)
			}
		}
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("δεν βρέθηκαν stablecoin pools για υπολογισμό τιμών")
	}

	return prices, nil
}

// GetAllPoolPrices επιστρέφει τις τιμές για όλα τα pools
func (c *OsmosisPoolClient) GetAllPoolPrices(pools []types.OsmosisPool, assetService *types.AssetService) ([]types.PoolPrice, error) {
	poolPrices := make([]types.PoolPrice, 0, len(pools))
	timestamp := time.Now()

	var skippedPools, processedPools int

	for _, pool := range pools {
		// Δουλεύουμε μόνο με pools 2 assets
		if len(pool.PoolAssets) != 2 {
			skippedPools++
			continue
		}

		asset0 := pool.PoolAssets[0]
		asset1 := pool.PoolAssets[1]

		// Parse amounts με error handling
		amount0, err := strconv.ParseFloat(asset0.Token.Amount, 64)
		if err != nil {
			skippedPools++
			continue
		}
		amount1, err := strconv.ParseFloat(asset1.Token.Amount, 64)
		if err != nil {
			skippedPools++
			continue
		}

		// Υπολόγισε την τιμή: πόσο token1 χρειάζεσαι για 1 token0
		var price float64
		if amount0 > 0 && amount1 > 0 {
			// Προσαρμογή με exponents
			exp0 := assetService.GetExponent(asset0.Token.Denom)
			exp1 := assetService.GetExponent(asset1.Token.Denom)

			adjustedAmount0 := amount0 / math.Pow10(exp0)
			adjustedAmount1 := amount1 / math.Pow10(exp1)

			if adjustedAmount0 > 0 {
				price = adjustedAmount1 / adjustedAmount0
			}
		}

		// Λήψη symbols
		symbol0 := assetService.GetSymbol(asset0.Token.Denom)
		symbol1 := assetService.GetSymbol(asset1.Token.Denom)

		// Αν δεν βρέθηκαν symbols, χρησιμοποίησε truncated denoms
		if symbol0 == "" {
			if len(asset0.Token.Denom) > 12 {
				symbol0 = asset0.Token.Denom[:12] + "..."
			} else {
				symbol0 = asset0.Token.Denom
			}
		}
		if symbol1 == "" {
			if len(asset1.Token.Denom) > 12 {
				symbol1 = asset1.Token.Denom[:12] + "..."
			} else {
				symbol1 = asset1.Token.Denom
			}
		}

		poolPrice := types.PoolPrice{
			PoolID:              pool.Id,
			Token0Symbol:        symbol0,
			Token0Denom:         asset0.Token.Denom,
			Token0Amount:        asset0.Token.Amount,
			Token1Symbol:        symbol1,
			Token1Denom:         asset1.Token.Denom,
			Token1Amount:        asset1.Token.Amount,
			PriceOSMO:           price,
			PriceToken0ToToken1: price,       // Ίδιο με PriceOSMO για συμβατότητα
			PriceToken1ToToken0: 1.0 / price, // Αντίστροφη τιμή
			LiquidityUSD:        0.0,         // Θα προστεθεί αργότερα
			Timestamp:           timestamp,
		}

		poolPrices = append(poolPrices, poolPrice)
		processedPools++
	}

	return poolPrices, nil
}
