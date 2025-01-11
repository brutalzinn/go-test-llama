package instruction

import (
	"testing"
)

func TestHandleCallback_SingleInstruction(t *testing.T) {
	input := `Some text with #instruction("command1", "arg1", "arg2", "arg3").`
	expected := []Instruction{
		{Name: "command1", Args: []string{"arg1", "arg2", "arg3"}},
	}

	actual := HandleCallback(input)
	if len(actual) != len(expected) {
		t.Errorf("Expected %d instructions, but got %d", len(expected), len(actual))
	}

	for i, expectedInstr := range expected {
		if actual[i].Name != expectedInstr.Name || !compareSlices(actual[i].Args, expectedInstr.Args) {
			t.Errorf("Mismatch in instruction %d:\nExpected: %+v\nGot: %+v", i, expectedInstr, actual[i])
		}
	}
}

func TestHandleCallback_MultipleInstructions(t *testing.T) {
	input := `Multiple instructions: #instruction("cmd1", "a1", "a2") and #instruction("cmd2", "b1").`
	expected := []Instruction{
		{Name: "cmd1", Args: []string{"a1", "a2"}},
		{Name: "cmd2", Args: []string{"b1"}},
	}

	actual := HandleCallback(input)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d instructions, but got %d", len(expected), len(actual))
	}

	for i, expectedInstr := range expected {
		if actual[i].Name != expectedInstr.Name || !compareSlices(actual[i].Args, expectedInstr.Args) {
			t.Errorf("Mismatch in instruction %d:\nExpected: %+v\nGot: %+v", i, expectedInstr, actual[i])
		}
	}
}

func TestHandleCallback_NoInstructions(t *testing.T) {
	input := `No instructions here.`
	expected := []Instruction{}

	actual := HandleCallback(input)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d instructions, but got %d", len(expected), len(actual))
	}
}

func TestHandleCallback_InvalidInstruction(t *testing.T) {
	input := `Invalid instruction: #instruction("cmd", "arg1" "arg2")` // Missing comma
	expected := []Instruction{}

	actual := HandleCallback(input)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d instructions, but got %d", len(expected), len(actual))
	}
}

func TestHandleCallback_BigText_Instruction(t *testing.T) {
	input := `Olá Roberto!\n\nAgradeço por compartilhar conosco como a luz fluorescente está te causando desconforto. Isso é muito importante para nós saber como podemos ajudar.\n\nVamos começar! \n\n#instruction(sensorial, Roberto, luz fluorescente me dando dor de cabeça, luz fluorescente, 8)\n\nEssa instrução significa que estamos coletando informações sobre a sua regulação sensorial e o fator que está causando desconforto. Nós vamos armazenar isso para poder monitorar seu progresso.\n\nAgora, gostaria de fazer uma pergunta: Você achou que o problema é só com a luz fluorescente ou se há outros fatores que podem estar contribuindo?`
	expected := []Instruction{
		{Name: "health", Args: []string{"Roberto", "triste"}},
	}

	actual := HandleCallback(input)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d instructions, but got %d", len(expected), len(actual))
	}

	for i, expectedInstr := range expected {
		if actual[i].Name != expectedInstr.Name || !compareSlices(actual[i].Args, expectedInstr.Args) {
			t.Errorf("Mismatch in instruction %d:\nExpected: %+v\nGot: %+v", i, expectedInstr, actual[i])
		}
	}
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
