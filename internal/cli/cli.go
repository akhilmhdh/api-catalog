package cli

import (
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"os"

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

			cmp, err := NewCompiler()
			if err != nil {
				log.Fatal("Error in setting up compiler: ", err)
			}

			switch apiType {
			case "openapi":
				// load the document and validate
				// for openapi we use kin-openapi package
				// Kudos: https://github.com/getkin/kin-openapi
				url, err := url.ParseRequestURI(apiURL)
				isValidURL := err == nil
				var doc *openapi3.T
				loader := openapi3.NewLoader()

				if isValidURL {
					doc, err = loader.LoadFromURI(url)
					if err != nil {
						log.Fatal("failed to parse document\n", err)
					}
					err = doc.Validate(loader.Context)
					if err != nil {
						log.Fatal("Invalid swagger document", err)
					}
				} else {
					doc, err = loader.LoadFromFile(apiURL)
					if err != nil {
						log.Fatal("failed to parse document\n", err)
					}
					err = doc.Validate(loader.Context)
					if err != nil {
						log.Fatal("Invalid swagger document\n", err)
					}
				}

				// this will soon be a root config dir specially for apic plugins
				files, err := os.ReadDir("./internal/builtin")
				if err != nil {
					log.Fatal("Failed to open builtin plugins dir: ", err)
				}

				for _, file := range files {
					if !file.IsDir() {
						data, err := os.ReadFile(fmt.Sprintf("./internal/builtin/%s", file.Name()))
						if err != nil {
							log.Fatal("Failed to open builtin plugin file: ", file.Name())
						}
						code, err := cmp.Transform(string(data))
						if err != nil {
							log.Fatal("Failed to : ", err)
						}

						err = cmp.Run(code)
						if err != nil {
							log.Fatal("Error in program: ", err)
						}
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
