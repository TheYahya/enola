package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type cmdOptions struct {
	username   string
	site       string
	outputPath string
}

var rootCmd = &cobra.Command{
	Use:   "enola {username}",
	Short: "A command-line tool to find username on websites",
	Args:  validateArgs,
	Run:   runCommand,
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("missing required argument: username")
	}
	return nil
}

func runCommand(cmd *cobra.Command, args []string) {
	username := args[0]
	siteFlag := cmd.Flag("site")
	outputPath := cmd.Flag("output")

	options := cmdOptions{
		username:   username,
		site:       siteFlag.Value.String(),
		outputPath: outputPath.Value.String(),
	}

	findAndShowResult(options)
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringP(
		"site",
		"s",
		"",
		"to only search an specific site",
	)
	rootCmd.Flags().StringP(
		"output",
		"o",
		"",
		"output path, supports json and csv, eg: C:\\test\\test.json",
	)
}
