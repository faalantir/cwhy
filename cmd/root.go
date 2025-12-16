package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings" // <--- Added this

	"github.com/faalantir/cwhy/internal/ai"
	"github.com/faalantir/cwhy/internal/data"
	"github.com/faalantir/cwhy/internal/ui"
	"github.com/spf13/cobra"
)

var (
	EmbeddedSupabaseURL string
	EmbeddedSupabaseKey string
)

var rootCmd = &cobra.Command{
	Use:   "cwhy",
	Short: "The AI Debugger for your terminal",
	Long: `cwhy is an AI-powered CLI tool that explains your error logs.
It pipes output from your commands (Terraform, Docker, Build logs) 
and uses OpenAI to find instant fixes.

It caches successful fixes to a shared Team Memory, 
so you never have to solve the same bug twice.`,
	Example: `  # 1. Pipe logs directly (Recommended)
  terraform apply | cwhy
  docker logs my-container | cwhy
  
  # 2. Analyze a specific log file
  cat build.log | cwhy

  # 3. Paste an error string
  cwhy "Error: AccessDeniedException..."`,
	Run: func(cmd *cobra.Command, args []string) {
		// SAFETY CHECK: Check if data is being piped to stdin
		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			// Case 1: Piped input (cat log.txt | cwhy)
			handlePipedInput("")
		} else if len(args) > 0 {
			// Case 2: Manual argument (cwhy "error text")
			handlePipedInput(strings.Join(args, " "))
		} else {
			// Case 3: No input? Show Help.
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

// Updated to accept an optional argument
func handlePipedInput(manualInput string) {
	// 1. Mandatory: OpenAI Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
		return
	}

	// 2. Optional: Supabase Keys (Check Env first, then Embedded)
	supaUrl := os.Getenv("SUPABASE_URL")
	if supaUrl == "" {
		supaUrl = EmbeddedSupabaseURL
	}

	supaKey := os.Getenv("SUPABASE_KEY")
	if supaKey == "" {
		supaKey = EmbeddedSupabaseKey
	}

	useMemory := false
	var suClient *data.Client

	if supaUrl != "" && supaKey != "" {
		var err error
		suClient, err = data.NewClient(supaUrl, supaKey)
		if err == nil {
			useMemory = true
		}
	}

	// 3. Get the Input (Manual or Piped?)
	var input string
	if manualInput != "" {
		input = manualInput
	} else {
		// Read from Pipe
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input += scanner.Text() + "\n"
		}
	}

	if len(strings.TrimSpace(input)) < 2 {
		return
	}

	// 4. CHECK MEMORY (Only if enabled)
	if useMemory {
		inputHash := data.GenerateHash(input)
		existingFix, err := suClient.GetExistingFix(inputHash)
		if err == nil && existingFix != nil {
			fmt.Println("⚡ Found in Team Memory")
			ui.RenderOutput(existingFix.FixSolution)
			return
		}
	}

	// 5. Ask AI
	fmt.Println("🧠 Analyzing logs...")
	explanation, err := ai.GetExplanation(apiKey, input)
	if err != nil {
		fmt.Printf("Error contacting AI: %v\n", err)
		return
	}

	ui.RenderOutput(explanation)

	// 6. SAVE TO MEMORY (Only if enabled)
	if useMemory {
		fmt.Print("💾 Saving to Team Memory... ")
		// Attempt to save (ignoring error for now to keep flow smooth)
		_ = suClient.SaveFix(input, explanation) 
		fmt.Println("Done.")
	} else {
		fmt.Println("\n(Tip: Configure SUPABASE_URL to enable Team Memory)")
	}
}