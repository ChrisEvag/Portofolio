package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"portofoliov1/types"
)

type EndpointInfo struct {
	URL     string
	Speed   time.Duration
	Working bool
	Chain   string
	Type    string // "lcd" ή "indexer"
}

type APIClient struct {
	HTTPClient    *http.Client
	LCDEndpoints  []EndpointInfo
	DydxEndpoints []EndpointInfo
}

func NewAPIClient() *APIClient {
	// Osmosis endpoints (παραμένουν ίδια)
	osmosisEndpoints := []string{
		"https://lcd.osmosis.zone",
		"https://rest.osmosis.goldenratiostaking.net",
		"https://rest.lavenderfive.com:443/osmosis",
		"https://osmosis-api.polkachu.com",
		"https://osmosis.rest.stakin-nodes.com",
		"https://api-osmosis-01.stakeflow.io",
		"https://osmosis-api.w3coins.io",
		"https://osmosis-rest.publicnode.com",
		"https://community.nuxian-node.ch:6797/osmosis/crpc",
		"https://osmosis-api.stake-town.com",
		"https://public.stakewolle.com/cosmos/osmosis/rest",
		"https://rest.cros-nest.com/osmosis",
		"https://osmosis-api.noders.services",
		"https://osmosis-api.highstakes.ch",
	}

	// dYdX Indexer API endpoints
	dydxEndpoints := []string{
		"https://indexer.dydx.trade",
		"https://dydx-indexer.kingnodes.com",
		"https://indexer.dydx.nodestake.org",
		"https://dydx-indexer.polkachu.com",
		"https://dydx-indexer.lavenderfive.com:443",
		"https://dydx-mainnet-lcd.autostake.com:443",
		"https://rest-dydx.ecostake.com:443",
		"https://dydx-rest.publicnode.com",
	}

	var osmosisEP []EndpointInfo
	for _, url := range osmosisEndpoints {
		osmosisEP = append(osmosisEP, EndpointInfo{URL: url, Chain: "osmosis", Type: "lcd"})
	}

	var dydxEP []EndpointInfo
	for _, url := range dydxEndpoints {
		dydxEP = append(dydxEP, EndpointInfo{URL: url, Chain: "dydx", Type: "indexer"})
	}

	return &APIClient{
		HTTPClient:    &http.Client{Timeout: 30 * time.Second},
		LCDEndpoints:  osmosisEP,
		DydxEndpoints: dydxEP,
	}
}

