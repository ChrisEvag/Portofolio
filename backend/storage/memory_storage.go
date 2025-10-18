package storage

import (
	"fmt"
	"portofoliov1/types"
	"sync"
	"time"
)

// MemoryStorage - In-memory cache για real-time data (χωρίς persistence)
type MemoryStorage struct {
	pools      map[string]types.OsmosisPool // pool_id -> pool
	poolPrices map[string]types.PoolPrice   // pool_id -> latest price
	tokenPools map[string][]string          // token_symbol -> []pool_ids
	lastUpdate time.Time
	mu         sync.RWMutex // Thread-safe access
}

// NewMemoryStorage - Δημιουργία νέου in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		pools:      make(map[string]types.OsmosisPool),
		poolPrices: make(map[string]types.PoolPrice),
		tokenPools: make(map[string][]string),
		lastUpdate: time.Now(),
	}
}

// SavePools - Αποθήκευση pools στη μνήμη
func (m *MemoryStorage) SavePools(pools []types.OsmosisPool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pool := range pools {
		m.pools[pool.Id] = pool
	}

	m.lastUpdate = time.Now()
	return nil
}

// SavePoolPrices - Αποθήκευση pool prices στη μνήμη
func (m *MemoryStorage) SavePoolPrices(prices []types.PoolPrice) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear old token->pools mappings
	m.tokenPools = make(map[string][]string)

	for _, price := range prices {
		m.poolPrices[price.PoolID] = price

		// Build token->pools index
		if price.Token0Symbol != "" {
			m.tokenPools[price.Token0Symbol] = append(m.tokenPools[price.Token0Symbol], price.PoolID)
		}
		if price.Token1Symbol != "" {
			m.tokenPools[price.Token1Symbol] = append(m.tokenPools[price.Token1Symbol], price.PoolID)
		}
	}

	m.lastUpdate = time.Now()
	return nil
}

// GetAllPoolsForToken - Επιστρέφει όλα τα pools που περιέχουν ένα token
func (m *MemoryStorage) GetAllPoolsForToken(symbol string) ([]types.PoolPrice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	poolIDs, exists := m.tokenPools[symbol]
	if !exists || len(poolIDs) == 0 {
		return nil, fmt.Errorf("no pools found for token %s", symbol)
	}

	var result []types.PoolPrice
	for _, poolID := range poolIDs {
		if price, ok := m.poolPrices[poolID]; ok {
			result = append(result, price)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no pool prices found for token %s", symbol)
	}

	return result, nil
}

// GetLatestPoolPrices - Επιστρέφει όλες τις τελευταίες τιμές pools
func (m *MemoryStorage) GetLatestPoolPrices() ([]types.PoolPrice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]types.PoolPrice, 0, len(m.poolPrices))
	for _, price := range m.poolPrices {
		result = append(result, price)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no pool prices available")
	}

	return result, nil
}

// GetDatabaseStats - Επιστρέφει stats για το cache
func (m *MemoryStorage) GetDatabaseStats() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"storage_type":      "in-memory",
		"pools_count":       len(m.pools),
		"pool_prices_count": len(m.poolPrices),
		"tokens_count":      len(m.tokenPools),
		"last_update":       m.lastUpdate,
		"uptime_seconds":    time.Since(m.lastUpdate).Seconds(),
	}

	return stats, nil
}

// Close - No-op για in-memory storage
func (m *MemoryStorage) Close() error {
	return nil
}

// GetName - Επιστρέφει το όνομα του storage
func (m *MemoryStorage) GetName() string {
	return "in-memory-cache"
}

// Save - Legacy method (not used)
func (m *MemoryStorage) Save(tokens []types.TokenInfo) error {
	return nil
}

// SaveTokenPrices - Not implemented για τώρα
func (m *MemoryStorage) SaveTokenPrices(prices []types.TokenPrice) error {
	return nil
}

// GetLatestTokenPrices - Not implemented για in-memory (δεν χρειάζεται)
func (m *MemoryStorage) GetLatestTokenPrices() ([]types.TokenPrice, error) {
	return []types.TokenPrice{}, nil
}

// GetTokenPrice - Not implemented για in-memory
func (m *MemoryStorage) GetTokenPrice(symbol string) (*types.TokenPrice, error) {
	return nil, fmt.Errorf("not implemented for in-memory storage")
}

// GetTokenPriceFromPools - Not implemented για in-memory
func (m *MemoryStorage) GetTokenPriceFromPools(symbol string) (*types.TokenPrice, error) {
	return nil, fmt.Errorf("not implemented for in-memory storage")
}

// GetAllUniqueTokens - Επιστρέφει όλα τα unique tokens από τα pools
func (m *MemoryStorage) GetAllUniqueTokens() ([]types.TokenPrice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokenMap := make(map[string]*types.TokenPrice)

	for _, price := range m.poolPrices {
		// Token 0
		if price.Token0Symbol != "" {
			if _, exists := tokenMap[price.Token0Symbol]; !exists {
				tokenMap[price.Token0Symbol] = &types.TokenPrice{
					Symbol:    price.Token0Symbol,
					Denom:     price.Token0Denom,
					Timestamp: price.Timestamp,
				}
			}
		}
		// Token 1
		if price.Token1Symbol != "" {
			if _, exists := tokenMap[price.Token1Symbol]; !exists {
				tokenMap[price.Token1Symbol] = &types.TokenPrice{
					Symbol:    price.Token1Symbol,
					Denom:     price.Token1Denom,
					Timestamp: price.Timestamp,
				}
			}
		}
	}

	result := make([]types.TokenPrice, 0, len(tokenMap))
	for _, token := range tokenMap {
		result = append(result, *token)
	}

	return result, nil
}

// GetMemoryUsage - Επιστρέφει εκτίμηση χρήσης μνήμης
func (m *MemoryStorage) GetMemoryUsage() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Εκτίμηση: κάθε pool ~1KB, κάθε price ~500 bytes
	poolsBytes := len(m.pools) * 1024
	pricesBytes := len(m.poolPrices) * 512
	totalMB := float64(poolsBytes+pricesBytes) / 1024 / 1024

	return map[string]interface{}{
		"pools_bytes":       poolsBytes,
		"prices_bytes":      pricesBytes,
		"total_mb":          totalMB,
		"pools_count":       len(m.pools),
		"pool_prices_count": len(m.poolPrices),
	}
}
