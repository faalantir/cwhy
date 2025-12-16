package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/faalantir/cwhy/internal/ui"
	"github.com/faalantir/cwhy/internal/data"
	"github.com/faalantir/cwhy/internal/ai" // <--- Make sure this matches your go.mod name
)

var (
    // These will be filled in during the build process!
    EmbeddedSupabaseURL string
    EmbeddedSupabaseKey string
)

var rootCmd = &cobra.Command{
	Use:   "cwhy",
	Short: "The AI Debugger for your terminal",
	Run: func(cmd *cobra.Command, args []string) {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			handlePipedInput()
		} else {
			cmd.Help()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handlePipedInput() {
	// 1. Mandatory: OpenAI Key (Still required for now)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
		return
	}

	// 2. Optional: Supabase Keys
	// We check if they exist, but we don't stop if they don't.
	supaUrl := os.Getenv("SUPABASE_URL")
	if supaUrl == "" { supaUrl = EmbeddedSupabaseURL } // Fallback to embedded

	supaKey := os.Getenv("SUPABASE_KEY")
	if supaKey == "" { supaKey = EmbeddedSupabaseKey } // Fallback to embedded

	useMemory := false
	var suClient *data.Client

	if supaUrl != "" && supaKey != "" {
		var err error
		suClient, err = data.NewClient(supaUrl, supaKey)
		if err == nil {
			useMemory = true
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		input += scanner.Text() + "\n"
	}

	if len(input) < 2 { return }

	// 3. CHECK MEMORY (Only if enabled)
	if useMemory {
		inputHash := data.GenerateHash(input)
		existingFix, err := suClient.GetExistingFix(inputHash)
		if err == nil && existingFix != nil {
			fmt.Println("⚡ Found in Team Memory")
			ui.RenderOutput(existingFix.FixSolution)
			return
		}
	}

	// 4. Ask AI
	fmt.Println("🧠 Analyzing logs...") 
	explanation, err := ai.GetExplanation(apiKey, input)
	if err != nil {
		fmt.Printf("Error contacting AI: %v\n", err)
		return
	}

	ui.RenderOutput(explanation)

	// 5. SAVE TO MEMORY (Only if enabled)
	if useMemory {
		fmt.Print("💾 Saving to Team Memory... ")
		suClient.SaveFix(input, explanation) // Ignore error, just try
		fmt.Println("Done.")
	} else {
		// Nice upsell message for users without DB configured
		fmt.Println("\n(Tip: Configure SUPABASE_URL to enable Team Memory)")
	}
}