// SpeedTestEndpoints - Ενημερωμένη για dYdX Indexer
func (c *APIClient) SpeedTestEndpoints(chain string) {
	var endpoints []EndpointInfo
	var chainName string
	var testURL string

	switch chain {
	case "osmosis":
		endpoints = c.LCDEndpoints
		chainName = "Osmosis"
		testURL = "/osmosis/gamm/v1beta1/pools?pagination.limit=1"
	case "dydx":
		endpoints = c.DydxEndpoints
		chainName = "dYdX"
		testURL = "/v4/perpetualMarkets"
	default:
		fmt.Printf("❌ Άγνωστο chain: %s\n", chain)
		return
	}

	fmt.Printf("🏎️  Ταχυμέτρηση %s endpoints...\n", chainName)

	type result struct {
		index int
		info  EndpointInfo
	}

	results := make(chan result, len(endpoints))

	for i, endpoint := range endpoints {
		go func(idx int, ep EndpointInfo) {
			start := time.Now()

			url := ep.URL + testURL
			resp, err := c.HTTPClient.Get(url)

			duration := time.Since(start)
			working := false

			if err != nil {
				fmt.Printf("    ❌ %s: %v\n", getHostname(ep.URL), err)
			} else {
				defer resp.Body.Close()
				working = resp.StatusCode == 200
				if working {
					fmt.Printf("    ✅ %s: %v\n", getHostname(ep.URL), duration.Round(time.Millisecond))
				} else {
					fmt.Printf("    ❌ %s: status %d\n", getHostname(ep.URL), resp.StatusCode)
				}
			}

			results <- result{
				index: idx,
				info: EndpointInfo{
					URL:     ep.URL,
					Speed:   duration,
					Working: working,
					Chain:   ep.Chain,
					Type:    ep.Type,
				},
			}
		}(i, endpoint)
	}

	var testedEndpoints []EndpointInfo
	for i := 0; i < len(endpoints); i++ {
		result := <-results
		testedEndpoints = append(testedEndpoints, result.info)
	}

	var workingEndpoints []EndpointInfo
	var failedEndpoints []EndpointInfo

	for _, ep := range testedEndpoints {
		if ep.Working {
			workingEndpoints = append(workingEndpoints, ep)
		} else {
			failedEndpoints = append(failedEndpoints, ep)
		}
	}

	sort.Slice(workingEndpoints, func(i, j int) bool {
		return workingEndpoints[i].Speed < workingEndpoints[j].Speed
	})

	// Update the appropriate endpoints slice
	switch chain {
	case "osmosis":
		c.LCDEndpoints = append(workingEndpoints, failedEndpoints...)
	case "dydx":
		c.DydxEndpoints = append(workingEndpoints, failedEndpoints...)
	}

	fmt.Printf("\n📊 ΑΠΟΤΕΛΕΣΜΑΤΑ ΤΑΧΥΜΕΤΡΗΣΗΣ %s:\n", chainName)
	fmt.Println("================================")

	workingCount := 0
	allEndpoints := append(workingEndpoints, failedEndpoints...)
	for i, ep := range allEndpoints {
		if ep.Working {
			workingCount++
			status := "✅"
			fmt.Printf("%2d. %s %-40s %v\n", i+1, status, getHostname(ep.URL), ep.Speed.Round(time.Millisecond))
		} else {
			status := "❌"
			fmt.Printf("%2d. %s %-40s (απέτυχε)\n", i+1, status, getHostname(ep.URL))
		}
	}

	fmt.Printf("\n📈 Σύνολο: %d/%d endpoints λειτουργούν\n", workingCount, len(allEndpoints))
	if workingCount > 0 {
		fmt.Printf("🚀 Ταχύτερο: %s (%v)\n", getHostname(workingEndpoints[0].URL), workingEndpoints[0].Speed.Round(time.Millisecond))
	}
}

// GetDydxMarketsWithFallback - Ενημερωμένη για Indexer API
func (c *APIClient) GetDydxMarketsWithFallback() ([]types.DydxMarket, error) {
	fmt.Println("🔄 dYdX Sequential Fallback...")

	for i, endpoint := range c.DydxEndpoints {
		if !endpoint.Working {
			fmt.Printf("  %d. ⏭️  Παράλειψη: %s (δεν λειτουργεί)\n", i+1, getHostname(endpoint.URL))
			continue
		}

		fmt.Printf("  %d. Δοκιμή: %s (%v)\n", i+1, getHostname(endpoint.URL), endpoint.Speed.Round(time.Millisecond))

		markets, err := c.getDydxMarketsFromEndpoint(endpoint.URL)
		if err == nil {
			fmt.Printf("  ✅ Επιτυχία με: %s\n", getHostname(endpoint.URL))
			return markets, nil
		}

		fmt.Printf("  ❌ Αποτυχία: %v\n", err)

		if i < len(c.DydxEndpoints)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil, fmt.Errorf("Όλα τα dYdX endpoints απέτυχαν")
}

// getDydxMarketsFromEndpoint - Ενημερωμένη έκδοση με prices
func (c *APIClient) getDydxMarketsFromEndpoint(baseURL string) ([]types.DydxMarket, error) {
	url := baseURL + "/v4/perpetualMarkets"

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Αίτηση απέτυχε: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Status: %d, Response: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}

	var marketResponse struct {
		Markets map[string]struct {
			ClobPairId  string `json:"clobPairId"`
			Ticker      string `json:"ticker"`
			Status      string `json:"status"`
			OraclePrice string `json:"oraclePrice"`
			Volume24H   string `json:"volume24H"`
		} `json:"markets"`
	}

	err = json.NewDecoder(resp.Body).Decode(&marketResponse)
	if err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	// Μετατροπή σε δική μας δομή
	var markets []types.DydxMarket
	for ticker, market := range marketResponse.Markets {
		// Εξαγωγή base και quote asset από το ticker
		baseAsset, quoteAsset := extractAssetsFromTicker(ticker)

		// Μετατροπή oracle price σε float64
		oraclePrice, _ := strconv.ParseFloat(market.OraclePrice, 64)
		volume24H, _ := strconv.ParseFloat(market.Volume24H, 64)

		markets = append(markets, types.DydxMarket{
			MarketID:     market.ClobPairId,
			Ticker:       market.Ticker,
			BaseAsset:    baseAsset,
			QuoteAsset:   quoteAsset,
			OraclePrice:  oraclePrice,
			Volume24H:    volume24H,
			Status:       market.Status,
			MinExchanges: 1,
		})
	}

	if len(markets) == 0 {
		return nil, fmt.Errorf("Δεν βρέθηκαν markets")
	}

	fmt.Printf("    ✅ Βρέθηκαν %d markets\n", len(markets))
	return markets, nil
}

