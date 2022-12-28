package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Run() {
	var apiType string
	var apiURL string

	rootCmd := &cobra.Command{
		Use:   "apc",
		Short: "One shot cli for your api schema security,performance and quality check",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(apiType)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&apiType, "apiType", "a", "", "Your API Type. Allowed values: rest | graphql")
	rootCmd.MarkPersistentFlagRequired("apiType")

	rootCmd.PersistentFlags().StringVar(&apiURL, "url", "", "URL or local file containing spec sheet")
	rootCmd.MarkPersistentFlagRequired("url")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
