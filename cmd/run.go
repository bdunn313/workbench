/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/bdunn313/workbench/embedpkg"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runScript(args[0])
	},
}

func init() {
	zxCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runScript(selectedScript string) {
	scripts, err := embedpkg.GetScripts()
	if err != nil {
		fmt.Println("Error getting scripts:", err)
		return
	}
	for _, script := range scripts {
		if script.Name == selectedScript {
			if err := executeScript(script); err != nil {
				fmt.Println("Error executing script:", err)
			}
			return
		}
	}
}

func executeScript(script embedpkg.Script) error {
	// Open the embedded script file
	scriptFile, err := embedpkg.Scripts().Open(script.Path)
	if err != nil {
		return fmt.Errorf("failed to open script: %w", err)
	}
	defer scriptFile.Close()

	// Create a temporary file to store the script
	tempFile, err := os.CreateTemp("", script.Name+"-*."+script.Extension)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write the embedded script to the temporary file
	if _, err := io.Copy(tempFile, scriptFile); err != nil {
		return fmt.Errorf("failed to write script to temporary file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	availableCommand, err := getAvailableCommand()
	if err != nil {
		return fmt.Errorf("Error getting available command: %w", err)
	}
	// Execute the script using the zx command
	cmd := exec.Command(availableCommand, tempFile.Name())
	cmd.Stdout = os.Stdout // Redirect stdout to the console
	cmd.Stderr = os.Stderr // Redirect stderr to the console

	fmt.Printf("Executing script: %s\n", script.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute script: %w", err)
	}
	return nil
}

func getAvailableCommand() (string, error) {
	isZxAvailable := isCommandAvailable("zx")
	if isZxAvailable {
		return "zx", nil
	}

	isNpxAvailable := isCommandAvailable("npx")
	if isNpxAvailable {
		return "npx zx", nil
	}
	fmt.Println("zx is not installed. Please install zx to use workbench.")
	return "", errors.New("zx is not installed")
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	if err != nil {
		return false // Command not found
	}
	return true // Command found
}
