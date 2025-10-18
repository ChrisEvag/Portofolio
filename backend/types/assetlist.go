package types

type AssetList struct {
	ChainName string  `json:"chain_name"`
	Assets    []Asset `json:"assets"`
}

type Asset struct {
	Description string      `json:"description"`
	DenomUnits  []DenomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Name        string      `json:"name"`
	Display     string      `json:"display"`
	Symbol      string      `json:"symbol"`
	LogoURIs    *LogoURIs   `json:"logo_URIs,omitempty"`
	Traces      []Trace     `json:"traces,omitempty"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent int      `json:"exponent"`
	Aliases  []string `json:"aliases,omitempty"`
}

type LogoURIs struct {
	PNG string `json:"png,omitempty"`
	SVG string `json:"svg,omitempty"`
}

type Trace struct {
	Type         string            `json:"type"`
	Counterparty TraceCounterparty `json:"counterparty"`
	Chain        TraceChain        `json:"chain"`
}

type TraceCounterparty struct {
	ChainName string `json:"chain_name"`
	BaseDenom string `json:"base_denom"`
	ChannelID string `json:"channel_id"`
}

type TraceChain struct {
	ChannelID string `json:"channel_id"`
	Path      string `json:"path"`
}

// GetDenomMapping returns a mapping from base denom to symbol
func GetDenomMapping(assets []Asset) map[string]string {
	mapping := make(map[string]string)

	for _, asset := range assets {
		// Store the mapping from base denom to symbol
		mapping[asset.Base] = asset.Symbol

		// Also store mappings for any aliases
		for _, denomUnit := range asset.DenomUnits {
			for _, alias := range denomUnit.Aliases {
				mapping[alias] = asset.Symbol
			}
		}
	}

	return mapping
}

// GetTokenMetadata returns detailed token information including logos and denominations
func GetTokenMetadata(assets []Asset) map[string]Asset {
	metadata := make(map[string]Asset)

	for _, asset := range assets {
		metadata[asset.Symbol] = asset

		// Also index by base denom
		metadata[asset.Base] = asset

		// And by any aliases
		for _, denomUnit := range asset.DenomUnits {
			for _, alias := range denomUnit.Aliases {
				metadata[alias] = asset
			}
		}
	}

	return metadata
}
