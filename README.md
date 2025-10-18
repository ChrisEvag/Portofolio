# 🚀 Osmosis Data Collector & Portfolio Tracker

Professional real-time data collector for Osmosis blockchain with in-memory cache and REST API.

## 📋 Features

- ⚡ **Real-time Data Collection** - Collects pool data every 1-2 seconds
- 💾 **In-Memory Cache** - Ultra-fast storage with zero disk I/O
- 🌐 **REST API** - Query tokens, pools, and prices instantly
- 📊 **1,894 Records/Second** - 1000 pools + 894 pool prices per cycle
- 🔒 **Thread-Safe** - Mutex protection for concurrent access
- 🎯 **Chain Registry Integration** - Automatic token metadata updates
- ⚡ **No Persistence** - Pure in-memory for maximum speed

## 🛠️ Installation

### Prerequisites

- Go 1.21 or higher
- No external dependencies (pure in-memory storage)

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

The in-memory cache is initialized automatically on startup.

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
│   ├── memory_storage.go  # In-memory cache operations
│   └── storage.go         # Storage interface
├── types/
│   ├── asset_service.go   # Token metadata service
│   ├── pool_types.go      # Pool data structures
│   └── price_types.go     # Price data structures
├── utils/
│   └── chain_registry_updater.go  # Auto-update chain registry
└── data/
    └── chain-registry/    # Token metadata from chain registry
```

## 💾 Storage

### In-Memory Cache

⚡ **All data is stored in memory** - No database files, no persistence.

**Advantages:**
- Ultra-fast API responses (<5ms)
- Zero disk I/O overhead
- No database locks or crashes
- Minimal memory footprint (~2 MB)

**Trade-offs:**
- Data is lost on restart (by design)
- Only latest prices are kept
- Perfect for real-time trading applications

### Memory Usage
- **Pools**: ~1 KB per pool × 1000 = ~1 MB
- **Pool Prices**: ~500 bytes per price × 894 = ~500 KB
- **Total**: ~2 MB (very lightweight!)

## 🔧 Configuration

Edit `main.go` to configure:

```go
var config = Config{
    DisplayLimit:   25,
    RequestTimeout: 30 * time.Second,
    RefreshMinutes: 1 * time.Second,      // Collection interval
    StorageType:    "memory",             // In-memory cache
    Chains:         []string{"osmosis"},   // Chains to monitor
}
```

## 📊 Performance

- **Update Interval**: ~1-2 seconds (API fetch + calculation time)
- **Records per Cycle**: 1,894 (1000 pools + 894 pool prices)
- **API Response Time**: <5ms (in-memory reads)
- **Memory Usage**: ~2 MB (stable)
- **Concurrent Safety**: Thread-safe with RWMutex

## 🐛 Troubleshooting

### Port Already in Use
If port 8080 is busy:
```go
// In main.go, change:
httpServer := api.NewHTTPServer(8080, ...)
// To:
httpServer := api.NewHTTPServer(8081, ...)
```

### Memory Usage
The application uses ~2 MB of memory. If you're concerned about memory:
- The cache only keeps latest data (1 snapshot)
- Memory usage is stable and predictable
- No memory leaks (Go's GC handles cleanup)

## 📝 TODO

- [ ] Swap simulation endpoints (`/api/swap/simulate`)
- [ ] Best route calculation for multi-hop swaps
- [ ] Wallet integration (Keplr support)
- [ ] Frontend UI (React/Next.js)
- [ ] WebSocket support for real-time push updates
- [ ] Optional: Add historical data persistence (if needed later)

## 📄 License

MIT License

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first.

