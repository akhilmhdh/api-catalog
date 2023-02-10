package cli

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/1-platform/api-catalog/internal/cli/pluginmanager"
	"github.com/spf13/cobra"
)

type ApiCatalogConfig struct {
	Title   string
	Rules   map[string]pluginmanager.PluginUserOverride
	Plugins pluginmanager.PluginConfFile
}

// cli flags
var apiType string
var apiSchemaURL string
var configFilePath string
var version string
var exportReportPath string

func Run(apiVersion string) {
	version = apiVersion

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the apic tests",
		Long:  "Marathon apic tests",
		Run:   runCommand,
	}

	runCmd.Flags().StringVarP(&apiType, "apiType", "a", "", "Your API Type. Allowed values: openapi")
	runCmd.MarkFlagRequired("apiType")

	runCmd.PersistentFlags().StringVar(&apiSchemaURL, "schema", "", "URL or local file containing spec sheet")
	runCmd.MarkPersistentFlagRequired("schema")

	runCmd.PersistentFlags().StringVar(&configFilePath, "config", ".", "Path to apic configuration file")
	runCmd.PersistentFlags().StringVar(&exportReportPath, "export", "", "File path to export data")

	rootCmd := &cobra.Command{
		Use:     "apic",
		Short:   "One shot cli for your api schema security,performance and quality check",
		Version: version,
	}
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
