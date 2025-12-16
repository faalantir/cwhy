package ui

import (
	"fmt"
	
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func RenderOutput(markdown string) {
	// 1. Configure the Markdown Renderer (Glamour)
	// This makes code blocks look like code blocks, adds bolding, etc.
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(), // Automatically detects if you use Dark or Light mode
		glamour.WithWordWrap(100),
	)

	renderedMD, err := r.Render(markdown)
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
		return
	}

	// 2. Add a Premium Border (Lipgloss)
	// This puts the answer inside a nice purple rounded box.
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")). // Nice Purple
		Padding(1, 2).
		MarginTop(1)

	// 3. Print the Result
	fmt.Println(boxStyle.Render(renderedMD))
}