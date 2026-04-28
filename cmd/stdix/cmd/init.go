package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/codref/stdix/internal/config"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialise stdix in the current directory",
		Long:  `Creates a .stdix.yaml configuration file in the current working directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			if config.Exists(cwd) {
				fmt.Fprint(cmd.OutOrStdout(), ".stdix.yaml already exists. Overwrite? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			cfg := config.Default()
			if err := config.Save(cwd, cfg); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Created .stdix.yaml")
			return nil
		},
	}
}
