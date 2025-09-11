/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var isActive bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List todo items",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiRoot := viper.GetString("api-root")
		return listAction(os.Stdout, apiRoot, isActive)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&isActive, "active", "a", false, "Filter list to show active tasks only")
}

func listAction(out io.Writer, apiRoot string, isActive bool) error {
	items, err := getAll(apiRoot)
	if err != nil {
		return err
	}

	if isActive {
		var activeItems []item
		for _, i := range items {
			if !i.Done {
				activeItems = append(activeItems, i)
			}
		}

		return printAll(out, activeItems)
	}

	return printAll(out, items)
}

func printAll(out io.Writer, items []item) error {
	w := tabwriter.NewWriter(out, 3, 2, 0, ' ', 0)
	for k, v := range items {
		done := "-"
		if v.Done {
			done = "X"
		}

		fmt.Fprintf(w, "%s\t%d\t%s\t\n", done, k+1, v.Task)
	}
	return w.Flush()
}
