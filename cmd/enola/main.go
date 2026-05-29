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

func validateArgs(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("missing required argument: username")
	}
	return nil
}

func runCommand(cmd *cobra.Command, args []string) {
	options := cmdOptions{
		username:   args[0],
		site:       cmd.Flag("site").Value.String(),
		outputPath: cmd.Flag("output").Value.String(),
	}
	findAndShowResult(options)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
		"output path, supports json and csv, eg: ./enola.json",
	)
}
