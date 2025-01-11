package instruction

import (
	"regexp"
	"strings"
)

type Instruction struct {
	Name string
	Args []string
}

func HandleCallback(ollamaResponse string) []Instruction {
	// Remove newlines for consistent matching
	responseWithoutNewlines := strings.ReplaceAll(ollamaResponse, "\n", " ")

	// Regular expression to match instructions
	re := regexp.MustCompile(`#instruction\(([^)]*)\)`)

	// Find all matches
	matches := re.FindAllStringSubmatch(responseWithoutNewlines, -1)
	instructions := []Instruction{}

	// Process each match
	for _, match := range matches {
		if len(match) == 2 {
			// Extract arguments from the match
			args := strings.Split(match[1], ",")
			if len(args) > 0 {
				name := strings.TrimSpace(args[0])
				arguments := []string{}
				for _, arg := range args[1:] {
					arguments = append(arguments, strings.TrimSpace(arg))
				}
				instructions = append(instructions, Instruction{
					Name: name,
					Args: arguments,
				})
			}
		}
	}

	return instructions
}
