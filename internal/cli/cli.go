package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/1-platform/api-catalog/internal/cli/compiler"
	"github.com/1-platform/api-catalog/internal/cli/filereader"
	"github.com/1-platform/api-catalog/internal/cli/pluginmanager"
	"github.com/1-platform/api-catalog/internal/cli/reportmanager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ApiCatalogConfig struct {
	Title   string
	Rules   map[string]pluginmanager.PluginUserOverride
	Plugins pluginmanager.PluginConfFile
}

func Run() {
	// cli flags
	var apiType string
	var apiURL string
	var configFilePath string

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the apic tests",
		Long:  "Marathon apic tests",
		Run: func(cmd *cobra.Command, args []string) {
			// find config file and load up the config
			var config ApiCatalogConfig
			viper.AddConfigPath(configFilePath)
			viper.SetConfigName("apic")

			if err := viper.ReadInConfig(); err != nil {
				log.Fatal("File not found", err)
			}

			if err := viper.Unmarshal(&config); err != nil {
				log.Fatal("Error in config file", err)
			}

			fr, err := filereader.New()
			if err != nil {
				log.Fatal("Failed to load filereader\n", err)
			}

			// setup cli logger
			logger := NewCliLogger()

			// set up the js script compiler
			cmp, err := compiler.New(logger)
			if err != nil {
				log.Fatal("Error in setting up compiler: ", err)
			}

			// loading up the plugins and corresponding rules
			pManager := pluginmanager.New(fr, apiType)
			if err := pManager.LoadBuiltinPlugin(); err != nil {
				log.Fatal(err)
			}
			logger.Completed("Loaded builtin plugins")

			if err := pManager.LoadUserPlugins(config.Plugins); err != nil {
				log.Fatal(err)
			}
			logger.Completed("Loaded user defined plugins")

			if err := pManager.OverrideRules(config.Rules); err != nil {
				log.Fatal(err)
			}

			var apiSchemaFile map[string]interface{}
			raw, err := fr.ReadFileReturnRaw(apiURL, &apiSchemaFile)
			if err != nil {
				log.Fatal("Failed to read file\n", err)
			}
			logger.Completed("Read and parsed API schema file")

			rm := reportmanager.New()

			// validation
			switch apiType {
			case "openapi":
				if err := ValidateOpenAPI(raw, apiSchemaFile, logger); err != nil {
					log.Fatal("Failed to validate openapi schema/n", err)
				}
				logger.Success("OpenAPI validation check passed")
				// iterate over rule
			default:
				logger.Error(fmt.Sprintf("Error api type not supported: %s", apiType))
				os.Exit(0)
			}

			for rule, opt := range pManager.Rules {
				if opt.Disable {
					logger.Info(fmt.Sprintf("%s has been disabled", rule))
					continue
				}

				// read original code
				rawCode, err := pManager.ReadPluginCode(opt.File)
				if err != nil {
					log.Fatal("Failed to : ", err)
				}

				// babel transpile
				code, err := cmp.Transform(rawCode)
				if err != nil {
					log.Fatal("Failed to : ", err)
				}

				// creating config for each rule because we also want rule name of each score and report setter
				runCfg := &compiler.RunConfig{
					Type:      apiType,
					ApiSchema: apiSchemaFile,
					SetScore: func(category string, score float32) {
						rm.SetScore(rule, reportmanager.Score{Category: category, Value: score})
					},
					Report: func(body *reportmanager.ReportDef) {
						rm.PushReport(rule, *body)
					},
				}

				// execute the code
				err = cmp.Run(code, runCfg, opt.Options)
				if err != nil {
					if errors.Is(err, compiler.ErrExceptionInPluginCode) {
						logger.Error(fmt.Sprintf("%s threw an exception", rule))
						logger.Error(err.Error())
					} else {
						log.Fatal("Failed to: ", err)
					}
				}
				logger.Info(fmt.Sprintf("%s check completed", rule))
			}

			logger.Title("Reports")
			for rule, r := range rm {
				for _, report := range r.Reports {
					logger.Report(rule, report.Method, report.Path, report.Message)
					logger.Divider()
				}
			}

			logger.Title("Score Card")
			scores := rm.GetTotalScore()
			logger.ScoreCard(scores)
		},
	}

	runCmd.Flags().StringVarP(&apiType, "apiType", "a", "", "Your API Type. Allowed values: rest | graphql")
	runCmd.MarkFlagRequired("apiType")

	runCmd.PersistentFlags().StringVar(&apiURL, "url", "", "URL or local file containing spec sheet")
	runCmd.MarkPersistentFlagRequired("url")

	runCmd.PersistentFlags().StringVar(&configFilePath, "config", ".", "Path to apic configuration file")

	rootCmd := &cobra.Command{
		Use:   "apic",
		Short: "One shot cli for your api schema security,performance and quality check",
	}
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
