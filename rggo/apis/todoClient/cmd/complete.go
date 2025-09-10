/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:          "complete <id>",
	Short:        "Complete a task in the list.",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiRoot := viper.GetString("api-root")

		return completeAction(os.Stdout, apiRoot, args[0])
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}

func completeAction(out io.Writer, apiRoot string, id string) error {
	if err := completeItem(apiRoot, id); err != nil {
		return err
	}

	return printComplete(out, id)
}

func printComplete(out io.Writer, id string) error {
	_, err := fmt.Fprintf(out, "Task %s completed.\n", id)

	return err
}
