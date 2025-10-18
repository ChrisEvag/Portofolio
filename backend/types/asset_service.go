package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type AssetService struct {
	DenomToSymbol map[string]string
	TokenMetadata map[string]Asset
	OsmoUsdPrice  float64
}

func NewAssetService() (*AssetService, error) {
	// Read assetlist.json
	assetList, err := loadAssetList()
	if err != nil {
		return nil, fmt.Errorf("failed to load asset list: %w", err)
	}

	// Create mappings
	denomToSymbol := GetDenomMapping(assetList.Assets)
	tokenMetadata := GetTokenMetadata(assetList.Assets)

	return &AssetService{
		DenomToSymbol: denomToSymbol,
		TokenMetadata: tokenMetadata,
		OsmoUsdPrice:  1.0, // Default τιμή, θα ενημερωθεί αργότερα
	}, nil
}

func loadAssetList() (*AssetList, error) {
	// Get project root directory by looking for go.mod
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
			break
		}
		parentDir := filepath.Dir(rootDir)
		if parentDir == rootDir {
			return nil, fmt.Errorf("could not find project root (no go.mod found)")
		}
		rootDir = parentDir
	}

	// Read assetlist.json
	assetListPath := filepath.Join(rootDir, "data", "chain-registry", "osmosis", "assetlist.json")
	content, err := os.ReadFile(assetListPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read assetlist.json: %w", err)
	}

	var assetList AssetList
	if err := json.Unmarshal(content, &assetList); err != nil {
		return nil, fmt.Errorf("failed to parse assetlist.json: %w", err)
	}

	return &assetList, nil
}

// GetSymbol returns the symbol for a given denom
func (s *AssetService) GetSymbol(denom string) string {
	if symbol, ok := s.DenomToSymbol[denom]; ok {
		return symbol
	}
	return denom // Return original denom if no mapping found
}

// GetAsset returns the full asset metadata for a given symbol or denom
func (s *AssetService) GetAsset(symbolOrDenom string) (Asset, bool) {
	asset, ok := s.TokenMetadata[symbolOrDenom]
	return asset, ok
}

// GetLogoURL returns the logo URL for a given symbol or denom
func (s *AssetService) GetLogoURL(symbolOrDenom string) string {
	if asset, ok := s.TokenMetadata[symbolOrDenom]; ok && asset.LogoURIs != nil {
		if asset.LogoURIs.PNG != "" {
			return asset.LogoURIs.PNG
		}
		return asset.LogoURIs.SVG
	}
	return ""
}

// GetDisplayDenom returns the display denomination for a given base denom
func (s *AssetService) GetDisplayDenom(baseDenom string) string {
	if asset, ok := s.TokenMetadata[baseDenom]; ok {
		return asset.Display
	}
	return baseDenom
}

// GetExponent returns the exponent for converting between base and display units
func (s *AssetService) GetExponent(denom string) int {
	if asset, ok := s.TokenMetadata[denom]; ok {
		for _, unit := range asset.DenomUnits {
			if unit.Denom == asset.Display {
				return unit.Exponent
			}
		}
	}
	return 0
}

// SetOsmoUsdPrice sets the current OSMO/USD price
func (s *AssetService) SetOsmoUsdPrice(price float64) {
	s.OsmoUsdPrice = price
}

// GetOsmoUsdPrice returns the current OSMO/USD price
func (s *AssetService) GetOsmoUsdPrice() float64 {
	return s.OsmoUsdPrice
}

// ConvertUsdToOsmo converts a USD price to OSMO
func (s *AssetService) ConvertUsdToOsmo(usdPrice float64) float64 {
	if s.OsmoUsdPrice <= 0 {
		return 0
	}
	return usdPrice / s.OsmoUsdPrice
}

// GetDenom returns the base denom for a given symbol
func (s *AssetService) GetDenom(symbol string) string {
	// Αναζήτηση στο TokenMetadata για το denom που αντιστοιχεί στο symbol
	for _, asset := range s.TokenMetadata {
		if asset.Symbol == symbol {
			return asset.Base
		}
	}
	return "" // Επιστρέφει κενό string αν δε βρεθεί το symbol
}

// GetDenomBySymbol is an alias for GetDenom for consistency
func (s *AssetService) GetDenomBySymbol(symbol string) string {
	return s.GetDenom(symbol)
}

// GetAllTokens returns all tokens from the asset service
func (s *AssetService) GetAllTokens() []Asset {
	tokens := make([]Asset, 0, len(s.TokenMetadata))
	for _, asset := range s.TokenMetadata {
		tokens = append(tokens, asset)
	}
	return tokens
}
