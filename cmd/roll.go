/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// rollCmd represents the roll command
var rollCmd = &cobra.Command{
	Use:   "roll",
	Short: "Roll some dice",
	Long: `Supports complex dice rolls, such as:
		roll 2d6+1d4+2`,
	Run: func(cmd *cobra.Command, args []string) {
		expression := args[0] // assuming the expression is the first argument
		fmt.Println()
		fmt.Println("Rolling:", expression)
		fmt.Println()
		result, err := RollDice(rand.New(rand.NewSource(time.Now().UnixNano())), expression)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("-------------------")
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(rollCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rollCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rollCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type RandIntn interface {
	Intn(n int) int
}

type Die struct {
	Sides    int
	Count    int
	Modifier int
}

func parseExpression(expression string) ([]Die, error) {
	var dice []Die
	var modifier int
	for _, die := range strings.Split(expression, "+") {
		if strings.Contains(die, "d") {
			parts := strings.Split(die, "d")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid die expression: %s", die)
			}
			count, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid die count: %s", parts[0])
			}
			sides, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid die sides: %s", parts[1])
			}
			dice = append(dice, Die{Count: count, Sides: sides})
		} else {
			value, err := strconv.Atoi(die)
			if err != nil {
				return nil, fmt.Errorf("invalid modifier: %s", die)
			}
			dice[len(dice)-1].Modifier = value
			modifier += value
		}

	}
	return dice, nil
}

type RollResult struct {
	Total     int
	RolledDie string
}

func RollDice(r RandIntn, expression string) (int, error) {
	dice, err := parseExpression(expression)
	if err != nil {
		return -1, err
	}
	var results []RollResult
	var total int
	var max int
	for _, die := range dice {
		diceTotal, diceMax, result := generateResult(&die, r)
		total += diceTotal
		max += diceMax
		results = append(results, result)
	}
	printRoll(total, results)
	if total >= max {
		return -1, fmt.Errorf("out of bounds somehow! %d > %d", total, max)
	}
	return total, nil
}

func printRoll(total int, results []RollResult) {
	width := int(math.Log10(float64(total))) + 1
	for _, result := range results {
		fmt.Printf("%*d %8s\n", width, result.Total, result.RolledDie)
	}
}

func generateResult(die *Die, r RandIntn) (int, int, RollResult) {
	var diceTotal int
	var diceMax int
	for i := 0; i < die.Count; i++ {
		diceTotal += r.Intn(die.Sides-1) + 1 + die.Modifier
		diceMax += die.Sides + die.Modifier
	}
	var strRepresentation string
	if die.Modifier != 0 {
		strRepresentation = fmt.Sprintf("%dd%d+%d", die.Count, die.Sides, die.Modifier)
	} else {
		strRepresentation = fmt.Sprintf("%dd%d", die.Count, die.Sides)
	}

	return diceTotal, diceMax, RollResult{Total: diceTotal, RolledDie: strRepresentation}
}
