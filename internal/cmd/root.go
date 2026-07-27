package cmd

import (
	"fmt"
	"os"

	"github.com/dagimg-dot/floww/internal/cmd/add"
	"github.com/dagimg-dot/floww/internal/cmd/apply"
	"github.com/dagimg-dot/floww/internal/cmd/edit"
	initpkg "github.com/dagimg-dot/floww/internal/cmd/init"
	"github.com/dagimg-dot/floww/internal/cmd/list"
	"github.com/dagimg-dot/floww/internal/cmd/remove"
	"github.com/dagimg-dot/floww/internal/cmd/validate"
	"github.com/dagimg-dot/floww/internal/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "floww",
	Short: "floww - your workflow automations in one place",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(FLOWW_ART)
		_ = cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Root().Flags().Changed("version") {
			fmt.Println(utils.VersionDisplay())
			os.Exit(0)
		}
		level, _ := cmd.Flags().GetString("log-level")
		SetupLogging(level)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("log-level", "l", "WARNING", "Log level (DEBUG, INFO, WARNING, ERROR)")
	rootCmd.Flags().BoolP("version", "v", false, "version for floww")

	rootCmd.AddCommand(initpkg.Command)
	rootCmd.AddCommand(list.Command)
	rootCmd.AddCommand(add.Command)
	rootCmd.AddCommand(edit.Command)
	rootCmd.AddCommand(remove.Command)
	rootCmd.AddCommand(validate.Command)
	rootCmd.AddCommand(apply.Command)
}

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
