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
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTableRollCommand(t *testing.T) {
	// Get the absolute path to the test fixtures directory
	testFixturesPath := filepath.Join("..", "testfixtures")

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "roll with file",
			args:           []string{"table", "roll", filepath.Join(testFixturesPath, "sample.csv")},
			expectedOutput: "Name:",
			expectError:    false,
		},
		{
			name:           "roll with stdin",
			args:           []string{"table", "roll", "-"},
			expectedOutput: "Name:",
			expectError:    false,
		},
		{
			name:           "roll with plain output",
			args:           []string{"table", "roll", filepath.Join(testFixturesPath, "sample.csv"), "--plain"},
			expectedOutput: ",",
			expectError:    false,
		},
		{
			name:           "roll with non-existent file",
			args:           []string{"table", "roll", "nonexistent.csv"},
			expectedOutput: "Error:",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Set up stdin if needed
			if tt.args[2] == "-" {
				file, err := os.Open(filepath.Join(testFixturesPath, "sample.csv"))
				if err != nil {
					t.Fatalf("Failed to open test file: %v", err)
				}
				defer file.Close()
				rootCmd.SetIn(file)
			}

			// Execute the command
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			// Check for expected error
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output
			output := buf.String()
			if !tt.expectError && !bytes.Contains([]byte(output), []byte(tt.expectedOutput)) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestTableRollCommandWithEmptyFile(t *testing.T) {
	// Get the absolute path to the test fixtures directory
	testFixturesPath := filepath.Join("..", "testfixtures")

	// Create an empty CSV file
	emptyFile := filepath.Join(testFixturesPath, "empty.csv")
	err := os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	defer os.Remove(emptyFile)

	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Execute the command
	rootCmd.SetArgs([]string{"table", "roll", emptyFile})
	err = rootCmd.Execute()

	// Check for expected error
	if err == nil {
		t.Error("Expected error for empty file but got none")
	}

	// Check output
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Error: CSV file is empty")) {
		t.Errorf("Expected error message about empty file, got %q", output)
	}
}
