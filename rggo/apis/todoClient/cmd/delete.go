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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:          "delete <id>",
	Short:        "Delete a task in the list.",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiRoot := viper.GetString("api-root")

		return deleteAction(os.Stdout, apiRoot, args[0])
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteAction(out io.Writer, apiRoot, id string) error {
	if err := deleteItem(apiRoot, id); err != nil {
		return err
	}

	return printDelete(out, id)
}

func printDelete(out io.Writer, id string) error {
	_, err := fmt.Fprintf(out, "Deleted task id %s.\n", id)

	return err
}
