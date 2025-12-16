package ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

func GetExplanation(apiKey string, input string) (string, error) {
	client := openai.NewClient(apiKey)

	// The Prompt Strategy:
	// We tell the AI it is a "Senior SRE" (Site Reliability Engineer).
	systemPrompt := `You are cwhy, a Senior SRE assistant in the terminal.
	Analyze the following error log or code snippet.
	1. concise explanation of what went wrong (1 sentence).
	2. The exact shell command or code change to fix it.
	3. Keep it short. Use Markdown for formatting.`

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini, // Cheap and fast
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}