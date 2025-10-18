package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"portofoliov1/types"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStorage - Professional SQLite storage για historical data
type SQLiteStorage struct {
	db        *sql.DB
	dbPath    string
	batchSize int
	mu        sync.RWMutex // Προστασία για concurrent access
}

// NewSQLiteStorage - Δημιουργία νέου SQLite storage
func NewSQLiteStorage(dataFolder string) (*SQLiteStorage, error) {
	// Δημιουργία φακέλου αν δεν υπάρχει
	if err := os.MkdirAll(dataFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data folder: %w", err)
	}

	dbPath := filepath.Join(dataFolder, "osmosis_history.db")

	// Άνοιγμα/Δημιουργία database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite optimizations για write performance
	pragmas := []string{
		"PRAGMA journal_mode = WAL",   // Write-Ahead Logging για καλύτερη performance
		"PRAGMA synchronous = NORMAL", // Ισορροπία μεταξύ ταχύτητας και ασφάλειας
		"PRAGMA cache_size = -64000",  // 64MB cache
		"PRAGMA temp_store = MEMORY",  // Temp tables στη μνήμη
		"PRAGMA busy_timeout = 5000",  // 5s timeout για locks
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Printf("⚠️  Warning: failed to set pragma: %s - %v", pragma, err)
		}
	}

	storage := &SQLiteStorage{
		db:        db,
		dbPath:    dbPath,
		batchSize: 1000, // Batch inserts για performance
	}

	// Δημιουργία tables
	if err := storage.createTables(); err != nil {
		return nil, err
	}

	// log.Printf("✅ SQLite database initialized: %s", dbPath) // Silent mode
	return storage, nil
}

