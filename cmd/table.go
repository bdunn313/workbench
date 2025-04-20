/*
Copyright © 2024 Brad Dunn <brad@braddunn.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// tableCmd represents the table command
var tableCmd = &cobra.Command{
	Use:   "table",
	Short: "Roll on a table",
	Long: `Roll on a table from a CSV file or stdin.
	
Examples:
  workbench table roll path/to/table.csv
  cat table.csv | workbench table roll -`,
}

// tableRollCmd represents the roll subcommand
var tableRollCmd = &cobra.Command{
	Use:   "roll [file]",
	Short: "Roll on a table from a CSV file or stdin",
	Long: `Roll on a table from a CSV file or stdin.
	
If file is "-", read from stdin.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		plain, err := cmd.Flags().GetBool("plain")
		if err != nil {
			return fmt.Errorf("error getting plain flag: %w", err)
		}

		var reader io.Reader
		if file == "-" {
			reader = cmd.InOrStdin()
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}
			defer f.Close()
			reader = f
		}

		csvReader := csv.NewReader(reader)
		records, err := csvReader.ReadAll()
		if err != nil {
			return fmt.Errorf("error reading CSV: %w", err)
		}

		if len(records) == 0 {
			return fmt.Errorf("CSV file is empty")
		}

		// Skip header row if it exists
		startRow := 0
		if len(records) > 1 {
			startRow = 1
		}

		// Randomly select a row
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		selectedRow := r.Intn(len(records)-startRow) + startRow
		selectedRecord := records[selectedRow]

		if plain {
			// Print as comma-separated values
			for i, field := range selectedRecord {
				if i > 0 {
					cmd.Print(",")
				}
				cmd.Print(field)
			}
			cmd.Println()
		} else {
			// Print formatted output
			if startRow == 1 {
				cmd.Println("\nSelected row:")
				for i, field := range selectedRecord {
					if i < len(records[0]) {
						cmd.Printf("%s: %s\n", records[0][i], field)
					} else {
						cmd.Printf("Column %d: %s\n", i+1, field)
					}
				}
			} else {
				cmd.Println("\nSelected row:")
				for i, field := range selectedRecord {
					cmd.Printf("Column %d: %s\n", i+1, field)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)
	tableCmd.AddCommand(tableRollCmd)
	tableRollCmd.Flags().BoolP("plain", "p", false, "Enable plain output")
}
