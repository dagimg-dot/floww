package init

import (
	"fmt"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/spf13/cobra"
)

var (
	createExample bool
	fileType      string
)

var Command = &cobra.Command{
	Use:   "init",
	Short: "Initialize floww configuration",
	Long: `Create the configuration directory (~/.config/floww/), the workflows
sub-directory, and the default config file.

Use --example to also create a sample workflow file.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg := config.NewConfigManager()
		if err := cfg.Init(createExample, fileType); err != nil {
			return fmt.Errorf("initialization failed: %w", err)
		}
		cmd.Println("Initialized config at " + cfg.ConfigPath())
		return nil
	},
}

func init() {
	Command.Flags().BoolVarP(&createExample, "example", "e", false, "Create an example workflow")
	Command.Flags().StringVarP(&fileType, "type", "t", "yaml", "File format for the example workflow")
}