// createTables - Δημιουργία database schema
func (s *SQLiteStorage) createTables() error {
	schema := `
	-- Pools History Table
	CREATE TABLE IF NOT EXISTS pools_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pool_id TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		type TEXT,
		total_liquidity_usd REAL,
		volume_24h REAL,
		apr REAL,
		swap_fee TEXT,
		assets_count INTEGER,
		assets_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Tokens History Table
	CREATE TABLE IF NOT EXISTS tokens_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		denom TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		price_usd REAL,
		price_osmo REAL,
		liquidity REAL,
		volume_24h REAL,
		pool_count INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Pool Prices History Table (για κάθε pool pair)
	CREATE TABLE IF NOT EXISTS pool_prices_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pool_id TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		token0_symbol TEXT,
		token0_denom TEXT,
		token0_amount TEXT,
		token1_symbol TEXT,
		token1_denom TEXT,
		token1_amount TEXT,
		price_token0_to_token1 REAL,
		price_token1_to_token0 REAL,
		liquidity_usd REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes για γρήγορες queries
	CREATE INDEX IF NOT EXISTS idx_pools_pool_id_timestamp ON pools_history(pool_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_pools_timestamp ON pools_history(timestamp DESC);
	
	CREATE INDEX IF NOT EXISTS idx_tokens_symbol_timestamp ON tokens_history(symbol, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_tokens_denom_timestamp ON tokens_history(denom, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_tokens_timestamp ON tokens_history(timestamp DESC);
	
	CREATE INDEX IF NOT EXISTS idx_pool_prices_pool_id_timestamp ON pool_prices_history(pool_id, timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_pool_prices_timestamp ON pool_prices_history(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_pool_prices_symbols ON pool_prices_history(token0_symbol, token1_symbol);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// SaveTokenPrices - Αποθήκευση token prices με timestamp
func (s *SQLiteStorage) SaveTokenPrices(prices []types.TokenPrice) error {
	if len(prices) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO tokens_history 
		(symbol, denom, timestamp, price_usd, price_osmo, liquidity, volume_24h, pool_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	timestamp := time.Now()
	for _, price := range prices {
		_, err = stmt.Exec(
			price.Symbol,
			price.Denom,
			timestamp,
			price.PriceUSD,
			price.PriceOSMO,
			0.0, // liquidity - θα προστεθεί αργότερα
			0.0, // volume_24h - θα προστεθεί αργότερα
			0,   // pool_count - θα προστεθεί αργότερα
		)
		if err != nil {
			return fmt.Errorf("failed to insert token price: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("💾 Saved %d token prices to database", len(prices))
	return nil
}

// SavePoolPrices - Αποθήκευση pool prices με timestamp
func (s *SQLiteStorage) SavePoolPrices(prices []types.PoolPrice) error {
	if len(prices) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO pool_prices_history 
		(pool_id, timestamp, token0_symbol, token0_denom, token0_amount, 
		 token1_symbol, token1_denom, token1_amount, 
		 price_token0_to_token1, price_token1_to_token0, liquidity_usd)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	timestamp := time.Now()
	for _, price := range prices {
		_, err = stmt.Exec(
			price.PoolID,
			timestamp,
			price.Token0Symbol,
			price.Token0Denom,
			price.Token0Amount,
			price.Token1Symbol,
			price.Token1Denom,
			price.Token1Amount,
			price.PriceToken0ToToken1,
			price.PriceToken1ToToken0,
			price.LiquidityUSD,
		)
		if err != nil {
			return fmt.Errorf("failed to insert pool price: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// log.Printf("💾 Saved %d pool prices to database", len(prices))
	return nil
}

// SavePools - Αποθήκευση ΟΛΩΝ των pools (raw data από Osmosis API)
func (s *SQLiteStorage) SavePools(pools []types.OsmosisPool) error {
	if len(pools) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO pools_history 
		(pool_id, timestamp, type, total_liquidity_usd, volume_24h, apr, swap_fee, assets_count, assets_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	timestamp := time.Now()
	for _, pool := range pools {
		// Serialize pool assets to JSON για να τα αποθηκεύσουμε
		assetsJSON := ""
		if len(pool.PoolAssets) > 0 {
			// Δημιουργία απλής JSON representation
			assetsStr := "["
			for i, asset := range pool.PoolAssets {
				if i > 0 {
					assetsStr += ","
				}
				assetsStr += fmt.Sprintf(`{"token":"%s","weight":"%s","amount":"%s"}`,
					asset.Token.Denom, asset.Weight, asset.Token.Amount)
			}
			assetsStr += "]"
			assetsJSON = assetsStr
		}

		_, err = stmt.Exec(
			pool.Id,
			timestamp,
			pool.Type,
			0.0, // total_liquidity_usd - θα υπολογιστεί αν χρειαστεί
			0.0, // volume_24h - δεν το έχουμε από το API
			0.0, // apr - δεν το έχουμε από το API
			pool.PoolParams.SwapFee,
			len(pool.PoolAssets),
			assetsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert pool: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// log.Printf("💾 Saved %d pools to database", len(pools))
	return nil
}

// GetTokenPriceHistory - Ανάκτηση historical prices για ένα token
func (s *SQLiteStorage) GetTokenPriceHistory(symbol string, from, to time.Time) ([]types.TokenPrice, error) {
	query := `
		SELECT symbol, denom, timestamp, price_usd, price_osmo
		FROM tokens_history
		WHERE symbol = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, symbol, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query token history: %w", err)
	}
	defer rows.Close()

	var prices []types.TokenPrice
	for rows.Next() {
		var price types.TokenPrice
		err := rows.Scan(
			&price.Symbol,
			&price.Denom,
			&price.Timestamp,
			&price.PriceUSD,
			&price.PriceOSMO,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// GetPoolPriceHistory - Ανάκτηση historical prices για ένα pool
func (s *SQLiteStorage) GetPoolPriceHistory(poolID string, from, to time.Time) ([]types.PoolPrice, error) {
	query := `
		SELECT pool_id, timestamp, token0_symbol, token0_denom, token0_amount,
		       token1_symbol, token1_denom, token1_amount,
		       price_token0_to_token1, price_token1_to_token0, liquidity_usd
		FROM pool_prices_history
		WHERE pool_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, poolID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query pool history: %w", err)
	}
	defer rows.Close()

	var prices []types.PoolPrice
	for rows.Next() {
		var price types.PoolPrice
		err := rows.Scan(
			&price.PoolID,
			&price.Timestamp,
			&price.Token0Symbol,
			&price.Token0Denom,
			&price.Token0Amount,
			&price.Token1Symbol,
			&price.Token1Denom,
			&price.Token1Amount,
			&price.PriceToken0ToToken1,
			&price.PriceToken1ToToken0,
			&price.LiquidityUSD,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// GetLatestTokenPrices - Οι πιο πρόσφατες τιμές tokens
func (s *SQLiteStorage) GetLatestTokenPrices() ([]types.TokenPrice, error) {
	query := `
		SELECT DISTINCT ON (symbol) 
			symbol, denom, timestamp, price_usd, price_osmo
		FROM tokens_history
		ORDER BY symbol, timestamp DESC
	`

	// SQLite doesn't support DISTINCT ON, so we use a workaround
	query = `
		SELECT symbol, denom, timestamp, price_usd, price_osmo
		FROM tokens_history t1
		WHERE timestamp = (
			SELECT MAX(timestamp) 
			FROM tokens_history t2 
			WHERE t2.symbol = t1.symbol
		)
		GROUP BY symbol
		ORDER BY timestamp DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest prices: %w", err)
	}
	defer rows.Close()

	var prices []types.TokenPrice
	for rows.Next() {
		var price types.TokenPrice
		err := rows.Scan(
			&price.Symbol,
			&price.Denom,
			&price.Timestamp,
			&price.PriceUSD,
			&price.PriceOSMO,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// GetAllUniqueTokens - Επιστροφή όλων των unique tokens που έχουν αποθηκευτεί στη database
func (s *SQLiteStorage) GetAllUniqueTokens() ([]types.TokenPrice, error) {
	query := `
		SELECT 
			symbol,
			denom,
			COUNT(*) as record_count,
			MIN(timestamp) as first_seen,
			MAX(timestamp) as last_seen,
			MAX(price_usd) as latest_price_usd,
			MAX(price_osmo) as latest_price_osmo
		FROM tokens_history
		GROUP BY symbol
		ORDER BY last_seen DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unique tokens: %w", err)
	}
	defer rows.Close()

	var tokens []types.TokenPrice
	for rows.Next() {
		var token types.TokenPrice
		var recordCount int
		var firstSeen, lastSeen time.Time

		err := rows.Scan(
			&token.Symbol,
			&token.Denom,
			&recordCount,
			&firstSeen,
			&lastSeen,
			&token.PriceUSD,
			&token.PriceOSMO,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Χρησιμοποιούμε το last_seen ως timestamp
		token.Timestamp = lastSeen

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// GetTokenPrice - Επιστροφή τιμής συγκεκριμένου token (latest)
func (s *SQLiteStorage) GetTokenPrice(symbol string) (*types.TokenPrice, error) {
	query := `
		SELECT symbol, denom, timestamp, price_usd, price_osmo
		FROM tokens_history
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var price types.TokenPrice
	err := s.db.QueryRow(query, symbol).Scan(
		&price.Symbol,
		&price.Denom,
		&price.Timestamp,
		&price.PriceUSD,
		&price.PriceOSMO,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token %s not found", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query token price: %w", err)
	}

	return &price, nil
}

// GetLatestPoolPrices - Επιστροφή των πιο πρόσφατων pool prices
func (s *SQLiteStorage) GetLatestPoolPrices() ([]types.PoolPrice, error) {
	query := `
		SELECT pool_id, timestamp, token0_symbol, token0_denom, token0_amount,
		       token1_symbol, token1_denom, token1_amount,
		       price_token0_to_token1, price_token1_to_token0, liquidity_usd
		FROM pool_prices_history p1
		WHERE timestamp = (
			SELECT MAX(timestamp)
			FROM pool_prices_history p2
			WHERE p2.pool_id = p1.pool_id
		)
		ORDER BY pool_id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest pool prices: %w", err)
	}
	defer rows.Close()

	var prices []types.PoolPrice
	for rows.Next() {
		var price types.PoolPrice
		err := rows.Scan(
			&price.PoolID,
			&price.Timestamp,
			&price.Token0Symbol,
			&price.Token0Denom,
			&price.Token0Amount,
			&price.Token1Symbol,
			&price.Token1Denom,
			&price.Token1Amount,
			&price.PriceToken0ToToken1,
			&price.PriceToken1ToToken0,
			&price.LiquidityUSD,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// GetTokenPriceFromPools - Υπολογισμός τιμής token από pool prices (OPTIMIZED)
func (s *SQLiteStorage) GetTokenPriceFromPools(symbol string) (*types.TokenPrice, error) {
	// OPTIMIZED: Παίρνουμε μόνο το latest global timestamp και φιλτράρουμε pools με αυτό
	// Πολύ πιο γρήγορο από JOIN!
	query := `
		WITH latest_ts AS (
			SELECT MAX(timestamp) as ts FROM pool_prices_history
		)
		SELECT 
			pool_id,
			token0_symbol,
			token0_denom,
			token1_symbol,
			token1_denom,
			price_token0_to_token1,
			price_token1_to_token0,
			liquidity_usd,
			timestamp
		FROM pool_prices_history
		WHERE (token0_symbol = ? OR token1_symbol = ?)
		AND timestamp = (SELECT ts FROM latest_ts)
		ORDER BY liquidity_usd DESC
		LIMIT 1
	`

	var p struct {
		poolID              int64
		token0Symbol        string
		token0Denom         string
		token1Symbol        string
		token1Denom         string
		priceToken0ToToken1 float64
		priceToken1ToToken0 float64
		liquidityUSD        float64
		timestamp           time.Time
	}

	err := s.db.QueryRow(query, symbol, symbol).Scan(
		&p.poolID,
		&p.token0Symbol,
		&p.token0Denom,
		&p.token1Symbol,
		&p.token1Denom,
		&p.priceToken0ToToken1,
		&p.priceToken1ToToken0,
		&p.liquidityUSD,
		&p.timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no pools found for token %s", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query pool for token %s: %w", symbol, err)
	}

	// Καθορισμός ποιο token ζητήθηκε και υπολογισμός τιμής
	var denom string
	var priceInOtherToken float64

	if p.token0Symbol == symbol {
		denom = p.token0Denom
		priceInOtherToken = p.priceToken0ToToken1
	} else {
		denom = p.token1Denom
		priceInOtherToken = p.priceToken1ToToken0
	}

	return &types.TokenPrice{
		Symbol:    symbol,
		Denom:     denom,
		PriceUSD:  0,                 // Θα το υπολογίσουμε αν ξέρουμε την τιμή του other token
		PriceOSMO: priceInOtherToken, // Τιμή σε σχέση με το άλλο token
		Timestamp: p.timestamp,
	}, nil
}

// GetAllPoolsForToken - Επιστρέφει όλα τα pools που περιέχουν ένα token
func (s *SQLiteStorage) GetAllPoolsForToken(symbol string) ([]types.PoolPrice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		WITH latest_ts AS (
			SELECT MAX(timestamp) as ts FROM pool_prices_history
		)
		SELECT 
			pool_id,
			token0_symbol,
			token0_denom,
			token0_amount,
			token1_symbol,
			token1_denom,
			token1_amount,
			price_token0_to_token1,
			price_token1_to_token0,
			liquidity_usd,
			timestamp
		FROM pool_prices_history
		WHERE (token0_symbol = ? OR token1_symbol = ?)
		AND timestamp = (SELECT ts FROM latest_ts)
		ORDER BY liquidity_usd DESC
	`

	rows, err := s.db.Query(query, symbol, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query pools for token %s: %w", symbol, err)
	}
	defer rows.Close()

	var pools []types.PoolPrice
	for rows.Next() {
		var p types.PoolPrice
		err := rows.Scan(
			&p.PoolID,
			&p.Token0Symbol,
			&p.Token0Denom,
			&p.Token0Amount,
			&p.Token1Symbol,
			&p.Token1Denom,
			&p.Token1Amount,
			&p.PriceToken0ToToken1,
			&p.PriceToken1ToToken0,
			&p.LiquidityUSD,
			&p.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pool: %w", err)
		}
		pools = append(pools, p)
	}

	if len(pools) == 0 {
		return nil, fmt.Errorf("no pools found for token %s", symbol)
	}

	return pools, nil
}

// GetDatabaseStats - Στατιστικά του database
func (s *SQLiteStorage) GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count records
	var tokenCount, poolPriceCount int64
	var oldestTimestamp, newestTimestamp time.Time

	// Token counts
	err := s.db.QueryRow("SELECT COUNT(*) FROM tokens_history").Scan(&tokenCount)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM pool_prices_history").Scan(&poolPriceCount)
	if err != nil {
		return nil, err
	}

	// Time range - Παίρνουμε από pool_prices_history (όχι tokens_history που είναι άδειο)
	err = s.db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM pool_prices_history").Scan(&oldestTimestamp, &newestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		// Fallback σε tokens_history αν υπάρχει
		err = s.db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM tokens_history").Scan(&oldestTimestamp, &newestTimestamp)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}

	// Database file size
	fileInfo, err := os.Stat(s.dbPath)
	if err != nil {
		return nil, err
	}

	// Count unique tokens and pools
	var uniqueTokens, uniquePools int64
	s.db.QueryRow("SELECT COUNT(DISTINCT token0_symbol) + COUNT(DISTINCT token1_symbol) FROM pool_prices_history").Scan(&uniqueTokens)
	s.db.QueryRow("SELECT COUNT(DISTINCT pool_id) FROM pool_prices_history").Scan(&uniquePools)

	stats["token_records"] = tokenCount
	stats["pool_price_records"] = poolPriceCount
	stats["unique_tokens"] = uniqueTokens
	stats["unique_pools"] = uniquePools
	stats["oldest_record"] = oldestTimestamp
	stats["latest_record"] = newestTimestamp
	stats["database_size"] = fmt.Sprintf("%.2f MB", float64(fileInfo.Size())/1024/1024)
	stats["database_path"] = s.dbPath
	stats["total_records"] = tokenCount + poolPriceCount

	return stats, nil
}

// Close - Κλείσιμο database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// GetName - Implementation του StorageInterface
func (s *SQLiteStorage) GetName() string {
	return "SQLite"
}

// Save - Implementation του StorageInterface (για backward compatibility)
func (s *SQLiteStorage) Save(tokens []types.TokenInfo) error {
	// Convert TokenInfo to TokenPrice για compatibility
	prices := make([]types.TokenPrice, len(tokens))
	for i, token := range tokens {
		prices[i] = types.TokenPrice{
			Symbol:    token.Symbol,
			Denom:     token.Denom,
			PriceUSD:  token.Price,
			Timestamp: time.Now(),
		}
	}
	return s.SaveTokenPrices(prices)
}