// extractAssetsFromTicker - Βοηθητική function για εξαγωγή assets από ticker
func extractAssetsFromTicker(ticker string) (string, string) {
	// Το ticker είναι σε μορφή "BASE-QUOTE" (π.χ. "BTC-USD", "ETH-USD")
	parts := strings.Split(ticker, "-")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return ticker, "USD" // Fallback
}

// GetLatestBlockHeight - Ενημερωμένο για dYdX
func (c *APIClient) GetLatestBlockHeight(baseURL, chain string) (int64, error) {
	var url string

	switch chain {
	case "osmosis":
		url = baseURL + "/cosmos/base/tendermint/v1beta1/blocks/latest"
	case "dydx":
		url = baseURL + "/cosmos/base/tendermint/v1beta1/blocks/latest"
	default:
		return 0, fmt.Errorf("Άγνωστο chain: %s", chain)
	}

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("Block height request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Block height status: %d", resp.StatusCode)
	}

	var blockResponse struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	}

	err = json.NewDecoder(resp.Body).Decode(&blockResponse)
	if err != nil {
		return 0, fmt.Errorf("Block height JSON decode failed: %w", err)
	}

	height, err := strconv.ParseInt(blockResponse.Block.Header.Height, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Block height parse failed: %w", err)
	}

	return height, nil
}

// Οι υπόλοιπες μέθοδοι παραμένουν όπως είναι...
// [Keep all existing methods like GetNumiaTokens, GetPoolsWithFallback, etc.]

// GetPoolsWithFallback - Sequential fallback που επιστρέφει μόνο pools
func (c *APIClient) GetPoolsWithFallback() ([]types.PoolResponse, error) {
	fmt.Println("🔄 Sequential Fallback (με βελτιωμένο pagination)...")

	for i, endpoint := range c.LCDEndpoints {
		if !endpoint.Working {
			fmt.Printf("  %d. ⏭️  Παράλειψη: %s (δεν λειτουργεί)\n", i+1, getHostname(endpoint.URL))
			continue
		}

		fmt.Printf("  %d. Δοκιμή: %s (%v)\n", i+1, getHostname(endpoint.URL), endpoint.Speed.Round(time.Millisecond))

		pools, err := c.getPoolsFromEndpoint(endpoint.URL)
		if err == nil {
			fmt.Printf("  ✅ Επιτυχία με: %s\n", getHostname(endpoint.URL))
			return pools, nil
		}

		fmt.Printf("  ❌ Αποτυχία: %v\n", err)

		if i < len(c.LCDEndpoints)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil, fmt.Errorf("Όλα τα endpoints απέτυχαν")
}

// getPoolsFromEndpoint - Βελτιωμένη function με καλύτερο pagination handling
func (c *APIClient) getPoolsFromEndpoint(baseURL string) ([]types.PoolResponse, error) {
	var allPools []types.PoolResponse
	limit := 100
	nextKey := ""
	totalPools := 0
	maxPages := 20 // Προστασία από infinite loop

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s/osmosis/gamm/v1beta1/pools?pagination.limit=%d", baseURL, limit)
		if nextKey != "" {
			url += "&pagination.key=" + nextKey
		}

		resp, err := c.HTTPClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("Αίτηση απέτυχε: %w", err)
		}

		// Διαβάζουμε το body ανεξάρτητα από το status code
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("Αδυναμία ανάγνωσης response: %w", err)
		}

		// Προσπαθούμε να decode-άρουμε το JSON
		var poolResp types.PoolResponse
		err = json.Unmarshal(body, &poolResp)
		if err != nil {
			return nil, fmt.Errorf("JSON decode failed: %w", err)
		}

		// Αν δεν υπάρχουν pools, σταματάμε (όχι error)
		if len(poolResp.Pools) == 0 {
			fmt.Printf("    ℹ️  Τέλος pools (σελίδα %d, status: %d)\n", page, resp.StatusCode)
			break
		}

		allPools = append(allPools, poolResp)
		totalPools += len(poolResp.Pools)

		fmt.Printf("    📥 Σελίδα %d: %d pools (status: %d)\n", page, len(poolResp.Pools), resp.StatusCode)

		// Έλεγχος για επόμενη σελίδα
		if poolResp.Pagination.NextKey == "" {
			fmt.Printf("    ✅ Τέλος pagination, σύνολο: %d pools\n", totalPools)
			break
		}
		nextKey = poolResp.Pagination.NextKey

		// Αν φτάσαμε στο τελευταίο page, σταματάμε
		if page == maxPages {
			fmt.Printf("    ⚠️  Φτάσαμε το μέγιστο αριθμό σελίδων (%d), σύνολο: %d pools\n", maxPages, totalPools)
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Επιστρέφουμε τα pools μόνο αν βρήκαμε τουλάχιστον ένα
	if totalPools == 0 {
		return nil, fmt.Errorf("Δεν βρέθηκαν pools")
	}

	return allPools, nil
}

