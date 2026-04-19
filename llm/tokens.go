package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TokenUsage struct {
	TotalTokens int       `toml:"total_tokens"`
	LastUsed    time.Time `toml:"last_used"`
	CallCount   int       `toml:"call_count"`
}

type TokenStore map[string]TokenUsage

const DefaultTokenLimit = 4000

func tokenStorePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cura", "token_usage.json")
}

func LoadTokenStore() TokenStore {
	store := make(TokenStore)
	data, err := os.ReadFile(tokenStorePath())
	if err != nil {
		return store
	}
	json.Unmarshal(data, &store)
	return store
}

func SaveTokenStore(store TokenStore) error {
	path := tokenStorePath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func RecordUsage(repoPath string, tokens int) error {
	store := LoadTokenStore()
	usage := store[repoPath]
	usage.TotalTokens += tokens
	usage.LastUsed = time.Now()
	usage.CallCount++
	store[repoPath] = usage
	return SaveTokenStore(store)
}

func GetUsage(repoPath string) TokenUsage {
	return LoadTokenStore()[repoPath]
}

func EstimateTokens(text string) int {
	return len(text) / 4
}

func CheckLimit(repoPath string, estimatedNext int, limit int) (bool, error) {
	usage := GetUsage(repoPath)
	if usage.TotalTokens+estimatedNext > limit {
		fmt.Printf("\n⚠ Token limit reached (%d used, limit: %d).\n", usage.TotalTokens, limit)
		fmt.Printf("This call will use ~%d more tokens.\n", estimatedNext)
		fmt.Print("Continue anyway? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			return false, nil
		}
	}
	return true, nil
}
