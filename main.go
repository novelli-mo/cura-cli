package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/novelli-mo/cura/config"
	"github.com/novelli-mo/cura/llm"
	"github.com/novelli-mo/cura/scanner"
)

func main() {
	rootCmd := &cobra.Command{
		Use:           "cura",
		Short:         "A skill manager for your repos",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize cura in the current repo",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := os.Getwd()

			ctx, err := scanner.ScanRepo(dir)
			if err != nil {
				fmt.Println("Error scanning repo:", err)
				os.Exit(1)
			}

			summary := fmt.Sprintf("Extensions: %v\nRoot files: %v\nFolders: %v\nTotal files: %d\n\nDocs:\n%s",
				ctx.Extensions, ctx.RootFiles, ctx.FolderNames, ctx.TotalFiles, ctx.DocContent)

			cfg, _ := config.Load()

			if llm.ClaudeAvailable() && !cfg.ClaudeApproved {
				fmt.Println("\ncura wants to use your Claude Code session for repo analysis.")
				fmt.Println("We promise this will use only a very small amount of your Claude tokens.")
				fmt.Print("Allow? (y/n): ")
				var answer string
				fmt.Scanln(&answer)
				if answer == "y" || answer == "Y" {
					cfg.ClaudeApproved = true
					config.Save(cfg)
					fmt.Println("Saved. You won't be asked again.")
					fmt.Println("Revoke permission using cura revoke -i claude")
				}
			}

			if !(llm.ClaudeAvailable() && cfg.ClaudeApproved) && cfg.GeminiAPIKey == "" {
				fmt.Println("\nNo LLM provider configured.")
				fmt.Println("→ Get a free Gemini key at: https://aistudio.google.com")
				fmt.Print("Paste your Gemini API key: ")
				fmt.Scanln(&cfg.GeminiAPIKey)
				config.Save(cfg)
				fmt.Println("Saved. You won't be asked again.")
				fmt.Println("Revoke permission using cura revoke -i gemini")
			}

			estimatedNext := llm.EstimateTokens(llm.BuildPrompt(summary))
			limit := 10000
			ok, _ := llm.CheckLimit(dir, estimatedNext, limit)
			if !ok {
				fmt.Printf("Token limit reached (%d total). Use `cura status` to check usage.\n", limit)
				fmt.Println("Aborted.")
				return
			}

			var resp llm.LLMResponse

			if llm.ClaudeAvailable() && cfg.ClaudeApproved {
				fmt.Println("Using Claude Code (your existing session)...")
				resp, err = llm.CallWithClaude(llm.BuildPrompt(summary))
			} else {
				fmt.Println("Using Gemini Flash...")
				resp, err = llm.CallWithGemini(cfg.GeminiAPIKey, llm.BuildPrompt(summary))
			}

			if err != nil {
				fmt.Println("LLM error:", err)
				os.Exit(1)
			}

			llm.RecordUsage(dir, resp.TokensUsed)

			tokenLabel := "tokens"
			if resp.IsEstimated {
				tokenLabel = "tokens (estimated)"
			}
			fmt.Printf("Tokens used this call: ~%d %s\n", resp.TokensUsed, tokenLabel)
			fmt.Println("\nSuggested skills (raw):")
			fmt.Println(resp.Text)
		},
	}

	revokeCmd := &cobra.Command{
		Use:   "revoke",
		Short: "Clear saved credentials and API keys",
		Run: func(cmd *cobra.Command, args []string) {
			provider, _ := cmd.Flags().GetString("provider")
			cfg, _ := config.Load()

			switch provider {
			case "claude":
				cfg.ClaudeApproved = false
				config.Save(cfg)
				fmt.Println("Claude permission revoked.")
			case "gemini":
				cfg.GeminiAPIKey = ""
				config.Save(cfg)
				fmt.Println("Gemini API key cleared.")
			case "":
				os.Remove(os.ExpandEnv("$HOME/.cura/config.toml"))
				fmt.Println("All credentials cleared.")
			default:
				fmt.Printf("Unknown provider: %s\nAvailable: claude, gemini\n", provider)
			}
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show token usage for current repo",
		Run: func(cmd *cobra.Command, args []string) {
			dir, _ := os.Getwd()
			usage := llm.GetUsage(dir)
			if usage.CallCount == 0 {
				fmt.Println("No token usage recorded for this repo.")
				return
			}
			fmt.Printf("Repo: %s\n", dir)
			fmt.Printf("Total tokens used: %d\n", usage.TotalTokens)
			fmt.Printf("LLM calls made: %d\n", usage.CallCount)
			fmt.Printf("Last used: %s\n", usage.LastUsed.Format("2006-01-02 15:04"))
		},
	}

	revokeCmd.Flags().StringP("provider", "i", "", "Provider to revoke (claude, gemini)")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(revokeCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
