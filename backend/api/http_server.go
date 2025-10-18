package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"portofoliov1/types"
)

type PriceData struct {
	mu        sync.RWMutex
	AllTokens []types.Asset
}

type HTTPServer struct {
	port                 int
	priceData            *PriceData
	server               *http.Server
	chainRegistryUpdater ChainRegistryUpdater
	sqliteStorage        SQLiteStorageReader
}

type SQLiteStorageReader interface {
	GetLatestTokenPrices() ([]types.TokenPrice, error)
	GetAllUniqueTokens() ([]types.TokenPrice, error)
	GetTokenPrice(symbol string) (*types.TokenPrice, error)
	GetTokenPriceFromPools(symbol string) (*types.TokenPrice, error)
	GetAllPoolsForToken(symbol string) ([]types.PoolPrice, error)
	GetLatestPoolPrices() ([]types.PoolPrice, error)
	GetDatabaseStats() (map[string]interface{}, error)
}

type ChainRegistryUpdater interface {
	ForceUpdate() error
	GetLastUpdateTime() (time.Time, error)
}

func NewHTTPServer(port int, updater ChainRegistryUpdater, storage SQLiteStorageReader) *HTTPServer {
	return &HTTPServer{
		port: port,
		priceData: &PriceData{
			AllTokens: []types.Asset{},
		},
		chainRegistryUpdater: updater,
		sqliteStorage:        storage,
	}
}

func (s *HTTPServer) Start() error {
	if err := s.loadChainRegistryTokens(); err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/tokens", s.handleGetAllTokens)
	mux.HandleFunc("/api/tokens/", s.handleGetToken)
	mux.HandleFunc("/api/pools", s.handleGetPools)
	mux.HandleFunc("/api/convert", s.handleConvert)
	mux.HandleFunc("/api/chain-registry/update", s.handleForceUpdateChainRegistry)
	mux.HandleFunc("/api/chain-registry/status", s.handleChainRegistryStatus)
	mux.Handle("/", http.FileServer(http.Dir("static")))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Println("🌐 HTTP Server started on port", s.port)
	log.Println("📍 Endpoints:")
	log.Println("   GET  /api/health")
	log.Println("   GET  /api/tokens")
	log.Println("   GET  /api/tokens/{symbol}/pools")
	log.Println("   GET  /api/pools")
	log.Println()

	return s.server.ListenAndServe()
}

func (s *HTTPServer) loadChainRegistryTokens() error {
	assetService, err := types.NewAssetService()
	if err != nil {
		return err
	}

	s.priceData.mu.Lock()
	defer s.priceData.mu.Unlock()

	s.priceData.AllTokens = make([]types.Asset, 0, len(assetService.TokenMetadata))
	for _, asset := range assetService.TokenMetadata {
		s.priceData.AllTokens = append(s.priceData.AllTokens, asset)
	}

	log.Printf("✅ Loaded %d tokens from chain-registry", len(s.priceData.AllTokens))
	return nil
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats, err := s.sqliteStorage.GetDatabaseStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":   "healthy",
		"database": stats,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGetToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/tokens/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Token symbol required", http.StatusBadRequest)
		return
	}

	symbol := strings.ToUpper(pathParts[0])

	if len(pathParts) > 1 && pathParts[1] == "pools" {
		s.handleGetTokenPools(w, r, symbol)
		return
	}

	http.Error(w, "Use /api/tokens/{symbol}/pools", http.StatusBadRequest)
}

func (s *HTTPServer) handleGetTokenPools(w http.ResponseWriter, r *http.Request, symbol string) {
	pools, err := s.sqliteStorage.GetAllPoolsForToken(symbol)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed: %v", err), http.StatusInternalServerError)
		return
	}

	type PoolWithPrice struct {
		PoolID       string    `json:"pool_id"`
		PairedWith   string    `json:"paired_with"`
		PairedDenom  string    `json:"paired_denom"`
		TokenPrice   float64   `json:"token_price"`
		InversePrice float64   `json:"inverse_price"`
		LiquidityUSD float64   `json:"liquidity_usd"`
		Timestamp    time.Time `json:"timestamp"`
	}

	result := make([]PoolWithPrice, 0, len(pools))
	for _, pool := range pools {
		var pairedSymbol, pairedDenom string
		var tokenPrice, inversePrice float64

		if pool.Token0Symbol == symbol {
			pairedSymbol = pool.Token1Symbol
			pairedDenom = pool.Token1Denom
			tokenPrice = pool.PriceToken0ToToken1
			inversePrice = pool.PriceToken1ToToken0
		} else {
			pairedSymbol = pool.Token0Symbol
			pairedDenom = pool.Token0Denom
			tokenPrice = pool.PriceToken1ToToken0
			inversePrice = pool.PriceToken0ToToken1
		}

		result = append(result, PoolWithPrice{
			PoolID:       pool.PoolID,
			PairedWith:   pairedSymbol,
			PairedDenom:  pairedDenom,
			TokenPrice:   tokenPrice,
			InversePrice: inversePrice,
			LiquidityUSD: pool.LiquidityUSD,
			Timestamp:    pool.Timestamp,
		})
	}

	var latestUpdate time.Time
	if len(pools) > 0 {
		latestUpdate = pools[0].Timestamp
	}

	response := map[string]interface{}{
		"symbol":        symbol,
		"pools":         result,
		"count":         len(result),
		"latest_update": latestUpdate,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGetAllTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats, err := s.sqliteStorage.GetDatabaseStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":       "Token list",
		"latest_update": stats["latest_record"],
	}

	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGetPools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pools, err := s.sqliteStorage.GetLatestPoolPrices()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	var latestUpdate time.Time
	if len(pools) > 0 {
		latestUpdate = pools[0].Timestamp
	}

	response := map[string]interface{}{
		"pools":         pools,
		"count":         len(pools),
		"latest_update": latestUpdate,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleConvert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Coming soon"})
}

func (s *HTTPServer) handleForceUpdateChainRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := s.chainRegistryUpdater.ForceUpdate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.loadChainRegistryTokens(); err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Updated",
	})
}

func (s *HTTPServer) handleChainRegistryStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	lastUpdate, err := s.chainRegistryUpdater.GetLastUpdateTime()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	s.priceData.mu.RLock()
	tokenCount := len(s.priceData.AllTokens)
	s.priceData.mu.RUnlock()

	response := map[string]interface{}{
		"last_update": lastUpdate,
		"token_count": tokenCount,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
