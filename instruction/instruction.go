package instruction

import (
	"regexp"
)

type Instruction struct {
	Name string
	Args []string
}

func HandleCallback(ollamaResponse string) []Instruction {
	re := regexp.MustCompile(`#instruction\("([^"]*)", ((?:"[^"]*",\s*)*"[^"]*")\)`)
	matches := re.FindAllStringSubmatch(ollamaResponse, -1)
	instructions := []Instruction{}
	for _, match := range matches {
		if len(match) == 3 {
			instruction := Instruction{
				Name: match[1],
			}
			args := regexp.MustCompile(`"([^"]*)"`).FindAllStringSubmatch(match[2], -1)
			for _, arg := range args {
				instruction.Args = append(instruction.Args, arg[1])
			}
			instructions = append(instructions, instruction)
		}
	}
	return instructions
}