// Προσθήκη της μεθόδου GetNumiaTokens στο APIClient
func (c *APIClient) GetNumiaTokens() ([]types.TokenInfo, error) {
	fmt.Println("⏱️  Δοκιμή Numia API...")

	// Δοκιμή διαφορετικών Numia endpoints
	numiaEndpoints := []string{
		"https://api-osmosis.imperator.co/tokens/v2/all",
		"https://public-osmosis-api.numia.xyz/tokens",
		"https://api.numia.xyz/osmosis/tokens",
	}

	var lastError error
	for _, url := range numiaEndpoints {
		fmt.Printf("  🔍 Δοκιμή: %s\n", getHostname(url))

		start := time.Now()
		resp, err := c.HTTPClient.Get(url)
		duration := time.Since(start)

		if err != nil {
			lastError = fmt.Errorf("%s: %w", getHostname(url), err)
			fmt.Printf("  ❌ Αποτυχία: %v\n", err)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			lastError = fmt.Errorf("%s: status %d", getHostname(url), resp.StatusCode)
			fmt.Printf("  ❌ Status: %d\n", resp.StatusCode)
			continue
		}

		var numiaTokens []types.NumiaToken
		err = json.NewDecoder(resp.Body).Decode(&numiaTokens)
		resp.Body.Close()

		if err != nil {
			lastError = fmt.Errorf("%s: JSON decode failed: %w", getHostname(url), err)
			fmt.Printf("  ❌ JSON error: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ Numia API: %d tokens (%v)\n", len(numiaTokens), duration.Round(time.Millisecond))

		var tokens []types.TokenInfo
		for _, nt := range numiaTokens {
			tokens = append(tokens, types.TokenInfo{
				Denom:     nt.Denom,
				Symbol:    nt.Symbol,
				Name:      nt.Name,
				Price:     nt.Price,
				Liquidity: nt.Liquidity,
				PoolCount: 1,
				Source:    "Numia API",
			})
		}

		return tokens, nil
	}

	return nil, fmt.Errorf("Όλα τα Numia endpoints απέτυχαν: %w", lastError)
}

// Βοηθητική function για min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getHostname - Βοηθητική function για εμφάνιση μόνο του hostname
func getHostname(url string) string {
	// Αφαίρεση protocol
	cleanURL := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")

	// Πάρτε μόνο το hostname
	parts := strings.Split(cleanURL, "/")
	hostname := parts[0]

	// Συντομογραφία αν είναι πολύ μακρύ
	if len(hostname) > 35 {
		return hostname[:32] + "..."
	}
	return hostname
}
