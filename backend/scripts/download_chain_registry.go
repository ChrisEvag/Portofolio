package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	githubAPIURL = "https://api.github.com/repos/cosmos/chain-registry/contents/osmosis"
)

type GithubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}

func DownloadChainRegistry() error {
	// Create base directory
	baseDir := filepath.Join("data", "chain-registry", "osmosis")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return err
	}

	// Get repository contents
	contents, err := getRepositoryContents()
	if err != nil {
		fmt.Printf("Error getting repository contents: %v\n", err)
		return err
	}

	// Download each file
	for _, content := range contents {
		if err := downloadFile(content, baseDir); err != nil {
			fmt.Printf("Error downloading %s: %v\n", content.Name, err)
			continue
		}
		fmt.Printf("Successfully downloaded: %s\n", content.Name)
	}

	return nil
}

func getRepositoryContents() ([]GithubContent, error) {
	// Create HTTP client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers for better rate limiting
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	// If you have a GitHub token, add it here
	// req.Header.Add("Authorization", "token YOUR_GITHUB_TOKEN")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var contents []GithubContent
	if err := json.Unmarshal(body, &contents); err != nil {
		return nil, err
	}

	return contents, nil
}

func downloadFile(content GithubContent, baseDir string) error {
	// Skip directories
	if content.Type != "file" {
		return nil
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", content.DownloadURL, nil)
	if err != nil {
		return err
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create file
	filePath := filepath.Join(baseDir, content.Name)
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Save content
	_, err = io.Copy(out, resp.Body)
	return err
}