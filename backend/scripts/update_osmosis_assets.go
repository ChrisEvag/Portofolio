package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const (
	assetListURL = "https://raw.githubusercontent.com/cosmos/chain-registry/master/osmosis/assetlist.json"
)

type Asset struct {
	Description string `json:"description"`
	DenomUnits  []struct {
		Denom    string `json:"denom"`
		Exponent int    `json:"exponent"`
	} `json:"denom_units"`
	Base     string   `json:"base"`
	Name     string   `json:"name"`
	Display  string   `json:"display"`
	Symbol   string   `json:"symbol"`
	Traces   []Trace  `json:"traces,omitempty"`
	LogoURIs struct{} `json:"logo_URIs"`
}

type Trace struct {
	Type    string `json:"type"`
	Counter int    `json:"counter"`
	Path    string `json:"path"`
}

type AssetList struct {
	Assets []Asset `json:"assets"`
}

// Template για το token_mapping.go αρχείο
const mappingTemplate = `package types

// TokenDenomMapping maps IBC denoms to their actual token symbols
// Auto-generated from Osmosis Chain Registry
var TokenDenomMapping = map[string]string{
{{- range .}}
	"{{.Denom}}": "{{.Symbol}}", // {{.Name}}
{{- end}}
}

// GetTokenSymbol returns the token symbol for a given denom
func GetTokenSymbol(denom string) string {
	if symbol, ok := TokenDenomMapping[denom]; ok {
		return symbol
	}
	return denom // Return original denom if no mapping found
}

// IsKnownToken checks if a given denom is mapped to a known token symbol
func IsKnownToken(denom string) bool {
	_, ok := TokenDenomMapping[denom]
	return ok
}
`

type TokenInfo struct {
	Denom  string
	Symbol string
	Name   string
}

func main() {
	// Download assetlist.json
	fmt.Println("📥 Downloading Osmosis assetlist.json...")
	resp, err := http.Get(assetListURL)
	if err != nil {
		fmt.Printf("❌ Failed to download assetlist: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var assetList AssetList
	if err := json.Unmarshal(body, &assetList); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Extract token info
	var tokens []TokenInfo
	seenDenoms := make(map[string]bool)

	// Πρώτα πρόσθεσε το OSMO
	tokens = append(tokens, TokenInfo{
		Denom:  "uosmo",
		Symbol: "OSMO",
		Name:   "Osmosis",
	})
	seenDenoms["uosmo"] = true

	for _, asset := range assetList.Assets {
		// Skip if no traces (meaning it's not an IBC token)
		if len(asset.Traces) == 0 && asset.Base != "uosmo" {
			continue
		}

		var denom string
		if len(asset.Traces) > 0 {
			// For IBC tokens, use the full IBC denom
			denom = asset.Base
		} else {
			// Skip OSMO as it's already added
			continue
		}

		// Skip if we've seen this denom before
		if seenDenoms[denom] {
			continue
		}

		tokens = append(tokens, TokenInfo{
			Denom:  denom,
			Symbol: strings.ToUpper(asset.Symbol),
			Name:   asset.Name,
		})
		seenDenoms[denom] = true
	}

	// Sort by symbol for consistent output
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Symbol < tokens[j].Symbol
	})

	// Generate token_mapping.go
	tmpl, err := template.New("mapping").Parse(mappingTemplate)
	if err != nil {
		fmt.Printf("❌ Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	// Create types directory if it doesn't exist
	typesDir := filepath.Join("..", "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create types directory: %v\n", err)
		os.Exit(1)
	}

	// Write to token_mapping.go
	outPath := filepath.Join(typesDir, "token_mapping.go")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("❌ Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	if err := tmpl.Execute(outFile, tokens); err != nil {
		fmt.Printf("❌ Failed to write output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully updated %s with %d tokens\n", outPath, len(tokens))
	fmt.Println("🔍 Token mappings:")
	for _, token := range tokens {
		fmt.Printf("  • %-6s: %s\n", token.Symbol, token.Denom)
	}
}
