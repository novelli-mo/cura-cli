package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type Caller interface {
	Call(prompt string) (string, error)
}

type Response struct {
	Skills []string `json:"skills"`
}

type ClaudeCaller struct{}

func NewClaude() Caller { return ClaudeCaller{} }

func (c ClaudeCaller) Call(prompt string) (string, error) {
	cmd := exec.Command("claude", "-p", prompt, "--model", "claude-haiku-4-5")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

type GeminiCaller struct{ APIKey string }

func NewGemini(apiKey string) Caller { return GeminiCaller{APIKey: apiKey} }

func (g GeminiCaller) Call(prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + g.APIKey

	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]any{{"text": prompt}}},
		},
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	text := result["candidates"].([]any)[0].(map[string]any)["content"].(map[string]any)["parts"].([]any)[0].(map[string]any)["text"].(string)
	return strings.TrimSpace(text), nil
}

func ClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func BuildPrompt(scanSummary string) string {
	return fmt.Sprintf(`You are a developer tools assistant. Based on this repository scan, suggest 3-5 skill IDs that would help developers working on this repo.

Repository scan:
%s

Respond ONLY with a valid JSON object like this (no explanation, no markdown):
{"skills": ["go-backend", "docker", "rest-api"]}

Available skill categories: go-backend, python-backend, node-backend, rust-backend, frontend-react, frontend-vue, docker, kubernetes, rest-api, graphql, postgres, mongodb, redis, grpc, cli-tool, data-science, machine-learning, senior-arch`, scanSummary)
}

type LLMResponse struct {
	Text        string
	TokensUsed  int
	IsEstimated bool
}

func CallWithClaude(prompt string) (LLMResponse, error) {
	cmd := exec.Command("claude", "-p", prompt, "--model", "claude-haiku-4-5")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	text := out.String()
	// estimate tokens for claude
	tokens := EstimateTokens(prompt + text)
	return LLMResponse{Text: text, TokensUsed: tokens, IsEstimated: true}, err
}

func CallWithGemini(apiKey, prompt string) (LLMResponse, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]any{{"text": prompt}}},
		},
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return LLMResponse{}, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return LLMResponse{}, err
	}

	text := result["candidates"].([]any)[0].(map[string]any)["content"].(map[string]any)["parts"].([]any)[0].(map[string]any)["text"].(string)

	tokens := 0
	if meta, ok := result["usageMetadata"].(map[string]any); ok {
		if total, ok := meta["totalTokenCount"].(float64); ok {
			tokens = int(total)
		}
	}

	return LLMResponse{Text: strings.TrimSpace(text), TokensUsed: tokens, IsEstimated: false}, nil
}
