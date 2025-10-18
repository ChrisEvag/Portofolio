# 🚀 Osmosis Data Collector & Portfolio Tracker

Professional real-time data collector for Osmosis blockchain with SQLite storage and REST API.

## 📋 Features

- ⚡ **Real-time Data Collection** - Collects pool data every second
- 💾 **SQLite Storage** - Historical data storage with WAL mode for performance
- 🌐 **REST API** - Query tokens, pools, and prices
- 📊 **1,894 Records/Second** - 1000 pools + 894 pool prices per cycle
- 🔒 **Thread-Safe** - Mutex protection for concurrent access
- 🎯 **Chain Registry Integration** - Automatic token metadata updates

## 🛠️ Installation

### Prerequisites

- Go 1.21 or higher
- No CGO required (uses pure Go SQLite driver)

### Setup

```bash
# Clone the repository
git clone <your-repo-url>
cd Portofolio/backend

# Install dependencies
go mod download

# Run the application
go run main.go
```

The database will be automatically created on first run.

## 🚀 Usage

### Running the Application

```bash
cd backend
go run main.go
```

Or build and run the executable:

```bash
go build -o osmosis-tracker.exe .
.\osmosis-tracker.exe
```

### API Endpoints

Once running, the API is available at `http://localhost:8080`:

#### Get Token Pools
```bash
GET /api/tokens/{SYMBOL}/pools
```
Returns all pools containing the specified token with real-time prices.

**Example:**
```bash
curl http://localhost:8080/api/tokens/ATOM/pools
```

**Response:**
```json
{
  "symbol": "ATOM",
  "pools": [
    {
      "pool_id": "1",
      "paired_with": "OSMO",
      "paired_denom": "uosmo",
      "token_price": 26.251288,
      "inverse_price": 0.038093,
      "liquidity_usd": 0,
      "timestamp": "2025-10-18T23:48:20Z"
    }
  ],
  "count": 312,
  "latest_update": "2025-10-18T23:48:20Z"
}
```

#### Get All Pool Prices
```bash
GET /api/pools
```
Returns all latest pool prices.

#### Get All Tokens
```bash
GET /api/tokens
```
Returns list of all available tokens.

#### Health Check
```bash
GET /api/health
```
Returns database statistics.

#### Chain Registry Update
```bash
POST /api/chain-registry/update
GET /api/chain-registry/status
```

## 📁 Project Structure

```
backend/
├── main.go                 # Main application entry point
├── api/
│   ├── http_server.go     # REST API server
│   └── osmosis_pool_client.go  # Osmosis API client
├── storage/
│   ├── sqlite_storage.go  # SQLite operations
│   └── storage.go         # Storage interface
├── types/
│   ├── asset_service.go   # Token metadata service
│   ├── pool_types.go      # Pool data structures
│   └── price_types.go     # Price data structures
├── utils/
│   └── chain_registry_updater.go  # Auto-update chain registry
└── data/
    ├── database/          # SQLite database files (not in Git)
    └── chain-registry/    # Token metadata from chain registry
```

## 💾 Database

### Important Notes

⚠️ **Database files are NOT included in Git** due to their size (can grow to GBs).

### Database Location
```
backend/data/database/osmosis_history.db
```

### Tables
- `pools_history` - Raw pool data from Osmosis API
- `pool_prices_history` - Calculated prices between token pairs
- `tokens_history` - Token price history (future use)

### Database Growth
- **Per hour**: ~50-100 MB (1-second intervals)
- **Per day**: ~1-2 GB (continuous collection)

### Reset Database
```bash
# Stop the application first (Ctrl+C)
Remove-Item backend/data/database/*.db*
```

## 🔧 Configuration

Edit `main.go` to configure:

```go
var config = Config{
    DisplayLimit:   25,
    RequestTimeout: 30 * time.Second,
    RefreshMinutes: 1 * time.Second,      // Collection interval
    StorageType:    "sqlite",
    DataFolder:     "data/database",
    Chains:         []string{"osmosis"},   // Chains to monitor
}
```

## 📊 Performance

- **Update Interval**: ~1-2 seconds (not exactly 1s due to API call time)
- **Records per Cycle**: 1,894 (1000 pools + 894 pool prices)
- **API Response Time**: <100ms (with SQLite indexes)
- **Concurrent Safety**: Mutex-protected database access

## 🐛 Troubleshooting

### Port Already in Use
If port 8080 is busy:
```go
// In main.go, change:
httpServer := api.NewHTTPServer(8080, ...)
// To:
httpServer := api.NewHTTPServer(8081, ...)
```

### Database Locked
If you get "database is locked" errors:
1. Stop all running instances
2. Delete `.db-shm` and `.db-wal` files
3. Restart the application

### Out of Memory
For long-running instances, consider:
1. Implementing data retention policies
2. Archiving old data periodically
3. Using VACUUM to reclaim space

## 📝 TODO

- [ ] Analytics API endpoints (`/api/history/*`, `/api/analytics/*`)
- [ ] Connection pooling optimization
- [ ] Data retention policies
- [ ] Historical data export
- [ ] Multi-chain support expansion
- [ ] WebSocket support for real-time updates

## 📄 License

MIT License

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first.

