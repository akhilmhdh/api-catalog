package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Run() {
	rootCmd := &cobra.Command{
		Use:   "apc",
		Short: "One shot cli for your api schema security,performance and quality check",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("API Catalog CLI is running....")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
