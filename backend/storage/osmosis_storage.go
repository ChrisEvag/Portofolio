package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"portofoliov1/types"
)

type OsmosisCSVStorage struct {
	BaseDir string
}

func NewOsmosisCSVStorage(dataFolder string) *OsmosisCSVStorage {
	return &OsmosisCSVStorage{
		BaseDir: dataFolder,
	}
}

// SavePoolStats αποθηκεύει τα στατιστικά των pools σε CSV
func (s *OsmosisCSVStorage) SavePoolStats(stats []types.PoolStats) error {
	// Δημιουργία φακέλου για pool stats
	poolStatsFolder := filepath.Join(s.BaseDir, "crypto-tokens", "osmosis", "pool_stats")
	if err := s.ensureDataFolder(poolStatsFolder); err != nil {
		return fmt.Errorf("δεν μπόρεσα να δημιουργήσω φάκελο για pool stats: %v", err)
	}

	// Δημιουργία ονόματος αρχείου με timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("osmosis_pool_stats_%s.csv", timestamp)
	filepath := filepath.Join(poolStatsFolder, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("σφάλμα δημιουργίας αρχείου: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Γράψε το header
	header := []string{
		"Pool_ID", "Volume_24h", "Volume_7d", "Fees_24h", "TVL",
		"Token_APRs", "Timestamp",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Γράψε τα δεδομένα
	timestampStr := time.Now().Format("2006-01-02 15:04:05")
	for _, stat := range stats {
		// Μετατροπή του TokenAPRs map σε string
		aprs := ""
		for token, apr := range stat.TokenAPRs {
			if aprs != "" {
				aprs += "|"
			}
			aprs += fmt.Sprintf("%s:%.2f%%", token, apr*100)
		}

		record := []string{
			stat.PoolId,
			strconv.FormatFloat(stat.Volume24h, 'f', 2, 64),
			strconv.FormatFloat(stat.Volume7d, 'f', 2, 64),
			strconv.FormatFloat(stat.Fees24h, 'f', 2, 64),
			strconv.FormatFloat(stat.TVL, 'f', 2, 64),
			aprs,
			timestampStr,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("💾 Pool stats αποθηκεύτηκαν στο %s\n", filepath)
	return nil
}

// SaveSpotPrices αποθηκεύει τις spot τιμές σε CSV
func (s *OsmosisCSVStorage) SaveSpotPrices(ticks []types.SpotPriceTick) error {
	// Δημιουργία φακέλου για τιμές
	pricesFolder := filepath.Join(s.BaseDir, "crypto-tokens", "osmosis", "spot_prices")
	if err := s.ensureDataFolder(pricesFolder); err != nil {
		return fmt.Errorf("δεν μπόρεσα να δημιουργήσω φάκελο για τιμές: %v", err)
	}

	// Δημιουργία ονόματος αρχείου με timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("osmosis_spot_prices_%s.csv", timestamp)
	filepath := filepath.Join(pricesFolder, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("σφάλμα δημιουργίας αρχείου: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Γράψε το header
	header := []string{
		"Pool_ID", "Token0", "Token1", "Price", "Timestamp",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Γράψε τα δεδομένα
	for _, tick := range ticks {
		record := []string{
			strconv.FormatUint(tick.PoolId, 10),
			tick.Token0,
			tick.Token1,
			strconv.FormatFloat(tick.Price, 'f', 6, 64),
			time.Unix(tick.Timestamp, 0).Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("💾 Spot prices αποθηκεύτηκαν στο %s\n", filepath)
	return nil
}

// SavePools αποθηκεύει τα pools σε CSV
func (s *OsmosisCSVStorage) SavePools(pools []types.OsmosisPool) error {
	// Δημιουργία φακέλου για pools
	poolsFolder := filepath.Join(s.BaseDir, "crypto-tokens", "osmosis", "pools")
	if err := s.ensureDataFolder(poolsFolder); err != nil {
		return fmt.Errorf("δεν μπόρεσα να δημιουργήσω φάκελο για pools: %v", err)
	}

	// Δημιουργία ονόματος αρχείου με timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("osmosis_pools_%s.csv", timestamp)
	filepath := filepath.Join(poolsFolder, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("σφάλμα δημιουργίας αρχείου: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Γράψε το header
	header := []string{
		"Pool_ID", "Type", "Assets", "Total_Weight", "Swap_Fee",
		"Exit_Fee", "Total_Shares", "Timestamp",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Γράψε τα δεδομένα
	timestampStr := time.Now().Format("2006-01-02 15:04:05")
	for _, pool := range pools {
		// Μετατροπή των assets σε string
		assets := ""
		for i, asset := range pool.PoolAssets {
			if i > 0 {
				assets += "|"
			}
			assets += fmt.Sprintf("%s:%s", asset.Token.Denom, asset.Token.Amount)
		}

		record := []string{
			pool.Id,
			pool.Type,
			assets,
			"", // TotalWeight δεν υπάρχει πια
			pool.PoolParams.SwapFee,
			pool.PoolParams.ExitFee,
			fmt.Sprintf("%s:%s", pool.TotalShares.Denom, pool.TotalShares.Amount),
			timestampStr,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("💾 Pools αποθηκεύτηκαν στο %s\n", filepath)
	return nil
}

// SaveTokenPrices αποθηκεύει τις τιμές των tokens σε CSV
func (s *OsmosisCSVStorage) SaveTokenPrices(prices map[string]float64, assetService *types.AssetService) error {
	// Δημιουργία φακέλου για τιμές tokens
	pricesFolder := filepath.Join(s.BaseDir, "crypto-tokens", "osmosis", "token_prices")
	if err := s.ensureDataFolder(pricesFolder); err != nil {
		return fmt.Errorf("δεν μπόρεσα να δημιουργήσω φάκελο για τιμές tokens: %v", err)
	}

	// Δημιουργία ονόματος αρχείου με timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("osmosis_token_prices_%s.csv", timestamp)
	filepath := filepath.Join(pricesFolder, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("σφάλμα δημιουργίας αρχείου: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Γράψε το header
	header := []string{
		"Symbol", "Denom", "Price_USD", "Timestamp",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Γράψε τα δεδομένα
	timestampStr := time.Now().Format("2006-01-02 15:04:05")
	for symbol, price := range prices {
		denom := assetService.GetDenom(symbol)
		if denom == "" {
			continue // Skip αν δεν βρέθηκε το denom
		}

		record := []string{
			symbol,
			denom,
			strconv.FormatFloat(price, 'f', 6, 64),
			timestampStr,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("💾 Token prices αποθηκεύτηκαν στο %s\n", filepath)
	return nil
}

func (s *OsmosisCSVStorage) SaveTokenPricesOSMO(prices []types.TokenPrice) error {
	if len(prices) == 0 {
		return fmt.Errorf("δεν υπάρχουν token prices για αποθήκευση")
	}

	// Αφού το BaseDir είναι ήδη "data/crypto-tokens/osmosis", προσθέτουμε μόνο "token_prices"
	dir := filepath.Join(s.BaseDir, "token_prices")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("αποτυχία δημιουργίας φακέλου: %w", err)
	}

	fname := fmt.Sprintf("osmosis_token_prices_osmo_%s.csv", time.Now().Format("20060102_150405"))
	fp := filepath.Join(dir, fname)
	f, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("αποτυχία δημιουργίας αρχείου: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	if err := w.Write([]string{"Symbol", "Denom", "Price_OSMO", "Timestamp"}); err != nil {
		return fmt.Errorf("αποτυχία εγγραφής header: %w", err)
	}

	for _, p := range prices {
		rec := []string{
			p.Symbol,
			p.Denom,
			fmt.Sprintf("%.12f", p.PriceOSMO),
			p.Timestamp.Format(time.RFC3339),
		}
		if err := w.Write(rec); err != nil {
			return fmt.Errorf("αποτυχία εγγραφής δεδομένων: %w", err)
		}
	}

	if err := w.Error(); err != nil {
		return fmt.Errorf("σφάλμα CSV writer: %w", err)
	}

	fmt.Printf("   💾 Token prices (OSMO): %s (%d tokens)\n", filepath.Base(fp), len(prices))
	return nil
}

// SaveAllPoolPrices αποθηκεύει τις τιμές όλων των pools σε CSV
func (s *OsmosisCSVStorage) SaveAllPoolPrices(poolPrices []types.PoolPrice) error {
	if len(poolPrices) == 0 {
		return fmt.Errorf("δεν υπάρχουν pool prices για αποθήκευση")
	}

	// Δημιουργία φακέλου για pool prices
	dir := filepath.Join(s.BaseDir, "pool_prices")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("αποτυχία δημιουργίας φακέλου: %w", err)
	}

	// Δημιουργία αρχείου
	fname := fmt.Sprintf("osmosis_all_pool_prices_%s.csv", time.Now().Format("20060102_150405"))
	fp := filepath.Join(dir, fname)
	f, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("αποτυχία δημιουργίας αρχείου: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header με καθαρά ονόματα
	if err := w.Write([]string{
		"Pool_ID",
		"Token0_Symbol",
		"Token0_Denom",
		"Token0_Amount",
		"Token1_Symbol",
		"Token1_Denom",
		"Token1_Amount",
		"Price_Token1_per_Token0",
		"Timestamp",
	}); err != nil {
		return fmt.Errorf("αποτυχία εγγραφής header: %w", err)
	}

	// Γράψε τα δεδομένα
	for _, p := range poolPrices {
		rec := []string{
			p.PoolID,
			p.Token0Symbol,
			p.Token0Denom,
			p.Token0Amount,
			p.Token1Symbol,
			p.Token1Denom,
			p.Token1Amount,
			fmt.Sprintf("%.18f", p.PriceOSMO), // Υψηλή ακρίβεια για μικρές τιμές
			p.Timestamp.Format(time.RFC3339),
		}
		if err := w.Write(rec); err != nil {
			return fmt.Errorf("αποτυχία εγγραφής δεδομένων: %w", err)
		}
	}

	if err := w.Error(); err != nil {
		return fmt.Errorf("σφάλμα CSV writer: %w", err)
	}

	fmt.Printf("   💾 Pool prices: %s (%d pools)\n", fp, len(poolPrices))
	return nil
}

func (s *OsmosisCSVStorage) ensureDataFolder(folderPath string) error {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return os.MkdirAll(folderPath, 0755)
	}
	return nil
}
