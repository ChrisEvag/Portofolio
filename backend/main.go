package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"portofoliov1/api"
	"portofoliov1/storage"
	"portofoliov1/types"
	"portofoliov1/utils"
)

type Config struct {
	DisplayLimit   int
	RequestTimeout time.Duration
	RefreshMinutes time.Duration
	StorageType    string
	DataFolder     string
	Chains         []string // ["osmosis"]
}

var config = Config{
	DisplayLimit:   25,
	RequestTimeout: 30 * time.Second,
	RefreshMinutes: 1 * time.Second,     // ⚡ REAL-TIME: Ανανέωση κάθε 1 δευτερόλεπτο
	StorageType:    "memory",            // 💾 In-Memory Cache (No persistence)
	DataFolder:     "data/database",     // Not used for memory storage
	Chains:         []string{"osmosis"}, // Αλυσίδες που θα παρακολουθούμε
}

func main() {
	// Initialize chain registry updater (1 φορά την εβδομάδα)
	chainRegistryUpdater := utils.NewChainRegistryUpdater()
	if err := chainRegistryUpdater.Start(); err != nil {
		log.Printf("⚠️  Προειδοποίηση: Αποτυχία εκκίνησης chain registry updater: %v", err)
	}
	defer chainRegistryUpdater.Stop()

	// Initialize asset service from chain registry
	assetService, err := types.NewAssetService()
	if err != nil {
		log.Fatalf("❌ Σφάλμα αρχικοποίησης AssetService: %v", err)
	}

	// Initialize In-Memory storage (για real-time data χωρίς persistence)
	memoryStorage := storage.NewMemoryStorage()
	log.Println("✅ In-Memory cache initialized")

	// Initialize HTTP server με access στο memory cache
	httpServer := api.NewHTTPServer(8080, chainRegistryUpdater, memoryStorage)

	// Start HTTP server σε ξεχωριστό goroutine
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Σφάλμα HTTP server: %v", err)
		}
	}()

	showWelcomeMessage()

	if config.RefreshMinutes > 0 {
		startAutoRefresh(assetService, httpServer, memoryStorage)
	} else {
		runSingleExecution(assetService, httpServer, memoryStorage)
	}
}

func showWelcomeMessage() {
	fmt.Println("🚀 Professional Osmosis Data Collector")
	fmt.Printf("💾 Storage: In-Memory Cache (Real-time)\n")
	fmt.Printf("⛓️  Chains: %v\n", config.Chains)
	fmt.Println("⚡ Update Interval: 1 SECOND (Real-time)")
	fmt.Println("================================")
}

func runSingleExecution(assetService *types.AssetService, httpServer *api.HTTPServer, memoryStorage *storage.MemoryStorage) {
	// Εκτέλεση για κάθε αλυσίδα
	for _, chain := range config.Chains {
		// fmt.Printf("\n🎯 ΕΠΕΞΕΡΓΑΣΙΑ ΑΛΥΣΙΔΑΣ: %s\n", strings.ToUpper(chain))
		// fmt.Println("------------------------------")

		_, err := fetchChainData(chain, assetService, httpServer, memoryStorage)
		if err != nil {
			log.Printf("❌ Σφάλμα για %s: %v", chain, err)
			continue
		}
	}
}

func fetchChainData(chain string, assetService *types.AssetService, httpServer *api.HTTPServer, memoryStorage *storage.MemoryStorage) ([]types.TokenInfo, error) {
	switch chain {
	case "osmosis":
		return fetchOsmosisData(assetService, httpServer, memoryStorage)
	default:
		return nil, fmt.Errorf("μη υποστηριζόμενη αλυσίδα: %s", chain)
	}
}

func fetchOsmosisData(assetService *types.AssetService, httpServer *api.HTTPServer, memoryStorage *storage.MemoryStorage) ([]types.TokenInfo, error) {
	// Αρχικοποίηση του νέου Osmosis Pool Client
	osmosisClient := api.NewOsmosisPoolClient()

	// 1. Λήψη pools
	pools, err := osmosisClient.GetAllPools(1000, 0)
	if err != nil {
		return nil, err
	}

	// 2. Υπολογισμός τιμών για ΟΛΟΥΣ τους pools (στη μνήμη)
	poolPrices, err := osmosisClient.GetAllPoolPrices(pools, assetService)
	if err != nil {
		fmt.Printf("   ⚠️  Προειδοποίηση: αποτυχία υπολογισμού τιμών pools: %v\n", err)
		poolPrices = []types.PoolPrice{}
	}

	// 3. ⚡ ΑΠΟΘΗΚΕΥΣΗ ΣΕ MEMORY CACHE (Real-time)
	// Αποθήκευση ΟΛΩΝ των pools (raw data)
	if err := memoryStorage.SavePools(pools); err != nil {
		log.Printf("❌ Failed to save pools: %v", err)
	}

	// Αποθήκευση pool prices (υπολογισμένες τιμές μεταξύ tokens)
	if err := memoryStorage.SavePoolPrices(poolPrices); err != nil {
		log.Printf("❌ Failed to save pool prices: %v", err)
	}

	// Silent mode - μόνο errors

	// Επιστρέφουμε κενό slice καθώς δε χρειαζόμαστε πλέον τα token infos
	return []types.TokenInfo{}, nil
}

// startAutoRefresh - Αρχή auto-refresh λειτουργίας
func startAutoRefresh(assetService *types.AssetService, httpServer *api.HTTPServer, memoryStorage *storage.MemoryStorage) {
	fmt.Printf("⚡ Real-Time Mode - Update κάθε %v\n", config.RefreshMinutes)
	fmt.Printf("💾 Storage: In-Memory Cache (No persistence)\n")
	fmt.Println("🌐 API: http://localhost:8080")
	fmt.Println("📊 Cache ανανεώνεται κάθε δευτερόλεπτο...")
	fmt.Println("   Πατήστε Ctrl+C για διακοπή")
	fmt.Println()

	// Τρέχει αμέσως την πρώτη φορά
	runSingleExecution(assetService, httpServer, memoryStorage)

	// Δημιουργία ticker για auto-refresh
	ticker := time.NewTicker(config.RefreshMinutes)
	defer ticker.Stop()

	executionCount := 1

	for range ticker.C {
		executionCount++
		runSingleExecution(assetService, httpServer, memoryStorage)

		// Κάθε 60 δευτερόλεπτα δείχνε stats
		if executionCount%60 == 0 {
			fmt.Printf("\n📊 [%d cache updates] - %s\n",
				executionCount, time.Now().Format("15:04:05"))
		}
	}
}
