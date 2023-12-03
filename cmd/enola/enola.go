package main

import (
	"errors"
	"fmt"
	"github.com/theyahya/enola"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

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

type elonaCommandOption struct {
	site       string
	outputPath string
	printFound bool
}

func runCommand(cmd *cobra.Command, args []string) {
	username := args[0]

	printFound, err := strconv.ParseBool(cmd.Flag("print-found").Value.String())
	if err != nil {
		fmt.Printf("error: %v", enola.ErrInvalidFlag)
		os.Exit(1)
	}
	option := elonaCommandOption{
		site:       cmd.Flag("site").Value.String(),
		outputPath: cmd.Flag("output").Value.String(),
		printFound: printFound,
	}

	findAndShowResult(username, &option)
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
		"relative path of output folder",
	)

	rootCmd.Flags().BoolP(
		"print-found",
		"p",
		false,
		"output sites where the username was found",
	)
}
