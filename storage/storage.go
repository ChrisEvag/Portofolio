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

// StorageInterface - Interface για αποθήκευση δεδομένων
type StorageInterface interface {
	Save(tokens []types.TokenInfo) error
	GetName() string
}

// CSVStorage - Αποθήκευση σε CSV
type CSVStorage struct {
	DataFolder string
}

func NewCSVStorage(dataFolder string) *CSVStorage {
	return &CSVStorage{
		DataFolder: dataFolder,
	}
}

func (s *CSVStorage) GetName() string {
	return "CSV"
}

func (s *CSVStorage) Save(tokens []types.TokenInfo) error {
	// Ομαδοποίηση tokens ανά chain
	tokensByChain := make(map[string][]types.TokenInfo)
	for _, token := range tokens {
		tokensByChain[token.Chain] = append(tokensByChain[token.Chain], token)
	}

	// Αποθήκευση για κάθε chain ξεχωριστά
	for chain, chainTokens := range tokensByChain {
		if err := s.saveChainTokens(chain, chainTokens); err != nil {
			return err
		}
	}

	return nil
}

func (s *CSVStorage) saveChainTokens(chain string, tokens []types.TokenInfo) error {
	// Δημιουργία chain-specific folder
	chainFolder := filepath.Join(s.DataFolder, chain)
	if err := s.ensureDataFolder(chainFolder); err != nil {
		return fmt.Errorf("δεν μπόρεσα να δημιουργήσω φάκελο για %s: %v", chain, err)
	}

	// Προσθήκη timestamp στο filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_tokens_%s.csv", chain, timestamp)
	filepath := filepath.Join(chainFolder, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("σφάλμα δημιουργίας αρχείου: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Chain", "Symbol", "Name", "Price_USD", "Liquidity", "Pool_Count",
		"Denom", "Source", "Block_Height", "Timestamp",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	timestampStr := time.Now().Format("2006-01-02 15:04:05")
	for _, token := range tokens {
		liquidityStr := ""
		if token.Liquidity > 0 {
			liquidityStr = strconv.FormatFloat(token.Liquidity, 'f', 2, 64)
		}

		record := []string{
			token.Chain,
			token.Symbol,
			token.Name,
			strconv.FormatFloat(token.Price, 'f', 6, 64),
			liquidityStr,
			strconv.Itoa(token.PoolCount),
			token.Denom,
			token.Source,
			strconv.FormatInt(token.BlockHeight, 10),
			timestampStr,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("💾 ΑΠΟΘΗΚΕΥΣΗ %s: Τα %d tokens αποθηκεύτηκαν στο %s\n",
		chain, len(tokens), filepath)
	return nil
}

func (s *CSVStorage) ensureDataFolder(folderPath string) error {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return os.MkdirAll(folderPath, 0755)
	}
	return nil
}
