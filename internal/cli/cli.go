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
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Plugin API Design Spec Sheet
// Unique Name
// tags - string[]
// scores - Array(Object({tag:string, value: string}}))
// metadata - Array(Object({ key:string, value:string, type:string }))

type ApiCatalogConfig struct {
	Title string
}

var errNotAbsolute = errors.New("path is not absolute")

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

			fmt.Println(config.Title)
			fmt.Println(apiType)

			fr, err := filereader.New()
			if err != nil {
				log.Fatal("Failed to load filereader\n", err)
			}

			// set up the js script compiler
			cmp, err := compiler.New()
			if err != nil {
				log.Fatal("Error in setting up compiler: ", err)
			}

			// loading up the plugins and corresponding rules
			pluginManager := pluginmanager.New(fr)
			if err := pluginManager.LoadBuiltinPlugin(); err != nil {
				log.Fatal(err)
			}

			var apiSchemaFile map[string]interface{}
			fd, err := fr.ReadFileAdvanced(apiURL, &apiSchemaFile)
			if err != nil {
				log.Fatal("Fail to read file\n", err)
			}

			switch apiType {
			case "openapi":
				loader := openapi3.NewLoader()
				doc, err := loader.LoadFromData(fd.Raw)
				if err != nil {
					log.Fatal(err)
				}
				if err := doc.Validate(loader.Context); err != nil {
					log.Fatal(err)
				}

				runCfg := &compiler.RunConfig{Type: apiType, ApiSchema: apiSchemaFile}
				// iterate over rule
				for _, p := range pluginManager.Rules {
					// read original code
					rawCode, err := pluginManager.ReadPluginCode(p.File)
					if err != nil {
						log.Fatal("Failed to : ", err)
					}
					// babel transpile
					code, err := cmp.Transform(rawCode)
					if err != nil {
						log.Fatal("Failed to : ", err)
					}

					// execute the code
					err = cmp.Run(code, runCfg)
					if err != nil {
						log.Fatal("Error in program: ", err)
					}
				}

			default:
				log.Fatal("Error api type not supported: ", apiType)
			}
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
