package main

import (
	"fmt"
	"log"
	"strings"
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
	Chains         []string // ["osmosis", "dydx"]
}

var config = Config{
	DisplayLimit:   25,
	RequestTimeout: 30 * time.Second,
	RefreshMinutes: 100000 * time.Millisecond,
	StorageType:    "csv",
	DataFolder:     "data/crypto-tokens",
	Chains:         []string{"osmosis", "dydx"}, // Αλυσίδες που θα παρακολουθούμε
}

// StorageManager - Διαχειριστής αποθήκευσης
type StorageManager struct {
	storage storage.StorageInterface
}

func NewStorageManager(storageType, dataFolder string) (*StorageManager, error) {
	var store storage.StorageInterface

	switch storageType {
	case "csv":
		store = storage.NewCSVStorage(dataFolder)
	default:
		return nil, fmt.Errorf("μη υποστηριζόμενος τύπος αποθήκευσης: %s", storageType)
	}

	return &StorageManager{
		storage: store,
	}, nil
}

func (sm *StorageManager) Save(tokens []types.TokenInfo) error {
	return sm.storage.Save(tokens)
}

func (sm *StorageManager) GetStorageName() string {
	return sm.storage.GetName()
}

func main() {
	storageManager, err := NewStorageManager(config.StorageType, config.DataFolder)
	if err != nil {
		log.Fatalf("❌ Σφάλμα αρχικοποίησης αποθήκευσης: %v", err)
	}

	showWelcomeMessage(storageManager.GetStorageName())

	if config.RefreshMinutes > 0 {
		startAutoRefresh(storageManager)
	} else {
		runSingleExecution(storageManager)
	}
}

func showWelcomeMessage(storageType string) {
	fmt.Println("🚀 Multi-Chain Portfolio Tracker")
	fmt.Printf("📁 Τύπος Αποθήκευσης: %s\n", storageType)
	fmt.Printf("⛓️  Αλυσίδες: %v\n", config.Chains)
	fmt.Println("================================")
	fmt.Println("🔄 Λήψη tokens από όλες τις αλυσίδες...")
	fmt.Println()
}

func runSingleExecution(storageManager *StorageManager) {
	var allTokens []types.TokenInfo

	// Αρχικοποίηση client
	client := api.NewAPIClient()

	// Εκτέλεση για κάθε αλυσίδα
	for _, chain := range config.Chains {
		fmt.Printf("\n🎯 ΕΠΕΞΕΡΓΑΣΙΑ ΑΛΥΣΙΔΑΣ: %s\n", strings.ToUpper(chain))
		fmt.Println("------------------------------")

		tokens, err := fetchChainData(client, chain)
		if err != nil {
			log.Printf("❌ Σφάλμα για %s: %v", chain, err)
			continue
		}

		allTokens = append(allTokens, tokens...)
	}

	// Αποθήκευση όλων των tokens
	if len(allTokens) > 0 {
		if err := storageManager.Save(allTokens); err != nil {
			log.Printf("❌ Σφάλμα αποθήκευσης: %v", err)
		} else {
			showCompletionMessage(allTokens, storageManager.GetStorageName())
		}
	}
}

func fetchChainData(client *api.APIClient, chain string) ([]types.TokenInfo, error) {
	switch chain {
	case "osmosis":
		return fetchOsmosisData(client)
	case "dydx":
		return fetchDydxData(client)
	default:
		return nil, fmt.Errorf("μη υποστηριζόμενη αλυσίδα: %s", chain)
	}
}

func fetchOsmosisData(client *api.APIClient) ([]types.TokenInfo, error) {
	// Ταχυμέτρηση endpoints
	client.SpeedTestEndpoints("osmosis")

	// Πρώτα δοκιμάζουμε Numia API
	tokens, blockHeight, err := tryNumiaAPI(client)
	if err == nil {
		// Προσθήκη chain info
		for i := range tokens {
			tokens[i].Chain = "osmosis"
		}
		return tokens, nil
	}

	// Fallback σε LCD API
	fmt.Println("2. 🔄 ΕΦΕΔΡΙΚΗ: Speed-Optimized Fallback System...")

	pools, err := client.GetPoolsWithFallback()
	if err != nil {
		return nil, err
	}

	blockHeight, err = getCurrentBlockHeight(client, "osmosis")
	if err != nil {
		return nil, fmt.Errorf("δεν μπόρεσε να ληφθεί block height: %w", err)
	}

	fmt.Printf("   ✅ Βρέθηκαν %d pools από block #%d\n", len(pools), blockHeight)

	// Εξαγωγή tokens
	tokensMap := utils.ExtractTokensFromPools(pools)
	tokensWithPrices := utils.GetTokenPrices(tokensMap)

	// Προσθήκη metadata
	for i := range tokensWithPrices {
		tokensWithPrices[i].Source = "LCD API"
		tokensWithPrices[i].BlockHeight = blockHeight
		tokensWithPrices[i].Chain = "osmosis"
	}

	return tokensWithPrices, nil
}

