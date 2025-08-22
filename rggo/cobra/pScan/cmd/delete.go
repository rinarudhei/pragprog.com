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
	"pragprog.com/rggo/cobra/pScan/scan"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:          "delete <host1> ... <hostn>",
	Aliases:      []string{"d"},
	Short:        "Delete host(s) from host list",
	SilenceUsage: true,
	Args:         cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostsFile := viper.GetString("hosts-file")

		return deleteAction(os.Stdout, hostsFile, args)
	},
}

func init() {
	hostsCmd.AddCommand(deleteCmd)
}

func deleteAction(out io.Writer, hostsFile string, args []string) error {
	hl := &scan.HostList{}
	if err := hl.Load(hostsFile); err != nil {
		return err
	}
	for _, h := range args {
		if err := hl.Remove(h); err != nil {
			return err
		}

		fmt.Fprintf(out, "Deleted host: %s\n", h)
	}

	return hl.Save(hostsFile)
}
