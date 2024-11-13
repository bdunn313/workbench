package cmd

import (
	"testing"
)

// mockRand is a mock random number generator that always returns a fixed value
type mockRand struct {
	value int
}

func (r *mockRand) Intn(n int) int {
	return r.value
}

func TestBasicSixSidedDie(t *testing.T) {
	r := &mockRand{value: 3}
	input := "1d6"
	expected := 4
	result, err := RollDice(r, input)

	if err != nil {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}

	if result.Total != expected {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}
}

func Test_parseExpression(t *testing.T) {
	input := "1d6"
	expected := Die{Sides: 6, Count: 1}
	result, err := parseExpression(input)

	if err != nil {
		t.Errorf("parseExpression(%s) = %v; want %v", input, result, expected)
	}

	if result[0] != expected {
		t.Errorf("parseExpression(%s) = %v; want %v", input, result, expected)
	}
}

func Test_Complex_parseExpression(t *testing.T) {
	input := "1d6+2d4+2+1d8+4"
	expectedDie := []Die{
		{Sides: 6, Count: 1, Modifier: 0},
		{Sides: 4, Count: 2, Modifier: 2},
		{Sides: 8, Count: 1, Modifier: 4},
	}
	result, err := parseExpression(input)

	if err != nil {
		t.Errorf("parseExpression(%s) = %v; want %v", input, result, expectedDie)
	}

	if len(result) != len(expectedDie) {
		t.Errorf("parseExpression(%s) = %v; want %v", input, result, expectedDie)
	}

	for i := range result {
		if result[i] != expectedDie[i] {
			t.Errorf("parseExpression(%s) = %v; want %v", input, result, expectedDie)
			break
		}
	}
}

func TestOutOfBoundsReturnsError(t *testing.T) {
	r := &mockRand{value: 8}
	input := "1d6"
	expected := -1
	result, err := RollDice(r, input)

	if result.Total != expected {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
		return
	}

	if err == nil {
		t.Errorf("RollDice(%s) = %d; want error", input, result.Total)
		return
	}

	if err.Error() != "out of bounds somehow! 9 > 6" {
		t.Errorf("RollDice(%s) = %s; want error", input, err)
		return
	}
}

func TestAddingModifiers(t *testing.T) {
	r := &mockRand{value: 3}
	input := "1d6+2"
	expected := 6
	result, err := RollDice(r, input)

	if err != nil {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}

	if result.Total != expected {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}
}

func TestNegativeModifier(t *testing.T) {
	r := &mockRand{value: 3}
	input := "1d6-2"
	expected := 2
	result, err := RollDice(r, input)
	if err != nil {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}
	if result.Total != expected {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}
}

func TestMissingDieCount(t *testing.T) {
	r := &mockRand{value: 3}
	input := "d6"
	expected := 4
	result, err := RollDice(r, input)
	if err != nil {
		t.Errorf("RollDice(%s) = %d; want %d; error %v", input, result.Total, expected, err)
	}
	if result.Total != expected {
		t.Errorf("RollDice(%s) = %d; want %d", input, result.Total, expected)
	}
}