func fetchDydxData(client *api.APIClient) ([]types.TokenInfo, error) {
	// Ταχυμέτρηση endpoints
	client.SpeedTestEndpoints("dydx")

	fmt.Println("1. 🔍 Λήψη δεδομένων από dYdX API...")

	markets, err := client.GetDydxMarketsWithFallback()
	if err != nil {
		return nil, err
	}

	blockHeight, err := getCurrentBlockHeight(client, "dydx")
	if err != nil {
		fmt.Printf("   ⚠️  Προειδοποίηση: %v\n", err)
		blockHeight = 0
	}

	fmt.Printf("   ✅ Βρέθηκαν %d markets από block #%d\n", len(markets), blockHeight)

	// Εξαγωγή tokens από markets
	tokensMap := utils.ExtractDydxTokensFromMarkets(markets)
	tokensWithPrices := utils.GetTokenPrices(tokensMap)

	// Προσθήκη metadata
	for i := range tokensWithPrices {
		tokensWithPrices[i].Source = "dYdX API"
		tokensWithPrices[i].BlockHeight = blockHeight
		tokensWithPrices[i].Chain = "dydx"
	}

	return tokensWithPrices, nil
}

// tryNumiaAPI - Προσπάθεια λήψης δεδομένων από Numia API
func tryNumiaAPI(client *api.APIClient) ([]types.TokenInfo, int64, error) {
	fmt.Println("1. 🔍 ΠΡΩΤΕΥΟΝ: Δοκιμή Numia API...")

	tokens, err := client.GetNumiaTokens()
	if err != nil {
		fmt.Printf("   ❌ Numia API απέτυχε: %v\n\n", err)
		return nil, 0, err
	}

	fmt.Printf("   ✅ Numia API: Βρέθηκαν %d tokens\n", len(tokens))

	// Λήψη block height για Numia data
	blockHeight, err := getCurrentBlockHeight(client, "osmosis")
	if err != nil {
		fmt.Printf("   ⚠️  Προειδοποίηση: %v\n", err)
		blockHeight = 0 // Default value αν αποτύχει
	}

	return tokens, blockHeight, nil
}

func getCurrentBlockHeight(client *api.APIClient, chain string) (int64, error) {
	var endpoints []api.EndpointInfo

	switch chain {
	case "osmosis":
		endpoints = client.LCDEndpoints
	case "dydx":
		endpoints = client.DydxEndpoints
	}

	for _, endpoint := range endpoints {
		if endpoint.Working {
			height, err := client.GetLatestBlockHeight(endpoint.URL, chain)
			if err == nil {
				fmt.Printf("   📦 Τρέχον Block Height: %d\n", height)
				return height, nil
			}
		}
	}
	return 0, fmt.Errorf("δεν μπόρεσε να ληφθεί block height")
}

// startAutoRefresh - Αρχή auto-refresh λειτουργίας
func startAutoRefresh(storageManager *StorageManager) {
	fmt.Printf("🔄 Λειτουργία Auto-Refresh - Ανανέωση κάθε %v\n", config.RefreshMinutes)
	fmt.Printf("📁 Αποθήκευση σε: %s\n", storageManager.GetStorageName())
	fmt.Println("   Πατήστε Ctrl+C για διακοπή")
	fmt.Println()

	// Τρέχει αμέσως την πρώτη φορά
	runSingleExecution(storageManager)

	// Δημιουργία ticker για auto-refresh
	ticker := time.NewTicker(config.RefreshMinutes)
	defer ticker.Stop()

	executionCount := 1

	for range ticker.C {
		executionCount++
		fmt.Printf("\n" + repeatString(50, "="))
		fmt.Printf("\n🔄 ΑΥΤΟΜΑΤΗ ΑΝΑΝΕΩΣΗ #%d - %s\n",
			executionCount, time.Now().Format("02/01/2006 15:04:05"))
		fmt.Println(repeatString(50, "="))

		runSingleExecution(storageManager)

		nextRun := time.Now().Add(config.RefreshMinutes)
		fmt.Printf("\n⏰ Επόμενη ανανέωση: %s\n", nextRun.Format("15:04:05"))
	}
}

// showCompletionMessage - Εμφάνιση μηνύματος ολοκλήρωσης
func showCompletionMessage(tokens []types.TokenInfo, storageType string) {
	tokenCount := len(tokens)
	var blockHeight int64
	if tokenCount > 0 {
		blockHeight = tokens[0].BlockHeight
	}

	fmt.Println("\n🎯 ΟΛΟΚΛΗΡΩΣΗ ΕΠΙΤΥΧΗΣ!")
	fmt.Println("=======================")
	fmt.Printf("✅ Λήφθηκαν %d tokens\n", tokenCount)
	fmt.Printf("📦 Block Height: #%d\n", blockHeight)
	fmt.Printf("💾 Αποθήκευση σε: %s\n", storageType)
	fmt.Printf("🕐 Χρόνος εκτέλεσης: %s\n", time.Now().Format("15:04:05"))
}

// Βοηθητικές συναρτήσεις

// repeatString - Επανάληψη string
func repeatString(n int, char string) string {
	result := ""
	for i := 0; i < n; i++ {
		result += char
	}
	return result
}

// min - Επιστροφή ελάχιστης τιμής
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
